"""Entity Alias / Focus Resolver（Step 6）—— 自然指代 → 既有 entity；唯讀、不建 entity、不推 reveal。"""
from __future__ import annotations

import core.constants as C
from core.world.model import WorldModel, OBJECT, FACT, ACTOR, AREA
from core.world.alias_resolver import (
    normalize_label, add_alias, entity_alias_set, extract_reference,
    resolve_entity_reference,
)


def _world_with(objs=(), facts=(), actors=()):
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    for oid, label, *st in objs:
        m.register(OBJECT, label, id=oid, state=(st[0] if st else "noticed"), props={"area": "area.site"})
    for fid, label, src in facts:
        m.register(FACT, label, id=fid, props={"source": src, "confidence": "npc_claim"})
    for aid, label in actors:
        m.register(ACTOR, label, id=aid, props={"area": "area.site"})
    return m


# ── 1. label normalization ───────────────────────────────────────────────────
def test_label_alias_normalization():
    assert normalize_label("WU 袖扣") == normalize_label("WU袖扣")
    assert normalize_label("Ｗ Ｕ　袖扣，") == "wu袖扣"          # 全形 + 標點 + 空白
    m = _world_with(objs=[("object.WU袖扣", "WU袖扣")])
    # 「WU 袖扣」（有空白）解析到同一 entity，不新增
    r = resolve_entity_reference("WU 袖扣", world=m)
    assert r["resolved_entity_id"] == "object.WU袖扣" and not r["ambiguous"]
    assert len(m.by_kind(OBJECT)) == 1                         # 沒新增重複 entity


# ── 2. aliases ───────────────────────────────────────────────────────────────
def test_aliases_no_pollution():
    m = _world_with(objs=[("object.cuff", "袖扣")])
    assert add_alias(m, "object.cuff", "WU 信物")
    assert not add_alias(m, "object.cuff", "袖扣")            # 與 label 同 → 不加
    assert not add_alias(m, "object.cuff", "WU 信物")        # 重複 → 不加
    assert "wu信物" in entity_alias_set(m.get("object.cuff"))
    # 用 alias 指涉 → 對到同一 entity，不新增
    r = resolve_entity_reference("WU信物", world=m)
    assert r["resolved_entity_id"] == "object.cuff"
    assert len(m.by_kind(OBJECT)) == 1


# ── 3. 「那枚」→ current_focus ────────────────────────────────────────────────
def test_demonstrative_resolves_to_current_focus():
    m = _world_with(objs=[("object.cuff", "WU袖扣")])
    focus = {"id": "object.cuff", "label": "WU袖扣", "kind": "object"}
    r = resolve_entity_reference("那枚", world=m, current_focus=focus)
    assert r["resolved_entity_id"] == "object.cuff" and r["resolution_source"] == "current_focus"


# ── 4. 「剛才那個」→ recent_entities ──────────────────────────────────────────
def test_recent_resolves_first_recent():
    m = _world_with(objs=[("object.note", "筆記本")])
    recent = [{"id": "object.note", "label": "筆記本", "kind": "object"}]
    r = resolve_entity_reference("剛才那個", world=m, current_focus={"id": "object.x"}, recent_entities=recent)
    # 含「剛才」→ 偏 recent（不取 current_focus）
    assert r["resolved_entity_id"] == "object.note" and r["resolution_source"] == "recent_entities"


# ── 5. 「他說的地方」→ recent fact，不新增 area ───────────────────────────────
def test_fact_reference_resolves_to_recent_fact():
    m = _world_with(facts=[("fact.comm", "通訊設備在機房", "npc.李")])
    recent = [{"id": "fact.comm", "label": "通訊設備在機房", "kind": "fact"}]
    n_area = len(m.by_kind(AREA))
    r = resolve_entity_reference("他說的地方", world=m, recent_entities=recent,
                                 known_facts=[{"id": "fact.comm", "label": "通訊設備在機房"}])
    assert r["resolved_entity_id"] == "fact.comm" and r["resolution_source"] == "known_facts"
    assert len(m.by_kind(AREA)) == n_area                     # 不新增 area
    assert "fact.comm" in m.entities and len(m.by_kind(FACT)) == 1


# ── 6. 「那個 NPC」→ 最近對話 NPC ─────────────────────────────────────────────
def test_npc_reference_resolves_to_recent_actor():
    m = _world_with(actors=[("actor.xie", "謝博仁")])
    recent = [{"id": "actor.xie", "label": "謝博仁", "kind": "actor"}]
    r = resolve_entity_reference("那個 NPC", world=m, recent_entities=recent)
    assert r["resolved_entity_id"] == "actor.xie"
    r2 = resolve_entity_reference("剛才那個人", world=m, recent_entities=recent)
    assert r2["resolved_entity_id"] == "actor.xie"


# ── 7. 兩個筆記本 → ambiguous，不亂選 ─────────────────────────────────────────
def test_ambiguous_two_notebooks():
    m = _world_with(objs=[("object.n1", "工作筆記本"), ("object.n2", "私人筆記本")])
    r = resolve_entity_reference("那本筆記本", world=m)
    assert r["ambiguous"] is True and r["resolved_entity_id"] is None
    assert set(r["candidates"]) == {"object.n1", "object.n2"}


# ── 8. unresolved（無法判斷不亂猜）────────────────────────────────────────────
def test_unresolved_when_no_candidate():
    m = _world_with(objs=[("object.cuff", "袖扣")])
    r = resolve_entity_reference("那把鑰匙", world=m)         # 沒有鑰匙 entity
    assert r["resolved_entity_id"] is None and r["resolution_source"] == "unresolved"


# ══ Focus-Scope Patch：指代依 scope 解析，不被 current_focus=NPC 卡住 ════════════

def test_object_ref_not_captured_by_npc_focus():
    """UX #5：focus 是 NPC 時，「那個東西」仍應對到 object（recent），不被 NPC focus 卡住。"""
    m = _world_with(objs=[("object.badge", "徽章")], actors=[("actor.doc", "醫生")])
    focus = {"id": "actor.doc", "label": "醫生", "kind": "actor"}
    recent = [{"id": "object.badge", "label": "徽章", "kind": "object"}]
    r = resolve_entity_reference("那個東西", world=m, current_focus=focus, recent_entities=recent)
    assert r["resolved_entity_id"] == "object.badge"          # object，非 NPC
    assert r["resolution_source"] == "recent_entities"


def test_object_measure_ref_not_captured_by_npc_focus():
    m = _world_with(objs=[("object.cuff", "袖扣")], actors=[("actor.doc", "醫生")])
    focus = {"id": "actor.doc", "kind": "actor"}
    recent = [{"id": "object.cuff", "label": "袖扣", "kind": "object"}]
    for q in ("那枚", "這東西", "剛才那個東西"):
        r = resolve_entity_reference(q, world=m, current_focus=focus, recent_entities=recent)
        assert r["resolved_entity_id"] == "object.cuff", q     # 一律對到 object


def test_person_ref_prefers_actor_over_object_focus():
    """focus 是 object 時，「那個人」應對到 actor，不被 object focus 卡住。"""
    m = _world_with(objs=[("object.badge", "徽章")], actors=[("actor.doc", "醫生")])
    focus = {"id": "object.badge", "kind": "object"}
    recent = [{"id": "actor.doc", "label": "醫生", "kind": "actor"}]
    r = resolve_entity_reference("那個人", world=m, current_focus=focus, recent_entities=recent)
    assert r["resolved_entity_id"] == "actor.doc"


def test_npc_focus_actor_ref_uses_focus():
    """focus 是剛對話的 NPC 時，「那個人」直接用 focus actor。"""
    m = _world_with(actors=[("actor.doc", "醫生")])
    focus = {"id": "actor.doc", "kind": "actor"}
    r = resolve_entity_reference("那個人", world=m, current_focus=focus)
    assert r["resolved_entity_id"] == "actor.doc" and r["resolution_source"] == "current_focus"


def test_fact_direction_ref_resolves_to_fact_no_new_area():
    """「他說的方向」→ recent fact（route-related），不新增 area/exit。"""
    m = _world_with(facts=[("fact.route", "出口往東邊走", "npc.李")])
    recent = [{"id": "fact.route", "label": "出口往東邊走", "kind": "fact"}]
    n_area = len(m.by_kind(AREA))
    r = resolve_entity_reference("他說的方向", world=m, recent_entities=recent,
                                 known_facts=[{"id": "fact.route", "label": "出口往東邊走"}])
    assert r["resolved_entity_id"] == "fact.route" and r["resolution_source"] == "known_facts"
    assert len(m.by_kind(AREA)) == n_area                     # 不新增 area
    from core.world.model import EXIT
    assert len(m.by_kind(EXIT)) == 0                          # 不新增 exit


def test_object_ref_unresolved_when_no_object_not_npc():
    """object scope 但沒有任何 object → unresolved（不退回 NPC focus、不 ambiguous）。"""
    m = _world_with(actors=[("actor.doc", "醫生")])
    focus = {"id": "actor.doc", "kind": "actor"}
    r = resolve_entity_reference("那個東西", world=m, current_focus=focus)
    assert r["resolved_entity_id"] is None
    assert r["resolution_source"] == "unresolved" and not r["ambiguous"]


def test_object_ref_falls_to_visible_then_inventory():
    m = _world_with(actors=[("actor.doc", "醫生")])
    focus = {"id": "actor.doc", "kind": "actor"}
    vis = [{"id": "actor.doc", "kind": "actor", "label": "醫生"},
           {"id": "object.key", "kind": "object", "label": "鑰匙"}]
    r = resolve_entity_reference("那個東西", world=m, current_focus=focus, visible_entities=vis)
    assert r["resolved_entity_id"] == "object.key"            # 跳過 NPC，取 visible object


def test_ambiguous_still_returned_within_scope():
    """object scope 下兩個同名 object label 命中 → 仍 ambiguous，不亂猜。"""
    m = _world_with(objs=[("object.n1", "工作筆記本"), ("object.n2", "私人筆記本")])
    r = resolve_entity_reference("那本筆記本", world=m)
    assert r["ambiguous"] and r["resolved_entity_id"] is None


# ── extract_reference（從整句抽指代片語）────────────────────────────────────────
def test_extract_reference():
    assert extract_reference("我檢查那枚袖扣").startswith("那枚")
    assert extract_reference("我去他說的機房").startswith("他說的")
    assert extract_reference("我跟那個人說話").startswith("那個")
    assert extract_reference("我往北走") is None             # 無指代


# ══ loop 整合 ═════════════════════════════════════════════════════════════════
def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def _levels(loop):
    return {tid: t.level for tid, t in (getattr(loop._reveal_ledger, "truths", {}) or {}).items()}


def test_observation_entity_resolution_and_no_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    # 登記並檢查一個物件 → current_focus
    loop._world_model_tick("我四處看", "桌上撿起一本筆記本。", None)
    loop._world_model_tick("我檢查那本筆記本", "你翻開筆記本。", None)
    before = _levels(loop)
    out = loop.step("我再仔細看那枚")                          # 「那枚」→ current_focus
    er = out["entity_resolution"]
    for k in ("query", "resolved_entity_id", "resolution_source", "ambiguous", "candidates"):
        assert k in er
    assert er["resolved_entity_id"] == loop._current_focus["id"]
    assert _levels(loop) == before                            # 指代解析不推 reveal


def test_review_mode_resolve_inventory_no_new_fact(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop._world.register(OBJECT, "鑰匙", id="object.key", state="taken")
    facts_before = {e.id for e in loop._world.by_kind(FACT)}
    out = loop.step("先退到外面整理，看看我那把鑰匙")
    # review 模式仍能解析隨身物（inventory），且不新增 fact、不新增 entity
    assert {e.id for e in loop._world.by_kind(FACT)} == facts_before
    assert "object.key" in loop._world.entities
    assert "entity_resolution" in out
