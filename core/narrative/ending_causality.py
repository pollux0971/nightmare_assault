"""core.narrative.ending_causality — 結局因果硬閘（HA2，Runtime Hard-Gate v0.3.1）。

所有結局先成 `EndingCandidate`，再過閘。核心硬規則：
- **`death_physical` 必須有明確死因**：warden hard_trigger==death / progress.death_cause_event /
  玩家明確致命行為且場景規則確認致命。**danger_level 達標不得直接致死**——只能降級
  danger_warning / injury / failed_escape（attractors.dominant_ending 的 death←danger 在此被攔）。
- escape 須 commit step；0 confirmed → ambiguous_escape（與 NR3/NR4 一致）。

零 LLM、規則版。對應 dev/CONTRACTS.md §十五（EndingCausalityGate）。
"""
from __future__ import annotations

from dataclasses import dataclass

DEATH_TYPES = {"death_physical", "death_mental", "death"}


@dataclass
class EndingCandidate:
    type: str
    source: str                      # warden|attractor|...
    confidence: float = 1.0
    cause_event_id: str | None = None
    requires_commit: bool = False
    debug_reason: str = ""


@dataclass
class GateResult:
    allowed: bool
    reason: str
    downgrade_to: str | None = None


def _get(o, k, d=None):
    return o.get(k, d) if isinstance(o, dict) else getattr(o, k, d)


class EndingCausalityGate:
    def check_death(self, candidate: EndingCandidate, *, warden_result=None,
                    progress_result=None, state=None) -> GateResult:
        if candidate.type not in DEATH_TYPES:
            return GateResult(True, "not a death ending")

        # 明確死因：warden 硬觸發死亡 / kernel 死因事件 / 明確致命行為
        warden_hard_death = (_get(warden_result, "ending_triggered") in DEATH_TYPES
                             or _get(warden_result, "hard_trigger") == "death")
        death_cause = _get(progress_result, "death_cause_event") is not None
        lethal_action = bool(_get(progress_result, "explicit_lethal_action", False))
        if warden_hard_death or death_cause or lethal_action or candidate.cause_event_id:
            return GateResult(True, "death has explicit cause")

        # 只有 danger 達標 → 不准直接死，降級為警告
        return GateResult(False, "danger threshold alone cannot trigger death_physical",
                          downgrade_to="danger_warning")

    def check_escape(self, candidate: EndingCandidate, *, escape_step: str,
                     reveal_progress: dict | None = None) -> GateResult:
        if candidate.type != "escape":
            return GateResult(True, "not escape")
        if escape_step != "commit":
            return GateResult(False, "escape requires commit step", downgrade_to="exit_candidate")
        rp = reveal_progress or {}
        # 相容兩種計數鍵：public_recap 用 'confirmed'，RevealLedger.counts() 用 'confirmed_or_better'
        confirmed = int(rp.get("confirmed", rp.get("confirmed_or_better", 0)))
        if confirmed <= 0:
            return GateResult(True, "low-truth escape becomes ambiguous",
                              downgrade_to="ambiguous_escape")
        return GateResult(True, "escape accepted")
