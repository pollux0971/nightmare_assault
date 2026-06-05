"""core.agents.compactor — Compactor Agent（U14，無限模式的承重牆）。

職責：把滑出視窗的舊 beat + 既有滾動摘要，壓縮成精煉記憶，**絕不丟失保護清單**
（伏筆 / anchor / 未揭露線索）。讓故事能無限走下去而 context 不爆。

兩個關鍵不變式（30 beat 驗收，E3 / C11）：
1. 滾動摘要 token 有界——compactor 產出取代式新摘要，不隨 beat 數線性成長。
2. 保護清單永不遺失——protected_foreshadowings 每輪原樣帶回並併入輸出。

非同步運作（A5）：compact() 只產 CompactorOutput；apply_to_blackboard() 用
submit_patch 提交 patch，安全點 merge_and_bump 才實際寫入 Blackboard。
"""
from __future__ import annotations

from typing import Any

from core.constants import (
    CONTEXT_THRESHOLD_L1,
    CONTEXT_THRESHOLD_L2,
    CONTEXT_THRESHOLD_L3,
)
from core.models import CompactorOutput, LedgerFact


# ─────────────────────────────────────────────────────────────────────────────
# 壓縮等級
# ─────────────────────────────────────────────────────────────────────────────

def compression_level(usage_pct: float) -> str:
    """依 context 使用率決定壓縮等級。

    - usage < L1            → "none"（尚未到門檻，仍可做輕量折疊）
    - L1 <= usage < L2      → "L1"（一般壓縮，背景跑）
    - L2 <= usage < L3      → "L2"（較激進）
    - usage >= L3           → "L3"（緊急，最激進，可阻塞下一 beat）
    """
    if usage_pct >= CONTEXT_THRESHOLD_L3:
        return "L3"
    if usage_pct >= CONTEXT_THRESHOLD_L2:
        return "L2"
    if usage_pct >= CONTEXT_THRESHOLD_L1:
        return "L1"
    return "none"


class Compactor:
    """記憶壓縮 agent。

    Args:
        caller: SkillCaller（提供 .call("compactor", context, output_model=...)）。
                可為 None → 走離線降級（不呼叫 LLM，僅保骨幹 + 保護清單）。
        protected_foreshadowings: 啟動時即知的保護清單（伏筆 / anchor / 線索）。
                每次 compact 都會把這份清單原樣帶入 context 與輸出，確保不遺失。
    """

    def __init__(
        self,
        caller: Any = None,
        protected_foreshadowings: list[str] | None = None,
    ) -> None:
        self.caller = caller
        # 內部維護一份累積的保護清單（去重、保序）
        self._protected: list[str] = list(protected_foreshadowings or [])

    # ── 保護清單管理 ──────────────────────────────────────────────────────

    @property
    def protected_foreshadowings(self) -> list[str]:
        return list(self._protected)

    def _absorb_protected(self, items: list[str] | None) -> None:
        """把新出現的伏筆併入保護清單（去重、保序）。"""
        for it in items or []:
            if it and it not in self._protected:
                self._protected.append(it)

    # ── 壓縮（非同步，只產 patch 用的 output）────────────────────────────

    def compact(
        self,
        evicted_beats: list,
        current_summary: str,
        ledger: list[LedgerFact],
        usage_pct: float,
    ) -> CompactorOutput:
        """把滑出的 beat + 現有摘要壓縮成新摘要 + ledger 更新 + 保護清單。

        依 usage_pct 決定壓縮等級，並餵給 caller。caller 回的
        preserved_foreshadowings 會與 compactor 自持的保護清單聯集，
        確保「已埋未揭露的伏筆」不論多舊都留存（C11）。

        caller 為 None 或呼叫失敗時降級：保留現有摘要骨幹 + 保護清單，
        不發明新事實（SKILL.md 邊界），不阻塞主線。
        """
        level = compression_level(usage_pct)

        context = {
            "compression_level": level,
            "usage_pct": usage_pct,
            "evicted_beats": list(evicted_beats),
            "current_summary": current_summary,
            "ledger": [self._fact_pair(f) for f in ledger],
            "protected_foreshadowings": list(self._protected),
        }

        output: CompactorOutput | None = None
        if self.caller is not None:
            try:
                output = self.caller.call(
                    "compactor",
                    context,
                    output_model=CompactorOutput,
                )
            except Exception:
                output = None

        if output is None:
            output = self._offline_fallback(
                evicted_beats, current_summary, ledger, usage_pct, level
            )

        # ── 保護清單聯集：caller 帶回的 + compactor 自持的，一個都不能少 ──
        self._absorb_protected(output.preserved_foreshadowings)
        output.preserved_foreshadowings = list(self._protected)

        return output

    # ── 套用到 Blackboard（非同步隔離，A5）──────────────────────────────

    def apply_to_blackboard(self, blackboard: Any, output: CompactorOutput) -> None:
        """以 submit_patch 提交 rolling_summary / ledger 的 patch。

        不直接寫 Blackboard——非同步 agent 只產 patch，安全點 merge_and_bump
        才生效。base_version 取當前 blackboard.version，過期則於 merge 時被丟棄。
        """
        base_version = getattr(blackboard, "version", 0)

        # rolling_summary：取代為壓縮後摘要
        blackboard.submit_patch({
            "base_version": base_version,
            "writer": "compactor",
            "target": "rolling_summary",
            "value": output.compressed_summary,
        })

        # ledger：把更新後的事實清單（dict 形式）整批覆寫
        ledger_value = [self._fact_dict(f) for f in output.ledger_updates]
        blackboard.submit_patch({
            "base_version": base_version,
            "writer": "compactor",
            "target": "ledger",
            "value": ledger_value,
        })

    # ── 內部輔助 ──────────────────────────────────────────────────────────

    @staticmethod
    def _fact_pair(fact: Any) -> list:
        """LedgerFact → [type, content]（餵 LLM 用二元組形式）。"""
        if isinstance(fact, LedgerFact):
            return [fact.type, fact.content]
        if isinstance(fact, dict):
            return [fact.get("type", ""), fact.get("content", "")]
        if isinstance(fact, (list, tuple)) and len(fact) >= 2:
            return [fact[0], fact[1]]
        return [str(fact), ""]

    @staticmethod
    def _fact_dict(fact: Any) -> dict:
        """LedgerFact → {type, content}（寫回 blackboard 用）。"""
        if isinstance(fact, LedgerFact):
            return {"type": fact.type, "content": fact.content}
        if isinstance(fact, dict):
            return {"type": fact.get("type", ""), "content": fact.get("content", "")}
        if isinstance(fact, (list, tuple)) and len(fact) >= 2:
            return {"type": fact[0], "content": fact[1]}
        return {"type": str(fact), "content": ""}

    def _offline_fallback(
        self,
        evicted_beats: list,
        current_summary: str,
        ledger: list[LedgerFact],
        usage_pct: float,
        level: str,
    ) -> CompactorOutput:
        """離線降級：不呼叫 LLM 時的保守壓縮。

        - 保留現有摘要骨幹（不串接 evicted 原文，避免摘要無限成長）。
        - ledger 原樣透傳（不發明、不刪除）。
        - 保護清單原樣帶回。
        """
        return CompactorOutput(
            compressed_summary=current_summary or "",
            ledger_updates=list(ledger),
            archived_beats=[self._beat_id(b) for b in evicted_beats],
            preserved_foreshadowings=list(self._protected),
            final_usage_estimate=usage_pct,
        )

    @staticmethod
    def _beat_id(beat: Any) -> str:
        """從 beat 取一個可封存的 id（dict 取 id/beat_number，否則 str()）。"""
        if isinstance(beat, dict):
            for k in ("id", "beat_id", "beat_number"):
                if k in beat:
                    return str(beat[k])
        return str(beat)
