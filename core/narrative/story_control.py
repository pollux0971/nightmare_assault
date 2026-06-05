"""core.narrative.story_control — Story Agent 降權（NC3）。

Story Agent 只執行 blueprint / obligations，不發明世界觀。本模組把降權欄位加進 story context：
allowed_new_elements / forbidden_new_elements / beat_purpose / truth_reveal_limit / player_motive，
並要求每 beat 只新增一個主要敘事資訊。**不含 hidden_truth / real_bible。**
"""
from __future__ import annotations

from typing import Any


def apply_story_downgrade(ctx: dict, contract: Any,
                          beat_purpose: str = "advance_one_main_thread") -> dict:
    """把降權欄位加進 story context（回傳新 ctx）。"""
    palette = getattr(contract, "motif_palette", None)
    motive = getattr(contract, "protagonist_motive", None)
    out = dict(ctx)
    out["player_motive"] = getattr(motive, "immediate_goal", "") if motive else ""
    out["allowed_new_elements"] = list(getattr(palette, "primary", []) or []) if palette else []
    out["forbidden_new_elements"] = list(getattr(palette, "forbidden_or_limited", []) or []) if palette else []
    out["beat_purpose"] = ctx.get("beat_purpose") or beat_purpose
    # 揭露上限：開場用 opening_reveal_limit；其餘 beat 由 loop 的全域 ceiling（reveal_manager
    # allowed_reveal_for/next_level_no_skip, NC4）填進 ctx['truth_reveal_limit']，否則保守 observed
    out["truth_reveal_limit"] = (
        ctx.get("truth_reveal_limit") or ctx.get("opening_reveal_limit")
        or getattr(contract, "opening_reveal_limit", "observed")
    )
    out["element_limit"] = 1                       # 每 beat 只新增一個主要敘事資訊
    out["instruction"] = (
        (ctx.get("instruction", "") + " ") +
        "只執行 blueprint / obligations，不自行新增核心設定（你不是世界觀發明者）；"
        "每 beat 只新增一個主要敘事資訊；選項必須關聯 player_motive / 線索 / 危險。"
    ).strip()
    return out
