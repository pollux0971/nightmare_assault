"""HA3 — Hidden Recap Masking 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：public_recap 不含任何 hidden truth content / 真實標題；hidden_count + 遮罩標題正確；
agent_play observation（預設）不洩 hidden content；--debug-reveal-truth 才露 full。
"""
from __future__ import annotations

import json

from core.narrative.revelation import (
    build_ledger_from_bible, RevelationBridge, EvidenceEvent, public_recap,
)


def _bible():
    return {"revelation_pool": [
        {"id": "t.sig", "title": "432.7 的意義", "content": "聲納校準值，極機密ZZZ"},
        {"id": "t.lin", "title": "林晨的下落", "content": "他在水下設備室SECRET"},
        {"id": "t.core", "title": "核心真相", "content": "你也是實驗體HIDDEN"}]}


def _ledger_with_one_confirmed():
    led = build_ledger_from_bible(_bible())
    RevelationBridge().apply(led, [EvidenceEvent("e", "kernel", "t.sig", evidence_strength=1.6)])
    return led


# ── public_recap 不洩 hidden content / 真實標題 ─────────────────────────────
def test_public_recap_hides_unfound_content():
    led = _ledger_with_one_confirmed()      # t.sig confirmed；t.lin/t.core hidden
    rc = public_recap(led)
    blob = json.dumps(rc, ensure_ascii=False)
    # 已確認的可露（玩家已得）
    assert "聲納校準值" in blob
    # 未發現的 content / 真實標題一律不得出現
    assert "SECRET" not in blob and "水下設備室" not in blob
    assert "HIDDEN" not in blob and "實驗體" not in blob
    assert "林晨的下落" not in blob and "核心真相" not in blob
    assert rc["hidden_count"] == 2
    assert rc["hidden_titles"] == ["未解的線索 #1", "未解的線索 #2"]


# ── 全 hidden 時也不洩 ───────────────────────────────────────────────────────
def test_all_hidden_no_leak():
    led = build_ledger_from_bible(_bible())
    rc = public_recap(led)
    blob = json.dumps(rc, ensure_ascii=False)
    for secret in ("ZZZ", "SECRET", "HIDDEN", "聲納校準值", "水下設備室"):
        assert secret not in blob
    assert rc["hidden_count"] == 3
    assert rc["found"] == []


# ── agent_play observation 預設遮罩、debug flag 才露 ─────────────────────────
def test_agent_play_ending_dict_masks_by_default():
    import importlib.util
    spec = importlib.util.spec_from_file_location("agent_play", "dev/tools/agent_play.py")
    ap = importlib.util.module_from_spec(spec); spec.loader.exec_module(ap)

    class _Loop:
        _reveal_ledger = _ledger_with_one_confirmed()
    ending = {"type": "escape", "ending_surface": "ambiguous_escape", "closing": "你走了。"}

    ap._REVEAL_TRUTH_DEBUG = False
    masked = ap._ending_dict(_Loop(), ending)
    blob = json.dumps(masked, ensure_ascii=False)
    assert "SECRET" not in blob and "HIDDEN" not in blob and "林晨的下落" not in blob
    assert masked["recap"]["hidden_count"] == 2

    ap._REVEAL_TRUTH_DEBUG = True
    full = ap._ending_dict(_Loop(), ending)
    # debug 模式允許 full（這裡只確認不崩、結構在）
    assert full["recap"] is not None
    ap._REVEAL_TRUTH_DEBUG = False
