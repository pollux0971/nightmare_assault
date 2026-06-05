"""HD1 — QualityGate Repair Once 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：pass → 直接收（repaired=False）；fail 一次後 pass → repaired=True、story_runner 跑 2 次；
fail 兩次 → deterministic fallback（fallback=True、有安全選項）；最多 repair 一次。
"""
from __future__ import annotations

from core.narrative.quality_repair import (
    StoryRepairPipeline, StoryResult, FALLBACK_OPTIONS,
)
from core.narrative.models import QualityGateResult


def _ok():
    return QualityGateResult(passed=True, severity="ok")


def _bad():
    return QualityGateResult(passed=False, severity="repairable",
                             violations=["x"], repair_instruction="修一下")


class _Runner:
    def __init__(self, *results):
        self._results = list(results); self.calls = 0
    def __call__(self, ctx):
        r = self._results[min(self.calls, len(self._results) - 1)]
        self.calls += 1
        return StoryResult(narrative=r, options=["a", "b"])


def _fallback(ctx, q):
    return StoryResult(narrative="安全方向。", options=list(FALLBACK_OPTIONS))


# ── pass → 直接收 ────────────────────────────────────────────────────────────
def test_pass_accepts_without_repair():
    runner = _Runner("好故事")
    s = StoryRepairPipeline(runner, lambda s, c: _ok(), _fallback).run({})
    assert runner.calls == 1
    assert s.meta["quality_repaired"] is False and s.meta["quality_fallback"] is False


# ── fail 一次 → repair → pass ───────────────────────────────────────────────
def test_repair_once_then_pass():
    runner = _Runner("壞", "修好的")
    checks = iter([_bad(), _ok()])
    s = StoryRepairPipeline(runner, lambda s, c: next(checks), _fallback).run({})
    assert runner.calls == 2                          # 原始 + repair 一次
    assert s.narrative == "修好的"
    assert s.meta["quality_repaired"] is True and s.meta["quality_fallback"] is False


# ── fail 兩次 → deterministic fallback（最多 repair 一次）───────────────────
def test_repair_fails_then_fallback():
    runner = _Runner("壞1", "壞2", "不該被呼叫")
    s = StoryRepairPipeline(runner, lambda s, c: _bad(), _fallback).run({})
    assert runner.calls == 2                          # 不會無限 retry
    assert s.meta["quality_fallback"] is True
    assert s.options == FALLBACK_OPTIONS              # 有安全選項


# ── loop fallback 產生合法 DecisionPoint ───────────────────────────────────
def test_loop_deterministic_result_has_dp():
    from core.orchestrator_loop import BeatLoop
    loop = BeatLoop.__new__(BeatLoop)
    loop.beat_number = 5
    res = loop._deterministic_story_result()
    assert res.payload is not None
    assert [o.text for o in res.payload.suggested_options] == FALLBACK_OPTIONS
    assert res.narrative
