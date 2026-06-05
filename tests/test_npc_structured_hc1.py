"""HC1 — NPCChat Structured Gate 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：合法結構化回覆通過；違規 new lore → repair once；repair 仍違規 → safe fallback；
JSON 解析失敗 → 純文字 fallback 不崩；evidence 經 bridge 推進 ledger。
"""
from __future__ import annotations

import json

import core.constants as C
from core.agents.npc_chat import run_npc_chat_structured, _parse_npc_response
from core.narrative.npc_chat_control import NPCChatResponse


class _BB:
    def __init__(self, contract=True):
        nc = {"motif_palette": {"primary": ["頻率", "鏽蝕"], "forbidden_or_limited": ["資料流"]},
              "protagonist_motive": {"immediate_goal": "找到林晨"}}
        self.game_meta = {"narrative_contract": nc} if contract else {}
        self._snap = {"npc_registry": [{"name": "吳靜", "profession": "醫師", "public_face": "冷靜",
                                        "secret_core": "SECRET", "evolving": {}}],
                      "revealed_bible": {"known_atmosphere": []}, "beat_window": [], "rolling_summary": ""}
    def snapshot(self):
        return self._snap


class _Caller:
    """回傳預先排好的字串序列（模擬 LLM 多次呼叫：原始 → repair）。"""
    def __init__(self, replies):
        self._replies = list(replies); self.calls = 0
    def call(self, agent, ctx, output_model=None, temperature=None):
        r = self._replies[min(self.calls, len(self._replies) - 1)]
        self.calls += 1
        return r


# ── 解析：JSON → 結構化；非 JSON → 純文字 ───────────────────────────────────
def test_parse_structured_and_text():
    r = _parse_npc_response('{"reply":"它是校準值。","answer_status":"partial","new_lore_terms":[]}')
    assert r.visible_reply == "它是校準值。" and r.answer_status == "partial"
    r2 = _parse_npc_response("我不太確定…")           # 純文字
    assert r2.visible_reply == "我不太確定…"


# ── 合法結構化回覆通過（flag on）─────────────────────────────────────────────
def test_valid_structured_passes(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    caller = _Caller(['{"reply":"它和頻率有關。","answer_status":"partial","new_lore_terms":[]}'])
    resp = run_npc_chat_structured(caller, _BB(), "吳靜", "432.7 是什麼？")
    assert "頻率" in resp.visible_reply
    assert caller.calls == 1                          # 沒觸發 repair


# ── 違規 new lore → repair once → 修好 ──────────────────────────────────────
def test_illegal_lore_repaired(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    caller = _Caller([
        '{"reply":"那是資料流的核心。","answer_status":"partial","new_lore_terms":["資料流"]}',
        '{"reply":"我只能說它和頻率有關。","answer_status":"partial","new_lore_terms":[]}'])
    resp = run_npc_chat_structured(caller, _BB(), "吳靜", "432.7 是什麼？")
    assert caller.calls == 2                          # repair 一次
    assert "資料流" not in resp.visible_reply


# ── repair 仍違規 → safe fallback ───────────────────────────────────────────
def test_repair_fails_safe_fallback(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    bad = '{"reply":"資料流。","answer_status":"none","new_lore_terms":["資料流"]}'
    caller = _Caller([bad, bad])
    resp = run_npc_chat_structured(caller, _BB(), "吳靜", "432.7 是什麼？")
    assert resp.answer_status == "actionable"         # safe fallback
    assert "資料流" not in resp.visible_reply
    assert "值班紀錄" in resp.visible_reply or "驗證" in resp.visible_reply


# ── flag OFF → 純文字包裝、不崩、不過 gate ──────────────────────────────────
def test_flag_off_plain_text(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", False)
    caller = _Caller(["就是個普通的數字吧。"])
    resp = run_npc_chat_structured(caller, _BB(contract=False), "吳靜", "432.7?")
    assert resp.visible_reply == "就是個普通的數字吧。"
    assert caller.calls == 1


# ── #10：NPC 結構化 evidence_events 走 bridge 推進 ledger（真實 truth_id 才算）─
def test_npc_structured_evidence_advances_ledger(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    real_tid = list(loop._reveal_ledger.truths)[0]    # 一個真實 revelation_pool 真相
    before = loop._reveal_ledger.level_of(real_tid)
    # 結構化 evidence_events（含真實 truth_id）→ 走 bridge → 推進
    resp = NPCChatResponse(visible_reply="它和頻率有關。", answer_status="partial",
                           evidence_events=[{"truth_id": real_tid, "max_level": "observed",
                                             "evidence_strength": 0.4}])
    updates = loop.bridge_npc_evidence(resp)
    assert updates                                    # 結構化路徑有 grant
    assert loop._reveal_ledger.level_of(real_tid) != "hidden"


def test_npc_invented_truth_id_rejected(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    n_before = len(loop._reveal_ledger.truths)
    # NPC 捏造一個不存在的 truth_id → 不得污染 ledger
    resp = NPCChatResponse(visible_reply="...", evidence_events=[
        {"truth_id": "truth.INVENTED_BY_NPC", "max_level": "observed", "evidence_strength": 0.5}])
    assert loop.bridge_npc_evidence(resp) == []
    assert len(loop._reveal_ledger.truths) == n_before    # 沒有新增假真相


def test_npc_keyword_scan_does_not_grant_by_default(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    real_tid = list(loop._reveal_ledger.truths)[0]
    loop._truth_index = {real_tid: ["頻率"]}
    # 純文字（無結構化 evidence_events）+ 命中關鍵詞 → 預設**不** grant
    before = loop._reveal_ledger.level_of(real_tid)
    resp = NPCChatResponse(visible_reply="頻率很重要。", answer_status="partial")
    assert loop.bridge_npc_evidence(resp) == []
    assert loop._reveal_ledger.level_of(real_tid) == before
    # 明確開 legacy fallback 才 grant
    loop.bridge_npc_evidence(resp, allow_keyword_fallback=True)
    assert loop._reveal_ledger.level_of(real_tid) != "hidden"
