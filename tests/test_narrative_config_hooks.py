"""NC7 — Config Center Hooks 驗收測試。

驗收：opening_budget/reveal_policy/forbidden_motifs/element_limit/ending_thresholds 可由 config 讀寫；
safe profile 為預設；未設時走安全值。
"""
from __future__ import annotations

from core.narrative.config import get_narrative_config, apply_to_budget, SAFE_DEFAULTS
from core.narrative.models import OpeningBudget
from core.narrative.contract import build_narrative_contract
from core.persistence.db import Database
from core.blackboard import Blackboard


# ── safe 預設（無 store）─────────────────────────────────────────────────────
def test_safe_defaults_when_no_store():
    cfg = get_narrative_config(None)
    assert cfg["opening_max_named_objects"] == 3
    assert cfg["opening_reveal_limit"] == "hinted"
    assert cfg["element_limit"] == 1
    assert cfg["ending_clean_requires_truth"] is True
    assert "菌絲" in cfg["forbidden_motifs"]


# ── config 中心數值覆寫 ──────────────────────────────────────────────────────
def test_config_store_overrides_numbers():
    store = Database(":memory:").config_store()
    store.set_flag("NC_OPENING_MAX_OBJECTS", 2)
    store.set_flag("NC_ELEMENT_LIMIT", 1)
    cfg = get_narrative_config(store)
    assert cfg["opening_max_named_objects"] == 2           # 由 config 讀
    assert cfg["opening_max_chars"] == SAFE_DEFAULTS["opening_max_chars"]  # 未設 → 安全值


# ── overrides（UI/JSON）覆寫列表/策略 ───────────────────────────────────────
def test_overrides_lists_and_policy():
    cfg = get_narrative_config(None, overrides={"forbidden_motifs": ["X母題"],
                                                "opening_reveal_limit": "observed"})
    assert cfg["forbidden_motifs"] == ["X母題"]
    assert cfg["opening_reveal_limit"] == "observed"


# ── 套到 budget + 進 contract ───────────────────────────────────────────────
def test_config_applies_to_contract_budget():
    b = OpeningBudget()
    apply_to_budget(b, {"opening_max_named_objects": 2, "opening_max_chars": 700})
    assert b.max_named_objects == 2 and b.max_opening_chars == 700

    bb = Blackboard()
    bb.write("setup", "real_bible", {"world_truth": {"what_really_happened": "X",
                                                     "the_threat_is": "Y", "deadly_rule": "Z"}})
    bb.write("setup", "protagonist", {"name": "周凱", "starting_situation": "弟弟失蹤"})
    nc = build_narrative_contract(bb, config={"opening_max_named_objects": 2,
                                              "opening_reveal_limit": "observed",
                                              "forbidden_motifs": ["X母題"]})
    assert nc.opening_budget.max_named_objects == 2
    assert nc.opening_reveal_limit == "observed"
    assert nc.motif_palette.forbidden_or_limited == ["X母題"]
