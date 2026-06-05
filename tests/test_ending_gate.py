"""NC6 — Ending Gate 驗收測試。

驗收：缺 escape action/出口/威脅未解 → 不給 clean_escape；0/8 碎片 → ambiguous；三條件齊全 + 有碎片 → clean。
"""
from __future__ import annotations

from core.narrative.ending_gate import (
    EndingGate, EndingState, build_ending_state, gate_escape_quality,
)
from core.blackboard import Blackboard


GATE = EndingGate()


def _full_state(truth=2):
    return EndingState(explicit_escape_action=True, exit_location_reached=True,
                       threat_avoided_or_resolved=True, truth_fragments_confirmed=truth,
                       total_beats=10)


# ── 因果門檻 ─────────────────────────────────────────────────────────────────
def test_clean_escape_needs_all_causality():
    assert GATE.check("clean_escape", _full_state()).ending_type == "clean_escape"

    s = _full_state(); s.explicit_escape_action = False
    assert GATE.check("clean_escape", s).ending_type == "ambiguous_escape"

    s = _full_state(); s.exit_location_reached = False
    assert GATE.check("clean_escape", s).reason == "exit_location_not_reached"

    s = _full_state(); s.threat_avoided_or_resolved = False
    assert GATE.check("clean_escape", s).reason == "threat_not_resolved"


def test_zero_truth_cannot_clean_escape():
    s = _full_state(truth=0)                                # 因果齊全但 0/8 碎片
    d = GATE.check("clean_escape", s)
    assert d.ending_type == "ambiguous_escape"
    assert "zero_truth_fragments" in d.reason


# ── build_ending_state（啟發式）────────────────────────────────────────────
class _GS:
    def __init__(self, scene="ward", danger=0):
        self.current_scene = scene; self.danger_level = danger


def _bb(revealed_ids=()):
    bb = Blackboard()
    bb.write("orchestrator", "revealed_bible",
             {"revealed_fragments": [{"id": i} for i in revealed_ids]})
    return bb


def test_build_state_detects_escape_action_and_exit():
    st = build_ending_state(_bb(("f1",)), _GS(scene="exit_gate", danger=0),
                            last_player_decision="我衝向大門逃出去", beat_number=8)
    assert st.explicit_escape_action and st.exit_location_reached
    assert st.threat_avoided_or_resolved and st.truth_fragments_confirmed == 1


# ── gate_escape_quality：0/8 → ambiguous；齊全 → clean ──────────────────────
def test_gate_quality_zero_truth_ambiguous():
    q, reason = gate_escape_quality(_bb(()), _GS("exit_gate", 0), "我逃出去", 9)
    assert q == "ambiguous" and "zero_truth" in reason


def test_gate_quality_full_clean():
    q, _ = gate_escape_quality(_bb(("f1", "f2")), _GS("停車場", 0), "我衝出大門逃離", 12)
    assert q == "clean"


def test_gate_quality_no_escape_action_ambiguous():
    q, _ = gate_escape_quality(_bb(("f1",)), _GS("ward", 0), "我繼續觀察", 5)
    assert q == "ambiguous"
