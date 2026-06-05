#!/usr/bin/env python3
"""spatial_ux_selfplay — 評估 spatial_summary 是否改善探索體驗（真 LLM、flag ON）。

「AI 玩家」**依 spatial_summary 做決策**（看可走/被阻/退路選下一步），逐 beat 自動評估：
  1. 看得懂自己在哪（目前位置 vs 實際 current_area）
  2. 看得懂可走/被阻路線（summary 段 vs 投影 routes/blocked）
  3. known_remote 沒被誤認成眼前物（眼前段不得含 remote 物件）
  4. review_mode 下是否清楚「在整理不是在調查」（current_area role + investigation_state）
  5. summary 是否過長 / 與 narrative 重複（字數 + 重疊率）
  6. summary 是否有不存在的地圖元素（每個具名項都要能對到 WorldModel 標籤）

用法：.venv/bin/python dev/tools/spatial_ux_selfplay.py
"""
from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))


def _seg(summary: str, header: str) -> str:
    """取 summary 裡某段（「header：...」）的右側內容（到行尾）。"""
    for line in (summary or "").split("\n"):
        if line.startswith(header):
            return line[len(header):].strip()
    return ""


def _named_items(seg_text: str) -> list:
    """把「A、B（鎖住）」拆成 ['A','B']（去掉狀態括號、空白）。"""
    if not seg_text or seg_text.startswith("（"):
        return []
    out = []
    for it in re.split(r"[、，,]", seg_text):
        it = re.sub(r"（[^）]*）", "", it).strip().replace(" ", "")
        if it:
            out.append(it)
    return out


def _valid_labels(world) -> set:
    s = set()
    for e in world.entities.values():
        if e.label:
            s.add(e.label.replace(" ", ""))
    return s


def _overlap_ratio(a: str, b: str) -> float:
    """summary 與 narrative 的字 bigram 重疊率（粗估重複度）。"""
    def grams(t):
        t = re.sub(r"\s", "", t or "")
        return {t[i:i+2] for i in range(len(t) - 1)}
    ga, gb = grams(a), grams(b)
    if not ga:
        return 0.0
    return len(ga & gb) / len(ga)


def main():
    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    from dev.tools.agent_play import _make_caller, _new_loop

    caller, note = _make_caller(str(ROOT / "config" / "config.json"), no_llm=False)
    loop = _new_loop(caller)
    log = open(ROOT / "dev" / "reports" / "spatial-ux-selfplay.jsonl", "w", encoding="utf-8")
    P = print
    P(f"[caller={note}  flag=ON]")
    res = loop.start({"theme": "午夜的廢棄海事研究站（弟弟林晨失蹤）",
                      "protagonist_name": "周凱", "npc_count": 2})

    records = []

    def snap(label, narrative, action, reason, out):
        sd = (out or {}).get("spatial_debug") or loop.spatial_debug()
        summary = sd.get("spatial_summary", "")
        wp = (out or {}).get("world_progress") or loop.world_progress()
        world = loop._world
        # ── 評估 ──
        pos_seg = _seg(summary, "目前位置：")
        cur_label = sd.get("current_area_label", "")
        eval1_pos_ok = bool(pos_seg) and pos_seg == cur_label
        route_items = _named_items(_seg(summary, "可走路線："))
        blocked_items = _named_items(_seg(summary, "被阻擋路線："))
        proj_routes = {r["label"].replace(" ", "") for r in sd.get("routes_from_here", [])}
        proj_blocked = {r["label"].replace(" ", "") for r in sd.get("blocked_routes", [])}
        eval2_routes_ok = (set(route_items) <= proj_routes) and (set(blocked_items) <= proj_blocked)
        # known_remote 不得出現在「眼前可互動物」
        visible_seg = _named_items(_seg(summary, "眼前可互動物："))
        remote_labels = {e["label"].replace(" ", "") for e in sd.get("known_remote_entities", [])}
        eval3_no_remote_in_visible = not (set(visible_seg) & remote_labels)
        # review 清晰度
        inv = wp.get("investigation_state")
        roles = sd.get("current_area_roles", [])
        review = inv in ("review_mode", "temporary_retreat")
        eval4_review_clear = (not review) or ("safe_zone" in roles)
        # phantom：每個具名項要能對到 WorldModel 標籤
        valid = _valid_labels(world)
        all_named = (route_items + blocked_items + visible_seg
                     + _named_items(_seg(summary, "安全撤退路線："))
                     + _named_items(_seg(summary, "已知但不在眼前：")))
        phantoms = [it for it in all_named
                    if not any(it in v or v in it for v in valid)]
        eval6_no_phantom = not phantoms
        # 長度 / 重疊
        length = len(summary)
        overlap = _overlap_ratio(summary, narrative)

        rec = {
            "label": label, "action": action, "reason_from_summary": reason,
            "current_area": sd.get("current_area"), "investigation_state": inv,
            "current_area_roles": roles, "spatial_summary": summary,
            "narrative_excerpt": (narrative or "")[:120],
            "eval": {
                "1_position_understandable": eval1_pos_ok,
                "2_routes_blocked_correct": eval2_routes_ok,
                "3_remote_not_in_visible": eval3_no_remote_in_visible,
                "4_review_mode_clear": eval4_review_clear,
                "5_length_chars": length, "5_narrative_overlap": round(overlap, 3),
                "6_no_phantom_elements": eval6_no_phantom, "phantoms": phantoms,
            },
        }
        records.append(rec)
        log.write(json.dumps(rec, ensure_ascii=False) + "\n"); log.flush()
        P("\n" + "=" * 72)
        P(f"▶ {label}  動作：{action}")
        P(f"  〔決策理由（讀 summary）〕{reason}")
        P(f"  〔narrative〕{(narrative or '')[:90]}")
        P(f"  〔spatial_summary〕\n    " + summary.replace("\n", "\n    "))
        P(f"  〔eval〕pos_ok={eval1_pos_ok} routes_ok={eval2_routes_ok} "
          f"remote_not_visible={eval3_no_remote_in_visible} review_clear={eval4_review_clear} "
          f"len={length} overlap={overlap:.2f} no_phantom={eval6_no_phantom} phantoms={phantoms}")
        return summary, sd

    # ── 開場 ─────────────────────────────────────────────────────────────────
    summary, sd = snap("opening", res.get("narrative"), "(start)", "開場觀察 summary",
                       {"spatial_debug": loop.spatial_debug(),
                        "world_progress": loop.world_progress(res["decision_point"])})

    def step(label, action, reason):
        out = loop.step(action)
        return snap(label, out.get("narrative"), action, reason, out)

    # beat1：先搜查讓世界登記實體
    summary, sd = step("beat1_search",
                       "我仔細搜查這個房間，翻看桌面、抽屜與地上，找任何能辨識身分或留下線索的小東西",
                       "summary 顯示眼前無可互動物，先搜查以發現物件")

    # beat2：依 summary「可走路線」決定往哪走（沒有就再探索）
    routes = [r["label"] for r in sd.get("routes_from_here", [])]
    if routes:
        act2, why2 = f"我走向{routes[0]}", f"summary 列出可走路線「{routes[0]}」，選它前進"
    else:
        act2, why2 = ("我往房間深處走，尋找可以離開這裡的出口",
                      "summary 顯示『沒有明顯可走的出口』，主動探索找路")
    summary, sd = step("beat2_route", act2, why2)

    # beat3：若 summary 有被阻擋路線就去查它；否則引用 known_remote
    blocked = [r["label"] for r in sd.get("blocked_routes", [])]
    remote = [e["label"] for e in sd.get("known_remote_entities", [])]
    if blocked:
        act3, why3 = (f"我檢查{blocked[0]}，看看它為什麼過不去",
                      f"summary 的被阻擋路線有「{blocked[0]}」，去查阻礙原因")
    elif remote:
        act3, why3 = (f"我去找{remote[0]}",
                      f"summary『已知但不在眼前』有「{remote[0]}」，前往該處")
    else:
        act3, why3 = "我繼續往前查看四周", "summary 無被阻路線/remote，繼續探索"
    summary, sd = step("beat3_blocked_or_remote", act3, why3)

    # beat4：撤退到安全區整理（測 review_mode summary 清晰度）
    summary, sd = step("beat4_withdraw",
                       "先退到外面整理線索，不結束本次調查",
                       "依 summary 安全撤退/退到外圍整理，進 review_mode")

    # beat5：返回現場
    summary, sd = step("beat5_return",
                       "我返回現場，繼續調查",
                       "整理完，summary 提示要返回現場 → 重新進入 active")

    # ── 期末總結 ───────────────────────────────────────────────────────────────
    def rate(key, pred):
        return sum(1 for r in records if pred(r["eval"])) , len(records)
    p1 = rate("", lambda e: e["1_position_understandable"])
    p2 = rate("", lambda e: e["2_routes_blocked_correct"])
    p3 = rate("", lambda e: e["3_remote_not_in_visible"])
    p4 = rate("", lambda e: e["4_review_mode_clear"])
    p6 = rate("", lambda e: e["6_no_phantom_elements"])
    lens = [r["eval"]["5_length_chars"] for r in records]
    overlaps = [r["eval"]["5_narrative_overlap"] for r in records]
    summary_report = {
        "beats": len(records),
        "1_position_ok": f"{p1[0]}/{p1[1]}",
        "2_routes_ok": f"{p2[0]}/{p2[1]}",
        "3_remote_not_in_visible": f"{p3[0]}/{p3[1]}",
        "4_review_clear": f"{p4[0]}/{p4[1]}",
        "5_len_max": max(lens), "5_len_avg": round(sum(lens) / len(lens), 1),
        "5_overlap_max": max(overlaps), "5_overlap_avg": round(sum(overlaps) / len(overlaps), 3),
        "6_no_phantom": f"{p6[0]}/{p6[1]}",
        "all_phantoms": [p for r in records for p in r["eval"]["phantoms"]],
    }
    log.write(json.dumps({"type": "summary", "report": summary_report}, ensure_ascii=False) + "\n")
    log.close()
    P("\n" + "#" * 72)
    P("【Spatial UX selfplay 總結】")
    P(json.dumps(summary_report, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
