#!/usr/bin/env python3
"""step4_patchb_validation — Patch B（NPC onboarding + beat rendering）完整 UX 驗證（真 LLM、flag ON）。

驗 6 重點：①NPC 首現有來歷線索 ②首問答非純 API ③personality 影響語氣
④一般 beat 不再短成 log（蒐 short_streak）⑤review_mode 仍短且不新增 fact ⑥TruthEvidenceGate 不受影響。
順便蒐集 beat_rendering 數據 → 決定要不要開 Step 3（P4）。
"""
from __future__ import annotations

import json
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))


def main():
    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    from dev.tools.agent_play import _make_caller, _new_loop, _reveal_progress, _do_chat
    from core.world.actor_profile import get_npc_profile

    caller, note = _make_caller(str(ROOT / "config" / "config.json"), no_llm=False)
    loop = _new_loop(caller)
    P = print
    P(f"[caller={note}  flag=ON]")
    res = loop.start({"theme": "午夜的廢棄海事研究站（弟弟林晨失蹤）",
                      "protagonist_name": "周凱", "npc_count": 2})
    P("【開場】" + (res.get("narrative") or "")[:120].replace("\n", " "))

    log = open(ROOT / "dev" / "reports" / "step4-patchb.jsonl", "w", encoding="utf-8")
    beats = []          # 每 beat 的 rendering + reveal gate

    def fact_ids():
        return {e.id for e in loop._world.by_kind("fact")} if loop._world else set()

    def step(label, action):
        before_rev = _reveal_progress(loop)
        out = loop.step(action)
        after_rev = _reveal_progress(loop)
        br = out.get("beat_rendering") or {}
        rec = {"label": label, "action": action,
               "beat_type": br.get("beat_type"), "actual_chars": br.get("actual_chars"),
               "too_short": br.get("too_short"), "short_streak": br.get("short_streak"),
               "reveal_gate_allowed": out.get("reveal_gate_allowed"),
               "action_class": out.get("action_class"),
               "reveal_delta": sum(after_rev[k] - before_rev[k]
                                   for k in ("hinted", "observed", "suspected", "confirmed")),
               "investigation_state": (out.get("world_progress") or {}).get("investigation_state"),
               "narrative_excerpt": (out.get("narrative") or "")[:90].replace("\n", " ")}
        beats.append(rec); log.write(json.dumps(rec, ensure_ascii=False) + "\n"); log.flush()
        P(f"\n▶ {label}（{action[:24]}…）")
        P(f"  beat_type={rec['beat_type']} chars={rec['actual_chars']} too_short={rec['too_short']} "
          f"streak={rec['short_streak']} | class={rec['action_class']} gate={rec['reveal_gate_allowed']} "
          f"reveal_Δ={rec['reveal_delta']} mode={rec['investigation_state']}")
        return out

    # ── 探索 beats（蒐 beat 長度 + 確認非 truth 行動不推 reveal）────────────────
    step("b1_search", "我搜查這個房間，翻看桌面抽屜地上，找能辨識身分或留線索的東西")
    step("b2_explore", "我往這層深處走，尋找其他人或可離開的路")

    # ── NPC 首次接觸（onboarding）+ personality ──────────────────────────────
    present = [n.get("name") for n in loop.bb.snapshot().get("npc_registry") or []
               if isinstance(n, dict) and n.get("presence", "present") == "present"]
    npc_eval = []
    for npc in present[:2]:
        pf = get_npc_profile(loop.bb, npc)
        rev_b = _reveal_progress(loop)
        intro_before = pf.intro_state
        reply = _do_chat(loop, {"npc": npc, "text": "你是誰？這裡的通訊設備還能用嗎？"})
        pf2 = get_npc_profile(loop.bb, npc)
        rev_a = _reveal_progress(loop)
        d = {"npc": npc, "personality": pf.personality_description,
             "intro_before": intro_before, "intro_after": pf2.intro_state,
             "reply_len": len(reply or ""), "reply": reply,
             "reveal_unchanged": rev_b == rev_a}
        npc_eval.append(d)
        P(f"\n💬 {npc}（首次, {pf.personality_description[:18]}…）")
        P(f"  reply({len(reply or '')}字): {(reply or '')[:160]}")
        P(f"  intro {intro_before}→{pf2.intro_state} | reveal 不變={rev_b == rev_a}")

    # ── 引用 NPC fact 找路（非 truth → 不該推 reveal）──────────────────────────
    step("b4_npcfact", "我根據他說的機房方向去找通訊設備，不碰真相")

    # ── 真相調查（truth_investigation → reveal 可推，證 gate 仍正常）────────────
    step("b5_truth", "我坐下來仔細研究這些實驗紀錄與異常頻率的數據")

    # ── 撤退整理（review_mode → 短 + 不新增 fact）─────────────────────────────
    facts_before = fact_ids()
    out6 = step("b6_review", "先退到外面整理線索，不結束本次調查，只整理已知")
    facts_after = fact_ids()
    review_no_new_fact = facts_after.issubset(facts_before)

    # ── 評估 ───────────────────────────────────────────────────────────────────
    general = [b for b in beats if b["beat_type"] not in ("review_mode", "ending")]
    short_streaks = [b["short_streak"] for b in beats]
    nontruth = [b for b in beats if b["label"] in ("b1_search", "b2_explore", "b3_listen", "b4_npcfact")]
    truth_beat = next((b for b in beats if b["label"] == "b5_truth"), {})
    report = {
        "1_npc_first_appearance": {
            "all_unintroduced_then_introduced": all(
                e["intro_before"] == "unintroduced" and e["intro_after"] == "introduced" for e in npc_eval),
            "npcs": [{"npc": e["npc"], "intro": f'{e["intro_before"]}→{e["intro_after"]}'} for e in npc_eval]},
        "2_first_qa_not_api": {
            "first_reply_lengths": [e["reply_len"] for e in npc_eval],
            "all_substantial": all(e["reply_len"] >= 60 for e in npc_eval)},  # 首問答夠長 ≈ 非純資訊
        "3_personality_differs": {
            "personalities": [e["personality"][:24] for e in npc_eval],
            "distinct": len({e["personality"] for e in npc_eval}) == len(npc_eval) if len(npc_eval) > 1 else None},
        "4_beats_not_log": {
            "short_streaks": short_streaks, "max_short_streak": max(short_streaks) if short_streaks else 0,
            "general_too_short_count": sum(1 for b in general if b["too_short"]),
            "general_beats": len(general)},
        "5_review_mode": {
            "beat_type": out6.get("beat_rendering", {}).get("beat_type"),
            "investigation_state": (out6.get("world_progress") or {}).get("investigation_state"),
            "no_new_fact": review_no_new_fact},
        "6_truth_gate_intact": {
            "nontruth_pushed_reveal": any(b["reveal_delta"] > 0 for b in nontruth),  # 應 False
            "nontruth_gate_allowed": [b["reveal_gate_allowed"] for b in nontruth],   # 應全 False
            "truth_class": truth_beat.get("action_class"),
            "truth_gate_allowed": truth_beat.get("reveal_gate_allowed")},
        "P4_verdict": ("建議開 Step 3 P4：一般 beat 連續 ≥2 過短" if (max(short_streaks) if short_streaks else 0) >= 2
                       else "P4 暫不需要：未出現連續 ≥2 過短"),
    }
    log.write(json.dumps({"type": "report", "report": report}, ensure_ascii=False) + "\n"); log.close()
    P("\n" + "#" * 72); P("【Step 4 驗證報告】")
    P(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
