"""core.events — Event 抽取（U11，程式碼層，零 LLM）。

不完全信 story 自報的 revelations_touched：用規則+關鍵詞比對玩家輸入與敘事，
抽出結構化事件供 orchestrator 佐證；衝突時信程式碼抽取（07 §四、CHECKLIST C12）。
"""
from __future__ import annotations

SEARCH_KW = ["搜索", "搜查", "翻找", "搜", "翻", "查看", "找"]
QUESTION_KW = ["質問", "詢問", "質疑", "問"]
PICK_KW = ["拾起", "撿起", "拿起", "撿", "拿", "取走", "取", "收起"]
REACH_KW = ["前往", "走到", "走進", "進入", "抵達", "去"]
REACH_STORY_KW = ["抵達", "來到", "走進", "踏入"]

_PUNCT = set("，。！？、；：「」『』（）()[]…\n\t .,!?\"'")


def _find_known(text: str, known) -> list[str]:
    return [k for k in (known or []) if k in text]


def _target_after(text: str, kws: list[str]):
    """啟發式：取動詞關鍵詞後到標點為止的短語當 target。"""
    best = None
    best_pos = len(text) + 1
    for kw in kws:
        idx = text.find(kw)
        if idx == -1 or idx >= best_pos:
            continue
        rest = text[idx + len(kw):]
        t = ""
        for ch in rest:
            if ch in _PUNCT:
                break
            t += ch
            if len(t) >= 6:
                break
        t = t.strip()
        if t:
            best, best_pos = t, idx
    return best


def _category(pd, so, kws, known, etype, key, story_kws=None):
    triggered = any(k in pd for k in kws) or bool(story_kws and any(k in so for k in story_kws))
    if not triggered:
        return []
    text = pd + " " + so
    targets = _find_known(text, known)
    if not targets:
        t = _target_after(pd, kws)
        if t:
            targets = [t]
    return [{"type": etype, key: tg} for tg in targets]


def event_extract(player_decision: str, story_output: str,
                  known_npcs=None, known_locations=None, known_items=None) -> list[dict]:
    """抽出結構化事件 list；去重；無可抽取回 []。"""
    pd = player_decision or ""
    so = story_output or ""
    events: list[dict] = []

    for ev in (
        _category(pd, so, QUESTION_KW, known_npcs, "questioned_npc", "npc"),
        _category(pd, so, SEARCH_KW, known_locations, "searched_location", "target"),
        _category(pd, so, PICK_KW, known_items, "picked_item", "item"),
        _category(pd, so, REACH_KW, known_locations, "reached_location", "location",
                  story_kws=REACH_STORY_KW),
    ):
        for e in ev:
            if e not in events:
                events.append(e)
    return events


def merge_with_self_report(extracted: list[dict], story_revelations_touched) -> dict:
    """取聯集；衝突時以程式碼 extracted 為準（C12）。

    回傳 {events: 程式碼抽取(權威), touched: 聯集訊號, authoritative: "code_extract"}。
    """
    code_targets = []
    for e in extracted:
        tg = e.get("target") or e.get("location") or e.get("item") or e.get("npc")
        if tg:
            code_targets.append(tg)
    union: list[str] = []
    for x in list(story_revelations_touched or []) + code_targets:
        if x not in union:
            union.append(x)
    return {"events": list(extracted), "touched": union, "authoritative": "code_extract"}
