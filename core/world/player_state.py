"""core.world.player_state — 玩家狀態投影層（Step 5；observation-only）。

從 WorldModel **投影**玩家面狀態：身上物件 / 已知主張 / 最近互動 / 當前焦點。
**不是新的世界狀態來源**：不改 WorldModel 權限、不推 reveal、不新增 fact、不顯示 hidden truth。

對應 dev/ROADMAP Step 5、docs/16-worldmodel §七。
"""
from __future__ import annotations

from core.world.model import OBJECT, FACT, ACTOR

# changed_entities reason enum（P3）
REASON_REGISTERED = "registered"
REASON_STATE_CHANGED = "state_changed"
REASON_INSPECTED = "inspected"
REASON_TAKEN = "taken"
REASON_USED = "used"
REASON_FACT_ASSERTED = "fact_asserted"
REASON_NPC_INTRODUCED = "npc_introduced"
REASON_TALKED = "talked"
REASON_AREA_CHANGED = "area_changed"
REASON_EXIT_STATE_CHANGED = "exit_state_changed"
REASON_UNKNOWN = "unknown"

SUMMARY_MAX_CHARS = 400
PLAYER_STATE_SUMMARY_SOURCE = "deterministic_projection"


def _carried(e) -> bool:
    return e.state == "taken" or bool(e.props.get("carried"))


def project_inventory(world) -> list:
    """state=taken 或 props.carried 的 object → 隨身物（即使在安全區也顯示）。"""
    out = []
    for e in world.by_kind(OBJECT):
        if _carried(e):
            out.append({"id": e.id, "label": e.label, "kind": e.kind, "state": e.state,
                        "carried": True, "affordances": list(e.affords or [])})
    return out


def project_known_facts(world) -> list:
    """WorldModel kind=fact → 結構化已知主張（保留自然語意 label + source/confidence/tags）。

    **不**收斂成粗 key、**不**顯示 hidden truth raw content（fact entity 本就只有 label/主張，無 truth 內容）。
    """
    out = []
    for e in world.by_kind(FACT):
        p = e.props or {}
        out.append({"id": e.id, "label": e.label, "state": e.state,
                    "source": p.get("source"), "confidence": p.get("confidence"),
                    "tags": list(p.get("tags") or [])})
    return out


def build_player_state(world, *, current_focus=None, recent_entities=None) -> dict:
    """組玩家狀態投影（純投影；無 LLM、不推 reveal）。"""
    if world is None:
        return {"inventory_entities": [], "known_facts": [],
                "current_focus": None, "recent_entities": []}
    return {
        "inventory_entities": project_inventory(world),
        "known_facts": project_known_facts(world),
        "current_focus": current_focus,
        "recent_entities": list(recent_entities or []),
    }


def player_state_summary(ps: dict, *, max_chars: int = SUMMARY_MAX_CHARS) -> tuple:
    """由 player_state **確定性**生成玩家可讀短摘要（不呼叫 LLM、不新增 fact、不取代 narrative）。"""
    lines = []
    inv = [e["label"] for e in (ps.get("inventory_entities") or [])]
    if inv:
        lines.append("你目前攜帶：" + "、".join(inv) + "。")
    facts = [f["label"] for f in (ps.get("known_facts") or [])]
    if facts:
        lines.append("你已知的主張：" + "；".join(facts[:8]) + "。")
    recent = [r["label"] for r in (ps.get("recent_entities") or [])]
    if recent:
        lines.append("最近互動：" + "、".join(recent[:6]) + "。")
    foc = ps.get("current_focus")
    if foc and foc.get("label"):
        lines.append("目前焦點：" + foc["label"] + "。")
    text = "\n".join(lines)
    truncated = False
    if len(text) > max_chars:
        text = text[:max_chars].rstrip() + "…"
        truncated = True
    return text, truncated
