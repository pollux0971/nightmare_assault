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


# ── 觀測投影：world_progress.world_model 含實體 + affordances_here ───────────
def test_observation_projects_world_model(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    # 餵含物件的敘事 → 登記 object
    loop._world_model_tick("我四處看", "桌上撿起一張員工證，旁邊有把鑰匙。")
    wp = loop.world_progress()
    wm = wp["world_model"]
    assert "current_area" in wm and "entities" in wm and "affordances_here" in wm
    labels = [e["label"] for e in wm["entities"] if e["kind"] == OBJECT]
    assert any("員工證" in l for l in labels) and any("鑰匙" in l for l in labels)
    # affordances_here 至少含一個 inspect
    assert any(a["verb"] == INSPECT for a in wm["affordances_here"])


# ── current_area ownership：WorldModel 是唯一權威 ──────────────────────────
def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


def test_withdraw_to_outside_persists_next_beat(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    assert not out.get("ended")
    assert loop._world.current_area == SAFE_ZONE_AREA_ID
    # 下一 beat（kernel 場景沒變）→ 不得被 scene sync 覆蓋回原區域
    loop._game_state.current_scene = loop._world_kernel_scene      # 模擬 kernel 沒移動
    loop._world_model_tick("我翻看手裡的筆記", "你在外面整理線索。")
    assert loop._world.current_area == SAFE_ZONE_AREA_ID
    assert loop.world_progress()["current_area"] == SAFE_ZONE_AREA_ID   # 觀測也從 WorldModel 投影


def test_kernel_sync_does_not_override_worldmodel_current_area(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop._world.withdraw_to_safe_zone()                            # WorldModel 說在 outside_dock
    assert loop._world.current_area == SAFE_ZONE_AREA_ID
    loop._game_state.current_scene = loop._world_kernel_scene      # kernel 本 beat 未移動
    loop._world_model_tick("我環顧四周", "海風從防波堤吹來。")
    assert loop._world.current_area == SAFE_ZONE_AREA_ID           # kernel scene sync **不覆蓋**


def test_stay_put_negative_intent_does_not_move_area(monkeypatch):
    loop = _started_loop(monkeypatch)
    area0 = loop._world.current_area
    loop._game_state.current_scene = loop._world_kernel_scene      # 玩家原地觀察 → kernel 不移動
    loop._world_model_tick("不進 B 區，先在原地觀察", "你站在原地，盯著走廊深處。")
    assert loop._world.current_area == area0                       # current_area 不變


def test_move_affordance_changes_area_via_world_delta(monkeypatch):
    loop = _started_loop(monkeypatch)
    start_area = loop._world.current_area
    loop._game_state.current_scene = "scene.newroom"              # 模擬 kernel 真的移動
    loop._world_model_tick("我推開門走進新房間", "你走進另一個房間。")
    assert loop._world.current_area == "scene.newroom"            # 透過 WorldDelta 改 current_area
    assert loop._world.current_area != start_area
    assert loop._world.get("scene.newroom").state == "current"


# ── 不過度登記：純氛圍敘事(無前景化線索)不亂登記物件 ───────────────────────
def test_no_over_registration():
    # 沒有「撿起/桌上/發現…」這類線索 → 不把氛圍裡的名詞當物件
    deltas = extract_entities("走廊很暗，空氣裡有鹽和鐵鏽的味道。")
    objs = [d for d in deltas if d.kind == OBJECT]
    assert objs == []


# ══ Exit / Route Entity Ownership ════════════════════════════════════════════
# WorldModel owns exits/routes；ExitResolver 只解析 affordance；唯 end_campaign 進結局。
from core.world.model import EXIT as _EXIT, MOVE_THROUGH as _MT
from core.narrative.exit_resolver import (
    resolve_exit_affordance, END_CAMPAIGN, WITHDRAW_TO, MOVE_THROUGH as _AFF_MT, OFFER, NO_EXIT,
)


# ── exit 狀態：unknown/known/available/locked/blocked/used + move_through ─────
def test_exit_states_and_move_through():
    m = WorldModel()
    m.set_current_area("area.room", label="房間")
    m.register_exit("北門", leads_to="area.hall", from_area="area.room", state="known")
    # known → 可通行
    moved, reason = m.move_through("北門")
    assert moved and reason == "moved_through"
    assert m.current_area == "area.hall"
    assert m.get("exit.北門").state == "used"          # 通行後標 used
    assert m.exits_here() == [] or all(e.state == "used" for e in m.by_kind(_EXIT))


# ── #2 locked exit 不移動 ───────────────────────────────────────────────────
def test_locked_exit_does_not_move():
    m = WorldModel()
    m.set_current_area("area.room", label="房間")
    m.register_exit("鐵閘", leads_to="area.vault", from_area="area.room", state="locked")
    moved, reason = m.move_through("鐵閘")
    assert not moved and reason == "locked"
    assert m.current_area == "area.room"               # 沒移動
    # blocked 同理
    m.register_exit("塌方通道", leads_to="area.cellar", from_area="area.room", state="blocked")
    moved2, reason2 = m.move_through("塌方通道")
    assert not moved2 and reason2 == "blocked" and m.current_area == "area.room"


# ── #3 negative intent 不選被否定 exit ───────────────────────────────────────
def test_negative_intent_does_not_select_negated_exit():
    from core.narrative.negative_intent import negated_targets
    m = WorldModel()
    m.set_current_area("area.room", label="房間")
    m.register_exit("通往 B 區的門", leads_to="area.B", from_area="area.room", state="known")
    neg = negated_targets("不進 B 區，留在原地")        # → ['B']
    moved, reason = m.move_through("通往 B 區的門", negated=neg)
    assert not moved and reason == "negated"
    assert m.current_area == "area.room"               # 被否定的 exit 不通行


# ── #1 + #4 affordance 分流：withdraw 不 ending / 唯 explicit end_campaign 才 ending ──
def test_exit_affordance_classification():
    # explicit 結束 → end_campaign（唯一進 EndingGate）
    assert resolve_exit_affordance("結束本次調查，接受目前結果").affordance == END_CAMPAIGN
    # 撤退/退到外面 → withdraw_to（續行，不結局）
    assert resolve_exit_affordance("先退到外面整理線索").affordance == WITHDRAW_TO
    assert resolve_exit_affordance("暫時撤退喘口氣").affordance == WITHDRAW_TO
    # 離開房間 → move_through（換場景，不結局）
    assert resolve_exit_affordance("離開這個房間").affordance == _AFF_MT
    # 「不結束本次調查」→ 不得 end_campaign（NegativeIntentGuard）
    assert resolve_exit_affordance("我不結束本次調查").affordance != END_CAMPAIGN
    # 純調查行動 → no_exit
    assert resolve_exit_affordance("我檢查桌上的筆記本").affordance == NO_EXIT


def test_withdraw_outside_does_not_end_via_loop(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    assert not out.get("ended")                        # withdraw → 不結局
    assert loop._world.current_area == SAFE_ZONE_AREA_ID
    # 本 beat 的 exit affordance 不是 end_campaign
    assert loop._exit_affordance.affordance != END_CAMPAIGN


def test_explicit_end_campaign_only_ending_via_loop(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("我結束本次調查，接受目前結果，頭也不回地離開")
    assert loop._exit_affordance.affordance == END_CAMPAIGN
    assert out.get("ended")                            # 唯 end_campaign 才結局
    assert (out.get("ending") or {}).get("via") == "player_exit"


# ══ Story Structured entity_delta ════════════════════════════════════════════
# story output 結構化 entity_delta（object/actor/fact，每 beat ≤3，malformed 不污染）。
from core.world.model import coerce_entity_deltas, STORY_ENTITY_KINDS, STORY_DELTA_CAP, ACTOR as _ACTOR


# ── coerce：合法 delta 通過、kind/op/cap/malformed 過濾 ───────────────────────
def test_coerce_entity_deltas_filters():
    raw = [
        {"op": "register", "kind": OBJECT, "label": "WU 袖扣", "affords": [INSPECT]},
        {"op": "register", "kind": "area", "label": "B 區"},        # area 不准 → 丟
        {"op": "register", "kind": EXIT, "label": "後門"},          # exit 不准 → 丟
        {"op": "register", "kind": OBJECT, "label": "   "},          # 空 label → 丟
        {"op": "move_current", "entity_id": "area.x"},              # op 不准 → 丟
        {"op": "set_state", "entity_id": "object.WU_袖扣", "state": "taken"},
        "not-a-dict",                                                # malformed → 丟
        {"kind": OBJECT, "label": "無 op"},                          # 無 op → 丟
    ]
    out = coerce_entity_deltas(raw)
    kinds_ok = all(d.op in ("register", "set_state") for d in out)
    assert kinds_ok
    regs = [d for d in out if d.op == "register"]
    assert len(regs) == 1 and regs[0].kind == OBJECT and regs[0].label == "WU 袖扣"
    assert any(d.op == "set_state" and d.state == "taken" for d in out)
    # 非 list / None → []
    assert coerce_entity_deltas(None) == [] and coerce_entity_deltas("x") == []


# ── cap：每 beat 最多 1–3 筆 ─────────────────────────────────────────────────
def test_coerce_entity_deltas_caps_at_three():
    raw = [{"op": "register", "kind": OBJECT, "label": f"物件{i}"} for i in range(6)]
    assert len(coerce_entity_deltas(raw)) == STORY_DELTA_CAP == 3


# ── apply_story_deltas：kind-guarded，回傳 changed id；不准改 area/exit 狀態 ────
def test_apply_story_deltas_kind_guard():
    m = WorldModel()
    m.set_current_area("area.room", label="房間")
    m.register_exit("北門", leads_to="area.hall", from_area="area.room", state="known")
    deltas = coerce_entity_deltas([
        {"op": "register", "kind": OBJECT, "label": "WU 袖扣"},
        {"op": "set_state", "entity_id": "exit.北門", "state": "used"},   # 改 exit → 被 guard 擋
    ])
    changed = m.apply_story_deltas(deltas)
    assert "object.WU_袖扣" in changed
    assert m.get("exit.北門").state == "known"          # exit 狀態未被 story 改動
    assert "exit.北門" not in changed


# ── #1 regression：story 提到「WU 袖扣」後，玩家下一 beat 可檢查它 ─────────────
def test_story_object_registered_and_inspectable_next_beat(monkeypatch):
    loop = _started_loop(monkeypatch)
    # beat A：story 結構化登記 object（直接走 tick，模擬 story 吐了 entity_delta）
    from core.models import DecisionPoint, BeatMeta
    dp = DecisionPoint(situation_recap="", decision_type="action",
                       beat_meta=BeatMeta(beat_number=1),
                       entity_delta=[{"op": "register", "kind": OBJECT,
                                      "label": "WU 袖扣", "affords": [INSPECT, "take"]}])
    loop._world_model_tick("我環顧四周", "桌上有一個袖扣。", dp)
    obj = loop._world.find("袖扣", kind=OBJECT)
    assert obj is not None and obj.state == "noticed"
    assert "object.WU_袖扣" in loop._changed_entities_this_beat   # 本 beat 變更含它
    # beat B：玩家檢查它 → 世界記得（noticed → inspected）
    loop._world_model_tick("我檢查那個袖扣", "你端詳袖扣。", None)
    assert loop._world.find("袖扣").state == "inspected"


# ── #2 regression：物件 state 可從 noticed 變 inspected（結構化 set_state）──────
def test_object_state_noticed_to_inspected_structured():
    m = WorldModel()
    m.apply_story_deltas(coerce_entity_deltas(
        [{"op": "register", "kind": OBJECT, "label": "員工證"}]))
    assert m.find("員工證").state == "noticed"
    m.apply_story_deltas(coerce_entity_deltas(
        [{"op": "set_state", "entity_id": "object.員工證", "state": "inspected"}]))
    assert m.find("員工證").state == "inspected"


# ── #3 regression：同物件再指涉仍對到同一 entity（label-slug 冪等）────────────
def test_same_object_reference_maps_to_same_entity():
    m = WorldModel()
    m.apply_story_deltas(coerce_entity_deltas(
        [{"op": "register", "kind": OBJECT, "label": "WU 袖扣"}]))
    first = m.find("袖扣", kind=OBJECT)
    # 下個 beat 再次登記同 label → 同一 entity，不重複新增
    m.apply_story_deltas(coerce_entity_deltas(
        [{"op": "register", "kind": OBJECT, "label": "WU 袖扣", "affords": [INSPECT]}]))
    assert len(m.by_kind(OBJECT)) == 1
    assert m.find("袖扣", kind=OBJECT).id == first.id


# ── #4 regression：malformed entity_delta 不得污染 WorldModel ─────────────────
def test_malformed_entity_delta_does_not_pollute(monkeypatch):
    loop = _started_loop(monkeypatch)
    n0 = len(loop._world.entities)
    from core.models import DecisionPoint, BeatMeta
    # entity_delta 是一坨垃圾：非 list、含壞 item、含禁止 kind
    bad = DecisionPoint(situation_recap="", decision_type="action",
                        beat_meta=BeatMeta(beat_number=1),
                        entity_delta="not-a-list-at-all")     # field_validator → []
    assert bad.entity_delta == []                              # DecisionPoint 不因此解析失敗
    loop._world_model_tick("我站著不動", "什麼也沒發生。", bad)
    # 沒有新增任何垃圾實體（fallback extractor 對純氛圍敘事也不登記）
    assert len(loop._world.entities) == n0
    # 直接餵壞 delta list 給 tick：禁止 kind / 壞 item 全被丟棄
    dp2 = DecisionPoint(situation_recap="", decision_type="action",
                        beat_meta=BeatMeta(beat_number=2),
                        entity_delta=[{"op": "register", "kind": "exit", "label": "鬼門"},
                                      "garbage", {"op": "drop_table"}])
    loop._world_model_tick("我看著牆", "牆上有霉斑。", dp2)
    assert loop._world.find("鬼門") is None                    # 禁止 kind 未進 WorldModel
    assert all(e.kind in STORY_ENTITY_KINDS | {AREA, EXIT}
               for e in loop._world.entities.values())        # 沒有非法 kind
