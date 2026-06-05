"""SK11 每個行動有後果 + context-aware sparse fallback 測試。"""
from core.scene_graph import StaticOpeningSceneGraphProvider, DEFAULT_GRAPH_PATH
from core.progress_kernel import ProgressKernel
from core.patch_validator import PatchValidator
from core.progress_models import GameState, NPCPresence


def _kernel():
    return ProgressKernel.from_provider(StaticOpeningSceneGraphProvider(DEFAULT_GRAPH_PATH))


def _void(**kw):
    base = dict(version=1, beat_number=0, current_scene="scene.void", scene_phase="beginning")
    base.update(kw)
    return GameState(**base)


def test_out_of_bounds_input_still_has_consequence():
    # 場景無候選（越界）+ 亂打 → 仍有 ≥1 delta 的 fallback
    res = _kernel().resolve_player_action("asdfgh@@@亂打一通", _void())
    assert res.patch.progress_delta
    assert PatchValidator().apply(res.patch, _void())   # 可提交（≥1 delta）


def test_no_clue_seeds_clue():
    res = _kernel().resolve_player_action("發呆", _void())   # 無線索
    assert "new_clue_added" in res.patch.progress_delta
    new = PatchValidator().apply(res.patch, _void())
    assert any(c.tags == ["fallback"] for c in new.clues.values())


def test_clue_present_no_npc_leaves_trace():
    st = _void(beat_number=5)
    # 先有一條近期線索
    from core.progress_models import LedgerEntry
    st.clues["c1"] = LedgerEntry(id="c1", title="t", content="", source_event="e", first_seen_beat=5)
    res = _kernel().resolve_player_action("亂走", st)
    assert "npc_trace_added" in res.patch.progress_delta


def test_clue_and_npc_present_escalates():
    st = _void(beat_number=6)
    from core.progress_models import LedgerEntry
    st.clues["c1"] = LedgerEntry(id="c1", title="t", content="", source_event="e", first_seen_beat=6)
    st.npcs["n1"] = NPCPresence(id="n1", name="x", visible=True)
    res = _kernel().resolve_player_action("發呆", st)
    assert res.patch.progress_delta == ["danger_level_changed"]


def test_every_input_produces_delta_across_inputs():
    k = _kernel()
    v = PatchValidator()
    st = _void()
    for txt in ["???", "隨便", "@@@@", "嗯", "看天花板", "原地轉圈"]:
        res = k.resolve_player_action(txt, st)
        assert res.patch.progress_delta, f"無後果: {txt}"
        st = v.apply(res.patch, st)
