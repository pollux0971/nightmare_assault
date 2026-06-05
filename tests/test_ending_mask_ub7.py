"""UB7 — 結局揭露政策（masked）驗收測試。

驗收：masked → 已發現露全文、未發現只露遮罩標題（不露全文）+ 重玩鉤子；
core 真相僅在發現比例足夠時才額外揭露；full → 全揭；預設 masked。
"""
from __future__ import annotations

from core.agents.ending import build_ending_sequence, render_ending_text, MASKED_TRUTH_RATIO
from core.blackboard import Blackboard

SECRET = "完整真相：記憶置換療法害死十一人SECRET"


def _bb(revealed_ids=(), titled=True):
    bb = Blackboard()
    pool = [
        {"id": "f1", "title": "地下室名單", "content": "地下室藏著一份死亡名單"},
        {"id": "f2", "title": "陳國棟筆記", "content": "院長知情並掩蓋"},
        {"id": "f3", "title": "B2 鑰匙", "content": "通往地下二層的鑰匙"},
    ]
    if not titled:
        for f in pool:
            f.pop("title")
    bb.write("setup", "real_bible", {
        "world_truth": {"what_really_happened": SECRET, "the_threat_is": "運轉中的機器",
                        "deadly_rule": "勿在聲波區久留"},
        "revelation_pool": pool})
    bb.write("orchestrator", "revealed_bible",
             {"revealed_fragments": [{"id": i} for i in revealed_ids]})
    return bb


# ── masked 預設：未發現只露遮罩標題、不露全文 + 重玩鉤子 ─────────────────────
def test_masked_hides_missed_content_shows_title():
    seq = build_ending_sequence(_bb(revealed_ids=("f1",)), {"type": "death_physical"})
    text = render_ending_text(seq)                          # 預設 masked
    assert "地下室藏著一份死亡名單" in text                  # 已發現 → 全文
    assert "已確認" in text and "未確認" in text
    assert "陳國棟筆記：？？？" in text                       # 未發現 → 遮罩標題
    assert "院長知情並掩蓋" not in text                      # 未發現 → 不露全文
    assert "通往地下二層的鑰匙" not in text
    assert "你還沒走到它面前" in text                        # 重玩鉤子


def test_masked_low_ratio_hides_core_truth():
    seq = build_ending_sequence(_bb(revealed_ids=("f1",)), {"type": "escape"})   # 1/3 < 0.6
    text = render_ending_text(seq, mode="masked")
    assert SECRET not in text                               # 探索不足 → 核心真相不揭露


def test_masked_high_ratio_reveals_core_truth():
    seq = build_ending_sequence(_bb(revealed_ids=("f1", "f2", "f3")), {"type": "truth_revealed"})  # 3/3
    assert (3 / 3) >= MASKED_TRUTH_RATIO
    text = render_ending_text(seq, mode="masked")
    assert SECRET in text                                   # 探索足夠 → 揭露核心真相
    assert "未確認" not in text                              # 沒有未發現的


# ── full 模式：全揭（含未發現全文 + 核心真相）──────────────────────────────
def test_full_mode_reveals_everything():
    seq = build_ending_sequence(_bb(revealed_ids=()), {"type": "death_physical"})
    text = render_ending_text(seq, mode="full")
    assert SECRET in text and "院長知情並掩蓋" in text        # 未發現全文也露


def test_default_mode_is_masked():
    seq = build_ending_sequence(_bb(revealed_ids=()), {"type": "death_physical"})
    assert render_ending_text(seq) == render_ending_text(seq, mode="masked")
    assert render_ending_text(seq) != render_ending_text(seq, mode="full")


# ── 無 title 的碎片 → 用遮罩通用標題（不露 content）─────────────────────────
def test_untitled_missed_uses_generic_label():
    seq = build_ending_sequence(_bb(revealed_ids=(), titled=False), {"type": "death_physical"})
    text = render_ending_text(seq)
    assert "未解的線索" in text                              # 通用遮罩標題
    assert "院長知情並掩蓋" not in text                      # 仍不露 content
