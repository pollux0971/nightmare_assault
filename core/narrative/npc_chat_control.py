"""core.narrative.npc_chat_control — NPC-Chat 敘事控制（NR1，敘事控制 v0.2）。

v0.1 只約束主 Story Agent；npc-chat 仍能無控擴張世界觀（新組織/協定/機制/怪物、真相跳級）。
本模組讓 npc-chat 收**同一份敘事契約**：限定母題、揭露上限、答債，並把有用提示轉成受控 EvidenceEvent。

防暴雷不變：控制契約只給「可用母題/上限/動機」，**不含 real_bible/secret_core**；違規新 lore → repair 重寫。
對應 dev/CONTRACTS.md §十四（NPCChatControl）。
"""
from __future__ import annotations

from dataclasses import dataclass, field

from core.narrative.revelation import EvidenceEvent


@dataclass
class NPCChatControlContext:
    """餵給 npc-chat 的敘事約束（無秘密內容）。"""
    allowed_terms: set[str] = field(default_factory=set)       # motif_palette + allowed_new_terms
    forbidden_terms: set[str] = field(default_factory=set)     # forbidden_or_limited motifs
    reveal_ceiling: str = "hinted"                             # 此 NPC 此刻可揭露的最高階
    answer_debt_level: int = 0                                 # 該話題目前答債（NR2 提供）
    active_motive: str = ""                                    # 主角動機（保持在場）


@dataclass
class NPCChatResponse:
    """npc-chat 的結構化輸出（visible_reply 給玩家；其餘供控制/橋接）。

    answer_status ∈ none|evasion|partial|actionable|confirmed（HC1/§十五）；
    舊值 answered/evaded/refused 仍相容（gate 只看是否「有償還」）。
    """
    visible_reply: str
    answer_status: str = "none"
    npc_emotion_delta: dict = field(default_factory=dict)
    evidence_events: list[dict] = field(default_factory=list)
    new_lore_terms: list[str] = field(default_factory=list)
    used_motifs: list[str] = field(default_factory=list)
    used_truth_ids: list[str] = field(default_factory=list)
    blocked_or_uncertain_claims: list[str] = field(default_factory=list)
    quality_flags: list[str] = field(default_factory=list)


# 償還答債視為「有回應」的 answer_status（含新舊命名）
_PAID_STATUS = ("answered", "partial", "actionable", "confirmed", "refused")


def safe_fallback_reply(question: str = "") -> NPCChatResponse:
    """gate 連 repair 後仍違規 → 決定性安全回覆：不新增 lore，但給 actionable 方向。"""
    return NPCChatResponse(
        visible_reply="我不能確定答案，但你應該先找能驗證的東西：值班紀錄、控制台紀錄，或剛才那張紙條。",
        answer_status="actionable", evidence_events=[], new_lore_terms=[])


class NPCChatControlGate:
    """規則版閘門：偵測未授權新 lore / forbidden 母題 / 未付答債。"""

    def validate(self, ctx: NPCChatControlContext, resp: NPCChatResponse) -> list[str]:
        flags: list[str] = []
        illegal = [t for t in resp.new_lore_terms if t not in ctx.allowed_terms]
        forbidden = [t for t in (resp.new_lore_terms + resp.used_motifs)
                     if t in ctx.forbidden_terms]
        if illegal:
            flags.append("illegal_new_lore_terms:" + ",".join(illegal))
        if forbidden:
            flags.append("forbidden_terms:" + ",".join(forbidden))
        # 答債：debt≥2 必須付（不可純迴避 none/evasion）
        if ctx.answer_debt_level >= 2 and resp.answer_status not in _PAID_STATUS:
            flags.append("answer_debt_not_paid")
        return flags

    def needs_repair(self, flags: list[str]) -> bool:
        return bool(flags)


REPAIR_INSTRUCTION = (
    "重寫這段 NPC 回答：移除所有未授權的新世界觀名詞（新組織/協定/機制/怪物），"
    "保留 NPC 的情緒與任何部分答案；把有用的提示改寫成既有母題能承載的線索。"
    "不得把暗示講成已確認真相。")


def build_control_context(narrative_contract, reveal_ceiling: str = "hinted",
                          answer_debt_level: int = 0,
                          allowed_new_terms: set[str] | None = None) -> NPCChatControlContext:
    """從 NarrativeContract 組 npc-chat 控制 context（無 real_bible）。"""
    nc = narrative_contract
    palette = getattr(nc, "motif_palette", None)
    allowed = set(allowed_new_terms or set())
    if palette is not None:
        allowed |= set(getattr(palette, "primary", []) or [])
        allowed |= set(getattr(palette, "secondary", []) or [])
    forbidden = set(getattr(palette, "forbidden_or_limited", []) or []) if palette else set()
    motive = ""
    pm = getattr(nc, "protagonist_motive", None)
    if pm is not None:
        motive = getattr(pm, "immediate_goal", "") or ""
    return NPCChatControlContext(
        allowed_terms=allowed, forbidden_terms=forbidden,
        reveal_ceiling=reveal_ceiling, answer_debt_level=int(answer_debt_level),
        active_motive=motive)


# NPC 永遠不能把真相推到 confirmed（決定性證明須靠玩家親自發現）
NPC_HARD_CAP = "observed"


def _safe_hint(title: str, max_level: str) -> str:
    t = title or "某條線索"
    return (f"你可以暗示與「{t}」有關的事（最多到 {max_level} 程度），"
            "但不可給出完整真相、不可發明新名詞、不可說穿決定性結論。")


def build_allowed_truth_refs(reveal_ledger, reveal_ceiling: str = "hinted", *,
                             core_truth_ids=frozenset(), npc_cap: str = NPC_HARD_CAP) -> dict:
    """產出 NPC 此刻**可安全引用**的 truth 清單（白名單）+ forbidden 清單。

    規則：NPC 只能引用「未達上限、非 core」的真相，且最多比現況高一階、不超過 min(reveal_ceiling, NPC 硬上限)。
    這讓 NPC 能安全推進 hinted/observed，但碰不到 confirmed 與 core 真相（不回到 keyword 亂猜）。
    """
    from core.narrative.revelation import REVEAL_RANK, REVEAL_ORDER
    # NPC 至少能「暗示」（hinted），最多到 min(全域 ceiling, NPC 硬上限 observed)
    cap_rank = min(max(REVEAL_RANK.get(reveal_ceiling, 1), REVEAL_RANK["hinted"]),
                   REVEAL_RANK.get(npc_cap, 2))
    allowed, forbidden = [], []
    for t in (getattr(reveal_ledger, "truths", {}) or {}).values():
        cur = REVEAL_RANK.get(t.level, 0)
        if t.truth_id in core_truth_ids or cur >= cap_rank:
            forbidden.append(t.truth_id)
            continue
        max_level = REVEAL_ORDER[min(cur + 1, cap_rank)]      # 最多高一階、且不超 cap
        allowed.append({"truth_id": t.truth_id, "max_level": max_level,
                        "safe_hint": _safe_hint(getattr(t, "title", ""), max_level)})
    return {"allowed_truth_refs": allowed, "forbidden_truth_refs": forbidden}


def validate_npc_evidence(events, allowed_refs: list) -> tuple[list, list]:
    """gate：只接受「truth_id ∈ 白名單」的 evidence，並把 level 夾到該 ref 的 max_level。

    回傳 (accepted_events, rejected)；rejected 為 [(event, reason)]，reason ∈
    no_truth_id / truth_id_not_allowed。level 超標 → 夾下來（不直接丟，仍尊重 ceiling）。
    """
    from core.narrative.revelation import REVEAL_RANK
    by_id = {r.get("truth_id"): r for r in (allowed_refs or [])}
    accepted, rejected = [], []
    for e in events or []:
        if not getattr(e, "truth_id", None):
            rejected.append((e, "no_truth_id"))               # 無 truth_id → conversation note，不 grant
            continue
        ref = by_id.get(e.truth_id)
        if ref is None:
            rejected.append((e, "truth_id_not_allowed"))      # 不在白名單（含 core/超 ceiling）→ reject
            continue
        if REVEAL_RANK.get(e.max_level, 5) > REVEAL_RANK.get(ref.get("max_level", "hinted"), 1):
            e.max_level = ref["max_level"]                    # 夾到允許上限
        accepted.append(e)
    return accepted, rejected


def control_context_from_meta(game_meta: dict | None, reveal_ceiling: str = "hinted",
                              answer_debt_level: int = 0) -> NPCChatControlContext | None:
    """從 blackboard.game_meta 存的 narrative_contract dict 組控制 context（run_npc_chat 用）。"""
    nc = (game_meta or {}).get("narrative_contract")
    if not isinstance(nc, dict):
        return None
    palette = nc.get("motif_palette") or {}
    allowed = set(palette.get("primary") or []) | set(palette.get("secondary") or [])
    forbidden = set(palette.get("forbidden_or_limited") or [])
    motive = ((nc.get("protagonist_motive") or {}).get("immediate_goal")) or ""
    return NPCChatControlContext(
        allowed_terms=allowed, forbidden_terms=forbidden,
        reveal_ceiling=reveal_ceiling, answer_debt_level=int(answer_debt_level),
        active_motive=motive)


def response_to_evidence(resp: NPCChatResponse, *, beat: int | None = None) -> list[EvidenceEvent]:
    """把結構化 npc-chat 回報的 evidence_events 轉成 EvidenceEvent（source=npc_chat），供 NR0 橋接。

    安全：每個 evidence 的揭露上限封到 reveal_ceiling（max_level）；NPC 不可信、強度偏低。
    """
    out: list[EvidenceEvent] = []
    for e in resp.evidence_events or []:
        if not isinstance(e, dict) or not e.get("truth_id"):
            continue
        out.append(EvidenceEvent(
            evidence_id=e.get("evidence_id") or f"ev.npc.{e['truth_id']}",
            source="npc_chat", truth_id=str(e["truth_id"]),
            evidence_strength=float(e.get("evidence_strength", 0.25)),
            max_level=e.get("max_level") or e.get("suggested_reveal_level", "observed"),
            surface_text=e.get("surface_text", ""), beat_number=beat))
    return out


def cap_evidence_to_ceiling(resp: NPCChatResponse, ceiling: str) -> NPCChatResponse:
    """把 evidence_events 的揭露上限（max_level）夾到 reveal_ceiling 以下（防 NPC 跳級暴雷）。"""
    from core.narrative.revelation import REVEAL_RANK, REVEAL_ORDER
    cap = REVEAL_RANK.get(ceiling, 1)
    for e in resp.evidence_events or []:
        if isinstance(e, dict):
            cur = e.get("max_level") or e.get("suggested_reveal_level", "observed")
            if REVEAL_RANK.get(cur, 2) > cap:
                e["max_level"] = REVEAL_ORDER[cap]
    return resp
