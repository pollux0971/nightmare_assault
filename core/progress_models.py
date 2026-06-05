"""Narrative Progress Kernel 資料模型（SK01，移植自穩定化補丁）。

刻意小而扁。world-state 由 kernel 決定，story 只 realize。GameState 是 kernel 的最小狀態，
與 Blackboard 經 core/progress_bridge.py 同步。
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Literal

ProgressDelta = Literal[
    "scene_phase_changed",
    "location_changed",
    "event_resolved",
    "obligation_closed",
    "new_clue_added",
    "item_added",
    "item_removed",
    "npc_spawned",
    "npc_trace_added",
    "danger_level_changed",
    "truth_fragment_revealed",
]

EventStatus = Literal["inactive", "active", "resolved", "failed", "blocked"]
ScenePhase = Literal["beginning", "middle", "turn", "ending"]


@dataclass
class PatchOp:
    op: Literal["set", "add", "remove", "inc"]
    path: str
    value: Any


@dataclass
class EventPatch:
    base_version: int
    event_id: str
    ops: list[PatchOp]
    progress_delta: list[ProgressDelta]
    narrative_obligations: list[str] = field(default_factory=list)
    forbidden_repeats: list[str] = field(default_factory=list)
    new_clues: list[str] = field(default_factory=list)
    new_items: list[str] = field(default_factory=list)
    spawned_npcs: list[str] = field(default_factory=list)
    debug_reason: str = ""


@dataclass
class Obligation:
    id: str
    kind: Literal["reveal", "spawn_npc", "grant_clue", "escalate", "resolve", "transition"]
    description: str
    status: Literal["open", "satisfied", "expired"] = "open"
    priority: int = 1
    due_scene: str | None = None


@dataclass
class EventCandidate:
    id: str
    scene_id: str
    intent_tags: list[str]
    preconditions: list[str] = field(default_factory=list)
    effects: list[PatchOp] = field(default_factory=list)
    satisfies: list[str] = field(default_factory=list)
    grants_clues: list[str] = field(default_factory=list)
    grants_items: list[str] = field(default_factory=list)
    spawns_npcs: list[str] = field(default_factory=list)
    forbidden_after: list[str] = field(default_factory=list)
    max_repeat: int = 1
    cooldown_beats: int = 0


@dataclass
class SceneState:
    id: str
    phase: ScenePhase = "beginning"
    beats_in_phase: int = 0
    max_beats_in_phase: int = 2


@dataclass
class LedgerEntry:
    id: str
    title: str
    content: str
    source_event: str
    first_seen_beat: int
    tags: list[str] = field(default_factory=list)
    # NR0：線索自帶的真相綁定（scene graph 從 real_bible 標註；供 RevelationBridge）
    truth_id: str | None = None
    evidence_strength: float = 0.4
    max_level: str = "actionable"


@dataclass
class InventoryItem:
    id: str
    name: str
    description: str
    source_event: str
    usable: bool = True
    tags: list[str] = field(default_factory=list)
    held_by: str | None = None        # None/"protagonist"=玩家持有；可綁 NPC id（命運跟著走）
    is_key_item: bool = False         # 內部標記：是否關鍵道具。**永不對玩家暴露**（get_inventory 不回傳）


@dataclass
class NPCPresence:
    id: str
    name: str
    visible: bool = False
    current_scene: str | None = None
    last_seen_beat: int | None = None
    entry_mode: str | None = None


@dataclass
class GameState:
    version: int
    beat_number: int
    current_scene: str
    scene_phase: ScenePhase
    scenes: dict[str, SceneState] = field(default_factory=dict)
    event_status: dict[str, EventStatus] = field(default_factory=dict)
    event_counts: dict[str, int] = field(default_factory=dict)
    recent_events: list[str] = field(default_factory=list)
    forbidden_repeats: set[str] = field(default_factory=set)
    open_obligations: list[Obligation] = field(default_factory=list)
    clues: dict[str, LedgerEntry] = field(default_factory=dict)
    inventory: dict[str, InventoryItem] = field(default_factory=dict)
    npcs: dict[str, NPCPresence] = field(default_factory=dict)
    danger_level: int = 0


@dataclass
class ProgressResult:
    accepted: bool
    patch: EventPatch
    committed_event: str
    explanation: str = ""
    soft_lookahead: list[str] = field(default_factory=list)   # 每 beat 重算的可能方向（非既定）
