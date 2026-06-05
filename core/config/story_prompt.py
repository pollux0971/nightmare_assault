"""core.config.story_prompt — Story Agent fragment 化的組裝 + 驗證（P3）。

把 Story Agent 的 system prompt 從**組裝後的 fragment**取得（wrap，不 rewrite）：
規則（role/objective/kernel_obedience/no_repetition/open_choice/context_policy/output_format）
進 system prompt；per-beat 的 runtime 變數（current_scene/forbidden_repeats/...）由 ContextBuilder
（SK03）放進 context（user 訊息）。兩半合起來保證「服從 kernel + 不重複已解決事件 + 保留自由選擇」。

static prompt（skills/story/SKILL.md）仍保留，來源切換在 P4。契約：dev/CONTRACTS.md §十一。
canonical：patch docs/03、docs/05。
"""
from __future__ import annotations

from typing import Any, Optional

from core.config.composer import PromptComposer, CompiledPrompt

# Story system prompt 必含的行為 fragment（缺任一 → 行為不完整）。
REQUIRED_STORY_FRAGMENTS = (
    "story.role", "story.objective", "story.kernel_obedience",
    "story.no_repetition", "story.open_choice", "story.context_policy",
    "story.output_format",
)

# docs/03「Required runtime variables」：組裝後 story 應收到的 per-beat 變數面。
# 這些主要由 ContextBuilder 放進 context；列在此供 preview/驗證/文件對齊。
STORY_RUNTIME_VARIABLES = (
    "current_scene", "current_location", "scene_phase", "narrative_obligations",
    "forbidden_repeats", "new_clues", "new_items", "visible_npcs",
    "recent_events", "future_direction_hint", "output_schema",
)


def map_context_to_variables(context: dict) -> dict:
    """把 ContextBuilder 的 context 鍵映射成 docs/03 的 runtime 變數面（供 composer 變數代入/快照）。

    僅做名稱對齊與淺取；不洩漏 real_bible（context 本就不含）。
    """
    ctx = context or {}
    return {
        "current_scene": ctx.get("current_scene"),
        "current_location": ctx.get("current_location") or ctx.get("current_scene"),
        "scene_phase": ctx.get("scene_phase"),
        "narrative_obligations": ctx.get("narrative_obligations", []),
        "forbidden_repeats": ctx.get("forbidden_repeats", []),
        "new_clues": ctx.get("new_clues", []),
        "new_items": ctx.get("new_items", []),
        "visible_npcs": ctx.get("visible_npcs", []),
        "recent_events": ctx.get("recent_events", []),
        "future_direction_hint": ctx.get("soft_lookahead", []),
        "output_schema": ctx.get("output_schema", "StoryBeatOutput"),
    }


def compose_story_prompt(store: Any, runtime_variables: Optional[dict] = None,
                         profile: Optional[str] = None, *, strict: bool = False) -> CompiledPrompt:
    """組裝 Story Agent 的 system prompt（由 fragment 來）。

    strict=False（preview/runtime 預設）：缺必填變數不拋，以 .missing_required 呈現，由呼叫端決定退 fallback。
    """
    return PromptComposer(store).compose("story", profile, runtime_variables or {}, strict=strict)


def validate_story_prompt(compiled: CompiledPrompt) -> list[str]:
    """回傳缺少的必含行為 fragment（空 list = 完整）。供 P4 runtime 健檢 / P6 回歸。"""
    present = set(compiled.enabled_fragments)
    return [k for k in REQUIRED_STORY_FRAGMENTS if k not in present]
