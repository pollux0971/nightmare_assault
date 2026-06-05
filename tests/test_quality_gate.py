"""NC5 — Quality Gate 驗收測試。

驗收：元素過載/缺動機/forbidden motif/reveal jump/無意義選項各被偵測；違規可 repair 或 fallback；合格放行。
"""
from __future__ import annotations

from core.narrative.quality_gate import (
    check_beat, evaluate_opening_text, should_repair,
)
from core.narrative.models import OpeningBlueprint


# ── 合格放行 ─────────────────────────────────────────────────────────────────
def test_clean_beat_passes():
    r = check_beat("你推開門，走廊很暗，地上有一行濕腳印。",
                   ["往深處走", "檢查腳印"],
                   {"forbidden_new_elements": ["菌絲"], "truth_reveal_limit": "hinted"})
    assert r.passed and r.severity == "ok" and not should_repair(r)


# ── forbidden motif → hard_fail ──────────────────────────────────────────────
def test_forbidden_term_hard_fail():
    r = check_beat("牆上爬滿了菌絲狀的東西。", ["走"], {"forbidden_new_elements": ["菌絲"]})
    assert not r.passed and r.severity == "hard_fail"
    assert any(v.startswith("forbidden_term") for v in r.violations)
    assert should_repair(r) and r.repair_instruction


# ── 元素過載（lore 過多）→ repairable ────────────────────────────────────────
def test_too_many_lore_terms():
    text = "核心協議啟動，系統感染了收容裝置，頻率與訊號開始實驗。"
    r = check_beat(text, ["走"], {"max_new_lore_terms": 3})
    assert not r.passed and any(v.startswith("too_many_lore_terms") for v in r.violations)


# ── 無意義選項 ───────────────────────────────────────────────────────────────
def test_no_or_empty_options():
    assert "no_options" in check_beat("敘事。", [], {}).violations
    assert "empty_option" in check_beat("敘事。", ["看", "  "], {}).violations


# ── reveal jump：hinted 上限卻講穿真相 ──────────────────────────────────────
def test_reveal_jump_detected():
    r = check_beat("真相是這裡曾做過記憶實驗。", ["走"], {"truth_reveal_limit": "hinted"})
    assert "reveal_jump" in r.violations
    # 上限夠高時不算 jump
    r2 = check_beat("真相是這裡曾做過記憶實驗。", ["走"], {"truth_reveal_limit": "confirmed"})
    assert "reveal_jump" not in r2.violations


# ── 開場文字檢查（移植 reference）──────────────────────────────────────────
def test_evaluate_opening_text():
    bp = OpeningBlueprint(beat_purpose="x", motive_evidence="弟弟的學生證", max_opening_chars=900)
    ok = evaluate_opening_text("你撿起弟弟的學生證，門縫透出紅光。", bp, forbidden_terms=["菌絲"])
    assert ok.passed
    miss = evaluate_opening_text("門縫透出紅光。", bp, forbidden_terms=[])
    assert "missing_motive_evidence" in miss.violations
    long = evaluate_opening_text("弟弟的學生證。" + "字" * 1200, bp, forbidden_terms=[])
    assert "opening_too_long" in long.violations
