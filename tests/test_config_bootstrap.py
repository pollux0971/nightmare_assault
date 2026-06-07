"""config 框架自動生成（clone 後第一次跑、無 config.json 時的 onboarding）。

驗收：缺檔 → 生成框架（api_key 空、8 個 agent 齊全）；已存在 → 不覆寫；
load_config 在缺檔時也不崩、且 api_key 為空 → configured=False（交給設定畫面/env）。
"""
from __future__ import annotations

import json
import tempfile
from pathlib import Path

import webview_app as W


def _tmp():
    return Path(tempfile.mkdtemp()) / "config.json"


# ── 缺檔 → 生成框架（空金鑰、8 agent 齊全）─────────────────────────────────
def test_bootstrap_creates_skeleton():
    p = _tmp()
    assert W.bootstrap_config_skeleton(p) is True
    cfg = json.loads(p.read_text(encoding="utf-8"))
    assert cfg["api_key"] == ""                            # 框架不含金鑰
    assert cfg["base_url"] and cfg["timeout"]
    # 8 個 agent 都有模型（即使 example 缺項也補齊）
    for agent in ("story", "setup", "orchestrator", "warden", "compactor",
                  "npc-chat", "dreaming", "offstage-fate"):
        assert cfg["agent_models"].get(agent), f"缺 agent 模型: {agent}"


# ── 已存在 → 不覆寫（保護使用者真實金鑰）──────────────────────────────────
def test_bootstrap_never_overwrites():
    p = _tmp()
    p.write_text('{"api_key":"REAL-KEY","base_url":"x"}', encoding="utf-8")
    assert W.bootstrap_config_skeleton(p) is False
    assert json.loads(p.read_text(encoding="utf-8"))["api_key"] == "REAL-KEY"


def _no_dotenv(monkeypatch):
    """隔離本機 .env：env 行為測試只看 monkeypatch 設定的環境變數。"""
    monkeypatch.setattr(W, "_load_dotenv", lambda *a, **k: None)


# ── load_config 缺檔不崩、api_key 空 → 未設定（走設定畫面/env）─────────────
def test_load_config_missing_is_unconfigured(monkeypatch):
    _no_dotenv(monkeypatch)
    monkeypatch.delenv("OPENROUTER_API_KEY", raising=False)
    monkeypatch.delenv("OPENAI_API_KEY", raising=False)
    p = _tmp()
    cfg = W.load_config(p)
    assert cfg.get("api_key", "") == ""                    # 未設定 → configured False
    assert cfg["agent_models"] and cfg["base_url"]
    assert p.is_file()                                      # 同時把框架落到磁碟


# ── 缺檔但有 env 金鑰 → 採用 env（OPENROUTER 與 OPENAI 皆支援）─────────────────
def test_load_config_uses_env_when_present(monkeypatch):
    _no_dotenv(monkeypatch)
    monkeypatch.delenv("OPENAI_API_KEY", raising=False)
    monkeypatch.setenv("OPENROUTER_API_KEY", "test-or-env-xyz")
    cfg = W.load_config(_tmp())
    assert cfg["api_key"] == "test-or-env-xyz"


def test_load_config_uses_openai_env(monkeypatch):
    _no_dotenv(monkeypatch)
    monkeypatch.delenv("OPENROUTER_API_KEY", raising=False)
    monkeypatch.setenv("OPENAI_API_KEY", "test-openai-env-abc")
    cfg = W.load_config(_tmp())
    assert cfg["api_key"] == "test-openai-env-abc"            # 優先採 OPENAI_API_KEY
