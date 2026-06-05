"""NR5 — Motif Cooldown 驗收測試（敘事控制 v0.2）。

驗收：同母題達上限後進 blocked（須演化）；連續 3 beat 同母題 → stagnant；
extract_motifs 偵測；換場景重置；QualityGate 對停滯母題告警。
"""
from __future__ import annotations

from core.narrative.motif_tracker import (
    MotifTracker, extract_motifs, motif_block_instruction,
)
from core.narrative.quality_gate import check_beat


# ── 超用 → blocked（reference 對齊）────────────────────────────────────────
def test_overused_motif_blocked():
    t = MotifTracker(max_uses_per_scene=2)
    t.register_beat({"stopped_clock"})
    t.register_beat({"stopped_clock"})
    assert t.is_overused("stopped_clock")
    assert "stopped_clock" in t.build_blocked_motifs()
    assert "stopped_clock" in motif_block_instruction(["stopped_clock"])


# ── 偵測母題關鍵詞 ───────────────────────────────────────────────────────────
def test_extract_motifs():
    motifs = extract_motifs("掛鐘停在 11:55，水窪裡映出你的臉。")
    assert "stopped_clock" in motifs
    assert "water_reflection" in motifs
    assert extract_motifs("一片寂靜。") == set()


# ── 連續 3 beat 同母題 → stagnant ───────────────────────────────────────────
def test_stagnant_after_three_beats():
    t = MotifTracker()
    t.register_beat({"stopped_clock", "water_reflection"})
    t.register_beat({"stopped_clock"})
    t.register_beat({"stopped_clock", "red_light"})
    assert t.stagnant_motifs(window=3) == ["stopped_clock"]
    # 只出現兩 beat 的不算停滯
    assert "water_reflection" not in t.stagnant_motifs(window=3)


# ── 換場景重置 ───────────────────────────────────────────────────────────────
def test_reset_on_scene_change():
    t = MotifTracker(max_uses_per_scene=2)
    t.register_beat({"metal_scraping"}); t.register_beat({"metal_scraping"})
    assert t.build_blocked_motifs() == ["metal_scraping"]
    t.reset_scene()
    assert t.build_blocked_motifs() == []
    assert t.stagnant_motifs() == []


# ── QualityGate 對停滯母題告警 ───────────────────────────────────────────────
def test_quality_gate_flags_stagnant():
    res = check_beat("掛鐘又停在同一刻。", ["看時鐘", "離開"],
                     {"stagnant_motifs": ["stopped_clock"]})
    assert not res.passed
    assert any("stagnant_motif:stopped_clock" in v for v in res.violations)
    # 無停滯時不因此 flag
    ok = check_beat("一個新的房間。", ["前進", "後退"], {})
    assert ok.passed
