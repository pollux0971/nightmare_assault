"""core.narrative.opening_director — Opening Director（NC2，移植自 patch reference_code）。

從 NarrativeContract 挑**少量**開場元素（原則 A：開頭四件事就夠），輸出 OpeningBlueprint。
refine UB6：UB6 產出候選表層種子，Director 依 opening_budget 收斂到 ≤ max_named_objects，
並把真相 seed 限制在 opening_reveal_limit（最多 hinted）。**全程只用表層字串，無 real_bible / hidden_truth。**
"""
from __future__ import annotations

from typing import Iterable

from core.narrative.models import NarrativeContract, OpeningBlueprint, TruthSeed


def select_opening_elements(contract: NarrativeContract,
                            candidate_elements: Iterable[str],
                            candidate_truth_ids: Iterable[str]) -> OpeningBlueprint:
    """依 budget 從候選元素挑 allowed（≤ max_named_objects），其餘 blocked；真相 seed ≤1、≤ reveal_limit。"""
    budget = contract.opening_budget
    forbidden = set(contract.motif_palette.forbidden_or_limited)

    allowed: list[str] = []
    blocked: list[str] = []

    # 動機證據（first_proof）永遠保留為第一個元素
    if contract.protagonist_motive.first_proof:
        allowed.append(contract.protagonist_motive.first_proof)

    for item in candidate_elements:
        if not item or item in allowed:
            continue
        if any(f and f in item for f in forbidden):
            blocked.append(item)
            continue
        if len(allowed) < budget.max_named_objects:
            allowed.append(item)
        else:
            blocked.append(item)

    truth_seeds = [
        TruthSeed(truth_id=tid, reveal_level=contract.opening_reveal_limit,
                  surface_form="用輕微異常暗示，不解釋完整真相。")
        for tid in list(candidate_truth_ids)[:1]
    ]

    return OpeningBlueprint(
        beat_purpose="establish_motive_and_first_actionable_mystery",
        motive_evidence=contract.protagonist_motive.first_proof,
        allowed_elements=allowed,
        blocked_elements=blocked,
        truth_seeds=truth_seeds,
        max_opening_chars=budget.max_opening_chars,
    )


def apply_to_context(ctx: dict, contract: NarrativeContract) -> dict:
    """把 Director 結果套進開場 context：用 budgeted 元素取代 UB6 全量種子（少元素）。

    候選元素取自 UB6 的 opening_seeds 表層；真相 id 取自 ctx['opening_truth_ids']（若有）。
    回傳新 ctx（含 opening_blueprint），**不含 hidden_truth / real_bible**。
    """
    seeds = ctx.get("opening_seeds") or []
    candidates = [s.get("surface", "") for s in seeds if isinstance(s, dict)]
    truth_ids = ctx.get("opening_truth_ids") or []
    bp = select_opening_elements(contract, candidates, truth_ids)

    out = dict(ctx)
    out["opening_blueprint"] = {
        "beat_purpose": bp.beat_purpose,
        "motive_evidence": bp.motive_evidence,
        "allowed_elements": bp.allowed_elements,
        "blocked_elements": bp.blocked_elements,
        "truth_seeds": [{"truth_id": t.truth_id, "reveal_level": t.reveal_level,
                         "surface_form": t.surface_form} for t in bp.truth_seeds],
        "max_opening_chars": bp.max_opening_chars,
    }
    # refine UB6：開場種子收斂到 budget 內的 allowed_elements（少而高價值）
    out["opening_seeds"] = [{"surface": e} for e in bp.allowed_elements]
    out["opening_element_budget"] = contract.opening_budget.max_named_objects
    out["opening_reveal_limit"] = contract.opening_reveal_limit
    return out
