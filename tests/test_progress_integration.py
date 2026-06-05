"""SK04 Progress Kernel 整合測試（ON/OFF/fallback/防暴雷）。"""
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus
from core.models import SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput

STORY = ('{"situation_recap":"前方走廊。","decision_type":"action",'
         '"suggested_options":[{"text":"往前走","tone":"bold"}],"beat_meta":{"beat_number":0}}')
FORBIDDEN_TRUTH = "這是絕不能洩漏的真相ZZZ"


class FakeCaller:
    def __init__(self):
        self.setup_out = SetupOutput(
            real_bible={"world_truth": {"what_really_happened": FORBIDDEN_TRUTH,
                                        "the_threat_is": "X", "deadly_rule": "不可喝藥水"},
                        "hard_triggers": [], "ending_conditions": [], "revelation_pool": []},
            npc_registry=[NPCBible(name="醫生", profession="醫生", personality="mysterious",
                                   voice_sample="…", public_face="冷", secret_core="兇手SECRET",
                                   self_aware=True, appearance="")],
            protagonist={"name": "林默"},
            scene_registry=SceneRegistry(current_location="hall",
                                         known_locations=[Location(id="hall", name="大廳", description="")]),
            opening_sequence=["你在病房醒來。"])
        self.story_contexts = []

    def call(self, agent, context, output_model=None, temperature=None):
        if agent == "setup":
            return self.setup_out
        if agent == "warden":
            return WardenOutput(directive_to_story="繼續")
        raise AssertionError(agent)

    def stream(self, agent, context, temperature=None):
        self.story_contexts.append(context)
        for t in ["走廊很暗。", "<<<DECISION>>>", STORY]:
            yield t


class BadProvider:
    def graph(self):
        raise RuntimeError("graph load fail")

    def start_scene(self):
        return "x"

    def default_obligations(self):
        return []


def _kernel_loop():
    c = FakeCaller()
    loop = BeatLoop(c, Blackboard(), Database(), SignalBus(), run_id="k", use_kernel=True)
    return loop, c


def test_kernel_starts_in_themed_scene():
    loop, _ = _kernel_loop()
    assert loop.use_kernel is True
    loop.start({"theme": "廢棄醫院", "npc_count": 1})
    # Patch 2：start_scene 來自 setup 的主題場景（FakeCaller 用 "hall"），非固定 ward
    assert loop._game_state.current_scene == "hall"


def test_exit_then_no_repeat():
    loop, _ = _kernel_loop()
    loop.start({"theme": "x", "npc_count": 1})
    out1 = loop.step("我打開病房門")
    assert out1["committed_event"] == "event.exit_start"
    assert loop._game_state.current_scene == "scene.beyond"   # 合成第二場景（只有 1 個 loc）
    assert out1["progress_delta"]                       # T3：≥1 delta
    # T1：下一 beat 不再是同一個離開事件
    out2 = loop.step("我往前走看看")
    assert out2["committed_event"] != "event.exit_start"
    assert "ask_exit_start" in loop._game_state.forbidden_repeats


def test_every_beat_has_progress_delta():
    loop, _ = _kernel_loop()
    loop.start({"theme": "x", "npc_count": 1})
    for txt in ["開門", "觀察走廊", "繼續走", "再走", "發呆"]:
        out = loop.step(txt)
        assert out["progress_delta"], f"beat 無 progress_delta: {txt}"


def test_npc_spawns_within_few_beats():
    loop, _ = _kernel_loop()
    loop.start({"theme": "x", "npc_count": 1})
    loop.step("開門")
    loop.step("我觀察四周")
    # 主題化 npc：來自 setup 的 npc_registry（FakeCaller 用 "醫生"）
    assert "npc.醫生" in loop._game_state.npcs
    assert loop._game_state.npcs["npc.醫生"].visible is True


def test_spoiler_safe_in_kernel_context():
    loop, c = _kernel_loop()
    loop.start({"theme": "x", "npc_count": 1})
    loop.step("我打開病房門")
    import json
    blob = json.dumps(c.story_contexts, ensure_ascii=False)
    assert FORBIDDEN_TRUTH not in blob          # real_bible 內容不入 kernel context
    assert "兇手SECRET" not in blob             # secret_core 不入


def test_graph_load_fail_falls_back_to_legacy():
    c = FakeCaller()
    loop = BeatLoop(c, Blackboard(), Database(), SignalBus(), run_id="bad",
                    scene_graph_provider=BadProvider(), use_kernel=True)
    loop.start({"theme": "x", "npc_count": 1})  # 在 start() 建圖失敗 → 回退 legacy，不崩
    assert loop.use_kernel is False
    out = loop.step("我開門")
    assert out["narrative"]


def test_legacy_flow_still_runs_when_disabled():
    c = FakeCaller()
    loop = BeatLoop(c, Blackboard(), Database(), SignalBus(), run_id="leg", use_kernel=False)
    loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我開門")
    assert out["narrative"] and out["decision_point"]
