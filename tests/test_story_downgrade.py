"""NC3 — Story Agent Downgrade 驗收測試。

驗收：story context 帶 allowed/forbidden_new_elements + beat_purpose + truth_reveal_limit + player_motive；
每 beat element_limit=1；SKILL.md 含 Delta 段；不含 hidden_truth。
"""
from __future__ import annotations

import json
from pathlib import Path

from core.narrative.models import NarrativeContract, ProtagonistMotive, MotifPalette
from core.narrative.story_control import apply_story_downgrade


def _contract():
    return NarrativeContract(
        core_premise="SECRET_PREMISE",
        protagonist_motive=ProtagonistMotive("弟弟失蹤", "找到他並逃出", "愧疚", "學生證"),
        central_question="他還記得你嗎？",
        motif_palette=MotifPalette(primary=["潮濕", "低頻嗡鳴"], forbidden_or_limited=["菌絲", "血字"]),
    )


def test_downgrade_adds_control_fields():
    ctx = apply_story_downgrade({"current_scene": "ward"}, _contract())
    assert ctx["player_motive"] == "找到他並逃出"
    assert ctx["allowed_new_elements"] == ["潮濕", "低頻嗡鳴"]
    assert ctx["forbidden_new_elements"] == ["菌絲", "血字"]
    assert ctx["beat_purpose"]
    assert ctx["truth_reveal_limit"]                       # 有揭露上限
    assert ctx["element_limit"] == 1                       # 每 beat 一個主要資訊
    assert "不自行新增核心設定" in ctx["instruction"]


def test_downgrade_respects_existing_reveal_limit():
    ctx = apply_story_downgrade({"opening_reveal_limit": "hinted"}, _contract())
    assert ctx["truth_reveal_limit"] == "hinted"


def test_downgrade_no_hidden_leak():
    ctx = apply_story_downgrade({}, _contract())
    assert "SECRET_PREMISE" not in json.dumps(ctx, ensure_ascii=False)
    assert "core_premise" not in ctx


def test_story_skill_has_delta_section():
    skill = (Path(__file__).resolve().parent.parent / "skills" / "story" / "SKILL.md").read_text(encoding="utf-8")
    assert "Story Agent Delta" in skill
    assert "你不是世界觀發明者" in skill
    assert "每 beat 只新增一個主要敘事資訊" in skill
