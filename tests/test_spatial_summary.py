"""Spatial UX Projection —— SpatialProjection → 玩家/QA 可讀短摘要（確定性、不餵 story、不 LLM）。"""
from __future__ import annotations

import core.constants as C
from core.world.model import (
    WorldModel, AREA, OBJECT, FACT, ROLE_SAFE_ZONE, ROLE_SITE,
)
from core.world.spatial import (
    build_spatial_projection, player_facing_spatial_summary, SPATIAL_SUMMARY_SOURCE,
)


def _site_world():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.set_area_role("area.site", ROLE_SITE)
    return m


# ── 基本分段 ─────────────────────────────────────────────────────────────────
def test_summary_sections_present():
    m = WorldModel()
    m.set_current_area("area.site", label="冷藏室")
    m.register_exit("北門", from_area="area.site", leads_to="area.hall", state="available")
    m.register_exit("鐵閘", from_area="area.site", leads_to="area.deep", state="locked")
    m.register(OBJECT, "員工證", id="object.badge", props={"area": "area.site"})
    txt, trunc = player_facing_spatial_summary(build_spatial_projection(m))
    assert "目前位置：冷藏室" in txt
    assert "可走路線：" in txt and "北門" in txt
    assert "被阻擋路線：" in txt and "鐵閘" in txt and "鎖住" in txt
    assert "眼前可互動物：" in txt and "員工證" in txt


# ── 不包含不存在的地圖元素 ───────────────────────────────────────────────────
def test_summary_no_phantom_elements():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    txt, _ = player_facing_spatial_summary(build_spatial_projection(m))
    assert "不存在的機房" not in txt
    assert "通往深處的密道" not in txt
    # 不捏造幻影路線/地點；補丁後真實 area 會有可推導的結構性 route（暫退安全區），
    # 故不再顯示「沒有明顯可走的出口」（該 marker 僅在真的無任何 route 時出現）。
    assert "暫退到安全區整理" in txt
    assert "沒有明顯可走的出口" not in txt


# ── safe_zone：site 物件不得寫成「眼前可互動物」，而是「已知但不在眼前」───────────
def test_safe_zone_site_objects_not_visible_in_summary():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(OBJECT, "掛鐘", id="object.clock", props={"area": "area.site"})
    m.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    m._set_current("area.safe")
    txt, _ = player_facing_spatial_summary(build_spatial_projection(m))
    # 掛鐘在站內，玩家在安全區 → 不在「眼前可互動物」
    assert "眼前可互動物" not in txt or "掛鐘" not in txt.split("眼前可互動物")[-1].split("\n")[0]
    # 應出現在「已知但不在眼前」
    assert "已知但不在眼前：" in txt and "掛鐘" in txt


# ── known_remote → 「已知但不在眼前」；locked → 「被阻擋路線」─────────────────────
def test_known_remote_and_locked_route_sections():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(FACT, "通訊設備在機房", id="fact.comm", props={"area": "area.site"})
    m.register(OBJECT, "WU袖扣", id="object.cuff", props={"area": "area.other"})  # 別區 → remote
    m.register_exit("深處的門", from_area="area.site", leads_to="area.deep", state="locked")
    txt, _ = player_facing_spatial_summary(build_spatial_projection(m))
    remote_seg = txt.split("已知但不在眼前：")[-1]
    assert "通訊設備在機房" in remote_seg          # fact → remote
    assert "WU袖扣" in remote_seg                  # 別區物件 → remote
    assert "被阻擋路線：" in txt and "深處的門" in txt and "鎖住" in txt


# ── safe retreat 段 ──────────────────────────────────────────────────────────
def test_safe_retreat_section():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(AREA, "安全區", id="area.safe", roles=[ROLE_SAFE_ZONE])
    m.register_exit("撤退通道", from_area="area.site", leads_to="area.safe", state="available")
    txt, _ = player_facing_spatial_summary(build_spatial_projection(m))
    assert "安全撤退路線：" in txt and "撤退通道" in txt


# ── taken/used 物件不列入「眼前可互動物」────────────────────────────────────────
def test_taken_object_not_in_interactables():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    m.register(OBJECT, "鑰匙", id="object.key", state="taken", props={"area": "area.site"})
    txt, _ = player_facing_spatial_summary(build_spatial_projection(m))
    assert "眼前可互動物" not in txt or "鑰匙" not in txt


# ── 字數上限 + 截斷旗標 ───────────────────────────────────────────────────────
def test_summary_char_cap_and_truncation_flag():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    for i in range(30):
        m.register(OBJECT, f"長名稱物件{i}" * 3, id=f"object.{i}", props={"area": "area.site"})
    txt, trunc = player_facing_spatial_summary(build_spatial_projection(m), max_chars=120)
    assert len(txt) <= 121                          # 120 + 省略號
    assert trunc is True


def test_truncation_flag_from_projection_limits():
    m = WorldModel()
    m.set_current_area("area.site", label="現場")
    for i in range(25):
        m.register(OBJECT, f"物{i}", id=f"object.{i}", props={"area": "area.site"})
    # 投影 visible 限額 20 → 截斷 → summary truncated 也為 True（即使文字沒超長）
    _, trunc = player_facing_spatial_summary(
        build_spatial_projection(m, limits={"visible_entities": 20}), max_chars=5000)
    assert trunc is True


# ══ observation 整合 ══════════════════════════════════════════════════════════
def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def test_observation_has_spatial_summary_fields(monkeypatch):
    import sys
    sys.path.insert(0, "dev/tools")
    from agent_play import _dp_to_obs
    loop = _started_loop(monkeypatch)
    out = loop.step("我往前走查看四周")
    obs = _dp_to_obs(loop, out["narrative"], out["decision_point"],
                     out.get("ended"), out.get("ending"), step_result=out)
    assert "spatial_summary" in obs and isinstance(obs["spatial_summary"], str)
    assert "spatial_summary_truncated" in obs
    assert obs["spatial_summary_source"] == SPATIAL_SUMMARY_SOURCE == "deterministic_projection"
    assert "目前位置" in obs["spatial_summary"]
    # 也在 spatial_debug 內
    assert out["spatial_debug"]["spatial_summary_source"] == "deterministic_projection"


def test_review_mode_has_summary_but_story_context_unchanged(monkeypatch):
    """review_mode 仍有玩家面 spatial_summary（面板用），但**不接 P5**——story context 不含它。"""
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    assert out["spatial_debug"]["spatial_summary"]          # review 模式也有摘要
    assert out["world_progress"]["investigation_state"] == "review_mode"
    # build_story_context 不含 spatial_summary（沒接 P5）
    from core.agents.story import build_story_context
    ctx = build_story_context(loop.bb, "整理線索")
    assert "spatial_summary" not in ctx and "spatial_debug" not in ctx


def test_summary_does_not_change_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    led = loop._reveal_ledger
    before = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    for _ in range(3):
        player_facing_spatial_summary(build_spatial_projection(loop._world))
        loop.spatial_debug()
    after = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    assert before == after                                  # 摘要不碰 reveal（gate 不受影響）
