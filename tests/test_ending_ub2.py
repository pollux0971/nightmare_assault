"""UB2 — 一種結局序列 驗收測試。

驗收：觸發後純敘述收尾、揭露完整 real truth、（提早觸發）劇情內合理化解、可回選單。
"""
from __future__ import annotations

from core.agents.ending import build_ending_sequence, render_ending_text
from core.blackboard import Blackboard


def _bb_with_truth(revealed_ids=()):
    bb = Blackboard()
    bb.write("setup", "real_bible", {
        "world_truth": {"what_really_happened": "這裡曾進行非法人體實驗，全員滅口",
                        "the_threat_is": "失敗實驗體的集體怨念",
                        "deadly_rule": "絕不能說出『代號七號』這個名字"},
        "revelation_pool": [
            {"id": "f1", "content": "地下室有一份名單"},
            {"id": "f2", "content": "護士長是唯一倖存者"},
            {"id": "f3", "content": "代號七號就是你弟弟"},
        ],
    })
    bb.write("orchestrator", "revealed_bible",
             {"revealed_fragments": [{"id": i, "content": "x"} for i in revealed_ids]})
    return bb


# ── 完整真相揭露 ─────────────────────────────────────────────────────────────
def test_ending_reveals_full_truth():
    seq = build_ending_sequence(_bb_with_truth(), {"type": "escape", "via": "attractor"})
    assert seq["is_ending"] is True
    assert seq["closing"]                                  # 純敘述收尾
    assert "非法人體實驗" in seq["truth"]["what_really_happened"]
    assert "代號七號" in seq["truth"]["deadly_rule"]
    assert len(seq["truth"]["all_fragments"]) == 3         # 完整碎片


# ── 復盤：發現 vs 全部 ───────────────────────────────────────────────────────
def test_recap_counts_discovered():
    seq = build_ending_sequence(_bb_with_truth(revealed_ids=("f1",)), {"type": "truth_revealed"})
    r = seq["recap"]
    assert r["found_count"] == 1 and r["total_count"] == 3
    assert r["early"] is True                              # 還有沒找到 → 提早觸發
    assert any("名單" in c["content"] for c in r["discovered"])
    assert any("護士長" in c["content"] for c in r["missed"])


# ── 提早觸發（什麼都沒發現就死）→ full 模式真相一次攤開（debug）─────────────
def test_early_trigger_full_mode_dumps_all_missed():
    seq = build_ending_sequence(_bb_with_truth(revealed_ids=()), {"type": "death_physical"})
    assert seq["recap"]["found_count"] == 0
    assert len(seq["recap"]["missed"]) == 3               # 資料層仍含全部
    text = render_ending_text(seq, mode="full")           # full = debug 全揭
    assert "真相" in text and "復盤" in text and "代號七號" in text


# ── 各結局型別都有收尾標題/敘述 ──────────────────────────────────────────────
def test_all_ending_types_have_closing():
    for et in ["death_physical", "death_mental", "escape", "truth_revealed", "transformation", "death"]:
        seq = build_ending_sequence(_bb_with_truth(), {"type": et})
        assert seq["title"] and seq["closing"]


# ── 整合：loop 觸發結局 → ending 被補上 is_ending/truth/recap ─────────────────
def test_loop_finalizes_ending():
    from core.orchestrator_loop import BeatLoop
    from core.persistence.db import Database
    loop = BeatLoop(caller=None, blackboard=_bb_with_truth(revealed_ids=("f1", "f2")),
                    db=Database(), run_id="t", use_kernel=False)
    loop.ended = True
    loop.ending = {"type": "escape", "via": "attractor"}
    loop._finalize_ending()
    assert loop.ending["is_ending"] and loop.ending["truth"]["the_threat_is"]
    assert loop.ending["recap"]["found_count"] == 2
    # 只補一次（idempotent）
    snapshot = dict(loop.ending)
    loop._finalize_ending()
    assert loop.ending == snapshot
