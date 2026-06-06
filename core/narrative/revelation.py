"""core.narrative.revelation — RevelationBridge（NR0，敘事控制 v0.2 · 重做版）。

修補核心斷鏈：玩家調查（kernel 線索 / NPC 對話 / 文件）必須轉成**官方真相進度並寫進 revealed_bible**。

設計（對應使用者 6 點要求）：
1. kernel clue/evidence **自帶 truth_id + evidence_strength**（由 scene graph 從 real_bible 標註）。
2. RevealManager 依 truth_id **累積 evidence_strength** 決定 reveal_level（strength 驅動，非只「升一階」）。
3. revealed_bible 記錄每個 truth 的 hidden→hinted→observed→suspected→confirmed→actionable。
4. 結局碎片數含 hinted/observed/suspected（不只 confirmed），且明細用標題顯示，不全是 ？？？。
5. NPC-chat 的有效資訊走**同一條 bridge**（source=npc_chat，強度較低/有上限）。
6. 檢查紙條 / 詢問 NPC / 查文件後，revealed_bible 不再 0/7。

防暴雷不變：bridge 只決定「等級」；content 僅在 confirmed 才全文進 revealed_fragments；story/npc 永不見 real_bible。
對應 dev/CONTRACTS.md §十四（EvidenceEvent / RevelationBridge / RevealLedger / EvidenceMapping）。
"""
from __future__ import annotations

from dataclasses import dataclass, field

# REVEAL_ORDER / REVEAL_RANK 單一來源在 core.narrative.models；此處 re-export 供既有 importer 沿用。
from core.narrative.models import REVEAL_ORDER, REVEAL_RANK

CONFIRMED_LEVEL = "confirmed"

# evidence_strength 累積 → reveal_level 門檻（strength 驅動；要求 #2）
STRENGTH_THRESHOLDS = [
    ("hinted", 0.25), ("observed", 0.60), ("suspected", 1.00),
    ("confirmed", 1.50), ("actionable", 2.20),
]


def strength_to_level(total_strength: float) -> str:
    """累積證據強度 → 揭露等級。"""
    level = "hidden"
    for name, thr in STRENGTH_THRESHOLDS:
        if total_strength >= thr:
            level = name
    return level


def _min_level(a: str, b: str) -> str:
    return a if REVEAL_RANK.get(a, 0) <= REVEAL_RANK.get(b, 0) else b


@dataclass
class EvidenceEvent:
    """玩家可見的線索事件。truth_id 自帶；evidence_strength 驅動升級；max_level 為來源可信度上限。

    source ∈ {kernel, story, npc_chat, document, item, fallback}（§十五擴充）。
    """
    evidence_id: str
    source: str
    truth_id: str | None                     # fallback 來源可為 None（map 不到真相，仍記為 unmapped）
    evidence_strength: float = 0.3
    max_level: str = "actionable"            # 此來源最高能推到的等級（NPC 通常設 hinted/observed）
    surface_text: str = ""
    player_action: str | None = None
    scene_id: str | None = None
    beat_number: int | None = None
    atmosphere_only: bool = False
    player_facing: bool = True               # 是否對玩家呈現（HA3/§十五）
    debug_reason: str = ""                   # 來由（debug 用，不洩 hidden content）


@dataclass
class TruthProgress:
    truth_id: str
    level: str = "hidden"
    strength: float = 0.0                    # 累積證據強度
    evidence_ids: list[str] = field(default_factory=list)
    sources: list[str] = field(default_factory=list)
    title: str = ""
    content: str = ""                        # 全文（僅 confirmed 才對玩家露）


@dataclass
class RevealLedger:
    truths: dict[str, TruthProgress] = field(default_factory=dict)

    def get_or_create(self, truth_id: str, *, title: str = "", content: str = "") -> TruthProgress:
        if truth_id not in self.truths:
            self.truths[truth_id] = TruthProgress(truth_id=truth_id, title=title, content=content)
        else:
            t = self.truths[truth_id]
            if title and not t.title:
                t.title = title
            if content and not t.content:
                t.content = content
        return self.truths[truth_id]

    def level_of(self, truth_id: str) -> str:
        t = self.truths.get(truth_id)
        return t.level if t else "hidden"

    def confirmed_truth_ids(self) -> list[str]:
        return [t.truth_id for t in self.truths.values()
                if REVEAL_RANK[t.level] >= REVEAL_RANK[CONFIRMED_LEVEL]]

    def counts(self) -> dict[str, int]:
        vals = list(self.truths.values())
        return {
            "total": len(vals),
            "hinted_or_better": sum(REVEAL_RANK[t.level] >= 1 for t in vals),
            "observed_or_better": sum(REVEAL_RANK[t.level] >= 2 for t in vals),
            "suspected_or_better": sum(REVEAL_RANK[t.level] >= 3 for t in vals),
            "confirmed_or_better": sum(REVEAL_RANK[t.level] >= 4 for t in vals),
            "actionable": sum(REVEAL_RANK[t.level] >= 5 for t in vals),
        }

    # ── 持久化：直接是 revealed_bible 的子結構（要求 #3）───────────────────────
    def to_progress_dict(self) -> dict:
        return {tid: {"level": t.level, "strength": round(t.strength, 3),
                      "evidence_ids": list(t.evidence_ids), "sources": list(t.sources),
                      "title": t.title} for tid, t in self.truths.items()}

    @classmethod
    def from_progress_dict(cls, data: dict | None, pool: list | None = None) -> "RevealLedger":
        led = cls()
        bytitle = {f.get("id"): f for f in (pool or []) if isinstance(f, dict) and f.get("id")}
        for tid, d in (data or {}).items():
            d = d or {}
            frag = bytitle.get(tid, {})
            led.truths[tid] = TruthProgress(
                truth_id=tid, level=d.get("level", "hidden"), strength=float(d.get("strength", 0.0)),
                evidence_ids=list(d.get("evidence_ids") or []), sources=list(d.get("sources") or []),
                title=d.get("title") or frag.get("title", ""), content=frag.get("content", ""))
        return led


class RevelationBridge:
    """累積 evidence_strength → reveal_level（要求 #2）。"""

    def apply(self, ledger: RevealLedger, events: list[EvidenceEvent]) -> list[dict]:
        updates: list[dict] = []
        for ev in events:
            if ev.atmosphere_only or not ev.truth_id:
                continue
            p = ledger.get_or_create(ev.truth_id)
            old = p.level
            p.strength += max(0.0, float(ev.evidence_strength))
            if ev.evidence_id and ev.evidence_id not in p.evidence_ids:
                p.evidence_ids.append(ev.evidence_id)
            if ev.source and ev.source not in p.sources:
                p.sources.append(ev.source)
            # strength 決定等級，但不超過此來源可信度上限（max_level）
            new_level = _min_level(strength_to_level(p.strength), ev.max_level or "actionable")
            if REVEAL_RANK[new_level] > REVEAL_RANK[p.level]:
                p.level = new_level
            if p.level != old:
                updates.append({"truth_id": ev.truth_id, "from": old, "to": p.level,
                                "strength": round(p.strength, 3), "source": ev.source})
        return updates


# ── Reveal Reward Loop（truth_investigation 的可觀測回報；不碰 TruthEvidenceGate）──────
# 真相型玩家持續調查卻只停 hinted 的問題：gate 放行但沒有 evidence 映射時，reveal ladder 不動。
# 這層在「gate 已放行的 truth_investigation」beat 上，對**已 hinted/observed** 的真相給 reward 強度，
# 讓它 hinted→observed→suspected 爬升——但**單靠 reward 永遠到不了 confirmed**（confirmed 仍須
# kernel clue.core / 結構化 evidence 的強證據）。
REWARD_CEILING_LEVEL = "suspected"       # reward 單獨最高只到 suspected
REWARD_STRENGTH = 0.4                     # 每次 gate-allowed truth_investigation 的 reward 強度
REWARD_STRENGTH_CAP = 1.15               # reward 累積使單一 truth strength 不超過此（留 confirmed 餘裕）


def reward_candidates(ledger: RevealLedger) -> list[TruthProgress]:
    """可被 reward 推進的 in-progress 真相：等級 ∈ [hinted, suspected) 且 strength 未達 reward 天花板。"""
    lo, hi = REVEAL_RANK["hinted"], REVEAL_RANK[REWARD_CEILING_LEVEL]
    return [t for t in ledger.truths.values()
            if lo <= REVEAL_RANK.get(t.level, 0) < hi and t.strength < REWARD_STRENGTH_CAP]


def apply_reveal_reward(ledger: RevealLedger, *, beat: int | None = None,
                        target_id: str | None = None) -> tuple[dict | None, str | None, str | None, str | None]:
    """對一個 in-progress 真相施加一次 reward evidence（capped at suspected + strength cap）。

    回 (update|None, target_id|None, previous_level|None, next_level|None)。
    update 為 None 表示「strength 有增加但未跨階」或「無候選」（由 target_id 是否 None 區分）。
    **單靠 reward 到不了 confirmed**：max_level=suspected 且 strength clamp 到 REWARD_STRENGTH_CAP。
    """
    cands = reward_candidates(ledger)
    if not cands:
        return None, None, None, None
    target = ledger.truths.get(target_id) if target_id else None
    if target is None or target not in cands:
        # 推進最接近下一階者（strength 最高、其次 rank 最高）→ 給玩家可見的穩定回報
        target = max(cands, key=lambda t: (t.strength, REVEAL_RANK.get(t.level, 0)))
    add = min(REWARD_STRENGTH, max(0.0, REWARD_STRENGTH_CAP - target.strength))
    ev = EvidenceEvent(
        evidence_id=f"ev.reward.{target.truth_id}.b{beat}",
        source="investigation_reward", truth_id=target.truth_id,
        evidence_strength=add, max_level=REWARD_CEILING_LEVEL,
        surface_text="（持續調查累積）", beat_number=beat,
        debug_reason="investigation_reward")
    prev = target.level
    updates = RevelationBridge().apply(ledger, [ev])
    return (updates[0] if updates else None), target.truth_id, prev, target.level


# ── Reveal → public-safe known_fact 投影（PlayerState Payoff）──────────────────
# 揭露進度（observed+）→ 可檢視的 public known_fact：**只用 title 標題，永不放 hidden content**。
# confirmed_public 只在 ledger 真的到 confirmed+（強證據/結構化 evidence）才出現——reward 上限 suspected。
_PUBLIC_LEVEL = {"observed": "observed", "suspected": "suspected",
                 "confirmed": "confirmed_public", "actionable": "confirmed_public"}


def reveal_public_facts(ledger: RevealLedger) -> list[dict]:
    """把 observed+ 的真相投影成 public-safe fact 描述（title + 公開等級，**無 content**）。

    回 [{truth_id, title, level}]；hidden/hinted → 不投影（太弱，只是暗示）。
    """
    out: list[dict] = []
    for t in ledger.truths.values():
        pub = _PUBLIC_LEVEL.get(t.level)
        if pub:
            out.append({"truth_id": t.truth_id, "title": t.title or "未命名的真相", "level": pub})
    return out


# ── 建帳本 / 寫進 revealed_bible ─────────────────────────────────────────────
def build_ledger_from_bible(real_bible: dict | None) -> RevealLedger:
    """從 real_bible.revelation_pool 種出各碎片（全 hidden，帶 title/content）。"""
    led = RevealLedger()
    for f in (real_bible or {}).get("revelation_pool") or []:
        if isinstance(f, dict) and f.get("id"):
            led.get_or_create(f["id"], title=f.get("title", ""), content=f.get("content", ""))
    return led


def write_ledger_to_revealed_bible(blackboard, ledger: RevealLedger, *, writer: str = "orchestrator"):
    """把帳本寫進 revealed_bible：truth_progress（全等級）+ revealed_fragments（confirmed 全文）。

    要求 #3：revealed_bible 記錄 hinted/observed/suspected/confirmed/actionable。
    要求 #6：confirmed 的碎片進 revealed_fragments（既有結局/逃脫門檻沿用）。
    """
    snap = blackboard.snapshot()
    rb = dict(snap.get("revealed_bible") or {})
    rb["truth_progress"] = ledger.to_progress_dict()
    # confirmed+ → revealed_fragments（去重；帶 id/title/content）
    have = {f.get("id") for f in (rb.get("revealed_fragments") or []) if isinstance(f, dict)}
    frags = list(rb.get("revealed_fragments") or [])
    for tid in ledger.confirmed_truth_ids():
        if tid not in have:
            t = ledger.truths[tid]
            frags.append({"id": tid, "title": t.title, "content": t.content})
            have.add(tid)
    rb["revealed_fragments"] = frags
    blackboard.write(writer, "revealed_bible", rb)


def evidence_from_clue_values(clue_items: list, *, source: str = "kernel",
                              beat: int | None = None, scene_id: str | None = None
                              ) -> tuple[list[EvidenceEvent], list[str]]:
    """從帶 truth_id 的 kernel clue 值建 EvidenceEvent（要求 #1：clue 自帶 truth_id）。

    clue_items：[(clue_id, clue_value_dict), ...]。回傳 (events, unmapped_clue_ids)。
    clue_value 需含 truth_id；缺 truth_id 的線索 → unmapped（debug 告警，不升級）。
    """
    def _g(o, k, d=None):
        return o.get(k, d) if isinstance(o, dict) else getattr(o, k, d)

    events: list[EvidenceEvent] = []
    unmapped: list[str] = []
    for cid, val in clue_items or []:
        tid = _g(val, "truth_id")
        if not tid:
            unmapped.append(cid)
            continue
        _str = _g(val, "evidence_strength")          # 別用 `or`：合法的 0.0 是 falsy，會被覆蓋
        _ml = _g(val, "max_level")
        events.append(EvidenceEvent(
            evidence_id=f"ev.{cid}", source=source, truth_id=str(tid),
            evidence_strength=float(_str if _str is not None else 0.4),
            max_level=_ml if _ml else "actionable",
            surface_text=_g(val, "content") or _g(val, "title") or "",
            scene_id=scene_id, beat_number=beat,
            atmosphere_only=bool(_g(val, "atmosphere_only"))))
    return events, unmapped


# ── NPC-chat 有效資訊 → evidence（要求 #5）────────────────────────────────────
def build_truth_keyword_index(real_bible: dict | None) -> dict[str, list[str]]:
    """從 revelation_pool 抽每個真相的關鍵詞（title + content 的可辨識片段），供掃 NPC 回覆。"""
    import re
    # 太常見、辨識度低的詞塊不當關鍵詞（避免 substring 誤判）
    _STOP = {"這裡", "什麼", "他們", "那是", "一個", "已經", "可能", "自己", "我們",
             "你們", "沒有", "知道", "因為", "所以", "但是", "如果", "實驗", "研究",
             "記憶", "聲音", "東西", "地方", "時候", "發生", "出現", "看到", "感覺"}

    def _keywords(fieldval: str) -> list[str]:
        kws: list[str] = []
        # 數字代號：只收**夠具體**者（含小數、單位 Hz、或 ≥3 位數），排除「3」「17」這種裸短數字
        for n in re.findall(r"[0-9]+(?:\.[0-9]+)?(?:Hz|hz)?", fieldval or ""):
            digits = re.sub(r"\D", "", n)
            if "." in n or n.lower().endswith("hz") or len(digits) >= 3:
                kws.append(n)
        # 中文詞塊：≥3 字（2 字詞太易誤撞），且不在停用詞表
        kws += [w for w in re.findall(r"[一-鿿]{3,6}", fieldval or "") if w not in _STOP]
        return kws

    idx: dict[str, list[str]] = {}
    for f in (real_bible or {}).get("revelation_pool") or []:
        if not (isinstance(f, dict) and f.get("id")):
            continue
        seen, uniq = set(), []
        for k in _keywords(f.get("title", "")) + _keywords(f.get("content", "")):
            if k and k not in seen:
                seen.add(k); uniq.append(k)
        if uniq:
            idx[f["id"]] = uniq[:6]
    return idx


def evidence_from_npc_reply(reply: str, truth_index: dict[str, list[str]], *,
                            beat: int | None = None, answer_status: str | None = None
                            ) -> list[EvidenceEvent]:
    """掃 NPC 回覆中的真相關鍵詞 → EvidenceEvent（source=npc_chat，強度低、上限封頂）。

    NPC 不可信：max_level 封到 observed；迴避/拒答（evaded）不給強度。
    """
    reply = reply or ""
    if answer_status in ("evaded", "evasion", "none"):   # 迴避/未答 → 不給證據
        return []
    out: list[EvidenceEvent] = []
    for tid, kws in (truth_index or {}).items():
        hits = sum(1 for k in kws if k and k in reply)
        if hits <= 0:
            continue
        # 命中越多關鍵詞、答得越實 → 強度略高（仍遠低於親見證據）
        strength = 0.2 + 0.1 * min(hits, 3)
        if answer_status == "partial":
            strength += 0.05
        out.append(EvidenceEvent(
            evidence_id=f"ev.npc.{tid}.b{beat}", source="npc_chat", truth_id=tid,
            evidence_strength=round(strength, 3), max_level="observed",
            surface_text="（NPC 透露）", beat_number=beat))
    return out


def public_recap(ledger: RevealLedger) -> dict:
    """HA3：玩家可見的**遮罩** recap——不含任何 hidden truth content。

    found：已達 hinted+ 者（confirmed 才附 content，較低只給 title+level）；
    hidden：只給 count + 遮罩標題（『未解的線索 #N』，連真實標題都不露）。
    """
    found, hidden_n = [], 0
    for t in ledger.truths.values():
        r = REVEAL_RANK[t.level]
        if r >= REVEAL_RANK["confirmed"]:
            found.append({"id": t.truth_id, "title": t.title or "未命名的真相",
                          "level": t.level, "content": t.content})    # 已確認 = 玩家已得
        elif r >= REVEAL_RANK["hinted"]:
            found.append({"id": t.truth_id, "title": t.title or "未命名的真相",
                          "level": t.level})                          # 接觸過/掌握：不附 content
        else:
            hidden_n += 1
    c = ledger.counts()
    return {
        "found": found, "hidden_count": hidden_n,
        "hidden_titles": [f"未解的線索 #{i + 1}" for i in range(hidden_n)],   # 遮罩，無真實標題/內容
        "total": c["total"], "confirmed": c["confirmed_or_better"],
        "line": (f"你發現了 {c['hinted_or_better']}/{c['total']} 條真相線索，"
                 f"其中 {c['confirmed_or_better']} 條已確認。") if c["total"] else "",
    }


def recap_from_ledger(ledger: RevealLedger) -> dict:
    """結局復盤：部分進度計數 + 分層明細（供 ending 顯示標題，不全 ？？？）。要求 #4。"""
    c = ledger.counts()
    discovered, suspected_list, hinted_list, hidden_list = [], [], [], []
    for t in ledger.truths.values():
        entry = {"id": t.truth_id, "title": t.title or "未命名的真相", "level": t.level,
                 "content": t.content}
        r = REVEAL_RANK[t.level]
        if r >= REVEAL_RANK["confirmed"]:
            discovered.append(entry)
        elif r >= REVEAL_RANK["observed"]:
            suspected_list.append(entry)
        elif r >= REVEAL_RANK["hinted"]:
            hinted_list.append(entry)
        else:
            hidden_list.append(entry)
    return {
        "total": c["total"], "found": c["hinted_or_better"], "confirmed": c["confirmed_or_better"],
        "suspected": c["suspected_or_better"], "observed": c["observed_or_better"],
        "confirmed_list": discovered, "suspected_list": suspected_list,
        "hinted_list": hinted_list, "hidden_list": hidden_list,
        "line": (f"你發現了 {c['hinted_or_better']}/{c['total']} 條真相線索，"
                 f"其中 {c['confirmed_or_better']} 條已確認、{c['observed_or_better']} 條已掌握。")
        if c["total"] else "",
    }
