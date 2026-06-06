"""Player State Surface（Step 5；observation-only）—— inventory / known_facts / focus / recent /
changed reason / deterministic summary。不改 WorldModel 權威、不推 reveal、不新增 fact。"""
from __future__ import annotations

import core.constants as C
from core.world.model import WorldModel, OBJECT, FACT
from core.world.player_state import (
    build_player_state, project_inventory, project_known_facts, player_state_summary,
    PLAYER_STATE_SUMMARY_SOURCE,
)
from core.world.spatial import build_spatial_projection


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def _levels(loop):
    return {tid: t.level for tid, t in (getattr(loop._reveal_ledger, "truths", {}) or {}).items()}


# ── P0 inventory_entities ─────────────────────────────────────────────────────
def test_inventory_entities_shows_taken_objects():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(OBJECT, "WU袖扣", id="object.cuff", state="taken", props={"area": "area.site"})
    m.register(OBJECT, "鑰匙", id="object.key", props={"carried": True, "area": "area.site"})
    m.register(OBJECT, "桌上的紙", id="object.paper", props={"area": "area.site"})  # 沒拿 → 不在 inventory
    inv = project_inventory(m)
    ids = {e["id"] for e in inv}
    assert "object.cuff" in ids and "object.key" in ids and "object.paper" not in ids
    assert all(e["carried"] for e in inv)
    # taken object **不**再被當作地面 visible entity（進 inventory，不在 spatial visible）
    p = build_spatial_projection(m)
    assert "object.cuff" not in [e.id for e in p.visible_entities]
    assert "object.paper" in [e.id for e in p.visible_entities]


# ── P1 known_facts structured ────────────────────────────────────────────────
def test_known_facts_structured_projection():
    m = WorldModel()
    m.register(FACT, "通訊設備在機房", id="fact.comm",
               props={"source": "npc.陳世和", "confidence": "npc_claim", "tags": ["location_claim"]})
    kf = project_known_facts(m)
    assert len(kf) == 1
    f = kf[0]
    assert f["label"] == "通訊設備在機房"                 # 保留自然語意，非粗 key
    assert f["source"] == "npc.陳世和" and f["confidence"] == "npc_claim"
    assert "location_claim" in f["tags"]


def test_npc_fact_in_known_facts_does_not_reveal(monkeypatch):
    from core.narrative.npc_chat_control import NPCChatResponse
    loop = _started_loop(monkeypatch)
    before = _levels(loop)
    loop.bridge_npc_evidence(
        NPCChatResponse(visible_reply="通訊設備在B2機房。", entity_delta=[]), npc_id="npc.王哲")
    ps = loop.player_state()
    labels = [f["label"] for f in ps["known_facts"]]
    assert "通訊設備在B2機房" in labels                    # NPC prose fact 出現在 known_facts
    assert _levels(loop) == before                        # 不推 reveal


def test_hidden_truth_not_in_known_facts(monkeypatch):
    loop = _started_loop(monkeypatch)
    ps = loop.player_state()
    # known_facts 只投影 WorldModel fact 實體（npc_claim/story），不含 real_bible 的 secret_core 原文
    blob = str(ps["known_facts"])
    real = loop.bb.snapshot().get("real_bible") or {}
    for f in (real.get("revelation_pool") or []):
        content = (f.get("content") or "")
        if content:
            assert content not in blob


# ── P2 current_focus / recent_entities ───────────────────────────────────────
def test_current_focus_updates_on_object_inspection(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop._world_model_tick("我四處看", "桌上撿起一本筆記本。", None)   # 登記
    loop._world_model_tick("我檢查那本筆記本", "你翻開筆記本。", None)  # 檢查 → focus
    foc = loop._current_focus
    assert foc and "筆記本" in foc["label"] and foc["reason"] == "inspected"
    assert any("筆記本" in r["label"] for r in loop._recent_entities)


def test_current_focus_updates_on_npc_chat(monkeypatch):
    loop = _started_loop(monkeypatch)
    name = (loop.bb.snapshot().get("npc_registry") or [{}])[0].get("name")
    loop.note_focus_npc(name)
    foc = loop._current_focus
    assert foc and foc["label"] == name and foc["kind"] == "actor" and foc["reason"] == "talked"


def test_recent_entities_bounded(monkeypatch):
    loop = _started_loop(monkeypatch)
    for i in range(12):
        loop._set_focus(f"object.{i}", "inspected", label=f"物{i}", kind="object")
    assert len(loop._recent_entities) <= 8                # recent 上限
    assert loop._recent_entities[0]["id"] == "object.11"  # 最新在前


# ── P3 changed_entities reason ───────────────────────────────────────────────
def test_changed_entities_include_reason(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("我四處張望")
    ch = out["world_progress"]["changed_entities_this_beat"]
    # 變更明細為 dict（含 reason）；area_changed 應出現（kernel 移動）
    assert all(isinstance(c, dict) and "reason" in c for c in ch) or ch == []
    # 直接驗 _build_changed_detail 的 reason 對映
    before = {"area.a": "known", "object.x": "noticed"}
    loop._world.register("area", "A", id="area.a")
    loop._world.register(OBJECT, "X", id="object.x")
    loop._world.set_state("object.x", "inspected")
    loop._world._set_current("area.a")
    detail = loop._build_changed_detail(before, {e.id: e.state for e in loop._world.entities.values()})
    reasons = {d["id"]: d["reason"] for d in detail}
    assert reasons.get("object.x") == "inspected"
    assert reasons.get("area.a") == "area_changed"


# ── P4 player_state_summary deterministic ────────────────────────────────────
def test_player_state_summary_is_deterministic():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(OBJECT, "WU袖扣", id="object.cuff", state="taken", props={"area": "area.site"})
    m.register(FACT, "通訊設備在機房", id="fact.comm", props={"confidence": "npc_claim"})
    ps = build_player_state(m, current_focus={"id": "object.cuff", "label": "WU袖扣", "kind": "object"},
                            recent_entities=[{"id": "object.cuff", "label": "WU袖扣"}])
    t1, tr1 = player_state_summary(ps)
    t2, tr2 = player_state_summary(ps)
    assert t1 == t2                                       # 確定性（同輸入同輸出）
    assert "你目前攜帶：WU袖扣" in t1
    assert "通訊設備在機房" in t1
    assert "目前焦點：WU袖扣" in t1
    # 字數上限 + 截斷旗標
    big = build_player_state(m)
    big["known_facts"] = [{"label": "很長的主張" * 5} for _ in range(40)]
    _, tr = player_state_summary(big, max_chars=80)
    assert tr is True


# ── review_mode：玩家狀態可見，但不新增 fact ─────────────────────────────────
def test_review_mode_player_state_visible_but_no_new_fact(monkeypatch):
    loop = _started_loop(monkeypatch)
    # 先在 active 拿一個物件 + 一條 fact
    loop._world.register(OBJECT, "鑰匙", id="object.key", state="taken")
    facts_before = {e.id for e in loop._world.by_kind(FACT)}
    out = loop.step("先退到外面整理線索，不結束本次調查，只整理已知")
    ps = out["player_state"]
    # review 仍能看到 inventory（隨身物即使在安全區也顯示）
    assert any(e["id"] == "object.key" for e in ps["inventory_entities"])
    # review 不新增 fact
    assert {e.id for e in loop._world.by_kind(FACT)} == facts_before
    assert out["player_state_summary_source"] == PLAYER_STATE_SUMMARY_SOURCE
