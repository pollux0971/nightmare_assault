"""core.narrative.answer_debt — 重複提問答債（NR2，敘事控制 v0.2）。

玩家反覆直問時，系統常以氛圍敷衍（卡關感）。本模組追蹤「已問問題 × 是否已償還」，
debt≥2 時強制下一相關回應給部分答 / 具體線索 / 指向證據 / 具理由拒答。

零 LLM、規則版。對應 dev/CONTRACTS.md §十四（AnswerDebt）。
"""
from __future__ import annotations

from dataclasses import dataclass, field

# 問題分類（docs/04）：category → 觸發關鍵詞
_CATEGORY_MARKERS = {
    "identity_question": ["你是誰", "是誰", "誰啊", "你叫", "是不是", "真的是"],
    "mechanism_question": ["432", "頻率", "為什麼", "怎麼回事", "原理", "校準", "整點", "報時", "聲音", "規則"],
    "threat_question": ["危險", "他們是", "怪", "會死", "威脅", "追", "在外面"],
    "location_question": ["在哪", "哪裡", "出口", "怎麼出去", "林晨", "弟弟", "他人呢"],
    "action_question": ["該怎麼辦", "怎麼做", "下一步", "該做什麼", "怎麼辦"],
}
# topic 細分（讓「432.7」與「林晨」分開記債）
_TOPIC_MARKERS = {
    "432": "signal", "頻率": "signal", "整點": "signal", "報時": "signal",
    "林晨": "linchen", "弟弟": "linchen",
    "出口": "exit", "出去": "exit",
}


def classify_question(text: str) -> str | None:
    """回傳 'category:topic'（如 mechanism_question:signal）；非問題回 None。"""
    t = (text or "").strip()
    if not t:
        return None
    is_question = ("?" in t or "？" in t or
                   any(w in t for ws in _CATEGORY_MARKERS.values() for w in ws))
    if not is_question:
        return None
    category = "action_question"
    for cat, markers in _CATEGORY_MARKERS.items():
        if any(m in t for m in markers):
            category = cat
            break
    topic = "general"
    for marker, top in _TOPIC_MARKERS.items():
        if marker in t:
            topic = top
            break
    return f"{category}:{topic}"


@dataclass
class AnswerDebtTracker:
    """逐 topic 記答債：0 無 / 1 可迴避 / 2 須部分答 / 3 須具體線索或具理由拒答。"""
    debts: dict[str, int] = field(default_factory=dict)

    def register_question(self, topic_key: str) -> int:
        self.debts[topic_key] = self.debts.get(topic_key, 0) + 1
        return self.debts[topic_key]

    def register_answer(self, topic_key: str, answer_status: str) -> None:
        if answer_status in ("answered", "partial"):
            self.debts[topic_key] = 0
        elif answer_status == "refused":
            self.debts[topic_key] = max(0, self.debts.get(topic_key, 0) - 1)

    def level(self, topic_key: str) -> int:
        return self.debts.get(topic_key, 0)

    def requires_payoff(self, topic_key: str) -> bool:
        return self.debts.get(topic_key, 0) >= 2

    def to_dict(self) -> dict:
        return dict(self.debts)

    @classmethod
    def from_dict(cls, data: dict | None) -> "AnswerDebtTracker":
        return cls(debts=dict(data or {}))


# debt≥2 時要求的償還方式（注入 story / npc-chat context）
def payoff_obligation(topic_key: str, level: int) -> str:
    return (f"玩家已重複追問「{topic_key}」（債務等級 {level}）。"
            "這次回應必須至少給出：一個部分答案、一條具體線索、指向證據的方向，或具理由的明確拒答——"
            "不得再用純氛圍敷衍。")
