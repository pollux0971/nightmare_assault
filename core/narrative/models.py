"""core.narrative.models — 敘事控制資料類（階段 N，移植自 patch reference_code/models.py）。

純 dataclass、零依賴。對應 dev/CONTRACTS.md §十二 `NarrativeContract` / `OpeningBlueprint`。
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Literal, Optional

# 真相揭露階梯（不可跳級；開場最多 hinted）——**單一事實來源**，其餘模組一律 import 此處，不得重複定義。
RevealLevel = Literal["hidden", "hinted", "observed", "suspected", "confirmed", "actionable"]
REVEAL_ORDER = ["hidden", "hinted", "observed", "suspected", "confirmed", "actionable"]
REVEAL_RANK = {lvl: i for i, lvl in enumerate(REVEAL_ORDER)}


@dataclass
class ProtagonistMotive:
    personal_loss: str          # 私人失去（誰失蹤/牽掛）
    immediate_goal: str         # 此刻要做什麼
    emotional_stake: str        # 情感賭注
    first_proof: str            # 第一個證據/動機物件（開場一定保留）


@dataclass
class MotifPalette:
    primary: list[str] = field(default_factory=list)
    secondary: list[str] = field(default_factory=list)
    forbidden_or_limited: list[str] = field(default_factory=list)


@dataclass
class OpeningBudget:
    max_core_anomalies: int = 1
    max_personal_hooks: int = 1
    max_false_leads: int = 1
    max_named_objects: int = 3      # 開場主要新元素（含具名物件）上限
    max_new_lore_terms: int = 3     # 新 lore 名詞上限
    max_opening_chars: int = 900


@dataclass
class NarrativeContract:
    core_premise: str
    protagonist_motive: ProtagonistMotive
    central_question: str
    motif_palette: MotifPalette = field(default_factory=MotifPalette)
    opening_budget: OpeningBudget = field(default_factory=OpeningBudget)
    opening_reveal_limit: RevealLevel = "hinted"


@dataclass
class TruthSeed:
    truth_id: str
    reveal_level: RevealLevel
    surface_form: str               # 表層提示（不解釋完整真相）


@dataclass
class OpeningBlueprint:
    beat_purpose: str
    motive_evidence: str
    allowed_elements: list[str] = field(default_factory=list)
    blocked_elements: list[str] = field(default_factory=list)
    truth_seeds: list[TruthSeed] = field(default_factory=list)
    max_opening_chars: int = 900


@dataclass
class QualityGateResult:
    passed: bool
    severity: Literal["ok", "warning", "repairable", "hard_fail"]
    violations: list[str] = field(default_factory=list)
    repair_instruction: Optional[str] = None
