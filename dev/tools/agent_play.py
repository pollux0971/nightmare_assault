#!/usr/bin/env python3
"""agent_play — 讓外部 AI / 腳本自動化遊玩本遊戲的無頭（headless）驅動器。

本遊戲的程式化介面是 `core.orchestrator_loop.BeatLoop`：
  - `loop.start(opts) -> {opening_sequence, narrative, decision_point}`（同步）
  - `loop.step(text) -> {narrative, decision_point, ended, ending, ...}`（同步）
webview 那層（webview_app.API.start_game/submit_decision）是 thread + 事件推播給前端，
不適合 AI 同步驅動；自動化請走 BeatLoop。本檔把它包成兩種好用的模式：

  1) JSON-over-stdio 協定（預設）——給「另一個 AI」用：
     · 啟動後，stdout 印一行開場 observation（JSON）。
     · 之後每從 stdin 讀一行 action（JSON 或純文字），就推進一個 beat，stdout 印一行 observation。
     · observation schema:
         {"beat": int, "narrative": str, "options": [str,...], "free_input_hint": str,
          "present_npcs": [str,...], "reveal_progress": {"hinted":n,"observed":n,"suspected":n,
          "confirmed":n,"total":n}, "ended": bool, "ending": {...}|null}
     · action schema（擇一）:
         "我檢查桌上的紙條"                        # 純文字 = 一個遊戲動作
         {"action": "我往前走"}                    # 同上
         {"chat": {"npc": "吳靜", "text": "432.7 是什麼？"}}   # 開聊天室問一句（不推進 beat）
         {"cmd": "quit"}                           # 結束

  2) --auto：內建簡單策略自動玩（煙霧測試 / 產生逐字稿），印人讀格式。

用法:
  # 互動 / 被另一個 AI 以管線驅動（JSONL）
  .venv/bin/python dev/tools/agent_play.py --theme "廢棄海事研究站" --name 周凱 --npc-count 2 --flag
  # 自我遊玩煙霧測試
  .venv/bin/python dev/tools/agent_play.py --auto --max-beats 12 --flag

旗標:
  --flag            開 ENABLE_NARRATIVE_CONTROL（敘事控制 v0.2 揭露橋接）
  --config PATH     config.json 路徑（預設 專案根 config/config.json；含 OpenRouter api_key）
  --no-llm          用內建假 caller（不呼叫 LLM；給 CI / 介面冒煙用，不需 api_key）
"""
from __future__ import annotations

import argparse
import json
import os
import random
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))


class JsonlLogger:
    """把整局 observation / action / assertion 落成 JSONL（每行一筆，供事後比對 / CI 斷言）。"""

    def __init__(self, path: str | None):
        self._fh = open(path, "w", encoding="utf-8") if path else None
        self.passed = 0
        self.failed = 0

    def _w(self, rec: dict):
        if self._fh:
            self._fh.write(json.dumps(rec, ensure_ascii=False) + "\n")
            self._fh.flush()

    def observation(self, beat: int, data: dict):
        self._w({"type": "observation", "beat": beat, "data": data})

    def action(self, beat: int, action: str):
        self._w({"type": "action", "beat": beat, "action": action})

    def assertion(self, case: str, ok: bool, notes: str = ""):
        self.passed += int(bool(ok)); self.failed += int(not ok)
        self._w({"type": "assertion", "case": case, "pass": bool(ok), "notes": notes})

    def close(self):
        if self._fh:
            self._w({"type": "summary", "assertions_passed": self.passed,
                     "assertions_failed": self.failed})
            self._fh.close()


def _reveal_delta_notes(prev: dict, cur: dict) -> str:
    """描述 reveal_progress 增量（如 'hinted +1, confirmed +1'）。無增量回空字串。"""
    parts = []
    for k in ("hinted", "observed", "suspected", "confirmed", "revealed_fragments"):
        d = int(cur.get(k, 0)) - int(prev.get(k, 0))
        if d > 0:
            parts.append(f"{k} +{d}")
    return ", ".join(parts)


def _emit_step_assertions(logger: JsonlLogger, prev: dict, cur: dict):
    """每 beat 後的內建斷言：揭露橋接有沒有把調查轉成進度。"""
    notes = _reveal_delta_notes(prev, cur)
    if notes:
        logger.assertion("RevealBridge", True, notes)


def _emit_ending_assertions(logger: JsonlLogger, ending: dict, reveal: dict):
    """結局時的內建斷言：不再 0/X；ambiguous 表層；0 confirmed 不得 clean。"""
    found = int((((ending or {}).get("recap") or {}).get("partial") or {}).get("found",
                reveal.get("hinted", 0)))
    logger.assertion("RevealNonZero", found >= 1,
                     f"found={found} (調查後結局碎片數應 ≥1)")
    surface = (ending or {}).get("ending_surface")
    if (ending or {}).get("type") == "escape":
        confirmed = int(reveal.get("confirmed", 0))
        if confirmed == 0:
            logger.assertion("AmbiguousEscape", surface == "ambiguous_escape",
                             f"0 confirmed → surface 應為 ambiguous_escape（實際 {surface}）")


def _make_caller(config_path: str, no_llm: bool):
    """回傳 (caller, note)。no_llm → 假 caller；否則用 config.json 的 OpenRouter 真 caller。"""
    if no_llm:
        from tests.test_narrative_v2_integration_nr import FakeCaller  # 既有的假 caller
        return FakeCaller(), "fake-caller(no-llm)"
    from core.llm.client import OpenRouterClient
    from core.agents.base import SkillCaller, SkillLoader
    from webview_app import DEFAULT_TEMPS
    cfg = json.load(open(config_path, encoding="utf-8"))
    client = OpenRouterClient({"api_key": cfg["api_key"], "base_url": cfg["base_url"],
                              "agent_models": cfg["agent_models"], "timeout": cfg.get("timeout", 120)})
    caller = SkillCaller(client, SkillLoader(str(ROOT / "skills")),
                         temperature_by_agent=DEFAULT_TEMPS)
    models = cfg.get("agent_models", {}).get("story", ["?"])
    return caller, f"openrouter(story={models[0]})"


def _new_loop(caller):
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database
    from core.signal import SignalBus
    db = Database(tempfile.mktemp(suffix=".db"))
    return BeatLoop(caller, Blackboard(), db, SignalBus(), run_id="agent-play", use_kernel=True)


def _reveal_progress(loop) -> dict:
    rb = loop.bb.snapshot().get("revealed_bible") or {}
    tp = rb.get("truth_progress") or {}
    from core.narrative.revelation import REVEAL_RANK
    def c(r):
        return sum(1 for v in tp.values() if REVEAL_RANK.get(v.get("level", "hidden"), 0) >= r)
    return {"hinted": c(1), "observed": c(2), "suspected": c(3), "confirmed": c(4),
            "total": len(tp), "revealed_fragments": len(rb.get("revealed_fragments") or [])}


def _present_npcs(loop) -> list:
    return [n.get("name") for n in loop.bb.snapshot().get("npc_registry") or []
            if isinstance(n, dict) and n.get("presence", "present") == "present"]


def _npc_tiers(loop) -> dict:
    """HE1：NPC 分層——visible（在場畫面中）/ known（玩家已知）/ chat_available（可對話）。"""
    reg = [n for n in loop.bb.snapshot().get("npc_registry") or [] if isinstance(n, dict) and n.get("name")]
    visible = [n["name"] for n in reg if n.get("presence", "present") == "present"]
    known = [n["name"] for n in reg]                  # 出現在 registry = 玩家已認知
    return {"visible_npcs": visible, "known_npcs": known, "chat_available_npcs": visible}


def _world_state(loop) -> dict:
    """世界狀態快照（給 AI 偵測導航 / 危險 / 後果問題）。"""
    gs = getattr(loop, "_game_state", None)
    if gs is not None:
        return {"current_scene": getattr(gs, "current_scene", None),
                "danger_level": int(getattr(gs, "danger_level", 0) or 0),
                "clue_count": len(getattr(gs, "clues", {}) or {}),
                "item_count": len(getattr(gs, "inventory", {}) or {})}
    snap = loop.bb.snapshot()
    return {"current_scene": None, "danger_level": 0, "clue_count": 0,
            "item_count": len((snap.get("shared_inventory") or {}).get("items", []))}


def annotate_world(obs: dict, track: dict) -> list:
    """算本 beat 的世界後果 / 停滯指標，加進 obs['deltas']；回傳該 beat 的內建斷言清單
    [(case, ok, notes), ...]（由 driver 在 log 完 observation 後再 emit，確保順序與配對）。

    track 跨 beat 保存 prev 與停滯計數（由 driver 持有）。
    """
    cur = obs.get("world_state") or {}
    prev = track.get("prev") or {}
    dbg = obs.get("debug") or {}
    scene_changed = bool(prev) and cur.get("current_scene") != prev.get("current_scene")
    new_clues = max(0, cur.get("clue_count", 0) - prev.get("clue_count", 0))
    new_items = max(0, cur.get("item_count", 0) - prev.get("item_count", 0))
    danger_delta = cur.get("danger_level", 0) - prev.get("danger_level", 0)
    reveal_up = len(dbg.get("reveal_updates_this_beat") or [])
    evidence = int(dbg.get("evidence_events_this_beat", 0) or 0)
    # P0 #4：world facts（NPC/story 寫入的 world_state_fact）也算後果
    wp = obs.get("world_progress") or {}
    new_world_facts = len(wp.get("new_world_facts_this_beat") or [])
    changed_exits = len(wp.get("changed_exits_this_beat") or [])
    is_exit_offer = dbg.get("escape_step") == "ambiguous"
    is_narration = bool(obs.get("ended"))
    had_consequence = bool(scene_changed or new_clues or new_items or danger_delta
                           or reveal_up or evidence or new_world_facts or changed_exits
                           or is_exit_offer)

    track["beats_in_scene"] = (track.get("beats_in_scene", 0) + 1) if not scene_changed else 1
    track["since_consequence"] = 0 if had_consequence else track.get("since_consequence", 0) + 1
    track["since_reveal"] = 0 if reveal_up else track.get("since_reveal", 0) + 1

    obs["world_state"] = cur
    obs["deltas"] = {
        "scene_changed": scene_changed, "danger_delta": danger_delta,
        "new_clues": new_clues, "new_items": new_items, "reveal_updates": reveal_up,
        "new_world_facts": new_world_facts, "changed_exits": changed_exits,
        "is_exit_offer": is_exit_offer, "had_consequence": had_consequence,
        "beats_in_scene": track["beats_in_scene"],
        "beats_since_consequence": track["since_consequence"],
        "beats_since_reveal": track["since_reveal"],
    }
    asserts = []
    if prev and not is_narration:
        # Player Sovereignty 核心：每個行動都該有可檢查的世界後果
        asserts.append(("WorldResponds", had_consequence,
                        "" if had_consequence else "本 beat 無任何世界後果（場景/線索/道具/危險/揭露/world_fact 皆未變）"))
        if track["beats_in_scene"] >= 5:
            asserts.append(("NotStuckInScene", False,
                            f"已連續 {track['beats_in_scene']} beat 停在 {cur.get('current_scene')}"))
        if cur.get("danger_level", 0) >= 5 and track["since_consequence"] >= 3:
            asserts.append(("DangerProducesThreat", False,
                            f"danger={cur.get('danger_level')} 但連 {track['since_consequence']} beat 無後果"))
    track["prev"] = cur
    return asserts


def _debug_block(loop, step_result) -> dict:
    """HE1：observation.debug——定位 bug 用；只含 truth_id，不含 hidden content。"""
    sr = step_result or {}
    qm = getattr(loop, "_quality_meta", None) or {}
    return {
        "committed_event": sr.get("committed_event"),
        "progress_delta": sr.get("progress_delta"),
        "escape_step": sr.get("escape_step", "none"),
        "evidence_events_this_beat": sr.get("evidence_events_this_beat", 0),
        "unmapped_evidence_this_beat": sr.get("unmapped_evidence_this_beat", 0),
        "reveal_updates_this_beat": sr.get("reveal_updates", []),   # 僅 truth_id/from/to，無 content
        # v0.7 P3：beat 渲染量測（beat_type/target/actual/too_short/short_streak）
        "beat_rendering": sr.get("beat_rendering", {}),
        # WorldConsequence vs TruthEvidence split：分類 + reveal 閘決定
        "action_class": sr.get("action_class", "unknown"),
        "no_truth_intent": sr.get("no_truth_intent", False),
        "reveal_gate_allowed": sr.get("reveal_gate_allowed", True),
        "reveal_gate_block_reason": sr.get("reveal_gate_block_reason", ""),
        "blocked_reveal_candidates": sr.get("blocked_reveal_candidates", []),
        "quality_gate": {"passed": qm.get("passed", True),
                         "repaired": qm.get("repaired", False),
                         "fallback": qm.get("fallback", False)},
        "model_used": {"story": (getattr(loop, "_last_prompt_meta", None) or {}).get("model", "")},
    }


def _dp_to_obs(loop, narrative, dp, ended, ending, step_result=None) -> dict:
    ended = bool(ended)
    # HA1：ended ⇒ options=[]、free_input_hint=null（輸出層再保險一次）
    options = [] if ended else [getattr(o, "text", "")
                                for o in (getattr(dp, "suggested_options", None) or [])]
    tiers = _npc_tiers(loop)
    _sd = ((step_result or {}).get("spatial_debug")
           or (loop.spatial_debug() if hasattr(loop, "spatial_debug") else {})) or {}
    return {
        "beat": loop.beat_number,
        "narrative": narrative,
        "options": options,
        "free_input_hint": None if ended else (getattr(dp, "free_input_hint", "") or ""),
        "present_npcs": tiers["visible_npcs"],        # 向後相容保留
        "visible_npcs": tiers["visible_npcs"],
        "known_npcs": tiers["known_npcs"],
        "chat_available_npcs": tiers["chat_available_npcs"],
        "reveal_progress": _reveal_progress(loop),
        "world_state": _world_state(loop),
        "world_progress": ((step_result or {}).get("world_progress")
                           or (loop.world_progress(dp) if hasattr(loop, "world_progress") else {})),
        "spatial_debug": _sd,
        # Step 5：玩家狀態投影 + 確定性摘要（observation-only）
        "player_state": ((step_result or {}).get("player_state")
                         or (loop.player_state() if hasattr(loop, "player_state") else {})),
        "player_state_summary": (step_result or {}).get("player_state_summary", ""),
        "player_state_summary_truncated": (step_result or {}).get("player_state_summary_truncated", False),
        "player_state_summary_source": (step_result or {}).get(
            "player_state_summary_source", "deterministic_projection"),
        # Spatial UX：玩家/QA 可讀摘要（top-level，方便顯示/記錄；deterministic、不餵 story）
        "spatial_summary": _sd.get("spatial_summary", ""),
        "spatial_summary_truncated": _sd.get("spatial_summary_truncated", False),
        "spatial_summary_source": _sd.get("spatial_summary_source", "deterministic_projection"),
        "debug": _debug_block(loop, step_result),
        "ended": ended,
        "ending": _ending_dict(loop, ending) if ended else None,
    }


# 預設遮罩 hidden truth（HA3）；只有 --debug-reveal-truth 才露 full hidden content。
_REVEAL_TRUTH_DEBUG = False


def _ending_dict(loop, ending) -> dict:
    if not ending:
        return {}
    try:
        from core.agents.ending import render_ending_text
        text = render_ending_text(ending, mode="full" if _REVEAL_TRUTH_DEBUG else "masked")
    except Exception:
        text = ending.get("closing", "")
    # HD2：ending rendered_text 也消毒
    try:
        from core.narrative.sanitizer import sanitize_text
        text = sanitize_text(text)
    except Exception:
        pass
    # HA3：玩家面 recap 用 public_recap（無 hidden content）；debug flag 才給 full partial
    recap = None
    led = getattr(loop, "_reveal_ledger", None)
    if led is not None:
        from core.narrative.revelation import public_recap, recap_from_ledger
        recap = recap_from_ledger(led) if _REVEAL_TRUTH_DEBUG else public_recap(led)
    elif _REVEAL_TRUTH_DEBUG:
        recap = (ending.get("recap") or {}).get("partial")
    return {"type": ending.get("type"), "ending_surface": ending.get("ending_surface"),
            "escape_quality": ending.get("escape_quality"),
            "recap": recap, "rendered_text": text}


# ── JSON-over-stdio 模式（給另一個 AI 用）────────────────────────────────────
def run_stdio(loop, opts, *, max_beats=20, logger=None, stop_on_ending=True):
    logger = logger or JsonlLogger(None)
    res = loop.start(opts)
    track = {}                                        # 跨 beat：世界狀態 + 停滯計數
    obs = _dp_to_obs(loop, res["narrative"], res["decision_point"], False, None)
    obs["opening_sequence"] = res.get("opening_sequence")
    asserts = annotate_world(obs, track)
    print(json.dumps(obs, ensure_ascii=False), flush=True)
    logger.observation(loop.beat_number, obs)
    for a in asserts:
        logger.assertion(*a)
    if loop.ended:
        return
    prev_reveal = obs["reveal_progress"]
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        action, chat = _parse_action(line)
        if action == "__quit__":
            break
        if chat is not None:                       # 聊天室一句（不推進 beat）
            reply = _do_chat(loop, chat)
            cur_reveal = _reveal_progress(loop)
            print(json.dumps({"chat_reply": reply, "npc": chat.get("npc"),
                              "reveal_progress": cur_reveal,
                              "npc_chat_debug": getattr(loop, "_npc_chat_debug", {})},
                             ensure_ascii=False), flush=True)
            notes = _reveal_delta_notes(prev_reveal, cur_reveal)
            if notes:
                logger.assertion("RevealBridge", True, "npc_chat: " + notes)
            prev_reveal = cur_reveal
            continue
        logger.action(loop.beat_number, action)
        out = loop.step(action)
        obs = _dp_to_obs(loop, out["narrative"], out["decision_point"],
                         out.get("ended"), out.get("ending"), step_result=out)
        asserts = annotate_world(obs, track)
        print(json.dumps(obs, ensure_ascii=False), flush=True)
        logger.observation(loop.beat_number, obs)
        for a in asserts:
            logger.assertion(*a)
        _emit_step_assertions(logger, prev_reveal, obs["reveal_progress"])
        prev_reveal = obs["reveal_progress"]
        if out.get("ended"):
            _emit_ending_assertions(logger, out.get("ending"), prev_reveal)
            if stop_on_ending:
                break
        if loop.beat_number >= max_beats:          # 防卡死：beat 上限
            logger.assertion("MaxBeatsReached", True, f"max_beats={max_beats}")
            print(json.dumps({"stopped": "max_beats", "max_beats": max_beats},
                             ensure_ascii=False), flush=True)
            break


def _parse_action(line: str):
    """回傳 (action_text|'__quit__', chat_dict|None)。"""
    if line.startswith("{"):
        try:
            obj = json.loads(line)
            if obj.get("cmd") == "quit":
                return "__quit__", None
            if "chat" in obj:
                return "", obj["chat"]
            return str(obj.get("action") or obj.get("text") or ""), None
        except Exception:
            return line, None
    return line, None


def _do_chat(loop, chat: dict) -> str:
    from core.agents.npc_chat import run_npc_chat_structured
    npc = chat.get("npc"); text = chat.get("text", "")
    try:
        resp = run_npc_chat_structured(loop.caller, loop.bb, npc, text)
    except Exception as e:                          # 對話失敗不可拖垮整個 session
        return f"（{npc} 沒有回應：{e}）"
    try:                                           # #10/P-plumbing：結構化 evidence 走 gate → bridge
        loop.bridge_npc_evidence(resp, npc_id=npc)
        loop.note_npc_answer(text, resp.answer_status)
        loop.note_focus_npc(npc)                          # Step 5：對話 → 焦點設為該 NPC
    except Exception:
        pass
    return resp.visible_reply


# ── --auto 模式（內建策略自我遊玩；煙霧測試）─────────────────────────────────
def _auto_choice(dp, beat, max_beats):
    opts = [getattr(o, "text", "") for o in (getattr(dp, "suggested_options", None) or [])]
    if beat >= max_beats - 1:
        return "我下定決心，頭也不回地走出去"
    if beat == max_beats - 2:
        return "我試圖離開這個地方，尋找出口"
    for o in opts:
        if any(k in o for k in ["檢查", "調查", "線索", "查看", "搜", "紙", "文件", "深", "痕跡"]):
            return o
    return opts[0] if opts else "我繼續探索並尋找線索"


def run_auto(loop, opts, max_beats, *, logger=None, stop_on_ending=True):
    logger = logger or JsonlLogger(None)
    P = print
    res = loop.start(opts)
    track = {}
    P("=" * 70); P("【開場】"); P(res["narrative"])
    obs = _dp_to_obs(loop, res["narrative"], res["decision_point"], False, None)
    for a in annotate_world(obs, track):
        logger.assertion(*a)
    P("〔reveal〕", obs["reveal_progress"], "〔world〕", obs["world_state"])
    logger.observation(loop.beat_number, obs)
    dp = res["decision_point"]
    prev_reveal = obs["reveal_progress"]
    for beat in range(1, max_beats + 1):
        act = _auto_choice(dp, beat, max_beats)
        P("\n" + "=" * 70); P(f"▶ beat{beat} 動作：{act}")
        logger.action(loop.beat_number, act)
        out = loop.step(act); dp = out["decision_point"]
        obs = _dp_to_obs(loop, out["narrative"], dp, out.get("ended"), out.get("ending"),
                         step_result=out)
        _aw = annotate_world(obs, track)
        P(out["narrative"])
        P("〔reveal〕", obs["reveal_progress"], "〔Δ〕", obs["deltas"])
        # Spatial UX：玩家/QA 面板摘要（debug 輔助，不取代 narrative）
        if obs.get("spatial_summary"):
            _t = "（已截斷）" if obs.get("spatial_summary_truncated") else ""
            P(f"〔spatial{_t}〕\n" + obs["spatial_summary"])
        logger.observation(loop.beat_number, obs)
        for a in _aw:
            logger.assertion(*a)
        _emit_step_assertions(logger, prev_reveal, obs["reveal_progress"])
        prev_reveal = obs["reveal_progress"]
        if out.get("ended"):
            _emit_ending_assertions(logger, out.get("ending"), prev_reveal)
            P("\n" + "#" * 70); P("【結局】", json.dumps(_ending_dict(loop, out["ending"]),
                                                      ensure_ascii=False, indent=2))
            if stop_on_ending:
                return
    P(f"（{loop.beat_number} beat 未觸發結局）")


def main(argv=None):
    ap = argparse.ArgumentParser(description="無頭驅動本遊戲（給外部 AI 自動化測試）")
    ap.add_argument("--theme", default="午夜的廢棄海事研究站（弟弟林晨失蹤）")
    ap.add_argument("--name", default="周凱")
    ap.add_argument("--npc-count", type=int, default=2)
    ap.add_argument("--flag", action="store_true", help="開 ENABLE_NARRATIVE_CONTROL")
    ap.add_argument("--config", default=str(ROOT / "config" / "config.json"))
    ap.add_argument("--no-llm", action="store_true", help="用假 caller（不需 api_key）")
    ap.add_argument("--auto", action="store_true", help="內建策略自我遊玩（否則 JSON-over-stdio）")
    ap.add_argument("--max-beats", type=int, default=20, help="beat 上限（防 agent 卡住一直玩）")
    ap.add_argument("--jsonl-log", default=None,
                    help="把 observation/action/assertion 落 JSONL（如 dev/reports/run-xxx.jsonl）")
    ap.add_argument("--stop-on-ending", action="store_true", default=True,
                    help="觸發結局即停（預設開）")
    ap.add_argument("--no-stop-on-ending", dest="stop_on_ending", action="store_false")
    ap.add_argument("--seed", type=int, default=None,
                    help="設 random 種子（影響離場命運擲骰等 stdlib 隨機；LLM 仍非決定性）")
    ap.add_argument("--debug-reveal-truth", action="store_true",
                    help="（除錯用）observation 露 full hidden truth content；預設遮罩")
    args = ap.parse_args(argv)

    if args.flag:
        os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    global _REVEAL_TRUTH_DEBUG
    _REVEAL_TRUTH_DEBUG = bool(args.debug_reveal_truth)
    if args.seed is not None:
        random.seed(args.seed)

    caller, note = _make_caller(args.config, args.no_llm)
    loop = _new_loop(caller)
    opts = {"theme": args.theme, "protagonist_name": args.name, "npc_count": args.npc_count}
    logger = JsonlLogger(args.jsonl_log)
    try:
        if args.auto:
            print(f"[caller={note}  flag={os.environ.get('ENABLE_NARRATIVE_CONTROL', 'false')}"
                  f"  seed={args.seed}  max_beats={args.max_beats}]")
            run_auto(loop, opts, args.max_beats, logger=logger, stop_on_ending=args.stop_on_ending)
        else:
            print(json.dumps({"meta": {"caller": note,
                                       "narrative_control": os.environ.get("ENABLE_NARRATIVE_CONTROL", "false"),
                                       "seed": args.seed, "max_beats": args.max_beats,
                                       "protocol": "send one action (text or JSON) per line on stdin; "
                                                   "read one observation JSON per line on stdout"}},
                             ensure_ascii=False), flush=True)
            run_stdio(loop, opts, max_beats=args.max_beats, logger=logger,
                      stop_on_ending=args.stop_on_ending)
    finally:
        logger.close()


if __name__ == "__main__":
    main()
