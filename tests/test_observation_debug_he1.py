"""HE1 — Observation Debug Fields 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：observation 含 debug（committed_event/progress_delta/escape_step/evidence_events_this_beat/
unmapped_evidence_this_beat/reveal_updates_this_beat/quality_gate/model_used）+ NPC 三分層；
debug 不含 hidden truth content；ended → options=[]。
"""
from __future__ import annotations

import importlib.util
import json


def _ap():
    spec = importlib.util.spec_from_file_location("agent_play", "dev/tools/agent_play.py")
    ap = importlib.util.module_from_spec(spec); spec.loader.exec_module(ap)
    return ap


def _loop_with(npcs, step_result=None, ledger=None):
    class _BB:
        def snapshot(self):
            return {"npc_registry": npcs,
                    "revealed_bible": {"truth_progress": ledger or {}}}
    class _Loop:
        beat_number = 3
        bb = _BB()
        _quality_meta = {"passed": True, "repaired": False, "fallback": False}
        _reveal_ledger = None
    return _Loop()


def test_observation_has_debug_and_npc_tiers():
    ap = _ap()
    npcs = [{"name": "吳靜", "presence": "present"}, {"name": "魏博文", "presence": "absent"}]
    from core.models import DecisionPoint, Option
    dp = DecisionPoint(situation_recap="走廊。", decision_type="action",
                       suggested_options=[Option(text="前進", tone="bold")],
                       free_input_hint="", beat_meta={"beat_number": 3})
    step = {"committed_event": "inspect_warning_note", "progress_delta": ["new_clue_added"],
            "escape_step": "none", "evidence_events_this_beat": 1,
            "unmapped_evidence_this_beat": 0,
            "reveal_updates": [{"truth_id": "t.sig", "from": "hidden", "to": "hinted"}]}
    obs = ap._dp_to_obs(_loop_with(npcs), "你找到紙條。", dp, False, None, step_result=step)

    # NPC 三分層
    assert obs["visible_npcs"] == ["吳靜"]
    assert set(obs["known_npcs"]) == {"吳靜", "魏博文"}
    assert obs["chat_available_npcs"] == ["吳靜"]
    # debug 欄位齊全
    d = obs["debug"]
    for k in ("committed_event", "progress_delta", "escape_step", "evidence_events_this_beat",
              "unmapped_evidence_this_beat", "reveal_updates_this_beat", "quality_gate", "model_used"):
        assert k in d, f"debug 缺欄位 {k}"
    assert d["committed_event"] == "inspect_warning_note"
    assert d["reveal_updates_this_beat"][0]["truth_id"] == "t.sig"
    assert d["quality_gate"]["passed"] is True


def test_debug_has_no_hidden_content():
    ap = _ap()
    # reveal_updates 只帶 truth_id/from/to——序列化後不該有任何「content」洩漏
    npcs = [{"name": "甲", "presence": "present"}]
    from core.models import DecisionPoint
    dp = DecisionPoint(situation_recap="x", decision_type="action",
                       suggested_options=[], free_input_hint="", beat_meta={"beat_number": 1})
    step = {"reveal_updates": [{"truth_id": "t.core", "from": "hidden", "to": "hinted"}]}
    obs = ap._dp_to_obs(_loop_with(npcs), "n", dp, False, None, step_result=step)
    blob = json.dumps(obs["debug"], ensure_ascii=False)
    assert "content" not in blob                       # debug 只露 truth_id
    assert "t.core" in blob                             # truth_id 可露


def test_ended_observation_no_options_with_debug():
    ap = _ap()
    from core.models import DecisionPoint, Option
    dp = DecisionPoint(situation_recap="結束。", decision_type="action",
                       suggested_options=[Option(text="不該出現", tone="bold")],
                       free_input_hint="x", beat_meta={"beat_number": 9})
    obs = ap._dp_to_obs(_loop_with([{"name": "甲", "presence": "present"}]),
                        "你逃出來了。", dp, True, {"type": "escape"}, step_result={})
    assert obs["ended"] is True and obs["options"] == []
    assert "debug" in obs
