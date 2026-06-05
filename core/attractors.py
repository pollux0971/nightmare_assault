"""Attractor-based ending（SK13）。

結局不是固定終點節點，而是**吸引子**：依累積狀態算各結局的「拉力」，
某吸引子越過門檻就觸發該結局。死亡←危險累積；真相←揭露碎片；逃脫←摸清出路+存活+有核心。
"""
from __future__ import annotations

from core.constants import (
    DANGER_DEATH_THRESHOLD, TRUTH_FRAGMENT_THRESHOLD, ESCAPE_CLUE_THRESHOLD,
)
from core.progress_models import GameState


def _truth_count(state: GameState) -> int:
    return sum(1 for c in state.clues.values() if "truth" in (c.tags or []))


def evaluate(state: GameState) -> dict[str, float]:
    """回傳各結局吸引子的拉力（0..∞，≥1 表示越過門檻）。

    註（#6 review）：此處 death 拉力仍由 danger_level 推導，**但這不是 live bug**——
    敘事控制開啟時，beat loop 在採用 attractor 死亡結局前會先過 `EndingCausalityGate.check_death`
    （HA2，orchestrator_loop 的 attractor 區塊），無明確死因的 danger-only 死亡會被降級為 danger_warning。
    本次刻意**不**把因果檢查下沉到這裡（依使用者指示，只清理 reveal 系統、不動 danger-death 行為）。
    """
    death = state.danger_level / DANGER_DEATH_THRESHOLD
    truth = _truth_count(state) / TRUTH_FRAGMENT_THRESHOLD
    escape = 0.0
    survived = state.danger_level < DANGER_DEATH_THRESHOLD
    if _truth_count(state) >= 1 and survived:
        escape = len(state.clues) / ESCAPE_CLUE_THRESHOLD
    return {"death_physical": death, "truth_revealed": truth, "escape": escape}


def dominant_ending(state: GameState, threshold: float = 1.0) -> str | None:
    """拉力最大且越過門檻的吸引子 → 結局型別；都未越過 → None。"""
    pulls = evaluate(state)
    name, val = max(pulls.items(), key=lambda kv: kv[1])
    return name if val >= threshold else None
