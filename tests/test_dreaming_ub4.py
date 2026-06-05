"""UB4 — 輕量 dreaming（在場 NPC）驗收測試。

驗收：在場 NPC 情緒演化、非同步只產 patch、沒戲份凍結（C7）、self_aware=false 不編謊（C5）、
只寫 evolving 碰不到 secret_core（C6）。
"""
from __future__ import annotations

import pytest

from core.agents.dreaming import run_dreaming, _merge_evolving
from core.blackboard import Blackboard
from core.models import DreamingOutput


def _npc(name, present=True, self_aware=True):
    return {
        "name": name, "profession": "醫生", "personality": "mysterious",
        "voice_sample": "…", "public_face": "冷靜", "secret_core": f"{name}的秘密SECRET",
        "self_aware": self_aware, "presence": "present" if present else "absent",
        "evolving": {"emotional_state": {}, "relationship": {}, "intent": "observe",
                     "revealed_layers": [], "emergent_lies": [], "personal_arc": ""},
    }


class FakeCaller:
    """每個 NPC 回固定 DreamingOutput（present test 用）。"""
    def __init__(self, out): self._out = out; self.calls = []
    def call(self, agent, context, output_model=None, temperature=None):
        self.calls.append(context)
        return self._out


def _bb(npcs):
    bb = Blackboard()
    bb.write("setup", "npc_registry", npcs)
    bb.version = 1
    return bb


OUT = DreamingOutput(emotional_update={"fear": 0.7}, relationship_update={"trust": 0.3},
                     intent_update="betray", emergent_lie="我整晚都在值班（謊）",
                     personal_arc_note="開始懷疑主角")


# ── 在場 NPC 演化 + 非同步只產 patch ─────────────────────────────────────────
def test_present_npc_evolves_via_patch():
    bb = _bb([_npc("張醫生", present=True, self_aware=True)])
    caller = FakeCaller(OUT)
    res = run_dreaming(caller, bb, beat_number=5)
    assert res and res[0][0] == "張醫生"
    # 非同步：merge 前快照看不到變化（只在 pending）
    before = bb.snapshot()["npc_registry"][0]["evolving"]
    assert before["intent"] == "observe" and not before["emotional_state"]
    assert len(bb.collect_pending()) == 1
    # 安全點 merge 後才生效
    bb.merge_and_bump()
    ev = bb.snapshot()["npc_registry"][0]["evolving"]
    assert ev["intent"] == "betray" and ev["emotional_state"]["fear"] == 0.7
    assert ev["relationship"]["trust"] == 0.3


# ── C7：沒戲份（不在場）凍結 ─────────────────────────────────────────────────
def test_absent_npc_frozen():
    bb = _bb([_npc("離場者", present=False, self_aware=True)])
    res = run_dreaming(FakeCaller(OUT), bb, beat_number=5)
    assert res == []                                 # 不在場 → 完全不跑
    assert bb.collect_pending() == []


# ── C5：self_aware=false 不編謊 ──────────────────────────────────────────────
def test_not_self_aware_no_emergent_lie():
    bb = _bb([_npc("誠實者", present=True, self_aware=False)])
    run_dreaming(FakeCaller(OUT), bb, beat_number=5)
    bb.merge_and_bump()
    ev = bb.snapshot()["npc_registry"][0]["evolving"]
    assert ev["emergent_lies"] == []                 # 誠實型不編謊（即使 LLM 給了 lie）
    # 但情緒仍會演化
    assert ev["emotional_state"]["fear"] == 0.7


def test_self_aware_can_lie():
    bb = _bb([_npc("說謊者", present=True, self_aware=True)])
    run_dreaming(FakeCaller(OUT), bb, beat_number=5)
    bb.merge_and_bump()
    assert "我整晚都在值班（謊）" in bb.snapshot()["npc_registry"][0]["evolving"]["emergent_lies"]


# ── 頻率：非第 5 beat 不跑 ───────────────────────────────────────────────────
def test_only_runs_every_5_beats():
    bb = _bb([_npc("A")])
    assert run_dreaming(FakeCaller(OUT), bb, beat_number=3) == []
    assert run_dreaming(FakeCaller(OUT), bb, beat_number=10) != []


# ── C6：dreaming 碰不到 secret_core（權限邊界）──────────────────────────────
def test_dreaming_cannot_write_secret_core():
    bb = _bb([_npc("張醫生")])
    bb.submit_patch({"base_version": 1, "writer": "dreaming",
                     "target": "npc_registry.張醫生.secret_core", "value": "竄改"})
    with pytest.raises(PermissionError):
        bb.merge_and_bump()


def test_caller_none_noop():
    bb = _bb([_npc("A")])
    assert run_dreaming(None, bb, beat_number=5) == []
