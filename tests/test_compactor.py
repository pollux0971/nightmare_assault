"""tests/test_compactor.py — U14 Compactor + 30 beat 承重牆驗收。

涵蓋（E3 / C11 / A5 / F7）：
  - BeatWindow：push 超過 size → 回傳被擠出最舊 beat；視窗永遠 <= size。
  - RollingSummary：取代語意 + 有界（needs_recompaction）。
  - FactLedger：add/merge 去重；protected 多輪後仍存在、拒絕刪除。
  - compression_level：usage_pct 過 L1/L2/L3 → 等級不同。
  - 30 beat 模擬（核心）：
      * rolling_summary 估算 token 不無限成長（有界 <= cap 級別）。
      * 標為「伏筆」的 protected 內容 30 beat 後仍在 preserved_foreshadowings。
      * ledger 硬事實 30 beat 後仍在。
      * window 始終 <= BEAT_WINDOW_SIZE。
  - apply_to_blackboard：submit_patch → merge 前 snapshot 不變、merge 後更新（A5）。
"""
from __future__ import annotations

from core.constants import (
    BEAT_WINDOW_SIZE,
    CONTEXT_THRESHOLD_L1,
    CONTEXT_THRESHOLD_L2,
    CONTEXT_THRESHOLD_L3,
    SUMMARY_TOKEN_CAP,
)
from core.agents.compactor import Compactor, compression_level
from core.blackboard import Blackboard
from core.memory.summary import (
    BeatWindow,
    FactLedger,
    RollingSummary,
    estimate_tokens,
)
from core.models import CompactorOutput, LedgerFact


# ─────────────────────────────────────────────────────────────────────────────
# 假 caller：回有界的固定短摘要，preserved_foreshadowings 原樣帶回保護清單
# ─────────────────────────────────────────────────────────────────────────────

# 固定短摘要：無論餵多少 beat，compressed_summary 都被壓到這個長度級別（有界）
_BOUNDED_SUMMARY = "故事至今：主角在廢棄醫院探索，氣氛緊繃，數名 NPC 行為可疑。"


class _BoundedCaller:
    """假 SkillCaller：

    - compressed_summary 永遠是有界的固定短摘要（模擬 LLM 把舊摘要 + 滑出 beat
      濃縮成定長散文）。
    - preserved_foreshadowings 原樣回傳 context 帶入的保護清單（C11：不丟伏筆）。
    - ledger_updates 把 context.ledger 二元組原樣轉回 LedgerFact（不發明、不刪）。
    - final_usage_estimate 隨壓縮略降。
    """

    def __init__(self) -> None:
        self.call_count = 0

    def call(self, agent: str, context: dict, output_model=None, temperature=None):
        assert agent == "compactor"
        self.call_count += 1
        protected = list(context.get("protected_foreshadowings", []))
        ledger_pairs = context.get("ledger", [])
        ledger_updates = [
            LedgerFact(type=p[0], content=p[1]) for p in ledger_pairs
        ]
        usage = context.get("usage_pct", 0.0)
        return CompactorOutput(
            compressed_summary=_BOUNDED_SUMMARY,
            ledger_updates=ledger_updates,
            archived_beats=[],
            preserved_foreshadowings=protected,
            final_usage_estimate=max(0.0, usage * 0.5),
        )


# ─────────────────────────────────────────────────────────────────────────────
# BeatWindow
# ─────────────────────────────────────────────────────────────────────────────

class TestBeatWindow:
    def test_push_under_size_no_eviction(self):
        w = BeatWindow(size=3)
        assert w.push("b1") == []
        assert w.push("b2") == []
        assert len(w) == 2

    def test_push_over_size_evicts_oldest(self):
        w = BeatWindow(size=3)
        for b in ("b1", "b2", "b3"):
            assert w.push(b) == []
        # 第 4 個 → 擠出最舊的 b1
        evicted = w.push("b4")
        assert evicted == ["b1"]
        assert w.items() == ["b2", "b3", "b4"]

    def test_window_never_exceeds_size(self):
        w = BeatWindow(size=BEAT_WINDOW_SIZE)
        for i in range(50):
            w.push(f"beat-{i}")
            assert len(w) <= BEAT_WINDOW_SIZE
        assert len(w) == BEAT_WINDOW_SIZE

    def test_items_is_copy(self):
        w = BeatWindow(size=3)
        w.push("b1")
        got = w.items()
        got.append("hacked")
        assert w.items() == ["b1"]

    def test_invalid_size(self):
        import pytest
        with pytest.raises(ValueError):
            BeatWindow(size=0)


# ─────────────────────────────────────────────────────────────────────────────
# RollingSummary
# ─────────────────────────────────────────────────────────────────────────────

class TestRollingSummary:
    def test_update_replaces(self):
        rs = RollingSummary()
        rs.update("第一版摘要")
        rs.update("第二版摘要")
        # 取代語意，非串接
        assert rs.text == "第二版摘要"

    def test_estimate_tokens_char_count(self):
        rs = RollingSummary()
        rs.update("一二三四五")
        assert rs.estimate_tokens() == 5
        assert estimate_tokens("一二三四五") == 5

    def test_needs_recompaction_when_over_cap(self):
        rs = RollingSummary(cap_tokens=10)
        rs.update("短")
        assert rs.needs_recompaction() is False
        rs.update("一" * 50)
        assert rs.needs_recompaction() is True

    def test_bounded_after_many_updates(self):
        """反覆 update 取代式摘要不會無限成長。"""
        rs = RollingSummary(cap_tokens=SUMMARY_TOKEN_CAP)
        for i in range(100):
            rs.update(_BOUNDED_SUMMARY)
        assert rs.estimate_tokens() == len(_BOUNDED_SUMMARY)
        assert rs.needs_recompaction() is False


# ─────────────────────────────────────────────────────────────────────────────
# FactLedger
# ─────────────────────────────────────────────────────────────────────────────

class TestFactLedger:
    def test_add_dedup(self):
        led = FactLedger()
        assert led.add(LedgerFact(type="世界事實", content="證件年份對不上")) is True
        # 同 (type, content) → 不重複
        assert led.add(LedgerFact(type="世界事實", content="證件年份對不上")) is False
        assert len(led) == 1

    def test_merge_dedup_returns_new_count(self):
        led = FactLedger()
        facts = [
            LedgerFact(type="npc", content="張醫生:戒備"),
            LedgerFact(type="npc", content="張醫生:戒備"),  # 重複
            LedgerFact(type="技能", content="鎖匠:只對機械鎖"),
        ]
        added = led.merge(facts)
        assert added == 2
        assert len(led) == 2

    def test_protected_not_removed(self):
        led = FactLedger()
        secret = LedgerFact(type="伏筆", content="地下室名單尚未揭露")
        led.add(secret, protected=True)
        # 嘗試移除 protected → 失敗
        assert led.remove(secret) is False
        assert secret in led
        assert secret in led.protected_facts()

    def test_non_protected_removable(self):
        led = FactLedger()
        f = LedgerFact(type="碎片", content="名單:已揭露")
        led.add(f)
        assert led.remove(f) is True
        assert f not in led

    def test_protected_survives_many_rounds(self):
        """protected fact 在多輪 add/merge/remove 嘗試後仍存在。"""
        led = FactLedger()
        secret = LedgerFact(type="伏筆", content="閣樓有第三具屍體")
        led.add(secret, protected=True)
        for i in range(30):
            led.merge([LedgerFact(type="世界事實", content=f"事件{i}")])
            led.remove(secret)  # 每輪都嘗試刪（應全部失敗）
        assert secret in led
        assert secret in led.protected_facts()


# ─────────────────────────────────────────────────────────────────────────────
# compression_level
# ─────────────────────────────────────────────────────────────────────────────

class TestCompressionLevel:
    def test_levels_by_usage(self):
        assert compression_level(0.10) == "none"
        assert compression_level(CONTEXT_THRESHOLD_L1) == "L1"
        assert compression_level((CONTEXT_THRESHOLD_L1 + CONTEXT_THRESHOLD_L2) / 2) == "L1"
        assert compression_level(CONTEXT_THRESHOLD_L2) == "L2"
        assert compression_level(CONTEXT_THRESHOLD_L3) == "L3"
        assert compression_level(0.99) == "L3"

    def test_level_monotonic(self):
        order = {"none": 0, "L1": 1, "L2": 2, "L3": 3}
        prev = -1
        for u in [0.0, 0.5, 0.70, 0.80, 0.85, 0.90, 0.95, 1.0]:
            cur = order[compression_level(u)]
            assert cur >= prev
            prev = cur

    def test_compact_returns_level_via_usage_estimate(self):
        """usage_pct 過不同門檻 → compact 的 final_usage_estimate 反映壓縮。"""
        comp = Compactor(caller=_BoundedCaller())
        out_low = comp.compact([], "s", [], usage_pct=CONTEXT_THRESHOLD_L1)
        out_high = comp.compact([], "s", [], usage_pct=CONTEXT_THRESHOLD_L3)
        # 高使用率輸入 → 壓縮後估算仍 >= 低使用率（caller 用 *0.5 反映）
        assert out_high.final_usage_estimate > out_low.final_usage_estimate


# ─────────────────────────────────────────────────────────────────────────────
# Compactor.compact 基本行為
# ─────────────────────────────────────────────────────────────────────────────

class TestCompactBasic:
    def test_preserves_init_foreshadowings(self):
        comp = Compactor(
            caller=_BoundedCaller(),
            protected_foreshadowings=["閣樓的腳步聲", "醫生口袋裡的鑰匙"],
        )
        out = comp.compact([], "舊摘要", [], usage_pct=0.72)
        assert "閣樓的腳步聲" in out.preserved_foreshadowings
        assert "醫生口袋裡的鑰匙" in out.preserved_foreshadowings

    def test_absorbs_new_foreshadowings_from_caller(self):
        class _AddsForeshadow(_BoundedCaller):
            def call(self, agent, context, output_model=None, temperature=None):
                out = super().call(agent, context, output_model)
                out.preserved_foreshadowings = out.preserved_foreshadowings + ["新埋的伏筆"]
                return out

        comp = Compactor(caller=_AddsForeshadow(), protected_foreshadowings=["舊伏筆"])
        out = comp.compact([], "s", [], usage_pct=0.72)
        assert "舊伏筆" in out.preserved_foreshadowings
        assert "新埋的伏筆" in out.preserved_foreshadowings
        # 累積進 compactor 自持清單
        assert "新埋的伏筆" in comp.protected_foreshadowings

    def test_offline_fallback_no_caller(self):
        comp = Compactor(caller=None, protected_foreshadowings=["關鍵伏筆"])
        ledger = [LedgerFact(type="世界事實", content="門上有血跡")]
        out = comp.compact(["beat-1"], "現有摘要骨幹", ledger, usage_pct=0.80)
        # 離線降級：保留摘要骨幹 + ledger 透傳 + 保護清單
        assert out.compressed_summary == "現有摘要骨幹"
        assert "關鍵伏筆" in out.preserved_foreshadowings
        assert any(f.content == "門上有血跡" for f in out.ledger_updates)

    def test_caller_failure_falls_back(self):
        class _BrokenCaller:
            def call(self, *a, **k):
                raise RuntimeError("LLM 爆了")

        comp = Compactor(caller=_BrokenCaller(), protected_foreshadowings=["不能丟的伏筆"])
        out = comp.compact([], "骨幹摘要", [], usage_pct=0.95)
        assert out.compressed_summary == "骨幹摘要"
        assert "不能丟的伏筆" in out.preserved_foreshadowings


# ─────────────────────────────────────────────────────────────────────────────
# 30 beat 模擬（核心 E3 / C11）
# ─────────────────────────────────────────────────────────────────────────────

class TestThirtyBeatSimulation:
    def _run(self):
        """模擬餵 30 個假 beat：window 溢出就 compact。回傳收集到的狀態。"""
        caller = _BoundedCaller()
        # 一個標記為「伏筆」的 protected 內容
        foreshadow = "地下室名單上有主角的名字"
        comp = Compactor(caller=caller, protected_foreshadowings=[foreshadow])

        window = BeatWindow(size=BEAT_WINDOW_SIZE)
        summary = RollingSummary(cap_tokens=SUMMARY_TOKEN_CAP)
        ledger = FactLedger()
        # 一條硬事實，應在 30 beat 後仍在
        hard_fact = LedgerFact(type="世界事實", content="張醫生證件年份對不上")
        ledger.add(hard_fact)

        max_window_len = 0
        max_summary_tokens = 0
        last_out = None

        for i in range(30):
            # 每個 beat 是一段不短的原文（模擬故事 beat），用來檢驗摘要不隨之膨脹
            beat = {"id": f"beat-{i}", "text": f"第{i}個 beat 的完整故事原文，" * 20}
            evicted = window.push(beat)
            max_window_len = max(max_window_len, len(window))

            if evicted:
                # 視窗溢出 → compact（usage_pct 隨 beat 緩升模擬 context 壓力）
                usage = min(0.99, 0.50 + i * 0.02)
                out = comp.compact(
                    evicted_beats=evicted,
                    current_summary=summary.text,
                    ledger=ledger.facts(),
                    usage_pct=usage,
                )
                summary.update(out.compressed_summary)
                ledger.merge(out.ledger_updates)
                last_out = out
                max_summary_tokens = max(max_summary_tokens, summary.estimate_tokens())

        return {
            "comp": comp,
            "window": window,
            "summary": summary,
            "ledger": ledger,
            "foreshadow": foreshadow,
            "hard_fact": hard_fact,
            "max_window_len": max_window_len,
            "max_summary_tokens": max_summary_tokens,
            "last_out": last_out,
            "caller": caller,
        }

    def test_window_bounded(self):
        st = self._run()
        assert st["max_window_len"] <= BEAT_WINDOW_SIZE
        assert len(st["window"]) == BEAT_WINDOW_SIZE

    def test_summary_token_bounded(self):
        """rolling_summary 估算 token 不無限成長（有界 <= cap 級別）。"""
        st = self._run()
        # 30 beat 後摘要 token 仍被壓在 cap 內（甚至遠低於）
        assert st["max_summary_tokens"] <= SUMMARY_TOKEN_CAP
        assert st["summary"].needs_recompaction() is False
        # 具體有界值：就是那段固定短摘要的長度
        assert st["summary"].estimate_tokens() == len(_BOUNDED_SUMMARY)

    def test_foreshadowing_survives_30_beats(self):
        """標為伏筆的 protected 內容 30 beat 後仍在 preserved_foreshadowings（C11/E3）。"""
        st = self._run()
        assert st["foreshadow"] in st["last_out"].preserved_foreshadowings
        assert st["foreshadow"] in st["comp"].protected_foreshadowings

    def test_hard_fact_survives_30_beats(self):
        """ledger 的硬事實 30 beat 後仍在。"""
        st = self._run()
        assert st["hard_fact"] in st["ledger"]
        assert any(
            f.content == "張醫生證件年份對不上" for f in st["ledger"].facts()
        )

    def test_compact_was_actually_exercised(self):
        """確認 30 beat 中確實觸發了多次 compact（視窗有溢出）。"""
        st = self._run()
        # 30 beat、視窗 6 → 至少 30-6=24 次 compact
        assert st["caller"].call_count >= 30 - BEAT_WINDOW_SIZE


# ─────────────────────────────────────────────────────────────────────────────
# apply_to_blackboard（非同步隔離 A5）
# ─────────────────────────────────────────────────────────────────────────────

class TestApplyToBlackboard:
    def test_patch_isolated_until_merge(self):
        bb = Blackboard()
        bb.rolling_summary = "舊摘要"
        comp = Compactor(caller=_BoundedCaller())
        out = comp.compact([], "舊摘要", [], usage_pct=0.72)

        comp.apply_to_blackboard(bb, out)

        # merge 前：snapshot 完全看不到 pending patch（A5 隔離）
        snap_before = bb.snapshot()
        assert snap_before["rolling_summary"] == "舊摘要"
        assert bb.collect_pending()  # 確實有 pending

    def test_merge_updates_rolling_summary(self):
        bb = Blackboard()
        bb.rolling_summary = "舊摘要"
        comp = Compactor(caller=_BoundedCaller())
        ledger = [LedgerFact(type="世界事實", content="走廊盡頭有光")]
        out = comp.compact([], "舊摘要", ledger, usage_pct=0.72)

        comp.apply_to_blackboard(bb, out)
        report = bb.merge_and_bump()

        # merge 後：rolling_summary 更新為壓縮摘要、ledger 寫入
        assert report["applied"] == 2
        assert bb.rolling_summary == _BOUNDED_SUMMARY
        assert {"type": "世界事實", "content": "走廊盡頭有光"} in bb.ledger
        assert bb.version == 1

    def test_stale_patch_discarded(self):
        """base_version 過期的 patch 於 merge 時被丟棄（A5）。"""
        bb = Blackboard()
        comp = Compactor(caller=_BoundedCaller())
        out = comp.compact([], "x", [], usage_pct=0.72)
        comp.apply_to_blackboard(bb, out)
        # 先讓 version 前進，使 pending patch 過期
        bb.version = 5
        report = bb.merge_and_bump()
        assert report["discarded"] == 2
        assert report["applied"] == 0
