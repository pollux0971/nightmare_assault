#!/usr/bin/env python3
"""alias_focus_scope_smoke — focused real-LLM 驗證 AliasResolver Focus-Scope Patch（commit d7a6614）。

只驗證、不改 core。流程：生成 object → 檢查/撿起 → 與 NPC 對話（focus→NPC）→ 連續指代：
「剛才那個東西」(應→object)、「他說的方向/地方」(應→recent fact，不新增 area/exit)、
「那個人」(應→actor)、兩本筆記本下「那本筆記本」(應→ambiguous)。全程驗 entity_resolution + 不推 reveal。

跑法：python3 dev/tools/alias_focus_scope_smoke.py --out /tmp/alias.jsonl
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
    ap.add_argument("--out", default="/tmp/alias.jsonl")
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
    from core.world.model import OBJECT, AREA, EXIT

    caller, note = _make_caller(args.config, no_llm=False)
    print(f"[alias_focus_scope_smoke] caller={note} flags=NC+OVC ON")
    loop = BeatLoop(caller, Blackboard(), Database(tempfile.mktemp(suffix=".db")), SignalBus(),
                    run_id="alias", use_kernel=True)
    out = open(args.out, "w", encoding="utf-8")

    def emit(rec):
        out.write(json.dumps(rec, ensure_ascii=False) + "\n"); out.flush()

    def kind_of(eid):
        e = loop._world.get(eid) if (eid and loop._world) else None
        return getattr(e, "kind", None)

    def step(action, label):
        o = loop.step(action)
        obs = _dp_to_obs(loop, o.get("narrative") or "", o.get("decision_point"),
                         o.get("ended"), o.get("ending"), step_result=o)
        er = obs.get("entity_resolution") or {}
        rid = er.get("resolved_entity_id")
        rec = {
            "kind": "beat", "label": label, "beat": loop.beat_number, "action": action,
            "action_class": o.get("action_class"), "reveal_gate_allowed": o.get("reveal_gate_allowed"),
            "reveal_progress": obs.get("reveal_progress"),
            "reveal_reward_ladder": (o.get("reveal_reward_debug") or {}).get("ladder_action"),
            "current_focus": (obs.get("player_state") or {}).get("current_focus"),
            "entity_resolution": {
                "query": er.get("query"), "resolved_entity_id": rid,
                "resolution_source": er.get("resolution_source"),
                "ambiguous": er.get("ambiguous"), "candidates": er.get("candidates"),
                "resolved_kind": kind_of(rid)},
            "area_count": len(loop._world.by_kind(AREA)),
            "exit_count": len(loop._world.by_kind(EXIT)),
        }
        emit(rec)
        er2 = rec["entity_resolution"]
        foc = rec["current_focus"]
        print(f"  [{label} b{rec['beat']}] cls={rec['action_class']} reveal={_rp(obs)} "
              f"focus={(foc or {}).get('kind')}:{(foc or {}).get('label')}")
        print(f"        er: q={er2['query']!r} → {er2['resolved_entity_id']} "
              f"({er2['resolved_kind']}) src={er2['resolution_source']} ambiguous={er2['ambiguous']} "
              f"cands={er2['candidates']}")
        return o, obs, rec

    res = loop.start({"theme": "午夜的廢棄海事研究站（弟弟失蹤）", "protagonist_name": "周凱", "npc_count": 1})
    emit({"kind": "opening", "head": (res.get("narrative") or "")[:120].replace("\n", " ")})
    print(f"[開場] reveal={_rp(_dp_to_obs(loop, res['narrative'], res['decision_point'], False, None))}")

    # 1+2. 生成 object → 撿起 + 端詳（focus → object）
    step("我仔細搜查這個房間，翻看桌面與抽屜，找出能辨識身分或留下線索的小東西", "gen_object")
    objs = [e for e in loop._world.by_kind(OBJECT) if e.state not in ("taken", "used")]
    obj_label = objs[0].label if objs else "東西"
    print(f"  生成物件：{[e.label for e in objs]}")
    step(f"我蹲下來，把那{obj_label}撿起來收進口袋，仔細端詳它", "take_inspect")

    # 3. 與 NPC 對話 → focus = NPC（並嘗試取得 location fact）
    obs0 = _dp_to_obs(loop, loop.last_story, None, loop.ended, loop.ending)
    npc = (obs0.get("visible_npcs") or obs0.get("chat_available_npcs") or [None])[0]
    rev_before_chat = _rp(obs0)
    npc_fact_materialized = False
    if npc:
        reply = _do_chat(loop, {"npc": npc, "text": "出口在哪個方向？那扇門是不是鎖住了？要先做什麼才能離開這裡？"})
        kf = [{"label": f.get("label"), "source": f.get("source")} for f in
              ((loop.player_state() or {}).get("known_facts") or [])]
        npc_fact_materialized = bool(kf)
        print(f"  [chat] {npc}: known_facts now={kf}")
        emit({"kind": "chat", "npc": npc, "reply_head": reply[:120].replace("\n", " "),
              "known_facts": kf, "npc_fact_materialized": npc_fact_materialized,
              "reveal_before": rev_before_chat,
              "reveal_after": _rp(_dp_to_obs(loop, loop.last_story, None, loop.ended, loop.ending))})

    # 注入一個 route-related fact（如 NPC 未自動產出 fact；純測 resolver 的 fact scope，與注入兩本筆記本同理）
    from core.world.model import FACT
    if not loop._world.get("fact.route_inject"):
        loop._world.register(FACT, "出口在東側走廊的盡頭", id="fact.route_inject",
                             props={"source": npc or "npc", "confidence": "npc_claim",
                                    "tags": ["location_claim"]}, origin="npc")
        loop.bb.game_meta = {**loop.bb.game_meta, "world_model": loop._world.to_dict()}

    # 4. 「剛才那個東西」→ 應解析到 object（非 NPC）；reference 置句尾避免尾隨動詞污染 noun
    area_b, exit_b = len(loop._world.by_kind(AREA)), len(loop._world.by_kind(EXIT))
    _, _, r4 = step("我確認自己沒看錯，再看一下剛才那個東西", "obj_ref_under_npc_focus")

    # 5. 「他說的方向 / 他說的地方」→ 應解析到 fact（非導航動作，避免 kernel 改 area）
    _, _, r5 = step("我在原地仔細回想他說的方向、他說的地方", "fact_ref")

    # 6. 「那個人」→ 應解析到 actor
    _, _, r6 = step("我想再問清楚，於是轉頭看向那個人", "actor_ref")

    # 7. 注入兩本同類筆記本 → 「那本筆記本」應 ambiguous（reference 置句尾）
    cur = loop._world.current_area
    loop._world.register(OBJECT, "紅色筆記本", id="object.nb_red", props={"area": cur})
    loop._world.register(OBJECT, "黑色筆記本", id="object.nb_black", props={"area": cur})
    loop.bb.game_meta = {**loop.bb.game_meta, "world_model": loop._world.to_dict()}
    print("  注入兩本筆記本（紅/黑）測 ambiguous")
    _, _, r7 = step("我的視線落在那本筆記本", "ambiguous_two_notebooks")

    # ── 檢查彙整 ──
    def _hinted(b):
        return (b["reveal_progress"] or {}).get("hinted", 0)
    alias_beats = [r4, r5, r6, r7]
    checks = {
        "4_object_ref_resolves_object_not_npc":
            r4["entity_resolution"]["resolved_kind"] == "object",
        "5_fact_ref_resolves_fact":
            r5["entity_resolution"]["resolved_kind"] == "fact",
        "5_no_new_area_exit":
            r5["area_count"] == area_b and r5["exit_count"] == exit_b,
        "6_actor_ref_resolves_actor":
            r6["entity_resolution"]["resolved_kind"] == "actor",
        "7_two_notebooks_ambiguous":
            bool(r7["entity_resolution"]["ambiguous"])
            and r7["entity_resolution"]["resolved_entity_id"] is None,
        "8_alias_beats_not_truth_investigation":
            all(b["action_class"] != "truth_investigation" for b in alias_beats),
        "8_reveal_not_pushed_on_alias_beats":
            all(_hinted(b) == 0 for b in alias_beats),
        "er_debug_fields_present":
            all(k in r4["entity_resolution"]
                for k in ("query", "resolved_entity_id", "resolution_source", "ambiguous", "candidates")),
        "npc_fact_materialized_live": npc_fact_materialized,
    }
    emit({"kind": "checks", **checks})
    print("\n=== CHECKS ===")
    for k, v in checks.items():
        print(f"  {'✅' if v else '❌'} {k}: {v}")

    out.close()
    print(f"\nwritten → {args.out}")


if __name__ == "__main__":
    main()
