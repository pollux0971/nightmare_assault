"""U16–U19 webview API 無頭測試：用 FakeWindow + mock loop 驗證 NA.* 推送契約。

不需要 GUI/display；驗證 API 邏輯與前後端事件契約。視覺需使用者實機跑 main.py。
"""
from webview_app import API
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus
from core.models import (
    SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput, CompactorOutput,
)

STORY = '{"situation_recap":"前方有門","decision_type":"action","suggested_options":[{"text":"開門","tone":"bold"}],"beat_meta":{"beat_number":0}}'


class FakeCaller:
    def __init__(self, hard_triggers=None):
        self.setup_out = SetupOutput(
            real_bible={"revelation_pool": [], "ending_conditions": [], "hard_triggers": hard_triggers or []},
            npc_registry=[NPCBible(name="A", profession="醫生", personality="mysterious",
                                   voice_sample="…", public_face="冷", secret_core="S",
                                   self_aware=True, appearance="")],
            protagonist={"name": "林默"},
            scene_registry=SceneRegistry(current_location="hall",
                                         known_locations=[Location(id="hall", name="大廳", description="")]),
            opening_sequence=["你在大廳醒來。"])

    def call(self, agent, context, output_model=None, temperature=None):
        if agent == "setup":
            return self.setup_out
        if agent == "warden":
            return WardenOutput(directive_to_story="繼續")
        if agent == "compactor":
            return CompactorOutput(compressed_summary="x", ledger_updates=[], archived_beats=[],
                                   preserved_foreshadowings=[], final_usage_estimate=0.1)
        raise AssertionError(agent)

    def stream(self, agent, context, temperature=None):
        for t in ["走廊很暗，盡頭有光。", "<<<DECISION>>>", STORY]:
            yield t


class FakeWindow:
    def __init__(self):
        self.calls = []

    def evaluate_js(self, js):
        self.calls.append(js)


def _api(hard_triggers=None):
    api = API(window=FakeWindow())
    api._config = {"api_key": "test", "base_url": "x", "agent_models": {}, "timeout": 5}
    api._make_loop = lambda: BeatLoop(FakeCaller(hard_triggers), Blackboard(),
                                      Database(), SignalBus(), run_id="t", use_kernel=False)
    return api


def test_check_config_true_when_key_present():
    assert _api().check_config()["configured"] is True


def test_check_config_false_without_key():
    api = API(window=FakeWindow())
    api._config = {}
    assert api.check_config()["configured"] is False


def test_start_pushes_opening_decision_and_complete():
    api = _api()
    api._loop = api._make_loop()
    api._run_start({"theme": "x", "npc_count": 1})
    joined = "\n".join(api._window.calls)
    assert "NA.onOpening" in joined
    assert "NA.appendToken" in joined
    assert "NA.onDecision" in joined
    assert "NA.onBeatComplete" in joined
    assert "NA.onStatus" in joined
    assert "NA.onProgress" in joined and "建構世界真相" in joined


def test_step_streams_and_advances():
    api = _api()
    api._loop = api._make_loop()
    api._run_start({"theme": "x", "npc_count": 1})
    api._window.calls.clear()
    api._run_step("我開門", "option")
    joined = "\n".join(api._window.calls)
    assert "NA.appendToken" in joined and "NA.onDecision" in joined
    # HUD 進度：守門人裁決 → 揭露閘門 → 編織夢魘
    assert "守門人裁決" in joined and "揭露閘門" in joined and "編織夢魘" in joined
    assert api._loop.beat_number == 2


def test_warden_ending_pushes_onEnding():
    api = _api(hard_triggers=["喝藥水"])
    api._loop = api._make_loop()
    api._run_start({"theme": "x", "npc_count": 1})
    api._window.calls.clear()
    api._run_step("我決定喝藥水", "free_text")
    assert "NA.onEnding" in "\n".join(api._window.calls)


def test_inventory_does_not_leak_key_item():
    api = _api()
    api._loop = api._make_loop()
    api._run_start({"theme": "x", "npc_count": 1})
    for it in api.get_inventory():
        assert "is_key_item" not in it


def test_game_state_shape():
    st = _api().get_game_state()
    for k in ("run_id", "state", "busy", "beat_number"):
        assert k in st


def test_save_config_roundtrip(tmp_path):
    api = API(window=FakeWindow(), config_path=tmp_path / "config.json")
    assert api.save_config({"api_key": "abc"})["ok"] is True
    assert api.check_config()["configured"] is True
