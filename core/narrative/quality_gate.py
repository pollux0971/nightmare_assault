"""core.narrative.quality_gate — 規則版敘事品質閘門（NC5，移植自 patch reference_code）。

輸出前/後檢查：元素數量 / 動機是否清楚 / forbidden motifs / reveal jump / 選項是否有意義。
違規 → repair instruction（供重生一次）或 fallback（降級安全 beat）。零 LLM。
"""
from __future__ import annotations

from typing import Iterable

from core.narrative.models import OpeningBlueprint, QualityGateResult

# 粗略 lore 名詞偵測（真實專案可換 term extractor）。
_LORE_MARKERS = ["協議", "核心", "系統", "資料流", "感染", "收容", "儀式", "頻率",
                 "實驗", "裝置", "序號", "代號", "病毒", "訊號"]
# 過早把暗示講成確認的強確認措辭。
_CONFIRM_WORDS = ["真相是", "其實是", "原來", "真正發生的是", "答案是"]

_REPAIR = "刪除未授權元素與 forbidden 母題，保留玩家動機、一個核心異常、一個可行動線索與有意義選項；揭露不得超過上限。"


def evaluate_opening_text(text: str, blueprint: OpeningBlueprint,
                          forbidden_terms: Iterable[str],
                          max_new_lore_terms: int = 3) -> QualityGateResult:
    """檢查開場文字（移植 reference）：缺動機證據 / forbidden / lore 過多 / 過長。"""
    violations: list[str] = []
    if blueprint.motive_evidence and blueprint.motive_evidence not in text:
        violations.append("missing_motive_evidence")
    for term in forbidden_terms:
        if term and term in text:
            violations.append(f"forbidden_term:{term}")
    lore_count = sum(1 for t in _LORE_MARKERS if t in text)
    if lore_count > max_new_lore_terms:
        violations.append(f"too_many_lore_terms:{lore_count}")
    if len(text) > blueprint.max_opening_chars * 1.25:
        violations.append("opening_too_long")
    return _result(violations)


def check_beat(narrative: str, option_texts: list[str], ctx: dict) -> QualityGateResult:
    """檢查任一 beat 輸出：forbidden 母題 / 元素過載 / 選項是否有意義 / reveal jump。"""
    ctx = ctx or {}
    text = narrative or ""
    violations: list[str] = []

    for term in (ctx.get("forbidden_new_elements") or []):
        if term and term in text:
            violations.append(f"forbidden_term:{term}")

    lore_count = sum(1 for m in _LORE_MARKERS if m in text)
    max_lore = int(ctx.get("max_new_lore_terms", 3))
    if lore_count > max_lore:
        violations.append(f"too_many_lore_terms:{lore_count}")

    opts = [t for t in (option_texts or []) if isinstance(t, str)]
    if not opts:
        violations.append("no_options")
    elif any(not t.strip() for t in opts):
        violations.append("empty_option")

    # reveal jump：揭露上限為 hinted/observed 時，不得用強確認措辭講穿真相
    limit = ctx.get("truth_reveal_limit")
    if limit in ("hinted", "observed") and any(w in text for w in _CONFIRM_WORDS):
        violations.append("reveal_jump")

    # NR5：母題停滯——連續多 beat 同一母題未演化（告警級）
    for m in (ctx.get("stagnant_motifs") or []):
        violations.append(f"stagnant_motif:{m}")

    return _result(violations)


def _result(violations: list[str]) -> QualityGateResult:
    if not violations:
        return QualityGateResult(passed=True, severity="ok")
    hard = any(v.startswith("forbidden_term") for v in violations)
    return QualityGateResult(passed=False,
                             severity="hard_fail" if hard else "repairable",
                             violations=violations, repair_instruction=_REPAIR)


def should_repair(result: QualityGateResult) -> bool:
    return (not result.passed) and result.severity in ("repairable", "hard_fail")
