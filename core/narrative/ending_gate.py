"""core.narrative.ending_gate — 結局因果門檻（NC6，移植自 patch reference_code）。

attractor / warden 提議結局；EndingGate 判斷是否允許：
- clean_escape 需 explicit_escape_action + exit_location_reached + threat_resolved。
- **0/8 真相碎片不可 clean_escape**（只可 ambiguous_escape）。
- 不滿足因果 → 降級 ambiguous_escape / fail-forward。refine UB7（masked render 不變）。
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import Any

# 玩家「明確嘗試逃離」的關鍵詞（explicit_escape_action 偵測）。
_ESCAPE_WORDS = ["逃", "離開", "出口", "出去", "逃出", "逃離", "escape", "離開這裡", "衝出", "跑出"]
# 場景看起來是出口/外部。
_EXIT_MARKERS = ["exit", "outside", "出口", "門外", "外面", "大門", "停車場", "碼頭"]


@dataclass
class EndingState:
    explicit_escape_action: bool = False
    exit_location_reached: bool = False
    threat_avoided_or_resolved: bool = False
    truth_fragments_confirmed: int = 0
    total_beats: int = 0


@dataclass
class EndingDecision:
    allowed: bool
    ending_type: str
    reason: str


class EndingGate:
    def check(self, proposed: str, state: EndingState) -> EndingDecision:
        if proposed == "clean_escape":
            if not state.explicit_escape_action:
                return EndingDecision(False, "ambiguous_escape", "missing_explicit_escape_action")
            if not state.exit_location_reached:
                return EndingDecision(False, "ambiguous_escape", "exit_location_not_reached")
            if not state.threat_avoided_or_resolved:
                return EndingDecision(False, "ambiguous_escape", "threat_not_resolved")
            if state.truth_fragments_confirmed == 0:
                return EndingDecision(False, "ambiguous_escape",
                                      "zero_truth_fragments_only_allows_ambiguous_escape")
            return EndingDecision(True, "clean_escape", "clean_escape_allowed")

        if proposed == "ambiguous_escape":
            if state.explicit_escape_action and state.exit_location_reached:
                return EndingDecision(True, "ambiguous_escape", "ambiguous_escape_allowed")
            return EndingDecision(False, "ambiguous_escape", "missing_escape_causality")

        return EndingDecision(True, proposed, "non_escape_ending_uses_other_rules")


def build_ending_state(blackboard: Any, game_state: Any, last_player_decision: str,
                       beat_number: int) -> EndingState:
    """從遊戲狀態 + 最後玩家行動推導 EndingState（啟發式）。"""
    snap = blackboard.snapshot() if hasattr(blackboard, "snapshot") else {}
    decision = (last_player_decision or "").lower()
    scene = str(getattr(game_state, "current_scene", "") or "").lower()
    danger = int(getattr(game_state, "danger_level", 0) or 0)
    revealed = (snap.get("revealed_bible") or {}).get("revealed_fragments") or []

    return EndingState(
        explicit_escape_action=any(w in decision for w in _ESCAPE_WORDS),
        exit_location_reached=any(m in scene for m in _EXIT_MARKERS),
        threat_avoided_or_resolved=danger <= 1,
        truth_fragments_confirmed=len(revealed),
        total_beats=beat_number,
    )


def gate_escape_quality(blackboard: Any, game_state: Any, last_player_decision: str,
                        beat_number: int) -> tuple[str, str]:
    """回傳 (escape_quality, reason)：'clean' 或 'ambiguous'。供 loop 標在 ending 上。"""
    state = build_ending_state(blackboard, game_state, last_player_decision, beat_number)
    decision = EndingGate().check("clean_escape", state)
    return ("clean" if decision.ending_type == "clean_escape" and decision.allowed else "ambiguous",
            decision.reason)


# ── NR3：結局表層變體（玩家看得出 clean vs ambiguous）─────────────────────────
ENDING_SURFACES = ["clean_escape", "ambiguous_escape", "failed_escape", "death", "truth_locked"]


def classify_ending_surface(ending: dict, confirmed_count: int) -> str:
    """由 ending_type + escape_quality（gate 給）+ 已確認真相數，決定**表層變體**。

    規則：死亡→death；逃脫且 clean 且 ≥1 confirmed→clean_escape，否則 ambiguous_escape；
    truth_revealed 但 0 confirmed→truth_locked。**0 confirmed 永不 clean_escape**（refine UB7/NC6）。
    """
    etype = (ending or {}).get("type") or "death"
    if etype in ("death_physical", "death_mental", "death"):
        return "death"
    if etype == "transformation":
        return "death"                       # 轉化＝認知意義上的失敗收束
    if etype == "escape":
        quality = ending.get("escape_quality")
        if quality == "clean" and confirmed_count >= 1:
            return "clean_escape"
        return "ambiguous_escape"            # 含 0 confirmed → 一律 ambiguous
    if etype == "truth_revealed":
        return "clean_escape" if confirmed_count >= 1 else "truth_locked"
    return etype if etype in ENDING_SURFACES else "ambiguous_escape"
