"""NPC onboarding — actor profile schema（patch v0.7 P1）。"""
from __future__ import annotations

import core.constants as C
from core.world.actor_profile import (
    ActorProfile, SpeechStyle, profile_from_npc, get_npc_profile, set_npc_profile,
    mark_introduced, build_first_contact_context, UNINTRODUCED, INTRODUCED,
)


def _npc(**kw):
    base = {"name": "陳博翰", "profession": "工程師", "personality": "nervous",
            "voice_sample": "我只懂機械。", "public_face": "想恢復部分電力",
            "secret_core": "其實動過手腳", "self_aware": True, "appearance": "戴口罩"}
    base.update(kw)
    return base


def test_profile_defaults_unintroduced():
    p = ActorProfile()
    assert p.intro_state == UNINTRODUCED
    assert isinstance(p.speech_style, dict) and "emotional_tone" in p.speech_style


# ── NPC actor entity registration consistency ───────────────────────────────

def _bb_with_npc(npc=None):
    from core.blackboard import Blackboard
    bb = Blackboard()
    bb.write("setup", "npc_registry", [npc or _npc()])
    return bb


def test_ensure_creates_actor_entity_public_fields_only():
    from core.world.model import WorldModel, ACTOR
    from core.world.actor_profile import ensure_actor_entity_from_npc_registry
    w = WorldModel()
    e = ensure_actor_entity_from_npc_registry(w, _bb_with_npc(), "陳博翰", origin="npc_chat")
    assert e is not None and e.kind == ACTOR and e.label == "陳博翰"
    assert e.id == "actor.陳博翰"
    assert e.props.get("known_role") == "工程師"             # 表層職業
    assert e.props.get("origin_kind") == "npc_chat"
    assert "戴口罩" in (e.props.get("aliases") or [])         # 外觀 alias
    blob = repr(e.props)
    assert "secret_core" not in blob and "其實動過手腳" not in blob   # hidden truth 永不進來


def test_ensure_is_idempotent_no_duplicate():
    from core.world.model import WorldModel, ACTOR
    from core.world.actor_profile import ensure_actor_entity_from_npc_registry
    w = WorldModel(); bb = _bb_with_npc()
    e1 = ensure_actor_entity_from_npc_registry(w, bb, "陳博翰")
    e2 = ensure_actor_entity_from_npc_registry(w, bb, "陳博翰")
    assert e1.id == e2.id and len(w.by_kind(ACTOR)) == 1


def test_ensure_merges_alias_and_upgrades_intro_only():
    from core.world.model import WorldModel, ACTOR
    from core.world.actor_profile import ensure_actor_entity_from_npc_registry, INTRODUCED
    w = WorldModel()
    e = w.register(ACTOR, "陳博翰", id="actor.陳博翰",
                   props={"intro_state": INTRODUCED, "aliases": ["老陳"]}, origin="npc")
    ensure_actor_entity_from_npc_registry(w, _bb_with_npc(), "陳博翰")  # registry intro=unintroduced
    assert e.props["intro_state"] == INTRODUCED               # 不降級
    assert "老陳" in e.props["aliases"] and "戴口罩" in e.props["aliases"]   # merge，不覆蓋


def test_ensure_adds_no_fact_area_exit():
    from core.world.model import WorldModel, FACT, AREA, EXIT
    from core.world.actor_profile import ensure_actor_entity_from_npc_registry
    w = WorldModel()
    ensure_actor_entity_from_npc_registry(w, _bb_with_npc(), "陳博翰")
    assert len(w.by_kind(FACT)) == 0 and len(w.by_kind(AREA)) == 0 and len(w.by_kind(EXIT)) == 0


# ── loop 整合：note_focus_npc 後 actor entity 存在、「那個人」→ actor、不推 reveal ──

def _loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop as mk
    loop = mk(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def _levels(loop):
    return {tid: t.level for tid, t in (getattr(loop._reveal_ledger, "truths", {}) or {}).items()}


def test_note_focus_npc_ensures_world_actor_entity(monkeypatch):
    from core.world.model import ACTOR
    loop = _loop(monkeypatch)
    npc = next(iter(n.get("name") for n in loop.bb.snapshot().get("npc_registry") or [] if n.get("name")))
    before = _levels(loop)
    loop.note_focus_npc(npc)
    # focus.id 對得到真正的 WorldModel actor entity
    fid = loop._current_focus["id"]
    assert loop._world.get(fid) is not None and loop._world.get(fid).kind == ACTOR
    assert _levels(loop) == before                           # 不推 reveal


def test_that_person_resolves_to_actor_after_focus(monkeypatch):
    from core.world.alias_resolver import resolve_entity_reference
    loop = _loop(monkeypatch)
    npc = next(iter(n.get("name") for n in loop.bb.snapshot().get("npc_registry") or [] if n.get("name")))
    loop.note_focus_npc(npc)
    r = resolve_entity_reference("那個人", world=loop._world, current_focus=loop._current_focus)
    rid = r["resolved_entity_id"]
    assert rid is not None and loop._world.get(rid) is not None
    assert loop._world.get(rid).kind == "actor"              # resolved_kind = actor（不再缺失）


def test_profile_from_npc_public_only():
    p = profile_from_npc(_npc())
    assert p.display_label == "陳博翰" and p.true_name == "陳博翰"
    assert p.known_role == "工程師"                      # 表層職業推測
    assert p.surface_motive == "想恢復部分電力"          # public_face，非 secret_core
    assert "戴口罩" in p.aliases                         # 外觀 alias
    # **player-visible 欄位不得含隱藏真相**
    blob = (p.known_role + p.surface_motive + p.personality_description
            + " ".join(p.aliases) + p.first_seen_context)
    assert "動過手腳" not in blob
    # personality → 語氣（nervous）
    assert p.speech_style["emotional_tone"] == "nervous"
    assert p.personality_description                      # 非空


def test_profile_serialize_round_trip():
    p = profile_from_npc(_npc(personality="mysterious"))
    p2 = ActorProfile.from_dict(p.to_dict())
    assert p2.known_role == p.known_role
    assert p2.speech_style["emotional_tone"] == "evasive"  # mysterious → evasive


def test_alias_maps_label():
    p = ActorProfile(display_label="戴口罩的維修員", aliases=["戴口罩的維修員"])
    p.aliases.append("陳博翰")                            # 之後對到真名
    assert "陳博翰" in p.aliases and "戴口罩的維修員" in p.aliases


# ── game_meta 存取（用真 blackboard）───────────────────────────────────────────
def _loop_bb(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop.bb


def test_get_set_profile_via_game_meta(monkeypatch):
    bb = _loop_bb(monkeypatch)
    name = (bb.snapshot().get("npc_registry") or [{}])[0].get("name")
    p = get_npc_profile(bb, name)
    assert p.intro_state == UNINTRODUCED                  # 預設未介紹
    assert p.display_label == name
    # mark_introduced
    p2 = mark_introduced(bb, name)
    assert p2.intro_state == INTRODUCED
    assert get_npc_profile(bb, name).intro_state == INTRODUCED   # 持久化


def test_first_contact_context_shape():
    p = profile_from_npc(_npc())
    fc = build_first_contact_context(p, "通訊設備在哪？")
    assert fc["is_first_contact"] is True
    assert any("位置" in m or "做什麼" in m for m in fc["must_include"])
    assert any("部分" in m for m in fc["must_include"])
    assert "隱藏真相 / 未揭露 bible 內容" in fc["must_not_include"]
    assert fc["player_question"] == "通訊設備在哪？"
    assert "動過手腳" not in str(fc)                      # 不洩隱藏真相
