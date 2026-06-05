"""core.agents.warden — Warden Agent（U12）。

裁判玩家動作，降級順序固定（B9 設計）：
  1. 本地 deterministic hard rule（關鍵詞/正則比對 hard_triggers 與硬結局）
     → 命中 → 直接觸發，不需要 LLM。
  2. LLM semantic judgment（via SkillCaller，模糊語義、技能宣稱、軟結局 gate）
  3. LLM 全失敗（拋例外或無 caller）→ 保守「正常推進」（不誤殺玩家）。

關鍵不變式：LLM 掛掉時，本地硬規則仍必須能觸發。
"""
from __future__ import annotations

import re
from typing import Any

from core.models import WardenOutput


def _match_trigger(player_decision: str, trigger: str) -> bool:
    """判斷 player_decision 是否命中 trigger。

    先嘗試以 trigger 作為正規表示式比對（IGNORECASE）；
    若 trigger 不是合法正則，退回純子字串比對。
    """
    try:
        return bool(re.search(trigger, player_decision, re.IGNORECASE))
    except re.error:
        return trigger.lower() in player_decision.lower()


def check_hard_rule(
    player_decision: str,
    real_bible: dict,
) -> WardenOutput | None:
    """純程式碼判斷。

    依序檢查：
    1. hard_triggers → rule_violation（death_physical）
    2. ending_conditions[*] 中 type 為硬結局（death_physical / death_mental）
       且無 gate 的條目，比對 trigger → ending_triggered。

    命中任一 → 回 WardenOutput；全未命中 → 回 None。
    """
    deadly_rule: str = real_bible.get("deadly_rule", "")
    hard_triggers: list[str] = real_bible.get("hard_triggers", [])
    ending_conditions: list[dict] = real_bible.get("ending_conditions", [])

    # ── 1. 硬違規觸發器 ───────────────────────────────────────────────────
    for trigger in hard_triggers:
        if _match_trigger(player_decision, trigger):
            return WardenOutput(
                rule_violation=True,
                violated_rule=deadly_rule or trigger,
                ending_triggered="death_physical",
                ending_is_soft=False,
                directive_to_story=(
                    f"寫死亡beat：玩家觸犯致命規則「{deadly_rule or trigger}」，"
                    f"行動「{player_decision}」命中禁止條件，立即死亡。"
                ),
            )

    # ── 2. 硬結局觸發條件（死亡類、無 gate 視為硬觸發）────────────────────
    HARD_ENDING_TYPES = {"death_physical", "death_mental"}
    for cond in ending_conditions:
        cond_type: str = cond.get("type", "")
        cond_trigger: str = cond.get("trigger", "")
        gate: Any = cond.get("gate", None)

        # 有 gate 的條目留給 LLM 判斷（軟結局邏輯）
        if gate:
            continue

        if cond_type in HARD_ENDING_TYPES and cond_trigger:
            if _match_trigger(player_decision, cond_trigger):
                return WardenOutput(
                    rule_violation=cond_type in HARD_ENDING_TYPES,
                    violated_rule=deadly_rule if cond_type in HARD_ENDING_TYPES else None,
                    ending_triggered=cond_type,  # type: ignore[arg-type]
                    ending_is_soft=False,
                    directive_to_story=(
                        f"結局序列：{cond_type}。"
                        f"玩家行動「{player_decision}」觸發結局條件。"
                    ),
                )

    return None


# ── 技能宣稱封頂（UB1）──────────────────────────────────────────────────────
# 玩家即興宣稱「破格能力」→ warden 不直接否定，而是**接受但加上具體、接劇情的侷限**
# （侷限本身變成謎題/線索）。本地 deterministic 偵測，LLM 掛掉也能封頂（B9 一致）。
# 每類給一個「具體、有代價、能推劇情」的侷限模板。
_SKILL_CAP_RULES: list[tuple[tuple[str, ...], str]] = [
    # (觸發關鍵詞, 侷限模板)
    (("瞬移", "瞬間移動", "傳送", "穿牆", "穿越牆", "隱形", "飛", "漂浮", "念力", "隔空",
      "讀心", "預知", "預言", "超能力", "魔法", "法術", "復活", "倒轉時間", "時間暫停", "變身"),
     "你的能力在這棟建築裡像被某種東西壓著——只能撐幾秒，且每次使用後太陽穴劇痛、視野邊緣浮起雜訊。"
     "它有效，但你付得起幾次？而且：是什麼在這裡壓制它？"),
    (("槍", "手槍", "步槍", "開槍", "射擊", "炸彈", "手榴彈", "核彈", "火箭", "無限子彈",
      "秒殺", "一槍", "必殺", "爆頭"),
     "彈匣裡只剩兩發，而且槍聲會在空蕩的走廊裡傳得很遠——把更多東西引過來。你得想清楚，那一槍值不值得開。"),
    (("無敵", "不死", "刀槍不入", "免疫", "百毒不侵", "金剛不壞", "永生"),
     "你的身體或許扛得住刀和拳頭，但這裡真正殺人的從來不是利器——窒息、低溫、看不見的東西，對你照樣有效。"),
    (("萬能鑰匙", "開所有", "開任何", "破解一切", "駭入", "駭進", "解鎖所有", "打開所有門"),
     "它只吻合舊式機械鎖；這裡更多的是要密碼、要線索、或要某個人才肯開的門。哪一道才是你真正需要的？"),
    (("我是神", "操控一切", "控制所有", "全知", "全能", "無所不能", "支配"),
     "你能撼動的，僅止於你看得見、摸得到的這一小塊現實；而這棟建築的規則，是更早就被誰寫好的。"),
]


def check_skill_claim(player_decision: str) -> WardenOutput | None:
    """本地偵測破格技能宣稱 → 接受但封頂（加具體侷限）。未命中回 None（交 LLM/正常推進）。"""
    text = (player_decision or "").lower()
    for keywords, limitation in _SKILL_CAP_RULES:
        for kw in keywords:
            if kw.lower() in text:
                return WardenOutput(
                    rule_violation=False,
                    skill_claim=player_decision.strip()[:80],
                    skill_verdict="allow",                 # 接受宣稱，但…
                    skill_limitation=limitation,           # …加上具體、接劇情的侷限（變謎題/線索）
                    directive_to_story=(
                        "玩家宣稱了一項破格能力：將其『接受但受限』地寫入敘事——"
                        f"能力部分生效，但立即顯現這個具體侷限：「{limitation}」。"
                        "把侷限自然轉成本 beat 的新阻礙或線索，不要直接否定玩家。"
                    ),
                )
    return None


def run_warden(
    player_decision: str,
    blackboard: Any,
    caller: Any = None,
) -> WardenOutput:
    """執行 Warden 完整降級流程（B9）。

    Args:
        player_decision: 本回合玩家輸入的行動文字。
        blackboard:       Blackboard 實例（提供 real_bible 等狀態）。
        caller:           SkillCaller 實例（可為 None，表示離線模式）。

    Returns:
        WardenOutput — 裁判結果。
    """
    # 取出 real_bible
    real_bible: dict = {}
    if hasattr(blackboard, "real_bible"):
        real_bible = blackboard.real_bible or {}
    elif isinstance(blackboard, dict):
        real_bible = blackboard.get("real_bible", {})

    # ── Step 1：本地硬規則（致命，不依賴 LLM）──────────────────────────────
    hard_result = check_hard_rule(player_decision, real_bible)
    if hard_result is not None:
        # 命中硬規則 → 直接回傳，不呼叫 LLM
        return hard_result

    # ── Step 1.5：本地技能宣稱封頂（UB1；破格宣稱→接受但加侷限，LLM 掛也能封頂）──
    skill_result = check_skill_claim(player_decision)
    if skill_result is not None:
        return skill_result

    # ── Step 2：LLM 語義判斷（若有 caller）──────────────────────────────────
    if caller is not None:
        try:
            context = {
                "player_decision": player_decision,
                "deadly_rule": real_bible.get("deadly_rule", ""),
                "ending_conditions": real_bible.get("ending_conditions", []),
                "hard_triggers": real_bible.get("hard_triggers", []),
            }
            llm_result: WardenOutput = caller.call(
                "warden",
                context,
                output_model=WardenOutput,
            )
            return llm_result
        except Exception:
            # LLM 失敗（網路/解析/驗證）→ 降到 Step 3
            pass

    # ── Step 3：保守預設（不誤殺玩家）──────────────────────────────────────
    return WardenOutput(
        rule_violation=False,
        directive_to_story="正常推進",
    )
