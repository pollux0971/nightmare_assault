"""NPC structured entity_delta —— NPC-chat 有用資訊落成 WorldModel fact/actor entity。

嚴格邊界：NPC 只准 fact/actor，不得新增 object/area/exit；fact 帶 source/confidence/origin；
malformed 不污染；NPC fact 寫進 WorldModel **不**自動 grant reveal（不變 confirmed truth）。
"""
from __future__ import annotations

import core.constants as C
from core.world.model import (
    WorldModel, coerce_npc_entity_deltas, NPC_ENTITY_KINDS, NPC_FACT_CONFIDENCE,
    FACT, ACTOR, OBJECT, AREA, EXIT,
)
from core.narrative.npc_chat_control import (
    NPCChatResponse, NPCChatControlContext, NPCChatControlGate,
)


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


# ── coerce：只准 fact/actor；object/area/exit 一律丟棄 ────────────────────────
def test_coerce_npc_only_fact_actor():
    raw = [
        {"op": "register", "kind": "fact", "label": "通訊設備在機房"},
        {"op": "register", "kind": "actor", "label": "王哲"},
        {"op": "register", "kind": "object", "label": "鑰匙"},       # NPC 不准 object
        {"op": "register", "kind": "area", "label": "機房"},          # NPC 不准 area
        {"op": "register", "kind": "exit", "label": "後門"},          # NPC 不准 exit
    ]
    out = coerce_npc_entity_deltas(raw, npc_id="npc.王哲")
    kinds = sorted({d.kind for d in out})
    assert kinds == [ACTOR, FACT]                       # 只剩 fact/actor
    assert all(d.kind in NPC_ENTITY_KINDS for d in out)


# ── fact entity 帶 source=npc id / confidence=npc_claim / origin=npc ──────────
def test_coerce_npc_fact_carries_source_confidence_origin():
    out = coerce_npc_entity_deltas(
        [{"op": "register", "kind": "fact", "label": "通訊設備在機房"}], npc_id="npc.王哲")
    assert len(out) == 1
    d = out[0]
    assert d.origin == "npc"
    assert d.props.get("source") == "npc.王哲"
    assert d.props.get("confidence") == NPC_FACT_CONFIDENCE == "npc_claim"


# ── malformed / 非 list → 不產生垃圾 delta ───────────────────────────────────
def test_coerce_npc_malformed_dropped():
    assert coerce_npc_entity_deltas(None, npc_id="x") == []
    assert coerce_npc_entity_deltas("not-a-list", npc_id="x") == []
    bad = ["junk", {"op": "drop"}, {"op": "register", "kind": "fact", "label": "   "},
           {"op": "register", "label": "無 kind"}, {"kind": "fact", "label": "無 op"}]
    assert coerce_npc_entity_deltas(bad, npc_id="x") == []


# ── cap：每輪最多 3 筆 ───────────────────────────────────────────────────────
def test_coerce_npc_caps_at_three():
    raw = [{"op": "register", "kind": "fact", "label": f"事實{i}"} for i in range(6)]
    assert len(coerce_npc_entity_deltas(raw, npc_id="x")) == 3


# ── gate：NPC 企圖登記 object/area/exit → 違規 flag ──────────────────────────
def test_gate_flags_illegal_entity_kind():
    gate = NPCChatControlGate()
    ctx = NPCChatControlContext()
    ok = NPCChatResponse(visible_reply="通訊設備在機房。",
                         entity_delta=[{"op": "register", "kind": "fact", "label": "通訊設備在機房"}])
    assert gate.validate(ctx, ok) == []                 # fact 合法 → 無 flag
    bad = NPCChatResponse(visible_reply="這裡有條密道。",
                          entity_delta=[{"op": "register", "kind": "exit", "label": "密道"},
                                        {"op": "register", "kind": "area", "label": "地下室"}])
    flags = gate.validate(ctx, bad)
    assert any(f.startswith("illegal_entity_kind:") for f in flags)
    f = next(f for f in flags if f.startswith("illegal_entity_kind:"))
    assert "area" in f and "exit" in f


# ── parser：JSON 回覆含 entity_delta → NPCChatResponse 帶上；非 list → [] ──────
def test_parse_npc_response_entity_delta():
    from core.agents.npc_chat import _parse_npc_response
    import json
    resp = _parse_npc_response(json.dumps({
        "reply": "通訊設備在機房。", "answer_status": "partial",
        "entity_delta": [{"op": "register", "kind": "fact", "label": "通訊設備在機房"}]}))
    assert resp.entity_delta and resp.entity_delta[0]["kind"] == "fact"
    # 壞 entity_delta（非 list）→ [] 不讓解析失敗
    resp2 = _parse_npc_response(json.dumps({"reply": "嗯。", "entity_delta": "garbage"}))
    assert resp2.entity_delta == [] and resp2.visible_reply == "嗯。"


# ══ loop 整合 ═════════════════════════════════════════════════════════════════
# ── NPC 說「通訊設備在機房」→ 登記 fact entity（帶 source/confidence）───────────
def test_npc_fact_registered_in_world_model(monkeypatch):
    loop = _started_loop(monkeypatch)
    resp = NPCChatResponse(
        visible_reply="通訊設備在機房，但門被鎖住了。", answer_status="partial",
        entity_delta=[{"op": "register", "kind": "fact", "label": "通訊設備在機房"}])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    fact = loop._world.find("通訊設備在機房", kind=FACT)
    assert fact is not None
    assert fact.origin == "npc"
    assert fact.props.get("source") == "npc.王哲"
    assert fact.props.get("confidence") == "npc_claim"
    # observation.world_progress 顯示 new fact entity
    wp = loop.world_progress()
    ids = [e["id"] for e in wp["world_model"]["entities_here"]]
    assert fact.id in ids
    assert fact.id in wp["changed_entities_this_beat"]


# ── NPC 嘗試新增 area/exit → 被拒，WorldModel 不出現該實體 ────────────────────
def test_npc_cannot_add_area_or_exit_via_loop(monkeypatch):
    loop = _started_loop(monkeypatch)
    n_areas = len(loop._world.by_kind(AREA))
    resp = NPCChatResponse(
        visible_reply="走那邊有條密道。",
        entity_delta=[{"op": "register", "kind": "exit", "label": "密道"},
                      {"op": "register", "kind": "area", "label": "地下室"},
                      {"op": "register", "kind": "object", "label": "撬棍"}])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    assert loop._world.find("密道", kind=EXIT) is None
    assert loop._world.find("地下室", kind=AREA) is None
    assert loop._world.find("撬棍", kind=OBJECT) is None
    assert len(loop._world.by_kind(AREA)) == n_areas    # NPC 沒新增任何 area


# ── 壞 entity_delta 不污染 WorldModel ────────────────────────────────────────
def test_npc_malformed_entity_delta_does_not_pollute(monkeypatch):
    loop = _started_loop(monkeypatch)
    n0 = len(loop._world.entities)
    resp = NPCChatResponse(visible_reply="呃……",
                           entity_delta=["junk", {"op": "x"}, {"kind": "fact"}, 42])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    assert len(loop._world.entities) == n0              # 沒新增任何垃圾實體
    assert all(e.kind in (FACT, ACTOR, AREA, EXIT, OBJECT)
               for e in loop._world.entities.values())


# ── NPC fact 不會自動 confirmed truth（entity_delta 與 reveal ledger 解耦）─────
def test_npc_fact_does_not_grant_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    led = loop._reveal_ledger
    before = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    resp = NPCChatResponse(
        visible_reply="通訊設備在機房。", answer_status="partial",
        entity_delta=[{"op": "register", "kind": "fact", "label": "通訊設備在機房"}])
    updates = loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    after = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    assert before == after                              # 沒有任何真相等級被推進
    assert not any(lv == "confirmed" for lv in after.values())   # 更別說 confirmed
    assert updates == [] or all("confirmed" not in str(u) for u in updates)
    # fact 仍進了 WorldModel（只是不算真相）
    assert loop._world.find("通訊設備在機房", kind=FACT) is not None
