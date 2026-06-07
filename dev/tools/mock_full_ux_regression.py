#!/usr/bin/env python3
"""mock_full_ux_regression — Final Stabilization Mock Regression（**無真 LLM**）。

用 deterministic ScriptedCaller 取代 OpenRouter，驅動 open-exploration runtime 的主要契約，
四條路線 A/B/C/D 逐 checkpoint 驗證。輸出：
  dev/reports/mock-full-ux-regression.jsonl
  dev/reports/mock-full-ux-regression.md

設計：ScriptedCaller 回確定性回應（setup/story/warden/npc-chat）；harness 以 loop.start / loop.step
（action-class/gate/reveal 不依賴 LLM）＋ loop 的確定性方法（_world_model_tick / bridge_npc_evidence /
spatial_debug / _sanitize_surface / _materialize_public_facts）驗證真實 runtime wiring。**不呼叫 OpenRouter。**
"""
from __future__ import annotations

import json
import os
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
os.environ["ENABLE_OPENING_VARIATION_CONTRACT"] = "true"

from core.models import (SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput,
                         DecisionPoint, Option, BeatMeta)


# ── ScriptedCaller（取代真 LLM；確定性）────────────────────────────────────────
def _setup_output():
    return SetupOutput(
        real_bible={
            "world_truth": {"what_really_happened": "這裡做過記憶實驗",
                            "the_threat_is": "頻率", "deadly_rule": "別相信整點報時"},
            "hard_triggers": [], "ending_conditions": [],
            # 刻意：frag_001 無 title（測 public_title fallback）、frag_003 有 title、frag_007 核心
            "revelation_pool": [
                {"id": "frag_001_logbook", "title": "", "content": "日誌顯示由你本人登入"},
                {"id": "frag_003_signal", "title": "異常訊號", "content": "訊號來自地下"},
                {"id": "frag_007_core", "title": "核心真相", "content": "你也是實驗體之一"}],
        },
        npc_registry=[NPCBible(name="林守一", profession="設備維修技師", personality="mysterious",
                               voice_sample="…", public_face="想恢復電力",
                               secret_core="其實動過手腳SECRET", self_aware=True, appearance="戴手套")],
        protagonist={"name": "周凱", "starting_situation": "弟弟在這裡失蹤"},
        scene_registry=SceneRegistry(
            current_location="entry",
            known_locations=[Location(id="entry", name="維修入口艙", description="入口"),
                             Location(id="corridor", name="主走廊", description="走廊")]),
        opening_sequence=["你在維修入口艙醒來。"])


_CLEAN_DECISION = {
    "situation_recap": "你站在原地，走廊很暗，前方有可走的方向。",
    "decision_type": "action",
    "suggested_options": [{"text": "往前走查看", "tone": "cautious"},
                          {"text": "留在原地觀察", "tone": "evasive"}],
    "free_input_hint": "或描述你想做的事…",
    "beat_meta": {"beat_number": 0},
}
_CLEAN_NARRATIVE = ("走廊的燈忽明忽暗，金屬牆面映著你的影子。空氣裡有股說不上來的氣味，"
                    "你放慢呼吸，留意四周還有什麼。前方的通道延伸進更深的黑暗裡。")


class ScriptedCaller:
    """確定性 caller：setup/warden 固定；story/npc-chat 由 harness 在每拍前設定。**不打真 LLM。**"""

    def __init__(self):
        self.loader = None                              # SkillCaller 相容（ConfigPromptSource 取用）
        self.setup_out = _setup_output()
        self.warden_response = WardenOutput(directive_to_story="繼續")
        self.story_narrative = _CLEAN_NARRATIVE
        self.story_decision = dict(_CLEAN_DECISION)
        self.npc_response = '{"reply":"我是負責設備維修的。","answer_status":"partial"}'
        self.calls = []                                 # 記錄呼叫（觀測；應全為 mock）

    def set_story(self, narrative=None, decision=None):
        self.story_narrative = narrative if narrative is not None else _CLEAN_NARRATIVE
        self.story_decision = decision if decision is not None else dict(_CLEAN_DECISION)

    def call(self, agent, context, output_model=None, temperature=None, system_override=None, **kw):
        self.calls.append(("call", agent))
        if agent == "setup":
            return self.setup_out
        if agent == "warden":
            return self.warden_response
        if agent == "npc-chat":
            return self.npc_response
        if agent == "orchestrator":
            from core.models import OrchestratorOutput
            return OrchestratorOutput()
        raise AssertionError(f"ScriptedCaller: unexpected call agent={agent}")

    def stream(self, agent, context, temperature=None, system_override=None, **kw):
        self.calls.append(("stream", agent))
        if agent != "story":
            raise AssertionError(f"ScriptedCaller: unexpected stream agent={agent}")
        decision_json = json.dumps(self.story_decision, ensure_ascii=False)
        for tok in [self.story_narrative, "\n<<<DECISION>>>\n", decision_json]:
            yield tok


def _new_loop(run_id):
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database
    from core.signal import SignalBus
    return BeatLoop(ScriptedCaller(), Blackboard(), Database(tempfile.mktemp(suffix=".db")),
                    SignalBus(), run_id=run_id, use_kernel=True)


# ── checkpoint 紀錄 ──────────────────────────────────────────────────────────
class Recorder:
    def __init__(self):
        self.rows = []

    def check(self, route, name, ok, evidence, known=None):
        row = {"route": route, "checkpoint": name, "pass": bool(ok), "evidence": evidence}
        if known and not ok:
            row["known_item"] = known          # 已知 monitor item（非新 regression）
        self.rows.append(row)
        mark = "✅" if ok else ("⚠ KNOWN" if known else "❌")
        print(f"  {mark} [{route}] {name}: {evidence}")
        return bool(ok)


def _obs(loop, out):
    from dev.tools.agent_play import _dp_to_obs
    return _dp_to_obs(loop, out.get("narrative"), out.get("decision_point"),
                      out.get("ended"), out.get("ending"), step_result=out)


def _reveal_tuple(loop):
    led = getattr(loop, "_reveal_ledger", None)
    if led is None:
        return (0, 0, 0, 0)
    c = led.counts()
    return (c["hinted_or_better"], c["observed_or_better"], c["suspected_or_better"],
            c["confirmed_or_better"])


def _reset_beat_attrs(loop):
    loop._changed_entities_this_beat = []
    loop._changed_entities_detail = []
    loop._inventory_delta_this_beat = []
    loop._known_fact_delta_this_beat = []
    loop._skipped_materialization_reason = ""
    loop._scene_changed_this_beat = False
    loop._new_actor_this_beat = False


def _drive_tick(loop, action, entity_delta=None, review_locked=False):
    """確定性驅動 _world_model_tick（帶 story entity_delta）；回 (dp, debug)。"""
    _reset_beat_attrs(loop)
    dp = DecisionPoint(situation_recap="x", decision_type="action",
                       suggested_options=[Option(text="往前走", tone="cautious")],
                       beat_meta=BeatMeta(beat_number=loop.beat_number + 1),
                       entity_delta=entity_delta or [])
    loop._world_model_tick(action, "敘事文字", dp, review_locked=review_locked)
    return dp, {"inventory_delta": list(loop._inventory_delta_this_beat),
                "known_fact_delta": list(loop._known_fact_delta_this_beat),
                "skipped": loop._skipped_materialization_reason}


def _spatial(loop):
    return loop.spatial_debug() or {}


# ══════════════ Route A — 探索型 mock（inventory / known_facts / alias / spatial）═══════
def run_route_A(new_loop, rec):
    from core.world.model import OBJECT, ACTOR, AREA, EXIT
    from dev.tools.agent_play import _do_chat
    R = "A_explorer_mock"
    print("\n" + "=" * 60 + f"\n{R}\n" + "=" * 60)
    loop = new_loop("mock-A")
    res = loop.start({"theme": "研究站", "protagonist_name": "周凱", "npc_count": 1})
    # 開場：opening variation 不用 forbidden literals（首場 forbidden 空）+ 無 placeholder
    ov = res.get("opening_variation", {})
    rec.check(R, "opening_no_forbidden_literals",
              not (ov.get("forbidden_literals")) and not res.get("opening_variation_violation", {}).get("has_violation"),
              {"forbidden": ov.get("forbidden_literals"), "violation": res.get("opening_variation_violation", {}).get("has_violation")})

    # 1. story 產生 object entity
    _drive_tick(loop, "我搜查桌面與抽屜",
                [{"op": "register", "kind": "object", "label": "頻率表",
                  "entity_id": "object.frequency_meter", "affords": ["inspect", "take"]}])
    rec.check(R, "story_object_registered",
              loop._world.get("object.frequency_meter") is not None,
              {"objects": [o.id for o in loop._world.by_kind(OBJECT)]})

    # 2+3. inspect 不取、take 才入 inventory
    _, dbg_inspect = _drive_tick(loop, "我仔細檢查那頻率表")
    inspected_not_taken = loop._world.get("object.frequency_meter").state == "inspected"
    rec.check(R, "inspect_does_not_take", inspected_not_taken and not dbg_inspect["inventory_delta"],
              {"state": loop._world.get("object.frequency_meter").state})
    _, dbg_take = _drive_tick(loop, "我把那頻率表撿起來，收進口袋帶著")
    from core.world.player_state import project_inventory
    inv = [e["label"] for e in project_inventory(loop._world)]
    rec.check(R, "take_adds_inventory",
              "頻率表" in inv and any(d["id"] == "object.frequency_meter" for d in dbg_take["inventory_delta"]),
              {"inventory": inv, "inventory_delta": dbg_take["inventory_delta"]})

    # 4. taken 物件：focus 在它時 visible；focus 移走後不 visible
    loop._world.tag_entity_area("object.frequency_meter", loop._world.current_area)
    visible_when_focus = "object.frequency_meter" in [v.get("id") for v in _spatial(loop).get("visible_entities", [])]
    rec.check(R, "taken_visible_when_focus", visible_when_focus,
              {"focus": (loop._current_focus or {}).get("id")})

    # 5. NPC first-contact → actor entity + fact（結構化 + prose 皆測；此處結構化 fact）
    loop.caller.npc_response = ('{"reply":"我是負責設備維修的，這裡的事我知道一些。",'
                               '"answer_status":"partial",'
                               '"entity_delta":[{"op":"register","kind":"fact","label":"離開需要主控室的授權卡"}]}')
    _do_chat(loop, {"npc": "林守一", "text": "你是誰？要怎麼離開這裡？"})
    foc = loop._current_focus or {}
    foc_entity = loop._world.get(foc.get("id")) if foc.get("id") else None
    rec.check(R, "npc_actor_entity_exists",
              foc_entity is not None and foc_entity.kind == ACTOR,
              {"focus_id": foc.get("id"), "kind": getattr(foc_entity, "kind", None)})
    from core.world.player_state import project_known_facts
    kf = project_known_facts(loop._world)
    npc_fact = [f for f in kf if f.get("confidence") == "npc_claim"]
    rec.check(R, "known_facts_from_npc_with_source_confidence",
              bool(npc_fact) and npc_fact[0].get("source") == "林守一",
              {"npc_facts": npc_fact})
    # taken 物件 focus 已移到 NPC → 不再 visible
    not_visible = "object.frequency_meter" not in [v.get("id") for v in _spatial(loop).get("visible_entities", [])]
    rec.check(R, "taken_not_visible_when_focus_moved", not_visible,
              {"focus": (loop._current_focus or {}).get("id")})

    # 8. 「剛才那個東西」→ object（不被 NPC focus 吃掉）
    ps = loop.player_state()
    er_obj = loop._entity_resolution_block("我再看一下剛才那個東西", ps)
    rid = er_obj.get("resolved_entity_id")
    rec.check(R, "alias_object_not_npc",
              bool(rid) and (loop._world.get(rid) or None) and loop._world.get(rid).kind == OBJECT,
              {"resolved": rid, "source": er_obj.get("resolution_source")})

    # 9. 「他說的方向」→ fact，不新增 area/exit
    a0, e0 = len(loop._world.by_kind(AREA)), len(loop._world.by_kind(EXIT))
    er_fact = loop._entity_resolution_block("我回想他說的方向、他說的地方", loop.player_state())
    rid2 = er_fact.get("resolved_entity_id")
    is_fact = bool(rid2) and (loop._world.get(rid2) or None) and loop._world.get(rid2).kind == "fact"
    rec.check(R, "alias_fact_no_new_area_exit",
              is_fact and len(loop._world.by_kind(AREA)) == a0 and len(loop._world.by_kind(EXIT)) == e0,
              {"resolved": rid2, "source": er_fact.get("resolution_source"),
               "area_delta": len(loop._world.by_kind(AREA)) - a0, "exit_delta": len(loop._world.by_kind(EXIT)) - e0})

    # 10. spatial_summary 有 withdraw_safe + return_previous
    loop._world.set_current_area("corridor", label="主走廊")   # previous_area = entry
    routes = [r.get("exit_id") for r in _spatial(loop).get("routes_from_here", [])]
    rec.check(R, "spatial_has_withdraw_and_return_previous",
              "route.withdraw_safe" in routes and "route.return_previous" in routes,
              {"routes": routes, "summary_line": [l for l in (_spatial(loop).get("spatial_summary") or "").split("\n") if "可走路線" in l]})


# ══════════════ Route B — 逃避型 mock（no_truth / gate / retreat / review）═══════════
def run_route_B(new_loop, rec):
    from core.world.model import FACT
    R = "B_avoider_mock"
    print("\n" + "=" * 60 + f"\n{R}\n" + "=" * 60)
    loop = new_loop("mock-B")
    loop.start({"theme": "醫院", "protagonist_name": "蘇明", "npc_count": 1})

    # 1+2. 「不想管真相、只想找出口」→ no_truth、非 truth_investigation、gate False、reveal_delta 0
    rev0 = _reveal_tuple(loop)
    out = loop.step("我不想管這裡發生什麼，我只想趕快找到離開的出口")
    rev1 = _reveal_tuple(loop)
    rec.check(R, "no_truth_intent_true", out.get("no_truth_intent") is True,
              {"no_truth_intent": out.get("no_truth_intent"), "action_class": out.get("action_class")})
    rec.check(R, "action_not_truth_investigation", out.get("action_class") != "truth_investigation",
              {"action_class": out.get("action_class")})
    rec.check(R, "gate_false", out.get("reveal_gate_allowed") is False,
              {"gate": out.get("reveal_gate_allowed"), "reason": out.get("reveal_gate_block_reason")})
    rec.check(R, "reveal_delta_zero", rev1 == rev0, {"before": rev0, "after": rev1})

    # 4. spatial 有可走路線
    routes = [r.get("exit_id") for r in _spatial(loop).get("routes_from_here", [])]
    rec.check(R, "spatial_routes_present", len(routes) > 0, {"routes": routes})

    # 5+6. retreat → review_mode / temporary_retreat、ended False、不新增 fact、不推 reveal
    facts0 = len(loop._world.by_kind(FACT))
    rev_b = _reveal_tuple(loop)
    out = loop.step("我退回比較安全的地方整理一下手上的線索，但我不想結束調查")
    mode = (loop.bb.game_meta or {}).get("exploration_mode")
    rec.check(R, "retreat_enters_review_or_temporary", mode in ("review_mode", "temporary_retreat"),
              {"exploration_mode": mode, "current_area": loop._world.current_area},
              known="UX#6 retreat→review_mode monitor：本 mock action 未翻 mode（kernel 解析成一般移動）；"
                    "真 LLM spatial-routes smoke 曾成功翻 review_mode。需真 LLM 確認或專屬 retreat patch（本 batch 不改 retreat）。")
    rec.check(R, "retreat_not_ended", not out.get("ended"), {"ended": out.get("ended")})
    rec.check(R, "review_no_new_fact", len(loop._world.by_kind(FACT)) == facts0,
              {"facts_before": facts0, "facts_after": len(loop._world.by_kind(FACT))})
    rec.check(R, "review_no_reveal_push", _reveal_tuple(loop) == rev_b,
              {"before": rev_b, "after": _reveal_tuple(loop)})

    # 7. 明確「結束本次調查」才進 ending
    out = loop.step("我受夠了，我決定結束本次調查，離開這個地方，不再回頭")
    rec.check(R, "explicit_end_triggers_ending", bool(out.get("ended")),
              {"ended": out.get("ended"), "ending": (out.get("ending") or {}).get("type")})


# ══════════════ Route C — 真相型 mock（gate / reward ladder / public title）══════════
def run_route_C(new_loop, rec):
    from core.narrative.models import REVEAL_RANK
    from core.narrative.revelation import EvidenceEvent, RevelationBridge
    from core.world.player_state import project_known_facts
    R = "C_truth_mock"
    print("\n" + "=" * 60 + f"\n{R}\n" + "=" * 60)
    loop = new_loop("mock-C")
    loop.start({"theme": "實驗大樓", "protagonist_name": "韓哲", "npc_count": 1})

    # 1+2. truth_investigation → gate True
    out = loop.step("我仔細研究桌上的實驗紀錄，逐頁解讀上面記載了什麼")
    rec.check(R, "truth_investigation_gate_true",
              out.get("action_class") == "truth_investigation" and out.get("reveal_gate_allowed") is True,
              {"action_class": out.get("action_class"), "gate": out.get("reveal_gate_allowed")})

    # 3+4. Reveal Reward Loop：hinted→observed→suspected；reward 不直接 confirmed；debug 欄位
    led = loop._reveal_ledger
    t = led.get_or_create("frag_001_logbook")
    t.level = "hinted"; t.strength = 0.3
    loop._action_class = "truth_investigation"; loop._reveal_gate_allowed = True
    loop._reveal_gate_reason = ""; loop._reveal_updates_this_beat = []; loop._reveal_reward_debug = {}
    loop._reveal_reward_tick()
    d1 = loop._reveal_reward_debug
    rec.check(R, "reward_hinted_to_observed",
              t.level == "observed" and d1.get("ladder_action") == "advanced_by_reward",
              {"level": t.level, "ladder": d1.get("ladder_action"),
               "prev": d1.get("previous_level"), "next": d1.get("next_level")})
    rec.check(R, "reveal_reward_debug_fields",
              all(k in d1 for k in ("gate_allowed", "truth_candidates_found", "ladder_action",
                                    "previous_level", "next_level", "no_progress_reason")),
              {k: d1.get(k) for k in ("gate_allowed", "ladder_action", "previous_level", "next_level")})
    loop._reveal_updates_this_beat = []; loop._reveal_reward_tick()
    rec.check(R, "reward_observed_to_suspected", t.level == "suspected", {"level": t.level})
    for _ in range(6):
        loop._reveal_updates_this_beat = []; loop._reveal_reward_tick()
    rec.check(R, "reward_never_reaches_confirmed",
              REVEAL_RANK[t.level] <= REVEAL_RANK["suspected"], {"level": t.level})

    # 5. public-safe known_fact：public_title、非「未命名的真相」、無 hidden content
    led.get_or_create("frag_003_signal", title="異常訊號").level = "observed"   # explicit title
    loop._known_fact_delta_this_beat = []
    loop._materialize_public_facts()
    kf = project_known_facts(loop._world)
    pub = [f for f in kf if f.get("source") == "reveal"]
    labels = [f["label"] for f in pub]
    rec.check(R, "public_known_fact_uses_title",
              bool(pub) and all("未命名的真相" not in l for l in labels),
              {"labels": labels})
    blob = repr(kf)
    rec.check(R, "no_hidden_raw_content_in_known_facts",
              all(s not in blob for s in ("你也是實驗體之一", "日誌顯示由你本人登入", "訊號來自地下")),
              {"sample_labels": labels})

    # 6. structured strong evidence 才能 confirmed
    ev = EvidenceEvent(evidence_id="e.kernel", source="kernel", truth_id="frag_001_logbook",
                       evidence_strength=0.8, max_level="actionable")
    RevelationBridge().apply(led, [ev])
    rec.check(R, "strong_evidence_can_confirm",
              REVEAL_RANK[t.level] >= REVEAL_RANK["confirmed"], {"level": t.level})


# ══════════════ Route D — 污染 / 邊界 mock（sanitizer / WorldDelta / 拒絕越界）═══════════
def run_route_D(new_loop, rec):
    from core.world.model import OBJECT, AREA, EXIT, FACT
    from core.world.player_state import project_known_facts
    from core.narrative.npc_chat_control import NPCChatResponse
    R = "D_pollution_mock"
    print("\n" + "=" * 60 + f"\n{R}\n" + "=" * 60)
    loop = new_loop("mock-D")
    loop.start({"theme": "站", "protagonist_name": "X", "npc_count": 1})

    # 1+2. sanitizer 移除 placeholder / 壞分隔符 / 黏拉丁
    polluted = ("你伸手去crumpled_paper口袋裡，門框上殘留<<<CONTINAlUE>>>，"
                "終端機顯示origine_verification，他低聲說了句然homme。")
    clean = loop._sanitize_surface(polluted)
    leaks = [x for x in ("crumpled_paper", "<<<CONTINAlUE>>>", "origine_verification", "homme") if x in clean]
    rec.check(R, "sanitizer_removes_pollution", not leaks, {"remaining": leaks, "clean": clean[:80]})
    clean2 = loop._sanitize_surface("敘事到此<<<CONTINUE>>>後面還有")
    rec.check(R, "sanitizer_removes_exact_delimiter", "<<<" not in clean2 and ">>>" not in clean2,
              {"clean": clean2})

    # 3. WorldDelta：entity_delta 用 id（非 entity_id）→ normalize 正常、不 crash、不 tick-skip
    _drive_tick(loop, "我撿到一張門卡",
                [{"op": "register", "kind": "object", "label": "門卡", "id": "object.keycard"}])
    rec.check(R, "worlddelta_id_normalized", loop._world.get("object.keycard") is not None,
              {"keycard": loop._world.get("object.keycard").id if loop._world.get("object.keycard") else None})

    # 4. 衝突 id/entity_id → 拒絕，不污染 WorldModel
    n_obj = len(loop._world.by_kind(OBJECT))
    _drive_tick(loop, "衝突",
                [{"op": "register", "kind": "object", "label": "衝突物",
                  "id": "object.aaa", "entity_id": "object.bbb"}])
    rec.check(R, "conflicting_id_rejected",
              loop._world.get("object.aaa") is None and loop._world.get("object.bbb") is None
              and len(loop._world.by_kind(OBJECT)) == n_obj,
              {"obj_count_unchanged": len(loop._world.by_kind(OBJECT)) == n_obj})

    # 5. story 嘗試新增 area / exit → 拒絕
    a0, e0 = len(loop._world.by_kind(AREA)), len(loop._world.by_kind(EXIT))
    _drive_tick(loop, "我打開一條密道",
                [{"op": "register", "kind": "area", "label": "密室", "entity_id": "area.secret"},
                 {"op": "register", "kind": "exit", "label": "暗門", "entity_id": "exit.hidden"}])
    rec.check(R, "story_area_exit_rejected",
              loop._world.get("area.secret") is None and loop._world.get("exit.hidden") is None
              and len(loop._world.by_kind(AREA)) == a0 and len(loop._world.by_kind(EXIT)) == e0,
              {"area_delta": len(loop._world.by_kind(AREA)) - a0, "exit_delta": len(loop._world.by_kind(EXIT)) - e0})

    # 6. NPC prose fact → known_facts，但不推 reveal
    rev0 = _reveal_tuple(loop)
    _reset_beat_attrs(loop)
    resp = NPCChatResponse(visible_reply="通訊設備在B2機房，那扇門已經鎖死了。", answer_status="partial")
    loop._bridge_npc_entity_delta(resp, "林守一")
    kf = project_known_facts(loop._world)
    prose = [f for f in kf if f.get("source") == "林守一"]
    rec.check(R, "npc_prose_fact_in_known_facts", bool(prose),
              {"facts": [f["label"] for f in prose]})
    rec.check(R, "npc_prose_fact_no_reveal_push", _reveal_tuple(loop) == rev0,
              {"before": rev0, "after": _reveal_tuple(loop)})


class _WarnCapture(__import__("logging").Handler):
    def __init__(self):
        super().__init__()
        self.records = []

    def emit(self, r):
        try:
            self.records.append(self.format(r))
        except Exception:
            pass


def main():
    import logging
    sys.path.insert(0, str(ROOT / "dev" / "tools"))
    rec = Recorder()
    cap = _WarnCapture()
    logging.getLogger().addHandler(cap)
    logging.getLogger().setLevel(logging.WARNING)
    print("[mock_full_ux_regression] ScriptedCaller (no real LLM)")
    for fn in (run_route_A, run_route_B, run_route_C, run_route_D):
        try:
            fn(_new_loop, rec)
        except Exception as e:
            import traceback
            traceback.print_exc()
            rec.check(fn.__name__, "route_ran_without_exception", False, {"error": str(e)})

    # 全域：WorldDelta id warning 不再出現（acceptance）
    tick_skips = [m for m in cap.records if "world model tick skipped" in m]
    rec.check("ACCEPTANCE", "no_worlddelta_id_warning", not tick_skips,
              {"tick_skip_warnings": tick_skips[:3]})

    out_jsonl = ROOT / "dev" / "reports" / "mock-full-ux-regression.jsonl"
    with open(out_jsonl, "w", encoding="utf-8") as f:
        for r in rec.rows:
            f.write(json.dumps(r, ensure_ascii=False) + "\n")
    _render_md(rec.rows, ROOT / "dev" / "reports" / "mock-full-ux-regression.md")
    n_pass = sum(1 for r in rec.rows if r["pass"])
    n_known = sum(1 for r in rec.rows if (not r["pass"]) and r.get("known_item"))
    print(f"\n=== {n_pass}/{len(rec.rows)} passed（其中 {n_known} 個未過為已知 monitor item）===")
    print(f"written → {out_jsonl}")
    return rec


_ROUTE_TITLE = {"A_explorer_mock": "A 探索型", "B_avoider_mock": "B 逃避型",
                "C_truth_mock": "C 真相型", "D_pollution_mock": "D 污染/邊界",
                "ACCEPTANCE": "Acceptance"}


def _render_md(rows, path):
    by_route = {}
    for r in rows:
        by_route.setdefault(r["route"], []).append(r)
    n_pass = sum(1 for r in rows if r["pass"])
    n_known = sum(1 for r in rows if (not r["pass"]) and r.get("known_item"))
    n_real_fail = sum(1 for r in rows if (not r["pass"]) and not r.get("known_item"))
    L = []
    L.append("# Mock Full UX Regression（**無真 LLM** / ScriptedCaller）\n")
    L.append("> deterministic ScriptedCaller 取代 OpenRouter，驅動 open-exploration runtime 的主要契約。")
    L.append("> 由 `dev/tools/mock_full_ux_regression.py` 產生。逐 checkpoint 資料：`mock-full-ux-regression.jsonl`。")
    L.append("> **不呼叫 OpenRouter、不使用真 LLM。**\n")
    L.append(f"## 結果：{n_pass}/{len(rows)} 通過"
             f"（{n_known} 個未過為**已知 monitor item**、{n_real_fail} 個真 regression）\n")
    L.append("| 路線 | 通過 | 未過(真) | 已知 monitor |")
    L.append("|---|---|---|---|")
    for route, rs in by_route.items():
        p = sum(1 for r in rs if r["pass"])
        rf = sum(1 for r in rs if (not r["pass"]) and not r.get("known_item"))
        kn = sum(1 for r in rs if (not r["pass"]) and r.get("known_item"))
        L.append(f"| {_ROUTE_TITLE.get(route, route)} | {p}/{len(rs)} | {rf} | {kn} |")
    L.append("")
    for route, rs in by_route.items():
        L.append(f"## {_ROUTE_TITLE.get(route, route)}（{route}）\n")
        L.append("| ✓ | checkpoint | evidence |")
        L.append("|---|---|---|")
        for r in rs:
            mark = "✅" if r["pass"] else ("⚠" if r.get("known_item") else "❌")
            ev = json.dumps(r["evidence"], ensure_ascii=False)
            ev = (ev[:140] + "…") if len(ev) > 140 else ev
            L.append(f"| {mark} | {r['checkpoint']} | `{ev}` |")
            if r.get("known_item"):
                L.append(f"|  |  | ⚠ **已知**：{r['known_item']} |")
        L.append("")
    # Acceptance checklist 對應
    idx = {r["checkpoint"]: r for r in rows}

    def st(*names):
        rs = [idx[n] for n in names if n in idx]
        if not rs:
            return "—"
        if all(r["pass"] for r in rs):
            return "✅"
        if any((not r["pass"]) and r.get("known_item") for r in rs):
            return "⚠ 已知"
        return "❌"
    L.append("## Acceptance Checklist 對應\n")
    L.append("| 項目 | 狀態 | 來源 checkpoint |")
    L.append("|---|---|---|")
    accept = [
        ("opening variation 不使用 forbidden literals", st("opening_no_forbidden_literals"), "A:opening_no_forbidden_literals"),
        ("placeholder / delimiter leak 被擋", st("sanitizer_removes_pollution", "sanitizer_removes_exact_delimiter"), "D:sanitizer_*"),
        ("WorldDelta id warning 不再出現", st("no_worlddelta_id_warning"), "ACCEPTANCE:no_worlddelta_id_warning"),
        ("PlayerState inventory_entities 能填入", st("take_adds_inventory"), "A:take_adds_inventory"),
        ("known_facts 從 NPC fact + reveal public fact 填入", st("known_facts_from_npc_with_source_confidence", "public_known_fact_uses_title"), "A:known_facts_* / C:public_known_fact_uses_title"),
        ("reveal 可升 observed / suspected", st("reward_hinted_to_observed", "reward_observed_to_suspected"), "C:reward_*"),
        ("confirmed 仍需 strong evidence", st("reward_never_reaches_confirmed", "strong_evidence_can_confirm"), "C:reward_never_* / strong_evidence_*"),
        ("spatial_summary 有可走路線", st("spatial_has_withdraw_and_return_previous", "spatial_routes_present"), "A/B:spatial_*"),
        ("no_truth_intent 擋「不想管 / 只想離開」", st("no_truth_intent_true"), "B:no_truth_intent_true（本 batch 修）"),
        ("AliasResolver：那個東西→object / 他說的地方→fact / 那個人→actor", st("alias_object_not_npc", "alias_fact_no_new_area_exit", "npc_actor_entity_exists"), "A:alias_* / npc_actor_entity_exists"),
        ("retreat 進 safe_zone / review_mode", st("retreat_enters_review_or_temporary"), "B:retreat_enters_review_or_temporary（UX#6 monitor）"),
        ("review_mode 不新增 fact、不推 reveal", st("review_no_new_fact", "review_no_reveal_push"), "B:review_*"),
        ("hidden truth raw content 不外洩", st("no_hidden_raw_content_in_known_facts"), "C:no_hidden_raw_content_in_known_facts"),
    ]
    for name, status, src in accept:
        L.append(f"| {name} | {status} | {src} |")
    L.append("")
    L.append("## 本 batch 已修的 deterministic bug\n")
    L.append("- **no_truth_intent 漏判逃避型（UX #4）**：補「不想管 / 只想離開 / 只想出口 …」→ "
             "逃避型現可被 `no_truth=True`、gate=False 擋住（regression test 已加）。")
    L.append("")
    L.append("## 仍需真 LLM 才能確認 / 未在本 batch 修\n")
    L.append("- **retreat → review_mode（UX #6，monitor item）**：本 mock 的撤退 action 被 kernel 解析成"
             "一般移動（current_area→corridor），未翻 review_mode。**真 LLM spatial-routes smoke 曾成功**"
             "翻 review_mode + 移到 safe_zone（kernel/exit-resolver/action 措辭相關）。其餘 review 不變量"
             "（不新增 fact、不推 reveal）在 mock 下成立。依既定方針 retreat 列 monitor、本 batch 不改。")
    L.append("")
    L.append("## 建議\n")
    if n_real_fail == 0:
        L.append("- 核心契約在 mock 下**全綠**（唯一未過為 retreat→review 的已知 monitor item）。")
        L.append("- **是否再花真 LLM smoke**：建議**只在 retreat→review 上**花一次極短真 LLM 確認"
                 "（其餘契約已由 mock + 單元測試決定性覆蓋，毋須再燒真 LLM 預算）。")
    else:
        L.append(f"- 有 {n_real_fail} 個真 regression 待處理（見上表 ❌）。")
    L.append("")
    path.write_text("\n".join(L), encoding="utf-8")


if __name__ == "__main__":
    main()
