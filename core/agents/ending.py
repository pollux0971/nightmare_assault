"""core.agents.ending — 結局序列（UB2）。

當結局被觸發（warden 硬結局 / attractor 累積），組一段**純敘述收尾** + **完整真相揭露** + **復盤**。

設計重點：
- 結局是遊戲**唯一**可以揭露 real_bible 的時機（遊戲已結束，不再有暴雷問題）。
- 程式碼組裝為主 → 一定有結局可玩（不依賴 LLM；caller 失敗也能收尾，B8）。
- 提早觸發（玩家還沒摸清就死/逃）→ 真相揭露本身就是劇情內的「合理化解」：把沒找到的碎片一次攤開。
- 收尾後回主選單（前端負責）。
"""
from __future__ import annotations

from typing import Any

# 每種結局的收尾基調（純敘述，不再給選項）
_CLOSING = {
    "death_physical": "黑暗收攏上來，像一直在等這一刻。你的呼吸慢下來，慢到聽不見——你的故事在這裡斷了。",
    "death_mental": "有什麼在你腦海裡碎開了。你還站著，但那個走進來的「你」，已經不在了。",
    "escape": "門在你身後闔上。冷夜的空氣灌進肺裡——你活著走出來了。只是你很清楚，有些東西被你留在了裡面，有些東西被你一起帶了出來。",
    "truth_revealed": "最後一塊拼圖落定。你站在原地，終於看清這個地方究竟發生過什麼——這份明白，比恐懼更沉。",
    "transformation": "你不再害怕了。因為你已經成了這個地方的一部分——現在，換你在黑暗裡，等下一個推開門的人。",
    "death": "終結降臨，無聲無息。",
}
_TITLE = {
    "death_physical": "你死了", "death_mental": "你的心智碎了", "escape": "你逃出來了",
    "truth_revealed": "真相", "transformation": "你成了它的一部分", "death": "終結",
}

# NR3：結局表層變體的收尾基調（ambiguous/truth_locked/failed 與 clean 明顯不同）
_SURFACE_CLOSING = {
    "clean_escape": "門在你身後闔上。冷夜的空氣灌進肺裡——你活著走出來了，而且你知道自己帶走了答案。",
    "ambiguous_escape": ("你走出去了。\n\n至少，你以為自己走出去了。\n\n"
                         "冷夜的空氣灌進肺裡，但有什麼跟著你一起出來了——你說不清是什麼，"
                         "只知道身後那扇門，其實從來沒有真正關上。"),
    "truth_locked": ("你活了下來，站在還亮著的燈下。可是你心裡很清楚：你逃過了它，"
                     "卻始終沒弄懂它是什麼。有些門，你連碰都還沒碰到。"),
    "failed_escape": ("出口就在眼前，近得能聞到外面的空氣。但你跨不過去——"
                      "有什麼把你留了下來，代價比你以為的更重。"),
    "death": "黑暗收攏上來，像一直在等這一刻。你的故事在這裡斷了。",
}
_SURFACE_TITLE = {
    "clean_escape": "你逃出來了", "ambiguous_escape": "你『逃出來了』",
    "truth_locked": "你活著，但真相留在裡面", "failed_escape": "你沒能離開", "death": "終結",
}


def build_ending_sequence(blackboard: Any, ending: dict, ledger: list | None = None,
                          reveal_recap: dict | None = None) -> dict:
    """把 ending dict 補上 closing 敘述 / 完整真相 / 復盤，回傳 enriched 結局 dict。

    reveal_recap（NR0，optional）：RevealLedger 的部分進度（found/confirmed/suspected/observed/line），
    讓復盤即使 0 個 confirmed 碎片也能顯示「接觸過/懷疑/觀察」的部分發現（解 0/X 問題）。
    """
    ending = dict(ending or {})
    etype = ending.get("type") or "death"
    snap = blackboard.snapshot() if hasattr(blackboard, "snapshot") else {}
    real = snap.get("real_bible") or {}
    world = real.get("world_truth") or {}
    pool = real.get("revelation_pool") or []
    revealed = (snap.get("revealed_bible") or {}).get("revealed_fragments") or []

    discovered_ids = {f.get("id") for f in revealed if isinstance(f, dict)}

    def _entry(f, idx):
        return {"id": f.get("id", f"f{idx}"),
                "title": f.get("title") or f"未解的線索 #{idx + 1}",
                "content": f.get("content", "")}
    found = [_entry(f, i) for i, f in enumerate(pool)
             if isinstance(f, dict) and f.get("id") in discovered_ids]
    missed = [_entry(f, i) for i, f in enumerate(pool)
              if isinstance(f, dict) and f.get("id") not in discovered_ids]

    # ── 完整真相（結局唯一可露 real_bible）；render 依 mode 決定露多少 ────────
    truth = {
        "what_really_happened": world.get("what_really_happened", ""),
        "the_threat_is": world.get("the_threat_is", ""),
        "deadly_rule": world.get("deadly_rule", ""),
        "all_fragments": [f.get("content", "") for f in pool if isinstance(f, dict)],
    }

    # ── 復盤（discovered/missed 帶 id+title+content，供 masked/full render）──
    recap = {
        "found_count": len(found),
        "total_count": len(pool),
        "discovered": found,            # [{id,title,content}]
        "missed": missed,               # [{id,title,content}]
        "early": bool(missed),          # 還有沒找到的 → 視為提早觸發
    }
    # NR0：揭露帳本的部分進度（接觸過/懷疑/觀察/確認）——即使無 confirmed 也非 0/X，
    # 且明細用標題顯示分層狀態（要求 #4），不再全是 ？？？。
    if reveal_recap:
        recap["partial"] = {
            "found": int(reveal_recap.get("found", 0)),
            "confirmed": int(reveal_recap.get("confirmed", 0)),
            "suspected": int(reveal_recap.get("suspected", 0)),
            "observed": int(reveal_recap.get("observed", 0)),
            "total": int(reveal_recap.get("total", len(pool))),
            "line": reveal_recap.get("line", ""),
            "confirmed_list": reveal_recap.get("confirmed_list", []),
            "suspected_list": reveal_recap.get("suspected_list", []),   # observed/suspected
            "hinted_list": reveal_recap.get("hinted_list", []),
            "hidden_list": reveal_recap.get("hidden_list", []),
        }

    # ── NR3：結局表層變體（僅在敘事控制已標記時生效；flag OFF 行為完全不變）──
    # 觸發條件：loop 已設 escape_quality（逃脫，flag on）或外部已指定 ending_surface。
    surface = ending.get("ending_surface")
    if not surface and ending.get("escape_quality") is not None:
        try:
            from core.narrative.ending_gate import classify_ending_surface
            surface = classify_ending_surface(ending, len(discovered_ids))
        except Exception:
            surface = None

    if surface in _SURFACE_CLOSING:
        ending["ending_surface"] = surface
        ending["title"] = _SURFACE_TITLE.get(surface, _TITLE.get(etype, "結局"))
        ending["closing"] = _SURFACE_CLOSING[surface]
    else:
        ending["title"] = _TITLE.get(etype, "結局")
        ending["closing"] = _CLOSING.get(etype, _CLOSING["death"])
    ending["truth"] = truth
    ending["recap"] = recap
    ending["is_ending"] = True
    return ending


# masked 模式下，發現比例 ≥ 此值才額外揭露核心真相（鼓勵探索/重玩）。
MASKED_TRUTH_RATIO = 0.6
_REPLAY_HOOK = "有些答案，你還沒走到它面前。"


def _content(item) -> str:
    return item.get("content", "") if isinstance(item, dict) else str(item)


def _title(item) -> str:
    return item.get("title", "") if isinstance(item, dict) else "未解的線索"


def render_ending_text(ending: dict, mode: str = "masked") -> str:
    """把 enriched 結局組成純文字收尾。

    mode="masked"（預設，正式體驗）：只露已發現碎片全文；未發現只露遮罩標題 + ？？？ + 重玩鉤子；
        核心真相僅在發現比例 ≥ MASKED_TRUTH_RATIO 時才額外揭露。
    mode="full"（debug）：完整攤開（真相 + 所有碎片）。
    """
    parts = [ending.get("closing", "")]
    t = ending.get("truth") or {}
    r = ending.get("recap") or {}
    total = r.get("total_count", 0)
    found = r.get("found_count", 0)
    ratio = (found / total) if total else 0.0

    def _truth_block():
        out = ["\n── 真相 ──", f"真正發生的事：{t['what_really_happened']}"]
        if t.get("the_threat_is"):
            out.append(f"真正的威脅：{t['the_threat_is']}")
        if t.get("deadly_rule"):
            out.append(f"那條致命規則：{t['deadly_rule']}")
        return out

    if mode == "full":
        if t.get("what_really_happened"):
            parts += _truth_block()
        if total:
            parts.append("\n── 復盤 ──")
            parts.append(f"你在這場夢魘裡發現了 {found}/{total} 個真相碎片。")
            if r.get("missed"):
                parts.append("還沒來得及找到的，現在一次攤在你面前：")
                parts += [f"・{_content(m)}" for m in r["missed"]]
        return "\n".join(p for p in parts if p)

    # ── masked（預設）──
    partial = r.get("partial")
    if partial and partial.get("total"):
        # NR0（要求 #4）：用揭露帳本的分層明細——confirmed 露全文，observed/suspected/hinted 顯示標題與狀態，
        # 只有完全沒接觸的才 ？？？。
        parts.append("\n── 復盤 ──")
        parts.append(partial.get("line") or
                     f"你在這場夢魘裡發現了 {partial.get('found', 0)}/{partial['total']} 條真相線索。")
        if partial.get("confirmed_list"):
            parts.append("已確認：")
            parts += [f"・{_title(d)}：{_content(d)}" for d in partial["confirmed_list"]]
        if partial.get("suspected_list"):
            parts.append("已掌握（尚未完全確認）：")
            parts += [f"・{_title(s)}：你拼湊出了大概，但還缺最後一塊。" for s in partial["suspected_list"]]
        if partial.get("hinted_list"):
            parts.append("接觸過（線索浮現，意義不明）：")
            parts += [f"・{_title(h)}：？" for h in partial["hinted_list"]]
        if partial.get("hidden_list"):
            parts.append("從未觸及：")
            parts += [f"・？？？" for _ in partial["hidden_list"]]
    elif total:
        parts.append("\n── 復盤 ──")
        parts.append(f"你在這場夢魘裡發現了 {found}/{total} 個真相碎片。")
        if r.get("discovered"):
            parts.append("已確認：")
            parts += [f"・{_content(d)}" for d in r["discovered"]]
        if r.get("missed"):
            parts.append("未確認：")
            parts += [f"・{_title(m)}：？？？" for m in r["missed"]]
    # 探索夠多才額外揭露核心真相
    if t.get("what_really_happened") and total and ratio >= MASKED_TRUTH_RATIO:
        parts += _truth_block()
    if r.get("missed"):
        parts.append(f"\n{_REPLAY_HOOK}")
    return "\n".join(p for p in parts if p)
