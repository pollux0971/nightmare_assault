"""UB6 — 序幕鉤子 + 真相種子 驗收測試。

驗收：種子涵蓋必要類型；開場 context 含 5 條義務 + 表層種子；**不洩漏 hidden_truth / real_bible**（防暴雷）；長度政策存在。
"""
from __future__ import annotations

import json

from core.agents.opening import (
    build_opening_seeds, build_opening_context, surface_seeds,
    OPENING_OBLIGATIONS, OPENING_LEN_MIN, OPENING_LEN_MAX,
)
from core.blackboard import Blackboard

SECRET = "真相是記憶置換療法導致十一名患者死亡SECRET"
DEADLY = "在低頻聲波區停留超過五回合會記憶侵蝕"
THREAT = "仍在運轉的記憶置換機器製造幻覺"


def _bb():
    bb = Blackboard()
    bb.write("setup", "real_bible", {
        "world_truth": {"what_really_happened": SECRET, "the_threat_is": THREAT, "deadly_rule": DEADLY},
        "revelation_pool": [{"id": "f1", "content": "地下室名單"}, {"id": "f2", "content": "病歷"}],
    })
    bb.write("setup", "protagonist", {"name": "周凱", "starting_situation": "弟弟林晨失蹤三週"})
    bb.write("orchestrator", "revealed_bible", {"revealed_fragments": [], "known_atmosphere": ["潮濕"]})
    return bb


# ── 種子組裝 ─────────────────────────────────────────────────────────────────
def test_seeds_cover_required_types():
    seeds = build_opening_seeds(_bb())
    types = {s["type"] for s in seeds}
    assert {"personal", "mechanical", "true", "imagery"} <= types
    assert len(seeds) >= 2
    # personal 鉤子帶到主角動機
    personal = next(s for s in seeds if s["type"] == "personal")
    assert "林晨" in personal["surface"] or "周凱" in personal["surface"]


# ── 開場 context：含義務 + 表層種子 ──────────────────────────────────────────
def test_opening_context_has_obligations_and_surface_seeds():
    ctx = build_opening_context(_bb(), {"current_scene": "lobby", "narrative_obligations": ["介紹場景"]})
    assert ctx["opening_obligations"] == OPENING_OBLIGATIONS and len(ctx["opening_obligations"]) == 5
    assert ctx["opening_length_policy"] == {"min_chars": OPENING_LEN_MIN, "max_chars": OPENING_LEN_MAX}
    assert ctx["opening_seeds"] and all("surface" in s for s in ctx["opening_seeds"])
    # 5 條義務併進 narrative_obligations，原有的也保留
    assert OPENING_OBLIGATIONS[0] in ctx["narrative_obligations"] and "介紹場景" in ctx["narrative_obligations"]


# ── ★ 防暴雷：開場 context / 表層種子 不含 hidden_truth / real_bible ──────────
def test_opening_context_does_not_leak_hidden_truth():
    ctx = build_opening_context(_bb(), {})
    blob = json.dumps(ctx, ensure_ascii=False, default=str)
    assert SECRET not in blob and DEADLY not in blob and THREAT not in blob
    assert "real_bible" not in ctx
    # surface_seeds 也不得含 hidden_truth 欄位
    for s in ctx["opening_seeds"]:
        assert "hidden_truth" not in s


def test_seeds_internally_hold_hidden_truth_but_surface_strips_it():
    """build_opening_seeds 內部保有 hidden_truth（供 real_bible 留存），surface_seeds 必須剝除。"""
    seeds = build_opening_seeds(_bb())
    assert any(s.get("hidden_truth") for s in seeds)          # 內部有
    for s in surface_seeds(seeds):                            # 表層無
        assert "hidden_truth" not in s


def test_length_policy_constants():
    assert OPENING_LEN_MIN == 600 and OPENING_LEN_MAX == 900


# ── SKILL.md 含序幕規則 ─────────────────────────────────────────────────────
def test_story_skill_has_opening_rules():
    from pathlib import Path
    skill = (Path(__file__).resolve().parent.parent / "skills" / "story" / "SKILL.md").read_text(encoding="utf-8")
    assert "序幕（開場 beat）特別規則" in skill
    assert "禁止在開場直接解釋完整真相" in skill
