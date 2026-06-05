"""core.agents.orchestrator — 揭露閘門（U10）。

每個 beat 決定哪些碎片從 real_bible 搬到 revealed_bible。
多數條件用程式碼判（min_beats / location_reached / requires_touched），
只有 requires_semantic 才呼 Light LLM（透過 caller）。
"""
from __future__ import annotations

from typing import Any

from core.models import OrchestratorOutput


def run_orchestrator(
    blackboard,
    beat_number: int,
    touched_ids: list[str] | None = None,
    reached_locations: list[str] | None = None,
    caller=None,
) -> list[dict]:
    """執行揭露閘門，回傳本 beat 新揭露的碎片 dict list（含 id）。

    Args:
        blackboard:        Blackboard 實例（可讀 real_bible，可寫 revealed_bible / turn_context）。
        beat_number:       當前 beat 編號（≥1）。
        touched_ids:       本 beat 已觸及的碎片 / 事件 id 集合（可為 None 視同空集合）。
        reached_locations: 本 beat 已抵達的地點 id 集合（可為 None 視同空集合）。
        caller:            SkillCaller 實例；None 時 requires_semantic 條件視為未滿足。

    Returns:
        新揭露的碎片 dict list（每個 dict 含 id 及原 revelation 欄位）。
    """
    touched_ids = set(touched_ids or [])
    reached_locations = set(reached_locations or [])

    snap = blackboard.snapshot()

    # ── 讀取 real_bible 的 revelation_pool ──────────────────────────────────
    real_bible: dict = snap.get("real_bible") or {}
    # 支援兩種鍵名：revelation_pool（推薦）或 revelations
    revelation_pool: list[dict] = (
        real_bible.get("revelation_pool")
        or real_bible.get("revelations")
        or []
    )

    # ── 讀取已揭露碎片 id 集合 ────────────────────────────────────────────
    revealed_bible: dict = snap.get("revealed_bible") or {}
    revealed_fragments: list[dict] = list(revealed_bible.get("revealed_fragments") or [])
    already_revealed_ids: set[str] = {f["id"] for f in revealed_fragments if "id" in f}

    # ── 逐碎片判斷 ───────────────────────────────────────────────────────
    newly_revealed: list[dict] = []

    for fragment in revelation_pool:
        frag_id: str = fragment.get("id", "")
        if frag_id in already_revealed_ids:
            continue  # 已揭露，跳過

        condition: dict = fragment.get("reveal_condition") or {}

        if not _evaluate_condition(
            condition=condition,
            beat_number=beat_number,
            touched_ids=touched_ids,
            reached_locations=reached_locations,
            fragment=fragment,
            caller=caller,
        ):
            continue

        # 條件全部滿足 → 加入揭露清單
        newly_revealed.append(dict(fragment))

    if not newly_revealed:
        return []

    # ── 更新 revealed_bible ──────────────────────────────────────────────
    new_revealed_fragments = revealed_fragments + newly_revealed
    new_revealed_bible = dict(revealed_bible)
    new_revealed_bible["revealed_fragments"] = new_revealed_fragments
    blackboard.write("orchestrator", "revealed_bible", new_revealed_bible)

    # ── 寫 turn_context.newly_revealed ──────────────────────────────────
    blackboard.write("orchestrator", "turn_context.newly_revealed", newly_revealed)

    return newly_revealed


# ─────────────────────────────────────────────────────────────────────────────
# 條件評估（AND 語意：所有出現的條件欄都必須滿足）
# ─────────────────────────────────────────────────────────────────────────────

def _evaluate_condition(
    condition: dict,
    beat_number: int,
    touched_ids: set[str],
    reached_locations: set[str],
    fragment: dict,
    caller: Any,
) -> bool:
    """所有出現的條件欄都滿足才回傳 True（AND）。

    未出現的條件欄視為「無此限制」，不影響判斷。
    """
    # 1. min_beats：beat_number >= 值
    if "min_beats" in condition:
        if beat_number < condition["min_beats"]:
            return False

    # 2. location_reached：地點 id 在 reached_locations
    if "location_reached" in condition:
        if condition["location_reached"] not in reached_locations:
            return False

    # 3. requires_touched：list ⊆ touched_ids
    if "requires_touched" in condition:
        required: list[str] = condition["requires_touched"] or []
        if not all(req in touched_ids for req in required):
            return False

    # 4. requires_semantic：只有有 caller 才判；無 caller 視為未滿足
    if "requires_semantic" in condition:
        semantic_prompt: str = condition["requires_semantic"]
        if caller is None:
            return False
        # 呼 LLM 判定
        try:
            result: OrchestratorOutput = caller.call(
                "orchestrator",
                {
                    "task": "semantic_reveal_check",
                    "fragment": fragment,
                    "prompt": semantic_prompt,
                },
                output_model=OrchestratorOutput,
            )
        except Exception:
            # LLM 呼叫失敗：保守不揭露
            return False

        # 若 fragments_to_reveal 包含此碎片 id → 揭露
        frag_id = fragment.get("id", "")
        revealed_ids = {fr.id for fr in result.fragments_to_reveal}
        if frag_id not in revealed_ids:
            return False

    return True
