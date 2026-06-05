"""SK12 soft lookahead（每 beat 重算、非既定、不持久）測試。"""
from dataclasses import fields

from core.scene_graph import StaticOpeningSceneGraphProvider, DEFAULT_GRAPH_PATH
from core.progress_kernel import ProgressKernel
from core.patch_validator import PatchValidator
from core.progress_context import ContextBuilder
from core.progress_models import GameState


def _kernel():
    return ProgressKernel.from_provider(StaticOpeningSceneGraphProvider(DEFAULT_GRAPH_PATH))


def _state():
    return GameState(version=1, beat_number=0, current_scene="scene.ward_101", scene_phase="beginning")


def test_lookahead_present_and_previews_next_scene():
    res = _kernel().resolve_player_action("我打開病房門", _state())
    # 開門後落地 corridor_2f；lookahead 應預覽走廊的事件（護士/拖行聲）
    assert res.soft_lookahead
    assert any("護士" in h or "拖行" in h or "壓力" in h for h in res.soft_lookahead)


def test_lookahead_recomputes_as_state_changes():
    k = _kernel()
    v = PatchValidator()
    st = _state()
    look1 = k.resolve_player_action("開門", st).soft_lookahead
    st = v.apply(k.resolve_player_action("開門", st).patch, st)   # 進 corridor
    look2 = k.resolve_player_action("觀察走廊", st).soft_lookahead
    assert look1 != look2                       # 隨狀態重算而改變


def test_lookahead_not_stored_in_gamestate():
    names = {f.name for f in fields(GameState)}
    assert "soft_lookahead" not in names        # 不持久於狀態（每 beat 重算）


def test_context_includes_soft_lookahead_as_nonbinding():
    res = _kernel().resolve_player_action("開門", _state())
    ctx = ContextBuilder().build_story_context(_state(), res, revealed_bible={})
    assert "soft_lookahead" in ctx
    assert "possible directions" in ctx["instruction"]   # 標示非既定
