#!/usr/bin/env python3
"""actor_consistency_reconfirm — very short real-LLM 驗證 NPC actor entity 一致性（commit 9156654）。

只驗證、不改 core。對話讓 NPC 成為 current_focus → 「那個人」指代 → 確認 resolved_kind=actor、
world.get(resolved_id) 不為 None、不新增 fact、不推 reveal。

跑法：python3 dev/tools/actor_consistency_reconfirm.py --out /tmp/actor.jsonl
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


def _rp(obs):
    r = obs.get("reveal_progress") or {}
    return f"{r.get('hinted',0)}/{r.get('observed',0)}/{r.get('suspected',0)}/{r.get('confirmed',0)}"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="/tmp/actor.jsonl")
    ap.add_argument("--config", default=str(ROOT / "config" / "config.json"))
    args = ap.parse_args()

    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    os.environ["ENABLE_OPENING_VARIATION_CONTRACT"] = "true"
    sys.path.insert(0, str(ROOT / "dev" / "tools"))
    from dev.tools.agent_play import _make_caller, _dp_to_obs, _do_chat
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database
    from core.signal import SignalBus
    from core.world.model import FACT, ACTOR

    caller, note = _make_caller(args.config, no_llm=False)
    print(f"[actor_consistency_reconfirm] caller={note} flags=NC+OVC ON")
    loop = BeatLoop(caller, Blackboard(), Database(tempfile.mktemp(suffix=".db")), SignalBus(),
                    run_id="actor", use_kernel=True)
    out = open(args.out, "w", encoding="utf-8")

    def emit(rec):
        out.write(json.dumps(rec, ensure_ascii=False) + "\n"); out.flush()

    res = loop.start({"theme": "午夜的廢棄海事研究站（弟弟失蹤）", "protagonist_name": "周凱", "npc_count": 2})
    emit({"kind": "opening", "head": (res.get("narrative") or "")[:120].replace("\n", " ")})
    dp = res["decision_point"]
    obs = _dp_to_obs(loop, res["narrative"], dp, False, None)

    # 確保有在場 NPC 可對話（必要時推進一拍）
    npc = (obs.get("visible_npcs") or obs.get("chat_available_npcs") or [None])[0]
    if not npc:
        o = loop.step("我四處查看，看看這裡還有沒有別人")
        dp = o.get("decision_point")
        obs = _dp_to_obs(loop, o.get("narrative"), dp, o.get("ended"), o.get("ending"), step_result=o)
        npc = (obs.get("visible_npcs") or obs.get("chat_available_npcs") or [None])[0]
    print(f"  NPC：{npc}")

    fact_before = len(loop._world.by_kind(FACT))
    reveal_before = _rp(obs)
    actor_before = [a.id for a in loop._world.by_kind(ACTOR)]

    # 1. 對話 → NPC 成為 current_focus（ensure_actor_entity 應在此建好 actor entity）
    chat_rec = {"kind": "chat", "npc": npc}
    if npc:
        reply = _do_chat(loop, {"npc": npc, "text": "你是誰？你剛才看到什麼了嗎？"})
        chat_rec["reply_head"] = reply[:100].replace("\n", " ")
    foc = loop._current_focus or {}
    foc_eid = foc.get("id")
    foc_entity = loop._world.get(foc_eid) if foc_eid else None
    chat_rec["focus_id"] = foc_eid
    chat_rec["focus_world_entity_kind"] = getattr(foc_entity, "kind", None)
    emit(chat_rec)
    print(f"  [chat] focus={foc_eid} world_entity_kind={getattr(foc_entity,'kind',None)}")

    # 2. 「那個人」指代 → 應 resolve 到 actor entity
    o = loop.step("我再問那個人，他剛才到底說了什麼")
    obs = _dp_to_obs(loop, o.get("narrative"), o.get("decision_point"), o.get("ended"),
                     o.get("ending"), step_result=o)
    er = obs.get("entity_resolution") or {}
    rid = er.get("resolved_entity_id")
    rentity = loop._world.get(rid) if rid else None
    reveal_after = _rp(obs)
    fact_after = len(loop._world.by_kind(FACT))
    rec = {
        "kind": "beat", "beat": loop.beat_number, "action": "我再問那個人，他剛才到底說了什麼",
        "action_class": o.get("action_class"),
        "entity_resolution": {"query": er.get("query"), "resolved_entity_id": rid,
                              "resolution_source": er.get("resolution_source"),
                              "ambiguous": er.get("ambiguous")},
        "resolved_world_entity_kind": getattr(rentity, "kind", None),
        "world_get_not_none": rentity is not None,
        "reveal_before": reveal_before, "reveal_after": reveal_after,
        "fact_before": fact_before, "fact_after": fact_after,
        "reveal_reward_ladder": (o.get("reveal_reward_debug") or {}).get("ladder_action"),
    }
    emit(rec)
    checks = {
        "resolved_is_actor_id": bool(rid) and str(rid).startswith("actor."),
        "resolved_kind_actor": rec["resolved_world_entity_kind"] == "actor",
        "world_get_not_none": rec["world_get_not_none"],
        "focus_maps_to_world_actor": chat_rec.get("focus_world_entity_kind") == "actor",
        "no_new_fact": fact_after == fact_before,
        "no_reveal_push": reveal_after == reveal_before,
    }
    emit({"kind": "checks", **checks})
    print(f"\n  [b{rec['beat']}] er: q={er.get('query')!r} → {rid} kind={rec['resolved_world_entity_kind']} "
          f"src={er.get('resolution_source')}")
    print("\n=== CHECKS ===")
    for k, v in checks.items():
        print(f"  {'✅' if v else '❌'} {k}: {v}")
    out.close()
    print(f"\nwritten → {args.out}")


if __name__ == "__main__":
    main()
