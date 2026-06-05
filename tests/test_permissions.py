"""tests/test_permissions.py — 權限表完整覆蓋測試（U03）。

每個 writer 的可寫 target 成功、禁寫 target 拋 PermissionError。
"""
from __future__ import annotations

import pytest
from core.blackboard import Blackboard, can_write


# ─────────────────────────────────────────────────────────────────────────────
# 輔助：直接用 can_write 函式做單元測試（不建 Blackboard）
# ─────────────────────────────────────────────────────────────────────────────

class TestCanWrite:
    """can_write() 函式的直接單元測試。"""

    # ── setup ──────────────────────────────────────────────────────────────
    def test_setup_can_write_real_bible(self):
        assert can_write("setup", "real_bible") is True

    def test_setup_can_write_npc_registry(self):
        assert can_write("setup", "npc_registry") is True

    def test_setup_can_write_protagonist(self):
        assert can_write("setup", "protagonist") is True

    def test_setup_can_write_beat_window(self):
        assert can_write("setup", "beat_window") is True

    def test_setup_can_write_secret_core_path(self):
        # setup 也可以寫含 secret_core 的路徑（初始化）
        assert can_write("setup", "npc_registry.Alice.secret_core") is True

    # ── orchestrator ───────────────────────────────────────────────────────
    def test_orchestrator_can_write_revealed_bible(self):
        assert can_write("orchestrator", "revealed_bible") is True

    def test_orchestrator_can_write_turn_context(self):
        assert can_write("orchestrator", "turn_context") is True

    def test_orchestrator_cannot_write_real_bible(self):
        assert can_write("orchestrator", "real_bible") is False

    def test_orchestrator_cannot_write_secret_core(self):
        assert can_write("orchestrator", "npc_registry.Alice.secret_core") is False

    def test_orchestrator_cannot_write_protagonist(self):
        # 不在 allowed_roots 裡
        assert can_write("orchestrator", "protagonist") is False

    # ── story ──────────────────────────────────────────────────────────────
    def test_story_can_write_beat_window(self):
        assert can_write("story", "beat_window") is True

    def test_story_can_write_turn_context(self):
        assert can_write("story", "turn_context") is True

    def test_story_cannot_write_real_bible(self):
        assert can_write("story", "real_bible") is False

    def test_story_cannot_write_secret_core(self):
        assert can_write("story", "npc_registry.Alice.secret_core") is False

    def test_story_cannot_write_npc_evolving(self):
        assert can_write("story", "npc_registry.Alice.evolving") is False

    def test_story_cannot_write_npc_evolving_subfield(self):
        assert can_write("story", "npc_registry.Bob.evolving.intent") is False

    # ── warden ─────────────────────────────────────────────────────────────
    def test_warden_can_write_turn_context(self):
        assert can_write("warden", "turn_context") is True

    def test_warden_can_write_ledger(self):
        assert can_write("warden", "ledger") is True

    def test_warden_cannot_write_real_bible(self):
        assert can_write("warden", "real_bible") is False

    def test_warden_cannot_write_secret_core(self):
        assert can_write("warden", "npc_registry.Bob.secret_core") is False

    def test_warden_cannot_write_npc_evolving(self):
        assert can_write("warden", "npc_registry.Alice.evolving") is False

    # ── npc_chat ───────────────────────────────────────────────────────────
    def test_npc_chat_can_write_chat_log(self):
        assert can_write("npc_chat", "chat_log") is True

    def test_npc_chat_cannot_write_real_bible(self):
        assert can_write("npc_chat", "real_bible") is False

    def test_npc_chat_cannot_write_secret_core(self):
        assert can_write("npc_chat", "npc_registry.Alice.secret_core") is False

    def test_npc_chat_cannot_write_protagonist(self):
        assert can_write("npc_chat", "protagonist") is False

    # ── dreaming ───────────────────────────────────────────────────────────
    def test_dreaming_can_write_npc_evolving(self):
        assert can_write("dreaming", "npc_registry.Alice.evolving") is True

    def test_dreaming_can_write_npc_evolving_subfield(self):
        assert can_write("dreaming", "npc_registry.Alice.evolving.intent") is True

    def test_dreaming_can_write_npc_offstage_intent(self):
        assert can_write("dreaming", "npc_registry.Alice.offstage_intent") is True

    def test_dreaming_cannot_write_real_bible(self):
        assert can_write("dreaming", "real_bible") is False

    def test_dreaming_cannot_write_secret_core(self):
        assert can_write("dreaming", "npc_registry.Alice.secret_core") is False

    # ── offstage_fate ──────────────────────────────────────────────────────
    def test_offstage_fate_can_write_npc_presence(self):
        assert can_write("offstage_fate", "npc_registry.Alice.presence") is True

    def test_offstage_fate_can_write_npc_alignment(self):
        assert can_write("offstage_fate", "npc_registry.Alice.alignment") is True

    def test_offstage_fate_can_write_npc_carried_fragment(self):
        assert can_write("offstage_fate", "npc_registry.Alice.carried_fragment") is True

    def test_offstage_fate_can_write_npc_offstage_intent(self):
        assert can_write("offstage_fate", "npc_registry.Alice.offstage_intent") is True

    def test_offstage_fate_can_write_scene_registry(self):
        assert can_write("offstage_fate", "scene_registry") is True

    def test_offstage_fate_cannot_write_real_bible(self):
        assert can_write("offstage_fate", "real_bible") is False

    def test_offstage_fate_cannot_write_secret_core(self):
        assert can_write("offstage_fate", "npc_registry.Alice.secret_core") is False

    # ── compactor ──────────────────────────────────────────────────────────
    def test_compactor_can_write_rolling_summary(self):
        assert can_write("compactor", "rolling_summary") is True

    def test_compactor_can_write_ledger(self):
        assert can_write("compactor", "ledger") is True

    def test_compactor_can_write_recent_chat_digest(self):
        assert can_write("compactor", "recent_chat_digest") is True

    def test_compactor_cannot_write_real_bible(self):
        assert can_write("compactor", "real_bible") is False

    def test_compactor_cannot_write_secret_core(self):
        assert can_write("compactor", "npc_registry.Alice.secret_core") is False

    def test_compactor_cannot_write_npc_registry(self):
        # compactor allowed_roots 不含 npc_registry
        assert can_write("compactor", "npc_registry") is False

    # ── 未知 writer ────────────────────────────────────────────────────────
    def test_unknown_writer_cannot_write_anything(self):
        assert can_write("unknown_agent", "rolling_summary") is False
        assert can_write("hacker", "real_bible") is False


# ─────────────────────────────────────────────────────────────────────────────
# Blackboard.write() 的 PermissionError 測試
# ─────────────────────────────────────────────────────────────────────────────

class TestBlackboardWritePermissions:
    """確認 Blackboard.write() 在違規時拋 PermissionError。"""

    @pytest.fixture
    def bb(self):
        return Blackboard()

    # setup 可寫全部
    def test_setup_write_real_bible_succeeds(self, bb):
        bb.write("setup", "real_bible", {"key": "value"})
        assert bb.real_bible == {"key": "value"}

    def test_setup_write_protagonist_succeeds(self, bb):
        bb.write("setup", "protagonist", {"name": "Hero"})
        assert bb.protagonist == {"name": "Hero"}

    # orchestrator 禁寫 real_bible
    def test_orchestrator_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("orchestrator", "real_bible", {"hack": True})

    # orchestrator 禁寫 secret_core
    def test_orchestrator_write_secret_core_raises(self, bb):
        # 先用 setup 新增 NPC
        bb.write("setup", "npc_registry", [{"name": "Alice", "secret_core": "original"}])
        with pytest.raises(PermissionError):
            bb.write("orchestrator", "npc_registry.Alice.secret_core", "hacked")

    # story 禁寫 npc_evolving
    def test_story_write_npc_evolving_raises(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Bob", "evolving": {}}])
        with pytest.raises(PermissionError):
            bb.write("story", "npc_registry.Bob.evolving", {"intent": "betray"})

    # dreaming 可寫 npc_evolving
    def test_dreaming_write_npc_evolving_succeeds(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Alice", "evolving": {}}])
        bb.write("dreaming", "npc_registry.Alice.evolving", {"intent": "flee"})
        npc = bb.npc_registry[0]
        assert npc["evolving"] == {"intent": "flee"}

    # dreaming 禁寫 real_bible
    def test_dreaming_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("dreaming", "real_bible", {"hack": True})

    # compactor 禁寫錨點
    def test_compactor_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("compactor", "real_bible", {"fake": True})

    def test_compactor_write_secret_core_raises(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Carol", "secret_core": "hidden"}])
        with pytest.raises(PermissionError):
            bb.write("compactor", "npc_registry.Carol.secret_core", "exposed")

    # compactor 可寫 rolling_summary
    def test_compactor_write_rolling_summary_succeeds(self, bb):
        bb.write("compactor", "rolling_summary", "compact text")
        assert bb.rolling_summary == "compact text"

    # warden 可寫 ledger
    def test_warden_write_ledger_succeeds(self, bb):
        bb.write("warden", "ledger", [{"type": "fact", "content": "Player died"}])
        assert len(bb.ledger) == 1

    # warden 禁寫 real_bible
    def test_warden_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("warden", "real_bible", {})

    # warden 禁寫 npc_evolving
    def test_warden_write_npc_evolving_raises(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Dave", "evolving": {}}])
        with pytest.raises(PermissionError):
            bb.write("warden", "npc_registry.Dave.evolving", {"intent": "betray"})

    # npc_chat 可寫 chat_log
    def test_npc_chat_write_chat_log_succeeds(self, bb):
        bb.write("npc_chat", "chat_log", [{"role": "npc", "text": "Hello"}])
        assert bb.chat_log[0]["role"] == "npc"

    # npc_chat 禁寫錨點
    def test_npc_chat_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("npc_chat", "real_bible", {})

    def test_npc_chat_write_secret_core_raises(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Eve", "secret_core": "secret"}])
        with pytest.raises(PermissionError):
            bb.write("npc_chat", "npc_registry.Eve.secret_core", "exposed")

    # offstage_fate 可寫 npc presence
    def test_offstage_fate_write_npc_presence_succeeds(self, bb):
        bb.write("setup", "npc_registry", [{"name": "Frank", "presence": "present"}])
        bb.write("offstage_fate", "npc_registry.Frank.presence", "absent")
        assert bb.npc_registry[0]["presence"] == "absent"

    # offstage_fate 禁寫 real_bible
    def test_offstage_fate_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("offstage_fate", "real_bible", {})

    # story 禁寫 real_bible
    def test_story_write_real_bible_raises(self, bb):
        with pytest.raises(PermissionError):
            bb.write("story", "real_bible", {})

    # story 可寫 beat_window（append 語意）
    def test_story_write_beat_window_appends(self, bb):
        bb.write("story", "beat_window", "beat1")
        bb.write("story", "beat_window", "beat2")
        assert bb.beat_window == ["beat1", "beat2"]
