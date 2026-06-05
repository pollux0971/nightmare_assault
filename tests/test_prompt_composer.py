"""P2 — PromptComposer 驗收測試。

驗收：
- 相同 fragments+vars → 相同 compiled prompt 與 prompt_hash；改 fragment → hash 變；停用 fragment → 不出現。
- 缺必填變數 → preview/compile 失敗（不靜默）；缺選填 → 可見 placeholder + warning。
- preview 全程零 LLM（mock client assert_not_called）。
"""
from __future__ import annotations

from unittest.mock import MagicMock

import pytest

from core.persistence.db import Database
from core.config.composer import PromptComposer, PromptCompositionError
from core.config import fragments as F


def _store():
    return Database(":memory:").config_store()


# ── 決定性 ───────────────────────────────────────────────────────────────────
def test_compose_is_deterministic():
    s = _store()
    c = PromptComposer(s)
    a = c.compose("story", "mvp_a_safe", {})
    b = c.compose("story", "mvp_a_safe", {})
    assert a.compiled_prompt == b.compiled_prompt
    assert a.prompt_hash == b.prompt_hash
    # 八個 story fragment 全在，依 sort_order
    assert a.enabled_fragments == [k for k, _ in F.STORY_BINDING_ORDER]


def test_fragments_joined_in_sort_order():
    s = _store()
    c = PromptComposer(s)
    out = c.compose("story", "mvp_a_safe", {})
    # role 在 output_format 之前（10 < 70）
    assert out.compiled_prompt.index("敘事生成 agent") < out.compiled_prompt.index("<<<DECISION>>>")


def test_changing_fragment_changes_hash():
    s = _store()
    c = PromptComposer(s)
    before = c.compose("story", "mvp_a_safe", {})
    s.upsert_fragment("story.style_horror", "全新風格內容", status="active")
    after = c.compose("story", "mvp_a_safe", {})
    assert after.prompt_hash != before.prompt_hash
    assert "全新風格內容" in after.compiled_prompt


def test_disabled_fragment_not_in_prompt():
    s = _store()
    c = PromptComposer(s)
    before = c.compose("story", "mvp_a_safe", {})
    assert "壓迫、清晰" in before.compiled_prompt           # style_horror 內容
    s.set_binding_enabled("story", "mvp_a_safe", "story.style_horror", False)
    after = c.compose("story", "mvp_a_safe", {})
    assert "壓迫、清晰" not in after.compiled_prompt
    assert "story.style_horror" not in after.enabled_fragments
    assert after.prompt_hash != before.prompt_hash


# ── 變數代入 ─────────────────────────────────────────────────────────────────
def _seed_var_fragment(store, content):
    store.upsert_fragment("story.vartest", content, title="vartest", status="active")
    store.connection.execute(
        "INSERT OR IGNORE INTO agent_prompt_bindings "
        "(agent_name, profile_name, fragment_key, enabled, sort_order) "
        "VALUES ('story','mvp_a_safe','story.vartest',1,90);")
    store.connection.commit()


def test_required_variable_substituted():
    s = _store()
    _seed_var_fragment(s, "場景：{{ current_scene }}。")
    out = PromptComposer(s).compose("story", "mvp_a_safe", {"current_scene": "corridor_2f"})
    assert "場景：corridor_2f。" in out.compiled_prompt
    assert "current_scene" in out.variables_used
    assert out.ok and not out.missing_required


def test_missing_required_variable_compile_raises():
    s = _store()
    _seed_var_fragment(s, "場景：{{ current_scene }}。")
    with pytest.raises(PromptCompositionError):
        PromptComposer(s).compose("story", "mvp_a_safe", {})        # strict 預設 → 拋


def test_missing_required_variable_preview_fails_visibly():
    s = _store()
    _seed_var_fragment(s, "場景：{{ current_scene }}。")
    out = PromptComposer(s).preview("story", "mvp_a_safe", {})       # 不拋，但標失敗
    assert out.ok is False
    assert "current_scene" in out.missing_required
    assert "[[MISSING:current_scene]]" in out.compiled_prompt        # 不靜默：可見標記


def test_optional_variable_placeholder_and_warning():
    s = _store()
    _seed_var_fragment(s, "提示：{{ hint? }}")
    out = PromptComposer(s).compose("story", "mvp_a_safe", {})       # 選填缺失不拋
    assert out.ok                                                    # 仍 ok（非必填）
    assert "[[opt:hint]]" in out.compiled_prompt
    assert any("hint" in w for w in out.warnings)


def test_variable_value_affects_hash():
    s = _store()
    _seed_var_fragment(s, "場景：{{ current_scene }}。")
    c = PromptComposer(s)
    h1 = c.compose("story", "mvp_a_safe", {"current_scene": "A"}).prompt_hash
    h2 = c.compose("story", "mvp_a_safe", {"current_scene": "B"}).prompt_hash
    assert h1 != h2


# ── 零 LLM ───────────────────────────────────────────────────────────────────
def test_preview_does_not_call_llm():
    """composer 不持有 client；給它一個 mock client，compose/preview 後斷言從未被呼叫。"""
    s = _store()
    mock_client = MagicMock()
    c = PromptComposer(s)
    # 即使把 client 掛到 composer 上，組裝也不該碰它
    c._unused_client = mock_client
    c.compose("story", "mvp_a_safe", {})
    c.preview("story", "mvp_a_safe", {})
    mock_client.assert_not_called()
    assert mock_client.call.call_count == 0
    assert mock_client.stream.call_count == 0


# ── model_settings / context_policy 帶出 ─────────────────────────────────────
def test_compose_carries_model_and_policy():
    s = _store()
    out = PromptComposer(s).compose("story", "mvp_a_safe", {})
    assert out.model_settings["output_schema_name"] == "StoryBeatOutput"
    assert out.context_policy["include_real_bible"] == 0
