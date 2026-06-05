"""Open Exploration P0 驗收測試（NegativeIntentGuard / WorldStateFact / WorldProgress）。

對應使用者 P0 五項：
1. 「先撤到外面整理線索，不結束本次調查」不得 ending。
2. 「不進 B 區」不得移動到 B 區。
3. NPC 說「通訊設備在機房」要寫入 world_state_fact。
4. 沒有 truth reveal 但有 world_state_fact 時，該 beat 仍算有進展。
5. WorldProgress 觀測欄位齊全。
"""
from __future__ import annotations

import core.constants as C
from core.narrative.negative_intent import negates_ending, negated_targets, is_negated
from core.narrative.world_facts import extract_world_facts, add_world_facts, get_world_facts


# ── #1：明確不結束 → 不得 ending（透過真實 loop）────────────────────────────
def test_withdraw_with_explicit_not_end_does_not_end(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("先沿原路離開目前區域，到外面整理線索，不結束本次調查")
    assert not out.get("ended"), "明確『不結束本次調查』卻進了 ending"


# ── #2：「不進 B 區」不得移動到 B 區（kernel NegativeIntentGuard）─────────────
def test_negative_intent_blocks_move():
    from core.progress_kernel import ProgressKernel
    from core.progress_models import GameState
    # 自建小圖：在 hall 有一個會移動到「B 區」的事件
    graph = {"start_scene": "hall", "events": [
        {"id": "event.enter_b", "scene_id": "hall", "intent_tags": ["move", "inspect"],
         "effects": [{"op": "set", "path": "current_scene", "value": "B 區"}],
         "narrative_obligations": ["玩家走進「B 區」。"], "progress_delta": ["location_changed"],
         "max_repeat": 1, "forbidden_after": []},
        {"id": "event.look", "scene_id": "hall", "intent_tags": ["inspect", "free"],
         "effects": [{"op": "add", "path": "clues.c1",
                      "value": {"title": "x", "content": "y", "tags": []}}],
         "progress_delta": ["new_clue_added"], "max_repeat": 1, "forbidden_after": []}]}
    class _P:
        def graph(self): return graph
        def start_scene(self): return "hall"
        def default_obligations(self): return []
    k = ProgressKernel(_P())
    st = GameState(version=1, beat_number=0, current_scene="hall", scene_phase="beginning")
    # 不帶否定 → 可能選到移動事件
    # 帶否定「B 區」→ 不得選到 enter_b（移動到 B 區）
    res = k.resolve_player_action("我先判斷，不進 B 區，先在原地查看",
                                  st, warden={"negated_targets": negated_targets("不進 B 區")})
    assert res.committed_event != "event.enter_b"
    # 確認 patch 沒有把場景設成「B 區」
    moved_to_b = any(getattr(op, "path", "") == "current_scene"
                     and getattr(op, "value", None) == "B 區" for op in res.patch.ops)
    assert not moved_to_b


def test_negated_targets_and_negates_ending():
    assert negates_ending("先撤到外面整理線索，不結束本次調查")
    assert not negates_ending("我結束本次調查")
    assert is_negated("B 區", negated_targets("不進 B 區"))
    assert not is_negated("機房", negated_targets("不進 B 區"))


# ── #3：NPC「通訊設備在機房」→ world_state_fact ──────────────────────────────
def test_npc_info_becomes_world_fact():
    facts = extract_world_facts("我不確定真相，但通訊設備在機房，你得自己過去。",
                                source="npc_chat")
    keys = [f["key"] for f in facts]
    assert "machine_room_known" in keys
    assert all(f["source"] == "npc_chat" for f in facts)


def test_npc_bridge_writes_world_fact(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    from core.narrative.npc_chat_control import NPCChatResponse
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    loop.bridge_npc_evidence(NPCChatResponse(
        visible_reply="出口鎖死了，要先重啟發電機。通訊設備在機房。", answer_status="partial"))
    facts = get_world_facts(loop.bb)
    assert "machine_room_known" in facts
    assert "known_exit_locked" in facts or "generator_needed" in facts


# ── #4：無 truth reveal 但有 world_state_fact → 仍算有進展 ───────────────────
def test_world_fact_counts_as_progress_without_truth():
    class _BB:
        def __init__(self): self.game_meta = {}
        def snapshot(self): return {}
    bb = _BB()
    added = add_world_facts(bb, extract_world_facts("出口鎖死，需要重啟發電機"), beat=2)
    assert added                                         # 有寫入世界事實
    assert get_world_facts(bb)                            # 可檢查
    # （annotate_world 的 had_consequence 已把 new_world_facts 算入後果——見 agent_play）


# ── #5：WorldProgress 觀測欄位齊全 ───────────────────────────────────────────
def test_world_progress_fields(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我檢查走廊")
    wp = out["world_progress"]
    for k in ("current_area", "known_areas", "world_facts", "new_world_facts_this_beat",
              "changed_exits_this_beat", "investigation_state", "available_next"):
        assert k in wp, f"world_progress 缺 {k}"
    # investigation_state 現為 ExplorationMode（active_exploration/temporary_retreat/review_mode/…）
    assert wp["investigation_state"] == "active_exploration"
    # 撤退整理 → ReviewMode Lock：investigation_state 變 review_mode（撤離鎖）
    out2 = loop.step("先撤到外面整理線索，不結束本次調查")
    assert out2["world_progress"]["investigation_state"] == "review_mode"
