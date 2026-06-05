"""P5 — 配置 UI 後端 API 驗收測試（draft → preview → activate；preview 不呼 LLM）。

驗收：
- 編輯 fragment 產 draft、不動 active run；activate 後才生效；preview 不呼 LLM；回傳 active profile + prompt_hash。
- JS 語法過（node --check，缺 node 則 skip）。
"""
from __future__ import annotations

import shutil
import subprocess
from pathlib import Path

import pytest

from webview_app import API
from core.persistence.db import Database


class FakeWindow:
    def evaluate_js(self, js):
        pass


def _api():
    """API + 單一共用 in-memory DB（讓 draft/activate 跨呼叫可見）。"""
    api = API(window=FakeWindow())
    api._config = {"api_key": "x", "base_url": "y", "agent_models": {}, "timeout": 5}
    db = Database()
    api._make_db = lambda: db
    return api


# ── 總覽 / 列表 ──────────────────────────────────────────────────────────────
def test_config_overview_shape():
    ov = _api().config_overview()
    assert ov["ok"] and ov["active_profile"] == "mvp_a_safe"
    assert ov["story"]["prompt_hash"]
    assert any(f["name"] == "ENABLE_CONFIG_CENTER" for f in ov["flags"])


def test_list_prompt_fragments_has_eight_story():
    frags = _api().list_prompt_fragments("story")
    keys = [f["fragment_key"] for f in frags]
    assert "story.kernel_obedience" in keys and "story.no_repetition" in keys
    assert len(frags) == 8


# ── draft 不動 active ────────────────────────────────────────────────────────
def test_save_draft_does_not_change_active():
    api = _api()
    before = api.preview_prompt("story")                      # active 編譯
    r = api.save_prompt_draft("story.style_horror", "DRAFT_SENTINEL_XYZ")
    assert r["ok"] and r["status"] == "draft"
    after = api.preview_prompt("story")                       # 仍是 active（無 override）
    assert "DRAFT_SENTINEL_XYZ" not in after["compiled_prompt"]
    assert after["prompt_hash"] == before["prompt_hash"]      # active hash 未變
    assert api.config_overview()["story"]["prompt_hash"] == before["prompt_hash"]


def test_preview_with_draft_shows_change_without_activating():
    api = _api()
    active_hash = api.preview_prompt("story")["prompt_hash"]
    prev = api.preview_prompt("story", None, "story.style_horror", "DRAFT_SENTINEL_XYZ")
    assert prev["ok"] and prev["llm_called"] is False
    assert "DRAFT_SENTINEL_XYZ" in prev["compiled_prompt"]
    assert prev["prompt_hash"] != active_hash                 # 預覽不同，但...
    assert api.preview_prompt("story")["prompt_hash"] == active_hash   # ...active 仍未變


def test_activate_draft_makes_it_active():
    api = _api()
    before = api.preview_prompt("story")["prompt_hash"]
    api.save_prompt_draft("story.style_horror", "ACTIVATED_SENTINEL")
    act = api.activate_prompt_draft("story.style_horror")
    assert act["ok"]
    after = api.preview_prompt("story")
    assert "ACTIVATED_SENTINEL" in after["compiled_prompt"]   # 啟用後 active 生效
    assert after["prompt_hash"] != before


def test_activate_without_draft_fails_gracefully():
    api = _api()
    r = api.activate_prompt_draft("story.role")               # 沒存過 draft
    assert r["ok"] is False


# ── preview 零 LLM ───────────────────────────────────────────────────────────
def test_preview_does_not_call_llm(monkeypatch):
    api = _api()
    # 若 preview 嘗試建 client / 呼 LLM，這裡會炸 → 測試會失敗
    api._make_client = lambda: (_ for _ in ()).throw(AssertionError("preview 不該建 LLM client"))
    r = api.preview_prompt("story")
    assert r["ok"] and r["llm_called"] is False
    assert r["compiled_prompt"]


# ── feature flags / profile ──────────────────────────────────────────────────
def test_toggle_feature_flag():
    api = _api()
    assert api.set_feature_flag("ENABLE_CONFIG_CENTER", True)["ok"]
    flags = {f["name"]: f["value"] for f in api.list_feature_flags()}
    assert flags["ENABLE_CONFIG_CENTER"] is True


def test_set_active_profile():
    api = _api()
    assert api.set_active_profile("debug")["ok"]
    assert api.get_active_profile()["active_profile"] == "debug"


# ── JS 語法 ──────────────────────────────────────────────────────────────────
def test_ui_js_syntax_valid():
    node = shutil.which("node")
    if not node:
        pytest.skip("node 不可用，略過 JS 語法檢查")
    js = Path(__file__).resolve().parent.parent / "ui" / "js" / "api.js"
    r = subprocess.run([node, "--check", str(js)], capture_output=True, text=True)
    assert r.returncode == 0, r.stderr
