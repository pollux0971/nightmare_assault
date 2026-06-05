"""HA2 — Death Causality Guard 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：danger 達標但無死因 → death_physical 被擋、downgrade_to=danger_warning；
有 warden hard_trigger / death_cause_event / 明確致命行為 → 允許；escape commit/ambiguous 邏輯。
"""
from __future__ import annotations

from core.narrative.ending_causality import EndingCausalityGate, EndingCandidate, GateResult


class _W:
    def __init__(self, ending_triggered=None, hard_trigger=None):
        self.ending_triggered = ending_triggered
        self.hard_trigger = hard_trigger


class _P:
    def __init__(self, death_cause_event=None, explicit_lethal_action=False):
        self.death_cause_event = death_cause_event
        self.explicit_lethal_action = explicit_lethal_action


# ── danger-only 不得致死 ─────────────────────────────────────────────────────
def test_danger_alone_cannot_trigger_death():
    g = EndingCausalityGate()
    cand = EndingCandidate(type="death_physical", source="attractor", confidence=0.9)
    res = g.check_death(cand, warden_result=_W(), progress_result=_P(), state=None)
    assert not res.allowed
    assert res.downgrade_to == "danger_warning"


# ── 有明確死因 → 允許 ────────────────────────────────────────────────────────
def test_death_with_explicit_cause_allowed():
    g = EndingCausalityGate()
    cand = EndingCandidate(type="death_physical", source="warden")
    # warden 硬觸發死亡
    assert g.check_death(cand, warden_result=_W(ending_triggered="death_physical"),
                         progress_result=_P()).allowed
    # kernel 死因事件
    assert g.check_death(cand, warden_result=_W(),
                         progress_result=_P(death_cause_event="event.drowned")).allowed
    # 明確致命行為
    assert g.check_death(cand, warden_result=_W(),
                         progress_result=_P(explicit_lethal_action=True)).allowed


# ── 非死亡結局不受此閘影響 ───────────────────────────────────────────────────
def test_non_death_passthrough():
    g = EndingCausalityGate()
    assert g.check_death(EndingCandidate(type="escape", source="attractor"),
                         warden_result=_W(), progress_result=_P()).allowed


# ── escape commit / ambiguous ───────────────────────────────────────────────
def test_escape_requires_commit_and_ambiguous():
    g = EndingCausalityGate()
    cand = EndingCandidate(type="escape", source="attractor")
    assert g.check_escape(cand, escape_step="await_commit").downgrade_to == "exit_candidate"
    r = g.check_escape(cand, escape_step="commit", reveal_progress={"confirmed": 0})
    assert r.allowed and r.downgrade_to == "ambiguous_escape"
    assert g.check_escape(cand, escape_step="commit", reveal_progress={"confirmed": 2}).downgrade_to is None


# ── loop 整合：flag ON 時 danger 超標不直接 death（用整合 harness）────────────
def test_loop_danger_does_not_kill(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    # 人為把危險拉爆，然後推進——不得單因 danger 觸發 death 結局
    loop._game_state.danger_level = 999
    out = loop.step("我站在原地不動")
    if out.get("ended"):
        assert out["ending"].get("type") not in ("death_physical", "death_mental", "death"), \
            "danger-only 不該致死"
