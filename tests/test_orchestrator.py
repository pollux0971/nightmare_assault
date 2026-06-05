"""tests/test_orchestrator.py — 揭露閘門測試（U10）。

覆蓋：
- min_beats 條件（beat=2 不揭露、beat=3 揭露）
- location_reached 條件（未到不揭露、到了才揭露）
- requires_touched 條件（不含→不揭露、含→揭露）
- 多條件 AND（缺一不揭露）
- 揭露後 blackboard 狀態正確（revealed_bible / turn_context.newly_revealed）
- real_bible 不變
- requires_semantic + 假 caller（揭露 / 不揭露 / 無 caller）
"""
from __future__ import annotations

import pytest
from core.blackboard import Blackboard
from core.agents.orchestrator import run_orchestrator
from core.models import OrchestratorOutput, FragmentReveal


# ─────────────────────────────────────────────────────────────────────────────
# 假 Caller
# ─────────────────────────────────────────────────────────────────────────────

class _FakeCaller:
    """可配置回傳 OrchestratorOutput 的假 caller。"""

    def __init__(self, reveal_ids: list[str] | None = None):
        """reveal_ids：這些碎片 id 會出現在 fragments_to_reveal 裡（視為「揭露」）。"""
        self._reveal_ids: set[str] = set(reveal_ids or [])

    def call(self, agent: str, context: dict, output_model=None, temperature=None):
        frag_id = context.get("fragment", {}).get("id", "")
        fragments = [FragmentReveal(id=frag_id, how_to_reveal="測試揭露")] if frag_id in self._reveal_ids else []
        return OrchestratorOutput(fragments_to_reveal=fragments, reasoning="fake")


# ─────────────────────────────────────────────────────────────────────────────
# Fixture 工廠
# ─────────────────────────────────────────────────────────────────────────────

def _make_bb(revelations: list[dict]) -> Blackboard:
    """建立含指定 revelation_pool 的 Blackboard。"""
    bb = Blackboard()
    bb.write("setup", "real_bible", {"revelation_pool": revelations})
    return bb


# ─────────────────────────────────────────────────────────────────────────────
# min_beats 條件
# ─────────────────────────────────────────────────────────────────────────────

class TestMinBeats:
    def test_not_revealed_below_threshold(self):
        bb = _make_bb([
            {"id": "frag_min3", "type": "knowledge", "content": "秘密A",
             "reveal_condition": {"min_beats": 3}},
        ])
        result = run_orchestrator(bb, beat_number=2)
        assert result == []

    def test_revealed_at_threshold(self):
        bb = _make_bb([
            {"id": "frag_min3", "type": "knowledge", "content": "秘密A",
             "reveal_condition": {"min_beats": 3}},
        ])
        result = run_orchestrator(bb, beat_number=3)
        assert len(result) == 1
        assert result[0]["id"] == "frag_min3"

    def test_revealed_above_threshold(self):
        bb = _make_bb([
            {"id": "frag_min3", "type": "knowledge", "content": "秘密A",
             "reveal_condition": {"min_beats": 3}},
        ])
        result = run_orchestrator(bb, beat_number=5)
        assert len(result) == 1


# ─────────────────────────────────────────────────────────────────────────────
# location_reached 條件
# ─────────────────────────────────────────────────────────────────────────────

class TestLocationReached:
    def test_not_revealed_without_location(self):
        bb = _make_bb([
            {"id": "frag_roof", "type": "item", "content": "頂樓線索",
             "reveal_condition": {"location_reached": "頂樓"}},
        ])
        result = run_orchestrator(bb, beat_number=1)
        assert result == []

    def test_not_revealed_wrong_location(self):
        bb = _make_bb([
            {"id": "frag_roof", "type": "item", "content": "頂樓線索",
             "reveal_condition": {"location_reached": "頂樓"}},
        ])
        result = run_orchestrator(bb, beat_number=1, reached_locations=["地下室"])
        assert result == []

    def test_revealed_when_location_reached(self):
        bb = _make_bb([
            {"id": "frag_roof", "type": "item", "content": "頂樓線索",
             "reveal_condition": {"location_reached": "頂樓"}},
        ])
        result = run_orchestrator(bb, beat_number=1, reached_locations=["頂樓"])
        assert len(result) == 1
        assert result[0]["id"] == "frag_roof"


# ─────────────────────────────────────────────────────────────────────────────
# requires_touched 條件
# ─────────────────────────────────────────────────────────────────────────────

class TestRequiresTouched:
    def test_not_revealed_without_required_touched(self):
        bb = _make_bb([
            {"id": "frag_b", "type": "knowledge", "content": "需先觸及A",
             "reveal_condition": {"requires_touched": ["frag_a"]}},
        ])
        result = run_orchestrator(bb, beat_number=1, touched_ids=[])
        assert result == []

    def test_not_revealed_partial_touched(self):
        bb = _make_bb([
            {"id": "frag_c", "type": "knowledge", "content": "需先觸及A和B",
             "reveal_condition": {"requires_touched": ["frag_a", "frag_b"]}},
        ])
        result = run_orchestrator(bb, beat_number=1, touched_ids=["frag_a"])
        assert result == []

    def test_revealed_when_all_touched(self):
        bb = _make_bb([
            {"id": "frag_b", "type": "knowledge", "content": "需先觸及A",
             "reveal_condition": {"requires_touched": ["frag_a"]}},
        ])
        result = run_orchestrator(bb, beat_number=1, touched_ids=["frag_a"])
        assert len(result) == 1
        assert result[0]["id"] == "frag_b"

    def test_revealed_with_multiple_touched(self):
        bb = _make_bb([
            {"id": "frag_c", "type": "knowledge", "content": "需先觸及A和B",
             "reveal_condition": {"requires_touched": ["frag_a", "frag_b"]}},
        ])
        result = run_orchestrator(bb, beat_number=1, touched_ids=["frag_a", "frag_b", "extra"])
        assert len(result) == 1
        assert result[0]["id"] == "frag_c"


# ─────────────────────────────────────────────────────────────────────────────
# 多條件 AND
# ─────────────────────────────────────────────────────────────────────────────

class TestMultipleConditionsAnd:
    FRAG = {
        "id": "frag_multi",
        "type": "knowledge",
        "content": "多條件",
        "reveal_condition": {
            "min_beats": 3,
            "location_reached": "頂樓",
            "requires_touched": ["frag_a"],
        },
    }

    def test_all_missing_not_revealed(self):
        bb = _make_bb([self.FRAG])
        result = run_orchestrator(bb, beat_number=1)
        assert result == []

    def test_only_min_beats_met_not_revealed(self):
        bb = _make_bb([self.FRAG])
        result = run_orchestrator(bb, beat_number=3)
        assert result == []

    def test_min_beats_and_location_met_not_revealed(self):
        bb = _make_bb([self.FRAG])
        result = run_orchestrator(bb, beat_number=3, reached_locations=["頂樓"])
        assert result == []

    def test_all_conditions_met_revealed(self):
        bb = _make_bb([self.FRAG])
        result = run_orchestrator(
            bb, beat_number=3,
            reached_locations=["頂樓"],
            touched_ids=["frag_a"],
        )
        assert len(result) == 1
        assert result[0]["id"] == "frag_multi"


# ─────────────────────────────────────────────────────────────────────────────
# 揭露後 Blackboard 狀態
# ─────────────────────────────────────────────────────────────────────────────

class TestBlackboardState:
    def test_revealed_bible_contains_fragment(self):
        bb = _make_bb([
            {"id": "frag_x", "type": "knowledge", "content": "秘密X",
             "reveal_condition": {"min_beats": 1}},
        ])
        run_orchestrator(bb, beat_number=1)
        revealed = bb.revealed_bible.get("revealed_fragments", [])
        ids = [f["id"] for f in revealed]
        assert "frag_x" in ids

    def test_turn_context_newly_revealed_set(self):
        bb = _make_bb([
            {"id": "frag_y", "type": "item", "content": "物品Y",
             "reveal_condition": {"min_beats": 1}},
        ])
        run_orchestrator(bb, beat_number=1)
        newly = bb.turn_context.get("newly_revealed", [])
        ids = [f["id"] for f in newly]
        assert "frag_y" in ids

    def test_real_bible_unchanged(self):
        pool = [
            {"id": "frag_z", "type": "person", "content": "人物Z",
             "reveal_condition": {"min_beats": 1}},
        ]
        bb = _make_bb(pool)
        original_real = bb.real_bible.copy()
        run_orchestrator(bb, beat_number=1)
        # real_bible 不應變動
        assert bb.real_bible == original_real
        assert "revelation_pool" in bb.real_bible
        assert len(bb.real_bible["revelation_pool"]) == 1

    def test_no_reveal_turn_context_not_set(self):
        bb = _make_bb([
            {"id": "frag_no", "type": "knowledge", "content": "不會揭露",
             "reveal_condition": {"min_beats": 99}},
        ])
        run_orchestrator(bb, beat_number=1)
        # 無揭露時 newly_revealed 不應被寫入（仍為空 / 不存在）
        assert bb.turn_context.get("newly_revealed") is None

    def test_already_revealed_not_double_added(self):
        bb = _make_bb([
            {"id": "frag_once", "type": "knowledge", "content": "只揭露一次",
             "reveal_condition": {"min_beats": 1}},
        ])
        run_orchestrator(bb, beat_number=1)
        run_orchestrator(bb, beat_number=2)  # 再跑一次
        revealed = bb.revealed_bible.get("revealed_fragments", [])
        ids = [f["id"] for f in revealed]
        assert ids.count("frag_once") == 1

    def test_multiple_fragments_reveal_all_eligible(self):
        bb = _make_bb([
            {"id": "frag_1", "type": "knowledge", "content": "一",
             "reveal_condition": {"min_beats": 1}},
            {"id": "frag_2", "type": "item", "content": "二",
             "reveal_condition": {"min_beats": 2}},
        ])
        result = run_orchestrator(bb, beat_number=2)
        ids = [f["id"] for f in result]
        assert "frag_1" in ids
        assert "frag_2" in ids
        assert len(ids) == 2


# ─────────────────────────────────────────────────────────────────────────────
# requires_semantic
# ─────────────────────────────────────────────────────────────────────────────

class TestRequiresSemantic:
    def test_semantic_no_caller_not_revealed(self):
        """無 caller 時 requires_semantic 視為未滿足→不揭露。"""
        bb = _make_bb([
            {"id": "frag_sem", "type": "knowledge", "content": "語義判定",
             "reveal_condition": {"requires_semantic": "玩家是否實質詢問了秘密"}},
        ])
        result = run_orchestrator(bb, beat_number=1, caller=None)
        assert result == []

    def test_semantic_caller_says_reveal(self):
        """caller 回「揭露」→ 揭露。"""
        bb = _make_bb([
            {"id": "frag_sem", "type": "knowledge", "content": "語義判定",
             "reveal_condition": {"requires_semantic": "玩家是否實質詢問了秘密"}},
        ])
        caller = _FakeCaller(reveal_ids=["frag_sem"])
        result = run_orchestrator(bb, beat_number=1, caller=caller)
        assert len(result) == 1
        assert result[0]["id"] == "frag_sem"

    def test_semantic_caller_says_no_reveal(self):
        """caller 回「不揭露」（fragments_to_reveal 空）→ 不揭露。"""
        bb = _make_bb([
            {"id": "frag_sem2", "type": "knowledge", "content": "語義判定2",
             "reveal_condition": {"requires_semantic": "玩家是否實質詢問了秘密"}},
        ])
        caller = _FakeCaller(reveal_ids=[])  # 不揭露任何碎片
        result = run_orchestrator(bb, beat_number=1, caller=caller)
        assert result == []

    def test_semantic_combined_with_min_beats_both_met(self):
        """requires_semantic AND min_beats，兩個都滿足才揭露。"""
        bb = _make_bb([
            {"id": "frag_combined", "type": "knowledge", "content": "複合",
             "reveal_condition": {
                 "min_beats": 2,
                 "requires_semantic": "玩家是否深入探索",
             }},
        ])
        caller = _FakeCaller(reveal_ids=["frag_combined"])
        result = run_orchestrator(bb, beat_number=2, caller=caller)
        assert len(result) == 1

    def test_semantic_combined_min_beats_not_met(self):
        """requires_semantic AND min_beats，min_beats 未滿足→不揭露（不呼 LLM）。"""
        bb = _make_bb([
            {"id": "frag_combined2", "type": "knowledge", "content": "複合2",
             "reveal_condition": {
                 "min_beats": 5,
                 "requires_semantic": "玩家是否深入探索",
             }},
        ])
        caller = _FakeCaller(reveal_ids=["frag_combined2"])
        result = run_orchestrator(bb, beat_number=2, caller=caller)
        assert result == []

    def test_semantic_revealed_bible_and_turn_context_updated(self):
        """語義揭露後 blackboard 狀態正確。"""
        bb = _make_bb([
            {"id": "frag_sem3", "type": "person", "content": "語義人物",
             "reveal_condition": {"requires_semantic": "深入觸及"}},
        ])
        caller = _FakeCaller(reveal_ids=["frag_sem3"])
        run_orchestrator(bb, beat_number=1, caller=caller)

        revealed = bb.revealed_bible.get("revealed_fragments", [])
        ids = [f["id"] for f in revealed]
        assert "frag_sem3" in ids

        newly = bb.turn_context.get("newly_revealed", [])
        assert any(f["id"] == "frag_sem3" for f in newly)


# ─────────────────────────────────────────────────────────────────────────────
# 無條件碎片（空 reveal_condition → 立即揭露）
# ─────────────────────────────────────────────────────────────────────────────

class TestNoCondition:
    def test_empty_condition_reveals_immediately(self):
        bb = _make_bb([
            {"id": "frag_free", "type": "knowledge", "content": "無條件",
             "reveal_condition": {}},
        ])
        result = run_orchestrator(bb, beat_number=1)
        assert len(result) == 1
        assert result[0]["id"] == "frag_free"


# ─────────────────────────────────────────────────────────────────────────────
# 空 real_bible / 空 revelation_pool
# ─────────────────────────────────────────────────────────────────────────────

class TestEdgeCases:
    def test_empty_real_bible_returns_empty(self):
        bb = Blackboard()
        # real_bible 完全空
        result = run_orchestrator(bb, beat_number=1)
        assert result == []

    def test_empty_revelation_pool_returns_empty(self):
        bb = Blackboard()
        bb.write("setup", "real_bible", {"revelation_pool": []})
        result = run_orchestrator(bb, beat_number=1)
        assert result == []

    def test_revelations_key_also_supported(self):
        """real_bible 用 revelations 鍵也應支援。"""
        bb = Blackboard()
        bb.write("setup", "real_bible", {
            "revelations": [
                {"id": "frag_alt", "type": "knowledge", "content": "替代鍵",
                 "reveal_condition": {}},
            ]
        })
        result = run_orchestrator(bb, beat_number=1)
        assert len(result) == 1
        assert result[0]["id"] == "frag_alt"
