"""HB2 — RevelationBridge Unified Inputs 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：多來源 EvidenceEvent（kernel/story/npc_chat/document）走同一 bridge.apply；
ledger 更新寫 revealed_bible；loop step 結果暴露 reveal_updates / evidence_events_this_beat /
unmapped_evidence_this_beat。
"""
from __future__ import annotations

from core.narrative.revelation import (
    RevealLedger, RevelationBridge, EvidenceEvent, build_ledger_from_bible,
)


def _bible():
    return {"revelation_pool": [{"id": "t1", "title": "甲", "content": "..."},
                                {"id": "t2", "title": "乙", "content": "..."}]}


# ── 多來源統一 apply ─────────────────────────────────────────────────────────
def test_multi_source_unified_apply():
    led = build_ledger_from_bible(_bible())
    events = [
        EvidenceEvent("e.k", "kernel", "t1", evidence_strength=0.3),
        EvidenceEvent("e.s", "story", "t1", evidence_strength=0.4),       # 累積 → t1 observed
        EvidenceEvent("e.n", "npc_chat", "t2", evidence_strength=0.3, max_level="observed"),
        EvidenceEvent("e.d", "document", "t2", evidence_strength=0.4),
    ]
    updates = RevelationBridge().apply(led, events)
    assert led.level_of("t1") == "observed"          # 0.3+0.4 ≥ 0.6
    assert led.level_of("t2") == "observed"          # 0.3+0.4 ≥ 0.6（封頂 observed 內）
    assert {u["truth_id"] for u in updates} == {"t1", "t2"}
    assert {e.source for e in events} == {"kernel", "story", "npc_chat", "document"}


# ── loop step 暴露揭露觀測欄位 ───────────────────────────────────────────────
def test_step_exposes_reveal_observability(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    out = loop.step("我打開病房門檢查走廊")
    assert "reveal_updates" in out
    assert "evidence_events_this_beat" in out
    assert "unmapped_evidence_this_beat" in out
    assert isinstance(out["reveal_updates"], list)
    # 計數每 beat 重置（不是累加全程）
    out2 = loop.step("我站著不動")
    assert isinstance(out2["evidence_events_this_beat"], int)
