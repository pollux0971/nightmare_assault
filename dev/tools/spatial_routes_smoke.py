#!/usr/bin/env python3
"""spatial_routes_smoke — focused real-LLM 驗證 Spatial Routes Projection（commit 25ab123）。

只驗證、不改 core。驅動玩家在區域間移動 + 撤退，逐 beat 抓 spatial_summary / routes_from_here /
blocked_routes / current_area / previous_area / exploration_mode，檢查結構性 route 是否如預期出現。

跑法：python3 dev/tools/spatial_routes_smoke.py --out /tmp/sr.jsonl
"""
from __future__ import annotations

import argparse
import json
import os
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

RID = {"route.return_previous": "返回上一區", "route.return_site": "返回現場",
       "route.withdraw_safe": "暫退安全區", "route.campaign_exit": "離場"}

STEPS = [
    "我推開眼前的門，往走廊深處走去探索下一個區域",
    "我繼續往更裡面走，進到下一個房間查看",
    "我沿著原路退回我剛才來的那個地方",
    "我想找個安全的地方停下來整理思緒，往出口方向退開危險",
    "我重新打起精神，往另一個還沒去過的方向走去",
    "我再往更深處推進，看看盡頭有什麼",
]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="/tmp/sr.jsonl")
    ap.add_argument("--config", default=str(ROOT / "config" / "config.json"))
    args = ap.parse_args()

    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    os.environ["ENABLE_OPENING_VARIATION_CONTRACT"] = "true"
    sys.path.insert(0, str(ROOT / "dev" / "tools"))
    from dev.tools.agent_play import _make_caller, _dp_to_obs
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database
    from core.signal import SignalBus

    caller, note = _make_caller(args.config, no_llm=False)
    print(f"[spatial_routes_smoke] caller={note} flags=NC+OVC ON")
    loop = BeatLoop(caller, Blackboard(), Database(tempfile.mktemp(suffix=".db")), SignalBus(),
                    run_id="sr", use_kernel=True)
    out = open(args.out, "w", encoding="utf-8")

    def emit(rec):
        out.write(json.dumps(rec, ensure_ascii=False) + "\n"); out.flush()

    def snap(loop, action, o=None):
        sd = loop.spatial_debug()
        routes = sd.get("routes_from_here") or []
        blocked = sd.get("blocked_routes") or []
        w = loop._world
        return {
            "action": action,
            "current_area": getattr(w, "current_area", None),
            "previous_area": getattr(w, "previous_area", None),
            "exploration_mode": (loop.bb.game_meta or {}).get("exploration_mode"),
            "routes_from_here": [{"exit_id": r.get("exit_id"), "label": r.get("label"),
                                  "to_area": r.get("to_area"), "state": r.get("state")} for r in routes],
            "blocked_routes": [{"label": r.get("label"), "state": r.get("state"),
                                "requires": r.get("requires")} for r in blocked],
            "structural_route_ids": [r.get("exit_id") for r in routes
                                     if str(r.get("exit_id", "")).startswith("route.")],
            "spatial_summary": sd.get("spatial_summary"),
        }

    res = loop.start({"theme": "午夜的廢棄海事研究站（弟弟失蹤，多個相連艙區）",
                      "protagonist_name": "周凱", "npc_count": 1})
    rec = snap(loop, "(opening)")
    rec["kind"] = "opening"
    emit(rec)
    sline = [l for l in (rec["spatial_summary"] or "").split("\n") if "可走路線" in l]
    print(f"[開場] area={rec['current_area']} routes={rec['structural_route_ids']}")
    print(f"       {sline[0] if sline else '(no route line)'}")
    dp = res["decision_point"]

    for i, step in enumerate(STEPS):
        o = loop.step(step)
        dp = o.get("decision_point")
        rec = snap(loop, step, o)
        rec["kind"] = "beat"
        rec["beat"] = loop.beat_number
        emit(rec)
        rline = [l for l in (rec["spatial_summary"] or "").split("\n") if "可走路線" in l]
        bline = [l for l in (rec["spatial_summary"] or "").split("\n") if "被阻擋" in l]
        print(f"  [b{rec['beat']}] area={rec['current_area']} prev={rec['previous_area']} "
              f"mode={rec['exploration_mode']}")
        print(f"        routes={[RID.get(x, x) for x in rec['structural_route_ids']]}  "
              f"+explicit={[r['label'] for r in rec['routes_from_here'] if not str(r['exit_id']).startswith('route.')]}")
        print(f"        {rline[0] if rline else '(no route line)'}")
        if bline:
            print(f"        {bline[0]}")
        if o.get("ended"):
            break

    out.close()
    print(f"\nwritten → {args.out}")


if __name__ == "__main__":
    main()
