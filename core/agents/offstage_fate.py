"""core.agents.offstage_fate — 離場 NPC 命運（MC4，生成式）。

離場 NPC 不凍結成紙片人。**程式碼加權擲骰**（受 alignment 影響）決定四種命運之一，LLM 只寫血肉。
每種命運攜帶一個未揭露的 `carried_fragment`。**只寫 npc_registry / scene_registry**（權限邊界 C6，
碰不到 secret_core/world_truth）；非同步只產 patch，安全點才 merge。命運對玩家隱藏直到重逢/搜屍（MC5）。
"""
from __future__ import annotations

import random
from typing import Any

from core.models import OffstageFateOutput

FATE_TYPES = ("opportunity_return", "missing", "corpse", "hostile_return")

# 依 alignment 的命運權重（機率可控、可平衡）
_FATE_WEIGHTS: dict[str, dict[str, int]] = {
    "allied":   {"opportunity_return": 5, "missing": 2, "corpse": 1, "hostile_return": 1},
    "neutral":  {"opportunity_return": 3, "missing": 3, "corpse": 2, "hostile_return": 2},
    "departed": {"opportunity_return": 2, "missing": 4, "corpse": 2, "hostile_return": 2},
    "hostile":  {"opportunity_return": 1, "missing": 2, "corpse": 2, "hostile_return": 5},
}

# 命運 → npc 狀態更新（程式碼決定，非 LLM）
_FATE_STATE = {
    "opportunity_return": {"presence": "present", "alignment": "allied", "offstage_intent": "帶著收穫回來"},
    "missing":            {"presence": "missing", "offstage_intent": "下落不明"},
    "corpse":             {"presence": "dead", "alignment": "dead", "offstage_intent": "死亡"},
    "hostile_return":     {"presence": "present", "alignment": "hostile", "offstage_intent": "變質歸來"},
}


def _get(o: Any, k: str, d=None):
    return o.get(k, d) if isinstance(o, dict) else getattr(o, k, d)


def roll_fate(npc: Any, rng: random.Random | None = None) -> str:
    """依 alignment 加權擲骰決定命運型別（程式碼，非 LLM）。"""
    rng = rng or random.Random()
    weights = _FATE_WEIGHTS.get(_get(npc, "alignment", "neutral"), _FATE_WEIGHTS["neutral"])
    items = list(weights.items())
    total = sum(w for _, w in items)
    r = rng.uniform(0, total)
    acc = 0
    for fate, w in items:
        acc += w
        if r <= acc:
            return fate
    return items[-1][0]


def _state_patches(npc_name: str, fate_type: str, fragment: str | None, base_version: int) -> list[dict]:
    """組命運 patch（只寫 npc_registry；writer=offstage_fate）。"""
    patches = []
    for attr, val in _FATE_STATE.get(fate_type, {}).items():
        patches.append({"base_version": base_version, "writer": "offstage_fate",
                        "target": f"npc_registry.{npc_name}.{attr}", "value": val})
    if fragment:
        patches.append({"base_version": base_version, "writer": "offstage_fate",
                        "target": f"npc_registry.{npc_name}.carried_fragment", "value": fragment})
    return patches


def run_offstage_fate(caller: Any, blackboard: Any, npc_name: str,
                      carried_fragment: str | None = None, fate_type: str | None = None,
                      rng: random.Random | None = None) -> OffstageFateOutput | None:
    """跑一個離場 NPC 的命運：程式碼定型別 → LLM 寫血肉 → 提交 patch（只寫 npc/scene）。"""
    snap = blackboard.snapshot()
    npc = None
    for n in snap.get("npc_registry") or []:
        if _get(n, "name") == npc_name:
            npc = n
            break
    if npc is None:
        return None
    fate_type = fate_type or roll_fate(npc, rng)        # 程式碼決定命運（非 LLM）

    ctx = {
        "fate_type": fate_type,                          # 已擲定，LLM 不得更改
        "npc": {"name": npc_name, "profession": _get(npc, "profession"),
                "voice_sample": _get(npc, "voice_sample"), "secret_core": _get(npc, "secret_core")},
        "carried_fragment": carried_fragment,
        "instruction": "命運型別已由程式碼擲定，你只寫血肉並包裝 carried_fragment；不要更改 fate_type。",
    }
    try:
        out: OffstageFateOutput = caller.call("offstage-fate", ctx, output_model=OffstageFateOutput)
    except Exception:
        return None
    out.fate_type = fate_type                            # 強制與程式碼一致

    base_version = snap.get("version", 0)
    for p in _state_patches(npc_name, fate_type, carried_fragment, base_version):
        blackboard.submit_patch(p)
    # 屍體：種 corpse interactable（scene_seed → scene_registry）
    if fate_type == "corpse" and out.scene_seed:
        blackboard.submit_patch({"base_version": base_version, "writer": "offstage_fate",
                                 "target": "scene_registry.pending_corpse", "value": out.scene_seed})
    return out


# ── MC5：命運隱藏 + 重逢揭曉 ────────────────────────────────────────────────
def store_hidden_fate(db: Any, run_id: str, npc_name: str, out: OffstageFateOutput,
                      carried_fragment: str | None) -> None:
    """把命運敘事存進隱藏 table（對玩家不可見，直到重逢/搜屍）。"""
    try:
        db.add_offstage_fate(run_id, npc_name, out.fate_type, out.fate_narrative, carried_fragment)
    except Exception:
        pass


def _fragment_content(blackboard: Any, frag_id: str | None) -> str | None:
    if not frag_id:
        return None
    pool = (blackboard.snapshot().get("real_bible") or {}).get("revelation_pool") or []
    for f in pool:
        if isinstance(f, dict) and f.get("id") == frag_id:
            return f.get("content", frag_id)
    return frag_id


def reveal_carried_fragment(blackboard: Any, db: Any, run_id: str, npc_name: str) -> str | None:
    """重逢/搜屍：把離場 NPC 攜帶的碎片從隱藏搬進 revealed_bible（過揭露），清掉 carried_fragment。"""
    try:
        fate = db.get_offstage_fate(run_id, npc_name)
    except Exception:
        fate = None
    if fate is None or fate.get("revealed"):
        return None
    frag_id = fate.get("carried_fragment")
    content = _fragment_content(blackboard, frag_id)
    if content:
        snap = blackboard.snapshot()
        rb = dict(snap.get("revealed_bible") or {})
        frags = list(rb.get("revealed_fragments") or [])
        if not any(isinstance(f, dict) and f.get("id") == frag_id for f in frags):
            frags.append({"id": frag_id, "content": content,
                          "how_to_reveal": f"由 {npc_name} 的命運揭曉"})
        rb["revealed_fragments"] = frags
        try:
            blackboard.write("orchestrator", "revealed_bible", rb)
        except Exception:
            blackboard.revealed_bible = rb
    try:                                            # 清掉 npc 的 carried_fragment（已交付）
        blackboard.write("offstage_fate", f"npc_registry.{npc_name}.carried_fragment", None)
    except Exception:
        pass
    try:
        db.mark_fate_revealed(run_id, npc_name)
    except Exception:
        pass
    return content


def check_reunions(blackboard: Any, db: Any, run_id: str) -> list:
    """掃在場（present）且有未揭曉隱藏命運的 NPC → 揭曉其碎片。回 [(npc, content)]。"""
    revealed = []
    for n in blackboard.snapshot().get("npc_registry") or []:
        if not isinstance(n, dict) or n.get("presence") != "present":
            continue
        try:
            fate = db.get_offstage_fate(run_id, n.get("name"))
        except Exception:
            fate = None
        if fate and not fate.get("revealed"):
            content = reveal_carried_fragment(blackboard, db, run_id, n.get("name"))
            if content:
                revealed.append((n.get("name"), content))
    return revealed


def offstage_fate_tick(caller: Any, blackboard: Any, beat_number: int, every: int = 8) -> list:
    """低頻命運 tick：對尚未定命運的離場 NPC（absent/missing）跑一次。回 [(npc, fate_type)]。"""
    if caller is None or (every and (beat_number <= 0 or beat_number % every != 0)):
        return []
    snap = blackboard.snapshot()
    pool = (snap.get("real_bible") or {}).get("revelation_pool") or []
    revealed_ids = {f.get("id") for f in (snap.get("revealed_bible") or {}).get("revealed_fragments", [])}
    unrevealed = [f.get("id") for f in pool if isinstance(f, dict) and f.get("id") not in revealed_ids]
    results = []
    for n in snap.get("npc_registry") or []:
        if not isinstance(n, dict):
            continue
        if n.get("presence") in ("absent", "missing") and not n.get("carried_fragment"):
            frag = unrevealed.pop(0) if unrevealed else None
            out = run_offstage_fate(caller, blackboard, n.get("name"), carried_fragment=frag)
            if out is not None:
                results.append((n.get("name"), out.fate_type))
    if results:
        blackboard.merge_and_bump()                      # 安全點 merge
    return results
