"""core.narrative.beat_rendering — Beat 渲染預算 + debug（patch v0.7 P3）。

量測每個 beat 的「類型 + 目標字數 + 實際字數 + 是否過短 + 連續過短」並曝進 observation。
**只量測、不修復**——P4 short-beat soft repair 預設不啟用；等 debug 顯示確實連續過短再開。

對應 docs/03-beat-rendering-policy.md。
"""
from __future__ import annotations

# beat 類型 → 目標字數區間（軟預算，非硬性塞字）
OPENING = "opening"
NEW_AREA_ENTRY = "new_area_entry"
NPC_FIRST_INTRO = "npc_first_intro"
NORMAL_EXPLORATION = "normal_exploration"
OBJECT_INSPECTION = "object_inspection"
NPC_CHAT = "npc_chat"
REVIEW_MODE = "review_mode"
DANGER_BEAT = "danger_beat"
ENDING = "ending"
SYSTEM_FEEDBACK = "system_feedback"

# 目標字數區間：max 維持原「理想」；min 校準到**真 LLM 實測**水準——
# too_short 應只標出「真的薄成狀態日誌」(<~140) 的 beat，而非合理的 ~200 字段落（見 step4 驗證）。
BUDGETS = {
    OPENING: (350, 900),
    NEW_AREA_ENTRY: (180, 650),
    NPC_FIRST_INTRO: (180, 550),
    NORMAL_EXPLORATION: (140, 420),
    OBJECT_INSPECTION: (120, 320),
    NPC_CHAT: (110, 380),
    REVIEW_MODE: (50, 260),
    DANGER_BEAT: (150, 450),
    ENDING: (280, 800),
    SYSTEM_FEEDBACK: (40, 160),
}

# 不納入「一般 beat 連續過短」統計的類型（review/系統回饋本就可短）
_EXCLUDED_FROM_STREAK = {REVIEW_MODE, SYSTEM_FEEDBACK}


def count_cjk_chars(text: str) -> int:
    """數實際內容字數（去空白）。敘事以中文為主，非空白字元數已足夠近似。"""
    return len("".join(ch for ch in (text or "") if not ch.isspace()))


def classify_beat_type(*, review_locked: bool = False, ended: bool = False,
                       is_opening: bool = False, action_class: str | None = None,
                       scene_changed: bool = False, npc_first_intro: bool = False,
                       danger_delta: int = 0) -> str:
    """依本 beat 的狀態分類 beat_type（優先序：結局 > 開場 > review > 危險 > 首見NPC >
    換區 > 物件檢查 > 一般探索）。"""
    if ended:
        return ENDING
    if is_opening:
        return OPENING
    if review_locked:
        return REVIEW_MODE
    if danger_delta and danger_delta > 0:
        return DANGER_BEAT
    if npc_first_intro:
        return NPC_FIRST_INTRO
    if scene_changed:
        return NEW_AREA_ENTRY
    if action_class == "object_inspection":
        return OBJECT_INSPECTION
    return NORMAL_EXPLORATION


def evaluate_beat_rendering(narrative: str, beat_type: str, *,
                            prev_short_streak: int = 0) -> dict:
    """量測單一 beat 的渲染指標。回 P3 debug dict（含更新後 short_streak）。**不修復。**"""
    lo, hi = BUDGETS.get(beat_type, BUDGETS[NORMAL_EXPLORATION])
    actual = count_cjk_chars(narrative)
    excluded = beat_type in _EXCLUDED_FROM_STREAK
    too_short = (actual < lo) and not excluded
    short_streak = (prev_short_streak + 1) if too_short else 0
    return {
        "beat_type": beat_type,
        "target_min_chars": lo,
        "target_max_chars": hi,
        "actual_chars": actual,
        "too_short": too_short,
        "short_streak": short_streak,
        "repair_attempted": False,        # P4 未啟用——永遠 False
    }
