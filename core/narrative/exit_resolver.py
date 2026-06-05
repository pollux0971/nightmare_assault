"""core.narrative.exit_resolver — 離開意圖解析（Player Sovereignty）。

玩家說「離開」時**預設不直接 ending**。分辨意圖；**不確定就問**（ExitOffer），永遠留一個
「明確結束本次調查」的選項，玩家不會被困在「想走走不掉」。

意圖：
- run_ending        明確結束本次調查 / 接受結果 / 放棄目標 / 頭也不回 → 進 EndingGate
- area_transition   只是離開房間 / 區域 → 換場景，遊戲繼續
- temporary_retreat 暫時撤退整理 → 降即時危險，遊戲繼續
- return_to_motive  回頭繼續追目的 → 遊戲繼續
- ambiguous         語意不明（只說「離開 / 找出口」）→ 輸出 ExitOffer（四選一）
- none              非離開行動

零 LLM、規則版。對應 docs/player-sovereignty-principles.md。
"""
from __future__ import annotations

RUN_ENDING = "run_ending"
AREA_TRANSITION = "area_transition"
TEMPORARY_RETREAT = "temporary_retreat"
RETURN_TO_MOTIVE = "return_to_motive"
AMBIGUOUS = "ambiguous"
NONE = "none"

# 順序：先比對「具體」意圖，最後才落到 generic → ambiguous
_END_RUN = ["結束本次調查", "結束調查", "結束這場夢魘", "結束探索", "結束這一切",
            "接受目前結果", "接受結果", "接受目前", "不再回頭", "頭也不回",
            "放棄調查", "放棄尋找", "放棄這次", "我放棄", "就此離開", "永遠離開",
            "離開研究站不再回頭", "結束本次"]
# 注意：避免過短/過泛的詞（如「逃」會誤中「逃避」、「繼續尋找」是繼續調查而非離開）。
_RETURN = ["回頭繼續", "回頭找", "回去找", "回頭尋找", "回頭追", "折返尋找"]
_RETREAT = ["暫時撤退", "退回外面", "喘口氣", "找地方休息", "先離開低頻", "暫時離開",
            "撤退到安全", "先退出", "退到外圍", "暫避"]
_AREA = ["離開這個房間", "離開房間", "離開辦公室", "離開這個區域", "離開目前區域", "離開此處",
         "離開這層", "離開走廊", "穿過門", "走出房間", "換個地方"]
_GENERIC_EXIT = ["離開", "出口", "逃出", "逃離", "走出去", "找出口", "尋找出口", "脫身", "逃出生天"]


def resolve_exit_intent(text: str) -> str:
    t = text or ""
    if any(p in t for p in _END_RUN):
        return RUN_ENDING
    if any(p in t for p in _RETURN):
        return RETURN_TO_MOTIVE
    if any(p in t for p in _RETREAT):
        return TEMPORARY_RETREAT
    if any(p in t for p in _AREA):
        return AREA_TRANSITION
    if any(p in t for p in _GENERIC_EXIT):
        return AMBIGUOUS
    return NONE


def exit_offer_options(motive_label: str = "你來這裡的目的") -> list[dict]:
    """ExitOffer 四選一（labels 會被 resolve_exit_intent 正確分類回對應意圖）。"""
    return [
        {"id": "leave_area", "label": "離開目前區域，繼續探索"},
        {"id": "temporary_retreat", "label": "暫時撤退到安全處整理線索"},
        {"id": "end_run", "label": "結束本次調查，接受目前結果"},
        {"id": "return_to_motive", "label": f"回頭繼續尋找{motive_label}"},
    ]


def build_exit_offer_decision_point(base_dp, motive_label: str = "你來這裡的目的"):
    """把當前 beat 的 decision_point 換成 ExitOffer（保留 story 敘事，只換選項）。"""
    from core.models import DecisionPoint, Option
    opts = exit_offer_options(motive_label)
    base_recap = (getattr(base_dp, "situation_recap", "") or "").strip()
    frame = "你看見一條可以離開的路。要怎麼做，由你決定。"
    recap = (base_recap + "\n" + frame) if base_recap else frame
    return DecisionPoint(
        situation_recap=recap, decision_type="action",
        suggested_options=[Option(text=o["label"], tone="cautious") for o in opts],
        free_input_hint="或自由描述你想做的事",
        beat_meta=getattr(base_dp, "beat_meta", None) or {"beat_number": 0},
        is_narration_only=False)
