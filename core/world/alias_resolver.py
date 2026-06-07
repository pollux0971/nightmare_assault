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
# 帶歸屬的 fact 指代（「他說的…」整段抓出，含主語）
_FACT_REF_TRIGGERS = ("他說的", "她說的", "他講的", "她講的", "他提到的", "她提到的",
                      "他說過的", "她說過的")
# 觸發「這是個指代」的詞（給 loop 從整句 action 抽出 reference 片語）——長詞優先
_REF_TRIGGERS = sorted(set(_WEAK) | set(_FACT_REF_CUE) | set(_FACT_REF_TRIGGERS)
                       | {"那個npc", "那名npc", "那個 npc", "那名 npc"},
                       key=len, reverse=True)


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


def extract_reference(action: str):
    """從整句玩家輸入抽出「指代片語」（給 loop 用）。找不到 → None。"""
    t = action or ""
    best_i, best = None, None
    for trig in _REF_TRIGGERS:
        i = t.find(trig)
        if i != -1 and (best_i is None or i < best_i):
            best_i, best = i, trig
    if best_i is None:
        return None
    return t[best_i:best_i + 12]                          # 片語 + 後續名詞窗口（保守）


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
