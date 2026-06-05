"""SK06 30-beat scripted 回歸 + 指標（kernel-only，無真 LLM，固定輸入）。

量化證明穩定化補丁修好「重複/無推進」：progression≥0.60、repetition≤0.10、npc≥0.70。
"""
from core.scene_graph import (
    StaticOpeningSceneGraphProvider, GeneratedSceneGraphProvider, DEFAULT_GRAPH_PATH,
)
from core.progress_kernel import ProgressKernel
from core.patch_validator import PatchValidator
from core import progress_bridge as bridge

INPUTS = ["我打開病房門", "我觀察走廊", "往走廊深處走", "檢查兩側房間",
          "呼喊有沒有人", "我站著不動", "再往前", "回頭看"]


def run_replay(beats: int = 30, provider=None, exit_event: str = "event.open_door_101") -> dict:
    provider = provider or StaticOpeningSceneGraphProvider(DEFAULT_GRAPH_PATH)
    kernel = ProgressKernel.from_provider(provider)
    validator = PatchValidator()
    gs = bridge.init_game_state(provider)

    committed: list[str] = []
    delta_counts: list[int] = []
    npc_visible_beats = 0
    failures: list[str] = []

    for i in range(beats):
        inp = INPUTS[i % len(INPUTS)]
        res = kernel.resolve_player_action(inp, gs)
        try:
            gs = validator.apply(res.patch, gs)
        except Exception as e:                      # 不該發生（kernel 保證 ≥1 delta）
            failures.append(f"beat {i}: {e}")
            continue
        committed.append(res.committed_event)
        delta_counts.append(len(res.patch.progress_delta))
        if any(n.visible for n in gs.npcs.values()):
            npc_visible_beats += 1

    n = len(committed)
    progression_rate = sum(1 for d in delta_counts if d >= 1) / n
    repeats = sum(1 for j in range(1, n) if committed[j] in committed[max(0, j - 3):j])
    repetition_rate = repeats / n
    npc_presence_rate = npc_visible_beats / n
    return {
        "run_id": "regression_001",
        "beats": n,
        "progression_rate": round(progression_rate, 3),
        "repetition_rate": round(repetition_rate, 3),
        "npc_presence_rate": round(npc_presence_rate, 3),
        "json_parse_success": 1.0,              # kernel-only，無 LLM JSON
        "exit_event_count": committed.count(exit_event),
        "failures": failures,
    }


def _assert_thresholds(r):
    assert not r["failures"]                     # 30 beat 不崩
    assert r["exit_event_count"] == 1            # 離開起始場景只發生一次（核心：不重複）
    assert r["progression_rate"] >= 0.60
    assert r["repetition_rate"] <= 0.10
    assert r["npc_presence_rate"] >= 0.70


def test_30beat_metrics_static_graph():
    r = run_replay(30, exit_event="event.open_door_101")
    print("\nstatic regression:", r)
    _assert_thresholds(r)


def test_30beat_metrics_themed_graph():
    snap = {
        "scene_registry": {"current_location": "scene.deck",
                           "known_locations": [{"id": "scene.deck", "name": "甲板"},
                                               {"id": "scene.lab", "name": "研究艙"}]},
        "npc_registry": [{"name": "陳博士"}],
        "real_bible": {"world_truth": {"the_threat_is": "低語"}},
    }
    r = run_replay(30, provider=GeneratedSceneGraphProvider(snap), exit_event="event.exit_start")
    print("\nthemed regression:", r)
    _assert_thresholds(r)
