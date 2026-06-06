"""WorldDelta input normalization（v0.8.1）—— LLM/entity_delta 的 `id` 別名 → `entity_id`。

修：raw entity_delta 帶 `id`（非 entity_id）時，apply 路徑會丟失 id（登成 slug），且某些路徑會
讓 `WorldDelta.__init__() got an unexpected keyword argument 'id'`。改在 coerce / apply 層 normalize。
不改 WorldDelta dataclass 欄位名。
"""
from __future__ import annotations

import logging

import core.constants as C
from core.world.model import (
    WorldModel, normalize_entity_delta_dict, coerce_entity_deltas, OBJECT, FACT,
)


# ── 1. 接受外部 id 別名 ───────────────────────────────────────────────────────
def test_world_delta_accepts_external_id_alias():
    # apply 路徑（raw dict 帶 id）→ 登在正確 entity_id，而非 label slug
    m = WorldModel()
    m.apply_deltas([{"op": "register", "kind": "object", "label": "血跡", "id": "object.blood"}])
    assert "object.blood" in m.entities
    assert m.get("object.blood").label == "血跡"
    # coerce 路徑（id → entity_id）
    out = coerce_entity_deltas([{"op": "register", "kind": "object", "label": "紙條", "id": "object.note"}])
    assert len(out) == 1 and out[0].entity_id == "object.note"
    # set_state 帶 id 也生效
    m.register(OBJECT, "鑰匙", id="object.key")
    m.apply({"op": "set_state", "id": "object.key", "state": "taken"})
    assert m.get("object.key").state == "taken"


# ── 2. entity_id 照舊可用 ─────────────────────────────────────────────────────
def test_world_delta_entity_id_still_works():
    m = WorldModel()
    m.apply({"op": "register", "kind": "object", "label": "x", "entity_id": "object.x"})
    assert "object.x" in m.entities
    out = coerce_entity_deltas([{"op": "register", "kind": "fact", "label": "f", "entity_id": "fact.f"}])
    assert out[0].entity_id == "fact.f"
    # normalize：無 id → 原樣
    assert normalize_entity_delta_dict({"op": "register", "entity_id": "object.x"}) == \
        {"op": "register", "entity_id": "object.x"}


# ── 3. id 與 entity_id 衝突 → 拒絕 ────────────────────────────────────────────
def test_world_delta_conflicting_id_rejected():
    # 相同 → 正常
    ok = normalize_entity_delta_dict({"op": "register", "id": "object.a", "entity_id": "object.a"})
    assert ok is not None and ok.get("entity_id") == "object.a" and "id" not in ok
    # 不同 → None（拒絕）
    assert normalize_entity_delta_dict({"op": "register", "id": "a", "entity_id": "b"}) is None
    # apply 衝突 delta → 不登記、不污染
    m = WorldModel()
    r = m.apply({"op": "register", "kind": "object", "label": "x", "id": "a", "entity_id": "b"})
    assert r is None and len(m.entities) == 0
    # coerce 衝突 delta → 丟棄
    out = coerce_entity_deltas([{"op": "register", "kind": "object", "label": "x",
                                 "id": "a", "entity_id": "b"}])
    assert out == []


# ── 4. malformed delta 不污染 ─────────────────────────────────────────────────
def test_malformed_delta_does_not_pollute_worldmodel():
    m = WorldModel()
    bad = ["junk", 42, {"op": "drop"}, {"kind": "object"}, None,
           {"op": "register", "kind": "object", "label": "x", "id": "a", "entity_id": "b"}]
    m.apply_deltas(bad)                                  # apply 路徑：壞 item 全略過
    assert len(m.entities) == 0
    assert coerce_entity_deltas(bad) == []               # coerce 路徑：全丟棄
    assert coerce_entity_deltas("not-a-list") == []
    # normalize 對非 dict 回 None（不拋）
    for x in ("s", 1, None, []):
        assert normalize_entity_delta_dict(x) is None


# ── 5. LLM-style entity_delta 走 _world_model_tick 不再 "skipped" ──────────────
def test_llm_style_entity_delta_no_world_model_tick_skipped(monkeypatch, caplog):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    from core.models import DecisionPoint, BeatMeta
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    dp = DecisionPoint(situation_recap="", decision_type="action",
                       beat_meta=BeatMeta(beat_number=1),
                       entity_delta=[{"op": "register", "kind": "object", "label": "血跡",
                                      "id": "object.blood", "affords": ["inspect"]}])
    with caplog.at_level(logging.WARNING):
        loop._world_model_tick("我看四周", "桌上有血跡。", dp)
    # 不得出現 "world model tick skipped"，且實體有登記
    assert not any("world model tick skipped" in r.message for r in caplog.records)
    assert loop._world.find("血跡", kind=OBJECT) is not None
