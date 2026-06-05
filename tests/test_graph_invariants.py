"""SK10 graph 不變式測試（多出口 / 多解法 / 多 intent / 每事件有後果）。"""
import pytest

from core.graph_invariants import validate_graph_invariants, check_graph_invariants, GraphInvariantError
from core.scene_graph import build_themed_graph, StaticOpeningSceneGraphProvider, DEFAULT_GRAPH_PATH

THEMED = {
    "scene_registry": {"current_location": "scene.deck",
                       "known_locations": [{"id": "scene.deck", "name": "甲板"},
                                           {"id": "scene.lab", "name": "研究艙"},
                                           {"id": "scene.core", "name": "反應爐"}]},
    "npc_registry": [{"name": "陳博士"}],
    "real_bible": {"world_truth": {"the_threat_is": "低語"}},
}


def _scene_events(graph):
    d = {}
    for e in graph["events"]:
        d.setdefault(e["scene_id"], []).append(e)
    return d


def test_generated_graph_passes_invariants():
    g = build_themed_graph(THEMED)               # builder 末端已 validate；這裡再確認
    validate_graph_invariants(g, start_scene=g["start_scene"])


def test_sparse_generated_graph_passes_invariants():
    g = build_themed_graph({})
    validate_graph_invariants(g, start_scene=g["start_scene"])


def test_each_scene_has_multiple_events():
    g = build_themed_graph(THEMED)
    for scene, evs in _scene_events(g).items():
        assert len(evs) >= 2, f"{scene} <2 events"


def test_start_scene_has_multiple_exits():
    g = build_themed_graph(THEMED)
    start = g["start_scene"]
    exits = [e for e in g["events"] if e["scene_id"] == start
             and any(op["path"] == "current_scene" for op in e["effects"])]
    assert len(exits) >= 2                        # 多出口


def test_leave_obligation_has_multiple_solutions():
    g = build_themed_graph(THEMED)
    sols = [e for e in g["events"] if "obl.leave_starting_room" in e.get("satisfies", [])]
    assert len(sols) >= 2                         # 多解法


def test_gateways_have_multiple_intents_and_deltas():
    g = build_themed_graph(THEMED)
    for e in g["events"]:
        if any(op["path"] == "current_scene" for op in e["effects"]):
            assert len(e["intent_tags"]) >= 2     # 多 intent
        assert e["progress_delta"]                # 每事件有後果


def test_invalid_graph_is_rejected():
    bad = {"start_scene": "s", "events": [
        {"id": "x", "scene_id": "s", "intent_tags": ["move"], "effects": [], "progress_delta": []},
    ]}
    with pytest.raises(GraphInvariantError):
        validate_graph_invariants(bad, start_scene="s")
    assert check_graph_invariants(bad, start_scene="s") is False


def test_static_graph_has_multi_events_per_scene():
    # 靜態開場圖（authored fallback）：至少每場景多事件（寬鬆，不強制多出口）
    g = StaticOpeningSceneGraphProvider(DEFAULT_GRAPH_PATH).graph()
    for scene, evs in _scene_events(g).items():
        assert len(evs) >= 2
