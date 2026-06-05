"""階段 R 整合驗收（敘事控制 v0.2）——透過真實 BeatLoop 驗證「調查 → 真相進度」鏈接通。

頭條驗收（patch docs/10 + README）：玩家發現有意義線索後，揭露帳本前進、結局 recap **不再 0/X**；
全程 flag OFF 時行為與現況一致。
"""
import core.constants as C
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus
from core.models import SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput

STORY = ('{"situation_recap":"前方走廊。","decision_type":"action",'
         '"suggested_options":[{"text":"往前走","tone":"bold"}],"beat_meta":{"beat_number":0}}')


class FakeCaller:
    def __init__(self):
        self.setup_out = SetupOutput(
            real_bible={"world_truth": {"what_really_happened": "這裡曾做過記憶實驗",
                                        "the_threat_is": "頻率", "deadly_rule": "別相信整點報時"},
                        "hard_triggers": [], "ending_conditions": [],
                        "revelation_pool": [
                            {"id": "f1", "title": "第一道警示", "content": "牆上刻著別相信整點報時"},
                            {"id": "f2", "title": "被忽略的細節", "content": "名單上有林晨的名字"},
                            {"id": "f3", "title": "核心真相", "content": "你也是實驗體之一"}]},
            npc_registry=[NPCBible(name="醫生", profession="醫生", personality="mysterious",
                                   voice_sample="…", public_face="冷", secret_core="兇手SECRET",
                                   self_aware=True, appearance="")],
            protagonist={"name": "林默", "starting_situation": "弟弟林晨在這裡失蹤"},
            scene_registry=SceneRegistry(current_location="hall",
                                         known_locations=[Location(id="hall", name="大廳", description="")]),
            opening_sequence=["你在病房醒來。"])

    def call(self, agent, context, output_model=None, temperature=None):
        if agent == "setup":
            return self.setup_out
        if agent == "warden":
            return WardenOutput(directive_to_story="繼續")
        raise AssertionError(agent)

    def stream(self, agent, context, temperature=None):
        for t in ["走廊很暗。", "<<<DECISION>>>", STORY]:
            yield t


def _loop():
    return BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(),
                    run_id="r", use_kernel=True)


def test_investigation_advances_reveal_ledger(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    loop = _loop()
    loop.start({"theme": "廢棄醫院", "npc_count": 1})
    # 帳本已種子化（3 個碎片，全 hidden）
    assert loop._reveal_ledger is not None
    assert loop._reveal_ledger.counts()["total"] == 3
    assert loop._reveal_ledger.counts()["hinted_or_better"] == 0

    # 玩家**真相調查**數個 beat（truth_investigation 才推 reveal；kernel 發現 clue.core 等線索）
    for txt in ["我打開病房門，研究裡面的病歷紀錄",
                "我仔細檢查走廊，分析牆上的異常痕跡",
                "我往更深處走，研判沿途的監控紀錄",
                "我搜查這個房間，解讀實驗紀錄與異常數據"]:
        loop.step(txt)

    # 頭條：揭露帳本前進（至少 1 條真相到 hinted+）
    led = loop._reveal_ledger.counts()
    assert led["hinted_or_better"] >= 1, "調查後揭露帳本仍為 0——斷鏈未修好"
    # 要求 #3：帳本寫進 revealed_bible（含分層），不是旁路 game_meta
    rb = loop.bb.snapshot().get("revealed_bible") or {}
    tp = rb.get("truth_progress") or {}
    assert any(v.get("level") != "hidden" for v in tp.values()), "revealed_bible 未記錄揭露進度"
    # 要求 #6：深入調查到 clue.core → 該真相 confirmed 進 revealed_fragments
    assert led["confirmed_or_better"] >= 1, "決定性線索未推到 confirmed"
    assert rb.get("revealed_fragments"), "confirmed 真相未進 revealed_bible.revealed_fragments"


def test_recap_not_zero_after_investigation(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    for txt in ["我研究牆上的刻字", "我分析監控紀錄", "我解讀異常頻率數據", "我研判實驗紀錄"]:
        loop.step(txt)
    # 強制結局，組裝復盤
    loop.ended = True
    loop.ending = {"type": "escape", "escape_quality": "ambiguous"}
    loop._finalize_ending()
    recap = loop.ending["recap"]
    partial = recap.get("partial") or {}
    assert partial.get("found", 0) >= 1                     # ← 不再 0/X
    # ambiguous 逃脫渲染與 clean 不同
    assert loop.ending["ending_surface"] == "ambiguous_escape"


def test_flag_off_unchanged(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", False)
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    loop.step("開門")
    # flag OFF：不建帳本、不標表層變體
    assert loop._reveal_ledger is None
    loop.ended = True
    loop.ending = {"type": "escape"}
    loop._finalize_ending()
    assert "partial" not in (loop.ending["recap"] or {})
    assert loop.ending.get("ending_surface") in (None,)
