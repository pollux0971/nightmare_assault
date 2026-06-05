"""NR0 — RevelationBridge 驗收測試（敘事控制 v0.2 · 重做版）。

對應使用者 6 點要求：
1. kernel clue 自帶 truth_id → evidence_from_clue_values 取得。
2. RevealManager 依 evidence_strength 累積決定 reveal_level（strength 驅動、可達 confirmed）。
3. revealed_bible 記 hidden→…→actionable（write_ledger_to_revealed_bible）。
4. recap 含 hinted/observed 分層明細（含標題）。
5. NPC-chat 有效資訊走同一 bridge（evidence_from_npc_reply）。
6. 檢查線索後 revealed_bible 不再 0。
"""
from __future__ import annotations

from core.narrative.revelation import (
    EvidenceEvent, RevealLedger, RevelationBridge, strength_to_level,
    build_ledger_from_bible, evidence_from_clue_values, evidence_from_npc_reply,
    build_truth_keyword_index, write_ledger_to_revealed_bible, recap_from_ledger,
)


def _bible():
    return {"revelation_pool": [
        {"id": "truth.signal", "title": "432.7 的意義", "content": "432.7 是聲納校準值"},
        {"id": "truth.linchen", "title": "林晨的下落", "content": "林晨沒有離開這裡"},
        {"id": "truth.core", "title": "核心真相", "content": "你也是實驗體之一"}]}


class _BB:
    def __init__(self, snap=None):
        self._snap = snap or {"revealed_bible": {"revealed_fragments": [], "truth_progress": {}}}
        self.writes = []
    def snapshot(self):
        return self._snap
    def write(self, writer, key, val):
        self._snap[key] = val
        self.writes.append((writer, key, val))


# ── #2：evidence_strength 累積驅動等級（可達 confirmed）─────────────────────
def test_strength_drives_level():
    assert strength_to_level(0.0) == "hidden"
    assert strength_to_level(0.3) == "hinted"
    assert strength_to_level(0.65) == "observed"
    assert strength_to_level(1.6) == "confirmed"
    led = RevealLedger()
    # 一個決定性證據（strength 1.6）→ 直接 confirmed（不卡 hinted）
    RevelationBridge().apply(led, [EvidenceEvent("e", "kernel", "t1", evidence_strength=1.6)])
    assert led.level_of("t1") == "confirmed"


def test_weak_evidence_accumulates():
    led = RevealLedger()
    b = RevelationBridge()
    b.apply(led, [EvidenceEvent("e1", "kernel", "t1", evidence_strength=0.3)])
    assert led.level_of("t1") == "hinted"
    b.apply(led, [EvidenceEvent("e2", "kernel", "t1", evidence_strength=0.4)])
    assert led.level_of("t1") == "observed"          # 0.3+0.4=0.7 ≥ 0.6


# ── #1：kernel clue 自帶 truth_id ──────────────────────────────────────────
def test_evidence_from_clue_values_reads_truth_id():
    clues = [("clue.core", {"title": "核心", "content": "…", "truth_id": "truth.core",
                            "evidence_strength": 1.6, "max_level": "actionable"}),
             ("clue.deco", {"title": "氛圍", "content": "…"})]      # 無 truth_id
    events, unmapped = evidence_from_clue_values(clues, beat=3)
    assert len(events) == 1 and events[0].truth_id == "truth.core"
    assert events[0].evidence_strength == 1.6
    assert unmapped == ["clue.deco"]


# ── #3 + #6：寫進 revealed_bible（含分層），confirmed 進 revealed_fragments ──
def test_writes_levels_into_revealed_bible():
    bb = _BB()
    led = build_ledger_from_bible(_bible())
    RevelationBridge().apply(led, evidence_from_clue_values(
        [("clue.core", {"truth_id": "truth.core", "evidence_strength": 1.6})])[0])
    write_ledger_to_revealed_bible(bb, led)
    rb = bb.snapshot()["revealed_bible"]
    assert rb["truth_progress"]["truth.core"]["level"] == "confirmed"     # #3 記等級
    assert rb["truth_progress"]["truth.signal"]["level"] == "hidden"
    ids = {f["id"] for f in rb["revealed_fragments"]}
    assert "truth.core" in ids                                            # #6 confirmed 進碎片


# ── #5：NPC 回覆有效資訊走同一 bridge ───────────────────────────────────────
def test_npc_reply_through_bridge():
    idx = build_truth_keyword_index(_bible())
    # NPC 提到 432.7 校準 → 命中 truth.signal
    evs = evidence_from_npc_reply("432.7 是聲納校準值，別問太多。", idx, beat=4,
                                  answer_status="partial")
    assert any(e.truth_id == "truth.signal" and e.source == "npc_chat" for e in evs)
    assert all(e.max_level == "observed" for e in evs)       # NPC 封頂 observed
    # 迴避不給證據
    assert evidence_from_npc_reply("我什麼都不知道。", idx, answer_status="evaded") == []
    # 走 bridge → 帳本前進
    led = build_ledger_from_bible(_bible())
    RevelationBridge().apply(led, evs)
    assert led.level_of("truth.signal") != "hidden"


# ── #4：recap 分層明細含標題（非全 ？？？）──────────────────────────────────
def test_recap_has_tiered_titles():
    led = build_ledger_from_bible(_bible())
    b = RevelationBridge()
    b.apply(led, [EvidenceEvent("e", "kernel", "truth.core", evidence_strength=1.6)])     # confirmed
    b.apply(led, [EvidenceEvent("e2", "kernel", "truth.signal", evidence_strength=0.3)])  # hinted
    recap = recap_from_ledger(led)
    assert recap["found"] == 2 and recap["confirmed"] == 1
    assert recap["confirmed_list"][0]["title"] == "核心真相"
    assert recap["hinted_list"][0]["title"] == "432.7 的意義"
    assert recap["hidden_list"][0]["title"] == "林晨的下落"
    assert "你發現了 2/3" in recap["line"]


# ── 持久化 round-trip（revealed_bible.truth_progress）───────────────────────
def test_progress_dict_roundtrip():
    led = build_ledger_from_bible(_bible())
    RevelationBridge().apply(led, [EvidenceEvent("e", "kernel", "truth.core", evidence_strength=0.7)])
    restored = RevealLedger.from_progress_dict(led.to_progress_dict(), _bible()["revelation_pool"])
    assert restored.level_of("truth.core") == "observed"
    assert restored.truths["truth.core"].title == "核心真相"
