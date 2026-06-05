"""SK05 持久化 + debug API 測試。"""
from webview_app import API
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus
from core.patch_validator import PatchValidator
from core.progress_models import EventPatch, GameState, PatchOp

from tests.test_progress_integration import FakeCaller


def _state():
    return GameState(version=1, beat_number=0, current_scene="s", scene_phase="beginning")


def test_clue_dedup_keeps_first():
    v = PatchValidator()
    st = _state()
    p1 = EventPatch(base_version=1, event_id="e1",
                    ops=[PatchOp("add", "clues.c1", {"title": "第一次", "content": "A"})],
                    progress_delta=["new_clue_added"])
    st = v.apply(p1, st)
    first_beat = st.clues["c1"].first_seen_beat
    p2 = EventPatch(base_version=st.version, event_id="e2",
                    ops=[PatchOp("add", "clues.c1", {"title": "第二次", "content": "B"})],
                    progress_delta=["new_clue_added"])
    st = v.apply(p2, st)
    assert st.clues["c1"].title == "第一次"            # 不被覆寫
    assert st.clues["c1"].content == "A"
    assert st.clues["c1"].first_seen_beat == first_beat


def test_get_debug_state_via_api():
    api = API(window=None)
    api._config = {"api_key": "x", "base_url": "y", "agent_models": {}, "timeout": 5}
    caller = FakeCaller()
    api._make_loop = lambda: BeatLoop(caller, Blackboard(), Database(), SignalBus(),
                                      run_id="t", use_kernel=True)
    api._loop = api._make_loop()
    api._run_start({"theme": "x", "npc_count": 1})
    api._run_step("我打開病房門", "free_text")
    d = api.get_debug_state()
    assert d["kernel"] is True
    assert d["scene"] == "scene.beyond"
    assert any(c["id"] == "clue.first_sign" for c in d["clues"])


def test_kernel_state_persisted_in_snapshot():
    db = Database()
    loop = BeatLoop(FakeCaller(), Blackboard(), db, SignalBus(), run_id="p", use_kernel=True)
    loop.start({"theme": "x", "npc_count": 1})
    loop.step("我打開病房門")
    saved = db.load_beat("p", loop.beat_number)
    assert saved is not None
    blob = saved["blackboard_snapshot_json"]
    assert "clue.first_sign" in blob       # 線索進 snapshot
    assert "progress_state" in blob           # 完整 progress state 進 snapshot（供回溯）
