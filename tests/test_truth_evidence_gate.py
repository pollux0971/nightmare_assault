"""WorldConsequence vs TruthEvidence Split —— 只有 truth_investigation / 合法 structured
evidence 才能推 reveal；找路/整理/引用 NPC fact/一般檢查 一律不推。"""
from __future__ import annotations

import core.constants as C
from core.narrative.action_intent import (
    classify_action, no_truth_intent,
    WORLD_NAVIGATION, WORLD_REVIEW, OBJECT_INSPECTION, TRUTH_INVESTIGATION,
    NPC_FACT_QUERY, CAMPAIGN_END, UNKNOWN,
)
from core.narrative.truth_evidence_gate import TruthEvidenceGate
from core.narrative.exit_resolver import resolve_exit_affordance
from core.world.model import FACT


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def _levels(loop):
    return {tid: t.level for tid, t in (getattr(loop._reveal_ledger, "truths", {}) or {}).items()}


# ── ActionIntent 分類 ────────────────────────────────────────────────────────
def test_classify_action():
    cl = lambda t: classify_action(t, exit_affordance=resolve_exit_affordance(t))
    assert cl("我研究實驗紀錄與異常頻率") == TRUTH_INVESTIGATION
    assert cl("我分析監控錄音的數據") == TRUTH_INVESTIGATION
    assert cl("只移動到B2通訊室方向") == WORLD_NAVIGATION
    assert cl("我往走廊深處走") == WORLD_NAVIGATION
    assert cl("根據他說的機房線索找路") == NPC_FACT_QUERY
    assert cl("在安全區整理已知資訊") == WORLD_REVIEW
    assert cl("我檢查桌上的東西") == OBJECT_INSPECTION
    assert cl("我結束本次調查，接受結果") == CAMPAIGN_END
    assert cl("嗯。") == UNKNOWN
    # 「不碰真相」即使含「真相」也不得判 truth_investigation
    assert cl("我找路，但不碰真相") != TRUTH_INVESTIGATION


def test_no_truth_intent():
    assert no_truth_intent("根據他說的機房線索找路，不碰真相")
    assert no_truth_intent("只整理已知，不新增調查")
    assert no_truth_intent("我只想找路")
    assert not no_truth_intent("我研究實驗紀錄")


# ── TruthEvidenceGate.evaluate ───────────────────────────────────────────────
def test_gate_allows_only_truth_or_structured():
    g = TruthEvidenceGate()
    assert g.evaluate(TRUTH_INVESTIGATION)[0] is True
    assert g.evaluate("anything", has_structured_evidence=True)[0] is True
    assert g.evaluate(OBJECT_INSPECTION, truth_bearing=True)[0] is True
    for ac in (WORLD_NAVIGATION, WORLD_REVIEW, NPC_FACT_QUERY, OBJECT_INSPECTION, UNKNOWN, CAMPAIGN_END):
        allowed, reason = g.evaluate(ac)
        assert allowed is False and reason


def test_gate_blocks_no_truth_and_review():
    g = TruthEvidenceGate()
    # 明確不碰真相 → block（即使 action 是 truth_investigation）
    assert g.evaluate(TRUTH_INVESTIGATION, no_truth=True) == (False, "explicit_no_truth_intent")
    # review/retreat 模式 → block
    assert g.evaluate(TRUTH_INVESTIGATION, exploration_mode="review_mode")[0] is False
    assert g.evaluate(TRUTH_INVESTIGATION, exploration_mode="temporary_retreat")[0] is False


# ── _revelation_tick：閘門擋下 vs 放行（注入 truth-mapped clue，確定性）────────
def test_revelation_tick_blocks_non_truth_action(monkeypatch):
    loop = _started_loop(monkeypatch)
    tid = next(t for t in loop._reveal_ledger.truths if t not in loop._core_truth_ids)
    before = loop._reveal_ledger.truths[tid].level
    loop._game_state.clues["clue.nav"] = {"truth_id": tid, "evidence_strength": 0.9,
                                          "max_level": "observed"}
    loop._action_class = WORLD_NAVIGATION
    loop._exploration_mode = "active_exploration"
    loop._no_truth_intent = False
    loop._reveal_updates_this_beat = []; loop._evidence_events_this_beat = 0
    loop._unmapped_evidence_this_beat = 0; loop._blocked_reveal_candidates = []
    changed = loop._revelation_tick(loop._game_state)
    assert changed is False
    assert loop._reveal_ledger.truths[tid].level == before     # 未推進
    assert loop._reveal_gate_allowed is False
    assert "clue.nav" in loop._blocked_reveal_candidates       # 記為 blocked（debug）


def test_revelation_tick_allows_truth_investigation(monkeypatch):
    loop = _started_loop(monkeypatch)
    tid = next(t for t in loop._reveal_ledger.truths if t not in loop._core_truth_ids)
    before = loop._reveal_ledger.truths[tid].level
    loop._game_state.clues["clue.study"] = {"truth_id": tid, "evidence_strength": 0.9,
                                            "max_level": "observed"}
    loop._action_class = TRUTH_INVESTIGATION
    loop._exploration_mode = "active_exploration"
    loop._no_truth_intent = False
    loop._reveal_updates_this_beat = []; loop._evidence_events_this_beat = 0
    loop._unmapped_evidence_this_beat = 0; loop._blocked_reveal_candidates = []
    changed = loop._revelation_tick(loop._game_state)
    assert changed is True
    assert loop._reveal_ledger.truths[tid].level != before     # 真相調查 → 推進
    assert loop._reveal_gate_allowed is True


# ── loop.step 整合：open-exploration 行為不推 reveal ──────────────────────────
def test_navigation_with_npc_fact_does_not_push_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    before = _levels(loop)
    out = loop.step("我根據他說的機房線索找路，不碰真相")
    assert _levels(loop) == before                             # reveal 不動
    assert out["reveal_gate_allowed"] is False
    assert out["no_truth_intent"] is True


def test_pure_navigation_does_not_push_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    before = _levels(loop)
    out = loop.step("只移動到B2通訊室方向")
    assert _levels(loop) == before
    assert out["reveal_gate_allowed"] is False
    assert out["action_class"] == WORLD_NAVIGATION


def test_review_does_not_push_reveal_nor_add_unaccounted_fact(monkeypatch):
    loop = _started_loop(monkeypatch)
    before = _levels(loop)
    facts_before = {e.id for e in loop._world.by_kind(FACT)}
    out = loop.step("在安全區整理已知資訊，不新增調查")
    assert _levels(loop) == before                             # reveal 不動
    assert out["reveal_gate_allowed"] is False
    # review 模式不新增未記帳 fact entity
    assert {e.id for e in loop._world.by_kind(FACT)} == facts_before


def test_npc_prose_fact_writes_worldmodel_but_no_reveal(monkeypatch):
    from core.narrative.npc_chat_control import NPCChatResponse
    loop = _started_loop(monkeypatch)
    before = _levels(loop)
    loop.bridge_npc_evidence(
        NPCChatResponse(visible_reply="通訊設備在B2機房。", entity_delta=[]), npc_id="npc.王哲")
    assert loop._world.find("通訊設備在B2機房", kind=FACT) is not None   # 進 WorldModel
    assert _levels(loop) == before                                       # 不推 reveal


def test_observation_debug_has_split_fields(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("只移動到B2通訊室方向")
    for k in ("action_class", "no_truth_intent", "reveal_gate_allowed",
              "reveal_gate_block_reason", "blocked_reveal_candidates"):
        assert k in out, f"step 缺 debug 欄位 {k}"
