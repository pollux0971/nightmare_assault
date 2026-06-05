"""SK08 GeneratedSceneGraphProvider / build_themed_graph 測試。"""
from core.scene_graph import build_themed_graph, GeneratedSceneGraphProvider
from core.progress_kernel import ProgressKernel
from core.patch_validator import PatchValidator
from core import progress_bridge as bridge


def _deep_sea_snapshot():
    return {
        "scene_registry": {
            "current_location": "scene.deck",
            "known_locations": [
                {"id": "scene.deck", "name": "潛水艙甲板", "description": "鏽蝕的金屬甲板"},
                {"id": "scene.lab", "name": "研究艙", "description": "閃爍的儀器"},
            ],
        },
        "npc_registry": [{"name": "陳博士", "self_aware": True}],
        "real_bible": {"world_truth": {"the_threat_is": "水壓中傳來不屬於人的低語"}},
    }


def test_build_graph_is_themed_and_valid():
    g = build_themed_graph(_deep_sea_snapshot())
    assert g["start_scene"] == "scene.deck"
    ids = {e["id"] for e in g["events"]}
    assert {"event.exit_start", "event.meet_npc", "event.mid_escalate"} <= ids
    assert "event.exit_alt" in ids               # 多出口
    # exit 事件主題化：進入研究艙
    exit_ev = next(e for e in g["events"] if e["id"] == "event.exit_start")
    assert any(op["path"] == "current_scene" and op["value"] == "scene.lab" for op in exit_ev["effects"])
    assert "研究艙" in " ".join(exit_ev["narrative_obligations"])
    assert exit_ev["progress_delta"]                       # ≥1 delta


def test_kernel_runs_through_themed_graph():
    provider = GeneratedSceneGraphProvider(_deep_sea_snapshot())
    kernel = ProgressKernel.from_provider(provider)
    validator = PatchValidator()
    gs = bridge.init_game_state(provider)
    assert gs.current_scene == "scene.deck"

    gs = validator.apply(kernel.resolve_player_action("我打開艙門前進", gs).patch, gs)
    assert gs.current_scene == "scene.lab"
    assert "clue.first_sign" in gs.clues

    res = kernel.resolve_player_action("我觀察研究艙", gs)
    gs = validator.apply(res.patch, gs)
    assert res.committed_event == "event.meet_npc"
    assert "npc.陳博士" in gs.npcs                          # 含中文的 npc id 正確（validator rsplit）
    assert gs.npcs["npc.陳博士"].visible is True


def test_no_repeat_exit_in_themed_graph():
    provider = GeneratedSceneGraphProvider(_deep_sea_snapshot())
    kernel = ProgressKernel.from_provider(provider)
    validator = PatchValidator()
    gs = bridge.init_game_state(provider)
    gs = validator.apply(kernel.resolve_player_action("開門", gs).patch, gs)
    res2 = kernel.resolve_player_action("往前走", gs)
    assert res2.committed_event != "event.exit_start"
    assert "ask_exit_start" in gs.forbidden_repeats


def test_sparse_setup_output_still_valid():
    # 幾乎空的 setup 輸出 → 仍 synthesize 出可用 graph
    g = build_themed_graph({})
    assert g["start_scene"]
    assert len(g["events"]) >= 3
    provider = GeneratedSceneGraphProvider({})
    kernel = ProgressKernel.from_provider(provider)
    validator = PatchValidator()
    gs = bridge.init_game_state(provider)
    gs = validator.apply(kernel.resolve_player_action("前進", gs).patch, gs)
    assert gs.current_scene == "scene.beyond"              # 合成的第二場景
