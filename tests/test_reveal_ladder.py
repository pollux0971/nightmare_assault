"""NC4 — 全域揭露上限助手測試（allowed_reveal_for / next_level_no_skip）。

驗收：依累積 evidence 漸進；不可跳級（一次最多升一階）；REVEAL_ORDER 為單一來源。
（per-truth 揭露階梯由 core.narrative.revelation 測試覆蓋；舊 RevealManager class 已移除。）
"""
from __future__ import annotations

from core.narrative.reveal_manager import allowed_reveal_for, next_level_no_skip, REVEAL_ORDER
from core.narrative.models import REVEAL_ORDER as MODELS_REVEAL_ORDER


# ── REVEAL_ORDER 單一來源 ────────────────────────────────────────────────────
def test_reveal_order_single_source():
    assert REVEAL_ORDER is MODELS_REVEAL_ORDER          # reveal_manager re-export models 的同一份
    assert REVEAL_ORDER == ["hidden", "hinted", "observed", "suspected", "confirmed", "actionable"]


# ── 全域 context 揭露上限：開場 hinted、漸進、不跳級 ─────────────────────────
def test_allowed_reveal_progression():
    assert allowed_reveal_for(0) == "hinted"            # 開場/無 evidence → 最多 hinted
    assert allowed_reveal_for(1) == "observed"
    assert allowed_reveal_for(3) == "suspected"
    assert allowed_reveal_for(5) == "confirmed"
    assert allowed_reveal_for(9) == "actionable"


def test_next_level_no_skip():
    assert next_level_no_skip("hidden", "confirmed") == "hinted"    # 一次只升一階
    assert next_level_no_skip("hinted", "actionable") == "observed"
    assert next_level_no_skip("observed", "observed") == "observed"  # 不降不停滯外升
    assert next_level_no_skip("confirmed", "hinted") == "confirmed"  # 目標較低 → 不降


def test_simulated_beats_never_jump():
    """模擬多 beat：evidence 暴增也只一階一階升（不跳級）。"""
    level = "hidden"
    seq = []
    for evidence in [0, 0, 5, 9, 9]:                    # 第 3 beat evidence 突增
        level = next_level_no_skip(level, allowed_reveal_for(evidence))
        seq.append(level)
    for a, b in zip(seq, seq[1:]):
        assert REVEAL_ORDER.index(b) - REVEAL_ORDER.index(a) <= 1
    assert seq[0] == "hinted"                           # 開場 hinted
