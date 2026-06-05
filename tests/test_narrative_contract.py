"""NC1 — Narrative Contract 驗收測試。

驗收：setup 輸出有明確 protagonist_motive 與 central_question；opening_budget 可被後續模組讀取；
flag OFF 不動既有行為（contract 不建）。
"""
from __future__ import annotations

import core.constants
from core.narrative.contract import build_narrative_contract, store_contract
from core.blackboard import Blackboard
from core.orchestrator_loop import BeatLoop
from core.persistence.db import Database


def _bb():
    bb = Blackboard()
    bb.write("setup", "real_bible", {
        "world_truth": {"what_really_happened": "記憶置換實驗害死十一人",
                        "the_threat_is": "仍在運轉的機器製造幻覺", "deadly_rule": "勿在聲波區久留"},
        "atmosphere": ["潮濕", "鏽蝕", "低頻嗡鳴"],
    })
    bb.write("setup", "protagonist", {"name": "周凱", "starting_situation": "弟弟林晨失蹤三週"})
    return bb


def test_contract_has_motive_and_question():
    nc = build_narrative_contract(_bb())
    m = nc.protagonist_motive
    assert m.personal_loss and m.immediate_goal and m.emotional_stake and m.first_proof
    assert "林晨" in m.personal_loss or "失蹤" in m.personal_loss
    assert nc.central_question                              # 明確核心問題
    assert nc.opening_reveal_limit == "hinted"             # 開場最多 hinted


def test_opening_budget_readable():
    nc = build_narrative_contract(_bb())
    b = nc.opening_budget
    assert b.max_named_objects == 3 and b.max_new_lore_terms == 3 and b.max_opening_chars == 900


def test_store_contract_into_game_meta():
    bb = _bb()
    store_contract(bb, build_narrative_contract(bb))
    meta = bb.snapshot()["game_meta"]
    assert "narrative_contract" in meta
    assert meta["narrative_contract"]["central_question"]


# ── 旁路 flag：ON 建、OFF 不建 ──────────────────────────────────────────────
def _loop():
    return BeatLoop(caller=None, blackboard=_bb(), db=Database(), run_id="t", use_kernel=False)


def test_flag_on_builds_contract(monkeypatch):
    monkeypatch.setattr(core.constants, "ENABLE_NARRATIVE_CONTROL", True)
    loop = _loop()
    loop._build_narrative_contract()
    assert loop._narrative_contract is not None
    assert "narrative_contract" in loop.bb.snapshot()["game_meta"]


def test_flag_off_skips_contract(monkeypatch):
    monkeypatch.setattr(core.constants, "ENABLE_NARRATIVE_CONTROL", False)
    loop = _loop()
    loop._build_narrative_contract()
    assert loop._narrative_contract is None
    assert "narrative_contract" not in (loop.bb.snapshot()["game_meta"] or {})
