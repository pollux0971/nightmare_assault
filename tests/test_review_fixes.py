"""Regression tests for code-review fixes (session-25 review)."""
from __future__ import annotations

import core.constants as C


# ── #2 真相關鍵詞索引：丟掉裸短數字與常見 2 字詞，保留具體詞 ─────────────────
def test_keyword_index_drops_noise_keeps_specific():
    from core.narrative.revelation import build_truth_keyword_index, evidence_from_npc_reply
    bible = {"revelation_pool": [
        {"id": "t.sig", "title": "432.7 的意義", "content": "聲納校準值，第3天開始的實驗"},
        {"id": "t.lin", "title": "林晨的下落", "content": "他在水下設備室"}]}
    idx = build_truth_keyword_index(bible)
    flat = [k for ks in idx.values() for k in ks]
    assert "432.7" in flat                       # 具體數字代號保留
    assert "3" not in flat                        # 裸短數字丟棄
    assert "實驗" not in flat                      # 常見 2 字詞丟棄（在停用表）
    # 一句只含「3」的無關回覆不該誤推 t.sig
    assert evidence_from_npc_reply("我在這等了3天。", idx, answer_status="partial") == []
    # 含具體關鍵詞才推進
    assert evidence_from_npc_reply("432.7 是校準值。", idx, answer_status="partial")


# ── #4 純文字 NPC 迴避不被當成 partial（不付答債、不洩證據）──────────────────
def test_freetext_evasion_not_partial():
    from core.agents.npc_chat import _parse_npc_response
    assert _parse_npc_response("我什麼都不知道。").answer_status == "evasion"
    assert _parse_npc_response("我不能說，別問了。").answer_status == "evasion"
    assert _parse_npc_response("它是校準值。").answer_status == "partial"
    assert _parse_npc_response("").answer_status == "none"


# ── #7 evidence_from_clue_values 不把合法 0.0 強度覆蓋成 0.4 ─────────────────
def test_zero_strength_not_clobbered():
    from core.narrative.revelation import evidence_from_clue_values
    clues = [("clue.x", {"truth_id": "t1", "evidence_strength": 0.0, "max_level": "hinted"})]
    events, _ = evidence_from_clue_values(clues)
    assert events[0].evidence_strength == 0.0       # 不被 `or 0.4` 蓋掉
    assert events[0].max_level == "hinted"


# ── #8 check_escape 相容 RevealLedger.counts() 的 'confirmed_or_better' 鍵 ───
def test_check_escape_counts_key_compat():
    from core.narrative.ending_causality import EndingCausalityGate, EndingCandidate
    g = EndingCausalityGate()
    cand = EndingCandidate(type="escape", source="attractor")
    # counts() 風格鍵
    r = g.check_escape(cand, escape_step="commit", reveal_progress={"confirmed_or_better": 2})
    assert r.allowed and r.downgrade_to is None     # 有 confirmed → 不降級
    r0 = g.check_escape(cand, escape_step="commit", reveal_progress={"confirmed_or_better": 0})
    assert r0.downgrade_to == "ambiguous_escape"


# ── #1 品質閘門：被否決的 beat 不串給玩家，只串最終接受版 ────────────────────
def test_quality_gate_streams_only_accepted(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    import core.orchestrator_loop as ol
    from core.models import DecisionPoint, Option

    # 假 run_story：第一次回含 forbidden 詞（會被 quality gate 否決），第二次回乾淨版
    calls = {"n": 0, "on_events": []}
    def _fake_run_story(caller, bb, pd, beat, context_override=None, on_event=None,
                        system_override=None):
        calls["n"] += 1
        calls["on_events"].append(on_event)
        bad = calls["n"] == 1
        text = "資料流的核心。" if bad else "走廊很安靜。"
        dp = DecisionPoint(situation_recap=text, decision_type="action",
                           suggested_options=[Option(text="前進", tone="bold")],
                           free_input_hint="", beat_meta={"beat_number": beat})
        return text, dp
    monkeypatch.setattr(ol, "run_story", _fake_run_story)

    loop = ol.BeatLoop.__new__(ol.BeatLoop)
    loop.caller = None; loop.bb = None; loop.beat_number = 2
    loop._narrative_contract = object()
    loop._motif_tracker = None
    loop._story_system = lambda c: None

    streamed = []
    def _on_event(ev):
        streamed.append(getattr(ev, "text", ""))

    class _GS:
        beat_number = 2
    # ctx 帶 forbidden_new_elements 讓第一版被否決
    ctx = {"forbidden_new_elements": ["資料流"]}
    narrative, dp, meta = loop._run_story_with_repair("看一下", _GS(), ctx, _on_event)

    # 生成期間 run_story 收到的 on_event 必須是 None（靜默生成，不外洩被否決的 beat）
    assert all(oe is None for oe in calls["on_events"])
    # 串給玩家的只有最終接受版，不含被否決的「資料流」
    joined = "".join(streamed)
    assert "資料流" not in joined
    assert "走廊很安靜" in joined
    assert meta["repaired"] is True
