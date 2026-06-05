"""Theme-Agnostic World Roles —— safe_zone/site/entry/return 改用 WorldModel role，不硬寫地名。"""
from __future__ import annotations

import json

import core.constants as C
from core.world.model import (
    WorldModel, AREA,
    ROLE_SAFE_ZONE, ROLE_SITE, ROLE_ENTRY, ROLE_ACTIVE_AREA,
    SAFE_ZONE_AREA_ID, LEGACY_SAFE_ZONE_ID,
)


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


# ── WorldModel role 查詢 ─────────────────────────────────────────────────────
def test_role_queries():
    m = WorldModel()
    m.set_current_area("area.lobby", label="大廳")
    m.set_area_role("area.lobby", ROLE_ENTRY)
    m.set_area_role("area.lobby", ROLE_SITE)
    assert m.entry_area_id() == "area.lobby"
    assert m.site_area_id() == "area.lobby"
    assert m.area_with_role(ROLE_SITE).id == "area.lobby"
    assert m.areas_with_role(ROLE_ENTRY) == [m.get("area.lobby")]
    # active_area 優先於 site
    m.register(AREA, "走廊", id="area.hall", roles=[ROLE_ACTIVE_AREA])
    assert m.site_area_id() == "area.hall"
    # exclusive：active_area 轉移
    m.set_area_role("area.lobby", ROLE_ACTIVE_AREA, exclusive=True)
    assert m.area_with_role(ROLE_ACTIVE_AREA).id == "area.lobby"
    assert ROLE_ACTIVE_AREA not in m.get("area.hall").roles


# ── withdraw 用 role=safe_zone，不依賴 outside_dock ───────────────────────────
def test_withdraw_uses_role_safe_zone():
    m = WorldModel()
    m.set_current_area("area.lobby", label="大廳")
    m.register(AREA, "停車場", id="area.parking", roles=[ROLE_SAFE_ZONE])   # 主題無關安全區
    sid = m.withdraw_to_safe_zone()
    assert sid == "area.parking"                       # 用 role，不是 outside_dock
    assert m.current_area == "area.parking"
    assert m.is_safe_zone("area.parking")
    assert "outside_dock" not in sid


# ── 預設安全區 id 是主題無關（非 outside_dock）────────────────────────────────
def test_default_safe_zone_is_theme_agnostic():
    m = WorldModel()
    assert m.safe_zone_id() == SAFE_ZONE_AREA_ID == "area.safe_zone"
    assert "outside_dock" not in SAFE_ZONE_AREA_ID
    sid = m.withdraw_to_safe_zone()
    assert sid == "area.safe_zone" and ROLE_SAFE_ZONE in m.get(sid).roles


# ── role 缺失時 fallback 舊 safe zone（相容，不破壞舊存檔/測試）────────────────
def test_legacy_safe_zone_fallback_compat():
    m = WorldModel()
    m.register(AREA, "外面", id=LEGACY_SAFE_ZONE_ID, origin="kernel")  # 舊存檔的 outside_dock
    assert m.safe_zone_id() == LEGACY_SAFE_ZONE_ID      # 沒 role → fallback 舊常數
    assert m.is_safe_zone(LEGACY_SAFE_ZONE_ID)          # 舊 id 仍被認得為安全區
    sid = m.withdraw_to_safe_zone()
    assert sid == LEGACY_SAFE_ZONE_ID                   # 沿用舊區域，標上 role
    assert ROLE_SAFE_ZONE in m.get(LEGACY_SAFE_ZONE_ID).roles


# ── 持久化保留 roles ─────────────────────────────────────────────────────────
def test_roles_persist_round_trip():
    m = WorldModel()
    m.set_current_area("area.a", label="A")
    m.set_area_role("area.a", ROLE_SITE)
    m2 = WorldModel.from_dict(m.to_dict())
    assert ROLE_SITE in m2.get("area.a").roles
    # 舊 dict（無 roles 欄位）也能載入
    old = {"current_area": "area.x",
           "entities": {"area.x": {"id": "area.x", "kind": "area", "label": "X",
                                   "state": "current", "props": {}, "affords": [], "origin": "kernel"}}}
    m3 = WorldModel.from_dict(old)
    assert m3.get("area.x").roles == []


# ══ loop 整合 ═════════════════════════════════════════════════════════════════
def test_seed_tags_entry_and_site(monkeypatch):
    loop = _started_loop(monkeypatch)
    start = loop._world.current_area
    assert ROLE_ENTRY in loop._world.get(start).roles
    assert ROLE_SITE in loop._world.get(start).roles


def test_withdraw_via_loop_uses_role_and_observation_area_roles(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    sid = loop._world.current_area
    assert loop._world.is_safe_zone(sid)
    assert ROLE_SAFE_ZONE in loop._world.get(sid).roles
    assert "outside_dock" not in sid
    # observation.world_progress 顯示 area_roles
    ar = out["world_progress"]["area_roles"]
    assert isinstance(ar, dict)
    assert any(ROLE_SAFE_ZONE in roles for roles in ar.values())
    assert any(ROLE_SITE in roles for roles in ar.values())
    # world_model entity projection 也帶 roles
    wm_ents = out["world_progress"]["world_model"]["entities"]
    assert any(ROLE_SAFE_ZONE in (e.get("roles") or []) for e in wm_ents)


def test_review_option_and_observation_have_no_hardcoded_theme(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    opt = [getattr(o, "text", "") for o in out["decision_point"].suggested_options]
    assert any("繼續調查" in t for t in opt)            # 返回現場（去主題化）
    assert not any("研究站" in t for t in opt)
    # 整份 world_progress 不含硬編碼主題字串
    blob = json.dumps(out["world_progress"], ensure_ascii=False, default=str)
    assert "outside_dock" not in blob
    assert "研究站" not in blob


def test_reenter_restores_site_area(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop.step("先退到外面整理線索，不結束本次調查")
    assert loop._world.is_safe_zone(loop._world.current_area)
    out = loop.step("我返回現場，繼續調查")
    assert loop._exploration_mode == "active_exploration"
    assert not loop._world.is_safe_zone(loop._world.current_area)   # 真的離開安全區
