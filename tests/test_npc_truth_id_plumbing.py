"""NPC Evidence Truth-ID Plumbing — 讓 NPC-chat 能在白名單內安全推進 reveal，
不回到 keyword-scan 誤判。

主線：
- NPC 只能引用 orchestrator 提供的 allowed_truth_refs 裡的 truth_id。
- 引用不在白名單（含 core / 超 ceiling）→ reject、不污染 ledger。
- evidence 無 truth_id → 只記 conversation note、不 grant。
- NPC 迴避（evasion）不重置答債。
"""
from __future__ import annotations

import core.constants as C
from core.narrative.npc_chat_control import (
    build_allowed_truth_refs, validate_npc_evidence, NPCChatResponse,
)
from core.narrative.revelation import (
    build_ledger_from_bible, RevelationBridge, EvidenceEvent,
)
from core.narrative.answer_debt import AnswerDebtTracker


def _bible():
    return {"revelation_pool": [
        {"id": "truth.signal_frequency", "title": "432.7 的意義", "content": "校準值"},
        {"id": "truth.missing_person", "title": "林晨的下落", "content": "他沒離開"},
        {"id": "truth.true_culprit", "title": "核心真相", "content": "你也是實驗體"}]}  # 最後 = core


# ── 白名單建構：core 與超 ceiling 被排除 ────────────────────────────────────
def test_build_allowed_truth_refs_excludes_core():
    led = build_ledger_from_bible(_bible())
    block = build_allowed_truth_refs(led, "observed",
                                     core_truth_ids={"truth.true_culprit"})
    allowed_ids = {r["truth_id"] for r in block["allowed_truth_refs"]}
    assert "truth.signal_frequency" in allowed_ids
    assert "truth.missing_person" in allowed_ids
    assert "truth.true_culprit" not in allowed_ids          # core → forbidden
    assert "truth.true_culprit" in block["forbidden_truth_refs"]
    # 每個 ref 最多到 ceiling/NPC 硬上限，且帶 safe_hint（無完整真相內容）
    for r in block["allowed_truth_refs"]:
        assert r["max_level"] in ("hinted", "observed")
        assert "校準值" not in r["safe_hint"] and "實驗體" not in r["safe_hint"]


# ── 1. 接受白名單內 truth_id ────────────────────────────────────────────────
def test_npc_evidence_accepts_allowed_truth_id():
    refs = [{"truth_id": "truth.signal_frequency", "max_level": "hinted"}]
    ev = EvidenceEvent("e", "npc_chat", "truth.signal_frequency", max_level="hinted",
                       evidence_strength=0.45)
    accepted, rejected = validate_npc_evidence([ev], refs)
    assert accepted and not rejected
    assert accepted[0].truth_id == "truth.signal_frequency"


# ── 2. 拒絕不在白名單的 truth_id ────────────────────────────────────────────
def test_npc_evidence_rejects_unallowed_truth_id():
    refs = [{"truth_id": "truth.signal_frequency", "max_level": "hinted"}]
    ev = EvidenceEvent("e", "npc_chat", "truth.true_culprit", max_level="observed")
    accepted, rejected = validate_npc_evidence([ev], refs)
    assert not accepted
    assert rejected[0][1] == "truth_id_not_allowed"


# ── 3. 無 truth_id → 不 grant（只能當 conversation note）─────────────────────
def test_npc_evidence_without_truth_id_does_not_grant():
    refs = [{"truth_id": "truth.signal_frequency", "max_level": "hinted"}]
    ev = EvidenceEvent("e", "npc_chat", None, surface_text="我聽過牆裡有聲音。")
    accepted, rejected = validate_npc_evidence([ev], refs)
    assert not accepted
    assert rejected[0][1] == "no_truth_id"


# ── level 超過 ref max_level → 夾下來（尊重 ceiling）─────────────────────────
def test_npc_evidence_level_clamped_to_ref():
    refs = [{"truth_id": "truth.signal_frequency", "max_level": "hinted"}]
    ev = EvidenceEvent("e", "npc_chat", "truth.signal_frequency", max_level="confirmed")
    accepted, _ = validate_npc_evidence([ev], refs)
    assert accepted[0].max_level == "hinted"                 # 被夾到 hinted


# ── 4. NPC 迴避不重置答債 ───────────────────────────────────────────────────
def test_npc_answer_status_evasion_does_not_reset_answer_debt():
    t = AnswerDebtTracker()
    k = "mechanism_question:signal"
    t.register_question(k); t.register_question(k)          # debt=2
    t.register_answer(k, "evasion")
    assert t.level(k) == 2                                   # 迴避 → 不重置
    t.register_answer(k, "none")
    assert t.level(k) == 2                                   # none 亦不重置
    t.register_answer(k, "partial")
    assert t.level(k) == 0                                   # 部分答 → 償還


# ── 端到端：loop 提供白名單，gate 接受合法、拒絕非法、不污染 ────────────────
def test_loop_plumbing_end_to_end(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    block = (loop.bb.game_meta or {}).get("npc_allowed_truth_refs") or {}
    allowed = block.get("allowed_truth_refs", [])
    assert allowed                                           # 開局就有白名單
    legal_id = allowed[0]["truth_id"]
    core_id = list(loop._core_truth_ids)[0]
    n_before = len(loop._reveal_ledger.truths)

    # 合法引用 → grant
    ok = NPCChatResponse(visible_reply="它和頻率有關。", answer_status="partial",
                         evidence_events=[{"truth_id": legal_id, "max_level": "hinted",
                                           "evidence_strength": 0.4}])
    assert loop.bridge_npc_evidence(ok)
    assert loop._reveal_ledger.level_of(legal_id) != "hidden"
    assert loop._npc_chat_debug["evidence_events_accepted"] == 1

    # 引用 core（不在白名單）→ reject、不污染
    bad = NPCChatResponse(visible_reply="真相是…", answer_status="partial",
                          evidence_events=[{"truth_id": core_id, "max_level": "observed",
                                            "evidence_strength": 0.6}])
    assert loop.bridge_npc_evidence(bad) == []
    assert loop._npc_chat_debug["evidence_events_rejected"] == 1
    assert "truth_id_not_allowed" in loop._npc_chat_debug["rejection_reasons"]

    # 捏造 truth_id → reject、ledger 不長新真相
    fake = NPCChatResponse(visible_reply="...", answer_status="partial",
                           evidence_events=[{"truth_id": "truth.MADE_UP", "max_level": "hinted",
                                             "evidence_strength": 0.5}])
    assert loop.bridge_npc_evidence(fake) == []
    assert len(loop._reveal_ledger.truths) == n_before
