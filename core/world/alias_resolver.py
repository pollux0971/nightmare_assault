"""core.world.alias_resolver — 自然指代 → WorldModel entity（Step 6；唯讀解析）。

讓「那枚 / 剛才那個 / 他說的地方 / 那個 NPC / 這東西」對到正確的既有 entity，
**不新增 entity、不推 reveal、不改世界狀態**。無 embedding、無 LLM、無大型同義詞庫。

解析順序：explicit id → exact label/alias → (fact-ref→known_facts) → (npc-ref→actor) →
current_focus → recent_entities → visible → inventory → unresolved；候選平手 → ambiguous（不亂選）。
"""
from __future__ import annotations

import re
import unicodedata

# 弱修飾/指示詞（從 query 剝除後剩「名詞」）——長詞優先
_WEAK = sorted([
    "剛才那個人", "剛剛那個人", "剛才那個", "剛剛那個", "之前那個", "剛才那條", "剛才那件",
    "剛才的那", "那一個", "這一個", "那名", "這名", "那位", "這位", "那枚", "這枚", "那本",
    "這本", "那件", "這件", "那條", "這條", "那張", "這張", "那把", "這把", "那台", "這台",
    "那東西", "這東西", "那玩意", "這玩意", "那傢伙", "這傢伙", "剛才的", "剛剛的", "之前的",
    "方才的", "那個", "這個", "那", "這", "剛才", "剛剛", "之前", "方才", "我的", "手裡的",
], key=len, reverse=True)

# query 類型線索
_NPC_CUE = ("npc", "人", "傢伙", "那位", "這位", "那名", "那傢伙")
_FACT_REF_CUE = ("說的", "講的", "提到的", "提及的", "說過的", "提的")
_FACT_NOUN = ("地方", "線索", "主張", "消息", "情報", "事", "出口", "機房", "位置", "方向", "路")
_PAST_CUE = ("剛才", "剛剛", "之前", "方才", "先前")
# object 指代線索（指向「物」而非「人/事」）——含泛稱物名 + 量詞式指示
_OBJECT_CUE = ("東西", "玩意", "物件", "那枚", "這枚", "那件", "這件", "那條", "這條",
               "那張", "這張", "那把", "這把", "那台", "這台", "那本", "這本", "那盞", "這盞")
# 泛稱物名（非具體 label）：noun 落在這裡 → 視為「無具體名詞」，走 object scope fallback
_GENERIC_OBJECT_NOUN = {"東西", "玩意", "物件"}
# scope → entity kind
_SCOPE_KIND = {"object": "object", "actor": "actor", "fact": "fact"}
# ── extract_reference：強指代 anchor + bounded noun capture（不被弱所有格/動詞污染）──────
# fact 歸屬 anchor（「他說的…」整段含主語）——之後抓 bounded noun
_FACT_ANCHORS = ("他說的", "她說的", "它說的", "他講的", "她講的", "他提到的", "她提到的",
                 "他說過的", "她說過的", "他提的", "她提的", "他提及的", "她提及的")
# 強指代 anchor（句中標示「這是個指代」的起點）——**不含**「我的/手裡的」這類弱所有格修飾。
# 含量詞者排前（長詞優先，同位置取最長）；bare「那/這」另做副詞守門（見 extract_reference）。
_DEMONSTRATIVE_ANCHORS = sorted([
    "剛才那", "剛剛那", "之前那", "方才那", "先前那",
    "那本", "這本", "那枚", "這枚", "那個", "這個", "那件", "這件", "那條", "這條",
    "那張", "這張", "那把", "這把", "那台", "這台", "那盞", "這盞", "那名", "這名",
    "那位", "這位", "那傢伙", "這傢伙", "那東西", "這東西", "那玩意", "這玩意", "那些", "這些",
], key=len, reverse=True)
# bare「那/這」後接這些字 → 視為副詞/方位（那裡/這時…），**不**當實體指代
_ADVERBIAL_AFTER = set("裡裏時邊麼樣兒會種頭天年月")
# noun 片段的停止字元（碰到動詞/助詞/標點/空白就停）——bounded noun capture。
# **刻意排除**會出現在名詞內部的字（向/對/把/從/往/給：方向/對講機/把手…），避免誤切名詞。
_NOUN_STOP = set("看瞧盯望拿撿拾取收放檢視查翻找走去到在就還想說講問摸碰握抓推拉搬扛背"
                 "是有被將也都很太更最，。、！？!?,.；;：:「」『』（）()【】"
                 "［］\"'`·•|/\\ \t\n的了著嗎呢吧啊喔欸")


def normalize_label(s: str) -> str:
    """正規化：全→半形（NFKC）、去空白與標點。**不做**繁簡轉換（保守）/ embedding。"""
    s = unicodedata.normalize("NFKC", s or "")
    s = re.sub(r"\s+", "", s)
    s = re.sub(r"[，。、！？!?,.\-—…「」『』（）()\[\]【】:：；;\"'`·•|/\\]", "", s)
    return s.lower()


def _strip_weak(nq: str) -> str:
    """從正規化 query 剝除弱修飾/指示詞，留下名詞核心。"""
    out = nq
    for w in _WEAK:
        out = out.replace(normalize_label(w), "")
    return out


def entity_alias_set(e) -> set:
    """某 entity 的可匹配別名集合（label + props.aliases，皆正規化）。"""
    s = {normalize_label(e.label)}
    for a in (e.props.get("aliases") or []):
        if a:
            s.add(normalize_label(a))
    s.discard("")
    return s


def add_alias(world, eid: str, alias: str) -> bool:
    """給既有 entity 加別名（存 props.aliases）。**不新增 entity**；空/重複不加。"""
    if world is None or not alias:
        return False
    e = world.get(eid)
    if e is None:
        return False
    cur = list(e.props.get("aliases") or [])
    if normalize_label(alias) in {normalize_label(a) for a in cur} or normalize_label(alias) == normalize_label(e.label):
        return False
    cur.append(alias)
    e.props["aliases"] = cur
    return True


def _capture_noun(t: str, start: int, maxlen: int = 6) -> str:
    """從 start 抓 bounded 名詞片段：連續 CJK 字，碰到停止字元（動詞/助詞/標點）即止；上限 maxlen。"""
    out = []
    for ch in t[start:start + maxlen + 2]:
        if ch in _NOUN_STOP or not ("一" <= ch <= "鿿"):
            break
        out.append(ch)
        if len(out) >= maxlen:
            break
    return "".join(out)


def extract_reference(action: str):
    """從整句玩家輸入抽出**最小可解析的指代片語**（強指代 anchor + bounded noun）。找不到 → None。

    不再用「最早出現的弱觸發詞 + 定長視窗」（會被「我的/視線落在…」污染整句）；改為：
    找最早出現的**強指代 anchor**（那本/那枚/那個/他說的/那名…，**不含** 我的/手裡的），同位置取最長，
    再抓 anchor 後的 bounded 名詞片段（碰動詞/標點即止）。bare「那/這」需後接實體名詞、非副詞（那裡/這時排除）。
    """
    t = action or ""
    if not t:
        return None
    hits: list = []                                      # (position, anchor)
    for a in _FACT_ANCHORS:
        i = t.find(a)
        if i != -1:
            hits.append((i, a))
    for a in _DEMONSTRATIVE_ANCHORS:
        i = t.find(a)
        if i != -1:
            hits.append((i, a))
    # bare「那/這」：只在後接「實體名詞字」（CJK 且非副詞/非量詞已被長 anchor 蓋掉）時才當 anchor
    for a in ("那", "這"):
        i = -1
        while True:
            i = t.find(a, i + 1)
            if i == -1:
                break
            nxt = t[i + 1] if i + 1 < len(t) else ""
            if nxt and "一" <= nxt <= "鿿" and nxt not in _ADVERBIAL_AFTER:
                hits.append((i, a))
                break
    if not hits:
        return None
    earliest = min(h[0] for h in hits)
    anchor = max((a for (i, a) in hits if i == earliest), key=len)   # 同位置取最長 anchor
    noun = _capture_noun(t, earliest + len(anchor))
    return anchor + noun


def _match_by_label(world, noun: str) -> list:
    """名詞 → 候選 entity id（label/alias 正規化雙向子字串）。noun 空 → []。"""
    if world is None or not noun:
        return []
    out = []
    for e in world.entities.values():
        for a in entity_alias_set(e):
            if a and (a == noun or noun in a or a in noun):
                out.append(e.id)
                break
    return out


def _first_of_kind(items: list, kind: str):
    for it in (items or []):
        if it.get("kind") == kind:
            return it.get("id"), it.get("label")
    return None, None


def resolve_entity_reference(query: str, *, world=None, current_focus=None,
                             recent_entities=None, visible_entities=None,
                             inventory_entities=None, known_facts=None) -> dict:
    """把自然指代解析成一個 entity id。回 entity_resolution dict（唯讀；不建 entity、不推 reveal）。"""
    recent = recent_entities or []
    vis = visible_entities or []
    inv = inventory_entities or []
    kf = known_facts or []

    def res(eid, src):
        return {"query": query, "resolved_entity_id": eid, "resolution_source": src,
                "ambiguous": False, "candidates": []}

    def ambiguous(cands):
        return {"query": query, "resolved_entity_id": None, "resolution_source": "ambiguous",
                "ambiguous": True, "candidates": list(cands)}

    unresolved = {"query": query, "resolved_entity_id": None, "resolution_source": "unresolved",
                  "ambiguous": False, "candidates": []}
    if not query:
        return unresolved

    # 1. explicit entity id
    if world is not None and query in world.entities:
        return res(query, "explicit_id")

    nq = normalize_label(query)
    is_past = any(c in query for c in _PAST_CUE)
    is_npc = any(c in nq for c in _NPC_CUE) or any(c in query for c in ("NPC", "那個人", "那名", "那位"))
    is_object = any(c in query for c in _OBJECT_CUE)
    is_fact = any(c in query for c in _FACT_REF_CUE) or any(n in query for n in _FACT_NOUN)
    noun = _strip_weak(nq)
    if noun in _GENERIC_OBJECT_NOUN:                     # 「東西/玩意」非具體 label → 走 object scope fallback
        noun = ""

    # scope：決定指代的目標 kind（人 > 物 > 事；其餘無 scope）
    scope = "actor" if is_npc else ("object" if is_object else ("fact" if is_fact else None))
    want_kind = _SCOPE_KIND.get(scope)

    def _focus_of_kind(k):
        if current_focus and current_focus.get("kind") == k and current_focus.get("id"):
            return current_focus["id"]
        return None

    def _recent_of_kind(k):
        return _first_of_kind(recent, k)[0]

    def _visible_of_kind(k):
        return _first_of_kind(vis, k)[0]

    # 2. exact / substring label or alias（有具體名詞時）；scope 設定時，命中候選優先過濾該 kind
    if noun:
        cands = list(dict.fromkeys(_match_by_label(world, noun)))
        if want_kind and world is not None:
            kind_cands = [c for c in cands
                          if world.get(c) is not None and world.get(c).kind == want_kind]
            if kind_cands:
                cands = kind_cands
        if len(cands) == 1:
            return res(cands[0], "label_alias")
        if len(cands) > 1:
            return ambiguous(cands)

    # 3. scope-based 解析（無具體名詞匹配時）；**object scope 絕不被 current_focus=NPC 卡住**
    if scope == "fact":
        fid, _ = _first_of_kind(recent, "fact")
        if fid is None and kf:
            fid = kf[0].get("id")
        if fid:
            return res(fid, "known_facts")
        return unresolved                                # 不亂猜（不退回 NPC/物件）
    if scope == "actor":
        if not is_past and _focus_of_kind("actor"):
            return res(_focus_of_kind("actor"), "current_focus")
        if _recent_of_kind("actor"):
            return res(_recent_of_kind("actor"), "recent_entities")
        if _visible_of_kind("actor"):
            return res(_visible_of_kind("actor"), "visible_entities")
        return unresolved
    if scope == "object":
        if not is_past and _focus_of_kind("object"):
            return res(_focus_of_kind("object"), "current_focus")
        if _recent_of_kind("object"):
            return res(_recent_of_kind("object"), "recent_entities")
        if _visible_of_kind("object"):
            return res(_visible_of_kind("object"), "visible_entities")
        if inv:                                          # inventory 全是 object
            return res(inv[0]["id"], "inventory_entities")
        return unresolved

    # 4. 無 scope 的純指代（「剛才那個」「這個」）→ current_focus（無「剛才」時）/ recent / visible / inv
    if not noun:
        if not is_past and current_focus and current_focus.get("id"):
            return res(current_focus["id"], "current_focus")
        if recent:
            return res(recent[0]["id"], "recent_entities")
        if vis:
            return res(vis[0]["id"], "visible_entities")
        if inv:
            return res(inv[0]["id"], "inventory_entities")

    return unresolved
