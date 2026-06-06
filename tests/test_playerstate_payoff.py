"""PlayerState Payoff Materialization（補丁）測試。

問題：Step 5 已有 inventory_entities / known_facts，但玩家檢查/研究/問 NPC 後玩家面仍空。
本補丁讓「取得 / 保存 / 得知 / 確認」沉澱進**既有** PlayerState surface（不新增 surface）。

必測（依規格）：
  - take_object_action_adds_inventory
  - inspect_object_does_not_take_inventory
  - npc_prose_fact_appears_in_known_facts
  - known_facts_do_not_push_reveal
  - reveal_reward_observed_adds_public_known_fact
  - hidden_truth_raw_not_in_known_facts
  - taken_object_removed_from_visible_entities
  - review_mode_can_show_inventory_and_known_facts_without_new_fact
"""
from __future__ import annotations

from core.narrative.revelation import (
    RevealLedger, apply_reveal_reward, reveal_public_facts)
from core.world.model import ACTOR, FACT, NPC_ENTITY_KINDS, OBJECT, WorldModel
from core.world.player_state import project_inventory, project_known_facts
from core.world.spatial import build_spatial_projection


# ── 純元件：WorldModel.take / reveal_public_facts ────────────────────────────

def test_take_sets_taken_and_carried():
    w = WorldModel()
    w.register(OBJECT, "徽章", id="object.badge", state="noticed")
    e = w.take("object.badge")
    assert e is not None and e.state == "taken" and e.props.get("carried") is True
    assert any(i["id"] == "object.badge" for i in project_inventory(w))


def test_take_does_not_revive_used_object():
    w = WorldModel()
    w.register(OBJECT, "電池", id="object.batt", state="used")
    assert w.take("object.batt") is None


def test_reveal_public_facts_only_observed_plus_no_content():
    led = RevealLedger()
    t = led.get_or_create("t1", title="核心真相", content="你也是實驗體之一")
    t.level = "hinted"
    assert reveal_public_facts(led) == []                 # hinted 太弱 → 不投影
    t.level = "observed"
    pf = reveal_public_facts(led)
    assert pf and pf[0]["title"] == "核心真相" and pf[0]["level"] == "observed"
    assert all("content" not in p for p in pf)            # 永不含 content


def test_reveal_public_fact_confirmed_maps_to_confirmed_public():
    led = RevealLedger()
    t = led.get_or_create("t1", title="x")
    t.level = "confirmed"
    assert reveal_public_facts(led)[0]["level"] == "confirmed_public"


# ── 迴圈整合：_world_model_tick / _materialize_inventory / _materialize_public_facts ──

def _mk_loop():
    from core.blackboard import Blackboard
    from core.orchestrator_loop import BeatLoop
    from core.persistence.db import Database
    from core.signal import SignalBus
    from tests.test_narrative_v2_integration_nr import FakeCaller
    loop = BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(), run_id="r", use_kernel=True)
    loop._world = WorldModel()
    loop._game_state = None
    loop._world_kernel_scene = None
    loop._scene_changed_this_beat = False
    loop._new_actor_this_beat = False
    loop._current_focus = None
    loop._recent_entities = []
    loop._changed_entities_this_beat = []
    loop._changed_entities_detail = []
    loop._inventory_delta_this_beat = []
    loop._known_fact_delta_this_beat = []
    loop._skipped_materialization_reason = ""
    loop.known_npcs = []
    loop._reveal_ledger = None
    return loop


def test_take_object_action_adds_inventory():
    loop = _mk_loop()
    loop._world.register(OBJECT, "徽章", id="object.badge", state="noticed")
    loop._world_model_tick("我撿起那枚徽章放進口袋", "", None)
    assert loop._world.get("object.badge").state == "taken"
    inv = project_inventory(loop._world)
    assert any(i["label"] == "徽章" for i in inv)
    assert any(d["id"] == "object.badge" for d in loop._inventory_delta_this_beat)


def test_inspect_object_does_not_take_inventory():
    loop = _mk_loop()
    loop._world.register(OBJECT, "徽章", id="object.badge", state="noticed")
    loop._world_model_tick("我仔細檢查徽章，端詳上面的紋路", "", None)
    assert loop._world.get("object.badge").state == "inspected"   # 檢查 → inspected，不是 taken
    assert project_inventory(loop._world) == []
    assert loop._inventory_delta_this_beat == []


def test_take_verb_without_object_records_skip_reason():
    loop = _mk_loop()                                     # 世界裡沒有可拿的物件
    loop._world_model_tick("我撿起地上的東西", "", None)
    assert loop._skipped_materialization_reason == "take_verb_no_object_resolved"


def test_npc_prose_fact_appears_in_known_facts():
    w = WorldModel()
    from core.narrative.npc_prose_facts import extract_npc_prose_facts
    deltas = extract_npc_prose_facts("通訊設備就在B2機房，那扇防火門已經鎖死了。", npc_id="醫生")
    w.apply_story_deltas(deltas, allowed_kinds=NPC_ENTITY_KINDS)
    kf = project_known_facts(w)
    assert any(f["label"] == "通訊設備就在B2機房" for f in kf)       # 保留自然語意 label
    f = [x for x in kf if x["label"] == "通訊設備就在B2機房"][0]
    assert f["source"] == "醫生" and f["confidence"] == "npc_claim"   # source/confidence 保留


def test_known_facts_do_not_push_reveal():
    loop = _mk_loop()
    led = RevealLedger()
    led.get_or_create("t1", title="x")                    # hidden
    loop._reveal_ledger = led
    before = led.counts()
    # 登記一個 NPC fact entity（得知主張）——不得動 reveal ledger
    loop._world.register(FACT, "通訊設備在B2機房", id="fact.comm",
                         props={"source": "醫生", "confidence": "npc_claim"}, origin="npc")
    assert project_known_facts(loop._world)                # known_facts 有它
    assert led.counts() == before                          # reveal ledger 完全不動


def test_reveal_reward_observed_adds_public_known_fact():
    loop = _mk_loop()
    led = RevealLedger()
    t = led.get_or_create("t1", title="被忽略的細節", content="名單上有某個名字")
    t.level = "hinted"
    t.strength = 0.3
    loop._reveal_ledger = led
    apply_reveal_reward(led, beat=1)                       # hinted → observed
    assert t.level == "observed"
    loop._materialize_public_facts()
    kf = project_known_facts(loop._world)
    pub = [f for f in kf if f["source"] == "reveal"]
    assert pub and pub[0]["label"] == "被忽略的細節" and pub[0]["state"] == "observed"
    assert any(d.get("source") == "reveal" for d in loop._known_fact_delta_this_beat)


def test_hidden_truth_raw_not_in_known_facts():
    loop = _mk_loop()
    led = RevealLedger()
    t = led.get_or_create("t1", title="核心真相", content="你也是實驗體之一")
    t.level = "observed"
    t.strength = 0.7
    loop._reveal_ledger = led
    loop._materialize_public_facts()
    kf = project_known_facts(loop._world)
    blob = repr(kf)
    assert "你也是實驗體之一" not in blob                   # hidden raw content 永不進 known_facts
    assert any("核心真相" == f["label"] for f in kf)        # 只露 public title


def test_taken_object_removed_from_visible_entities():
    w = WorldModel()
    w.set_current_area("area.hall", label="大廳")
    w.register(OBJECT, "徽章", id="object.badge", state="noticed")
    w.tag_entity_area("object.badge", "area.hall")
    # 未 taken → visible
    proj = build_spatial_projection(w)
    assert any(v.id == "object.badge" for v in proj.visible_entities)
    # taken 且非 focus → 不在 visible
    w.take("object.badge")
    proj2 = build_spatial_projection(w)
    assert not any(v.id == "object.badge" for v in proj2.visible_entities)
    # taken 但正是 focus → 仍 visible（剛撿起/正在端詳）
    proj3 = build_spatial_projection(w, focus_id="object.badge")
    assert any(v.id == "object.badge" for v in proj3.visible_entities)


def test_review_mode_can_show_inventory_and_known_facts_without_new_fact():
    loop = _mk_loop()
    loop.known_npcs = ["阿明"]
    # 先在非 review 下取得物件 + 既有 NPC fact
    loop._world.register(OBJECT, "鑰卡", id="object.keycard", state="noticed")
    loop._world.take("object.keycard")
    loop._world.register(FACT, "電梯在東側", id="fact.lift",
                         props={"source": "醫生", "confidence": "npc_claim"}, origin="npc")
    facts_before = len(loop._world.by_kind(FACT))
    actors_before = len(loop._world.by_kind(ACTOR))
    # review-locked tick：敘事提到新角色「阿明」，但 review 不得新增 entity（extractor 被跳過）
    loop._world_model_tick("我停下來整理手上的東西", "走廊另一頭站著阿明", None, review_locked=True)
    assert len(loop._world.by_kind(FACT)) == facts_before    # 無新 fact
    assert len(loop._world.by_kind(ACTOR)) == actors_before  # 無新 actor
    # 但既有 inventory / known_facts 仍可投影
    assert any(i["id"] == "object.keycard" for i in project_inventory(loop._world))
    assert any(f["id"] == "fact.lift" for f in project_known_facts(loop._world))
