"""Fragment Title / Public Reveal Label（補丁）測試。

問題：reveal-derived known_facts 可能顯示「未命名的真相」→ 真相型玩家無回報感。
本補丁：每個真相走 public_title（explicit fragment title 或 slug fallback），
用於 known_facts / reveal recap / player_state_summary；**永不「未命名的真相」、永不 hidden content**。

不改 TruthEvidenceGate / Reveal Reward 升階規則 / WorldModel ownership / AliasResolver；不新增故事內容。
"""
from __future__ import annotations

from core.narrative.models import REVEAL_RANK
from core.narrative.revelation import (
    RevealLedger, apply_reveal_reward, public_recap, public_title,
    recap_from_ledger, reveal_public_facts)


# ── public_title 核心 ────────────────────────────────────────────────────────

def test_public_title_explicit_wins():
    assert public_title("frag_001_logbook", "第一道警示") == ("第一道警示", "explicit")


def test_public_title_fallback_never_unnamed():
    for tid in ["frag_001_logbook", "frag_007_xue_is_gone", "f1", "f99", "真相_核心", "clue_42"]:
        label, src = public_title(tid, "")
        assert src == "fallback"
        assert label.strip() and "未命名的真相" not in label


def test_public_title_fallback_human_readable():
    assert public_title("frag_001_logbook", "")[0] == "線索：Logbook"
    assert public_title("frag_003_infection_sign", "")[0] == "線索：Infection Sign"
    assert public_title("f7", "")[0] == "未解線索 #7"


# ── reveal_public_facts：observed/suspected 用 public_title、無 content ────────

def test_reveal_public_facts_observed_meaningful_no_content():
    led = RevealLedger()
    t = led.get_or_create("frag_003_infection_sign", title="", content="病毒透過通風口擴散")
    t.level = "observed"
    pf = reveal_public_facts(led)
    assert pf and pf[0]["title_source"] == "fallback"
    assert "未命名的真相" not in pf[0]["title"]
    assert "病毒透過通風口擴散" not in pf[0]["title"]      # hidden content 永不入 title


def test_reveal_public_facts_suspected_meaningful():
    led = RevealLedger()
    t = led.get_or_create("frag_010_reactor", title="")
    t.level = "suspected"
    assert "未命名的真相" not in reveal_public_facts(led)[0]["title"]


def test_reveal_public_facts_explicit_title_source():
    led = RevealLedger()
    t = led.get_or_create("t1", title="被忽略的細節")
    t.level = "observed"
    pf = reveal_public_facts(led)
    assert pf[0]["title"] == "被忽略的細節" and pf[0]["title_source"] == "explicit"


# ── recap 也用 public_title ──────────────────────────────────────────────────

def test_recap_uses_public_title_not_unnamed():
    led = RevealLedger()
    led.get_or_create("frag_002_door", title="").level = "confirmed"
    led.get_or_create("frag_004_voice", title="").level = "observed"
    pr = public_recap(led)
    rc = recap_from_ledger(led)
    assert all("未命名的真相" not in f["title"] for f in pr["found"])
    assert all("未命名的真相" not in f["title"]
               for f in rc["confirmed_list"] + rc["suspected_list"] + rc["hinted_list"])


# ── title patch 不放寬升階：confirmed 仍須強 evidence ─────────────────────────

def test_title_patch_does_not_relax_ladder():
    led = RevealLedger()
    t = led.get_or_create("frag_x", title="")
    t.level = "hinted"; t.strength = 0.3
    for b in range(10):
        apply_reveal_reward(led, beat=b)              # reward 仍上限 suspected
    assert REVEAL_RANK[t.level] <= REVEAL_RANK["suspected"]


# ── 迴圈：materialize 產生 public-safe known_fact + observation debug ─────────

def _mk_loop():
    from core.blackboard import Blackboard
    from core.orchestrator_loop import BeatLoop
    from core.persistence.db import Database
    from core.signal import SignalBus
    from core.world.model import WorldModel
    from tests.test_narrative_v2_integration_nr import FakeCaller
    loop = BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(), run_id="r", use_kernel=True)
    loop._world = WorldModel()
    loop._known_fact_delta_this_beat = []
    return loop


def test_materialize_public_fact_label_and_debug():
    from core.world.player_state import project_known_facts
    loop = _mk_loop()
    led = RevealLedger()
    t = led.get_or_create("frag_003_infection_sign", title="", content="病毒透過通風口擴散")
    t.level = "observed"
    levels_before = {tid: tr.level for tid, tr in led.truths.items()}
    loop._reveal_ledger = led
    loop._materialize_public_facts()
    # known_fact label = public_title（有意義、非 未命名、非 content）
    kf = project_known_facts(loop._world)
    pub = [f for f in kf if f.get("source") == "reveal"][0]
    assert "未命名的真相" not in pub["label"] and "病毒透過通風口擴散" not in pub["label"]
    # observation debug 欄位齊全
    d = [x for x in loop._known_fact_delta_this_beat if x.get("source") == "reveal"][0]
    assert d["truth_id"] == "frag_003_infection_sign"
    assert d["public_title"] == pub["label"]
    assert d["title_source"] == "fallback"
    assert d["hidden_content_exposed"] is False
    # materialize 唯讀：不改 ledger 等級（不推 reveal）
    assert {tid: tr.level for tid, tr in led.truths.items()} == levels_before


def test_materialize_explicit_title_label():
    from core.world.player_state import project_known_facts
    loop = _mk_loop()
    led = RevealLedger()
    led.get_or_create("t1", title="被忽略的細節").level = "observed"
    loop._reveal_ledger = led
    loop._materialize_public_facts()
    pub = [f for f in project_known_facts(loop._world) if f.get("source") == "reveal"][0]
    assert pub["label"] == "被忽略的細節"
    d = [x for x in loop._known_fact_delta_this_beat if x.get("source") == "reveal"][0]
    assert d["title_source"] == "explicit"
