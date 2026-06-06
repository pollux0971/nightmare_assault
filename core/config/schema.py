"""core.config.schema — 配置中心 additive 配置表 + 種子 + ConfigStore（P1）。

**additive only**：所有 DDL 用 CREATE TABLE IF NOT EXISTS，不改名/刪除既有 U05 表；
種子用 INSERT OR IGNORE（冪等，重跑不重複）。掛在既有 Database 同一條 sqlite3 連線上，
既有存檔開啟時自動補表，舊資料不受影響。

契約：dev/CONTRACTS.md §十一 `ConfigSchema` / `AgentContextPolicy`；canonical = patch docs/04。
"""
from __future__ import annotations

import datetime
import json
import sqlite3
from typing import Optional

from core.config import fragments as F

CONFIG_SCHEMA_VERSION = "1"

_CONFIG_DDL = """
CREATE TABLE IF NOT EXISTS config_profiles (
    profile_name TEXT PRIMARY KEY,
    description  TEXT,
    is_active    INTEGER DEFAULT 0,
    created_at   TEXT,
    updated_at   TEXT
);

CREATE TABLE IF NOT EXISTS agent_configs (
    id                    INTEGER PRIMARY KEY,
    agent_name            TEXT NOT NULL,
    profile_name          TEXT NOT NULL,
    enabled               INTEGER DEFAULT 1,
    model                 TEXT,
    temperature           REAL,
    max_output_tokens     INTEGER,
    context_budget_tokens INTEGER,
    prompt_version        TEXT,
    output_schema_name    TEXT,
    stream_enabled        INTEGER DEFAULT 1,
    updated_at            TEXT
);

CREATE TABLE IF NOT EXISTS prompt_fragments (
    fragment_key       TEXT PRIMARY KEY,
    title              TEXT,
    category           TEXT,
    content            TEXT,
    description        TEXT,
    enabled_by_default INTEGER DEFAULT 1,
    status             TEXT DEFAULT 'active',
    version            INTEGER DEFAULT 1,
    updated_at         TEXT
);

CREATE TABLE IF NOT EXISTS prompt_fragment_versions (
    id           INTEGER PRIMARY KEY,
    fragment_key TEXT NOT NULL,
    version      INTEGER NOT NULL,
    content      TEXT,
    status       TEXT,
    created_at   TEXT
);

CREATE TABLE IF NOT EXISTS agent_prompt_bindings (
    id           INTEGER PRIMARY KEY,
    agent_name   TEXT NOT NULL,
    profile_name TEXT NOT NULL,
    fragment_key TEXT NOT NULL,
    enabled      INTEGER DEFAULT 1,
    sort_order   INTEGER DEFAULT 100
);

CREATE TABLE IF NOT EXISTS agent_context_policies (
    id                          INTEGER PRIMARY KEY,
    agent_name                  TEXT NOT NULL,
    profile_name                TEXT NOT NULL,
    max_recent_beats            INTEGER,
    max_relevant_clues          INTEGER,
    max_relevant_items          INTEGER,
    max_visible_npcs            INTEGER,
    include_full_history        INTEGER DEFAULT 0,
    include_full_revealed_bible INTEGER DEFAULT 1,
    include_real_bible          INTEGER DEFAULT 0,
    include_debug_trace         INTEGER DEFAULT 0,
    retrieval_strategy          TEXT DEFAULT 'relevant_only'
);

CREATE TABLE IF NOT EXISTS feature_flags (
    flag_name    TEXT NOT NULL,
    profile_name TEXT NOT NULL,
    value        INTEGER,
    updated_at   TEXT
);

CREATE TABLE IF NOT EXISTS run_config_snapshots (
    id                     INTEGER PRIMARY KEY,
    run_id                 TEXT NOT NULL,
    profile_name           TEXT,
    agent_name             TEXT,
    config_json            TEXT,
    compiled_prompt_hash   TEXT,
    compiled_prompt_preview TEXT,
    enabled_fragments_json TEXT,
    created_at             TEXT
);

CREATE TABLE IF NOT EXISTS prompt_test_cases (
    id           INTEGER PRIMARY KEY,
    name         TEXT NOT NULL,
    agent_name   TEXT,
    fragments_json TEXT,
    variables_json TEXT,
    expected_kind  TEXT,
    expected_value TEXT,
    created_at   TEXT
);

CREATE TABLE IF NOT EXISTS prompt_test_results (
    id           INTEGER PRIMARY KEY,
    test_case_id INTEGER,
    run_at       TEXT,
    passed       INTEGER,
    detail       TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_configs_uniq
    ON agent_configs(agent_name, profile_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bindings_uniq
    ON agent_prompt_bindings(agent_name, profile_name, fragment_key);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ctxpolicy_uniq
    ON agent_context_policies(agent_name, profile_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_flags_uniq
    ON feature_flags(flag_name, profile_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fragver_uniq
    ON prompt_fragment_versions(fragment_key, version);
CREATE INDEX IF NOT EXISTS idx_run_cfg_snap_run
    ON run_config_snapshots(run_id);
"""


def _now() -> str:
    return datetime.datetime.now().isoformat()


def init_config_schema(conn: sqlite3.Connection) -> None:
    """建立所有配置表/index（additive、冪等）。"""
    with conn:
        conn.executescript(_CONFIG_DDL)
        conn.execute(
            "INSERT OR IGNORE INTO schema_meta (key, value) VALUES ('config_schema_version', ?);",
            (CONFIG_SCHEMA_VERSION,),
        )


def seed_defaults(conn: sqlite3.Connection) -> None:
    """種子 mvp_a_safe + story 預設（冪等：INSERT OR IGNORE，不覆寫使用者改動）。"""
    now = _now()
    with conn:
        # profiles
        for name, desc, is_active in F.PROFILES:
            conn.execute(
                "INSERT OR IGNORE INTO config_profiles "
                "(profile_name, description, is_active, created_at, updated_at) "
                "VALUES (?, ?, ?, ?, ?);",
                (name, desc, is_active, now, now),
            )
        # story agent_config（綁 default profile）
        c = F.STORY_AGENT_CONFIG
        conn.execute(
            "INSERT OR IGNORE INTO agent_configs "
            "(agent_name, profile_name, enabled, model, temperature, max_output_tokens, "
            " context_budget_tokens, prompt_version, output_schema_name, stream_enabled, updated_at) "
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
            (c["agent_name"], F.DEFAULT_PROFILE, c["enabled"], c["model"], c["temperature"],
             c["max_output_tokens"], c["context_budget_tokens"], c["prompt_version"],
             c["output_schema_name"], c["stream_enabled"], now),
        )
        # story fragments + version 1
        for key, title, category, content, desc, enabled in F.STORY_FRAGMENTS:
            conn.execute(
                "INSERT OR IGNORE INTO prompt_fragments "
                "(fragment_key, title, category, content, description, enabled_by_default, status, version, updated_at) "
                "VALUES (?, ?, ?, ?, ?, ?, 'active', 1, ?);",
                (key, title, category, content, desc, enabled, now),
            )
            conn.execute(
                "INSERT OR IGNORE INTO prompt_fragment_versions "
                "(fragment_key, version, content, status, created_at) VALUES (?, 1, ?, 'active', ?);",
                (key, content, now),
            )
        # story bindings（依 STORY_BINDING_ORDER）
        for key, order in F.STORY_BINDING_ORDER:
            conn.execute(
                "INSERT OR IGNORE INTO agent_prompt_bindings "
                "(agent_name, profile_name, fragment_key, enabled, sort_order) "
                "VALUES ('story', ?, ?, 1, ?);",
                (F.DEFAULT_PROFILE, key, order),
            )
        # story context policy
        p = F.STORY_CONTEXT_POLICY
        conn.execute(
            "INSERT OR IGNORE INTO agent_context_policies "
            "(agent_name, profile_name, max_recent_beats, max_relevant_clues, max_relevant_items, "
            " max_visible_npcs, include_full_history, include_full_revealed_bible, include_real_bible, "
            " include_debug_trace, retrieval_strategy) "
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
            (p["agent_name"], F.DEFAULT_PROFILE, p["max_recent_beats"], p["max_relevant_clues"],
             p["max_relevant_items"], p["max_visible_npcs"], p["include_full_history"],
             p["include_full_revealed_bible"], p["include_real_bible"], p["include_debug_trace"],
             p["retrieval_strategy"]),
        )
        # feature flags（綁 default profile）
        for flag, value in F.FEATURE_FLAGS:
            conn.execute(
                "INSERT OR IGNORE INTO feature_flags (flag_name, profile_name, value, updated_at) "
                "VALUES (?, ?, ?, ?);",
                (flag, F.DEFAULT_PROFILE, value, now),
            )
        # ── 其他 agent（P7：warden/orchestrator/compactor/setup）──────────────
        for agent, frags in F.OTHER_AGENT_FRAGMENTS.items():
            for key, title, category, content, desc, enabled in frags:
                conn.execute(
                    "INSERT OR IGNORE INTO prompt_fragments "
                    "(fragment_key, title, category, content, description, enabled_by_default, status, version, updated_at) "
                    "VALUES (?, ?, ?, ?, ?, ?, 'active', 1, ?);",
                    (key, title, category, content, desc, enabled, now),
                )
                conn.execute(
                    "INSERT OR IGNORE INTO prompt_fragment_versions "
                    "(fragment_key, version, content, status, created_at) VALUES (?, 1, ?, 'active', ?);",
                    (key, content, now),
                )
            for key, order in F.OTHER_AGENT_BINDINGS[agent]:
                conn.execute(
                    "INSERT OR IGNORE INTO agent_prompt_bindings "
                    "(agent_name, profile_name, fragment_key, enabled, sort_order) VALUES (?, ?, ?, 1, ?);",
                    (agent, F.DEFAULT_PROFILE, key, order),
                )
            c = F.OTHER_AGENT_CONFIGS[agent]
            conn.execute(
                "INSERT OR IGNORE INTO agent_configs "
                "(agent_name, profile_name, enabled, model, temperature, max_output_tokens, "
                " context_budget_tokens, prompt_version, output_schema_name, stream_enabled, updated_at) "
                "VALUES (?, ?, 1, NULL, ?, ?, ?, 'v1', ?, 0, ?);",
                (agent, F.DEFAULT_PROFILE, c["temperature"], c["max_output_tokens"],
                 c["context_budget_tokens"], c["output_schema_name"], now),
            )
            p = F.OTHER_AGENT_CONTEXT_POLICIES[agent]
            conn.execute(
                "INSERT OR IGNORE INTO agent_context_policies "
                "(agent_name, profile_name, max_recent_beats, max_relevant_clues, max_relevant_items, "
                " max_visible_npcs, include_full_history, include_full_revealed_bible, include_real_bible, "
                " include_debug_trace, retrieval_strategy) "
                "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
                (agent, F.DEFAULT_PROFILE, p["max_recent_beats"], p["max_relevant_clues"],
                 p["max_relevant_items"], p["max_visible_npcs"], p["include_full_history"],
                 p["include_full_revealed_bible"], p["include_real_bible"], p["include_debug_trace"],
                 p["retrieval_strategy"]),
            )


class ConfigStore:
    """配置讀寫（掛在既有 Database 連線上）。讀為主；P5 寫入 fragment draft/activate 由本類提供。"""

    def __init__(self, conn: sqlite3.Connection, ensure: bool = True) -> None:
        self._conn = conn
        if ensure:
            init_config_schema(conn)
            seed_defaults(conn)

    @property
    def connection(self) -> sqlite3.Connection:
        """暴露底層連線（同一條 Database 連線）。"""
        return self._conn

    # ── profiles ──────────────────────────────────────────────────────
    def active_profile(self) -> str:
        row = self._conn.execute(
            "SELECT profile_name FROM config_profiles WHERE is_active = 1 "
            "ORDER BY updated_at DESC LIMIT 1;"
        ).fetchone()
        return row["profile_name"] if row else F.DEFAULT_PROFILE

    def default_profile(self) -> str:
        return F.DEFAULT_PROFILE

    def list_profiles(self) -> list[dict]:
        return [dict(r) for r in self._conn.execute(
            "SELECT * FROM config_profiles ORDER BY profile_name;").fetchall()]

    def set_active_profile(self, profile_name: str) -> None:
        with self._conn:
            self._conn.execute("UPDATE config_profiles SET is_active = 0;")
            self._conn.execute(
                "UPDATE config_profiles SET is_active = 1, updated_at = ? WHERE profile_name = ?;",
                (_now(), profile_name),
            )

    # ── agent_configs ─────────────────────────────────────────────────
    def get_agent_config(self, agent: str, profile: Optional[str] = None) -> Optional[dict]:
        profile = profile or self.active_profile()
        row = self._conn.execute(
            "SELECT * FROM agent_configs WHERE agent_name = ? AND profile_name = ?;",
            (agent, profile),
        ).fetchone()
        return dict(row) if row else None

    # ── fragments / bindings ──────────────────────────────────────────
    def get_fragment(self, fragment_key: str) -> Optional[dict]:
        row = self._conn.execute(
            "SELECT * FROM prompt_fragments WHERE fragment_key = ?;", (fragment_key,)
        ).fetchone()
        return dict(row) if row else None

    def get_bound_fragments(self, agent: str, profile: Optional[str] = None) -> list[dict]:
        """回傳該 agent/profile 啟用的 fragment（已 join 內容、依 sort_order 排序）。"""
        profile = profile or self.active_profile()
        rows = self._conn.execute(
            """
            SELECT b.fragment_key, b.sort_order, f.title, f.category, f.content, f.status, f.version
            FROM agent_prompt_bindings b
            JOIN prompt_fragments f ON f.fragment_key = b.fragment_key
            WHERE b.agent_name = ? AND b.profile_name = ? AND b.enabled = 1
            ORDER BY b.sort_order ASC, b.fragment_key ASC;
            """,
            (agent, profile),
        ).fetchall()
        return [dict(r) for r in rows]

    def get_all_bindings(self, agent: str, profile: Optional[str] = None) -> list[dict]:
        """回傳該 agent/profile 的**所有** binding（含 disabled），供 Config Center 顯示與切換。

        唯讀；含 enabled 旗標 + updated_at（供 Prompt Blocks 表）。
        """
        profile = profile or self.active_profile()
        rows = self._conn.execute(
            """
            SELECT b.fragment_key, b.sort_order, b.enabled,
                   f.title, f.category, f.content, f.status, f.version, f.updated_at
            FROM agent_prompt_bindings b
            JOIN prompt_fragments f ON f.fragment_key = b.fragment_key
            WHERE b.agent_name = ? AND b.profile_name = ?
            ORDER BY b.sort_order ASC, b.fragment_key ASC;
            """,
            (agent, profile),
        ).fetchall()
        return [dict(r) for r in rows]

    def upsert_fragment(self, fragment_key: str, content: str, *, title: str = "",
                        category: str = "rules", description: str = "",
                        status: str = "draft") -> int:
        """新增/更新 fragment 內容並 bump version（寫 prompt_fragment_versions）。回傳新 version。

        P5 draft 流程：status 預設 'draft'（不影響 active 組裝直到 activate_fragment）。
        """
        now = _now()
        existing = self.get_fragment(fragment_key)
        new_version = (existing["version"] + 1) if existing else 1
        with self._conn:
            if existing:
                self._conn.execute(
                    "UPDATE prompt_fragments SET content = ?, status = ?, version = ?, "
                    "title = COALESCE(NULLIF(?, ''), title), updated_at = ? WHERE fragment_key = ?;",
                    (content, status, new_version, title, now, fragment_key),
                )
            else:
                self._conn.execute(
                    "INSERT INTO prompt_fragments "
                    "(fragment_key, title, category, content, description, enabled_by_default, status, version, updated_at) "
                    "VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?);",
                    (fragment_key, title, category, content, description, status, new_version, now),
                )
            self._conn.execute(
                "INSERT OR IGNORE INTO prompt_fragment_versions "
                "(fragment_key, version, content, status, created_at) VALUES (?, ?, ?, ?, ?);",
                (fragment_key, new_version, content, status, now),
            )
        return new_version

    def activate_fragment(self, fragment_key: str) -> None:
        """把 fragment 從 draft 切到 active（P5 activate）。"""
        with self._conn:
            self._conn.execute(
                "UPDATE prompt_fragments SET status = 'active', updated_at = ? WHERE fragment_key = ?;",
                (_now(), fragment_key),
            )

    # ── draft 隔離（P5：draft 不影響 active 直到 activate）──────────────────
    def _max_version(self, fragment_key: str) -> int:
        row = self._conn.execute(
            "SELECT MAX(version) AS v FROM prompt_fragment_versions WHERE fragment_key = ?;",
            (fragment_key,)).fetchone()
        return int(row["v"]) if row and row["v"] is not None else 0

    def save_fragment_draft(self, fragment_key: str, content: str) -> int:
        """存一個 draft 版本到 prompt_fragment_versions，**不動 prompt_fragments（active）**。回傳 draft version。"""
        version = self._max_version(fragment_key) + 1
        with self._conn:
            self._conn.execute(
                "INSERT INTO prompt_fragment_versions (fragment_key, version, content, status, created_at) "
                "VALUES (?, ?, ?, 'draft', ?);",
                (fragment_key, version, content, _now()),
            )
        return version

    def get_latest_draft(self, fragment_key: str) -> Optional[dict]:
        row = self._conn.execute(
            "SELECT * FROM prompt_fragment_versions WHERE fragment_key = ? AND status = 'draft' "
            "ORDER BY version DESC LIMIT 1;", (fragment_key,)).fetchone()
        return dict(row) if row else None

    def list_fragment_versions(self, fragment_key: str) -> list[dict]:
        return [dict(r) for r in self._conn.execute(
            "SELECT * FROM prompt_fragment_versions WHERE fragment_key = ? ORDER BY version DESC;",
            (fragment_key,)).fetchall()]

    def activate_fragment_draft(self, fragment_key: str, version: Optional[int] = None) -> Optional[int]:
        """把指定（或最新）draft 版本的內容**提升為 active**：寫進 prompt_fragments + 標該版本 active。

        這是「draft → preview → activate」的 activate 步：在此之前 active 編譯 prompt 不受 draft 影響。
        """
        if version is None:
            d = self.get_latest_draft(fragment_key)
            if d is None:
                return None
            version = d["version"]
        row = self._conn.execute(
            "SELECT content FROM prompt_fragment_versions WHERE fragment_key = ? AND version = ?;",
            (fragment_key, version)).fetchone()
        if row is None:
            return None
        now = _now()
        with self._conn:
            self._conn.execute(
                "UPDATE prompt_fragments SET content = ?, version = ?, status = 'active', updated_at = ? "
                "WHERE fragment_key = ?;",
                (row["content"], version, now, fragment_key),
            )
            self._conn.execute(
                "UPDATE prompt_fragment_versions SET status = 'active' WHERE fragment_key = ? AND version = ?;",
                (fragment_key, version))
        return version

    def set_binding_enabled(self, agent: str, profile: str, fragment_key: str, enabled: bool) -> None:
        with self._conn:
            self._conn.execute(
                "UPDATE agent_prompt_bindings SET enabled = ? "
                "WHERE agent_name = ? AND profile_name = ? AND fragment_key = ?;",
                (1 if enabled else 0, agent, profile, fragment_key),
            )

    # ── context policies ──────────────────────────────────────────────
    def get_context_policy(self, agent: str, profile: Optional[str] = None) -> Optional[dict]:
        profile = profile or self.active_profile()
        row = self._conn.execute(
            "SELECT * FROM agent_context_policies WHERE agent_name = ? AND profile_name = ?;",
            (agent, profile),
        ).fetchone()
        return dict(row) if row else None

    # ── feature flags ─────────────────────────────────────────────────
    def get_flag(self, flag_name: str, profile: Optional[str] = None) -> Optional[int]:
        """回傳 DB 內 flag 值（profile 找不到 → default profile → None）。.env 覆寫不在此處理（見 flags.py）。"""
        for prof in [profile or self.active_profile(), F.DEFAULT_PROFILE]:
            row = self._conn.execute(
                "SELECT value FROM feature_flags WHERE flag_name = ? AND profile_name = ?;",
                (flag_name, prof),
            ).fetchone()
            if row is not None:
                return int(row["value"])
        return None

    def set_flag(self, flag_name: str, value: int, profile: Optional[str] = None) -> None:
        profile = profile or self.active_profile()
        with self._conn:
            self._conn.execute(
                "INSERT INTO feature_flags (flag_name, profile_name, value, updated_at) VALUES (?, ?, ?, ?) "
                "ON CONFLICT(flag_name, profile_name) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at;",
                (flag_name, profile, int(value), _now()),
            )

    # ── run config snapshots（P6 寫入）────────────────────────────────
    def write_run_config_snapshot(self, run_id: str, profile_name: str, agent_name: str,
                                   config_json: dict, compiled_prompt_hash: str,
                                   enabled_fragments: list[str],
                                   compiled_prompt_preview: Optional[str] = None) -> None:
        with self._conn:
            self._conn.execute(
                "INSERT INTO run_config_snapshots "
                "(run_id, profile_name, agent_name, config_json, compiled_prompt_hash, "
                " compiled_prompt_preview, enabled_fragments_json, created_at) "
                "VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
                (run_id, profile_name, agent_name,
                 json.dumps(config_json, ensure_ascii=False, default=str),
                 compiled_prompt_hash, compiled_prompt_preview,
                 json.dumps(enabled_fragments, ensure_ascii=False), _now()),
            )

    def get_run_config_snapshots(self, run_id: str) -> list[dict]:
        rows = self._conn.execute(
            "SELECT * FROM run_config_snapshots WHERE run_id = ? ORDER BY id ASC;", (run_id,)
        ).fetchall()
        return [dict(r) for r in rows]
