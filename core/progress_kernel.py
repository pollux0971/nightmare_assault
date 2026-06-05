"""ProgressKernel（SK02，移植自穩定化補丁）。

rule-first：用小型 graph + 簡單評分決定推進事件。world-state 由此決定，story 只 realize。
依賴 SceneGraphProvider 取得標準 graph，不直接綁死病房劇情。
"""
from __future__ import annotations

from pathlib import Path
from typing import Any

from core.progress_models import EventPatch, GameState, PatchOp, ProgressResult
from core.scene_graph import SceneGraphProvider, StaticOpeningSceneGraphProvider


class ProgressKernel:
    def __init__(self, provider: SceneGraphProvider):
        self.provider = provider
        self.graph = provider.graph()
        self.events = {e["id"]: e for e in self.graph.get("events", [])}

    @classmethod
    def from_provider(cls, provider: SceneGraphProvider) -> "ProgressKernel":
        return cls(provider)

    @classmethod
    def from_json(cls, path: str | Path) -> "ProgressKernel":
        return cls(StaticOpeningSceneGraphProvider(path))

    def resolve_player_action(self, player_input: str, state: GameState,
                              warden: dict | None = None) -> ProgressResult:
        intent_tags = self._normalize_intent(player_input)

        candidates = self._candidate_events(state, intent_tags)
        if not candidates:
            candidates = [self._make_fallback_event(state)]   # 越界/無候選 → 仍有後果

        chosen = max(candidates, key=lambda e: self._score_event(e, state, intent_tags))
        patch = self._event_to_patch(chosen, state)

        return ProgressResult(
            accepted=True,
            patch=patch,
            committed_event=chosen["id"],
            explanation=f"Chosen event {chosen['id']} for intents {intent_tags}.",
            soft_lookahead=self._soft_lookahead(state, chosen),   # 每 beat 重算
        )

    def _soft_lookahead(self, state: GameState, chosen: dict) -> list[str]:
        """**每 beat 重算**的可能方向（非既定事實）：預覽本 beat 後可能出現的事件。

        看 chosen 事件落地後的場景，列出該場景未禁/未滿次數的其他事件提示。不存進 GameState。
        """
        next_scene = state.current_scene
        for op in chosen.get("effects", []):
            if op.get("path") == "current_scene" and op.get("op") == "set":
                next_scene = op.get("value")
        hints: list[str] = []
        for ev in self.graph.get("events", []):
            if ev.get("scene_id") != next_scene or ev["id"] == chosen["id"]:
                continue
            if ev["id"] in state.forbidden_repeats:
                continue
            if state.event_counts.get(ev["id"], 0) >= ev.get("max_repeat", 1):
                continue
            obls = ev.get("narrative_obligations") or []
            hints.append((obls[0] if obls else ev["id"])[:42])
            if len(hints) >= 3:
                break
        return hints

    def _normalize_intent(self, text: str) -> list[str]:
        t = (text or "").lower()
        tags: list[str] = []
        if any(k in t for k in ["開門", "打開", "門", "open"]):
            tags.append("open")
        if any(k in t for k in ["觀察", "檢查", "看", "搜", "search", "inspect"]):
            tags.append("inspect")
        if any(k in t for k in ["呼喊", "叫", "喊", "help", "call"]):
            tags.append("call")
        if any(k in t for k in ["走", "前進", "離開", "進", "go", "move"]):
            tags.append("move")
        return tags or ["free"]

    def _candidate_events(self, state: GameState, intent_tags: list[str]) -> list[dict[str, Any]]:
        candidates = []
        for event in self.graph.get("events", []):
            if event.get("scene_id") != state.current_scene:
                continue
            if event["id"] in state.forbidden_repeats:
                continue
            if state.event_counts.get(event["id"], 0) >= event.get("max_repeat", 1):
                continue
            if not self._preconditions_ok(event.get("preconditions", []), state):
                continue
            if self._intent_overlap(event.get("intent_tags", []), intent_tags):
                candidates.append(event)
        return candidates

    def _preconditions_ok(self, preconditions: list[str], state: GameState) -> bool:
        for precond in preconditions:
            if precond.startswith("not_recent:"):
                event_id = precond.split(":", 1)[1]
                if event_id in state.recent_events:
                    return False
            elif precond.startswith("scene_phase:"):
                phase = precond.split(":", 1)[1]
                if state.scene_phase != phase:
                    return False
            elif precond.startswith("clue_missing:"):
                clue_id = precond.split(":", 1)[1]
                if clue_id in state.clues:
                    return False
            elif precond.startswith("npc_not_visible:"):
                npc_id = precond.split(":", 1)[1]
                if state.npcs.get(npc_id) and state.npcs[npc_id].visible:
                    return False
        return True

    def _intent_overlap(self, event_tags: list[str], input_tags: list[str]) -> bool:
        return bool(set(event_tags) & set(input_tags)) or "free" in input_tags or "free" in event_tags

    def _score_event(self, event: dict[str, Any], state: GameState, intent_tags: list[str]) -> float:
        score = 0.0
        open_obls = {o.id for o in state.open_obligations if o.status == "open"}
        score += 3.0 * len(set(event.get("satisfies", [])) & open_obls)
        score += 2.0 * len(event.get("grants_clues", []))
        score += 2.0 * len(event.get("spawns_npcs", []))
        score += 1.0 * len(set(event.get("intent_tags", [])) & set(intent_tags))
        score -= 3.0 if event["id"] in state.recent_events else 0.0
        return score

    def _event_to_patch(self, event: dict[str, Any], state: GameState) -> EventPatch:
        ops = [PatchOp(**op) for op in event.get("effects", [])]
        return EventPatch(
            base_version=state.version,
            event_id=event["id"],
            ops=ops,
            progress_delta=event.get("progress_delta", []),
            narrative_obligations=event.get("narrative_obligations", []),
            forbidden_repeats=event.get("forbidden_after", []),
            new_clues=event.get("grants_clues", []),
            new_items=event.get("grants_items", []),
            spawned_npcs=event.get("spawns_npcs", []),
            debug_reason=event.get("debug_reason", ""),
        )

    @staticmethod
    def _has_recent_clue(state: GameState, window: int = 2) -> bool:
        return any(c.first_seen_beat >= state.beat_number - window for c in state.clues.values())

    def _make_fallback_event(self, state: GameState) -> dict[str, Any]:
        """玩家越界/無合法候選時的 **context-aware sparse fallback**：

        永遠有後果（≥1 delta）；依當前缺口選最該推進的方向：
          缺線索 → 種一條小線索；缺 NPC 痕跡 → 留痕跡（壓力+痕跡）；否則 → 升壓。
        """
        beat = state.beat_number
        scene = state.current_scene
        base = {"scene_id": scene, "intent_tags": ["free"], "forbidden_after": []}

        if not state.clues or not self._has_recent_clue(state):
            cid = f"clue.trace_{beat}"
            return {**base, "id": f"fallback.clue.{scene}.{beat}",
                    "effects": [{"op": "add", "path": f"clues.{cid}",
                                 "value": {"title": "新的不安",
                                           "content": "你偏離了原本的念頭，卻因此注意到一個之前忽略的細節。",
                                           "tags": ["fallback"]}}],
                    "progress_delta": ["new_clue_added"],
                    "narrative_obligations": ["玩家偏離預期時，環境給出一條新的小線索，而不是停滯。"],
                    "debug_reason": "sparse fallback: seed clue (no recent clue)."}

        if not any(n.visible for n in state.npcs.values()) and beat >= 3:
            return {**base, "id": f"fallback.trace.{scene}.{beat}",
                    "effects": [{"op": "inc", "path": "danger_level", "value": 1}],
                    "progress_delta": ["npc_trace_added", "danger_level_changed"],
                    "narrative_obligations": [
                        "出現某物/某人曾經過的痕跡（腳印/拖痕/餘溫），暗示有東西不遠，且壓力升高。"],
                    "debug_reason": "sparse fallback: npc trace (no visible npc)."}

        return {**base, "id": f"fallback.escalate.{scene}.{beat}",
                "effects": [{"op": "inc", "path": "danger_level", "value": 1}],
                "progress_delta": ["danger_level_changed"],
                "narrative_obligations": [
                    "玩家沒有在原地停滯；環境主動升高壓力，讓下一個選擇不同於上一回合。"],
                "debug_reason": "sparse fallback: escalation."}
