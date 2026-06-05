"""WorldModel 抽象骨架 + 三個接線驗證案例。

1. stay-put：「不進 B 區」不得 move_to(area.B)。
2. withdraw：「先退到外面整理線索，不結束調查」→ current_area=area.outside_dock、ended=false。
3. object memory：story 提到「WU 袖扣」→ 登記 object，玩家之後可檢查。
"""
from __future__ import annotations

import core.constants as C
from core.world.model import (
    WorldModel, Entity, WorldDelta, Affordance,
    AREA, EXIT, OBJECT, ACTOR, FACT, INSPECT, MOVE_TO, SAFE_ZONE_AREA_ID,
)
from core.world.extractor import extract_entities


# ── 骨架基本：四型別 + 五 kind + 狀態/affordance + 持久化 ────────────────────
def test_model_basics():
    m = WorldModel()
    obj = m.register(OBJECT, "WU 袖扣")
    assert obj.kind == OBJECT and obj.state == "noticed" and INSPECT in obj.affords
    area = m.register(AREA, "B 區", id="area.B")
    assert m.get("area.B") is area
    assert m.find("袖扣").id == obj.id            # 以 label 子字串解析
    assert m.by_kind(OBJECT) == [obj]
    # WorldDelta 套用
    m.apply(WorldDelta(op="register", kind=FACT, label="出口鎖死", entity_id="fact.exit_locked"))
    assert m.get("fact.exit_locked").kind == FACT
    m.apply(WorldDelta(op="set_state", entity_id=obj.id, state="inspected"))
    assert m.get(obj.id).state == "inspected"
    # 持久化 round-trip
    m2 = WorldModel.from_dict(m.to_dict())
    assert m2.get("area.B").label == "B 區" and m2.get(obj.id).state == "inspected"


# ── #1 stay-put：被否定的區域不得 move_to ───────────────────────────────────
def test_stay_put_negation_blocks_move():
    from core.narrative.negative_intent import negated_targets
    m = WorldModel()
    m.register(AREA, "B 區", id="area.B")
    m.register(AREA, "走廊", id="area.corridor")
    neg = negated_targets("不進 B 區，先觀察")        # → ['B']
    moved, reason = m.move_to("B 區", negated=neg)
    assert not moved and reason == "negated"
    assert m.current_area != "area.B"
    # 沒被否定的區域可以移動
    moved2, _ = m.move_to("走廊", negated=neg)
    assert moved2 and m.current_area == "area.corridor"


# ── #2 withdraw：退到外面 → current_area=area.outside_dock，不結局 ───────────
def test_withdraw_to_safe_zone(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("先退到外面整理線索，不結束本次調查")
    assert not out.get("ended")                       # withdraw 不結局
    assert loop._world.current_area == SAFE_ZONE_AREA_ID   # 真的切到安全區
    assert loop._world.get(SAFE_ZONE_AREA_ID).state == "current"


# ── #3 object memory：story 提到袖扣 → 登記，可檢查 ─────────────────────────
def test_object_memory_extract_and_inspect():
    deltas = extract_entities("你蹲下時，在地板縫隙撿起一個刻著 WU 的袖扣。")
    m = WorldModel()
    m.apply_deltas(deltas)
    obj = m.find("袖扣", kind=OBJECT)
    assert obj is not None and "袖扣" in obj.label    # 世界登記了這個物件
    assert obj.state == "noticed"
    # 玩家之後檢查它 → 世界記得他查過
    m.inspect("袖扣")
    assert m.find("袖扣").state == "inspected"


def test_object_memory_via_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    # 直接餵含物件的敘事走 world model tick（FakeCaller 的固定敘事不含物件）
    loop._world_model_tick("我四處張望", "桌上攤開一本沾血的筆記本，旁邊掉落一張員工證。")
    assert loop._world.find("筆記本", kind=OBJECT) is not None
    assert loop._world.find("員工證", kind=OBJECT) is not None
    # 之後檢查
    loop._world_model_tick("我檢查那本筆記本", "你翻開筆記本。")
    assert loop._world.find("筆記本").state == "inspected"


# ── 不過度登記：純氛圍敘事(無前景化線索)不亂登記物件 ───────────────────────
def test_no_over_registration():
    # 沒有「撿起/桌上/發現…」這類線索 → 不把氛圍裡的名詞當物件
    deltas = extract_entities("走廊很暗，空氣裡有鹽和鐵鏽的味道。")
    objs = [d for d in deltas if d.kind == OBJECT]
    assert objs == []
