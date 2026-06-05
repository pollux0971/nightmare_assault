"""core.narrative.exploration_mode — ExplorationMode / ReviewMode Lock（撤離鎖）。

Player Sovereignty 缺口修補：玩家撤到安全區整理線索時，系統必須**停止自動調查推進**
（不推場景、不升壓、不發 reveal、不登記新 object），直到玩家**明確**選擇回去研究站或結束。

四個模式（持久化在 game_meta.exploration_mode，跨 beat 黏著）：
  active_exploration       正常調查推進（預設）
  temporary_retreat        撤到外面/安全區（不結束；停止自動推進）
  review_mode              在安全區整理既有線索（只生 notes，不新增 fact/truth/object）
  campaign_end_requested   玩家明確結束本次調查 → 進 EndingGate

零 LLM、規則版。對應 15-player-sovereignty.md（撤離鎖）。
"""
from __future__ import annotations

from core.world.model import SAFE_ZONE_AREA_ID

ACTIVE_EXPLORATION = "active_exploration"
TEMPORARY_RETREAT = "temporary_retreat"
REVIEW_MODE = "review_mode"
CAMPAIGN_END_REQUESTED = "campaign_end_requested"

# 進 review/retreat 即上「撤離鎖」的模式集
LOCKED_MODES = (TEMPORARY_RETREAT, REVIEW_MODE)

# ── 玩家措辭 → 模式（先比對「明確再入/結束」，再比對「撤退/整理」）─────────────
# 明確「返回現場 / 重新進入」才解鎖回 active（theme-agnostic；不寫死特定地名）
_REENTER = ["返回現場", "回到現場", "回現場", "返回", "回去", "重新進入", "再次進入",
            "再進去", "重新回到", "重新展開調查", "回到調查", "繼續深入", "繼續往裡",
            "繼續調查", "回到裡面", "回去裡面", "回到原處", "返回原處", "重新進去",
            # 相容：舊主題（研究站）措辭仍可解
            "回去研究站", "回到研究站", "回研究站", "返回站內", "回到站內"]
# 整理/盤點/不新增調查/不碰真相 → review_mode
_REVIEW = ["整理線索", "整理筆記", "盤點線索", "回顧線索", "梳理線索", "整理一下線索",
           "根據已知", "不新增調查", "不再深入", "不碰真相", "不追真相", "暫停調查",
           "整理思緒", "重新理一遍"]
# 退到外面/撤退 → temporary_retreat
_RETREAT = ["退到外面", "退出去", "到外面", "撤到外面", "退回外面", "到站外", "退到安全",
            "暫時撤退", "先離開到外面", "退出研究站", "撤退到安全"]
# 不解鎖（即使含「回研究站」子字串也視為仍鎖）的否定前綴
_NEG_PREFIX = ("不", "別", "拒", "非", "沒")


def _says_reenter(text: str) -> bool:
    """明確再入研究站（避開「不回研究站」這類否定）。"""
    t = text or ""
    for p in _REENTER:
        i = t.find(p)
        while i != -1:
            prev = t[i - 1] if i > 0 else ""
            if prev not in _NEG_PREFIX:
                return True
            i = t.find(p, i + 1)
    return False


def resolve_mode(text: str, current_mode: str = ACTIVE_EXPLORATION,
                 exit_decision=None) -> str:
    """依玩家輸入 + exit affordance + 當前模式，算出本 beat 的 exploration_mode。

    優先序：明確結束 > 明確再入 > 撤退/整理 > 黏著（已在鎖定模式則維持）> active。
    """
    t = text or ""
    from core.narrative.exit_resolver import END_CAMPAIGN, WITHDRAW_TO

    # 明確結束本次調查 → campaign_end_requested（唯一進 EndingGate）
    if exit_decision is not None and getattr(exit_decision, "affordance", None) == END_CAMPAIGN:
        return CAMPAIGN_END_REQUESTED

    # 明確回去研究站 / 重新進入 → 解鎖回 active（requirement 5）
    if _says_reenter(t):
        return ACTIVE_EXPLORATION

    # 整理線索（含「不新增調查 / 不碰真相 / 不回研究站」）→ review_mode
    if any(p in t for p in _REVIEW):
        return REVIEW_MODE

    # exit affordance 判定撤退 → 看是否同時整理（review）否則 temporary_retreat
    if exit_decision is not None and getattr(exit_decision, "affordance", None) == WITHDRAW_TO:
        return TEMPORARY_RETREAT
    if any(p in t for p in _RETREAT):
        return TEMPORARY_RETREAT

    # 黏著：已在鎖定模式且玩家沒明確再入/結束 → 維持鎖定（不自動恢復調查）
    if current_mode in LOCKED_MODES:
        return current_mode

    return ACTIVE_EXPLORATION


def is_review_locked(mode: str) -> bool:
    """是否處於「撤離鎖」狀態（停止自動調查推進）。"""
    return mode in LOCKED_MODES


def wants_notes(text: str) -> bool:
    """玩家是否要求『根據已知線索整理』（→ 只生 summary/notes，不新增 fact/truth）。"""
    t = text or ""
    return any(p in t for p in _REVIEW)


# ── review-mode 可用行動（available_next 的 canonical id）─────────────────────
RETURN_INSIDE = "return_inside"
REVIEW_NOTES = "review_notes"
INSPECT_INVENTORY = "inspect_inventory"
END_CAMPAIGN_ACTION = "end_campaign"
REVIEW_AFFORDANCES = [RETURN_INSIDE, REVIEW_NOTES, INSPECT_INVENTORY, END_CAMPAIGN_ACTION]

# id → 玩家面選項文字（**主題無關**；措辭須能被 resolve_mode / resolve_exit_intent 正確分類回去）
REVIEW_OPTION_LABELS = {
    RETURN_INSIDE: "返回現場，繼續調查",
    REVIEW_NOTES: "根據已知線索整理筆記",
    INSPECT_INVENTORY: "檢視我隨身帶的東西",
    END_CAMPAIGN_ACTION: "結束本次調查，接受目前結果",
}


def _review_option_label(aff: str, site_label: str) -> str:
    if aff == RETURN_INSIDE and site_label:
        return f"返回{site_label}，繼續調查"
    return REVIEW_OPTION_LABELS[aff]


def build_review_decision_point(base_dp, notes_text: str = "", site_label: str = "現場"):
    """把當前 beat 的 decision_point 換成 ReviewMode 四選一（不自動推進；永遠含再入與結束）。

    site_label：調查現場顯示名（由 WorldModel 提供；去主題化，不寫死「研究站」）。
    """
    from core.models import DecisionPoint, Option
    base_recap = (getattr(base_dp, "situation_recap", "") or "").strip()
    frame = (f"你退到安全的地方，暫時不再深入。你可以整理手上的線索，"
             f"檢視隨身的東西，返回{site_label or '現場'}繼續調查，或就此結束。")
    recap = (notes_text.strip() + "\n" + frame) if notes_text.strip() else (
        (base_recap + "\n" + frame) if base_recap else frame)
    opts = [Option(text=_review_option_label(a, site_label), tone="cautious")
            for a in REVIEW_AFFORDANCES]
    return DecisionPoint(
        situation_recap=recap, decision_type="action", suggested_options=opts,
        free_input_hint="或描述你想整理 / 檢視的事（仍在安全區，不會自動深入）",
        beat_meta=getattr(base_dp, "beat_meta", None) or {"beat_number": 0},
        is_narration_only=False)


def review_notes_text(reveal_ledger=None, world_facts: dict | None = None) -> str:
    """用**既有**線索生 review 筆記（純摘要，不新增 fact/truth；requirement 4）。"""
    lines: list[str] = []
    try:
        if reveal_ledger is not None:
            from core.narrative.revelation import public_recap
            rec = public_recap(reveal_ledger) or {}
            known = rec.get("known") or rec.get("partial") or []
            if isinstance(known, dict):
                known = known.get("found") or []
            for k in list(known)[:6]:
                lines.append(f"・{k}" if isinstance(k, str) else f"・{k}")
    except Exception:
        pass
    for k in sorted((world_facts or {}).keys())[:6]:
        lines.append(f"・（世界事實）{k}")
    if not lines:
        return "你翻看手上的線索，但目前能確定的很少——也許該回去找出更多，或就此打住。"
    return "你把目前掌握的線索理了一遍：\n" + "\n".join(lines)
