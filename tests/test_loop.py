"""U15 beat 主迴圈整合測試（mock caller，不打真網路）。"""
import json

from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus
from core.constants import EVT_BEAT_COMPLETED, EVT_ENDING_TRIGGERED, BEAT_WINDOW_SIZE
from core.models import (
    SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput, CompactorOutput,
)

STORY_JSON = ('{"situation_recap":"前方有一道門。","decision_type":"action",'
              '"suggested_options":[{"text":"開門","tone":"bold"}],'
              '"beat_meta":{"beat_number":0}}')


def make_setup_output(hard_triggers=None, forbidden="機密真相不可外洩XYZ"):
    return SetupOutput(
        real_bible={
            "world_truth": {"what_really_happened": forbidden,
                            "the_threat_is": "走廊裡的東西", "deadly_rule": "不可喝藥水"},
            "hard_triggers": hard_triggers or [],
            "ending_conditions": [],
            "revelation_pool": [
                {"id": "f1", "type": "knowledge", "content": "隱藏碎片ABC",
                 "reveal_condition": {"min_beats": 2}},
            ],
        },
        npc_registry=[NPCBible(name="張醫生", profession="醫生", personality="mysterious",
                               voice_sample="你不該來這。", public_face="冷靜",
                               secret_core="他就是兇手SECRET", self_aware=True, appearance="白袍")],
        protagonist={"name": "林默", "starting_situation": "在大廳醒來"},
        scene_registry=SceneRegistry(current_location="hall",
                                     known_locations=[Location(id="hall", name="大廳",
                                                               description="陰暗", exits=["corridor"])]),
        opening_sequence=["你在大廳醒來，血腥味瀰漫。"],
    )


class FakeCaller:
    """依 agent 分派的假 caller；記錄 story context 供防暴雷檢查。"""
    def __init__(self, setup_out, warden_out=None, compactor_out=None, story_tokens=None):
        self.setup_out = setup_out
        self.warden_out = warden_out or WardenOutput(directive_to_story="繼續，氣氛緊張")
        self.compactor_out = compactor_out or CompactorOutput(
            compressed_summary="主角仍在醫院，伏筆未解。", ledger_updates=[],
            archived_beats=[], preserved_foreshadowings=["地下室名單"], final_usage_estimate=0.3)
        self.story_tokens = story_tokens or ["你站在走廊。", "<<<DECISION>>>", STORY_JSON]
        self.story_contexts = []

    def call(self, agent, context, output_model=None, temperature=None):
        if agent == "setup":
            return self.setup_out
        if agent == "warden":
            return self.warden_out
        if agent == "compactor":
            return self.compactor_out
        raise AssertionError(f"未預期的 call agent={agent}")

    def stream(self, agent, context, temperature=None):
        assert agent == "story"
        self.story_contexts.append(context)
        for t in self.story_tokens:
            yield t


def _loop(caller, bus=None, run_id="run-test"):
    # 這些是 legacy 流程的整合測試（kernel 流程另見 test_progress_integration.py）
    return BeatLoop(caller, Blackboard(), Database(), signal_bus=bus, run_id=run_id,
                    use_kernel=False)


def test_start_creates_world_and_first_beat():
    loop = _loop(FakeCaller(make_setup_output()))
    res = loop.start({"theme": "廢棄醫院", "npc_count": 1})
    assert res["opening_sequence"] and "大廳醒來" in res["opening_sequence"][0]
    assert res["decision_point"].suggested_options[0].text == "開門"
    snap = loop.bb.snapshot()
    assert snap["real_bible"] and snap["npc_registry"][0]["name"] == "張醫生"
    assert loop.db.list_runs()[0]["run_id"] == "run-test"


def test_step_advances_and_persists():
    loop = _loop(FakeCaller(make_setup_output()))
    loop.start({"theme": "x", "npc_count": 1})
    b0 = loop.beat_number
    out = loop.step("我開門")
    assert loop.beat_number == b0 + 1
    assert out["narrative"] and out["decision_point"].decision_type in ("action", "dialogue")
    saved = loop.db.load_beat("run-test", loop.beat_number)
    assert saved is not None and saved["narrative"]


def test_continuous_beats_window_bounded_and_compaction():
    caller = FakeCaller(make_setup_output())
    loop = _loop(caller)
    loop.start({"theme": "x", "npc_count": 1})
    for _ in range(9):
        loop.step("繼續探索")
    # beat_window 受控（修剪到上限）
    assert len(loop.bb.beat_window) <= BEAT_WINDOW_SIZE
    # compaction 觸發 → patch 經 merge_and_bump 套用 → version 前進
    assert loop.bb.version >= 1
    # 伏筆保護清單仍在
    assert "地下室名單" in loop.compactor.protected_foreshadowings
    # 連續 10 beat 皆落 SQLite
    assert loop.db.load_beat("run-test", 10) is not None


def test_warden_hard_rule_ends_game():
    caller = FakeCaller(make_setup_output(hard_triggers=["喝藥水"]))
    bus = SignalBus()
    ended_events = []
    bus.subscribe(EVT_ENDING_TRIGGERED, lambda *a, **k: ended_events.append(1))
    loop = _loop(caller, bus=bus)
    loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我決定喝藥水")
    assert out["ended"] is True
    assert out["warden"].rule_violation is True
    assert ended_events  # ENDING_TRIGGERED 發出
    # 結束後再 step 不再推進
    assert loop.step("還想動作")["ended"] is True


def test_beat_completed_signal_fires_each_beat():
    bus = SignalBus()
    hits = []
    bus.subscribe(EVT_BEAT_COMPLETED, lambda *a, **k: hits.append(1))
    loop = _loop(FakeCaller(make_setup_output()), bus=bus)
    loop.start({"theme": "x", "npc_count": 1})   # 1
    loop.step("a"); loop.step("b")               # +2
    assert len(hits) == 3


def test_story_context_never_contains_real_bible():
    forbidden = "這是絕不能洩漏的真相ZZZ"
    caller = FakeCaller(make_setup_output(forbidden=forbidden))
    loop = _loop(caller)
    loop.start({"theme": "x", "npc_count": 1})
    loop.step("我四處查看")
    blob = json.dumps(caller.story_contexts, ensure_ascii=False)
    assert forbidden not in blob          # real_bible 內容
    assert "他就是兇手SECRET" not in blob   # secret_core
