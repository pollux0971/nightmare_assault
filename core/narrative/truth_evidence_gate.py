"""core.narrative.truth_evidence_gate — TruthEvidenceGate（reveal 准入閘）。

開放式探索鐵律：**只有** truth_investigation 或合法 structured evidence_events 才能推 reveal。
找路 / 整理 / 引用 NPC fact / 一般檢查 / review_mode / 明確「不碰真相」一律 block。
被擋下的線索仍可記為 world note / unmapped evidence（debug），但**不更新 reveal ledger**。

零 LLM、規則版。對應 15-player-sovereignty.md。
"""
from __future__ import annotations

from core.narrative.action_intent import (
    TRUTH_INVESTIGATION, OBJECT_INSPECTION, WORLD_NAVIGATION, WORLD_REVIEW,
    NPC_FACT_QUERY, CAMPAIGN_END, UNKNOWN,
)
from core.narrative.exploration_mode import is_review_locked


class TruthEvidenceGate:
    """判斷本 beat 是否准許推進 truth reveal。回 (allowed, reason)。"""

    def evaluate(self, action_class: str, *, exploration_mode: str = "active_exploration",
                 no_truth: bool = False, has_structured_evidence: bool = False,
                 truth_bearing: bool = False) -> tuple[bool, str]:
        # 1. 玩家明確「不碰真相 / 只找路 / 只整理」→ 一律 block（最高優先）
        if no_truth:
            return (False, "explicit_no_truth_intent")
        # 2. review / temporary_retreat 模式 → 整理階段，不推 reveal
        if is_review_locked(exploration_mode):
            return (False, "review_mode")
        # 3. 合法 structured evidence_events（自帶 valid truth_id）→ 永遠允許（獨立通道）
        if has_structured_evidence:
            return (True, "structured_evidence")
        # 4. 明確真相調查 → 允許
        if action_class == TRUTH_INVESTIGATION:
            return (True, "truth_investigation")
        # 5. 檢查「truth-bearing」物件 → 允許（一般物件檢查不行）
        if action_class == OBJECT_INSPECTION and truth_bearing:
            return (True, "truth_bearing_object")
        # 6. 其餘（navigation / review / npc_fact_query / 一般 inspection / end / unknown）→ block
        return (False, action_class or UNKNOWN)
