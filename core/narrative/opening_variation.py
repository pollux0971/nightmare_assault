"""core.narrative.opening_variation — Opening Variation Contract（補丁 v0.8）。

問題：prompt 裡的具體名詞被 LLM 照抄——`紙條` 幾乎每局出現、`林晨` 反覆出現、開場目標幾乎都是
`找人`。根因不是文筆，而是 prompt examples 形成高權重錨點（見 patch docs/01）。

解法：**不**用更多 prompt 提醒，而是把開場核心素材抽象化：

    抽象類別（archetype）+ 變體池（weighted pools）+ cooldown ledger + contract enforcement

本模組是 **schema + 選擇器**（純資料/純函式、零 LLM）：

  * `MotiveArchetype` / `PersonalAnchorType` / `MessageMedium`：開場素材的抽象類別列舉。
  * `OpeningVariationContract`：Setup/OpeningDirector 決定、StoryAgent 只能執行的開場契約。
  * `CooldownLedger`：近期用過的 literal/archetype → forbidden（防止連續重複）。
  * `weighted_choice` / `build_contract`：依變體池 + cooldown 權重抽樣（決定性 seed 可重現）。

**邊界（patch 非目標）**：不改 WorldModel / TruthEvidenceGate / SpatialProjection / PlayerState，
不新增固定故事內容，不做模板引擎，不強制收束劇情——contract 只提供「開場素材欄位」，
LLM 仍負責把素材寫自然。
"""
from __future__ import annotations

from dataclasses import asdict, dataclass, field
from enum import Enum
from random import Random
from typing import Iterable, Mapping


# ─────────────────────────────────────────────────────────────────────────────
# 抽象類別列舉（變體池的鍵；對應 patch docs/02-variation-taxonomy）
# ─────────────────────────────────────────────────────────────────────────────

class MotiveArchetype(str, Enum):
    MISSING_PERSON = "missing_person"
    RETRIEVE_OBJECT = "retrieve_object"
    VERIFY_EVENT = "verify_event"
    REPAIR_SYSTEM = "repair_system"
    ESCAPE_OR_WITHDRAW = "escape_or_withdraw"
    INVESTIGATE_SIGNAL = "investigate_signal"
    PROTECT_SOMEONE = "protect_someone"
    PROVE_INNOCENCE = "prove_innocence"
    RECOVER_MEMORY = "recover_memory"
    DELIVER_OR_HIDE = "deliver_or_hide"
    IDENTIFY_ENTITY = "identify_entity"
    MAP_ROUTE = "map_route"


class PersonalAnchorType(str, Enum):
    MISSING_SIBLING = "missing_sibling"
    FORMER_COLLEAGUE = "former_colleague"
    UNKNOWN_SENDER = "unknown_sender"
    PAST_SELF = "past_self"
    CLIENT = "client"
    SUPERVISOR = "supervisor"
    PATIENT_SUBJECT = "patient_subject"
    RECORDED_VOICE = "recorded_voice"
    ACCUSED_PERSON = "accused_person"
    NO_PERSON_ANCHOR = "no_person_anchor"


class MessageMedium(str, Enum):
    HANDWRITTEN_NOTE = "handwritten_note"
    VOICE_MESSAGE = "voice_message"
    CORRUPTED_LOG = "corrupted_log"
    ACCESS_RECORD = "access_record"
    RADIO_BURST = "radio_burst"
    CCTV_FRAME = "cctv_frame"
    PRINTED_RECEIPT = "printed_receipt"
    DEVICE_STATUS = "device_status"
    BODY_MARK = "body_mark"
    OBJECT_PLACEMENT = "object_placement"
    ENVIRONMENTAL_TRACE = "environmental_trace"
    NPC_CLAIM = "npc_claim"
    INVENTORY_ANOMALY = "inventory_anomaly"
    SCHEDULE_ENTRY = "schedule_entry"
    MAP_ANNOTATION = "map_annotation"
    PHOTO_ARTIFACT = "photo_artifact"
    TERMINAL_PROMPT = "terminal_prompt"
    EMERGENCY_BROADCAST = "emergency_broadcast"


def is_valid_motive(value: str) -> bool:
    return value in {m.value for m in MotiveArchetype}


def is_valid_anchor(value: str) -> bool:
    return value in {a.value for a in PersonalAnchorType}


def is_valid_medium(value: str) -> bool:
    return value in {m.value for m in MessageMedium}


# ─────────────────────────────────────────────────────────────────────────────
# OpeningVariationContract：開場核心素材契約（StoryAgent 只能執行，不可重抽）
# ─────────────────────────────────────────────────────────────────────────────

@dataclass(frozen=True)
class OpeningVariationContract:
    """開場核心素材契約。

    motive_archetype / personal_anchor_type / message_medium / first_interactable_type
    是**抽象類別**（非具體名詞）；StoryAgent 依此寫表層敘事，但不得自行改回紙條/找人/固定姓名。
    forbidden_* 是 cooldown 擋下的素材；recent_* 是近期使用紀錄（debug）。

    刻意**不含任何劇情走向/結局欄位**——契約只決定「開場用什麼素材」，不收束故事。
    """
    motive_archetype: str
    personal_anchor_type: str
    message_medium: str
    initial_goal: str
    first_interactable_type: str
    personal_anchor_label: str | None = None
    opening_objective_sentence: str | None = None
    forbidden_literals: list[str] = field(default_factory=list)
    forbidden_archetypes: list[str] = field(default_factory=list)
    recent_literals: list[str] = field(default_factory=list)
    recent_archetypes: list[str] = field(default_factory=list)
    cooldown_debug: dict = field(default_factory=dict)

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> "OpeningVariationContract":
        d = dict(data or {})
        known = {f for f in cls.__dataclass_fields__}  # type: ignore[attr-defined]
        return cls(**{k: v for k, v in d.items() if k in known})


# ─────────────────────────────────────────────────────────────────────────────
# CooldownLedger：近期用過的 literal/archetype → forbidden（見 patch docs/04）
# ─────────────────────────────────────────────────────────────────────────────

@dataclass
class CooldownLedger:
    """記錄 literal/archetype 最近一次被用在哪個 run；在冷卻窗內者轉為 forbidden。

    `recent_literals` / `recent_archetypes`: {素材 -> last_used_run}。
    冷卻判定：`current_run - last_used_run <= cooldown_runs` → 仍在冷卻 → forbidden。
    """
    recent_literals: dict[str, int] = field(default_factory=dict)
    recent_archetypes: dict[str, int] = field(default_factory=dict)
    literal_cooldown_runs: int = 3
    archetype_cooldown_runs: int = 2

    def forbidden_literals(self, current_run: int) -> list[str]:
        return [
            literal for literal, used_at in self.recent_literals.items()
            if 0 <= current_run - used_at <= self.literal_cooldown_runs
        ]

    def forbidden_archetypes(self, current_run: int) -> list[str]:
        return [
            archetype for archetype, used_at in self.recent_archetypes.items()
            if 0 <= current_run - used_at <= self.archetype_cooldown_runs
        ]

    def record(self, current_run: int, *, literals: Iterable[str] = (),
               archetypes: Iterable[str] = ()) -> None:
        for literal in literals:
            if literal:
                self.recent_literals[literal] = current_run
        for archetype in archetypes:
            if archetype:
                self.recent_archetypes[archetype] = current_run

    def to_dict(self) -> dict:
        return {
            "recent_literals": dict(self.recent_literals),
            "recent_archetypes": dict(self.recent_archetypes),
            "literal_cooldown_runs": self.literal_cooldown_runs,
            "archetype_cooldown_runs": self.archetype_cooldown_runs,
        }

    @classmethod
    def from_dict(cls, data: dict) -> "CooldownLedger":
        d = dict(data or {})
        return cls(
            recent_literals={str(k): int(v) for k, v in (d.get("recent_literals") or {}).items()},
            recent_archetypes={str(k): int(v) for k, v in (d.get("recent_archetypes") or {}).items()},
            literal_cooldown_runs=int(d.get("literal_cooldown_runs", 3)),
            archetype_cooldown_runs=int(d.get("archetype_cooldown_runs", 2)),
        )


# ─────────────────────────────────────────────────────────────────────────────
# 權重抽樣（決定性：同 rng 種子 + 同池 + 同 forbidden → 同結果）
# ─────────────────────────────────────────────────────────────────────────────

def weighted_choice(rng: Random, weighted_items: Mapping[str, float],
                    forbidden: set[str]) -> tuple[str, bool]:
    """從 weighted_items 依權重抽一個非 forbidden 的鍵。

    回傳 (key, exhausted)。若 forbidden 把所有候選擋光 → exhausted=True 並退回完整池
    （cooldown 不應導致「無開場」，見 patch docs/07 風險 4 fallback pool）。
    """
    candidates = [(k, max(float(w), 0.0)) for k, w in weighted_items.items() if k not in forbidden]
    exhausted = False
    if not candidates:
        exhausted = True
        candidates = [(k, max(float(w), 0.0)) for k, w in weighted_items.items()]
    if not candidates:
        raise ValueError("weighted_choice: 變體池為空")
    total = sum(w for _, w in candidates)
    if total <= 0:                                   # 全 0 權重 → 均勻退化為第一個
        return candidates[0][0], exhausted
    r = rng.random() * total
    upto = 0.0
    for key, weight in candidates:
        upto += weight
        if upto >= r:
            return key, exhausted
    return candidates[-1][0], exhausted


def violates_forbidden_literals(text: str, forbidden_literals: Iterable[str]) -> list[str]:
    """回傳 text 中出現的所有 forbidden literal（供斷言/gate 共用）。"""
    return [lit for lit in forbidden_literals if lit and lit in (text or "")]
