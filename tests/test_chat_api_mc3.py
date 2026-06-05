"""MC3 — 聊天室 UI + API 驗收測試。

驗收：list_present_npcs 只列在場；open/send/close round-trip；send_chat 回覆非空；close 觸發濃縮；不洩漏 secret。
"""
from __future__ import annotations

import json

from webview_app import API
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus

SECRET = "兇手是張醫生SECRET"


class FakeWindow:
    def evaluate_js(self, js): pass


class ChatCaller:
    def call(self, agent, context, output_model=None, temperature=None):
        assert agent == "npc-chat"
        return "「我只管照顧病人。」"
    def stream(self, *a, **k):
        yield ""


def _api():
    bb = Blackboard()
    bb.write("setup", "real_bible", {"world_truth": {"what_really_happened": SECRET}})
    bb.write("setup", "npc_registry", [
        {"name": "張醫生", "profession": "醫生", "personality": "mysterious", "voice_sample": "…",
         "public_face": "冷靜", "secret_core": SECRET, "self_aware": True, "presence": "present",
         "evolving": {"intent": "observe"}},
        {"name": "離場者", "profession": "警衛", "personality": "nervous", "voice_sample": "…",
         "public_face": "緊張", "secret_core": "x", "self_aware": False, "presence": "missing",
         "evolving": {}},
    ])
    api = API(window=FakeWindow())
    api._config = {"api_key": "x", "base_url": "y", "agent_models": {}, "timeout": 5}
    loop = BeatLoop(ChatCaller(), bb, Database(), SignalBus(), run_id="t", use_kernel=False)
    loop.beat_number = 3
    api._loop = loop
    return api


def test_list_present_only():
    npcs = [n["name"] for n in _api().list_present_npcs()]
    assert "張醫生" in npcs and "離場者" not in npcs       # 只列在場


def test_open_send_close_roundtrip():
    api = _api()
    assert api.open_chatroom("張醫生")["ok"]
    r = api.send_chat("張醫生", "這裡發生什麼事？")
    assert r["ok"] and r["reply"]                          # 回覆非空
    # 不洩漏 secret
    assert SECRET not in json.dumps(r, ensure_ascii=False)
    c = api.close_chatroom("張醫生")
    assert c["ok"] and c["digest"]                         # 退出濃縮
    # 完整紀錄入 chat_logs
    logs = api._loop.db.load_chat_logs("t", "張醫生")
    assert len(logs) == 2


def test_open_absent_fails():
    assert _api().open_chatroom("離場者")["ok"] is False    # 不在場不能開


def test_recent_digest_in_story_context_after_close():
    api = _api()
    api.open_chatroom("張醫生")
    api.send_chat("張醫生", "午夜會怎樣？")
    api.close_chatroom("張醫生")
    assert api._loop.bb.snapshot()["recent_chat_digest"]   # story 下個 beat 看得到
