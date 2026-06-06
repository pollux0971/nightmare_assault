"""Beat Rendering Debug（patch v0.7 P3）—— 量測 beat 類型/字數/連續過短；不修復（P4 未開）。"""
from __future__ import annotations

import core.constants as C
from core.narrative.beat_rendering import (
    classify_beat_type, evaluate_beat_rendering, count_cjk_chars,
    NORMAL_EXPLORATION, NEW_AREA_ENTRY, OBJECT_INSPECTION, NPC_FIRST_INTRO,
    REVIEW_MODE, DANGER_BEAT, ENDING, OPENING, BUDGETS,
)


# ── classify_beat_type ───────────────────────────────────────────────────────
def test_classify_priority():
    assert classify_beat_type(ended=True) == ENDING
    assert classify_beat_type(is_opening=True) == OPENING
    assert classify_beat_type(review_locked=True) == REVIEW_MODE
    assert classify_beat_type(danger_delta=1) == DANGER_BEAT
    assert classify_beat_type(npc_first_intro=True) == NPC_FIRST_INTRO
    assert classify_beat_type(scene_changed=True) == NEW_AREA_ENTRY
    assert classify_beat_type(action_class="object_inspection") == OBJECT_INSPECTION
    assert classify_beat_type() == NORMAL_EXPLORATION
    # review 優先於 danger（撤退整理時不算危險 beat）
    assert classify_beat_type(review_locked=True, danger_delta=2) == REVIEW_MODE


def test_count_cjk_chars():
    assert count_cjk_chars("你 好\n世界") == 4
    assert count_cjk_chars("") == 0


# ── evaluate：too_short + streak ─────────────────────────────────────────────
def test_short_beat_increments_streak():
    short = "走廊很暗。"                                 # 遠低於 normal 下限 220
    d1 = evaluate_beat_rendering(short, NORMAL_EXPLORATION, prev_short_streak=0)
    assert d1["too_short"] is True and d1["short_streak"] == 1
    assert d1["repair_attempted"] is False               # P4 未開
    assert d1["target_min_chars"] == BUDGETS[NORMAL_EXPLORATION][0]
    d2 = evaluate_beat_rendering(short, NORMAL_EXPLORATION, prev_short_streak=1)
    assert d2["short_streak"] == 2                        # 連續第二次 → 達 P4 門檻（但不修復）
    assert d2["repair_attempted"] is False


def test_long_beat_resets_streak():
    long = "字" * 300
    d = evaluate_beat_rendering(long, NORMAL_EXPLORATION, prev_short_streak=2)
    assert d["too_short"] is False and d["short_streak"] == 0


def test_review_mode_excluded_from_streak():
    short = "你理了一遍線索。"
    d = evaluate_beat_rendering(short, REVIEW_MODE, prev_short_streak=2)
    assert d["too_short"] is False                        # review 不算過短
    assert d["short_streak"] == 0                         # 不累計、且重置（不會觸發一般 repair）


# ── loop 整合 ─────────────────────────────────────────────────────────────────
def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def test_observation_has_beat_rendering(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("我往前走查看四周")
    br = out["beat_rendering"]
    for k in ("beat_type", "target_min_chars", "target_max_chars", "actual_chars",
              "too_short", "short_streak", "repair_attempted"):
        assert k in br
    assert br["repair_attempted"] is False                # P4 未啟用


def test_review_beat_not_in_general_streak(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop.step("我往前走查看四周")                          # 一般 beat（FakeCaller 敘事很短 → too_short）
    s_after_normal = loop._short_beat_streak
    out = loop.step("先退到外面整理線索，不結束本次調查")   # review beat
    assert out["beat_rendering"]["beat_type"] == REVIEW_MODE
    assert out["beat_rendering"]["too_short"] is False     # review 不算過短
    assert loop._short_beat_streak == 0                    # review 不累計一般 streak
    assert s_after_normal >= 1                             # 但先前一般 beat 有計到
