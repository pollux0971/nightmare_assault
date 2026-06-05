from pydantic import BaseModel, Field, field_validator
from typing import Literal


class Option(BaseModel):
    text: str
    tone: Literal["cautious", "bold", "evasive", "aggressive"]


class BeatMeta(BaseModel):
    beat_number: int
    revelations_touched: list[str] = Field(default_factory=list)
    npcs_present: list[str] = Field(default_factory=list)
    pacing: Literal["calm", "rising", "peak"] = "calm"
    audio_cue: Literal["normal", "silence", "sting", "swell"] = "normal"


class DecisionPoint(BaseModel):
    situation_recap: str
    decision_type: Literal["action", "dialogue"]
    suggested_options: list[Option] = Field(default_factory=list)
    free_input_hint: str = "或描述你想做的事…"
    beat_meta: BeatMeta
    is_narration_only: bool = False
    # story 可選輸出的結構化世界實體變更（object/actor/fact；由 WorldModel 套用）。
    # **容錯**：非 list 一律歸 []——malformed entity_delta 不得讓整個 DecisionPoint 解析失敗。
    entity_delta: list = Field(default_factory=list)

    @field_validator("entity_delta", mode="before")
    @classmethod
    def _tolerate_bad_entity_delta(cls, v):
        return v if isinstance(v, list) else []


class WorldTruth(BaseModel):
    what_really_happened: str
    the_threat_is: str
    deadly_rule: str


class Revelation(BaseModel):
    id: str
    type: Literal["knowledge", "item", "person"]
    content: str
    reveal_condition: dict = Field(default_factory=dict)


class Interactable(BaseModel):
    id: str
    type: Literal["item", "clue", "corpse", "door", "npc_trace"]
    linked_fragment: str | None = None
    revealed: bool = False


class Location(BaseModel):
    id: str
    name: str
    description: str
    discovered: bool = False
    exits: list[str] = Field(default_factory=list)
    interactables: list[Interactable] = Field(default_factory=list)


class SceneRegistry(BaseModel):
    current_location: str
    known_locations: list[Location] = Field(default_factory=list)


class NPCEvolving(BaseModel):
    emotional_state: dict = Field(default_factory=dict)
    relationship: dict = Field(default_factory=dict)   # 內含 trust/suspicion/affinity，float 0-1
    intent: Literal["observe", "befriend", "betray", "flee", "manipulate"] = "observe"
    revealed_layers: list[str] = Field(default_factory=list)
    emergent_lies: list[str] = Field(default_factory=list)
    personal_arc: str = ""


class NPC(BaseModel):
    name: str
    profession: str
    personality: Literal["leader", "nervous", "analytical", "optimistic", "mysterious"]
    voice_sample: str
    public_face: str
    secret_core: str
    self_aware: bool
    appearance: str = ""
    presence: Literal["present", "absent", "missing", "dead"] = "present"
    alignment: Literal["allied", "neutral", "departed", "hostile", "dead"] = "neutral"
    offstage_intent: str | None = None
    return_condition: str | None = None
    fate_pressure: float = 0.0
    carried_fragment: str | None = None
    evolving: NPCEvolving = Field(default_factory=NPCEvolving)


class NPCBible(BaseModel):   # setup 輸出用的 NPC 骨架（無 evolving/雙軸）
    name: str
    profession: str
    personality: Literal["leader", "nervous", "analytical", "optimistic", "mysterious"]
    voice_sample: str
    public_face: str
    secret_core: str
    self_aware: bool
    appearance: str = ""


class SetupOutput(BaseModel):
    real_bible: dict = Field(default_factory=dict)
    npc_registry: list[NPCBible] = Field(default_factory=list)
    protagonist: dict = Field(default_factory=dict)
    scene_registry: SceneRegistry
    opening_sequence: list[str] = Field(default_factory=list)


class FragmentReveal(BaseModel):
    id: str
    how_to_reveal: str


class OrchestratorOutput(BaseModel):
    fragments_to_reveal: list[FragmentReveal] = Field(default_factory=list)
    reasoning: str = ""


class WardenOutput(BaseModel):
    rule_violation: bool = False
    violated_rule: str | None = None
    ending_triggered: Literal["death_physical", "death_mental", "truth_revealed", "escape", "transformation"] | None = None
    ending_is_soft: bool = False
    skill_claim: str | None = None
    skill_verdict: Literal["allow", "reject"] | None = None
    skill_limitation: str | None = None
    directive_to_story: str


class DreamingOutput(BaseModel):
    emotional_update: dict = Field(default_factory=dict)
    relationship_update: dict = Field(default_factory=dict)
    intent_update: Literal["observe", "befriend", "betray", "flee", "manipulate"]
    revealed_layer: str | None = None
    emergent_lie: str | None = None
    personal_arc_note: str = ""
    reflection_log: str = ""


class OffstageFateOutput(BaseModel):
    fate_type: Literal["opportunity_return", "missing", "corpse", "hostile_return"]
    fate_narrative: str
    fragment_delivery: str
    state_update: dict = Field(default_factory=dict)
    scene_seed: dict | None = None
    reunion_hook: str = ""


class LedgerFact(BaseModel):
    type: str
    content: str


class CompactorOutput(BaseModel):
    compressed_summary: str
    ledger_updates: list[LedgerFact] = Field(default_factory=list)
    archived_beats: list[str] = Field(default_factory=list)
    preserved_foreshadowings: list[str] = Field(default_factory=list)
    final_usage_estimate: float


class LLMResult(BaseModel):
    text: str
    model_used: str
    input_tokens: int
    output_tokens: int
    latency_ms: int
    success: bool
    error: str | None = None
