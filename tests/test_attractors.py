"""SK13 attractor-based ending 測試。"""
from core.attractors import evaluate, dominant_ending
from core.progress_models import GameState, LedgerEntry
from core.constants import DANGER_DEATH_THRESHOLD

from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus
from tests.test_progress_integration import FakeCaller


def _state(**kw):
    base = dict(version=1, beat_number=0, current_scene="s", scene_phase="beginning")
    base.update(kw)
    return GameState(**base)


def _truth_clue(i):
    return LedgerEntry(id=f"t{i}", title="真相", content="", source_event="e",
                       first_seen_beat=0, tags=["truth"])


def test_no_ending_below_thresholds():
    assert dominant_ending(_state(danger_level=2)) is None


def test_danger_accumulation_triggers_death_attractor():
    st = _state(danger_level=DANGER_DEATH_THRESHOLD)
    assert dominant_ending(st) == "death_physical"        # 由累積危險觸發，非腳本節點


def test_truth_fragments_trigger_truth_attractor():
    st = _state(clues={"t1": _truth_clue(1), "t2": _truth_clue(2)})
    assert dominant_ending(st) == "truth_revealed"


def test_many_clues_survived_trigger_escape():
    clues = {f"c{i}": LedgerEntry(id=f"c{i}", title="x", content="", source_event="e",
                                  first_seen_beat=0, tags=(["truth"] if i == 0 else []))
             for i in range(5)}
    st = _state(danger_level=1, clues=clues)
    pulls = evaluate(st)
    assert pulls["escape"] >= 1.0
    assert dominant_ending(st) == "escape"


def test_loop_ends_via_attractor_when_danger_accumulates(monkeypatch):
    # attractor 自動收束是 **flag OFF** 的舊行為；flag ON 改由 Player Sovereignty（玩家明確結束）。
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", False)
    loop = BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(),
                    run_id="att", use_kernel=True)
    loop.start({"theme": "x", "npc_count": 1})
    ended_via = None
    for _ in range(40):
        out = loop.step("我胡亂衝撞、激怒四周")   # 反覆觸發升壓/fallback，累積 danger
        if out.get("ended"):
            ended_via = out["ending"].get("via")
            break
    assert ended_via == "attractor"               # 結局由累積狀態吸引，非固定終點
    assert loop.ending["type"] in {"death_physical", "truth_revealed", "escape"}
