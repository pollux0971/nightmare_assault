"""Spatial Routes Projection（補丁）測試。

問題：SpatialSummary 的「可走路線」幾乎永遠是「沒有明顯可走的出口」。
本補丁從 WorldModel roles / current_area / previous_area / exits 投影**最小可用** routes
（無 pathfinding、不擴 scene graph、不新增 entity）。

必測（依規格）：
  - previous_area → 返回上一區
  - active area 可撤退 → safe_zone
  - safe_zone → return_to_site
  - locked exit 在 blocked，不在 available
  - 無任何 route 才顯示「沒有明顯可走的出口」
  - 不改 reveal / PlayerState（投影唯讀）
"""
from __future__ import annotations

from core.world.model import (
    AREA, EXIT, ROLE_CAMPAIGN_EXIT, ROLE_ENTRY, ROLE_SAFE_ZONE, ROLE_SITE, WorldModel)
from core.world.spatial import (
    ROUTE_CAMPAIGN_EXIT, ROUTE_RETURN_PREVIOUS, ROUTE_RETURN_SITE, ROUTE_WITHDRAW_SAFE,
    build_spatial_projection, derive_structural_routes, player_facing_spatial_summary)


def _route_ids(routes):
    return [r.exit_id for r in routes]


def _labels(routes):
    return [r.label for r in routes]


# ── 1. previous_area → 返回上一區 ────────────────────────────────────────────

def test_previous_area_shows_return_previous():
    w = WorldModel()
    w.set_current_area("area.site", label="研究站大廳")
    w.set_current_area("area.deep", label="地下檔案層")     # previous_area = area.site
    assert w.previous_area == "area.site"
    p = build_spatial_projection(w)
    assert ROUTE_RETURN_PREVIOUS in _route_ids(p.routes_from_here)
    r = [x for x in p.routes_from_here if x.exit_id == ROUTE_RETURN_PREVIOUS][0]
    assert r.to_area == "area.site" and "研究站大廳" in r.label


def test_previous_area_persists_through_serialization():
    w = WorldModel()
    w.set_current_area("a1", label="一")
    w.set_current_area("a2", label="二")
    w2 = WorldModel.from_dict(w.to_dict())
    assert w2.previous_area == "a1" and w2.current_area == "a2"


# ── 2. active area 可撤退 → safe_zone ────────────────────────────────────────

def test_active_area_shows_withdraw_to_safe_zone():
    w = WorldModel()
    w.set_current_area("area.site", label="現場")
    w.set_area_role("area.site", ROLE_SITE)
    p = build_spatial_projection(w)
    assert ROUTE_WITHDRAW_SAFE in _route_ids(p.routes_from_here)
    assert any("安全區" in l for l in _labels(p.routes_from_here))
    # withdraw 同時是一條安全撤退路線
    assert any(r.exit_id == ROUTE_WITHDRAW_SAFE for r in p.safe_retreat_routes)


# ── 3. safe_zone → return_to_site ────────────────────────────────────────────

def test_safe_zone_shows_return_to_site():
    w = WorldModel()
    w.set_current_area("area.site", label="研究站")
    w.set_area_role("area.site", ROLE_SITE)
    w.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    w.set_current_area("area.safe")                        # 撤到安全區
    p = build_spatial_projection(w)
    assert ROUTE_RETURN_SITE in _route_ids(p.routes_from_here) or \
        ROUTE_RETURN_PREVIOUS in _route_ids(p.routes_from_here)
    # 在安全區 → 不再提供「暫退安全區」
    assert ROUTE_WITHDRAW_SAFE not in _route_ids(p.routes_from_here)
    assert any("返回現場" in l or "研究站" in l for l in _labels(p.routes_from_here))


def test_safe_zone_return_to_site_prefers_site_label():
    """polish：在 safe_zone 且上一區同時是 site/active_area → return_previous label 顯示「返回現場」。"""
    from core.world.model import ROLE_ACTIVE_AREA
    w = WorldModel()
    w.set_current_area("area.site", label="研究站")
    w.set_area_role("area.site", ROLE_ACTIVE_AREA)         # 撤退前的調查現場
    w.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    w.set_current_area("area.safe")                        # previous_area = area.site；cur = safe_zone
    p = build_spatial_projection(w)
    r = [x for x in p.routes_from_here if x.exit_id == ROUTE_RETURN_PREVIOUS][0]
    assert r.label.startswith("返回現場") and "研究站" in r.label   # 標籤優先「返回現場」
    assert r.to_area == "area.site"                        # route target 不變
    assert "return_site" in (r.roles or [])


def test_non_safe_zone_return_previous_keeps_previous_label():
    """非 safe_zone：即使上一區是 site，label 仍為「返回上一個區域」（polish 只在 safe_zone 生效）。"""
    from core.world.model import ROLE_ACTIVE_AREA
    w = WorldModel()
    w.set_current_area("area.site", label="研究站")
    w.set_area_role("area.site", ROLE_ACTIVE_AREA)
    w.set_current_area("area.deep", label="深處")          # cur 非 safe_zone
    p = build_spatial_projection(w)
    r = [x for x in p.routes_from_here if x.exit_id == ROUTE_RETURN_PREVIOUS][0]
    assert r.label.startswith("返回上一個區域")


def test_return_site_uses_entry_label_when_only_entry():
    w = WorldModel()
    w.set_current_area("area.entry", label="入口大廳")
    w.set_area_role("area.entry", ROLE_ENTRY)
    w.set_current_area("area.deep", label="深處")
    p = build_spatial_projection(w)
    # site_area_id() → entry；但 entry==previous → 由 return_previous 覆蓋（dedup），標籤含入口名
    labels = " ".join(_labels(p.routes_from_here))
    assert "入口大廳" in labels


# ── 4. locked exit 在 blocked，不在 available ────────────────────────────────

def test_locked_exit_blocked_not_available():
    w = WorldModel()
    w.set_current_area("area.x", label="X")
    w.register(EXIT, "防火門", id="exit.fire", state="locked",
               props={"area": "area.x", "leads_to": "area.y", "requires": ["鑰卡"]})
    p = build_spatial_projection(w)
    assert any(r.state == "locked" for r in p.blocked_routes)
    assert all(r.state != "locked" for r in p.routes_from_here)
    assert not any(r.to_area == "area.y" for r in p.routes_from_here)
    # 受阻路線 player-facing 文案含 requires
    txt, _ = player_facing_spatial_summary(p)
    assert "需要鑰卡" in txt and "防火門" in txt


def test_available_exit_in_routes_dedups_structural():
    w = WorldModel()
    w.set_current_area("area.a", label="A")
    w.set_current_area("area.b", label="B")               # previous = area.a
    # 明確可通行 exit 回到 area.a → 結構性 return_previous 應 dedup（不重複到 area.a）
    w.register(EXIT, "回頭的門", id="exit.back", state="available",
               props={"area": "area.b", "leads_to": "area.a"})
    p = build_spatial_projection(w)
    to_a_count = sum(1 for r in p.routes_from_here if r.to_area == "area.a")
    assert to_a_count == 1                                 # 不重複提供同一目標


# ── 5. 無任何 route 才顯示「沒有明顯可走的出口」───────────────────────────────

def test_no_routes_marker_only_when_truly_empty():
    w = WorldModel()                                       # 空世界，無 current_area
    p = build_spatial_projection(w)
    assert p.routes_from_here == []
    txt, _ = player_facing_spatial_summary(p)
    assert "沒有明顯可走的出口" in txt


def test_real_area_does_not_show_no_exits_marker():
    w = WorldModel()
    w.set_current_area("area.site", label="現場")
    txt, _ = player_facing_spatial_summary(build_spatial_projection(w))
    assert "沒有明顯可走的出口" not in txt                  # 有結構性 route → 不顯示


# ── campaign_exit ────────────────────────────────────────────────────────────

def test_campaign_exit_route_when_role_exists():
    w = WorldModel()
    w.set_current_area("area.site", label="現場")
    w.set_area_role("area.site", ROLE_SITE)
    w.register(AREA, "出口閘門", id="area.exit", roles=[ROLE_CAMPAIGN_EXIT])
    p = build_spatial_projection(w)
    assert ROUTE_CAMPAIGN_EXIT in _route_ids(p.routes_from_here)
    assert any("離開" in l or "結束" in l for l in _labels(p.routes_from_here))


# ── 6. 投影唯讀：不改 reveal / PlayerState / world entity ─────────────────────

def test_projection_is_read_only():
    w = WorldModel()
    w.set_current_area("area.site", label="現場")
    w.set_area_role("area.site", ROLE_SITE)
    n_before = len(w.entities)
    cur_before, prev_before = w.current_area, w.previous_area
    build_spatial_projection(w)
    build_spatial_projection(w)                            # 多次呼叫
    assert len(w.entities) == n_before                     # 不新增 area/exit entity
    assert w.current_area == cur_before and w.previous_area == prev_before
    # 結構性 route 的目標若是 fallback safe_zone，也不得被實際登記進 world
    assert "area.safe_zone" not in w.entities
