"""P3 — Story Agent fragment 化 驗收測試。

驗收：
- forbidden_repeats 含 ask_open_door_101 → 選項不再提供開門（開門回歸過）；current_scene 反映新地點。
- narrative_obligations 反映；free_input 保留；DecisionPoint 仍可解析。
- story context/輸出不含 real_bible（C2/E2）。
"""
from __future__ import annotations

import json

import pytest

from core.persistence.db import Database
from core.config.story_prompt import (
    REQUIRED_STORY_FRAGMENTS, compose_story_prompt, validate_story_prompt,
    map_context_to_variables,
)
from core.agents.story import (
    assert_respects_forbidden, build_story_context, run_story,
)
from core.blackboard import Blackboard


def _store():
    return Database(":memory:").config_store()


# ── 組裝後 system prompt 含所有必要行為規則 ──────────────────────────────────
def test_composed_story_prompt_has_required_fragments():
    compiled = compose_story_prompt(_store(), {})
    assert validate_story_prompt(compiled) == []           # 七條必含行為 fragment 全在
    for key in REQUIRED_STORY_FRAGMENTS:
        assert key in compiled.enabled_fragments


def test_composed_prompt_encodes_kernel_obedience_and_no_repetition():
    text = compose_story_prompt(_store(), {}).compiled_prompt
    assert "不是世界狀態裁判" in text                       # kernel_obedience
    assert "禁止重複已完成事件" in text                     # no_repetition
    assert "門已經打開，不得再次詢問" in text                # 開門範例（回歸保證）
    assert "保留玩家自由選擇" in text                       # open_choice
    assert "<<<DECISION>>>" in text                        # output_format


def test_missing_required_fragment_is_detected():
    s = _store()
    s.set_binding_enabled("story", "mvp_a_safe", "story.no_repetition", False)
    missing = validate_story_prompt(compose_story_prompt(s, {}))
    assert "story.no_repetition" in missing


# ── runtime 變數面對齊 docs/03 ───────────────────────────────────────────────
def test_map_context_to_variables_surface():
    ctx = {
        "current_scene": "corridor_2f", "forbidden_repeats": ["ask_open_door_101"],
        "narrative_obligations": ["門已開，描寫走廊"], "new_clues": ["scratch"],
        "visible_npcs": [{"id": "doc"}],
    }
    v = map_context_to_variables(ctx)
    assert v["current_scene"] == "corridor_2f"
    assert v["forbidden_repeats"] == ["ask_open_door_101"]
    assert v["narrative_obligations"] == ["門已開，描寫走廊"]
    assert v["new_clues"] == ["scratch"]


# ── 反重複斷言：開門回歸 ─────────────────────────────────────────────────────
class _DP:
    def __init__(self, options, is_narration_only=False):
        self.suggested_options = options
        self.is_narration_only = is_narration_only


def test_assert_respects_forbidden_catches_reoffer():
    dp = _DP([{"text": "推開那扇門", "tone": "bold"}, {"text": "往走廊深處走", "tone": "cautious"}])
    with pytest.raises(AssertionError):
        assert_respects_forbidden(dp, ["ask_open_door_101"])


def test_assert_respects_forbidden_passes_compliant_options():
    dp = _DP([{"text": "往走廊深處走", "tone": "cautious"},
              {"text": "檢查門框上的抓痕", "tone": "cautious"},
              {"text": "回頭確認病房", "tone": "evasive"}])
    assert_respects_forbidden(dp, ["ask_open_door_101"])   # 不拋


# ── 整合：run_story（mock caller）+ forbidden_repeats → 可解析 + 保留自由輸入 + 守門過 ──
def _compliant_tokens():
    decision = json.dumps({
        "situation_recap": "門已經開了，走廊的冷光把病房切成兩半。",
        "decision_type": "action",
        "suggested_options": [
            {"text": "往走廊深處走", "tone": "cautious"},
            {"text": "檢查門框上的抓痕", "tone": "cautious"},
        ],
        "free_input_hint": "或描述你想做的事…",
        "beat_meta": {"beat_number": 5, "revelations_touched": [],
                      "npcs_present": [], "pacing": "rising", "audio_cue": "swell"},
        "is_narration_only": False,
    }, ensure_ascii=False)
    return ["門已經開了。", "<<<", "DECISION", ">>>", decision]


class FakeCaller:
    def __init__(self, tokens):
        self._tokens = tokens
        self.seen_context = None

    def stream(self, agent, context, temperature=None):
        self.seen_context = context
        for t in self._tokens:
            yield t


def test_run_story_with_forbidden_repeats_parses_and_keeps_free_input():
    bb = Blackboard()
    caller = FakeCaller(_compliant_tokens())
    ctx = {
        "current_scene": "corridor_2f", "scene_phase": "承", "beat_number": 5,
        "committed_event": "door_opened",
        "narrative_obligations": ["門已開，描寫走廊冷光"],
        "forbidden_repeats": ["ask_open_door_101"],
        "new_clues": [], "new_items": [], "visible_npcs": [], "revealed_bible": {},
    }
    narrative, dp = run_story(caller, bb, "我推開門", 5, context_override=ctx)
    assert narrative and "<<<DECISION>>>" not in narrative   # 敘事不含分隔符
    assert dp.is_narration_only is False
    assert len(dp.suggested_options) >= 1                    # 仍有選項
    assert dp.free_input_hint                                 # 自由輸入保留
    assert_respects_forbidden(dp, ctx["forbidden_repeats"])  # 不重新提供開門


def test_story_context_excludes_real_bible():
    bb = Blackboard()
    bb.write("setup", "real_bible", {"what_really_happened": "醫生早就死了"})
    bb.write("orchestrator", "revealed_bible", {"atmosphere": "潮濕"})
    ctx = build_story_context(bb, "我四處看看")
    assert "real_bible" not in ctx
    blob = json.dumps(ctx, ensure_ascii=False, default=str)
    assert "real_bible" not in blob and "醫生早就死了" not in blob
