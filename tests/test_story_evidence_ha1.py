"""HB1 — Story Evidence Extraction 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：調查型 action + 具體 narrative + reveal 無變化 → 產 hinted evidence；命中 keyword → truth_id；
map 不到 → source=fallback、truth_id=None；reveal 已變化/非調查/無具體資訊 → 不產。
"""
from __future__ import annotations

from core.narrative.evidence_extractor import StoryEvidenceExtractor


def _ext():
    return StoryEvidenceExtractor({"truth.signal_frequency": ["432.7", "頻率", "赫茲"]})


# ── reference 對齊：有意義調查、無 reveal 變化 → hinted evidence ──────────────
def test_meaningful_investigation_produces_hint():
    events = _ext().extract(
        beat=2, action="拾起警告紙條仔細檢查",
        narrative="紙條上寫著：不要相信432.7。旁邊還有一段頻率紀錄。",
        reveal_changed=False)
    assert events
    assert events[0].truth_id == "truth.signal_frequency"
    assert events[0].max_level == "hinted"
    assert events[0].source == "story"


# ── map 不到 truth → fallback / truth_id None ───────────────────────────────
def test_unmapped_becomes_fallback():
    events = _ext().extract(
        beat=3, action="檢查牆上的儀器",
        narrative="儀器面板上有一組編號 B-12 和一段值班紀錄。",
        reveal_changed=False)
    assert events and events[0].truth_id is None
    assert events[0].source == "fallback"


# ── 抑制條件：reveal 已變化 / 非調查 / 無具體資訊 → 不產 ─────────────────────
def test_suppressed_cases():
    e = _ext()
    assert e.extract(beat=1, action="檢查紙條", narrative="紙條寫著432.7", reveal_changed=True) == []
    assert e.extract(beat=1, action="往前走", narrative="紙條寫著432.7", reveal_changed=False) == []
    assert e.extract(beat=1, action="檢查桌子", narrative="桌上空空如也。", reveal_changed=False) == []


# ── loop 整合：調查後 reveal_progress 不會完全沒反應 ────────────────────────
def test_loop_investigation_not_zero(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop()
    loop.start({"theme": "x", "npc_count": 1})
    # 連續調查（FakeCaller 的 narrative 含「頻率…」具體資訊）
    for txt in ["我檢查牆上的標誌", "我仔細查看走廊的頻率紀錄", "我研究地上的字條"]:
        loop.step(txt)
    # 至少有過 evidence（kernel 線索或 story 保底其一）
    rb = loop.bb.snapshot().get("revealed_bible") or {}
    tp = rb.get("truth_progress") or {}
    assert any(v.get("level") != "hidden" for v in tp.values())
