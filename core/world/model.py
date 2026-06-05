"""core.world.model — 抽象世界模型（WorldModel 骨架）。

主題無關的實體模型:世界裡所有「可被指涉的東西」都是 Entity 的實例,玩家行動 = 對某個
Entity 套用一個 affordance。這一層先當**平行的記憶 + 觀測層**(不取代 kernel),把
移動 / 物件 / 出口 / NPC / 可檢查事實收成同一機制。

四個核心型別:Entity / Affordance / WorldDelta / WorldModel(store)。
Entity kind 先只支援:area / exit / object / actor / fact。

對應 docs/player-sovereignty-principles.md §二、15-player-sovereignty.md。
"""
from __future__ import annotations

import re
from dataclasses import dataclass, field, asdict

# ── kinds ────────────────────────────────────────────────────────────────────
AREA, EXIT, OBJECT, ACTOR, FACT = "area", "exit", "object", "actor", "fact"
KINDS = {AREA, EXIT, OBJECT, ACTOR, FACT}

# ── affordances（玩家對某 Entity 能做的事）────────────────────────────────────
INSPECT, TAKE, USE, MOVE_TO, WITHDRAW_TO, TALK, MANIPULATE = (
    "inspect", "take", "use", "move_to", "withdraw_to", "talk", "manipulate")

# 每種 kind 的預設初始狀態 + 預設 affordances（抽象狀態機）
_DEFAULT_STATE = {OBJECT: "noticed", EXIT: "known", AREA: "known",
                  ACTOR: "present", FACT: "asserted"}
_DEFAULT_AFFORDS = {OBJECT: [INSPECT], EXIT: [MOVE_TO], AREA: [MOVE_TO, WITHDRAW_TO],
                    ACTOR: [TALK], FACT: []}
# 每種 kind 合法狀態（小狀態機）
_STATES = {
    OBJECT: ["unseen", "noticed", "inspected", "taken", "used"],
    EXIT: ["unknown", "known", "available", "locked", "blocked", "used"],
    AREA: ["unknown", "known", "visited", "current"],
    ACTOR: ["unknown", "present", "absent", "talked"],
    FACT: ["asserted"],
}

# 「外面 / 安全區」——結構性區域（撤退目的地），主題無關
SAFE_ZONE_AREA_ID = "area.outside_dock"
SAFE_ZONE_LABEL = "外面（安全區）"


def slug(label: str) -> str:
    s = re.sub(r"\s+", "_", (label or "").strip())
    s = re.sub(r"[^0-9A-Za-z一-鿿_]", "", s)
    return s[:32] or "x"


@dataclass
class Entity:
    id: str
    kind: str
    label: str
    state: str = ""
    props: dict = field(default_factory=dict)
    affords: list = field(default_factory=list)
    origin: str = "kernel"          # kernel | story | npc | extractor


@dataclass
class Affordance:
    """一個「此刻對某 Entity 可做的行動」(observation 的 available_next 由此來)。"""
    verb: str
    entity_id: str
    label: str


@dataclass
class WorldDelta:
    """對世界模型的結構化變更(story/npc 可輸出 entity_delta 清單,model.apply 套用)。"""
    op: str                          # register | set_state | move_current
    kind: str | None = None          # for register
    label: str | None = None         # for register
    entity_id: str | None = None     # for set_state / move_current / register(指定 id)
    state: str | None = None         # for set_state
    affords: list | None = None      # for register
    origin: str = "story"
    props: dict | None = None

    def to_dict(self) -> dict:
        return {k: v for k, v in asdict(self).items() if v is not None}


class WorldModel:
    """實體 store。平行於 kernel 的記憶/觀測層;不取代 kernel 的場景推進。"""

    def __init__(self):
        self.entities: dict[str, Entity] = {}
        self.current_area: str | None = None

    # ── 登記 / 查詢 ──────────────────────────────────────────────────────────
    def register(self, kind: str, label: str, *, id: str | None = None,
                 state: str | None = None, props: dict | None = None,
                 affords: list | None = None, origin: str = "story") -> Entity:
        if kind not in KINDS:
            raise ValueError(f"unknown entity kind: {kind}")
        eid = id or f"{kind}.{slug(label)}"
        if eid in self.entities:                       # 已存在 → 不重複登記(可補 label)
            e = self.entities[eid]
            if label and not e.label:
                e.label = label
            return e
        e = Entity(id=eid, kind=kind, label=label,
                   state=state or _DEFAULT_STATE.get(kind, "known"),
                   props=dict(props or {}),
                   affords=list(affords if affords is not None else _DEFAULT_AFFORDS.get(kind, [])),
                   origin=origin)
        self.entities[eid] = e
        return e

    def get(self, eid: str) -> Entity | None:
        return self.entities.get(eid)

    def by_kind(self, kind: str) -> list[Entity]:
        return [e for e in self.entities.values() if e.kind == kind]

    def find(self, ref: str, kind: str | None = None) -> Entity | None:
        """以 id 或 label(雙向子字串)解析實體。"""
        if not ref:
            return None
        if ref in self.entities and (kind is None or self.entities[ref].kind == kind):
            return self.entities[ref]
        ref_ns = ref.replace(" ", "")
        for e in self.entities.values():
            if kind and e.kind != kind:
                continue
            lab = (e.label or "").replace(" ", "")
            if lab and (lab in ref_ns or ref_ns in lab):
                return e
        return None

    # ── 狀態轉移 ─────────────────────────────────────────────────────────────
    def set_state(self, eid: str, state: str) -> bool:
        e = self.entities.get(eid)
        if e is None:
            return False
        if state not in _STATES.get(e.kind, [state]):   # 未知狀態仍允許(寬鬆),但合法集優先
            pass
        e.state = state
        return True

    def inspect(self, ref: str) -> Entity | None:
        """玩家檢查某物件 → 標 inspected(世界記得他查過)。"""
        e = self.find(ref, kind=OBJECT)
        if e is not None:
            e.state = "inspected"
        return e

    def move_to(self, area_ref: str, *, negated: list | None = None) -> tuple[bool, str]:
        """切換 current_area。respects NegativeIntentGuard:目標被否定 → 不移動。

        回傳 (moved, reason)。area_ref 可為 id 或 label;不存在則略過(不亂建)。
        """
        from core.narrative.negative_intent import is_negated
        if negated and is_negated(area_ref, negated):
            return (False, "negated")
        e = self.find(area_ref, kind=AREA)
        if e is None:
            return (False, "unknown_area")
        if negated and is_negated(e.label, negated):
            return (False, "negated")
        self._set_current(e.id)
        return (True, "moved")

    def withdraw_to_safe_zone(self) -> str:
        """撤退到結構性安全區(若無則建);回傳該 area id。"""
        if SAFE_ZONE_AREA_ID not in self.entities:
            self.register(AREA, SAFE_ZONE_LABEL, id=SAFE_ZONE_AREA_ID, origin="kernel")
        self._set_current(SAFE_ZONE_AREA_ID)
        return SAFE_ZONE_AREA_ID

    def set_current_area(self, area_id: str, label: str | None = None) -> Entity:
        """kernel 場景變動 → 同步 current_area(沒有則登記)。"""
        e = self.entities.get(area_id) or self.register(
            AREA, label or area_id, id=area_id, origin="kernel")
        self._set_current(e.id)
        return e

    def _set_current(self, area_id: str):
        for e in self.by_kind(AREA):
            if e.state == "current":
                e.state = "visited"
        self.current_area = area_id
        if area_id in self.entities:
            self.entities[area_id].state = "current"

    # ── affordances / 投影 ───────────────────────────────────────────────────
    def affordances_here(self) -> list[Affordance]:
        """當前區域可做的事(物件 inspect/take、在場 NPC talk、出口 move_to)。"""
        out: list[Affordance] = []
        for e in self.entities.values():
            loc = e.props.get("area")
            if e.kind in (OBJECT, ACTOR, EXIT) and loc and loc != self.current_area:
                continue                              # 有標定位置且不在當前區域 → 跳過
            if e.kind == OBJECT and e.state in ("taken", "used"):
                continue
            for verb in e.affords:
                out.append(Affordance(verb=verb, entity_id=e.id, label=e.label))
        return out

    # ── 套用 WorldDelta(story/npc entity_delta 的入口)─────────────────────────
    def apply(self, delta) -> Entity | None:
        d = delta if isinstance(delta, dict) else delta.to_dict()
        op = d.get("op")
        if op == "register":
            return self.register(d.get("kind"), d.get("label", ""), id=d.get("entity_id"),
                                 state=d.get("state"), props=d.get("props"),
                                 affords=d.get("affords"), origin=d.get("origin", "story"))
        if op == "set_state" and d.get("entity_id") and d.get("state"):
            self.set_state(d["entity_id"], d["state"]); return self.entities.get(d["entity_id"])
        if op == "move_current" and d.get("entity_id"):
            self.set_current_area(d["entity_id"]); return self.entities.get(d["entity_id"])
        return None

    def apply_deltas(self, deltas) -> list:
        return [self.apply(d) for d in (deltas or [])]

    # ── 持久化(存進 game_meta)────────────────────────────────────────────────
    def to_dict(self) -> dict:
        return {"current_area": self.current_area,
                "entities": {eid: asdict(e) for eid, e in self.entities.items()}}

    @classmethod
    def from_dict(cls, data: dict | None) -> "WorldModel":
        m = cls()
        data = data or {}
        m.current_area = data.get("current_area")
        for eid, ed in (data.get("entities") or {}).items():
            m.entities[eid] = Entity(**ed)
        return m
