"""core.memory.summary — 記憶結構（U14 承重牆基礎件）。

提供三個無限模式記憶命脈的容器：

- BeatWindow：最近 N 個 beat 的滑動視窗；溢出的最舊 beat 由 compactor 折進摘要。
- RollingSummary：散文主線摘要，**有界**（cap_tokens），滿了標記需更激進壓縮。
- FactLedger：硬事實二元組清單，(type, content) 去重，protected 永不刪。

這三者本身不呼叫 LLM——壓縮邏輯在 core.agents.compactor。這裡只負責結構與
不變式（視窗 ≤ size、摘要有界、protected 不滅），讓 30 beat 模擬可被驗證。
"""
from __future__ import annotations

from core.constants import BEAT_WINDOW_SIZE, SUMMARY_TOKEN_CAP
from core.models import LedgerFact


# ─────────────────────────────────────────────────────────────────────────────
# BeatWindow — 滑動視窗
# ─────────────────────────────────────────────────────────────────────────────

class BeatWindow:
    """最近 size 個 beat 的滑動視窗。

    push(beat) 後若超過 size，從頭（最舊）擠出多餘的 beat 並回傳（list）。
    呼叫端（compactor 觸發點）拿被擠出的 beat 去折進滾動摘要。

    不變式：任何時刻 len(items()) <= size。
    """

    def __init__(self, size: int = BEAT_WINDOW_SIZE) -> None:
        if size < 1:
            raise ValueError("BeatWindow size 必須 >= 1")
        self.size = size
        self._items: list = []

    def push(self, beat) -> list:
        """加入一個 beat。回傳因溢出被擠出的最舊 beat（可能多個；通常 0 或 1 個）。"""
        self._items.append(beat)
        evicted: list = []
        while len(self._items) > self.size:
            evicted.append(self._items.pop(0))
        return evicted

    def items(self) -> list:
        """回傳目前視窗內的 beat（淺拷貝，避免外部改動內部清單）。"""
        return list(self._items)

    def __len__(self) -> int:
        return len(self._items)


# ─────────────────────────────────────────────────────────────────────────────
# RollingSummary — 滾動摘要（有界）
# ─────────────────────────────────────────────────────────────────────────────

def estimate_tokens(text: str) -> int:
    """中文 token 粗估（CHECKLIST F7）：用字元數近似。

    中文約 1 字 ≈ 1 token，這裡直接用 len(text) 當粗估上界，足夠驅動壓縮門檻。
    """
    return len(text or "")


class RollingSummary:
    """散文主線摘要，持有單一 str，受 cap_tokens 上限約束。

    update(new_summary) 直接取代為壓縮後的新摘要（compactor 已做合併/濃縮，
    這裡不再串接，否則摘要會無限成長——這正是承重牆的關鍵不變式）。

    estimate_tokens() 回傳目前摘要的粗估 token 數；超 cap 時 needs_recompaction()
    為 True，呼叫端（compactor）應做更激進的壓縮。
    """

    def __init__(self, cap_tokens: int = SUMMARY_TOKEN_CAP) -> None:
        self.cap_tokens = cap_tokens
        self.text: str = ""

    def update(self, new_summary: str) -> None:
        """以 compactor 產出的新摘要**取代**舊摘要（取代語意，非串接）。

        compactor 的職責是把「舊摘要 + 滑出的 beat」濃縮成 new_summary，
        因此這裡只需取代。取代語意保證摘要不隨 beat 數線性成長。
        """
        self.text = new_summary or ""

    def estimate_tokens(self) -> int:
        """目前摘要的粗估 token 數。"""
        return estimate_tokens(self.text)

    def needs_recompaction(self) -> bool:
        """摘要超過 cap → 需要更激進壓縮。"""
        return self.estimate_tokens() > self.cap_tokens

    def __str__(self) -> str:
        return self.text


# ─────────────────────────────────────────────────────────────────────────────
# FactLedger — 硬事實帳本（去重、protected 不刪）
# ─────────────────────────────────────────────────────────────────────────────

class FactLedger:
    """硬事實二元組帳本：list[LedgerFact]，以 (type, content) 為鍵去重。

    - add / merge：同 (type, content) 視為同一條，不重複加入。
    - protected fact：標記為保護的條目永不被移除（伏筆 / anchor / 未揭露線索）。
      protected 由 (type, content) 鍵集合追蹤，與內容本身分離，方便外部標記。
    """

    def __init__(self) -> None:
        self._facts: list[LedgerFact] = []
        self._keys: set[tuple[str, str]] = set()
        self._protected: set[tuple[str, str]] = set()

    @staticmethod
    def _key(fact: LedgerFact) -> tuple[str, str]:
        return (fact.type, fact.content)

    def add(self, fact: LedgerFact, *, protected: bool = False) -> bool:
        """加入一條事實。同 (type, content) 已存在則不重複加入。

        protected=True 時把該鍵標為保護（即使該事實已存在亦補標）。
        回傳 True 表示新加入，False 表示已存在（僅可能補標 protected）。
        """
        key = self._key(fact)
        if protected:
            self._protected.add(key)
        if key in self._keys:
            return False
        self._facts.append(fact)
        self._keys.add(key)
        return True

    def merge(self, facts: list[LedgerFact], *, protected: bool = False) -> int:
        """合併一批事實（去重）。回傳實際新加入的條數。"""
        added = 0
        for f in facts:
            if self.add(f, protected=protected):
                added += 1
        return added

    def mark_protected(self, fact: LedgerFact) -> None:
        """把某條事實標為 protected（不要求其已存在）。"""
        self._protected.add(self._key(fact))

    def is_protected(self, fact: LedgerFact) -> bool:
        return self._key(fact) in self._protected

    def remove(self, fact: LedgerFact) -> bool:
        """移除一條事實；protected 條目拒絕移除（回傳 False）。"""
        key = self._key(fact)
        if key in self._protected:
            return False
        if key not in self._keys:
            return False
        self._facts = [f for f in self._facts if self._key(f) != key]
        self._keys.discard(key)
        return True

    def facts(self) -> list[LedgerFact]:
        """回傳所有事實（淺拷貝清單）。"""
        return list(self._facts)

    def protected_facts(self) -> list[LedgerFact]:
        return [f for f in self._facts if self.is_protected(f)]

    def __len__(self) -> int:
        return len(self._facts)

    def __contains__(self, fact: LedgerFact) -> bool:
        return self._key(fact) in self._keys
