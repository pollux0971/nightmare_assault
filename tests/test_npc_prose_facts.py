"""NPC Prose Fact Extraction（fallback）—— NPC 沒吐 structured entity_delta 時，
把散文裡明確、可用、非真相的資訊抽成 WorldModel fact entity。"""
from __future__ import annotations

import core.constants as C
from core.world.model import FACT, OBJECT, AREA, EXIT
from core.narrative.npc_prose_facts import (
    extract_npc_prose_facts, LOCATION_CLAIM, LOCKED_EXIT_CLAIM, ACTION_REQUIRED_CLAIM,
)
from core.narrative.npc_chat_control import NPCChatResponse


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


# ── 抽取器單元 ───────────────────────────────────────────────────────────────
def test_extract_location_claim():
    out = extract_npc_prose_facts("通訊設備在B2機房，但線路早就斷了。", npc_id="npc.王哲")
    assert len(out) == 1
    d = out[0]
    assert d.kind == FACT and d.origin == "npc"
    assert d.label == "通訊設備在B2機房"               # 保留自然語意（非 machine_room_known）
    assert d.props.get("source") == "npc.王哲"
    assert d.props.get("confidence") == "npc_claim"
    assert LOCATION_CLAIM in d.props.get("tags", [])


def test_extract_locked_exit_claim_is_fact_not_exit():
    out = extract_npc_prose_facts("我跟你說，東側閘門鎖死了，別想從那走。", npc_id="npc.王哲")
    assert len(out) == 1 and out[0].kind == FACT       # 是 fact，不是 exit
    assert LOCKED_EXIT_CLAIM in out[0].props.get("tags", [])
    assert "閘門" in out[0].label


def test_extract_action_required_claim():
    out = extract_npc_prose_facts("要先重啟發電機，不然什麼都動不了。", npc_id="npc.王哲")
    assert len(out) == 1 and out[0].kind == FACT
    assert ACTION_REQUIRED_CLAIM in out[0].props.get("tags", [])


def test_atmosphere_yields_no_fact():
    for prose in ("空氣裡瀰漫著腐爛的海腥味，牆上的影子像在扭曲蠕動。",
                  "他的聲音彷彿從很遠的地方傳來，像某種低語。",
                  "我不知道……這裡讓我覺得不對勁，有種說不出的寒意。"):
        assert extract_npc_prose_facts(prose, npc_id="x") == []


def test_cap_at_two():
    prose = ("通訊設備在B2機房；發電機在配電室；東側閘門鎖死了；要先重啟發電機。")
    out = extract_npc_prose_facts(prose, npc_id="x")
    assert len(out) <= 2                               # 每次最多 2 個


def test_empty_or_malformed_input():
    assert extract_npc_prose_facts("", npc_id="x") == []
    assert extract_npc_prose_facts("嗯。對。我不確定。", npc_id="x") == []


# ══ loop 整合 ═════════════════════════════════════════════════════════════════
def test_prose_fallback_registers_fact_via_loop(monkeypatch):
    loop = _started_loop(monkeypatch)
    resp = NPCChatResponse(visible_reply="通訊設備在B2機房，但線路早就斷了。",
                           answer_status="partial", entity_delta=[])   # 無 structured
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    f = loop._world.find("通訊設備在B2機房", kind=FACT)
    assert f is not None and f.origin == "npc"
    assert f.props.get("source") == "npc.王哲" and f.props.get("confidence") == "npc_claim"
    assert f.state == "asserted"


def test_locked_exit_prose_does_not_add_exit_entity(monkeypatch):
    loop = _started_loop(monkeypatch)
    n_exit = len(loop._world.by_kind(EXIT))
    resp = NPCChatResponse(visible_reply="東側閘門鎖死了，走不通。", entity_delta=[])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    assert len(loop._world.by_kind(EXIT)) == n_exit    # 沒新增 exit
    assert any(e.kind == FACT and "閘門" in e.label for e in loop._world.entities.values())


def test_structured_entity_delta_skips_prose_extractor(monkeypatch):
    """structured entity_delta 存在 → 不跑 prose fallback（避免重複登記）。"""
    loop = _started_loop(monkeypatch)
    resp = NPCChatResponse(
        visible_reply="通訊設備在B2機房。",            # 散文裡有可抽的 location claim
        entity_delta=[{"op": "register", "kind": "fact", "label": "我親手關了主閘"}])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    # 只登記 structured 那筆，不額外從散文抽「通訊設備在B2機房」
    assert loop._world.find("我親手關了主閘", kind=FACT) is not None
    assert loop._world.find("通訊設備在B2機房", kind=FACT) is None


def test_prose_fact_does_not_grant_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    led = loop._reveal_ledger
    before = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    resp = NPCChatResponse(visible_reply="要先重啟發電機才能供電。", entity_delta=[])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    after = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    assert before == after                              # prose fact 不推進任何 reveal


def test_prose_fact_does_not_change_current_area(monkeypatch):
    loop = _started_loop(monkeypatch)
    area0 = loop._world.current_area
    resp = NPCChatResponse(visible_reply="通訊設備在B2機房，東側閘門鎖死了。", entity_delta=[])
    loop.bridge_npc_evidence(resp, npc_id="npc.王哲")
    assert loop._world.current_area == area0            # current_area 不變
