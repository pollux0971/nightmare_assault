"""core.agents.dreaming — 輕量 Dreaming Agent（UB4）。

讓**在場（present）的 active NPC** 在自己的演化層長出情緒/意圖/關係，世界因此「活著」。
鐵律：
- C6 權限邊界：**只寫 npc_registry.<name>.evolving**，碰不到 secret_core（由 Blackboard `dreaming` writer policy 強制）。
- C5 self_aware=false 不編謊：誠實型 NPC 不產 emergent_lie。
- C7 沒戲份凍結：只跑在場 NPC；absent/missing/dead 一律跳過（省成本、狀態不動）。
- 非同步只產 patch：透過 `submit_patch` 提交，安全點 `merge_and_bump` 才生效（不污染同步 story 讀取）。

頻率：每 `every`（預設 5）個 beat 一次。skills/dreaming/SKILL.md 是其 prompt。
"""
from __future__ import annotations

from typing import Any

from core.models import DreamingOutput

DREAMING_EVERY = 5          # 每 N beat 跑一次在場 NPC 的演化


def _get(npc: Any, key: str, default=None):
    return npc.get(key, default) if isinstance(npc, dict) else getattr(npc, key, default)


def _present_npcs(npc_registry: list) -> list:
    """C7：只挑在場 NPC（absent/missing/dead 凍結）。"""
    return [n for n in (npc_registry or []) if _get(n, "presence", "present") == "present"]


def _merge_evolving(current: dict, out: DreamingOutput, self_aware: bool) -> dict:
    """把 DreamingOutput 疊到既有 evolving 上，回傳新的 evolving dict（不就地改）。"""
    cur = dict(current or {})
    ev = {
        "emotional_state": {**cur.get("emotional_state", {}), **(out.emotional_update or {})},
        "relationship": {**cur.get("relationship", {}), **(out.relationship_update or {})},
        "intent": out.intent_update or cur.get("intent", "observe"),
        "revealed_layers": list(cur.get("revealed_layers", [])),
        "emergent_lies": list(cur.get("emergent_lies", [])),
        "personal_arc": (out.personal_arc_note or cur.get("personal_arc", "")),
    }
    if out.revealed_layer and out.revealed_layer not in ev["revealed_layers"]:
        ev["revealed_layers"].append(out.revealed_layer)
    # C5：只有 self_aware=true 才可能說謊；誠實型 NPC 不加 emergent_lie
    if self_aware and out.emergent_lie and out.emergent_lie not in ev["emergent_lies"]:
        ev["emergent_lies"].append(out.emergent_lie)
    return ev


def run_dreaming(caller: Any, blackboard: Any, beat_number: int,
                 every: int = DREAMING_EVERY) -> list:
    """每 every beat 跑一次在場 NPC 的演化（非同步只產 patch）。

    回傳 [(npc_name, DreamingOutput), …]（給 log/測試）；無事可做回 []。
    呼叫端在安全點 `merge_and_bump()` 才讓 patch 生效。
    """
    if caller is None:
        return []
    if every and (beat_number <= 0 or beat_number % every != 0):
        return []

    snap = blackboard.snapshot()
    base_version = snap.get("version", 0)
    recent = snap.get("beat_window", [])[-3:]
    ledger = snap.get("ledger", [])

    results: list = []
    for npc in _present_npcs(snap.get("npc_registry", [])):
        name = _get(npc, "name")
        if not name:
            continue
        self_aware = bool(_get(npc, "self_aware", False))
        context = {
            # dreaming 可讀該 NPC 全狀態（含 secret_core）以反思，但**只能寫 evolving**
            "npc": npc,
            "self_aware": self_aware,
            "recent_beats": [b.get("narrative", "") if isinstance(b, dict) else str(b) for b in recent],
            "ledger": ledger,
            "instruction": (
                "更新這個在場 NPC 的演化層（情緒/關係/意圖/個人線）。"
                + ("" if self_aware else "此 NPC self_aware=false：不得編造謊言（emergent_lie 留空），只會真誠地反應、可能真誠地說錯。")
            ),
        }
        try:
            out: DreamingOutput = caller.call("dreaming", context, output_model=DreamingOutput)
        except Exception:
            continue                                  # 單一 NPC 失敗不影響其他（graceful）
        # C5 在 _merge_evolving 內把關（self_aware=false 不收 emergent_lie）；不就地改 out。
        new_ev = _merge_evolving(_get(npc, "evolving", {}) or {}, out, self_aware)
        # C6：只提交 evolving patch（writer=dreaming，碰不到 anchor/secret_core）
        blackboard.submit_patch({
            "base_version": base_version, "writer": "dreaming",
            "target": f"npc_registry.{name}.evolving", "value": new_ev,
        })
        results.append((name, out))
    return results
