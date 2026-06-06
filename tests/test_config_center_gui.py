"""Config Center GUI completion（v0.8）—— 後端 glue smoke + 前端 DOM 靜態檢查。

無 JS engine，故前端以「HTML/JS 結構斷言」當 DOM smoke：Config Center 能開、有 Agent Models /
Prompt Blocks 表、切 agent 刷新、Preview 零 LLM、API key 不出現在 config table。
"""
from __future__ import annotations

import json
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def _api_with_tmp():
    tmp = Path(tempfile.mkdtemp())
    cfg = tmp / "config.json"
    cfg.write_text(json.dumps({
        "api_key": "sk-or-test-SECRET", "base_url": "https://x",
        "agent_models": {"story": ["deepseek/x", "qwen/y"], "warden": ["minimax/z"]},
        "timeout": 90}), encoding="utf-8")
    from webview_app import API
    from core.persistence.db import Database
    api = API(config_path=cfg)
    db = Database(str(tmp / "game.db"))
    api._make_db = lambda: db                            # 用臨時 DB，不碰共享 game.db
    return api, cfg


# ── 後端 glue ─────────────────────────────────────────────────────────────────
def test_agent_models_overview_no_key():
    api, _ = _api_with_tmp()
    mo = api.agent_models_overview()
    assert mo["ok"] and {a["agent"] for a in mo["agents"]} >= {"story", "warden"}
    story = next(a for a in mo["agents"] if a["agent"] == "story")
    assert story["primary"] == "deepseek/x" and story["fallbacks"] == ["qwen/y"]
    assert story["temperature"] == 0.8                   # DEFAULT_TEMPS
    # **API key 絕不出現在表格資料**
    assert "sk-or-test" not in json.dumps(mo, ensure_ascii=False)
    assert all("api_key" not in a for a in mo["agents"])


def test_list_prompt_blocks_has_rows():
    api, _ = _api_with_tmp()
    lb = api.list_prompt_blocks("story")
    assert lb["ok"] and len(lb["blocks"]) >= 1           # 有 fragment rows
    b = lb["blocks"][0]
    for k in ("fragment_key", "enabled", "sort_order", "status", "version", "preview"):
        assert k in b


def test_switch_agent_refreshes_blocks():
    api, _ = _api_with_tmp()
    s = api.list_prompt_blocks("story")
    w = api.list_prompt_blocks("warden")
    assert s["agent"] == "story" and w["agent"] == "warden"   # 切 agent → 不同結果集


def test_preview_no_llm():
    api, _ = _api_with_tmp()
    pv = api.preview_prompt("story")
    assert pv["ok"] and pv["llm_called"] is False and pv["compiled_prompt"]


def test_save_agent_models_preserves_key():
    api, cfg = _api_with_tmp()
    r = api.save_agent_models([{"agent": "story", "primary": "new/model",
                                "fallbacks": ["fb/1"], "temperature": 0.9,
                                "max_tokens": 1000, "enabled": True}])
    saved = json.loads(cfg.read_text(encoding="utf-8"))
    assert r["ok"]
    assert saved["agent_models"]["story"] == ["new/model", "fb/1"]
    assert saved["api_key"] == "sk-or-test-SECRET"       # key 保留、不被清掉
    assert saved["agent_settings"]["story"]["temperature"] == 0.9


def test_set_fragment_enabled_roundtrip():
    api, _ = _api_with_tmp()
    blocks = api.list_prompt_blocks("story")["blocks"]
    key = blocks[0]["fragment_key"]
    assert api.set_fragment_enabled("story", key, False)["ok"]
    after = {b["fragment_key"]: b["enabled"] for b in api.list_prompt_blocks("story")["blocks"]}
    assert after[key] is False                            # 停用後仍可見（含 disabled）
    assert api.set_fragment_enabled("story", key, True)["ok"]


# ── 前端 DOM 靜態檢查（無 JS engine）──────────────────────────────────────────
def _html():
    return (ROOT / "ui" / "index.html").read_text(encoding="utf-8")


def _apijs():
    return (ROOT / "ui" / "js" / "api.js").read_text(encoding="utf-8")


def test_config_center_dialog_and_tabs_exist():
    h = _html()
    assert 'id="dlg-config"' in h
    for tab in ("models", "blocks", "preview", "test", "flags"):
        assert f'data-tab="{tab}"' in h and f'data-panel="{tab}"' in h
    assert 'id="cfg-models-body"' in h and 'id="cfg-blocks-body"' in h
    assert 'id="cfg-preview"' in h
    # Test Prompt 後端尚未支援 → 按鈕 disabled + 標示
    assert 'id="cfg-btn-test"' in h and "disabled" in h.split('id="cfg-btn-test"')[1][:120]
    assert "backend API missing" in h


def test_apijs_wires_new_apis():
    j = _apijs()
    for fn in ("agent_models_overview", "save_agent_models", "list_prompt_blocks",
               "set_fragment_enabled", "preview_prompt", "test_model"):
        assert fn in j, f"api.js 未接 {fn}"
    assert "ConfigUI.open" in j and "cfg-blocks-agent" in j   # 開啟 + 切 agent 刷新
    assert "loadBlocks" in j


def test_apijs_models_table_has_no_api_key():
    j = _apijs()
    # Agent Models 渲染路徑不得碰 api_key（key 不入 config table）
    render = j.split("renderModels")[1].split("_collectModels")[0] if "renderModels" in j else ""
    assert "api_key" not in render and "apiKey" not in render
