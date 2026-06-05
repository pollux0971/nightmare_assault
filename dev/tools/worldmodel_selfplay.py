#!/usr/bin/env python3
"""worldmodel_selfplay — WorldModel 專用 selfplay（不測文筆，專測「世界實體是否可重訪」）。

劇本（flag ON、真 LLM）：
  beat1  搜查 → 讓 story 前景化一個可互動 object（袖扣/筆記本/照片/控制台碎片…）
  beat2  自由輸入檢查同一物件 → 確認 resolver 對到**同一 entity**、state 升級
  chat   與在場 NPC 對話，誘導其說出一條 fact（通訊設備在機房 / 某出口鎖死）
  beat3  引用該 fact（「我根據他說的機房線索尋找通訊設備」）
  beat4  「先退到外面整理線索，不結束本次調查」→ current_area 應穩定在 safe zone、ended=false
  beat5  續留外面再觀察一次 → current_area 應仍為 safe zone（不被 kernel scene sync 覆蓋）

輸出：每步的 world_progress 投影 + 期末報告（登記/重訪/NPC fact/reveal 誤推/current_area/污染）。

用法：
  .venv/bin/python dev/tools/worldmodel_selfplay.py --jsonl-log dev/reports/wm-selfplay.jsonl
"""
from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))


def _wm_snapshot(loop) -> dict:
    """直接讀 WorldModel 的權威狀態（不經觀測投影，作為事實核對）。"""
    w = getattr(loop, "_world", None)
    if w is None:
        return {"current_area": None, "entities": []}
    return {
        "current_area": w.current_area,
        "entities": [{"id": e.id, "kind": e.kind, "label": e.label, "state": e.state,
                      "origin": e.origin, "source": e.props.get("source"),
                      "confidence": e.props.get("confidence")}
                     for e in w.entities.values()],
    }


def _wp_compact(wp: dict) -> dict:
    """把 observation.world_progress 摘成報告要看的七個欄位。"""
    wm = wp.get("world_model") or {}
    def labels(items):
        return [f'{e.get("label")}[{e.get("kind")}:{e.get("state")}]' for e in (items or [])]
    return {
        "current_area": wp.get("current_area"),
        "entities_here": labels(wm.get("entities_here")),
        "interactables_here": labels(wm.get("interactables_here")),
        "changed_entities_this_beat": wp.get("changed_entities_this_beat"),
        "fact_entities": [e.get("label") for e in (wm.get("entities") or [])
                          if e.get("kind") == "fact"],
        "world_facts": wp.get("world_facts"),
        "affordances_here": [f'{a.get("verb")}:{a.get("label")}'
                             for a in (wm.get("affordances_here") or [])],
    }


def main(argv=None):
    ap = argparse.ArgumentParser()
    ap.add_argument("--theme", default="午夜的廢棄海事研究站（弟弟林晨失蹤）")
    ap.add_argument("--name", default="周凱")
    ap.add_argument("--npc-count", type=int, default=2)
    ap.add_argument("--config", default=str(ROOT / "config" / "config.json"))
    ap.add_argument("--jsonl-log", default=str(ROOT / "dev" / "reports" / "wm-selfplay.jsonl"))
    args = ap.parse_args(argv)

    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"   # 本測必須 flag ON
    from dev.tools.agent_play import _make_caller, _new_loop, _reveal_progress, _do_chat

    caller, note = _make_caller(args.config, no_llm=False)
    loop = _new_loop(caller)
    log = open(args.jsonl_log, "w", encoding="utf-8")

    def emit(rec):
        log.write(json.dumps(rec, ensure_ascii=False) + "\n"); log.flush()
        return rec

    P = print
    timeline = []          # 每步 (label, world_progress_compact, wm_snapshot)
    reveal_trace = []      # reveal_progress 軌跡（含 NPC chat 前後）

    def record(label, wp):
        wpc = _wp_compact(wp or {})
        wm = _wm_snapshot(loop)
        rp = _reveal_progress(loop)
        timeline.append({"label": label, "wp": wpc, "wm": wm})
        reveal_trace.append({"label": label, "reveal": rp})
        emit({"type": "step", "label": label, "world_progress": wpc,
              "world_model": wm, "reveal_progress": rp})
        P("\n" + "=" * 72)
        P(f"▶ {label}")
        P(f"  current_area      : {wpc['current_area']}")
        P(f"  entities_here     : {wpc['entities_here']}")
        P(f"  interactables_here: {wpc['interactables_here']}")
        P(f"  changed_this_beat : {wpc['changed_entities_this_beat']}")
        P(f"  fact_entities     : {wpc['fact_entities']}")
        P(f"  affordances_here  : {wpc['affordances_here']}")
        P(f"  reveal_progress   : {rp}")
        return wpc, wm

    # ── 開場 ─────────────────────────────────────────────────────────────────
    P(f"[caller={note}  flag=ON]")
    res = loop.start({"theme": args.theme, "protagonist_name": args.name,
                      "npc_count": args.npc_count})
    P("【開場】" + (res.get("narrative") or "")[:200])
    record("opening", loop.world_progress(res["decision_point"]))

    def step(label, action):
        emit({"type": "action", "label": label, "action": action})
        out = loop.step(action)
        P(f"\n〔story〕{(out.get('narrative') or '')[:220]}")
        wpc, wm = record(label, out.get("world_progress"))
        return out, wpc, wm

    # ── beat1：搜查，誘導 story 前景化一個可互動 object ──────────────────────────
    step("beat1_search",
         "我仔細搜查這個房間：翻看桌面、抽屜和地上，想找出任何能辨識身分或留下線索的小東西"
         "（例如證件、筆記本、照片、袖扣之類）。我只觀察與拾取，不離開房間。")

    # 找出第一個被登記的 object 當作「待重訪」目標
    def first_object():
        w = loop._world
        objs = [e for e in w.entities.values() if e.kind == "object"]
        return objs[0] if objs else None

    obj = first_object()
    obj_label = obj.label if obj else None
    obj_id_before = obj.id if obj else None
    obj_state_before = obj.state if obj else None
    P(f"\n[偵測] 第一個 object entity = {obj_label!r} (id={obj_id_before}, state={obj_state_before})")

    # ── beat2：自由輸入檢查同一物件 → 應對到同一 entity、state 升級 ───────────────
    if obj_label:
        step("beat2_inspect_same_object",
             f"我蹲下來，把那個{obj_label}拿到眼前，仔細地翻看它的每一面，研究上面的細節。")
        obj_after = loop._world.find(obj_label, kind="object")
        revisit_ok = bool(obj_after) and obj_after.id == obj_id_before
        revisit_state = obj_after.state if obj_after else None
    else:
        # story 沒給可互動 object（連 extractor 也沒抓到）→ 用通用檢查動作再試一次
        step("beat2_inspect_generic",
             "我把剛才注意到的那件小東西拿到眼前，仔細檢查它的每一面與上面的痕跡。")
        revisit_ok, revisit_state = None, None

    # ── chat：與在場 NPC 對話，誘導 fact（通訊設備在機房 / 出口狀態）─────────────
    present = [n.get("name") for n in loop.bb.snapshot().get("npc_registry") or []
               if isinstance(n, dict) and n.get("presence", "present") == "present"]
    npc = present[0] if present else None
    reveal_before_chat = _reveal_progress(loop)
    wm_facts_before_chat = [e for e in _wm_snapshot(loop)["entities"] if e["kind"] == "fact"]
    chat_reply = None
    if npc:
        emit({"type": "chat_action", "npc": npc})
        chat_reply = _do_chat(loop, {
            "npc": npc,
            "text": "這裡的通訊設備在哪？還有，往外的出口現在還走得通嗎？"
                    "如果你知道確切位置或哪條路被封死了，直接告訴我。"})
        P(f"\n〔chat {npc}〕{(chat_reply or '')[:200]}")
        P(f"〔npc_chat_debug〕{getattr(loop, '_npc_chat_debug', {})}")
        record(f"after_chat_{npc}", loop.world_progress())
    reveal_after_chat = _reveal_progress(loop)
    wm_facts_after_chat = [e for e in _wm_snapshot(loop)["entities"] if e["kind"] == "fact"]
    npc_fact_new = [f for f in wm_facts_after_chat
                    if f["id"] not in {b["id"] for b in wm_facts_before_chat}]
    npc_origin_fact = [f for f in wm_facts_after_chat if f.get("origin") == "npc"]

    # ── beat3：引用該 fact ───────────────────────────────────────────────────
    step("beat3_reference_fact",
         "我根據他剛才說的、關於機房和通訊設備的線索，動身去找那台通訊設備。")

    # ── beat4：撤到外面整理（不結束）→ current_area 應為 safe zone、ended=false ───
    out4, wp4, wm4 = step("beat4_withdraw_outside",
                          "我先退到外面整理線索，喘口氣，但我不結束這次調查。")
    withdraw_area = wp4["current_area"]
    withdraw_ended = bool(out4.get("ended"))

    # ── beat5：續留外面再觀察 → current_area 應仍為 safe zone（不被 kernel 覆蓋）────
    out5, wp5, wm5 = step("beat5_stay_outside",
                          "我站在外面的空地上，背靠著牆，把手裡的線索重新理一遍，暫時不進去。")
    stay_area = wp5["current_area"]
    stay_ended = bool(out5.get("ended"))

    # ── 期末報告 ───────────────────────────────────────────────────────────────
    final_wm = _wm_snapshot(loop)
    by_kind = {}
    for e in final_wm["entities"]:
        by_kind[e["kind"]] = by_kind.get(e["kind"], 0) + 1

    from core.world.model import SAFE_ZONE_AREA_ID as SAFE
    report = {
        "registered_entities": [f'{e["label"]}[{e["kind"]}:{e["state"]}|{e["origin"]}]'
                                for e in final_wm["entities"]],
        "entity_counts_by_kind": by_kind,
        "object_revisit": {
            "target_label": obj_label,
            "same_entity_on_revisit": revisit_ok,
            "state_before": obj_state_before,
            "state_after_revisit": revisit_state,
        },
        "npc_fact": {
            "npc": npc,
            "reply_excerpt": (chat_reply or "")[:160],
            "new_fact_entities_after_chat": [f["label"] for f in npc_fact_new],
            "npc_origin_fact_entities": [
                {"label": f["label"], "source": f["source"], "confidence": f["confidence"]}
                for f in npc_origin_fact],
            "npc_chat_debug": getattr(loop, "_npc_chat_debug", {}),
        },
        "reveal_not_pushed_by_npc": {
            "before_chat": reveal_before_chat,
            "after_chat": reveal_after_chat,
            "unchanged": reveal_before_chat == reveal_after_chat,
        },
        "current_area_stability": {
            "withdraw_area": withdraw_area, "withdraw_ended": withdraw_ended,
            "stay_area": stay_area, "stay_ended": stay_ended,
            "stable_safe_zone": withdraw_area == stay_area == SAFE
                                and not withdraw_ended and not stay_ended,
        },
        "pollution_check": {
            "total_entities": len(final_wm["entities"]),
            "object_count": by_kind.get("object", 0),
            "fact_count": by_kind.get("fact", 0),
            "actor_count": by_kind.get("actor", 0),
            # NPC 不得新增 area/exit → 來自 npc 的 area/exit 應為 0
            "npc_added_area_or_exit": [e["label"] for e in final_wm["entities"]
                                       if e["origin"] == "npc" and e["kind"] in ("area", "exit")],
        },
        "reveal_trace": reveal_trace,
    }
    emit({"type": "report", "report": report})
    P("\n" + "#" * 72)
    P("【WorldModel selfplay 報告】")
    P(json.dumps(report, ensure_ascii=False, indent=2))
    log.close()
    P(f"\n[jsonl → {args.jsonl_log}]")


if __name__ == "__main__":
    main()
