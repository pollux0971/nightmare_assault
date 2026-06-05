"""P6 — Prompt 回歸測試集（決定性、零 LLM）。

五項（task_card P6）：開門 anti-repetition、story 無 real_bible、prompt hash 決定性、
new clue 出現在 story context、free input 保留。每項可回答「哪個 prompt 版本造成這個行為」。
"""
from __future__ import annotations

import json

import pytest

from core.persistence.db import Database
from core.config.composer import PromptComposer
from core.config.story_prompt import compose_story_prompt
from core.agents.story import assert_respects_forbidden, build_story_context, run_story
from core.blackboard import Blackboard
from core.progress_context import ContextBuilder
from core.progress_models import GameState, EventPatch, ProgressResult


def _store():
    return Database(":memory:").config_store()


# ── 1. 開門 anti-repetition ──────────────────────────────────────────────────
def test_regression_open_door_not_reoffered():
    # 規則在組裝 prompt 內
    assert "門已經打開，不得再次詢問" in compose_story_prompt(_store(), {}).compiled_prompt
    # 守門斷言抓到重新提供開門、放過合規選項
    bad = type("DP", (), {"suggested_options": [{"text": "再次推開那扇門"}], "is_narration_only": False})()
    good = type("DP", (), {"suggested_options": [{"text": "往走廊深處走"}], "is_narration_only": False})()
    with pytest.raises(AssertionError):
        assert_respects_forbidden(bad, ["ask_open_door_101"])
    assert_respects_forbidden(good, ["ask_open_door_101"])


# ── 2. story context 無 real_bible ───────────────────────────────────────────
def test_regression_no_real_bible_in_story_context():
    bb = Blackboard()
    bb.write("setup", "real_bible", {"what_really_happened": "兇手是醫生SECRET"})
    bb.write("orchestrator", "revealed_bible", {"atmosphere": "潮濕"})
    ctx = build_story_context(bb, "我環顧四周")
    blob = json.dumps(ctx, ensure_ascii=False, default=str)
    assert "real_bible" not in ctx and "real_bible" not in blob
    assert "兇手是醫生SECRET" not in blob
    # context policy 也硬性禁止
    assert _store().get_context_policy("story", "mvp_a_safe")["include_real_bible"] == 0


# ── 3. prompt hash 決定性 ────────────────────────────────────────────────────
def test_regression_prompt_hash_deterministic():
    s = _store()
    c = PromptComposer(s)
    h1 = c.preview("story", "mvp_a_safe", {}).prompt_hash
    h2 = c.preview("story", "mvp_a_safe", {}).prompt_hash
    assert h1 == h2
    s.upsert_fragment("story.style_horror", "改寫風格", status="active")
    assert c.preview("story", "mvp_a_safe", {}).prompt_hash != h1   # 改 fragment → hash 變


# ── 4. new clue 出現在 story context ─────────────────────────────────────────
def test_regression_new_clue_surfaces_in_context():
    state = GameState(version=1, beat_number=5, current_scene="corridor_2f", scene_phase="承")
    patch = EventPatch(base_version=1, event_id="e1", ops=[], progress_delta=[],
                       narrative_obligations=["描寫新線索"], new_clues=["scratch_marks"])
    progress = ProgressResult(accepted=True, patch=patch, committed_event="e1")
    ctx = ContextBuilder().build_story_context(state, progress, revealed_bible={})
    assert "scratch_marks" in ctx["new_clues"]
    assert "描寫新線索" in ctx["narrative_obligations"]


# ── 5. free input 保留 ───────────────────────────────────────────────────────
def _compliant_tokens():
    dj = json.dumps({
        "situation_recap": "走廊冷光。", "decision_type": "action",
        "suggested_options": [{"text": "往走廊深處走", "tone": "cautious"}],
        "free_input_hint": "或描述你想做的事…",
        "beat_meta": {"beat_number": 5, "pacing": "rising", "audio_cue": "swell"},
        "is_narration_only": False,
    }, ensure_ascii=False)
    return ["門已開。", "<<<", "DECISION", ">>>", dj]


class _Caller:
    def __init__(self, toks):
        self._t = toks

    def stream(self, agent, context, temperature=None, system_override=None):
        for t in self._t:
            yield t


def test_regression_free_input_preserved():
    bb = Blackboard()
    ctx = {"current_scene": "corridor_2f", "beat_number": 5, "committed_event": "door_opened",
           "narrative_obligations": [], "forbidden_repeats": ["ask_open_door_101"],
           "new_clues": [], "new_items": [], "visible_npcs": [], "revealed_bible": {}}
    _, dp = run_story(_Caller(_compliant_tokens()), bb, "我推門", 5, context_override=ctx)
    assert dp.free_input_hint                       # 自由輸入提示仍在
    assert dp.is_narration_only is False
    assert_respects_forbidden(dp, ctx["forbidden_repeats"])
