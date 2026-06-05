"""SK02 ProgressKernel + SceneGraphProvider 測試（移植 + 調整路徑）。"""
from pathlib import Path

from core.progress_kernel import ProgressKernel
from core.scene_graph import StaticOpeningSceneGraphProvider, DEFAULT_GRAPH_PATH
from core.patch_validator import PatchValidator
from core.progress_models import GameState, SceneState, Obligation
from core.progress_context import ContextBuilder


def _provider():
    return StaticOpeningSceneGraphProvider(DEFAULT_GRAPH_PATH)


def make_state():
    return GameState(
        version=1, beat_number=0,
        current_scene="scene.ward_101", scene_phase="beginning",
        scenes={"scene.ward_101": SceneState(id="scene.ward_101")},
        open_obligations=[
            Obligation(id="obl.leave_starting_room", kind="transition", description="Leave the starting ward."),
            Obligation(id="obl.seed_first_clue", kind="grant_clue", description="Seed the first clue."),
            Obligation(id="obl.introduce_first_npc", kind="spawn_npc", description="Introduce first NPC."),
        ],
    )


def test_provider_loads_and_start_scene():
    p = _provider()
    assert p.start_scene() == "scene.ward_101"
    assert p.graph().get("events")
    assert len(p.default_obligations()) == 3


def test_open_door_resolves_and_forbids_repeat():
    kernel = ProgressKernel.from_provider(_provider())
    validator = PatchValidator()
    state = make_state()

    result = kernel.resolve_player_action("我打開病房門", state)
    new_state = validator.apply(result.patch, state)

    assert result.committed_event == "event.open_door_101"
    assert new_state.current_scene == "scene.corridor_2f"
    assert "event.open_door_101" in new_state.forbidden_repeats
    assert "ask_open_door_101" in new_state.forbidden_repeats
    assert "clue.scratch_marks" in new_state.clues


def test_next_candidate_is_not_open_same_door():
    kernel = ProgressKernel.from_provider(_provider())
    validator = PatchValidator()
    state = validator.apply(kernel.resolve_player_action("開門", make_state()).patch, make_state())

    result2 = kernel.resolve_player_action("我往前走，看看走廊", state)
    assert result2.committed_event != "event.open_door_101"


def test_npc_spawn_in_corridor():
    kernel = ProgressKernel.from_provider(_provider())
    validator = PatchValidator()
    state = validator.apply(kernel.resolve_player_action("開門", make_state()).patch, make_state())

    result2 = kernel.resolve_player_action("我觀察走廊", state)
    state2 = validator.apply(result2.patch, state)

    assert "npc.night_nurse" in state2.npcs
    assert state2.npcs["npc.night_nurse"].visible is True


def test_dummy_event_when_no_candidate():
    kernel = ProgressKernel.from_provider(_provider())
    # 不存在的場景 → 無候選 → dummy escalation（仍有 progress_delta）
    state = GameState(version=1, beat_number=0, current_scene="scene.void", scene_phase="beginning")
    result = kernel.resolve_player_action("發呆", state)
    assert result.patch.progress_delta  # 至少一個 delta（sparse fallback 有後果）
    assert "fallback" in result.committed_event


def test_context_builder_minimal_no_real_bible():
    kernel = ProgressKernel.from_provider(_provider())
    state = make_state()
    result = kernel.resolve_player_action("開門", state)
    ctx = ContextBuilder().build_story_context(state, result, revealed_bible={"fragments": []})
    assert "real_bible" not in ctx
    assert "archive_transcript" not in ctx
    assert ctx["committed_event"] == "event.open_door_101"
    assert "narrative_obligations" in ctx
