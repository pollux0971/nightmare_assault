"""P1 — 配置表（additive SQLite + 種子）驗收測試。

驗收：
- additive：既有 runs/beats/save_points 結構不變、舊存檔可讀回；migration 冪等。
- 種子：mvp_a_safe is_active；story agent_config 預設符 docs/04；story include_real_bible=0。
- 既有 app 仍可開（import core + 開既有存檔 smoke 不崩）。
"""
from __future__ import annotations

import sqlite3

from core.persistence.db import Database
from core.config.schema import init_config_schema, seed_defaults, ConfigStore
from core.config import fragments as F


CONFIG_TABLES = [
    "config_profiles", "agent_configs", "prompt_fragments", "prompt_fragment_versions",
    "agent_prompt_bindings", "agent_context_policies", "feature_flags",
    "run_config_snapshots", "prompt_test_cases", "prompt_test_results",
]
BASE_TABLES = ["runs", "beats", "npc_states", "inventory_snapshots",
               "chat_logs", "save_points", "llm_traces", "schema_meta"]


def _tables(conn) -> set[str]:
    return {r[0] for r in conn.execute(
        "SELECT name FROM sqlite_master WHERE type='table';").fetchall()}


# ── additive ────────────────────────────────────────────────────────────────
def test_config_tables_created_on_open():
    db = Database(":memory:")
    names = _tables(db.connection)
    for t in CONFIG_TABLES:
        assert t in names, f"配置表 {t} 未建立"
    for t in BASE_TABLES:
        assert t in names, f"既有表 {t} 不應消失"
    db.close()


def test_base_tables_intact_and_functional():
    """既有 runs/beats/save_points round-trip 不受配置表影響。"""
    db = Database(":memory:")
    db.create_run("r1", theme="ship")
    db.save_beat("r1", 1, "narr", '{"x":1}', "sum", '{"v":1}', is_narration_only=False)
    got = db.load_beat("r1", 1)
    assert got["narrative"] == "narr"
    db.add_save_point("r1", 1, "auto")
    assert db.list_save_points("r1")[0]["label"] == "auto"
    assert db.schema_version() == "1"
    db.close()


def test_old_save_reopen_is_additive_and_idempotent(tmp_path):
    """模擬舊存檔：寫 run/beat → 關 → 重開 → 既有資料在、配置表補上、種子不重複。"""
    path = str(tmp_path / "save.db")
    db = Database(path)
    db.create_run("r1", theme="hospital")
    db.save_beat("r1", 3, "old beat", "{}", "", "{}")
    db.close()

    # 重開（再次跑 additive migration + seed）
    db2 = Database(path)
    assert db2.load_beat("r1", 3)["narrative"] == "old beat"   # 舊資料完好
    store = db2.config_store()
    # 種子冪等：profiles 仍是 5（沒因重開翻倍）
    assert len(store.list_profiles()) == len(F.PROFILES)
    n_bindings = db2.connection.execute(
        "SELECT COUNT(*) FROM agent_prompt_bindings WHERE agent_name='story';").fetchone()[0]
    assert n_bindings == len(F.STORY_BINDING_ORDER)
    db2.close()


def test_init_and_seed_idempotent_on_raw_conn():
    conn = sqlite3.connect(":memory:")
    conn.row_factory = sqlite3.Row
    conn.execute("CREATE TABLE IF NOT EXISTS schema_meta (key TEXT PRIMARY KEY, value TEXT);")
    for _ in range(3):                       # 跑三次不應重複種子
        init_config_schema(conn)
        seed_defaults(conn)
    n = conn.execute("SELECT COUNT(*) FROM config_profiles;").fetchone()[0]
    assert n == len(F.PROFILES)
    n_frag = conn.execute("SELECT COUNT(*) FROM prompt_fragments WHERE fragment_key LIKE 'story.%';").fetchone()[0]
    assert n_frag == len(F.STORY_FRAGMENTS)
    conn.close()


# ── 種子內容 ─────────────────────────────────────────────────────────────────
def test_seed_active_profile_is_mvp_a_safe():
    store = Database(":memory:").config_store()
    assert store.active_profile() == "mvp_a_safe"


def test_story_agent_config_defaults_match_docs04():
    store = Database(":memory:").config_store()
    cfg = store.get_agent_config("story", "mvp_a_safe")
    assert cfg is not None
    assert 0.65 <= cfg["temperature"] <= 0.8           # docs/04
    assert 5000 <= cfg["context_budget_tokens"] <= 8000
    assert cfg["stream_enabled"] == 1
    assert cfg["output_schema_name"] == "StoryBeatOutput"


def test_story_context_policy_forbids_real_bible():
    store = Database(":memory:").config_store()
    pol = store.get_context_policy("story", "mvp_a_safe")
    assert pol is not None
    assert pol["include_real_bible"] == 0               # ★ 硬規則 C2/E2


def test_story_bindings_ordered_eight_fragments():
    store = Database(":memory:").config_store()
    frags = store.get_bound_fragments("story", "mvp_a_safe")
    assert [f["fragment_key"] for f in frags] == [k for k, _ in F.STORY_BINDING_ORDER]
    orders = [f["sort_order"] for f in frags]
    assert orders == sorted(orders)                     # 已依 sort_order 排序
    # join 後內容非空（composer 才能組裝）
    assert all(f["content"].strip() for f in frags)


def test_feature_flags_seeded():
    store = Database(":memory:").config_store()
    assert store.get_flag("ENABLE_CONFIG_CENTER", "mvp_a_safe") == 0
    assert store.get_flag("ENABLE_PROMPT_PREVIEW", "mvp_a_safe") == 1
