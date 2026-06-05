"""core.narrative.quality_repair — QualityGate 從 monitor 變 gate（HD1，Runtime Hard-Gate v0.3.1）。

`check → pass 收 / fail → repair_once → 仍 fail → deterministic_fallback`。**最多 repair 一次**
（避免 LLM 成本/延遲失控）。fallback 不求文采，只要系統正確（給可驗證方向 + 安全選項）。

對應 dev/CONTRACTS.md §十五（StoryRepairPipeline）。
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Callable


@dataclass
class StoryResult:
    """story 一次產出的標準封裝。payload 帶 loop 需要的原物件（如 DecisionPoint）。"""
    narrative: str
    options: list[str] = field(default_factory=list)
    meta: dict = field(default_factory=dict)
    payload: Any = None


class StoryRepairPipeline:
    def __init__(self, story_runner: Callable, quality_check: Callable,
                 deterministic_fallback: Callable):
        # story_runner(ctx)->StoryResult；quality_check(StoryResult,ctx)->QualityGateResult；
        # deterministic_fallback(ctx,quality_result)->StoryResult
        self.story_runner = story_runner
        self.quality_check = quality_check
        self.deterministic_fallback = deterministic_fallback

    def run(self, ctx: dict) -> StoryResult:
        s = self.story_runner(ctx)
        q = self.quality_check(s, ctx)
        if q.passed:
            s.meta["quality_repaired"] = False
            s.meta["quality_fallback"] = False
            return s
        # repair once（帶 repair_instruction 重生一次）
        repair_ctx = dict(ctx)
        repair_ctx["repair_instruction"] = q.repair_instruction or ""
        s2 = self.story_runner(repair_ctx)
        q2 = self.quality_check(s2, ctx)
        if q2.passed:
            s2.meta["quality_repaired"] = True
            s2.meta["quality_fallback"] = False
            return s2
        # 仍 fail → deterministic fallback
        f = self.deterministic_fallback(ctx, q2)
        f.meta["quality_repaired"] = True
        f.meta["quality_fallback"] = True
        return f


# ── deterministic fallback 文案（docs/06）──────────────────────────────────
FALLBACK_NARRATIVE = ("你沒有得到完整答案，但取得了一個可驗證的方向："
                      "控制室的儀器記錄，可能保存著剛才那個數字的來源。")
FALLBACK_OPTIONS = ["前往控制室查記錄", "繼續追問在場的人", "暫時撤離到安全區"]
