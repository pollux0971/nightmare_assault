"""NR1 — NPCChat 敘事控制驗收測試（敘事控制 v0.2）。

驗收：npc-chat 收同一敘事契約；無法引入未授權新 lore（gate flag / repair）；
forbidden 母題被擋；有用 hint → evidence_event 轉換並夾到 reveal_ceiling；
context 注入控制欄位但仍無 real_bible / secret_core。
"""
from __future__ import annotations

from core.narrative.npc_chat_control import (
    NPCChatControlContext, NPCChatResponse, NPCChatControlGate,
    build_control_context, response_to_evidence, cap_evidence_to_ceiling,
    REPAIR_INSTRUCTION,
)
from core.narrative.models import NarrativeContract, ProtagonistMotive, MotifPalette
from core.agents.npc_chat import build_npc_chat_context


class _BB:
    def __init__(self, snap):
        self._snap = snap
    def snapshot(self):
        return self._snap


def _contract():
    return NarrativeContract(
        core_premise="（內部前提，不外洩）",
        protagonist_motive=ProtagonistMotive(
            personal_loss="弟弟林晨失蹤", immediate_goal="找到林晨並活著離開",
            emotional_stake="不能再失去他", first_proof="林晨的紙條"),
        central_question="這裡發生過什麼？",
        motif_palette=MotifPalette(primary=["頻率", "鏽蝕"], secondary=["靜電"],
                                   forbidden_or_limited=["資料流", "菌絲"]))


# ── gate：未授權新 lore + 答債未付（reference 對齊）─────────────────────────
def test_gate_flags_unapproved_lore_and_answer_debt():
    gate = NPCChatControlGate()
    ctx = NPCChatControlContext(allowed_terms={"432.7"}, forbidden_terms={"資料流"},
                                answer_debt_level=2)
    resp = NPCChatResponse(visible_reply="...", answer_status="evaded",
                           new_lore_terms=["資料流"])
    flags = gate.validate(ctx, resp)
    assert any("forbidden_terms" in f for f in flags)
    assert any("illegal_new_lore_terms" in f for f in flags)
    assert "answer_debt_not_paid" in flags
    assert gate.needs_repair(flags) and REPAIR_INSTRUCTION


# ── gate：合規回答不 flag ────────────────────────────────────────────────────
def test_gate_passes_compliant_response():
    gate = NPCChatControlGate()
    ctx = NPCChatControlContext(allowed_terms={"頻率", "432.7"}, answer_debt_level=2)
    resp = NPCChatResponse(visible_reply="它是校準值。", answer_status="partial",
                           new_lore_terms=[], used_motifs=["頻率"])
    assert gate.validate(ctx, resp) == []


# ── 從 NarrativeContract 組控制 context ─────────────────────────────────────
def test_build_control_context_from_contract():
    ctx = build_control_context(_contract(), reveal_ceiling="observed", answer_debt_level=1)
    assert "頻率" in ctx.allowed_terms and "靜電" in ctx.allowed_terms
    assert "資料流" in ctx.forbidden_terms
    assert ctx.reveal_ceiling == "observed"
    assert "找到林晨" in ctx.active_motive


# ── evidence 轉換 + 夾到 reveal_ceiling（防跳級）──────────────────────────────
def test_evidence_capped_to_ceiling():
    resp = NPCChatResponse(visible_reply="…", evidence_events=[
        {"truth_id": "truth.sig", "max_level": "confirmed",
         "surface_text": "他承認頻率不對勁"}])
    cap_evidence_to_ceiling(resp, "hinted")
    evs = response_to_evidence(resp, beat=4)
    assert evs[0].source == "npc_chat"
    assert evs[0].max_level == "hinted"     # 上限被夾，不跳到 confirmed


# ── context 注入控制欄位，但無 real_bible / secret_core ─────────────────────
def test_context_has_control_but_no_secrets():
    snap = {
        "npc_registry": [{"name": "吳靜", "profession": "醫師", "public_face": "冷靜",
                          "secret_core": "她其實知道真相SECRET", "self_aware": True,
                          "evolving": {"emotional_state": "緊張"}}],
        "revealed_bible": {"known_atmosphere": ["潮濕"]},
        "real_bible": {"world_truth": {"what_really_happened": "TOP SECRET"}},
        "beat_window": [], "rolling_summary": "",
    }
    cctx = build_control_context(_contract(), reveal_ceiling="hinted")
    ctx = build_npc_chat_context(_BB(snap), "吳靜", "432.7 是什麼？", control_ctx=cctx)
    blob = str(ctx)
    assert "narrative_control" in ctx
    assert ctx["narrative_control"]["reveal_ceiling"] == "hinted"
    assert "SECRET" not in blob and "TOP SECRET" not in blob      # 防暴雷不變
    assert "secret_core" not in ctx["cognition_card"]


# ── 不給 control_ctx 時行為與 MC1 一致（向後相容）────────────────────────────
def test_backward_compatible_without_control():
    snap = {"npc_registry": [{"name": "甲", "profession": "警衛"}],
            "revealed_bible": {}, "beat_window": [], "rolling_summary": ""}
    ctx = build_npc_chat_context(_BB(snap), "甲", "你好")
    assert "narrative_control" not in ctx
    assert ctx["cognition_card"]["name"] == "甲"
