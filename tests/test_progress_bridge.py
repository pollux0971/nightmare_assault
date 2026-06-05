"""SK03 progress_bridge 測試。"""
from core.blackboard import Blackboard
from core.scene_graph import StaticOpeningSceneGraphProvider, DEFAULT_GRAPH_PATH
from core.progress_kernel import ProgressKernel
from core.patch_validator import PatchValidator
from core import progress_bridge as B


def _provider():
    return StaticOpeningSceneGraphProvider(DEFAULT_GRAPH_PATH)


def test_init_game_state():
    gs = B.init_game_state(_provider())
    assert gs.current_scene == "scene.ward_101"
    assert len(gs.open_obligations) == 3
    assert gs.version == 1 and gs.beat_number == 0


def test_snapshot_roundtrip():
    kernel = ProgressKernel.from_provider(_provider())
    gs = B.init_game_state(_provider())
    gs = PatchValidator().apply(kernel.resolve_player_action("開門", gs).patch, gs)
    d = B.to_snapshot_dict(gs)
    assert isinstance(d["forbidden_repeats"], list)   # set → list（JSON-able）
    gs2 = B.from_snapshot_dict(d)
    assert gs2.current_scene == gs.current_scene
    assert gs2.forbidden_repeats == gs.forbidden_repeats
    assert "clue.scratch_marks" in gs2.clues


def test_sync_to_blackboard():
    kernel = ProgressKernel.from_provider(_provider())
    validator = PatchValidator()
    gs = B.init_game_state(_provider())
    gs = validator.apply(kernel.resolve_player_action("開門", gs).patch, gs)        # → corridor + clue
    gs = validator.apply(kernel.resolve_player_action("觀察走廊", gs).patch, gs)    # → nurse spawn

    bb = Blackboard()
    B.sync_to_blackboard(gs, bb)
    snap = bb.snapshot()
    assert snap["scene_registry"]["current_location"] == "scene.corridor_2f"
    assert any(c["id"] == "clue.scratch_marks" for c in snap["game_meta"]["clues"])
    assert any(n["id"] == "npc.night_nurse" for n in snap["game_meta"]["visible_npcs"])
    assert snap["game_meta"]["progress_kernel"] is True


def test_debug_state_shape():
    gs = B.init_game_state(_provider())
    d = B.debug_state(gs)
    for k in ("scene", "phase", "open_obligations", "clues", "inventory", "visible_npcs"):
        assert k in d
