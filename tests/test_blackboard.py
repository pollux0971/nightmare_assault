"""tests/test_blackboard.py — Blackboard 機制測試（U03）。

覆蓋：
- A4 過期 patch 丟棄（base_version 不符）
- 有效 patch merge 生效、version +1
- A5 隔離：submit_patch 後 snapshot 不含未 merge 的變更
- merge_and_bump 後 collect_pending() 為空
- 違規 patch 拋 PermissionError
- merge 摘要計數正確
"""
from __future__ import annotations

import copy
import pytest
from core.blackboard import Blackboard


# ─────────────────────────────────────────────────────────────────────────────
# Fixtures
# ─────────────────────────────────────────────────────────────────────────────

@pytest.fixture
def bb():
    return Blackboard()


# ─────────────────────────────────────────────────────────────────────────────
# 初始狀態
# ─────────────────────────────────────────────────────────────────────────────

class TestInitialState:
    def test_version_starts_at_zero(self, bb):
        assert bb.version == 0

    def test_beat_number_starts_at_zero(self, bb):
        assert bb.beat_number == 0

    def test_real_bible_empty(self, bb):
        assert bb.real_bible == {}

    def test_npc_registry_empty_list(self, bb):
        assert bb.npc_registry == []

    def test_rolling_summary_empty_string(self, bb):
        assert bb.rolling_summary == ""

    def test_recent_chat_digest_none(self, bb):
        assert bb.recent_chat_digest is None

    def test_pending_empty(self, bb):
        assert bb.collect_pending() == []


# ─────────────────────────────────────────────────────────────────────────────
# A5 隔離：submit_patch 後 snapshot 不含變更
# ─────────────────────────────────────────────────────────────────────────────

class TestSnapshotIsolation:
    """A5：pending patch 在 merge 前完全不出現在 snapshot 裡。"""

    def test_snapshot_does_not_contain_pending_patch(self, bb):
        # 提交一個更改 rolling_summary 的 patch
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "this should not appear yet",
        })
        snap = bb.snapshot()
        # snapshot 不應包含 patch 的值
        assert snap["rolling_summary"] == ""

    def test_snapshot_does_not_contain_pending_protagonist_patch(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "setup",
            "target": "protagonist",
            "value": {"name": "Hidden Hero"},
        })
        snap = bb.snapshot()
        assert snap["protagonist"] == {}

    def test_snapshot_after_merge_contains_patch(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "merged summary",
        })
        bb.merge_and_bump()
        snap = bb.snapshot()
        assert snap["rolling_summary"] == "merged summary"

    def test_snapshot_is_deep_copy_not_reference(self, bb):
        bb.write("setup", "protagonist", {"name": "Hero", "items": []})
        snap = bb.snapshot()
        # 修改 snapshot 不應影響 blackboard
        snap["protagonist"]["name"] = "Mutated"
        snap["protagonist"]["items"].append("sword")
        assert bb.protagonist["name"] == "Hero"
        assert bb.protagonist["items"] == []

    def test_multiple_snapshots_are_independent(self, bb):
        bb.write("setup", "rolling_summary", "original")
        snap1 = bb.snapshot()
        bb.write("compactor", "rolling_summary", "updated")
        snap2 = bb.snapshot()
        assert snap1["rolling_summary"] == "original"
        assert snap2["rolling_summary"] == "updated"


# ─────────────────────────────────────────────────────────────────────────────
# A4 過期 patch 測試
# ─────────────────────────────────────────────────────────────────────────────

class TestExpiredPatch:
    """A4：base_version 不符的 patch 必須被丟棄，不生效。"""

    def test_expired_patch_is_discarded(self, bb):
        # 先 bump 一次，version 變成 1
        bb.merge_and_bump()
        assert bb.version == 1

        # 提交一個 base_version=0 的 patch（過期）
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "stale patch",
        })
        result = bb.merge_and_bump()
        # 應被丟棄
        assert result["discarded"] == 1
        assert result["applied"] == 0
        # rolling_summary 不應改變
        assert bb.rolling_summary == ""

    def test_expired_patch_does_not_write_real_bible(self, bb):
        # version = 0
        # 先 bump 讓 version = 1
        bb.merge_and_bump()

        # 過期 patch（即使 writer=setup，base_version 不符也丟棄）
        bb.submit_patch({
            "base_version": 0,
            "writer": "setup",
            "target": "real_bible",
            "value": {"stale": True},
        })
        result = bb.merge_and_bump()
        assert result["discarded"] == 1
        assert bb.real_bible == {}

    def test_mixed_valid_and_expired_patches(self, bb):
        # version=0
        # 提交一個有效 patch 和一個過期 patch
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "valid",
        })
        # 先 merge 讓 version 變 1
        bb.merge_and_bump()
        assert bb.version == 1
        assert bb.rolling_summary == "valid"

        # 現在提交 base_version=0（過期）與 base_version=1（有效）
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "stale override",
        })
        bb.submit_patch({
            "base_version": 1,
            "writer": "compactor",
            "target": "recent_chat_digest",
            "value": "new digest",
        })
        result = bb.merge_and_bump()
        assert result["applied"] == 1
        assert result["discarded"] == 1
        assert bb.rolling_summary == "valid"  # stale patch 沒覆蓋
        assert bb.recent_chat_digest == "new digest"

    def test_future_version_patch_is_discarded(self, bb):
        # version=0，提交 base_version=5（未來版本）→ 也算不符 → 丟棄
        bb.submit_patch({
            "base_version": 5,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "future",
        })
        result = bb.merge_and_bump()
        assert result["discarded"] == 1
        assert bb.rolling_summary == ""


# ─────────────────────────────────────────────────────────────────────────────
# 有效 patch merge 測試
# ─────────────────────────────────────────────────────────────────────────────

class TestValidPatchMerge:
    """有效 patch（base_version==version）merge 後生效、version+1。"""

    def test_valid_patch_applies_and_bumps_version(self, bb):
        assert bb.version == 0
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "summary v1",
        })
        result = bb.merge_and_bump()
        assert result["applied"] == 1
        assert result["discarded"] == 0
        assert result["version_after"] == 1
        assert bb.version == 1
        assert bb.rolling_summary == "summary v1"

    def test_multiple_valid_patches_all_apply(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "summary",
        })
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "recent_chat_digest",
            "value": "digest",
        })
        result = bb.merge_and_bump()
        assert result["applied"] == 2
        assert bb.rolling_summary == "summary"
        assert bb.recent_chat_digest == "digest"
        assert bb.version == 1

    def test_version_increments_by_one_per_merge(self, bb):
        for i in range(3):
            bb.submit_patch({
                "base_version": i,
                "writer": "compactor",
                "target": "rolling_summary",
                "value": f"v{i+1}",
            })
            bb.merge_and_bump()
        assert bb.version == 3
        assert bb.rolling_summary == "v3"

    def test_setup_patch_writes_real_bible(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "setup",
            "target": "real_bible",
            "value": {"threat": "monster"},
        })
        bb.merge_and_bump()
        assert bb.real_bible == {"threat": "monster"}

    def test_ledger_patch_applies(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "warden",
            "target": "ledger",
            "value": [{"type": "death", "content": "player fell"}],
        })
        result = bb.merge_and_bump()
        assert result["applied"] == 1
        assert bb.ledger[0]["type"] == "death"


# ─────────────────────────────────────────────────────────────────────────────
# merge_and_bump 後 collect_pending() 為空
# ─────────────────────────────────────────────────────────────────────────────

class TestPendingClearedAfterMerge:
    def test_collect_pending_empty_after_merge(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "x",
        })
        assert len(bb.collect_pending()) == 1
        bb.merge_and_bump()
        assert bb.collect_pending() == []

    def test_collect_pending_empty_after_merge_with_expired(self, bb):
        bb.merge_and_bump()  # bump to version=1
        bb.submit_patch({
            "base_version": 0,  # 過期
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "stale",
        })
        bb.merge_and_bump()
        assert bb.collect_pending() == []

    def test_no_patches_merge_still_bumps_version(self, bb):
        result = bb.merge_and_bump()
        assert result["applied"] == 0
        assert result["discarded"] == 0
        assert bb.version == 1
        assert bb.collect_pending() == []


# ─────────────────────────────────────────────────────────────────────────────
# 違規 patch 拋 PermissionError
# ─────────────────────────────────────────────────────────────────────────────

class TestPermissionViolatingPatch:
    """merge 時遇到違規 patch 必須拋 PermissionError。"""

    def test_patch_violating_anchor_raises(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",  # compactor 禁寫 real_bible
            "target": "real_bible",
            "value": {"hack": True},
        })
        with pytest.raises(PermissionError):
            bb.merge_and_bump()

    def test_patch_violating_npc_evolving_by_story_raises(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Zoe", "evolving": {}}])
        bb.submit_patch({
            "base_version": 0,
            "writer": "story",  # story 禁寫 npc_evolving
            "target": "npc_registry.Zoe.evolving",
            "value": {"intent": "betray"},
        })
        with pytest.raises(PermissionError):
            bb.merge_and_bump()

    def test_patch_violating_secret_core_by_orchestrator_raises(self, bb):
        bb.write("setup", "npc_registry", [{"name": "X", "secret_core": "hidden"}])
        bb.submit_patch({
            "base_version": 0,
            "writer": "orchestrator",
            "target": "npc_registry.X.secret_core",
            "value": "exposed",
        })
        with pytest.raises(PermissionError):
            bb.merge_and_bump()


# ─────────────────────────────────────────────────────────────────────────────
# collect_pending 快照不共享引用
# ─────────────────────────────────────────────────────────────────────────────

class TestCollectPendingIsolation:
    def test_collect_pending_returns_copy(self, bb):
        bb.submit_patch({
            "base_version": 0,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": "test",
        })
        pending = bb.collect_pending()
        # 修改回傳的清單不影響內部
        pending.clear()
        assert len(bb.collect_pending()) == 1

    def test_submit_patch_deep_copies_value(self, bb):
        mutable = {"key": "original"}
        bb.submit_patch({
            "base_version": 0,
            "writer": "setup",
            "target": "protagonist",
            "value": mutable,
        })
        # 修改原始物件
        mutable["key"] = "mutated"
        # patch 內應保留 "original"
        pending = bb.collect_pending()
        assert pending[0]["value"]["key"] == "original"


# ─────────────────────────────────────────────────────────────────────────────
# submit_patch 格式驗證
# ─────────────────────────────────────────────────────────────────────────────

class TestSubmitPatchValidation:
    def test_missing_base_version_raises_value_error(self, bb):
        with pytest.raises(ValueError):
            bb.submit_patch({"writer": "compactor", "target": "rolling_summary", "value": "x"})

    def test_missing_writer_raises_value_error(self, bb):
        with pytest.raises(ValueError):
            bb.submit_patch({"base_version": 0, "target": "rolling_summary", "value": "x"})

    def test_missing_target_raises_value_error(self, bb):
        with pytest.raises(ValueError):
            bb.submit_patch({"base_version": 0, "writer": "compactor", "value": "x"})

    def test_missing_value_raises_value_error(self, bb):
        with pytest.raises(ValueError):
            bb.submit_patch({"base_version": 0, "writer": "compactor", "target": "rolling_summary"})
