"""core.persistence.db — SQLite 持久化層（schema v1）。

標準庫 sqlite3，不依賴外部套件。
"""

import datetime
import sqlite3
from typing import Optional


_SCHEMA_DDL = """
CREATE TABLE IF NOT EXISTS schema_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runs (
    run_id           TEXT PRIMARY KEY,
    title            TEXT,
    theme            TEXT,
    difficulty       TEXT,
    created_at       TEXT,
    updated_at       TEXT,
    current_beat     INTEGER DEFAULT 0,
    current_location TEXT,
    status           TEXT DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS beats (
    id                           INTEGER PRIMARY KEY,
    run_id                       TEXT    NOT NULL,
    beat_number                  INTEGER NOT NULL,
    narrative                    TEXT,
    decision_json                TEXT,
    rolling_summary_snapshot     TEXT,
    blackboard_snapshot_json     TEXT,
    is_narration_only            INTEGER DEFAULT 0,
    created_at                   TEXT
);

CREATE TABLE IF NOT EXISTS npc_states (
    id               INTEGER PRIMARY KEY,
    run_id           TEXT NOT NULL,
    beat_number      INTEGER NOT NULL,
    npc_name         TEXT NOT NULL,
    state_json       TEXT,
    hidden_state_json TEXT
);

CREATE TABLE IF NOT EXISTS inventory_snapshots (
    id             INTEGER PRIMARY KEY,
    run_id         TEXT NOT NULL,
    beat_number    INTEGER NOT NULL,
    inventory_json TEXT
);

CREATE TABLE IF NOT EXISTS chat_logs (
    id         INTEGER PRIMARY KEY,
    run_id     TEXT NOT NULL,
    npc_name   TEXT,
    beat_number INTEGER,
    role       TEXT,
    content    TEXT,
    created_at TEXT
);

CREATE TABLE IF NOT EXISTS save_points (
    id          INTEGER PRIMARY KEY,
    run_id      TEXT    NOT NULL,
    beat_number INTEGER NOT NULL,
    label       TEXT,
    created_at  TEXT
);

CREATE TABLE IF NOT EXISTS offstage_fates (
    id                INTEGER PRIMARY KEY,
    run_id            TEXT NOT NULL,
    npc_name          TEXT NOT NULL,
    fate_type         TEXT,
    fate_narrative    TEXT,
    carried_fragment  TEXT,
    revealed          INTEGER DEFAULT 0,
    created_at        TEXT
);

CREATE TABLE IF NOT EXISTS llm_traces (
    id            INTEGER PRIMARY KEY,
    run_id        TEXT    NOT NULL,
    beat_number   INTEGER,
    agent         TEXT,
    model         TEXT,
    prompt_hash   TEXT,
    input_tokens  INTEGER,
    output_tokens INTEGER,
    latency_ms    INTEGER,
    success       INTEGER,
    error         TEXT,
    created_at    TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_beats_run_beat
    ON beats(run_id, beat_number);

CREATE INDEX IF NOT EXISTS idx_npc_states_run_beat
    ON npc_states(run_id, beat_number);

CREATE INDEX IF NOT EXISTS idx_inventory_run_beat
    ON inventory_snapshots(run_id, beat_number);

CREATE INDEX IF NOT EXISTS idx_chat_logs_run_npc
    ON chat_logs(run_id, npc_name);

CREATE INDEX IF NOT EXISTS idx_llm_traces_run_beat
    ON llm_traces(run_id, beat_number);
"""


def _now() -> str:
    return datetime.datetime.now().isoformat()


class Database:
    """SQLite 持久化封裝，所有操作走同一連線。"""

    def __init__(self, path: str = ":memory:") -> None:
        self._conn = sqlite3.connect(path, check_same_thread=False)
        self._conn.row_factory = sqlite3.Row
        self._conn.execute("PRAGMA journal_mode=WAL;")
        self._conn.execute("PRAGMA foreign_keys=ON;")
        self._init_schema()

    # ------------------------------------------------------------------
    # 初始化
    # ------------------------------------------------------------------

    def _init_schema(self) -> None:
        """若尚未初始化則建立所有 table/index 並寫入 schema_version。"""
        with self._conn:
            self._conn.executescript(_SCHEMA_DDL)
            # 只在不存在時才寫入 schema_version（冪等）
            self._conn.execute(
                "INSERT OR IGNORE INTO schema_meta (key, value) VALUES ('schema_version', '1');"
            )
        # ── 配置中心 additive migration（階段 P / P1）──────────────────────
        # additive only：只 CREATE IF NOT EXISTS + INSERT OR IGNORE，既有存檔不受影響。
        # 失敗不可拖垮既有 DB 開啟（B8 graceful degradation）。
        try:
            from core.config.schema import init_config_schema, seed_defaults
            init_config_schema(self._conn)
            seed_defaults(self._conn)
        except Exception:  # pragma: no cover - 配置層失敗不影響核心持久化
            import logging
            logging.getLogger("nightmare.db").warning(
                "config-center additive migration skipped", exc_info=True)

    # ------------------------------------------------------------------
    # 配置中心存取（階段 P）
    # ------------------------------------------------------------------

    @property
    def connection(self) -> sqlite3.Connection:
        """暴露底層連線（配置中心 ConfigStore 掛同一條連線用）。"""
        return self._conn

    def config_store(self):
        """回傳掛在本 DB 連線上的 ConfigStore（配置表已於 _init_schema 建好）。"""
        from core.config.schema import ConfigStore
        return ConfigStore(self._conn, ensure=False)

    # ------------------------------------------------------------------
    # meta
    # ------------------------------------------------------------------

    def schema_version(self) -> str:
        row = self._conn.execute(
            "SELECT value FROM schema_meta WHERE key = 'schema_version';"
        ).fetchone()
        if row is None:
            raise RuntimeError("schema_meta 表尚未初始化")
        return row["value"]

    # ------------------------------------------------------------------
    # runs
    # ------------------------------------------------------------------

    def create_run(
        self,
        run_id: str,
        title: Optional[str] = None,
        theme: Optional[str] = None,
        difficulty: Optional[str] = None,
    ) -> None:
        now = _now()
        with self._conn:
            self._conn.execute(
                """
                INSERT INTO runs (run_id, title, theme, difficulty, created_at, updated_at, status)
                VALUES (?, ?, ?, ?, ?, ?, 'active');
                """,
                (run_id, title, theme, difficulty, now, now),
            )

    def list_runs(self) -> list[dict]:
        """回傳所有 run 的摘要欄位（list_saves 入口）。"""
        rows = self._conn.execute(
            "SELECT run_id, current_beat, current_location, updated_at FROM runs ORDER BY updated_at DESC;"
        ).fetchall()
        return [dict(r) for r in rows]

    def _touch_run(self, run_id: str, beat_number: int) -> None:
        """更新 runs 的 current_beat / updated_at（每次寫 beat 時呼叫）。"""
        self._conn.execute(
            "UPDATE runs SET current_beat = MAX(current_beat, ?), updated_at = ? WHERE run_id = ?;",
            (beat_number, _now(), run_id),
        )

    # ------------------------------------------------------------------
    # beats
    # ------------------------------------------------------------------

    def save_beat(
        self,
        run_id: str,
        beat_number: int,
        narrative: Optional[str],
        decision_json: Optional[str],
        rolling_summary_snapshot: Optional[str],
        blackboard_snapshot_json: Optional[str],
        is_narration_only: bool = False,
    ) -> None:
        """Upsert：同 run_id + beat_number 則覆寫（保留唯一 index）。"""
        now = _now()
        with self._conn:
            self._conn.execute(
                """
                INSERT INTO beats
                    (run_id, beat_number, narrative, decision_json,
                     rolling_summary_snapshot, blackboard_snapshot_json,
                     is_narration_only, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                ON CONFLICT(run_id, beat_number) DO UPDATE SET
                    narrative                = excluded.narrative,
                    decision_json            = excluded.decision_json,
                    rolling_summary_snapshot = excluded.rolling_summary_snapshot,
                    blackboard_snapshot_json = excluded.blackboard_snapshot_json,
                    is_narration_only        = excluded.is_narration_only,
                    created_at               = excluded.created_at;
                """,
                (
                    run_id,
                    beat_number,
                    narrative,
                    decision_json,
                    rolling_summary_snapshot,
                    blackboard_snapshot_json,
                    1 if is_narration_only else 0,
                    now,
                ),
            )
            self._touch_run(run_id, beat_number)

    def load_beat(self, run_id: str, beat_number: int) -> Optional[dict]:
        """回傳該 beat 完整列；找不到則回傳 None。"""
        row = self._conn.execute(
            "SELECT * FROM beats WHERE run_id = ? AND beat_number = ?;",
            (run_id, beat_number),
        ).fetchone()
        return dict(row) if row is not None else None

    # ------------------------------------------------------------------
    # inventory snapshots（UB5：道具庫隨 beat 快照保存）
    # ------------------------------------------------------------------

    def save_inventory_snapshot(self, run_id: str, beat_number: int, inventory_json: str) -> None:
        """每 beat 存一份道具庫快照（含 held_by/is_key_item 等完整欄位，供回溯/讀檔）。"""
        with self._conn:
            self._conn.execute(
                "INSERT INTO inventory_snapshots (run_id, beat_number, inventory_json) VALUES (?, ?, ?);",
                (run_id, beat_number, inventory_json),
            )

    def load_inventory_snapshot(self, run_id: str, beat_number: int) -> Optional[list]:
        """讀回某 beat 的道具庫快照（找不到回 None）。"""
        row = self._conn.execute(
            "SELECT inventory_json FROM inventory_snapshots WHERE run_id = ? AND beat_number = ? "
            "ORDER BY id DESC LIMIT 1;",
            (run_id, beat_number),
        ).fetchone()
        if row is None:
            return None
        import json as _json
        try:
            return _json.loads(row["inventory_json"])
        except Exception:
            return None

    # ------------------------------------------------------------------
    # save_points
    # ------------------------------------------------------------------

    def add_save_point(self, run_id: str, beat_number: int, label: str) -> None:
        now = _now()
        with self._conn:
            self._conn.execute(
                "INSERT INTO save_points (run_id, beat_number, label, created_at) VALUES (?, ?, ?, ?);",
                (run_id, beat_number, label, now),
            )

    def list_save_points(self, run_id: str) -> list[dict]:
        rows = self._conn.execute(
            "SELECT * FROM save_points WHERE run_id = ? ORDER BY beat_number ASC;",
            (run_id,),
        ).fetchall()
        return [dict(r) for r in rows]

    # ------------------------------------------------------------------
    # chat_logs（MC1：NPC 聊天完整紀錄，cold）
    # ------------------------------------------------------------------

    def add_chat_log(self, run_id: str, npc_name: str, beat_number: int,
                     role: str, content: str) -> None:
        """追加一則聊天（role: 'player' | 'npc'）。"""
        with self._conn:
            self._conn.execute(
                "INSERT INTO chat_logs (run_id, npc_name, beat_number, role, content, created_at) "
                "VALUES (?, ?, ?, ?, ?, ?);",
                (run_id, npc_name, beat_number, role, content, _now()),
            )

    def load_chat_logs(self, run_id: str, npc_name: Optional[str] = None) -> list[dict]:
        """讀回聊天紀錄（可選某 NPC），依時間排序。"""
        if npc_name is None:
            rows = self._conn.execute(
                "SELECT * FROM chat_logs WHERE run_id = ? ORDER BY id ASC;", (run_id,)).fetchall()
        else:
            rows = self._conn.execute(
                "SELECT * FROM chat_logs WHERE run_id = ? AND npc_name = ? ORDER BY id ASC;",
                (run_id, npc_name)).fetchall()
        return [dict(r) for r in rows]

    # ------------------------------------------------------------------
    # offstage_fates（MC5：離場命運隱藏紀錄，重逢才揭曉）
    # ------------------------------------------------------------------

    def add_offstage_fate(self, run_id: str, npc_name: str, fate_type: str,
                          fate_narrative: str, carried_fragment: Optional[str]) -> None:
        """存一筆隱藏命運（對玩家不可見，直到重逢/搜屍揭曉）。"""
        with self._conn:
            self._conn.execute(
                "INSERT INTO offstage_fates (run_id, npc_name, fate_type, fate_narrative, "
                "carried_fragment, revealed, created_at) VALUES (?, ?, ?, ?, ?, 0, ?);",
                (run_id, npc_name, fate_type, fate_narrative, carried_fragment, _now()),
            )

    def get_offstage_fate(self, run_id: str, npc_name: str) -> Optional[dict]:
        row = self._conn.execute(
            "SELECT * FROM offstage_fates WHERE run_id = ? AND npc_name = ? ORDER BY id DESC LIMIT 1;",
            (run_id, npc_name)).fetchone()
        return dict(row) if row else None

    def mark_fate_revealed(self, run_id: str, npc_name: str) -> None:
        with self._conn:
            self._conn.execute(
                "UPDATE offstage_fates SET revealed = 1 WHERE run_id = ? AND npc_name = ?;",
                (run_id, npc_name))

    # ------------------------------------------------------------------
    # llm_traces
    # ------------------------------------------------------------------

    def write_llm_trace(
        self,
        run_id: str,
        beat_number: Optional[int],
        agent: Optional[str],
        model: Optional[str],
        prompt_hash: Optional[str],
        input_tokens: Optional[int],
        output_tokens: Optional[int],
        latency_ms: Optional[int],
        success: bool,
        error: Optional[str] = None,
    ) -> None:
        now = _now()
        with self._conn:
            self._conn.execute(
                """
                INSERT INTO llm_traces
                    (run_id, beat_number, agent, model, prompt_hash,
                     input_tokens, output_tokens, latency_ms, success, error, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
                """,
                (
                    run_id,
                    beat_number,
                    agent,
                    model,
                    prompt_hash,
                    input_tokens,
                    output_tokens,
                    latency_ms,
                    1 if success else 0,
                    error,
                    now,
                ),
            )

    def llm_cost_summary(self, run_id: str) -> dict:
        """加總該 run 所有 llm_traces 的 input/output tokens（成本監控用）。"""
        row = self._conn.execute(
            """
            SELECT
                COUNT(*)              AS call_count,
                COALESCE(SUM(input_tokens),  0) AS total_input_tokens,
                COALESCE(SUM(output_tokens), 0) AS total_output_tokens
            FROM llm_traces
            WHERE run_id = ?;
            """,
            (run_id,),
        ).fetchone()
        return dict(row) if row is not None else {"call_count": 0, "total_input_tokens": 0, "total_output_tokens": 0}

    # ------------------------------------------------------------------
    # 資源管理
    # ------------------------------------------------------------------

    def close(self) -> None:
        self._conn.close()

    def __enter__(self) -> "Database":
        return self

    def __exit__(self, *_) -> None:
        self.close()
