"""core.narrative.evidence_extractor — 調查保底證據抽取（HB1，Runtime Hard-Gate v0.3.1）。

修「玩家做了有意義調查、story 也吐出具體新資訊，但本 beat reveal 無變化 → 最後仍 0/X」。
條件齊備時保底產一個 hinted EvidenceEvent；map 不到 truth_id → source=fallback、計 unmapped。

零 LLM、規則版。對應 dev/CONTRACTS.md §十五（StoryEvidenceExtractor）。
"""
from __future__ import annotations

import re

from core.narrative.revelation import EvidenceEvent

# 調查型動詞（玩家在「主動查」而非「移動/逃」）
INVESTIGATION_VERBS = ["檢查", "查看", "翻找", "詢問", "追問", "比對", "拆", "閱讀",
                       "研究", "觀察", "搜", "看", "讀", "問", "查", "拾起", "撿"]
# 具體新資訊樣式（頻率/編號/紀錄/名單/儀器…）
CONCRETE_PATTERNS = [
    r"\d+(?:\.\d+)?\s*(?:Hz|赫茲|秒|分鐘|號|層|室)",
    r"B-?\d+",
    r"登入|紀錄|記錄|值班表|名單|檔案|錄音|頻率|控制台|儀器|鑰匙|地圖|字條|紙條|日誌|照片",
]


class StoryEvidenceExtractor:
    def __init__(self, truth_keyword_index: dict[str, list[str]] | None = None):
        self.truth_keyword_index = truth_keyword_index or {}

    def is_investigation(self, action: str) -> bool:
        return any(v in (action or "") for v in INVESTIGATION_VERBS)

    def has_concrete_info(self, narrative: str) -> bool:
        return any(re.search(p, narrative or "", re.IGNORECASE) for p in CONCRETE_PATTERNS)

    def map_truth_id(self, text: str) -> str | None:
        for truth_id, keywords in self.truth_keyword_index.items():
            if any(k and k in text for k in keywords):
                return truth_id
        return None

    def extract(self, *, beat: int, action: str, narrative: str,
                reveal_changed: bool) -> list[EvidenceEvent]:
        """reveal 本 beat 已變化 / 非調查 / 無具體資訊 → 不產。否則保底 hinted evidence。"""
        if reveal_changed:
            return []
        if not self.is_investigation(action):
            return []
        if not self.has_concrete_info(narrative):
            return []
        truth_id = self.map_truth_id((action or "") + "\n" + (narrative or ""))
        return [EvidenceEvent(
            evidence_id=f"ev.story.{beat}.fallback",
            source="story" if truth_id else "fallback",
            truth_id=truth_id,
            evidence_strength=0.30 if truth_id else 0.0,   # 對得上真相才推進；對不上只記錄
            max_level="hinted",                            # 保底僅 hinted（弱證據）
            surface_text=self._clip(narrative),
            debug_reason="meaningful investigation produced concrete narrative info "
                         "without reveal update")]

    @staticmethod
    def _clip(text: str, n: int = 160) -> str:
        return (text or "").strip().replace("\n", " ")[:n]
