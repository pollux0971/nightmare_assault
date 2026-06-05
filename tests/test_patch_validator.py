"""SK01 PatchValidator 測試。"""
import pytest

from core.patch_validator import PatchValidator, PatchValidationError
from core.progress_models import EventPatch, GameState, PatchOp


def _state(**kw):
    base = dict(version=1, beat_number=0, current_scene="scene.a", scene_phase="beginning")
    base.update(kw)
    return GameState(**base)


def _patch(**kw):
    base = dict(base_version=1, event_id="ev1", ops=[], progress_delta=["event_resolved"])
    base.update(kw)
    return EventPatch(**base)


def test_reject_no_progress_delta():
    with pytest.raises(PatchValidationError):
        PatchValidator().validate(_patch(progress_delta=[]), _state())


def test_reject_base_version_mismatch():
    with pytest.raises(PatchValidationError):
        PatchValidator().validate(_patch(base_version=99), _state())


def test_reject_forbidden_event():
    st = _state(forbidden_repeats={"ev1"})
    with pytest.raises(PatchValidationError):
        PatchValidator().validate(_patch(event_id="ev1"), st)


def test_apply_scene_and_version_bump():
    st = _state()
    patch = _patch(ops=[PatchOp("set", "current_scene", "scene.b")],
                   progress_delta=["location_changed"], forbidden_repeats=["ev1"])
    new = PatchValidator().apply(patch, st)
    assert new.current_scene == "scene.b"
    assert new.version == 2 and new.beat_number == 1
    assert "ev1" in new.recent_events and "ev1" in new.forbidden_repeats
    assert new.event_status["ev1"] == "resolved"


def test_apply_clue_add_dedup():
    st = _state()
    patch = _patch(ops=[PatchOp("add", "clues.clue.scratch_marks",
                                {"title": "抓痕", "content": "從內側往外"})],
                   progress_delta=["new_clue_added"])
    new = PatchValidator().apply(patch, st)
    assert "clue.scratch_marks" in new.clues
    assert new.clues["clue.scratch_marks"].title == "抓痕"


def test_apply_npc_dotted_id():
    st = _state()
    patch = _patch(ops=[PatchOp("set", "npcs.npc.night_nurse.visible", True),
                        PatchOp("set", "npcs.npc.night_nurse.current_scene", "scene.b")],
                   progress_delta=["npc_spawned"])
    new = PatchValidator().apply(patch, st)
    assert "npc.night_nurse" in new.npcs            # 含點的 npc id 正確
    assert new.npcs["npc.night_nurse"].visible is True
    assert new.npcs["npc.night_nurse"].current_scene == "scene.b"


def test_apply_inventory_and_danger():
    st = _state()
    patch = _patch(ops=[PatchOp("add", "inventory.item.key", {"name": "鏽鑰匙"}),
                        PatchOp("inc", "danger_level", 2)],
                   progress_delta=["item_added", "danger_level_changed"])
    new = PatchValidator().apply(patch, st)
    assert new.inventory["item.key"].name == "鏽鑰匙"
    assert new.danger_level == 2
