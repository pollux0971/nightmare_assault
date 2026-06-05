"""NR3 — EndingGate Surface Variant 驗收測試（敘事控制 v0.2）。

驗收：0 confirmed 逃脫 → ambiguous_escape 且文字含不確定收尾（≠ clean）；
clean（≥1 confirmed + quality=clean）→ clean_escape；死亡→death；
未標 escape_quality（flag OFF）→ 不套表層變體（行為不變）。
"""
from __future__ import annotations

from core.narrative.ending_gate import classify_ending_surface, ENDING_SURFACES
from core.agents.ending import build_ending_sequence, render_ending_text


class _BB:
    def __init__(self, snap):
        self._snap = snap
    def snapshot(self):
        return self._snap


def _bb(revealed_ids=()):
    pool = [{"id": "f1", "title": "碎片一", "content": "真相一"},
            {"id": "f2", "title": "碎片二", "content": "真相二"}]
    revealed = [f for f in pool if f["id"] in revealed_ids]
    return _BB({
        "real_bible": {"world_truth": {"what_really_happened": "這裡曾經…",
                                       "the_threat_is": "它"}, "revelation_pool": pool},
        "revealed_bible": {"revealed_fragments": revealed},
    })


# ── 表層分類純函式 ───────────────────────────────────────────────────────────
def test_classify_zero_truth_escape_is_ambiguous():
    assert classify_ending_surface({"type": "escape", "escape_quality": "clean"}, 0) == "ambiguous_escape"
    assert classify_ending_surface({"type": "escape", "escape_quality": "ambiguous"}, 3) == "ambiguous_escape"


def test_classify_clean_escape_needs_confirmed():
    assert classify_ending_surface({"type": "escape", "escape_quality": "clean"}, 2) == "clean_escape"


def test_classify_death_and_truth_locked():
    assert classify_ending_surface({"type": "death_physical"}, 0) == "death"
    assert classify_ending_surface({"type": "truth_revealed"}, 0) == "truth_locked"
    assert classify_ending_surface({"type": "truth_revealed"}, 1) == "clean_escape"
    assert set(ENDING_SURFACES) >= {"clean_escape", "ambiguous_escape", "truth_locked"}


# ── 0/X 逃脫渲染為 ambiguous，文字 ≠ clean ──────────────────────────────────
def test_ambiguous_escape_renders_differently():
    amb = build_ending_sequence(_bb(revealed_ids=()), {"type": "escape", "escape_quality": "clean"})
    assert amb["ending_surface"] == "ambiguous_escape"
    amb_text = render_ending_text(amb)
    assert "以為自己走出去了" in amb_text          # ambiguous 專屬收尾

    clean = build_ending_sequence(_bb(revealed_ids=("f1", "f2")),
                                  {"type": "escape", "escape_quality": "clean"})
    assert clean["ending_surface"] == "clean_escape"
    clean_text = render_ending_text(clean)
    assert "帶走了答案" in clean_text
    assert amb_text != clean_text                  # 兩種結局文字明顯不同


# ── flag OFF（無 escape_quality）→ 不套表層變體，行為不變 ────────────────────
def test_no_surface_when_flag_off():
    e = build_ending_sequence(_bb(revealed_ids=()), {"type": "escape"})
    assert "ending_surface" not in e or e.get("ending_surface") is None
    # 仍是原本的 escape 收尾
    assert "走出來了" in e["closing"]
