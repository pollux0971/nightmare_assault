"""ContextBuilder（SK03，移植自穩定化補丁）。

把 story-agent context 維持最小且聚焦。**不放完整 chat log 或 real_bible**（結構性防暴雷）。
"""
from __future__ import annotations

from core.progress_models import GameState, ProgressResult


class ContextBuilder:
    def build_story_context(self, state: GameState, progress: ProgressResult,
                            revealed_bible: dict | None = None) -> dict:
        patch = progress.patch
        visible_npcs = [
            {"id": n.id, "name": n.name, "scene": n.current_scene, "entry_mode": n.entry_mode}
            for n in state.npcs.values()
            if n.visible or n.id in patch.spawned_npcs
        ]
        relevant_clues = [
            {"id": c.id, "title": c.title, "content": c.content}
            for c in state.clues.values()
            if c.id in patch.new_clues or c.source_event in state.recent_events[-2:]
        ]
        relevant_items = [
            {"id": i.id, "name": i.name, "description": i.description, "usable": i.usable}
            for i in state.inventory.values()
        ]
        return {
            "current_scene": state.current_scene,
            "scene_phase": state.scene_phase,
            "beat_number": state.beat_number,
            "recent_events": state.recent_events[-2:],
            "committed_event": progress.committed_event,
            "narrative_obligations": patch.narrative_obligations,
            "forbidden_repeats": list(state.forbidden_repeats | set(patch.forbidden_repeats)),
            "new_clues": patch.new_clues,
            "new_items": patch.new_items,
            "spawned_npcs": patch.spawned_npcs,
            "visible_npcs": visible_npcs,
            "relevant_clues": relevant_clues,
            "relevant_items": relevant_items,
            "danger_level": state.danger_level,
            "revealed_bible": revealed_bible or {},
            # soft_lookahead：可能方向（非既定事實，story 不得當成已發生）
            "soft_lookahead": list(getattr(progress, "soft_lookahead", []) or []),
            "instruction": (
                "Realize only the committed event. Do not repeat forbidden events. "
                "Do not decide new world-state changes outside the provided obligations. "
                "soft_lookahead is only possible directions; never present it as already happened."
            ),
        }
