"""Scene graph 不變式（SK10）— 強制「多出口 / 多解法 / 多 intent」。

generated graph 一律過這些不變式（builder 末端 assert），守住敘事規劃品質：
- INV1 每個場景 ≥2 個事件（玩家永遠有多個選擇，非單線）。
- INV2 起始場景 ≥2 個「改變場景」的出口；其餘非終端場景 ≥1 出口。
- INV3 「離開起始場景」obligation 由 ≥2 個事件可滿足（多解法）。
- INV4 每個 gateway（改變場景）事件 ≥2 個 intent_tags（多種觸發方式）。
- INV5 每個事件 ≥1 progress_delta（每個行動有後果，與 PatchValidator 一致）。
"""
from __future__ import annotations


class GraphInvariantError(ValueError):
    pass


def _changes_scene(ev: dict) -> bool:
    return any(op.get("path") == "current_scene" and op.get("op") == "set"
              for op in ev.get("effects", []))


def validate_graph_invariants(graph: dict, start_scene: str | None = None,
                              require_multi_exit: bool = True,
                              leave_obligation: str = "obl.leave_starting_room") -> None:
    events = graph.get("events", [])
    if not events:
        raise GraphInvariantError("graph has no events")
    start = start_scene or graph.get("start_scene")

    by_scene: dict[str, list[dict]] = {}
    for ev in events:
        by_scene.setdefault(ev.get("scene_id"), []).append(ev)

    for scene, evs in by_scene.items():
        # INV1
        if len(evs) < 2:
            raise GraphInvariantError(f"scene '{scene}' has <2 events (INV1)")
        exits = [e for e in evs if _changes_scene(e)]
        if require_multi_exit:
            if scene == start and len(exits) < 2:
                raise GraphInvariantError(f"start scene '{scene}' has <2 exits (INV2)")
            if scene != start and len(exits) < 1:
                raise GraphInvariantError(f"scene '{scene}' has no exit (INV2)")

    # INV3：離開起始場景 ≥2 解法
    if require_multi_exit:
        solutions = [e for e in events if leave_obligation in e.get("satisfies", [])]
        if len(solutions) < 2:
            raise GraphInvariantError(f"obligation '{leave_obligation}' has <2 solutions (INV3)")

    for ev in events:
        # INV4：gateway 多 intent
        if _changes_scene(ev) and len(ev.get("intent_tags", [])) < 2:
            raise GraphInvariantError(f"gateway event '{ev.get('id')}' has <2 intent_tags (INV4)")
        # INV5：每個事件有後果
        if not ev.get("progress_delta"):
            raise GraphInvariantError(f"event '{ev.get('id')}' has no progress_delta (INV5)")


def check_graph_invariants(graph: dict, **kw) -> bool:
    try:
        validate_graph_invariants(graph, **kw)
        return True
    except GraphInvariantError:
        return False
