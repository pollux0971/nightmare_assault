#!/usr/bin/env python3
"""p5_onboarding_tryit — v0.7 P5 試玩：看 NPC onboarding + beat 渲染（真 LLM、flag ON）。

逐 beat 顯示 beat_rendering（type/實際字數/too_short/short_streak），
並與一個在場 NPC 首次對話 → 看 onboarding 介紹 + intro_state 轉換 + 不推 reveal。
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
    P("【開場】" + (res.get("narrative") or "")[:160].replace("\n", " "))

    streaks = []

    def show(label, out):
        br = out.get("beat_rendering") or {}
        streaks.append(br.get("short_streak", 0))
        P("\n" + "=" * 70)
        P(f"▶ {label}")
        P(f"  〔story〕{(out.get('narrative') or '')[:140].replace(chr(10),' ')}…")
        P(f"  〔beat_rendering〕type={br.get('beat_type')} "
          f"actual={br.get('actual_chars')} / 目標 {br.get('target_min_chars')}–{br.get('target_max_chars')} "
          f"too_short={br.get('too_short')} short_streak={br.get('short_streak')} "
          f"repair={br.get('repair_attempted')}")

    show("beat1 搜查", loop.step("我搜查這個房間，找任何能辨識身分或留下線索的東西"))
    show("beat2 前進", loop.step("我往研究站深處走，找其他人或出口"))
    show("beat3 觀察", loop.step("我停下來仔細聽周圍的聲音，注意有沒有人在附近"))

    # ── 與在場 NPC 首次對話（onboarding）──────────────────────────────────────
    present = [n.get("name") for n in loop.bb.snapshot().get("npc_registry") or []
               if isinstance(n, dict) and n.get("presence", "present") == "present"]
    if present:
        npc = present[0]
        prof_before = get_npc_profile(loop.bb, npc)
        rev_before = _reveal_progress(loop)
        P("\n" + "#" * 70)
        P(f"💬 與 {npc} 首次對話　intro_state(前)={prof_before.intro_state}　"
          f"個性={prof_before.personality_description[:24]}…")
        reply = _do_chat(loop, {"npc": npc, "text": "你是誰？這裡的通訊設備還能用嗎？"})
        P(f"  〔{npc}〕{reply}")
        prof_after = get_npc_profile(loop.bb, npc)
        rev_after = _reveal_progress(loop)
        P(f"  → intro_state(後)={prof_after.intro_state}　"
          f"reveal 不變={rev_before == rev_after}")
        # 第二次對話（已 introduced，不再走 onboarding）
        reply2 = _do_chat(loop, {"npc": npc, "text": "那機房在哪個方向？"})
        P(f"  〔{npc}・第二輪〕{reply2[:120]}")

    P("\n" + "#" * 70)
    P("【試玩總結】")
    P(json.dumps({
        "beats_played": len(streaks),
        "short_streaks": streaks,
        "max_short_streak": max(streaks) if streaks else 0,
        "建議": ("一般 beat 連續過短達 ≥2 → 可考慮開 P4 repair"
                if (max(streaks) if streaks else 0) >= 2 else
                "尚未出現連續過短，P4 暫不需要"),
    }, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
