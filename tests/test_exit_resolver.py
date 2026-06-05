"""ExitResolver / Player Sovereignty 驗收測試。

核心：玩家說「離開」預設不直接 ending；不確定 → ExitOffer（四選一，含「結束」）；
只有明確「結束本次調查」才進 EndingGate。
"""
from __future__ import annotations

import core.constants as C
from core.narrative.exit_resolver import (
    resolve_exit_intent, exit_offer_options, build_exit_offer_decision_point,
    RUN_ENDING, AREA_TRANSITION, TEMPORARY_RETREAT, RETURN_TO_MOTIVE, AMBIGUOUS, NONE,
)
from core.models import DecisionPoint, Option


# ── 意圖分類 ─────────────────────────────────────────────────────────────────
def test_intent_classification():
    assert resolve_exit_intent("我結束本次調查，接受目前結果") == RUN_ENDING
    assert resolve_exit_intent("我下定決心，頭也不回地走出去") == RUN_ENDING
    assert resolve_exit_intent("我放棄調查，直接離開") == RUN_ENDING
    assert resolve_exit_intent("我離開這個房間") == AREA_TRANSITION
    assert resolve_exit_intent("先暫時撤退到外面整理線索") == TEMPORARY_RETREAT
    assert resolve_exit_intent("回頭繼續尋找林晨") == RETURN_TO_MOTIVE
    assert resolve_exit_intent("我試圖離開這個地方，找出口") == AMBIGUOUS   # 語意不明 → 問
    assert resolve_exit_intent("我檢查桌上的紙條") == NONE


# ── ExitOffer 四個選項（含「結束」）、labels 能被分類回對應意圖 ──────────────
def test_exit_offer_has_four_options():
    opts = exit_offer_options()
    assert len(opts) == 4
    ids = {o["id"] for o in opts}
    assert ids == {"leave_area", "temporary_retreat", "end_run", "return_to_motive"}
    by = {o["id"]: o["label"] for o in opts}
    assert resolve_exit_intent(by["end_run"]) == RUN_ENDING            # 「結束本次調查…」→ run_ending
    assert resolve_exit_intent(by["leave_area"]) == AREA_TRANSITION
    assert resolve_exit_intent(by["temporary_retreat"]) == TEMPORARY_RETREAT
    assert resolve_exit_intent(by["return_to_motive"]) == RETURN_TO_MOTIVE


def test_build_exit_offer_dp_keeps_narrative_replaces_options():
    base = DecisionPoint(situation_recap="你站在門邊。", decision_type="action",
                         suggested_options=[Option(text="原本選項", tone="bold")],
                         free_input_hint="", beat_meta={"beat_number": 5})
    dp = build_exit_offer_decision_point(base)
    labels = [o.text for o in dp.suggested_options]
    assert len(labels) == 4 and "結束本次調查，接受目前結果" in labels
    assert "你站在門邊。" in dp.situation_recap                        # 保留原敘事


# ── 玩家說「離開房間」不得 ending ───────────────────────────────────────────
def test_leave_room_does_not_end(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我離開這個房間，到隔壁看看")
    assert not out.get("ended")                                       # area_transition → 續行


# ── 玩家說「我結束調查，接受目前結果」才 ending ──────────────────────────────
def test_explicit_end_run_ends(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我結束本次調查，接受目前結果")
    assert out.get("ended")
    assert out["ending"]["type"] == "escape"
    assert out["ending"]["via"] == "player_exit"


# ── ambiguous → ExitOffer（不 ending，dp 給四個選項）─────────────────────────
def test_ambiguous_exit_offers_four_options(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我試圖離開這個地方，尋找出口")
    assert not out.get("ended")                                       # 不自動收束
    labels = [o.text for o in out["decision_point"].suggested_options]
    assert len(labels) == 4
    assert any("結束本次調查" in l for l in labels)                    # 永遠有「結束」出口
    # 接著選「結束」→ 才真的 ending
    out2 = loop.step("結束本次調查，接受目前結果")
    assert out2.get("ended")


# ── danger 累積 + flag ON 不自動致死（吸引子不收束）──────────────────────────
def test_danger_does_not_auto_end_when_flag_on(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    loop._game_state.danger_level = 9999
    out = loop.step("我站在原地發抖")
    assert not out.get("ended")                                       # 危險拉力不自動結局
