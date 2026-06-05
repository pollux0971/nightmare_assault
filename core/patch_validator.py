"""PatchValidator（SK01，移植自穩定化補丁）。

保守設計：**沒有 progress_delta 的 beat 一律拒絕**——那通常代表敘事在原地打轉。
"""
from __future__ import annotations

from copy import deepcopy
from typing import Any

from core.progress_models import EventPatch, GameState, InventoryItem, LedgerEntry, NPCPresence


class PatchValidationError(ValueError):
    pass


class PatchValidator:
    REQUIRED_DELTA = "A beat must produce at least one progress_delta."

    def validate(self, patch: EventPatch, state: GameState) -> None:
        if patch.base_version != state.version:
            raise PatchValidationError(
                f"Patch base_version={patch.base_version} does not match state.version={state.version}."
            )
        if not patch.progress_delta:
            raise PatchValidationError(self.REQUIRED_DELTA)
        if patch.event_id in state.forbidden_repeats:
            raise PatchValidationError(f"Event is forbidden from repeating: {patch.event_id}")

    def apply(self, patch: EventPatch, state: GameState) -> GameState:
        self.validate(patch, state)
        new_state = deepcopy(state)

        for op in patch.ops:
            self._apply_op(new_state, op.op, op.path, op.value, patch.event_id)

        new_state.event_status[patch.event_id] = "resolved"
        new_state.event_counts[patch.event_id] = new_state.event_counts.get(patch.event_id, 0) + 1
        new_state.recent_events.append(patch.event_id)
        new_state.recent_events = new_state.recent_events[-8:]
        new_state.forbidden_repeats.update(patch.forbidden_repeats)
        new_state.version += 1
        new_state.beat_number += 1
        return new_state

    def _apply_op(self, state: GameState, op: str, path: str, value: Any, event_id: str) -> None:
        if path == "current_scene" and op == "set":
            state.current_scene = str(value)
            return
        if path == "scene_phase" and op == "set":
            state.scene_phase = value
            return
        if path == "danger_level" and op == "inc":
            state.danger_level += int(value)
            return
        if path.startswith("event_status.") and op == "set":
            eid = path.split(".", 1)[1]
            state.event_status[eid] = value
            return
        if path.startswith("clues.") and op == "add":
            clue_id = path.split(".", 1)[1]
            if clue_id not in state.clues:
                if isinstance(value, dict):
                    state.clues[clue_id] = LedgerEntry(
                        id=clue_id,
                        title=value.get("title", clue_id),
                        content=value.get("content", ""),
                        source_event=event_id,
                        first_seen_beat=state.beat_number,
                        tags=value.get("tags", []),
                        truth_id=value.get("truth_id"),               # NR0：帶上真相綁定
                        evidence_strength=float(value.get("evidence_strength", 0.4)),
                        max_level=value.get("max_level", "actionable"),
                    )
            return
        if path.startswith("inventory.") and op == "add":
            item_id = path.split(".", 1)[1]
            if item_id not in state.inventory:
                if isinstance(value, dict):
                    state.inventory[item_id] = InventoryItem(
                        id=item_id,
                        name=value.get("name", item_id),
                        description=value.get("description", ""),
                        source_event=event_id,
                        usable=bool(value.get("usable", True)),
                        tags=value.get("tags", []),
                        held_by=value.get("held_by"),                    # 預設玩家持有
                        is_key_item=bool(value.get("is_key_item", False)),
                    )
            return
        if path.startswith("inventory.") and op == "remove":
            # path: inventory.<id>（道具被消耗/移除；item 增刪查的『刪』）
            item_id = path.split(".", 1)[1]
            state.inventory.pop(item_id, None)
            return
        if path.startswith("inventory.") and op == "set":
            # path: inventory.<id>.<attr>（如 held_by 轉移：道具易主，命運跟著走）
            rest = path[len("inventory."):]
            if "." in rest:
                item_id, attr = rest.rsplit(".", 1)
                item = state.inventory.get(item_id)
                if item is not None and hasattr(item, attr):
                    setattr(item, attr, value)
            return
        if path.startswith("npcs.") and op == "set":
            # path: npcs.<npc_id>.<attr>，npc_id 可能含「.」（如 npc.night_nurse）→ rsplit 取末段屬性
            rest = path[len("npcs."):]
            npc_id, attr = rest.rsplit(".", 1)
            npc = state.npcs.setdefault(npc_id, NPCPresence(id=npc_id, name=npc_id))
            setattr(npc, attr, value)
            if attr == "visible" and value:
                npc.last_seen_beat = state.beat_number
            return

        raise PatchValidationError(f"Unsupported patch op/path: {op} {path}")
