"""NR4 — Escape Commit Gate 驗收測試（敘事控制 v0.2）。

驗收：首次「我試圖離開」→ await_commit（產出口發現，不結局）；
已發現出口 + 明確提交 → commit（可結算）；非逃脫 → none；
loop tick 注入出口發現義務並標記 _exit_candidate_found。
"""
from __future__ import annotations

from core.narrative.escape_commit import (
    EscapeCommitGate, is_escape_intent, is_explicit_commit, EXIT_CANDIDATE_OBLIGATION,
)


# ── 兩段式：首次逃脫意圖 → await_commit；不即結局 ───────────────────────────
def test_first_attempt_awaits_commit():
    g = EscapeCommitGate()
    assert g.decide("我試圖離開這裡", exit_candidate_found=False) == "await_commit"
    # 已找到出口後再表達離開 → commit
    assert g.decide("我頭也不回地走出去", exit_candidate_found=True) == "commit"


# ── 已找到出口 + 明確提交 → commit ──────────────────────────────────────────
def test_commit_after_exit_found():
    g = EscapeCommitGate()
    assert g.decide("確定離開", exit_candidate_found=True) == "commit"
    # 已找到出口但沒有任何離開意圖 → none（不誤觸）
    assert g.decide("我檢查桌子", exit_candidate_found=True) == "none"


# ── 非逃脫行動 → none ────────────────────────────────────────────────────────
def test_non_escape_is_none():
    g = EscapeCommitGate()
    assert g.decide("我往地下室走", exit_candidate_found=False) == "none"


def test_intent_and_commit_detectors():
    assert is_escape_intent("我要從出口逃出去")
    assert not is_escape_intent("我坐下休息")
    assert is_explicit_commit("我下定決心離開")
    assert not is_explicit_commit("這裡有出口嗎？")


# ── loop tick：await_commit 注入義務 + 標記出口已找到 ───────────────────────
def test_loop_tick_injects_obligation():
    from core.orchestrator_loop import BeatLoop
    loop = BeatLoop.__new__(BeatLoop)              # 不跑 __init__，直接測 tick
    loop._exit_candidate_found = False
    ctx = {}
    step = loop._escape_commit_tick("我試圖離開這裡", ctx)
    assert step == "await_commit"
    assert loop._exit_candidate_found is True
    assert EXIT_CANDIDATE_OBLIGATION in ctx["narrative_obligations"]
    # 第二次（已找到出口）+ 提交 → commit
    assert loop._escape_commit_tick("確定離開", {}) == "commit"
