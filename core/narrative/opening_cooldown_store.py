"""core.narrative.opening_cooldown_store — CooldownLedger 跨 run 持久化（補丁 v0.8 P2）。

cooldown 要跨「整場遊戲」生效（這一局用過紙條，下一局才該避開），所以 ledger 不能只活在記憶體。
本模組把 ledger + 一個單調遞增的 run 計數器存進一張 **additive** SQLite 表，掛在既有 DB 連線上：

    opening_variation_state(key PRIMARY KEY, value TEXT)   # key: "run_counter" / "ledger"(json)

只 `CREATE TABLE IF NOT EXISTS`（冪等、不動既有 schema、舊存檔可讀）。任一步失敗都 graceful：
退回純記憶體 ledger（B8——cooldown 失效頂多重複，不該讓開局崩）。
"""
from __future__ import annotations

import json
import logging
import sqlite3

from core.narrative.opening_variation import CooldownLedger

log = logging.getLogger("nightmare.opening_cooldown")

_DDL = """
CREATE TABLE IF NOT EXISTS opening_variation_state (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
"""


class OpeningVariationStore:
    """掛在既有 DB 連線上的 cooldown 持久化（additive、冪等）。"""

    def __init__(self, conn: sqlite3.Connection, ensure: bool = True):
        self._conn = conn
        if ensure:
            self._ensure()

    def _ensure(self) -> None:
        with self._conn:
            self._conn.execute(_DDL)

    def _get(self, key: str) -> str | None:
        row = self._conn.execute(
            "SELECT value FROM opening_variation_state WHERE key = ?;", (key,)
        ).fetchone()
        if row is None:
            return None
        # row 可能是 sqlite3.Row 或 tuple
        try:
            return row["value"]
        except (TypeError, IndexError, KeyError):
            return row[0]

    def _set(self, key: str, value: str) -> None:
        with self._conn:
            self._conn.execute(
                "INSERT INTO opening_variation_state (key, value) VALUES (?, ?) "
                "ON CONFLICT(key) DO UPDATE SET value = excluded.value;",
                (key, value),
            )

    # ── run 計數器（每次「生成一份開場契約」遞增）──────────────────────────────
    def next_run_index(self) -> int:
        """遞增並回傳新的 run 序號（從 1 起算）。"""
        cur = 0
        raw = self._get("run_counter")
        if raw is not None:
            try:
                cur = int(raw)
            except (TypeError, ValueError):
                cur = 0
        nxt = cur + 1
        self._set("run_counter", str(nxt))
        return nxt

    def current_run_index(self) -> int:
        raw = self._get("run_counter")
        try:
            return int(raw) if raw is not None else 0
        except (TypeError, ValueError):
            return 0

    # ── ledger 讀寫 ────────────────────────────────────────────────────────
    def load_ledger(self) -> CooldownLedger:
        raw = self._get("ledger")
        if not raw:
            return CooldownLedger()
        try:
            return CooldownLedger.from_dict(json.loads(raw))
        except Exception as e:                       # 壞資料不該擋開局
            log.warning("load_ledger failed, fresh ledger: %s", e)
            return CooldownLedger()

    def save_ledger(self, ledger: CooldownLedger) -> None:
        self._set("ledger", json.dumps(ledger.to_dict(), ensure_ascii=False))

    def reset(self) -> None:
        """admin：清空 ledger 與計數器（patch docs/07 風險 4：cooldown 過強可重置）。"""
        with self._conn:
            self._conn.execute("DELETE FROM opening_variation_state;")


def open_store(db) -> "OpeningVariationStore | None":
    """從 Database 包裝取連線建 store；失敗回 None（caller 退記憶體 ledger）。"""
    try:
        conn = getattr(db, "connection", None)
        if conn is None:
            conn = getattr(db, "_conn", None)
        if conn is None:
            return None
        return OpeningVariationStore(conn)
    except Exception as e:
        log.warning("open_store failed: %s", e)
        return None
