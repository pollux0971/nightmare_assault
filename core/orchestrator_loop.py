"""core.orchestrator_loop — beat 主迴圈（U15 + SK04 Narrative Progress Kernel 整合）。

兩條流程，**入口單一分流**（不做 flag 分支地獄）：
- _step_kernel：ProgressKernel 決定推進（EventPatch + obligations），story 只 realize；PatchValidator 提交。
- _step_legacy：原 LLM 自由流程（warden→event 抽取→orchestrator→story）。

`ENABLE_PROGRESS_KERNEL` 預設 ON。kernel/graph/validation 失敗 → log + 回退 legacy（不 crash）。
"""
from __future__ import annotations

import json
import logging
from typing import Any

from core.agents.compactor import Compactor
from core.agents.orchestrator import run_orchestrator
from core.agents.setup import run_setup
from core.agents.story import run_story
from core.agents.warden import run_warden
from core.constants import (
    BEAT_WINDOW_SIZE, CONTEXT_THRESHOLD_L1, SUMMARY_TOKEN_CAP, ENABLE_PROGRESS_KERNEL,
    EVT_BEAT_COMPLETED, EVT_CONTEXT_THRESHOLD, EVT_ENDING_TRIGGERED,
    EVT_RULE_VIOLATION, EVT_SKILL_CLAIMED, EVT_NPC_EVOLVED,
)
from core.attractors import dominant_ending
from core.events import event_extract
from core.memory.summary import estimate_tokens
from core.patch_validator import PatchValidator
from core.progress_context import ContextBuilder
from core.progress_kernel import ProgressKernel
from core.scene_graph import StaticOpeningSceneGraphProvider, GeneratedSceneGraphProvider
from core import progress_bridge as bridge

log = logging.getLogger("nightmare.beatloop")


class BeatLoop:
    """一場遊戲的後端核心迴圈。start() 開局，step() 推進一個 beat。"""

    def __init__(self, caller, blackboard, db, signal_bus=None, run_id: str = "run-1",
                 scene_graph_provider=None, use_kernel: bool | None = None):
        self.caller = caller
        self.bb = blackboard
        self.db = db
        self.bus = signal_bus
        self.run_id = run_id
        self.compactor = Compactor(caller)
        self.beat_number = 0
        self.ended = False
        self.ending: dict | None = None
        self.last_story = ""
        self.known_npcs: list[str] = []
        self.known_locations: list[str] = []
        self.known_items: list[str] = []

        # ── Progress Kernel（旁路層；kernel 在 start() 依 setup 輸出建主題化圖）──
        self.use_kernel = ENABLE_PROGRESS_KERNEL if use_kernel is None else use_kernel
        self._explicit_provider = scene_graph_provider   # 明確指定（測試/特例）
        self._provider = None
        self._kernel = None
        self._validator = PatchValidator()
        self._ctx_builder = ContextBuilder()
        self._game_state = None
        log.info("ProgressKernel enabled (pending setup): %s", self.use_kernel)

        # ── 配置中心（P4）：config-first story prompt 來源（旁路；flag off / 失敗 → static）──
        self._prompt_source = None
        self._last_prompt_meta: dict | None = None
        try:
            from core.config.runtime import ConfigPromptSource
            ps = ConfigPromptSource(self.db.config_store(), getattr(caller, "loader", None))
            if ps.enabled:
                self._prompt_source = ps
            log.info("ConfigPromptSource enabled: %s", bool(self._prompt_source))
        except Exception as e:                      # 配置層失敗不可拖垮 beat loop（B8）
            log.warning("ConfigPromptSource init failed, static story prompt: %s", e)

    def _snapshot_run_config(self):
        """P6：開局時把每個 enabled agent 的 active config + 編譯 prompt hash 存進 run_config_snapshots。

        用 PromptComposer.preview（零 LLM）。受 ENABLE_RUN_CONFIG_SNAPSHOT 控制；失敗不影響開局（B8）。
        """
        try:
            from core.config.flags import run_config_snapshot_enabled
            from core.config.composer import PromptComposer
            store = self.db.config_store()
            if not run_config_snapshot_enabled(store):
                return
            profile = store.active_profile()
            composer = PromptComposer(store)
            # story + 其他已配置 agent（P7）；只快照有綁定 fragment 的 agent
            for agent in ("story", "warden", "orchestrator", "compactor", "setup"):
                if not store.get_bound_fragments(agent, profile):
                    continue
                cfg = store.get_agent_config(agent, profile) or {}
                compiled = composer.preview(agent, profile, {})
                store.write_run_config_snapshot(
                    self.run_id, profile, agent,
                    config_json={"agent_config": cfg, "model_settings": compiled.model_settings,
                                 "context_policy": compiled.context_policy},
                    compiled_prompt_hash=compiled.prompt_hash,
                    enabled_fragments=compiled.enabled_fragments,
                    compiled_prompt_preview=compiled.compiled_prompt[:500],
                )
            log.info("run config snapshot stored (profile=%s)", profile)
        except Exception as e:                       # 快照失敗不可拖垮開局
            log.warning("run config snapshot skipped: %s", e)

    def run_config_snapshots(self) -> list:
        """回傳本 run 已存的 config 快照（載入存檔時優先用此重現；P6/P5）。"""
        try:
            return self.db.config_store().get_run_config_snapshots(self.run_id)
        except Exception:
            return []

    def _story_system(self, ctx: dict | None = None):
        """取 config-first 的 story system prompt（None → 讓 SkillCaller 用 static SKILL.md）。"""
        if self._prompt_source is None:
            return None
        try:
            from core.config.story_prompt import map_context_to_variables
            system, meta = self._prompt_source.story_system_prompt(map_context_to_variables(ctx or {}))
            self._last_prompt_meta = meta
            log.info("story prompt source=%s profile=%s hash=%s",
                     meta.get("source"), meta.get("profile"), meta.get("prompt_hash"))
            # source=static 時回 None：交回 SkillCaller 讀 SKILL.md，行為與未啟用一致
            return system if meta.get("source") == "config" else None
        except Exception as e:
            log.warning("story prompt resolve failed, static fallback: %s", e)
            return None

    def _build_narrative_contract(self):
        """NC1：旁路敘事控制——從 setup 輸出組裝 NarrativeContract 存進 blackboard。

        受 ENABLE_NARRATIVE_CONTROL 控制（預設 OFF）；失敗不影響開局（B8）。
        """
        self._narrative_contract = None
        self._reveal_level = "hidden"            # NC4：context 全域揭露上限（一次升一階）
        # NR0：揭露橋接（旁路；flag OFF 時保持 None/空，行為與現況一致）
        self._reveal_ledger = None
        self._truth_index = {}
        self._known_clue_ids: set[str] = set()
        self._core_truth_ids: set[str] = set()       # NPC 不得引用的決定性真相
        self._npc_chat_debug: dict = {}              # 最近一次 NPC evidence 驗證統計（observation 用）
        # NR2：答債（旁路）
        self._answer_debt = None
        self._world = None                           # WorldModel（current_area/area/exit 的唯一權威）
        self._world_kernel_scene = None              # WorldModel 上次同步到的 kernel 場景（偵測真正移動）
        # NR5/NR6：母題冷卻 + 動機心跳（旁路）
        self._motif_tracker = None
        self._motive_heartbeat = None
        self._last_scene = None
        try:
            from core.constants import ENABLE_NARRATIVE_CONTROL
            if not ENABLE_NARRATIVE_CONTROL:
                return
            from core.narrative.contract import build_narrative_contract, store_contract
            # NC7：敘事控制參數可配置（safe profile 為預設）
            try:
                from core.narrative.config import get_narrative_config
                cfg = get_narrative_config(self.db.config_store())
            except Exception:
                cfg = None
            nc = build_narrative_contract(self.bb, config=cfg)
            store_contract(self.bb, nc)
            self._narrative_contract = nc
            log.info("narrative contract built (motive=%s)", nc.protagonist_motive.immediate_goal)
            # NR0：從 real_bible 建 RevealLedger（種子全 hidden）+ NPC 真相關鍵詞索引（要求 #5）
            try:
                from core.narrative.revelation import (
                    build_ledger_from_bible, build_truth_keyword_index,
                    write_ledger_to_revealed_bible)
                real = self.bb.snapshot().get("real_bible") or {}
                self._reveal_ledger = build_ledger_from_bible(real)
                self._truth_index = build_truth_keyword_index(real)
                # core 真相（決定性、對應 clue.core）= revelation_pool 最後一個碎片，NPC 不得引用
                pool = [f for f in (real.get("revelation_pool") or [])
                        if isinstance(f, dict) and f.get("id")]
                self._core_truth_ids = {pool[-1]["id"]} if pool else set()
                write_ledger_to_revealed_bible(self.bb, self._reveal_ledger)
                self._update_npc_truth_refs()        # P-plumbing：種出 NPC 可引用白名單
                log.info("reveal ledger seeded into revealed_bible (%d truths, %d keyword index)",
                         len(self._reveal_ledger.truths), len(self._truth_index))
            except Exception as e:                   # 橋接初始化失敗不影響開局
                log.warning("reveal ledger init skipped: %s", e)
            # NR2：答債追蹤器
            try:
                from core.narrative.answer_debt import AnswerDebtTracker
                self._answer_debt = AnswerDebtTracker()
            except Exception as e:
                log.warning("answer debt init skipped: %s", e)
            # WorldModel：抽象實體層（平行記憶，不取代 kernel）
            try:
                from core.world.model import WorldModel
                self._world = WorldModel()
            except Exception as e:
                log.warning("world model init skipped: %s", e)
            # NR5/NR6：母題冷卻 + 動機心跳
            try:
                from core.narrative.motif_tracker import MotifTracker, MotiveHeartbeat
                self._motif_tracker = MotifTracker()
                self._motive_heartbeat = MotiveHeartbeat()
            except Exception as e:
                log.warning("motif/motive init skipped: %s", e)
        except Exception as e:                       # 敘事控制失敗不影響開局
            log.warning("narrative contract build skipped: %s", e)

    def _init_kernel_from_setup(self):
        """依 setup 輸出建 scene graph provider：explicit → generated（主題化）→ static → legacy。"""
        snap = self.bb.snapshot()
        if self._explicit_provider is not None:
            candidates = [("explicit", lambda: self._explicit_provider)]
        else:
            candidates = [("generated", lambda: GeneratedSceneGraphProvider(snap)),
                          ("static", StaticOpeningSceneGraphProvider)]
        for name, make in candidates:
            try:
                provider = make()
                kernel = ProgressKernel.from_provider(provider)
                self._provider, self._kernel = provider, kernel
                log.info("ProgressKernel ready via %s provider (start_scene=%s)",
                         name, provider.start_scene())
                return
            except Exception as e:
                log.warning("scene graph provider '%s' failed: %s", name, e)
        self.use_kernel = False
        log.warning("all scene graph providers failed; kernel disabled, using legacy")

    # ── 對外 ────────────────────────────────────────────────────────────
    def start(self, opts: dict, on_event=None, on_progress=None, on_opening=None) -> dict:
        """開局：setup 生世界 → 產第一個 beat（決策點）。setup 失敗不降級（B8 往上拋）。"""
        _p = on_progress or (lambda *_: None)
        _p("setup")
        opening = run_setup(self.caller, self.bb, opts)
        self.db.create_run(self.run_id, theme=opts.get("theme"))
        self._snapshot_run_config()                 # P6：每場 run 存 config/prompt-hash 快照
        self._derive_known()
        self._build_narrative_contract()            # NC1：旁路敘事控制（flag OFF 時不動）
        if on_opening:
            on_opening(opening)

        if self.use_kernel:
            self._init_kernel_from_setup()         # 依 setup 主題建圖（失敗會關掉 use_kernel）

        if self.use_kernel and self._kernel is not None:
            self._game_state = bridge.init_game_state(self._provider)
            bridge.sync_to_blackboard(self._game_state, self.bb)
            self._seed_world_area()                  # WorldModel：起始區域當權威來源（種子）
            _p("story")
            narrative, dp = self._kernel_intro_beat(on_event)
        else:
            _p("orchestrator")
            newly = run_orchestrator(self.bb, self.beat_number,
                                     touched_ids=[], reached_locations=[], caller=self.caller)
            _p("story")
            narrative, dp = run_story(self.caller, self.bb, "（序章）", self.beat_number,
                                      directive=None, newly_revealed=newly, on_event=on_event,
                                      system_override=self._story_system(None))
        self.last_story = narrative
        self._safe_point(narrative, dp)
        return {"opening_sequence": opening, "narrative": narrative, "decision_point": dp}

    def step(self, player_decision: str, input_path: str = "free_text",
             on_event=None, on_progress=None) -> dict:
        """單一分流點：kernel 流程 vs legacy 流程。"""
        if self.ended:
            return {"ended": True, "ending": self.ending}
        if self.use_kernel and self._kernel is not None and self._game_state is not None:
            try:
                return self._step_kernel(player_decision, input_path, on_event, on_progress)
            except Exception as e:                      # kernel/validation 失敗 → 回退 legacy
                log.warning("kernel step failed, fallback to legacy: %s", e)
        return self._step_legacy(player_decision, input_path, on_event, on_progress)

    # ── Kernel 流程 ─────────────────────────────────────────────────────
    def _kernel_intro_beat(self, on_event):
        """開場 beat：介紹起始場景、呈現出口（含門），不替玩家決定（不跑 kernel resolve）。"""
        revealed = self.bb.snapshot().get("revealed_bible") or {}
        gs = self._game_state
        ctx = {
            "current_scene": gs.current_scene, "scene_phase": gs.scene_phase, "beat_number": 0,
            "recent_events": [], "committed_event": None,
            "narrative_obligations": [
                "介紹起始場景與可見的出口（包含一扇門），以恐怖氛圍鋪陳，"
                "最後提供玩家可選行動（如開門、檢查、呼喊），不要替玩家做決定。"],
            "forbidden_repeats": [], "new_clues": [], "new_items": [], "spawned_npcs": [],
            "visible_npcs": [], "relevant_clues": [], "relevant_items": [],
            "danger_level": 0, "revealed_bible": revealed,
            "instruction": "Realize the opening scene; present exits including a door; do not decide for the player.",
        }
        # UB6：序幕鉤子 + 真相種子（注入表層義務/種子；hidden_truth 結構性不外洩）
        try:
            from core.agents.opening import build_opening_context
            ctx = build_opening_context(self.bb, ctx)
            # NC2：Opening Director——啟用敘事控制時，把開場元素收斂到 budget（少而高價值）
            from core.constants import ENABLE_NARRATIVE_CONTROL
            if ENABLE_NARRATIVE_CONTROL and getattr(self, "_narrative_contract", None):
                from core.narrative.opening_director import apply_to_context
                ctx = apply_to_context(ctx, self._narrative_contract)
        except Exception as e:                       # 開場富化失敗也要能開局（保底用原 ctx）
            log.warning("opening hook enrich skipped: %s", e)
        return run_story(self.caller, self.bb, "（序章）", 0,
                         context_override=ctx, on_event=on_event,
                         system_override=self._story_system(ctx))

    def _step_kernel(self, player_decision, input_path, on_event, on_progress) -> dict:
        _p = on_progress or (lambda *_: None)
        from core.constants import ENABLE_NARRATIVE_CONTROL   # 全方法共用（避免 UnboundLocal）
        gs = self._game_state
        self._escape_step = "none"               # 本 beat 離開意圖（預設無）
        self._exit_intent = "none"               # Player Sovereignty：離開意圖分類
        self._exit_affordance = None             # 本 beat exit affordance（唯 end_campaign 進結局）
        # HB1/HB2/HE1：本 beat 揭露觀測（每 beat 重置）
        self._reveal_updates_this_beat = []
        self._evidence_events_this_beat = 0
        self._unmapped_evidence_this_beat = 0
        # P0 WorldStateFact：本 beat 新增的世界事實（每 beat 重置）
        self._new_world_facts_this_beat = []
        # WorldModel：本 beat 被新增/變更狀態的實體 id（每 beat 重置）
        self._changed_entities_this_beat = []

        # 1. warden（致命規則/技能/結局，僅玩家）
        _p("warden")
        verdict = run_warden(player_decision, self.bb, caller=self.caller)
        self._record_skill_claim(verdict)
        if verdict.rule_violation:
            self._emit(EVT_RULE_VIOLATION, verdict)
        if verdict.ending_triggered:
            self._emit(EVT_ENDING_TRIGGERED, verdict)

        # 2. ProgressKernel：決定本 beat 推進的事件（EventPatch + obligations）
        # NegativeIntentGuard（P0）：把玩家明確拒絕的目標傳進 kernel，避免選到「移動到該目標」的事件
        _negated = []
        if ENABLE_NARRATIVE_CONTROL:
            try:
                from core.narrative.negative_intent import negated_targets
                _negated = negated_targets(player_decision)
            except Exception:
                _negated = []
        _p("kernel")
        progress = self._kernel.resolve_player_action(
            player_decision, gs,
            warden={"directive": verdict.directive_to_story, "rule_violation": verdict.rule_violation,
                    "negated_targets": _negated},
        )

        # 3. story = realizer（context 只給 revealed + obligations，防暴雷）
        revealed = self.bb.snapshot().get("revealed_bible") or {}
        ctx = self._ctx_builder.build_story_context(gs, progress, revealed_bible=revealed)
        # NC3：Story Agent 降權（啟用敘事控制時，加 allowed/forbidden 元素 + beat_purpose + 動機）
        from core.constants import ENABLE_NARRATIVE_CONTROL
        if ENABLE_NARRATIVE_CONTROL and getattr(self, "_narrative_contract", None):
            try:
                # NC4：reveal ladder——依累積 evidence（線索數）一次升一階，不跳級
                from core.narrative.reveal_manager import allowed_reveal_for, next_level_no_skip
                target = allowed_reveal_for(len(getattr(gs, "clues", {}) or {}))
                self._reveal_level = next_level_no_skip(getattr(self, "_reveal_level", "hidden"), target)
                ctx["truth_reveal_limit"] = self._reveal_level
                # NC3：Story Agent 降權
                from core.narrative.story_control import apply_story_downgrade
                ctx = apply_story_downgrade(ctx, self._narrative_contract)
            except Exception as e:
                log.warning("narrative control (reveal/downgrade) skipped: %s", e)
            # NR2：答債——分類玩家提問，債≥2 時加償還義務進 story context
            self._answer_debt_tick(player_decision, ctx)
            # Player Sovereignty：解析成 exit affordance（ExitResolver 不直接 ending；只 end_campaign 進結局）
            from core.narrative.exit_resolver import resolve_exit_intent, resolve_exit_affordance
            self._exit_intent = resolve_exit_intent(player_decision)
            self._exit_affordance = resolve_exit_affordance(player_decision)
            self._escape_step = self._exit_intent     # observation 沿用此欄位
            # NR5：母題冷卻——換場景重置；把超用母題注入 context（story 須演化或換意象）
            self._motif_cooldown_pre(gs, ctx)
            # NR6：動機心跳——逾期未提動機 → 加提醒義務
            self._motive_heartbeat_pre(ctx)
        _p("story")
        # HD1：story + QualityGate repair pipeline（check → repair once → deterministic fallback）
        narrative, dp, self._quality_meta = self._run_story_with_repair(
            player_decision, gs, ctx, on_event)
        # NR7/HD2：表層消毒（後端，render 前；flag-gated）——narrative + 選項 + situation_recap
        narrative = self._sanitize_surface(narrative)
        dp = self._sanitize_decision_point(dp)
        # NR5/NR6：登記本 beat 的母題與動機提及（供下一 beat 冷卻/心跳判斷）
        if ENABLE_NARRATIVE_CONTROL and getattr(self, "_narrative_contract", None):
            self._motif_motive_post(narrative, ctx)

        # 4. PatchValidator.commit（≥1 delta 否則拒；kernel 已保證 dummy 有 delta）
        self._game_state = self._validator.apply(progress.patch, gs)
        bridge.sync_to_blackboard(self._game_state, self.bb)
        self.bb.game_meta = {**self.bb.game_meta,
                             "progress_state": bridge.to_snapshot_dict(self._game_state)}

        # NR0/HB1：揭露橋接——kernel 線索 → bridge；若本 beat 無 reveal 變化且玩家在調查 → 保底 evidence
        if ENABLE_NARRATIVE_CONTROL and self._reveal_ledger is not None:
            _changed = self._revelation_tick(self._game_state)
            self._story_evidence_tick(player_decision, narrative, _changed)
        # P0 WorldStateFact：從 story 敘事抽世界事實（NPC fact 由 bridge_npc_evidence 處理）
        if ENABLE_NARRATIVE_CONTROL:
            self._world_facts_tick(narrative, source="story")
            self._world_model_tick(player_decision, narrative, dp)   # WorldModel：實體記憶

        self.last_story = narrative
        # Player Sovereignty：離開意圖處理（ambiguous → ExitOffer；retreat → 降危險；皆不自動收束）
        _nc_on = ENABLE_NARRATIVE_CONTROL and getattr(self, "_narrative_contract", None)
        if _nc_on:
            # 用**已套用 patch 的** self._game_state（gs 是 pre-apply 舊物件，會覆蓋本 beat 進度）
            dp = self._apply_exit_intent(self._exit_intent, self._game_state, dp)
        self._safe_point(narrative, dp)
        self._log_beat(progress)

        if verdict.ending_triggered or verdict.rule_violation:
            # warden 硬觸發（致命/不可逆）——允許結局（Player Sovereignty 原則 5）
            self.ended = True
            self.ending = {"type": verdict.ending_triggered or "death_physical",
                           "soft": bool(verdict.ending_is_soft), "via": "warden"}
        elif not self.ended:
            if _nc_on:
                # Player Sovereignty：吸引子拉力**不自動收束**；唯 end_campaign affordance 才結算
                from core.narrative.exit_resolver import END_CAMPAIGN
                if getattr(self, "_exit_affordance", None) and \
                        self._exit_affordance.affordance == END_CAMPAIGN:
                    self._trigger_player_ending(player_decision, gs)
            else:
                # flag OFF：保留原 attractor 自動收束行為（向後相容）
                et = dominant_ending(self._game_state)
                if et:
                    self.ended = True
                    self.ending = {"type": et, "soft": False, "via": "attractor"}
                    self._emit(EVT_ENDING_TRIGGERED, et)

        self._finalize_ending()
        dp = self._enforce_ended_invariant(dp)       # HA1：ended ⇒ 無 options
        return {"narrative": narrative, "decision_point": dp, "warden": verdict,
                "committed_event": progress.committed_event,
                "progress_delta": progress.patch.progress_delta,
                # HB2：統一揭露橋接的本 beat 觀測（多來源 evidence 走同一 bridge）
                "reveal_updates": list(self._reveal_updates_this_beat),
                "evidence_events_this_beat": self._evidence_events_this_beat,
                "unmapped_evidence_this_beat": self._unmapped_evidence_this_beat,
                "escape_step": getattr(self, "_escape_step", "none"),
                # P0 #4：WorldProgress（current_area/known_areas/世界事實/investigation_state）
                "world_progress": self.world_progress(dp),
                "new_world_facts_this_beat": list(self._new_world_facts_this_beat),
                "ended": self.ended, "ending": self.ending}

    def _run_story_with_repair(self, player_decision, gs, ctx, on_event):
        """HD1：run_story 包進 QualityGate repair pipeline。回 (narrative, dp, quality_meta)。

        flag OFF / 無 contract → 純 run_story（行為不變）。
        """
        from core.constants import ENABLE_NARRATIVE_CONTROL

        # flag OFF / 無 contract：原樣串流（happy path、行為不變）
        if not (ENABLE_NARRATIVE_CONTROL and getattr(self, "_narrative_contract", None)):
            from core.narrative.quality_repair import StoryResult
            n, d = run_story(self.caller, self.bb, player_decision, gs.beat_number,
                             context_override=ctx, on_event=on_event,
                             system_override=self._story_system(ctx))
            return n, d, {"passed": True, "repaired": False, "fallback": False}

        # 敘事控制：pipeline 內**靜默生成**（on_event=None），避免把被否決的 beat 串給玩家；
        # 只有最終被接受的敘事才串流（修：品質閘門對串流輸出形同虛設）。
        def _runner(c):
            from core.narrative.quality_repair import StoryResult
            n, d = run_story(self.caller, self.bb, player_decision, gs.beat_number,
                             context_override=c, on_event=None,
                             system_override=self._story_system(c))
            opts = [getattr(o, "text", "") for o in (getattr(d, "suggested_options", None) or [])]
            return StoryResult(narrative=n, options=opts, payload=d)

        def _check(s, c):
            from core.narrative.quality_gate import check_beat
            if self._motif_tracker is not None:
                c["stagnant_motifs"] = self._motif_tracker.stagnant_motifs()
            return check_beat(s.narrative, s.options, c)

        def _fallback(c, q):
            return self._deterministic_story_result()

        try:
            from core.narrative.quality_repair import StoryRepairPipeline
            s = StoryRepairPipeline(_runner, _check, _fallback).run(ctx)
            if s.meta.get("quality_repaired") or s.meta.get("quality_fallback"):
                log.info("quality gate beat=%s repaired=%s fallback=%s", gs.beat_number,
                         s.meta.get("quality_repaired"), s.meta.get("quality_fallback"))
            self._replay_narrative(on_event, s.narrative)   # 串流最終被接受的敘事
            return s.narrative, s.payload, {
                "passed": not s.meta.get("quality_fallback"),
                "repaired": bool(s.meta.get("quality_repaired")),
                "fallback": bool(s.meta.get("quality_fallback"))}
        except Exception as e:                       # pipeline 失敗 → 退回純 run_story（含串流，不崩）
            log.warning("quality repair pipeline skipped: %s", e)
            n, d = run_story(self.caller, self.bb, player_decision, gs.beat_number,
                             context_override=ctx, on_event=on_event,
                             system_override=self._story_system(ctx))
            return n, d, {"passed": True, "repaired": False, "fallback": False}

    def _replay_narrative(self, on_event, text: str):
        """把最終被接受的敘事串流給前端（pipeline 內為靜默生成，故在此補送）。"""
        if on_event is None or not text:
            return
        try:
            from core.llm.parser import ParseEvent, NARRATIVE_CHUNK
            for para in str(text).split("\n\n"):     # 以段落為單位送，保留一點節奏
                if para.strip():
                    on_event(ParseEvent(NARRATIVE_CHUNK, para if para.endswith("\n") else para + "\n"))
        except Exception as e:
            log.warning("narrative replay skipped: %s", e)

    def _deterministic_story_result(self):
        """HD1：repair 仍失敗時的決定性 fallback——系統正確的方向 + 安全選項（不求文采）。"""
        from core.narrative.quality_repair import StoryResult, FALLBACK_NARRATIVE, FALLBACK_OPTIONS
        from core.models import DecisionPoint, Option
        dp = DecisionPoint(
            situation_recap=FALLBACK_NARRATIVE, decision_type="action",
            suggested_options=[Option(text=o, tone="cautious") for o in FALLBACK_OPTIONS],
            free_input_hint="或描述你想做的事", beat_meta={"beat_number": self.beat_number},
            is_narration_only=False)
        return StoryResult(narrative=FALLBACK_NARRATIVE, options=list(FALLBACK_OPTIONS), payload=dp)

    def _sanitize_decision_point(self, dp):
        """HD2：消毒 decision_point 的選項與 situation_recap（flag-gated）。"""
        if dp is None:
            return dp
        try:
            from core.constants import ENABLE_NARRATIVE_CONTROL
            if not ENABLE_NARRATIVE_CONTROL:
                return dp
            from core.narrative.sanitizer import SurfaceTextSanitizer, allowed_from_contract
            s = SurfaceTextSanitizer(
                allowed_terms=allowed_from_contract(getattr(self, "_narrative_contract", None)))
            new_opts = []
            for o in (getattr(dp, "suggested_options", None) or []):
                txt = s.sanitize(getattr(o, "text", ""))[0]
                new_opts.append(o.model_copy(update={"text": txt})
                                if hasattr(o, "model_copy") else o)
            recap = s.sanitize(getattr(dp, "situation_recap", "") or "")[0]
            return dp.model_copy(update={"suggested_options": new_opts, "situation_recap": recap})
        except Exception as e:
            log.warning("decision point sanitize skipped: %s", e)
            return dp

    def _sanitize_surface(self, text: str) -> str:
        """NR7：表層消毒（flag-gated）。flag OFF / 失敗 → 原文不動。"""
        if not text:
            return text
        try:
            from core.constants import ENABLE_NARRATIVE_CONTROL
            if not ENABLE_NARRATIVE_CONTROL:
                return text
            from core.narrative.sanitizer import SurfaceTextSanitizer, allowed_from_contract
            allowed = allowed_from_contract(getattr(self, "_narrative_contract", None))
            clean, leaks = SurfaceTextSanitizer(allowed_terms=allowed).sanitize(text)
            if leaks:
                log.info("surface sanitizer removed leaks: %s", leaks)
            return clean
        except Exception as e:                       # 消毒失敗不影響輸出
            log.warning("surface sanitize skipped: %s", e)
            return text

    def _motif_cooldown_pre(self, gs, ctx: dict):
        """NR5：換場景重置母題計數；把超用母題注入 context（story 須演化或換意象）。"""
        if self._motif_tracker is None:
            return
        try:
            scene = getattr(gs, "current_scene", None)
            if scene != self._last_scene:
                self._motif_tracker.reset_scene()
                self._last_scene = scene
            blocked = self._motif_tracker.build_blocked_motifs()
            if blocked:
                from core.narrative.motif_tracker import motif_block_instruction
                ctx["blocked_motifs"] = blocked
                ctx.setdefault("narrative_obligations", []).append(
                    motif_block_instruction(blocked))
        except Exception as e:
            log.warning("motif cooldown pre skipped: %s", e)

    def _motive_heartbeat_pre(self, ctx: dict):
        """NR6：逾期未提動機 → 加動機提醒義務（透過文件/NPC/道具/選項嵌入，不重複同句）。"""
        if self._motive_heartbeat is None or not self._motive_heartbeat.required():
            return
        try:
            nc = getattr(self, "_narrative_contract", None)
            goal = getattr(getattr(nc, "protagonist_motive", None), "immediate_goal", "") or "你來這裡的理由"
            loss = getattr(getattr(nc, "protagonist_motive", None), "personal_loss", "") or ""
            ctx["motive_heartbeat"] = True
            ctx.setdefault("narrative_obligations", []).append(
                f"提醒玩家為何而來（{loss}／{goal}），但**換一種方式**：透過一份文件、一個 NPC 的反應、"
                "一件道具或一個選項，而不是重複同一句旁白。")
            log.info("motive heartbeat due (beats_since=%s)", self._motive_heartbeat.beats_since_motive)
        except Exception as e:
            log.warning("motive heartbeat pre skipped: %s", e)

    def _motif_motive_post(self, narrative: str, ctx: dict):
        """NR5/NR6：登記本 beat 出現的母題、以及是否提及動機。"""
        try:
            if self._motif_tracker is not None:
                from core.narrative.motif_tracker import extract_motifs
                self._motif_tracker.register_beat(extract_motifs(narrative))
            if self._motive_heartbeat is not None:
                nc = getattr(self, "_narrative_contract", None)
                loss = getattr(getattr(nc, "protagonist_motive", None), "personal_loss", "") or ""
                # 動機物件/失蹤者名字出現 → 視為提及；或本 beat 已下達心跳義務
                referenced = bool(ctx.get("motive_heartbeat")) or (loss and any(
                    tok and tok in (narrative or "") for tok in loss.replace("，", " ").split()))
                self._motive_heartbeat.register_beat(bool(referenced))
        except Exception as e:
            log.warning("motif/motive post skipped: %s", e)

    def _apply_exit_intent(self, intent: str, gs, dp):
        """Player Sovereignty：依離開意圖調整本 beat（不自動收束）。

        ambiguous → 把 decision_point 換成 ExitOffer（保留 story 敘事、只換選項，永遠含「結束」選項）；
        temporary_retreat → 降即時危險、遊戲繼續；其餘 → dp 不變。
        """
        try:
            from core.narrative.exit_resolver import (
                AMBIGUOUS, TEMPORARY_RETREAT, SAFE_ZONE_REACHED, build_exit_offer_decision_point)
            if intent == AMBIGUOUS:
                log.info("exit intent ambiguous → ExitOffer（不自動結局，交還玩家）")
                return build_exit_offer_decision_point(dp)
            if intent in (TEMPORARY_RETREAT, SAFE_ZONE_REACHED):
                d = int(getattr(gs, "danger_level", 0) or 0)
                if d > 0:
                    gs.danger_level = max(0, d - 1)
                    bridge.sync_to_blackboard(gs, self.bb)
                    log.info("temporary retreat → danger %d→%d（續行，不結局）", d, gs.danger_level)
                # WorldModel：撤到外面 → current_area 切到結構性安全區（不結局）
                if intent == SAFE_ZONE_REACHED and self._world is not None:
                    self._world.withdraw_to_safe_zone()
                    self.bb.game_meta = {**self.bb.game_meta, "world_model": self._world.to_dict()}
        except Exception as e:
            log.warning("apply exit intent skipped: %s", e)
        return dp

    def _trigger_player_ending(self, player_decision: str, gs):
        """玩家明確「結束本次調查」→ 結算結局（escape；品質由 reveal 帳本決定 clean/ambiguous）。"""
        if self.ended:                               # 防禦：warden 硬結局已先觸發 → 不覆寫
            return
        self.ended = True
        self.ending = {"type": "escape", "soft": False, "via": "player_exit"}
        try:
            from core.narrative.ending_gate import gate_escape_quality
            quality, reason = gate_escape_quality(
                self.bb, self._game_state, player_decision, gs.beat_number)
            self.ending["escape_quality"] = quality
            self.ending["gate_reason"] = reason
            log.info("player ended run (escape, %s)", quality)
        except Exception as e:
            log.warning("ending gate skipped: %s", e)
        self._emit(EVT_ENDING_TRIGGERED, "escape")

    def _answer_debt_tick(self, player_decision: str, ctx: dict):
        """NR2：分類玩家提問→記債；債≥2 加償還義務進 story context；債存 game_meta 供 npc-chat。"""
        if self._answer_debt is None:
            return
        try:
            from core.narrative.answer_debt import classify_question, payoff_obligation
            topic = classify_question(player_decision)
            if topic:
                lvl = self._answer_debt.register_question(topic)
                ctx["answer_debt"] = {"topic": topic, "level": lvl}
                if self._answer_debt.requires_payoff(topic):
                    ctx.setdefault("narrative_obligations", []).append(
                        payoff_obligation(topic, lvl))
                    log.info("answer debt due: %s level=%s", topic, lvl)
                # 規則版：假定 story 已依義務償還 → 重置該 topic（避免永久嘮叨）
                self._answer_debt.register_answer(topic, "partial")
            self.bb.game_meta = {**self.bb.game_meta,
                                 "answer_debt": self._answer_debt.to_dict()}
        except Exception as e:                       # 答債失敗不影響推進
            log.warning("answer debt tick skipped: %s", e)

    def _revelation_tick(self, gs) -> bool:
        """NR0：本 beat 新發現的 kernel 線索（自帶 truth_id）→ bridge → 寫進 revealed_bible。

        回傳本 beat 是否有 reveal 升級（供 HB1 判斷是否需要保底 story evidence）。
        """
        try:
            from core.narrative.revelation import (
                RevelationBridge, evidence_from_clue_values, write_ledger_to_revealed_bible)
            clues = getattr(gs, "clues", {}) or {}
            new_ids = [cid for cid in clues if cid not in self._known_clue_ids]
            if not new_ids:
                return False
            self._known_clue_ids.update(new_ids)
            items = [(cid, clues.get(cid)) for cid in new_ids]
            events, unmapped = evidence_from_clue_values(
                items, source="kernel",
                beat=getattr(gs, "beat_number", self.beat_number),
                scene_id=getattr(gs, "current_scene", None))
            updates = RevelationBridge().apply(self._reveal_ledger, events)
            self._reveal_updates_this_beat += updates
            self._evidence_events_this_beat += len(events)
            self._unmapped_evidence_this_beat += len(unmapped)
            if updates:
                log.info("reveal updates (kernel): %s", updates)
            if unmapped:
                log.info("unmapped clues (no truth_id on clue): %s", unmapped)
            write_ledger_to_revealed_bible(self.bb, self._reveal_ledger)
            self._update_npc_truth_refs()            # 帳本變動 → 更新 NPC 可引用白名單
            return bool(updates)
        except Exception as e:                       # 橋接失敗不影響推進（B8）
            log.warning("revelation tick skipped: %s", e)
            return False

    def _update_npc_truth_refs(self):
        """P-plumbing：把 NPC 此刻可安全引用的 truth 白名單寫進 game_meta（供 npc-chat context + gate）。"""
        if self._reveal_ledger is None:
            return
        try:
            from core.narrative.npc_chat_control import build_allowed_truth_refs
            refs = build_allowed_truth_refs(
                self._reveal_ledger, getattr(self, "_reveal_level", "hinted"),
                core_truth_ids=getattr(self, "_core_truth_ids", set()))
            self.bb.game_meta = {**self.bb.game_meta, "npc_allowed_truth_refs": refs}
        except Exception as e:
            log.warning("npc truth refs update skipped: %s", e)

    def _seed_world_area(self):
        """開局把起始 kernel 場景登記為 area 並設 current（WorldModel 是 current_area 權威）。"""
        if self._world is None:
            return
        scene = getattr(self._game_state, "current_scene", None)
        if scene:
            self._world.set_current_area(scene, label=self._area_label(scene))
            self._world_kernel_scene = scene
            self.bb.game_meta = {**self.bb.game_meta, "world_model": self._world.to_dict()}

    def _area_label(self, scene_id) -> str:
        """從 scene_registry 取場景顯示名（無則用 id）。"""
        try:
            sc = (self.bb.snapshot().get("scene_registry") or {})
            for l in sc.get("known_locations", []) if isinstance(sc, dict) else []:
                if isinstance(l, dict) and l.get("id") == scene_id and l.get("name"):
                    return l["name"]
        except Exception:
            pass
        return str(scene_id)

    def _world_model_tick(self, action: str, narrative: str, dp=None):
        """WorldModel：current_area 的**唯一權威**。kernel 只在「真正移動」時透過 WorldDelta 改區域；
        **不得用 scene sync 覆蓋** WorldModel 已套用的 current_area（如撤退到 outside_dock）。

        物件/NPC/事實登記**優先走 story 結構化 entity_delta**（object/actor/fact，每 beat ≤3，
        malformed 丟棄）；story 沒給結構化 delta 時，才退回 EntityExtractor 掃敘事（fallback）。
        本 beat 被新增/變更狀態的實體記進 `_changed_entities_this_beat`（供 observation 投影）。
        """
        if self._world is None:
            return
        try:
            from core.world.extractor import extract_entities
            from core.world.model import WorldDelta, OBJECT, coerce_entity_deltas
            gs = self._game_state
            scene = getattr(gs, "current_scene", None)
            before = {eid: e.state for eid, e in self._world.entities.items()}
            # 只有 kernel **真的移動了**（場景 ≠ 上次 WorldModel 同步的 kernel 場景）才改 current_area。
            # 若玩家原地觀察 / 否定移動 / 已撤退到 outside_dock，kernel 場景不變 → WorldModel 不被覆蓋。
            if scene and scene != self._world_kernel_scene:
                self._world.apply(WorldDelta(op="move_current", entity_id=scene, origin="kernel"))
                # 補 label（apply 用 id 當 label 登記時補回顯示名）
                _e = self._world.get(scene)
                if _e is not None and _e.label == scene:
                    _e.label = self._area_label(scene)
                self._world_kernel_scene = scene
            # 優先：story 結構化 entity_delta（object/actor/fact，kind-guarded）。
            story_deltas = coerce_entity_deltas(getattr(dp, "entity_delta", None))
            if story_deltas:
                self._world.apply_story_deltas(story_deltas)
            else:
                # fallback：EntityExtractor 掃敘事（story 沒給結構化 delta 時才用）。
                self._world.apply_deltas(extract_entities(narrative, npc_names=self.known_npcs))
            # 玩家檢查物件 → 標 inspected（世界記得他查過）
            if any(v in (action or "") for v in ("檢查", "查看", "檢視", "端詳", "研究", "翻看")):
                for o in self._world.by_kind(OBJECT):
                    if any(tok and tok in action for tok in (o.label or "").split()):
                        self._world.inspect(o.id)
            # 算本 beat 被新增/變更狀態的實體（含 register 新增、set_state、inspect）
            after = {eid: e.state for eid, e in self._world.entities.items()}
            self._changed_entities_this_beat = [
                eid for eid, st in after.items() if before.get(eid) != st]
            self.bb.game_meta = {**self.bb.game_meta, "world_model": self._world.to_dict()}
        except Exception as e:
            log.warning("world model tick skipped: %s", e)

    def _world_facts_tick(self, text: str, *, source: str = "story") -> list:
        """P0：從文字抽 world_state_fact 寫進 game_meta（即使沒 truth reveal 也留可檢查後果）。"""
        try:
            from core.narrative.world_facts import extract_world_facts, add_world_facts
            facts = extract_world_facts(text, source=source, npc_names=self.known_npcs)
            added = add_world_facts(self.bb, facts, beat=self.beat_number)
            if added:
                self._new_world_facts_this_beat += added
                log.info("world facts added (%s): %s", source, added)
            return added
        except Exception as e:
            log.warning("world facts tick skipped: %s", e)
            return []

    def world_progress(self, dp=None) -> dict:
        """P0 #4：WorldProgress 觀測。**current_area / known_areas / exits / affordances 由 WorldModel 投影**
        （WorldModel 是這些的唯一權威；不從 kernel scene 直取）。"""
        from core.world.model import AREA, EXIT
        w = getattr(self, "_world", None)
        gs = getattr(self, "_game_state", None)
        # current_area / known_areas / exits ← WorldModel（無 world 時退回 kernel scene）
        if w is not None:
            current_area = w.current_area
            known = [e.id for e in w.by_kind(AREA)]
            exits = [{"id": e.id, "label": e.label, "state": e.state} for e in w.by_kind(EXIT)]
        else:
            current_area = getattr(gs, "current_scene", None) if gs is not None else None
            known, exits = ([current_area] if current_area else []), []
        from core.narrative.world_facts import get_world_facts
        all_facts = get_world_facts(self.bb)
        new_facts = getattr(self, "_new_world_facts_this_beat", []) or []
        changed_exits = [k for k in new_facts
                         if (all_facts.get(k) or {}).get("category") == "exit"]
        intent = getattr(self, "_exit_intent", "none")
        from core.narrative.exit_resolver import WITHDRAW_STATES
        inv_state = "paused" if intent in WITHDRAW_STATES else "active"
        avail = [getattr(o, "text", "") for o in (getattr(dp, "suggested_options", None) or [])]
        return {
            "current_area": current_area,
            "known_areas": known,
            "exits": exits,
            "world_facts": sorted(all_facts.keys()),
            "new_world_facts_this_beat": list(new_facts),
            "changed_exits_this_beat": changed_exits,
            "investigation_state": inv_state,
            "available_next": avail,
            "changed_entities_this_beat": list(
                getattr(self, "_changed_entities_this_beat", []) or []),
            "world_model": self._world_model_projection(),
        }

    def _world_model_projection(self) -> dict:
        """把 WorldModel 投影成觀測（AI 可看到世界記得哪些實體、此處有什麼可互動、本 beat 改了什麼）。"""
        w = getattr(self, "_world", None)
        if w is None:
            return {}
        try:
            def _e(e):
                return {"id": e.id, "kind": e.kind, "label": e.label, "state": e.state}
            ents = [_e(e) for e in w.entities.values()]
            affs = [{"verb": a.verb, "entity_id": a.entity_id, "label": a.label}
                    for a in w.affordances_here()]
            by_kind = {}
            for e in w.entities.values():
                by_kind[e.kind] = by_kind.get(e.kind, 0) + 1
            changed = list(getattr(self, "_changed_entities_this_beat", []) or [])
            return {"current_area": w.current_area, "counts": by_kind,
                    "entities": ents[:40], "affordances_here": affs[:20],
                    # story structured entity_delta：此處的實體 / 可互動物 / 本 beat 變更
                    "entities_here": [_e(e) for e in w.entities_here()][:40],
                    "interactables_here": [_e(e) for e in w.interactables_here()][:20],
                    "changed_entities_this_beat": changed}
        except Exception:
            return {}

    def _story_evidence_tick(self, action: str, narrative: str, reveal_changed: bool):
        """HB1：玩家做了有意義調查但本 beat reveal 無變化 → 保底產 hinted evidence 走 bridge。"""
        if self._reveal_ledger is None:
            return
        try:
            from core.narrative.evidence_extractor import StoryEvidenceExtractor
            from core.narrative.revelation import RevelationBridge, write_ledger_to_revealed_bible
            ext = StoryEvidenceExtractor(getattr(self, "_truth_index", {}) or {})
            events = ext.extract(beat=self.beat_number, action=action,
                                 narrative=narrative, reveal_changed=reveal_changed)
            if not events:
                return
            self._evidence_events_this_beat += len(events)
            mapped = [e for e in events if e.truth_id]
            self._unmapped_evidence_this_beat += len(events) - len(mapped)
            if mapped:
                updates = RevelationBridge().apply(self._reveal_ledger, mapped)
                self._reveal_updates_this_beat += updates
                if updates:
                    log.info("reveal updates (story fallback): %s", updates)
                # 即使未跨等級，累積的 strength 也要持久化（否則存讀檔後歸零）
                write_ledger_to_revealed_bible(self.bb, self._reveal_ledger)
            else:
                log.info("unmapped story evidence this beat: investigation w/o truth mapping")
        except Exception as e:
            log.warning("story evidence tick skipped: %s", e)

    def bridge_npc_evidence(self, reply_or_resp, *, answer_status: str | None = None,
                            allow_keyword_fallback: bool = False, npc_id: str | None = None):
        """HC1/#10：把 NPC 對話的**結構化** evidence_events 走同一條 bridge 進 revealed_bible。

        正規路徑：`NPCChatResponse.evidence_events` → cap_to_ceiling → response_to_evidence
        → 過濾「只准真實 revelation_pool 真相」（NPC 不得無中生有 truth）→ RevelationBridge.apply。
        keyword scan（evidence_from_npc_reply）只當 **debug warning / legacy fallback**，
        預設**不 grant**（`allow_keyword_fallback=True` 才作為過渡回退）。
        flag OFF / 無帳本 / 純文字無結構化 evidence → 不 grant。

        另：NPC 結構化 `entity_delta`（只 fact/actor）走**獨立**通道進 WorldModel，
        **不**經 reveal ledger（NPC fact 不得自動 grant 真相，見 _bridge_npc_entity_delta）。
        """
        try:
            from core.constants import ENABLE_NARRATIVE_CONTROL
            led = getattr(self, "_reveal_ledger", None)
            if not ENABLE_NARRATIVE_CONTROL or led is None:
                return []
            # P0 WorldStateFact：NPC 的有用資訊（出口鎖死/機房/發電機/某人來過）→ world_state_fact
            self._world_facts_tick(getattr(reply_or_resp, "visible_reply", reply_or_resp),
                                   source="npc_chat")
            # NPC structured entity_delta → WorldModel（fact/actor only；不 grant reveal）
            self._bridge_npc_entity_delta(reply_or_resp, npc_id)
            from core.narrative.revelation import RevelationBridge, write_ledger_to_revealed_bible
            allowed_block = (self.bb.game_meta or {}).get("npc_allowed_truth_refs") or {}
            allowed_refs = allowed_block.get("allowed_truth_refs", [])

            # ── 正規路徑：結構化 evidence_events → 過白名單 gate → bridge ──
            resp = reply_or_resp if hasattr(reply_or_resp, "evidence_events") else None
            accepted, rejected = [], []
            if resp is not None and getattr(resp, "evidence_events", None):
                from core.narrative.npc_chat_control import (
                    response_to_evidence, validate_npc_evidence)
                proposed = response_to_evidence(resp, beat=self.beat_number)
                # gate：只准白名單 truth_id（含 core/超 ceiling 排除）；level 夾到 ref 上限
                accepted, rejected = validate_npc_evidence(proposed, allowed_refs)
                accepted = [e for e in accepted if e.truth_id in led.truths]   # 雙重保險：真實碎片
            else:
                proposed = []

            # conversation note：無 truth_id 的提示寫成對話筆記（不污染 truth ledger）
            notes = [getattr(e, "surface_text", "") for e, why in rejected if why == "no_truth_id"]
            if notes:
                self._append_conversation_notes(notes)

            self._npc_chat_debug = {
                "allowed_truth_refs": [r.get("truth_id") for r in allowed_refs],
                "evidence_events_proposed": len(proposed),
                "evidence_events_accepted": len(accepted),
                "evidence_events_rejected": len(rejected),
                "rejection_reasons": sorted({why for _, why in rejected}),
            }

            if accepted:
                updates = RevelationBridge().apply(led, accepted)
                if updates:
                    log.info("reveal updates (npc structured, gated): %s", updates)
                write_ledger_to_revealed_bible(self.bb, led)     # 累積 strength 也持久化
                self._update_npc_truth_refs()
                return updates

            # ── keyword scan：僅 debug warning / 選擇性 legacy fallback（預設不 grant）──
            reply = getattr(reply_or_resp, "visible_reply", reply_or_resp)
            status = answer_status or getattr(reply_or_resp, "answer_status", None)
            from core.narrative.revelation import evidence_from_npc_reply
            kw = [e for e in evidence_from_npc_reply(
                      reply, getattr(self, "_truth_index", {}) or {},
                      beat=self.beat_number, answer_status=status)
                  if e.truth_id in led.truths]
            if kw:
                log.info("npc keyword-scan evidence detected (debug only, NOT granted unless "
                         "allow_keyword_fallback): %s", [e.truth_id for e in kw])
                if allow_keyword_fallback:
                    updates = RevelationBridge().apply(led, kw)
                    write_ledger_to_revealed_bible(self.bb, led)
                    return updates
            return []
        except Exception as e:
            log.warning("npc evidence bridge skipped: %s", e)
            return []

    def _bridge_npc_entity_delta(self, resp, npc_id: str | None) -> list:
        """NPC 結構化 entity_delta → WorldModel（**只 fact/actor**；**不碰 reveal ledger**）。

        NPC 不得新增 object/area/exit（coerce 過濾 + apply kind-guard 雙重擋）；fact 帶
        source=npc_id / confidence=npc_claim / origin=npc。malformed 丟棄、不污染。
        **關鍵**：這條通道完全獨立於 evidence/ledger——NPC 講的事只成「世界裡可被指涉的主張」，
        不會自動把任何真相推進到 confirmed（決定性證明仍須玩家親自發現）。回新增/變更的 entity id。
        """
        if self._world is None:
            return []
        try:
            from core.world.model import coerce_npc_entity_deltas, NPC_ENTITY_KINDS
            raw = getattr(resp, "entity_delta", None)
            deltas = coerce_npc_entity_deltas(raw, npc_id=str(npc_id or "npc"))
            if not deltas:
                return []
            changed = self._world.apply_story_deltas(deltas, allowed_kinds=NPC_ENTITY_KINDS)
            if changed:
                # 併進本 beat 變更集（observation.world_progress 投影 new fact/actor entities）
                self._changed_entities_this_beat = list(dict.fromkeys(
                    list(getattr(self, "_changed_entities_this_beat", []) or []) + changed))
                self.bb.game_meta = {**self.bb.game_meta, "world_model": self._world.to_dict()}
                log.info("npc entity_delta → WorldModel (no reveal grant): %s", changed)
            return changed
        except Exception as e:                           # 失敗不影響對話（B8）
            log.warning("npc entity delta bridge skipped: %s", e)
            return []

    def _append_conversation_notes(self, notes: list):
        """把 NPC 無 truth_id 的提示寫成對話筆記（不進 truth ledger，避免污染）。"""
        try:
            cur = list((self.bb.game_meta or {}).get("npc_conversation_notes") or [])
            cur += [n for n in notes if n]
            self.bb.game_meta = {**self.bb.game_meta, "npc_conversation_notes": cur[-20:]}
        except Exception:
            pass

    def note_npc_answer(self, player_question: str, answer_status: str | None):
        """NPC 回應後更新答債：partial/answered 償還；evasion/none 不償還（不重置答債）。"""
        if self._answer_debt is None or not answer_status:
            return
        try:
            from core.narrative.answer_debt import classify_question
            topic = classify_question(player_question)
            if topic:
                self._answer_debt.register_answer(topic, answer_status)
                self.bb.game_meta = {**self.bb.game_meta,
                                     "answer_debt": self._answer_debt.to_dict()}
        except Exception as e:
            log.warning("note npc answer skipped: %s", e)

    def _log_beat(self, progress):
        gs = self._game_state
        log.info("beat use_kernel=true scene=%s event=%s delta=%s obligations=%s forbidden=%s",
                 gs.current_scene, progress.committed_event, progress.patch.progress_delta,
                 progress.patch.narrative_obligations, sorted(gs.forbidden_repeats))

    # ── Legacy 流程（原 LLM 自由流程，原封）─────────────────────────────
    def _step_legacy(self, player_decision, input_path, on_event, on_progress) -> dict:
        _p = on_progress or (lambda *_: None)
        _p("warden")
        verdict = run_warden(player_decision, self.bb, caller=self.caller)
        self._record_skill_claim(verdict)
        if verdict.rule_violation:
            self._emit(EVT_RULE_VIOLATION, verdict)
        if verdict.ending_triggered:
            self._emit(EVT_ENDING_TRIGGERED, verdict)

        events = event_extract(player_decision, self.last_story,
                               known_npcs=self.known_npcs,
                               known_locations=self.known_locations,
                               known_items=self.known_items)
        touched = [e.get("target") or e.get("npc") or e.get("item") or e.get("location") for e in events]
        touched = [t for t in touched if t]
        reached = [e["location"] for e in events if e.get("type") == "reached_location"]

        _p("orchestrator")
        newly = run_orchestrator(self.bb, self.beat_number,
                                 touched_ids=touched, reached_locations=reached, caller=self.caller)
        _p("story")
        narrative, dp = run_story(self.caller, self.bb, player_decision, self.beat_number,
                                  directive=verdict.directive_to_story, newly_revealed=newly,
                                  on_event=on_event, system_override=self._story_system(None))
        self.last_story = narrative
        self._safe_point(narrative, dp)
        log.info("beat use_kernel=false (legacy)")

        if verdict.ending_triggered or verdict.rule_violation:
            self.ended = True
            self.ending = {"type": verdict.ending_triggered or "death_physical",
                           "soft": bool(verdict.ending_is_soft), "via": "warden"}
        self._finalize_ending()
        dp = self._enforce_ended_invariant(dp)       # HA1：ended ⇒ 無 options
        return {"narrative": narrative, "decision_point": dp, "warden": verdict,
                "ended": self.ended, "ending": self.ending}

    # ── 內部 ────────────────────────────────────────────────────────────
    def _emit(self, evt, *args, **kwargs):
        if self.bus is not None:
            self.bus.publish(evt, *args, **kwargs)

    def _enforce_ended_invariant(self, dp):
        """HA1：結局不變式——ended=true ⇒ decision_point 不再帶 options / free_input_hint。

        regression 斷言 `not (ended and options)`。輸出層（agent_play/webview）也應再保險一次。
        """
        if not self.ended or dp is None:
            return dp
        try:
            # free_input_hint 型別為 str（非 Optional）→ 用 "" 而非 None，避免 schema/consumer 破裂
            return dp.model_copy(update={"suggested_options": [], "free_input_hint": ""})
        except Exception:                            # 非 pydantic / 失敗也要安全
            try:
                dp.suggested_options = []
            except Exception:
                pass
            return dp

    def _finalize_ending(self):
        """UB2：結局觸發時補上純敘述收尾 + 完整真相揭露 + 復盤（只做一次）。"""
        if self.ended and self.ending and not self.ending.get("is_ending"):
            try:
                from core.agents.ending import build_ending_sequence
                # NR0：把揭露帳本的部分進度帶進復盤（hinted/observed/suspected，不再只算 confirmed）
                reveal_recap = None
                _led = getattr(self, "_reveal_ledger", None)
                if _led is not None:
                    from core.narrative.revelation import recap_from_ledger
                    reveal_recap = recap_from_ledger(_led)
                self.ending = build_ending_sequence(self.bb, self.ending, list(self.bb.ledger),
                                                    reveal_recap=reveal_recap)
            except Exception as e:                   # 結局組裝失敗也要能收尾（保底 ending dict）
                log.warning("ending sequence build failed: %s", e)

    def _record_skill_claim(self, verdict):
        """UB1：玩家破格技能被封頂 → 發 SKILL_CLAIMED + 寫 (技能,侷限) 進 ledger。"""
        if not getattr(verdict, "skill_claim", None):
            return
        self._emit(EVT_SKILL_CLAIMED, verdict)
        try:
            self.bb.ledger.append({
                "type": "skill_limit",
                "content": f"宣稱「{verdict.skill_claim}」→ 侷限：{verdict.skill_limitation or '(未指定)'}",
            })
        except Exception as e:                       # 寫 ledger 失敗不影響推進
            log.warning("skill ledger write failed: %s", e)

    def _derive_known(self):
        snap = self.bb.snapshot()
        self.known_npcs = [n.get("name") for n in (snap.get("npc_registry") or [])
                           if isinstance(n, dict) and n.get("name")]
        sc = snap.get("scene_registry") or {}
        locs = sc.get("known_locations", []) if isinstance(sc, dict) else []
        names = [l.get("name") for l in locs if isinstance(l, dict) and l.get("name")]
        ids = [l.get("id") for l in locs if isinstance(l, dict) and l.get("id")]
        self.known_locations = list(dict.fromkeys(names + ids))

    def _safe_point(self, narrative: str, dp: Any):
        """beat 完成：落 SQLite 快照 → 修剪視窗 → compactor → merge patch → 發信號。"""
        self.beat_number += 1
        self.bb.beat_number = self.beat_number

        snap = self.bb.snapshot()
        self.db.save_beat(
            self.run_id, self.beat_number, narrative,
            decision_json=dp.model_dump_json(),
            rolling_summary_snapshot=str(snap.get("rolling_summary", "")),
            blackboard_snapshot_json=json.dumps(snap, ensure_ascii=False, default=str),
            is_narration_only=dp.is_narration_only,
        )
        # UB5：道具庫隨 beat 快照保存（kernel 模式存完整欄位含 held_by/is_key_item）
        try:
            from dataclasses import asdict as _asdict
            if self._game_state is not None:
                inv = [_asdict(i) for i in self._game_state.inventory.values()]
            else:
                inv = (snap.get("shared_inventory") or {}).get("items", [])
            self.db.save_inventory_snapshot(
                self.run_id, self.beat_number, json.dumps(inv, ensure_ascii=False, default=str))
        except Exception as e:  # pragma: no cover - 快照失敗不影響推進
            log.warning("inventory snapshot skipped: %s", e)

        bw = self.bb.beat_window
        if len(bw) > BEAT_WINDOW_SIZE:
            evicted = bw[:-BEAT_WINDOW_SIZE]
            del bw[:-BEAT_WINDOW_SIZE]
            usage = min(1.0, estimate_tokens(str(snap.get("rolling_summary", ""))) / SUMMARY_TOKEN_CAP)
            out = self.compactor.compact(evicted, str(snap.get("rolling_summary", "")),
                                         list(self.bb.ledger), usage)
            self.compactor.apply_to_blackboard(self.bb, out)
            self.bb.merge_and_bump()
            if usage >= CONTEXT_THRESHOLD_L1:
                self._emit(EVT_CONTEXT_THRESHOLD, usage)

        # UB4：輕量 dreaming（每 5 beat，在場 NPC，非同步只產 patch，安全點 merge）
        try:
            from core.agents.dreaming import run_dreaming
            dreamt = run_dreaming(self.caller, self.bb, self.beat_number)
            if dreamt:
                self.bb.merge_and_bump()             # 安全點才讓 evolving patch 生效
                self._emit(EVT_NPC_EVOLVED, [n for n, _ in dreamt])
        except Exception as e:                       # dreaming 掛掉不影響遊戲續行（B8/F8）
            log.warning("dreaming skipped: %s", e)

        self._emit(EVT_BEAT_COMPLETED, self.beat_number)
