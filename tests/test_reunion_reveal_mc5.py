"""MC5 — 離場命運隱藏 + 重逢揭曉 驗收測試。

驗收：離場期間 story context 看不到命運敘事/碎片（隱藏）；NPC 重新 present → carried_fragment 進 revealed；
隱藏紀錄存讀一致；不主動劇透。
"""
from __future__ import annotations

import json

from core.agents.offstage_fate import (
    store_hidden_fate, reveal_carried_fragment, check_reunions,
)
from core.agents.story import build_story_context
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.models import OffstageFateOutput

FRAG_CONTENT = "地下室藏著一份死亡名單SECRET"
FATE_NARR = "警衛在地下室發現了名單，卻再也沒回來——他死在那裡。"


def _bb(presence="missing"):
    bb = Blackboard()
    bb.write("setup", "real_bible", {"revelation_pool": [{"id": "f1", "content": FRAG_CONTENT}]})
    bb.write("setup", "npc_registry", [
        {"name": "警衛", "profession": "警衛", "personality": "nervous", "voice_sample": "…",
         "public_face": "緊張", "secret_core": "x", "self_aware": False,
         "presence": presence, "alignment": "neutral", "carried_fragment": "f1"}])
    bb.write("orchestrator", "revealed_bible", {"revealed_fragments": []})
    return bb


def _out():
    return OffstageFateOutput(fate_type="corpse", fate_narrative=FATE_NARR, fragment_delivery="屍體上的線索")


# ── 隱藏紀錄存讀 ─────────────────────────────────────────────────────────────
def test_hidden_fate_stored():
    db = Database(":memory:"); db.create_run("r1")
    store_hidden_fate(db, "r1", "警衛", _out(), "f1")
    fate = db.get_offstage_fate("r1", "警衛")
    assert fate["fate_narrative"] == FATE_NARR and fate["revealed"] == 0


# ── 離場期間 story 看不到命運敘事/碎片 ──────────────────────────────────────
def test_hidden_from_story_during_absence():
    bb = _bb(presence="missing")
    db = Database(":memory:"); db.create_run("r1")
    store_hidden_fate(db, "r1", "警衛", _out(), "f1")
    ctx = build_story_context(bb, "我四處看看")
    blob = json.dumps(ctx, ensure_ascii=False, default=str)
    assert FRAG_CONTENT not in blob                     # 碎片內容不外洩
    assert FATE_NARR not in blob                         # 命運敘事不外洩
    assert "carried_fragment" not in blob                # 連欄位都不在 story 投影


# ── 重逢揭曉：present → 碎片進 revealed ──────────────────────────────────────
def test_reveal_on_reunion():
    bb = _bb(presence="present")                         # 機遇/敵對歸來 → present
    db = Database(":memory:"); db.create_run("r1")
    store_hidden_fate(db, "r1", "警衛", _out(), "f1")
    revealed = check_reunions(bb, db, "r1")
    assert revealed and revealed[0][0] == "警衛"
    rb = bb.snapshot()["revealed_bible"]["revealed_fragments"]
    assert any(f["content"] == FRAG_CONTENT for f in rb)  # 碎片過揭露
    assert db.get_offstage_fate("r1", "警衛")["revealed"] == 1
    # 已交付 → carried_fragment 清掉
    assert bb.snapshot()["npc_registry"][0]["carried_fragment"] is None


def test_no_reveal_while_absent():
    bb = _bb(presence="missing")                         # 仍離場 → 不揭曉
    db = Database(":memory:"); db.create_run("r1")
    store_hidden_fate(db, "r1", "警衛", _out(), "f1")
    assert check_reunions(bb, db, "r1") == []
    assert bb.snapshot()["revealed_bible"]["revealed_fragments"] == []


def test_reveal_idempotent():
    bb = _bb(presence="present")
    db = Database(":memory:"); db.create_run("r1")
    store_hidden_fate(db, "r1", "警衛", _out(), "f1")
    reveal_carried_fragment(bb, db, "r1", "警衛")
    assert reveal_carried_fragment(bb, db, "r1", "警衛") is None   # 已揭曉不重複
