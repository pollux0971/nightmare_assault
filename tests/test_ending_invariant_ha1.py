"""HA1 — Ending Observation Invariant 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：ended=true ⇒ decision_point 無 options、free_input_hint=None；agent_play observation
ended 時 options=[]；regression 斷言 not (ended and options)。
"""
from __future__ import annotations

from core.orchestrator_loop import BeatLoop
from core.models import DecisionPoint, Option


def _dp(opts=("往前走", "後退")):
    return DecisionPoint(
        situation_recap="走廊盡頭。", decision_type="action",
        suggested_options=[Option(text=t, tone="cautious") for t in opts],
        free_input_hint="或自由輸入", beat_meta={"beat_number": 5})


# ── loop 不變式：ended ⇒ dp 無 options ──────────────────────────────────────
def test_loop_enforces_ended_invariant():
    loop = BeatLoop.__new__(BeatLoop)
    loop.ended = True
    dp = _dp()
    out = loop._enforce_ended_invariant(dp)
    assert out.suggested_options == []
    assert out.free_input_hint == ""                       # str 型別 → "" 而非 None
    assert not (loop.ended and out.suggested_options)     # regression 不變式


def test_loop_keeps_options_when_not_ended():
    loop = BeatLoop.__new__(BeatLoop)
    loop.ended = False
    dp = _dp()
    out = loop._enforce_ended_invariant(dp)
    assert len(out.suggested_options) == 2                 # 未結局 → 保留選項


# ── agent_play observation 層：ended ⇒ options=[] ───────────────────────────
def test_agent_play_observation_no_options_when_ended():
    import sys, types
    # 直接測 _dp_to_obs（不需真 loop；用最小 stub）
    sys.path.insert(0, "dev/tools")
    import importlib.util
    spec = importlib.util.spec_from_file_location("agent_play", "dev/tools/agent_play.py")
    ap = importlib.util.module_from_spec(spec); spec.loader.exec_module(ap)

    class _Loop:
        beat_number = 11
        class bb:
            @staticmethod
            def snapshot(): return {"revealed_bible": {"truth_progress": {}}, "npc_registry": []}
    obs = ap._dp_to_obs(_Loop(), "你逃出來了。", _dp(), True, {"type": "escape"})
    assert obs["ended"] is True
    assert obs["options"] == []
    assert obs["free_input_hint"] is None
    assert not (obs["ended"] and obs["options"])
