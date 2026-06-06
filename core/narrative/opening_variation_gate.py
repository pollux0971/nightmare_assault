"""core.narrative.opening_variation_gate — 開場輸出違規偵測（補丁 v0.8 P6）。

掃 StoryAgent 的開場 narrative，檢查它有沒有偷換回被 contract 擋掉的素材：

  * `forbidden_literal`：出現了 cooldown 擋下的具體字串（紙條 / 林晨 / …）。
  * `forbidden_archetype`：用了被擋的 archetype（例如 missing_person 被擋卻仍寫找失蹤的人）。
  * `message_medium_mismatch`：contract 指定 corrupted_log，卻寫成手寫紙條/便條/手寫留言。

設計邊界：這是**只讀的偵測器**，不改世界、不推 reveal、不收束劇情。違規後的 repair/fallback
策略由 orchestrator_loop 決定（repair once → deterministic fallback）。

別名群（patch docs/07 風險 2）：MVP 用 literal exact + 常見別稱，不做繁簡/embedding。
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable


# missing_person archetype 的語意線索（被擋時不得出現）。**繁簡都收**——真實模型常輸出簡體，
# 只放繁體會讓 gate 在簡體輸出上漏判（實測 deepseek 輸出簡體「失踪/寻找」會繞過）。
_MISSING_PERSON_CUES = (
    "失蹤", "失踪", "找人", "尋找他", "寻找他", "尋找她", "寻找她",
    "尋人", "寻人", "下落不明", "失蹤的", "失踪的", "失聯", "失联",
)

# handwritten_note 的別稱群（非 handwritten_note medium 時不得出現）。**繁簡都收。**
_HANDNOTE_CUES = (
    "紙條", "纸条", "便條", "便条", "手寫留言", "手写留言",
    "字條", "字条", "便籤", "便签", "手寫紙條", "手写纸条",
)

# archetype → 該 archetype 被擋時不得出現的語意線索群。
_ARCHETYPE_CUES: dict[str, tuple[str, ...]] = {
    "missing_person": _MISSING_PERSON_CUES,
    "handwritten_note": _HANDNOTE_CUES,
}


@dataclass(frozen=True)
class OpeningViolation:
    type: str
    value: str


def check_opening_output(
    text: str,
    *,
    forbidden_literals: Iterable[str],
    forbidden_archetypes: Iterable[str],
    expected_message_medium: str | None = None,
) -> list[OpeningViolation]:
    """回傳 text 中所有違規（空 list = 合規）。"""
    text = text or ""
    violations: list[OpeningViolation] = []

    for literal in forbidden_literals:
        if literal and literal in text:
            violations.append(OpeningViolation("forbidden_literal", literal))

    forbidden_set = set(forbidden_archetypes)
    for archetype in forbidden_set:
        cues = _ARCHETYPE_CUES.get(archetype)
        if cues and any(c in text for c in cues):
            violations.append(OpeningViolation("forbidden_archetype", archetype))

    # 指定了非手寫的 medium，卻寫成手寫紙條家族 → mismatch。
    if expected_message_medium and expected_message_medium != "handwritten_note":
        if any(c in text for c in _HANDNOTE_CUES):
            violations.append(OpeningViolation("message_medium_mismatch", "handwritten_note"))

    return violations


def has_violation(violations: Iterable[OpeningViolation]) -> bool:
    return bool(list(violations))


def violations_to_debug(violations: Iterable[OpeningViolation]) -> list[dict]:
    return [{"type": v.type, "value": v.value} for v in violations]
