"""MC4 — Offstage-Fate agent 驗收測試。

驗收：roll_fate 受 alignment 影響、機率可控；四種命運各正確寫 presence/alignment；carried_fragment；
只寫 npc/scene、碰不到 secret_core（C6）；非同步只產 patch。
"""
from __future__ import annotations

import random

import pytest

from core.agents.offstage_fate import (
    roll_fate, run_offstage_fate, offstage_fate_tick, FATE_TYPES, _state_patches,
)
from core.blackboard import Blackboard
from core.models import OffstageFateOutput


def _npc(name="警衛", alignment="neutral", presence="absent"):
    return {"name": name, "profession": "警衛", "personality": "nervous", "voice_sample": "…",
            "public_face": "緊張", "secret_core": f"{name}的秘密SECRET", "self_aware": False,
            "presence": presence, "alignment": alignment, "evolving": {}}


def _bb(npcs):
    bb = Blackboard()
    bb.write("setup", "npc_registry", npcs)
    bb.write("setup", "real_bible", {"revelation_pool": [{"id": "f1", "content": "地下室名單"}]})
    bb.write("orchestrator", "revealed_bible", {"revealed_fragments": []})
    return bb


# ── roll_fate 受 alignment 影響 ─────────────────────────────────────────────
def test_roll_fate_weighted_by_alignment():
    rng = random.Random(42)
    allied = [roll_fate(_npc(alignment="allied"), rng) for _ in range(200)]
    hostile = [roll_fate(_npc(alignment="hostile"), rng) for _ in range(200)]
    assert allied.count("opportunity_return") > allied.count("hostile_return")
    assert hostile.count("hostile_return") > hostile.count("opportunity_return")
    assert all(f in FATE_TYPES for f in allied + hostile)


# ── 四種命運各正確寫狀態 ─────────────────────────────────────────────────────
class FateCaller:
    def __init__(self, narrative="他在三樓找到了這個。"):
        self._n = narrative
    def call(self, agent, context, output_model=None, temperature=None):
        assert agent == "offstage-fate"
        return OffstageFateOutput(fate_type=context["fate_type"], fate_narrative=self._n,
                                  fragment_delivery="碎片包裝", scene_seed={"type": "corpse"})


@pytest.mark.parametrize("fate,expect", [
    ("opportunity_return", ("present", "allied")),
    ("missing", ("missing", None)),
    ("corpse", ("dead", "dead")),
    ("hostile_return", ("present", "hostile")),
])
def test_each_fate_writes_state(fate, expect):
    bb = _bb([_npc()])
    out = run_offstage_fate(FateCaller(), bb, "警衛", carried_fragment="f1", fate_type=fate)
    assert out.fate_type == fate                         # 強制與程式碼一致
    bb.merge_and_bump()                                  # 非同步：安全點才生效
    npc = bb.snapshot()["npc_registry"][0]
    assert npc["presence"] == expect[0]
    if expect[1]:
        assert npc["alignment"] == expect[1]
    assert npc["carried_fragment"] == "f1"               # 領到碎片


# ── 非同步只產 patch（merge 前不可見）────────────────────────────────────────
def test_async_patch_only():
    bb = _bb([_npc()])
    run_offstage_fate(FateCaller(), bb, "警衛", carried_fragment="f1", fate_type="missing")
    assert bb.snapshot()["npc_registry"][0]["presence"] == "absent"   # merge 前不變
    assert bb.collect_pending()                                       # 在 pending
    bb.merge_and_bump()
    assert bb.snapshot()["npc_registry"][0]["presence"] == "missing"


# ── C6：offstage_fate 碰不到 secret_core ─────────────────────────────────────
def test_cannot_write_secret_core():
    bb = _bb([_npc()])
    bb.submit_patch({"base_version": 0, "writer": "offstage_fate",
                     "target": "npc_registry.警衛.secret_core", "value": "竄改"})
    with pytest.raises(PermissionError):
        bb.merge_and_bump()


# ── tick：只跑離場 NPC ───────────────────────────────────────────────────────
def test_tick_only_offstage_npcs():
    bb = _bb([_npc("在場者", presence="present"), _npc("離場者", presence="absent")])
    res = offstage_fate_tick(FateCaller(), bb, beat_number=8)
    names = [r[0] for r in res]
    assert "離場者" in names and "在場者" not in names
    assert offstage_fate_tick(FateCaller(), bb, beat_number=3) == []   # 非 tick beat
