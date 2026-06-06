"""Reveal Reward Loop（補丁）測試。

問題：gate=True 的 truth_investigation 常停在 hinted，不升 observed/suspected → 真相型無回報。
本補丁：在「gate 已放行的 truth_investigation」beat 上對已 hinted/observed 真相給 reward
（hinted→observed→suspected；**單靠 reward 到不了 confirmed**），並一律記 reveal_reward_debug。
**不碰 TruthEvidenceGate 的擋/放邏輯。**

必測（依規格）：
  - truth_investigation 可 hinted→observed
  - 重複 truth_investigation 可 observed→suspected
  - 無 mapped truth → no_progress_reason
  - world_navigation 仍不推 reveal
  - review_mode 仍不推 reveal
  - npc_fact_query 仍不推 reveal
  - 結構化 evidence_events 仍可推進（含 confirmed）
"""
from __future__ import annotations

from core.narrative.models import REVEAL_RANK
from core.narrative.revelation import (
    EvidenceEvent,
    RevealLedger,
    RevelationBridge,
    apply_reveal_reward,
    reward_candidates,
)


def _ledger(level: str, strength: float, tid: str = "t1"):
    led = RevealLedger()
    t = led.get_or_create(tid, title="某真相")
    t.level = level
    t.strength = strength
    return led, t


# ── 核心 reward（revelation.py）──────────────────────────────────────────────

def test_reward_moves_hinted_to_observed():
    led, t = _ledger("hinted", 0.3)
    update, tid, prev, nxt = apply_reveal_reward(led, beat=1)
    assert prev == "hinted" and nxt == "observed"
    assert update and update["to"] == "observed" and tid == "t1"


def test_repeated_reward_moves_observed_to_suspected():
    led, t = _ledger("hinted", 0.3)
    apply_reveal_reward(led, beat=1)               # hinted → observed
    assert t.level == "observed"
    update, _, prev, nxt = apply_reveal_reward(led, beat=2)
    assert prev == "observed" and nxt == "suspected"


def test_reward_alone_never_reaches_confirmed():
    """confirmed 不可只靠 reward——max_level=suspected。"""
    led, t = _ledger("hinted", 0.3)
    for b in range(12):
        apply_reveal_reward(led, beat=b)
    assert REVEAL_RANK[t.level] <= REVEAL_RANK["suspected"]
    assert t.level != "confirmed"


def test_no_candidate_when_all_hidden():
    led = RevealLedger()
    led.get_or_create("h1")                          # 仍 hidden
    assert reward_candidates(led) == []
    update, tid, prev, nxt = apply_reveal_reward(led, beat=1)
    assert update is None and tid is None


def test_no_candidate_when_all_suspected():
    led, t = _ledger("suspected", 1.2)
    assert reward_candidates(led) == []


def test_structured_evidence_can_confirm():
    """規格：結構化 evidence_events 仍可推進（含 confirmed），不受 reward 上限約束。"""
    led, t = _ledger("suspected", 1.0)
    ev = EvidenceEvent(evidence_id="e.kernel", source="kernel", truth_id="t1",
                       evidence_strength=0.6, max_level="actionable")
    RevelationBridge().apply(led, [ev])
    assert REVEAL_RANK[t.level] >= REVEAL_RANK["confirmed"]


# ── 迴圈整合（_reveal_reward_tick 的 gating）─────────────────────────────────

def _loop_with_hinted_truth():
    from core.blackboard import Blackboard
    from core.orchestrator_loop import BeatLoop
    from core.persistence.db import Database
    from core.signal import SignalBus
    from tests.test_narrative_v2_integration_nr import FakeCaller
    loop = BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(),
                    run_id="r", use_kernel=True)
    led, t = _ledger("hinted", 0.3)
    loop._reveal_ledger = led
    loop.beat_number = 1
    loop._reveal_updates_this_beat = []
    loop._reveal_reward_debug = {}
    loop._reveal_gate_reason = ""
    loop.bb.write("setup", "revealed_bible",
                  {"revealed_fragments": [], "truth_progress": {}})
    return loop, led, t


def test_loop_truth_investigation_rewards_hinted_to_observed():
    loop, led, t = _loop_with_hinted_truth()
    loop._action_class = "truth_investigation"
    loop._reveal_gate_allowed = True
    loop._reveal_reward_tick()
    d = loop._reveal_reward_debug
    assert t.level == "observed"
    assert d["ladder_action"] == "advanced_by_reward"
    assert d["gate_allowed"] is True
    assert d["previous_level"] == "hinted" and d["next_level"] == "observed"
    assert d["mapped_truth_ids"] == ["t1"]


def test_loop_world_navigation_does_not_push_reveal():
    loop, led, t = _loop_with_hinted_truth()
    loop._action_class = "world_navigation"
    loop._reveal_gate_allowed = True                 # 即使 gate 放行，class 不對也不推
    loop._reveal_reward_tick()
    assert t.level == "hinted"
    assert loop._reveal_reward_debug["ladder_action"] == "not_truth_investigation"


def test_loop_npc_fact_query_does_not_push_reveal():
    loop, led, t = _loop_with_hinted_truth()
    loop._action_class = "npc_fact_query"
    loop._reveal_gate_allowed = True
    loop._reveal_reward_tick()
    assert t.level == "hinted"
    assert loop._reveal_reward_debug["ladder_action"] == "not_truth_investigation"


def test_loop_review_mode_does_not_push_reveal():
    loop, led, t = _loop_with_hinted_truth()
    loop._action_class = "truth_investigation"
    loop._reveal_gate_allowed = False                # review_mode/no_truth → gate 擋
    loop._reveal_gate_reason = "review_mode"
    loop._reveal_reward_tick()
    assert t.level == "hinted"
    d = loop._reveal_reward_debug
    assert d["ladder_action"] == "gate_blocked"
    assert d["no_progress_reason"] == "review_mode"


def test_loop_no_progress_reason_when_capped():
    loop, led, t = _loop_with_hinted_truth()
    t.level = "suspected"
    t.strength = 1.2
    loop._action_class = "truth_investigation"
    loop._reveal_gate_allowed = True
    loop._reveal_reward_tick()
    d = loop._reveal_reward_debug
    assert d["ladder_action"] == "no_candidate"
    assert d["no_progress_reason"] == "all_candidates_capped"


def test_loop_no_progress_reason_when_no_hint_yet():
    loop, led, t = _loop_with_hinted_truth()
    t.level = "hidden"
    t.strength = 0.0
    loop._action_class = "truth_investigation"
    loop._reveal_gate_allowed = True
    loop._reveal_reward_tick()
    d = loop._reveal_reward_debug
    assert d["ladder_action"] == "no_candidate"
    assert d["no_progress_reason"] == "no_hinted_truth_yet"


def test_loop_advanced_by_evidence_when_prior_update():
    loop, led, t = _loop_with_hinted_truth()
    loop._action_class = "truth_investigation"
    loop._reveal_gate_allowed = True
    loop._reveal_updates_this_beat = [{"truth_id": "t1", "from": "hinted", "to": "observed"}]
    t.level = "observed"
    loop._reveal_reward_tick()
    d = loop._reveal_reward_debug
    assert d["ladder_action"] == "advanced_by_evidence"
    assert d["next_level"] == "observed"
