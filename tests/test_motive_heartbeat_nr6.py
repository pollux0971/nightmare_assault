"""NR6 — Motive Heartbeat 驗收測試（敘事控制 v0.2）。

驗收：連 3 beat 無動機提及 → required；提及即重置；loop pre 注入動機提醒義務（換管道、不重複句）；
post 依敘事是否提及動機更新心跳。
"""
from __future__ import annotations

from core.narrative.motif_tracker import MotiveHeartbeat


# ── 連 3 beat 無提及 → required（reference 對齊）─────────────────────────────
def test_required_after_three_silent_beats():
    hb = MotiveHeartbeat(max_beats_without_motive=3)
    hb.register_beat(False); hb.register_beat(False); hb.register_beat(False)
    assert hb.required()


# ── 提及即重置 ───────────────────────────────────────────────────────────────
def test_reference_resets():
    hb = MotiveHeartbeat(max_beats_without_motive=3)
    hb.register_beat(False); hb.register_beat(False)
    hb.register_beat(True)                          # 提到動機
    assert not hb.required()
    assert hb.beats_since_motive == 0


# ── loop pre：逾期 → 注入動機提醒義務（含「換一種方式」）───────────────────
def test_loop_motive_heartbeat_injects_obligation():
    from core.orchestrator_loop import BeatLoop
    from core.narrative.models import NarrativeContract, ProtagonistMotive, MotifPalette
    loop = BeatLoop.__new__(BeatLoop)
    loop._motive_heartbeat = MotiveHeartbeat(max_beats_without_motive=3)
    loop._narrative_contract = NarrativeContract(
        core_premise="x",
        protagonist_motive=ProtagonistMotive("弟弟林晨失蹤", "找到林晨", "不能再失去他", "林晨的紙條"),
        central_question="?", motif_palette=MotifPalette())
    # 尚未逾期 → 不注入
    ctx = {}
    loop._motive_heartbeat_pre(ctx)
    assert "motive_heartbeat" not in ctx
    # 逾期 → 注入
    loop._motive_heartbeat._beats_since = None
    loop._motive_heartbeat.beats_since_motive = 3
    ctx2 = {}
    loop._motive_heartbeat_pre(ctx2)
    assert ctx2.get("motive_heartbeat") is True
    obl = " ".join(ctx2["narrative_obligations"])
    assert "換一種方式" in obl and "重複同一句" in obl


# ── loop post：敘事提到失蹤者名字 → 視為提及、重置心跳 ──────────────────────
def test_loop_post_resets_when_motive_referenced():
    from core.orchestrator_loop import BeatLoop
    from core.narrative.models import NarrativeContract, ProtagonistMotive, MotifPalette
    loop = BeatLoop.__new__(BeatLoop)
    loop._motif_tracker = None
    loop._motive_heartbeat = MotiveHeartbeat(max_beats_without_motive=3)
    loop._motive_heartbeat.beats_since_motive = 2
    loop._narrative_contract = NarrativeContract(
        core_premise="x",
        protagonist_motive=ProtagonistMotive("林晨", "找到林晨", "賭注", "紙條"),
        central_question="?", motif_palette=MotifPalette())
    loop._motif_motive_post("你在牆上看到林晨刻下的字。", {})
    assert loop._motive_heartbeat.beats_since_motive == 0
