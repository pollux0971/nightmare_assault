"""MC2 — 聊天退出濃縮 驗收測試。

驗收：退出 → recent_chat_digest 含 3–4 句濃縮、非空；完整紀錄仍在 chat_logs；CHATROOM_CLOSED 發出。
"""
from __future__ import annotations

from core.agents.npc_chat import condense_chat, apply_chat_exit
from core.blackboard import Blackboard
from core.signal import SignalBus
from core.constants import EVT_CHATROOM_CLOSED


HISTORY = [
    {"role": "player", "content": "這裡發生什麼事？"},
    {"role": "npc", "content": "我只管照顧病人，別問太多。"},
    {"role": "player", "content": "你怕什麼？"},
    {"role": "npc", "content": "午夜之後，別待在三樓。"},
]


def test_condense_3to4_sentences():
    digest = condense_chat("張醫生", HISTORY)
    assert digest and "張醫生" in digest
    assert digest.count("。") + digest.count("？") <= 6     # 大致 3–4 句
    assert "午夜" in digest or "照顧病人" in digest          # 含 NPC 重點


def test_condense_empty_history():
    assert condense_chat("護士", []) and "護士" in condense_chat("護士", [])


def test_apply_chat_exit_writes_digest_and_signal():
    bb = Blackboard()
    bus = SignalBus()
    closed = []
    bus.subscribe(EVT_CHATROOM_CLOSED, lambda npc=None, *a, **k: closed.append(npc))
    digest = apply_chat_exit(bb, "張醫生", HISTORY, signal_bus=bus)
    # recent_chat_digest 進 story hot context
    assert bb.snapshot()["recent_chat_digest"] == digest and digest
    assert closed == ["張醫生"]                              # CHATROOM_CLOSED 發出


def test_apply_chat_exit_no_bus_ok():
    bb = Blackboard()
    digest = apply_chat_exit(bb, "護士", HISTORY)             # 無 signal bus 不炸
    assert digest
