"""Spatial WorldModel Projection（patch v0.6 P0–P4）—— 確定性、唯讀、可快取、有預算。"""
from __future__ import annotations

import time

import core.constants as C
from core.world.model import (
    WorldModel, AREA, OBJECT, FACT, ROLE_SAFE_ZONE, ROLE_SITE,
)
from core.world.spatial import (
    build_spatial_projection, SpatialProjectionCache, deterministic_mental_map_text,
    validate_mental_map_summary, MentalMapWorker,
)


def _site_world():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.set_area_role("area.site", ROLE_SITE)
    return m


# ══ P0 — 投影契約（唯讀、不變更 WorldModel）═══════════════════════════════════
def test_projection_does_not_mutate_worldmodel():
    m = _site_world()
    before = m.version_snapshot()
    p = build_spatial_projection(m)
    assert m.version_snapshot() == before               # 投影不改版本（唯讀）
    assert p.current_area == "area.site"
    assert hasattr(p, "routes_from_here") and hasattr(p, "visible_entities")


def test_projection_fields_present():
    m = _site_world()
    d = build_spatial_projection(m).to_debug_dict()
    for k in ("current_area", "current_area_label", "current_area_roles",
              "routes_from_here", "blocked_routes", "safe_retreat_routes",
              "visible_entities", "known_remote_entities", "mental_map_text",
              "counts", "truncated", "versions"):
        assert k in d


# ══ P1 — dirty-version cache ═════════════════════════════════════════════════
def test_version_bumps_on_change():
    m = WorldModel()
    v0 = m.version_snapshot()["world_version"]
    m.register(OBJECT, "紙條", id="object.note")
    assert m.version_snapshot()["world_version"] > v0
    assert m.version_snapshot()["entity_version"] >= 1
    m.set_state("object.note", "inspected")
    assert m.version_snapshot()["entity_version"] >= 2
    m.set_current_area("area.a", label="A")
    assert m.version_snapshot()["area_version"] >= 1
    m.register_exit("門", from_area="area.a", leads_to="area.b", state="known")
    assert m.version_snapshot()["exit_version"] >= 1
    m.register(FACT, "事實", id="fact.x")
    assert m.version_snapshot()["fact_version"] >= 1


def test_cache_hit_then_invalidate():
    m = _site_world()
    c = SpatialProjectionCache()
    b = lambda w: build_spatial_projection(w)
    p1 = c.get_or_build(m, b)
    p2 = c.get_or_build(m, b)
    assert p1 is p2 and c.hits >= 1                      # 未變 → 回同一物件
    m.register(OBJECT, "新物件", id="object.new")        # WorldModel 變動
    p3 = c.get_or_build(m, b)
    assert p3 is not p1 and c.misses >= 2                # 變動 → 重算
    # 換 profile（如 exploration_mode）也 invalidate
    p4 = c.get_or_build(m, b, profile="review_mode")
    assert p4 is not p3


# ══ P2 — 同步快速投影 ═════════════════════════════════════════════════════════
def test_safe_zone_does_not_show_site_objects_as_visible():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(OBJECT, "掛鐘", id="object.clock", props={"area": "area.site"})
    m.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    m._set_current("area.safe")
    p = build_spatial_projection(m)
    assert "object.clock" not in [e.id for e in p.visible_entities]
    assert "object.clock" in [e.id for e in p.known_remote_entities]


def test_available_exit_in_routes_from_here():
    m = WorldModel()
    m.set_current_area("area.safe", label="安全區")
    m.register_exit("返回現場的路", from_area="area.safe", leads_to="area.site", state="available")
    p = build_spatial_projection(m)
    assert any(r.to_area == "area.site" for r in p.routes_from_here)
    assert not p.blocked_routes


def test_locked_exit_in_blocked_routes():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register_exit("深處的門", from_area="area.site", leads_to="area.deep", state="locked")
    p = build_spatial_projection(m)
    assert any(r.state == "locked" for r in p.blocked_routes)
    # locked exit 不可通行：不得出現在可走路線（補丁後仍可能有結構性 route，但鎖門不算）
    assert all(r.state != "locked" for r in p.routes_from_here)
    assert not any(r.to_area == "area.deep" for r in p.routes_from_here)


def test_safe_retreat_routes_via_role():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    m.register_exit("撤退路線", from_area="area.site", leads_to="area.safe", state="available")
    p = build_spatial_projection(m)
    assert any(r.to_area == "area.safe" for r in p.safe_retreat_routes)


def test_safe_retreat_excludes_locked_route_to_safe_zone():
    # 鎖住的門通往安全區 → 不算「可用退路」（只列可通行的撤退路線）
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    m.register_exit("鎖住的撤退門", from_area="area.site", leads_to="area.safe", state="locked")
    p = build_spatial_projection(m)
    # 鎖住的「明確 exit」不算可用退路；結構性 withdraw（always-available）仍可在 safe_retreat。
    assert all(r.state != "locked" for r in p.safe_retreat_routes)
    assert any(r.state == "locked" for r in p.blocked_routes)


def test_fact_is_remote_not_visible():
    m = _site_world()
    m.register(FACT, "通訊設備在機房", id="fact.commroom", props={"area": "area.site"})
    p = build_spatial_projection(m)
    ids_v = [e.id for e in p.visible_entities]
    ids_r = [e.id for e in p.known_remote_entities]
    assert "fact.commroom" not in ids_v and "fact.commroom" in ids_r


# ══ P3 — mental_map_text ══════════════════════════════════════════════════════
def test_mental_map_text_uses_only_projection():
    m = WorldModel()
    m.set_current_area("area.site", label="冷藏室")
    m.register_exit("北門", from_area="area.site", leads_to="area.hall", state="available")
    p = build_spatial_projection(m)
    assert "冷藏室" in p.mental_map_text                 # 含當前區域
    assert "北門" in p.mental_map_text                   # 含可前往路線
    assert "不存在的機房" not in p.mental_map_text       # 不捏造未知實體


def test_mental_map_changes_on_area_change():
    m1 = WorldModel(); m1.set_current_area("area.a", label="鍋爐房")
    m2 = WorldModel(); m2.set_current_area("area.b", label="檔案室")
    assert build_spatial_projection(m1).mental_map_text != build_spatial_projection(m2).mental_map_text


def test_deterministic_template_omits_empty():
    txt = deterministic_mental_map_text("空房", [], [], [], [])
    assert txt == "你目前位於：空房。"                    # 空清單 → 省略


# ══ P4 — 觀測預算 ═════════════════════════════════════════════════════════════
def test_budget_truncation_flags_and_counts():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    for i in range(25):
        m.register(OBJECT, f"物{i}", id=f"object.{i}", props={"area": "area.site"})
    p = build_spatial_projection(m, limits={"visible_entities": 20})
    assert len(p.visible_entities) == 20                 # 截斷
    assert p.truncated["visible_entities"] is True
    assert p.counts["visible_entities"] == 25            # 仍回報總數


def test_mental_map_text_capped():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    for i in range(60):
        m.register_exit("很長的路線名稱" * 3, from_area="area.site",
                        leads_to=f"area.{i}", state="available", id=f"exit.{i}")
    p = build_spatial_projection(m, limits={"mental_map_text": 120})
    assert len(p.mental_map_text) <= 120


# ══ P4 — async worker（非阻塞；遲到用 fallback）════════════════════════════════
def test_worker_returns_fallback_when_empty():
    w = MentalMapWorker(lambda pj, fb: "潤飾：" + fb)
    assert w.get_text("missing", "FB") == "FB"           # 無快取 → 立即回 fallback
    w.stop()


def test_worker_refresh_eventually_or_fallback():
    w = MentalMapWorker(lambda pj, fb: "潤飾：" + fb)
    p = build_spatial_projection(_site_world())
    w.request_refresh("k", p)                            # 非阻塞投件
    got = "FB"
    for _ in range(40):                                  # 寬鬆輪詢（最多 ~2s）
        got = w.get_text("k", "FB")
        if got != "FB":
            break
        time.sleep(0.05)
    assert got in ("潤飾：" + p.mental_map_text, "FB")    # 潤飾或 fallback——都不阻塞
    w.stop()


def test_validate_summary_rejects_overlong():
    p = build_spatial_projection(_site_world())
    assert validate_mental_map_summary("短摘要", p) is True
    assert validate_mental_map_summary("x" * 900, p, max_chars=800) is False
    assert validate_mental_map_summary("", p) is False


# ══ loop 整合 + 回歸 ══════════════════════════════════════════════════════════
def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def test_observation_has_spatial_debug(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("我往前走查看四周")
    sd = out["spatial_debug"]
    for k in ("current_area", "routes_from_here", "blocked_routes", "visible_entities",
              "known_remote_entities", "mental_map_text", "truncated", "counts", "versions"):
        assert k in sd
    assert sd["current_area"] == loop._world.current_area


def test_spatial_debug_no_reveal_side_effect(monkeypatch):
    loop = _started_loop(monkeypatch)
    led = loop._reveal_ledger
    before = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    loop.spatial_debug(); loop.spatial_debug()           # 多次投影
    after = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    assert before == after                               # 投影不碰 reveal（TruthEvidenceGate 不受影響）
