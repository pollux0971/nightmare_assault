"""core.world.model — 抽象世界模型（WorldModel 骨架）。

主題無關的實體模型:世界裡所有「可被指涉的東西」都是 Entity 的實例,玩家行動 = 對某個
Entity 套用一個 affordance。這一層先當**平行的記憶 + 觀測層**(不取代 kernel),把
移動 / 物件 / 出口 / NPC / 可檢查事實收成同一機制。

四個核心型別:Entity / Affordance / WorldDelta / WorldModel(store)。
Entity kind 先只支援:area / exit / object / actor / fact。

對應 docs/player-sovereignty-principles.md §二、15-player-sovereignty.md。
"""
from __future__ import annotations

import logging
import re
from dataclasses import dataclass, field, asdict

log = logging.getLogger("nightmare.worldmodel")

# ── kinds ────────────────────────────────────────────────────────────────────
AREA, EXIT, OBJECT, ACTOR, FACT = "area", "exit", "object", "actor", "fact"
KINDS = {AREA, EXIT, OBJECT, ACTOR, FACT}

# ── affordances（玩家對某 Entity 能做的事）────────────────────────────────────
INSPECT, TAKE, USE, MOVE_TO, MOVE_THROUGH, WITHDRAW_TO, TALK, MANIPULATE = (
    "inspect", "take", "use", "move_to", "move_through", "withdraw_to", "talk", "manipulate")

# 每種 kind 的預設初始狀態 + 預設 affordances（抽象狀態機）
_DEFAULT_STATE = {OBJECT: "noticed", EXIT: "known", AREA: "known",
                  ACTOR: "present", FACT: "asserted"}
_DEFAULT_AFFORDS = {OBJECT: [INSPECT], EXIT: [MOVE_THROUGH], AREA: [MOVE_TO, WITHDRAW_TO],
                    ACTOR: [TALK], FACT: []}
# exit 可通行的狀態（locked/blocked/used 不可 move_through）
_PASSABLE_EXIT_STATES = {"known", "available"}
# 每種 kind 合法狀態（小狀態機）
_STATES = {
    OBJECT: ["unseen", "noticed", "inspected", "taken", "used"],
    EXIT: ["unknown", "known", "available", "locked", "blocked", "used"],
    AREA: ["unknown", "known", "visited", "current"],
    ACTOR: ["unknown", "present", "absent", "talked"],
    FACT: ["asserted"],
}

# ── Area roles（主題無關角色；取代硬編碼地名）────────────────────────────────
ROLE_SAFE_ZONE = "safe_zone"        # 撤退/整理的結構性安全區
ROLE_SITE = "site"                  # 調查現場（主要地點）
ROLE_ENTRY = "entry"                # 起始區域
ROLE_ACTIVE_AREA = "active_area"    # 玩家撤退前所在、返回時的目標區域
ROLE_CAMPAIGN_EXIT = "campaign_exit"  # 結束本次調查的離場區域
AREA_ROLES = {ROLE_SAFE_ZONE, ROLE_SITE, ROLE_ENTRY, ROLE_ACTIVE_AREA, ROLE_CAMPAIGN_EXIT}

# 預設結構性安全區（**主題無關** id/label；只在沒有 role=safe_zone 區域時當預設）
SAFE_ZONE_AREA_ID = "area.safe_zone"
SAFE_ZONE_LABEL = "外面（安全區）"
# 舊存檔相容：以前硬寫的 outside_dock 仍視為 safe_zone（is_safe_zone 認得）
LEGACY_SAFE_ZONE_ID = "area.outside_dock"

# story 結構化 entity_delta 只准登記/改這些 kind（area/exit 由 kernel 圖擁有，story 不得自由新增）
STORY_ENTITY_KINDS = {OBJECT, ACTOR, FACT}
# story entity_delta 每 beat 最多新增/變更的實體數
STORY_DELTA_CAP = 3
# story 允許的 op（move_current 屬 area/kernel 權限，排除）
_STORY_ALLOWED_OPS = ("register", "set_state")


def normalize_entity_delta_dict(d) -> dict | None:
    """正規化外部 entity_delta dict：`id` 別名 → `entity_id`，**內部一律用 entity_id**。

    - LLM 常吐 `id` 而非 `entity_id` → 自動搬成 entity_id（不改 WorldDelta 欄位名）。
    - `id` 與 `entity_id` 同時存在：相同 → 正常（丟掉 id）；**不同 → 拒絕該 delta（回 None）+ debug warning**。
    - 非 dict → None。回**新 dict**（移除 id；不就地改原物件），讓 malformed/衝突 delta 被丟棄不污染。
    """
    if not isinstance(d, dict):
        return None
    if "id" not in d:
        return d
    out = dict(d)
    alias = out.pop("id", None)
    eid = out.get("entity_id")
    if eid is None:
        if alias is not None:
            out["entity_id"] = alias
    elif alias is not None and eid != alias:
        log.warning("entity_delta id 衝突，拒絕該 delta：id=%r entity_id=%r", alias, eid)
        return None                                  # 衝突 → 拒絕，不污染 WorldModel
    return out


def coerce_entity_deltas(raw, *, allowed_kinds=STORY_ENTITY_KINDS,
                         cap: int = STORY_DELTA_CAP,
                         allow_ops=_STORY_ALLOWED_OPS) -> list:
    """把 story 輸出的原始 entity_delta 清洗成安全的 WorldDelta 清單。

    malformed item 一律丟棄(不污染 WorldModel)；非 list → []；超過 cap 截斷。
    register 需 kind ∈ allowed_kinds 且有 label；set_state 需 entity_id + state。
    """
    out: list = []
    if not isinstance(raw, list):
        return out
    for item in raw:
        if len(out) >= cap:
            break
        item = normalize_entity_delta_dict(item)     # id→entity_id；衝突/非 dict → 丟棄
        if item is None:
            continue
        op = item.get("op")
        if op not in allow_ops:
            continue
        if op == "register":
            kind, label = item.get("kind"), item.get("label")
            if kind not in allowed_kinds or not (isinstance(label, str) and label.strip()):
                continue
            eid = item.get("entity_id") or item.get("id")
            affords = item.get("affords")
            out.append(WorldDelta(
                op="register", kind=kind, label=label.strip(),
                entity_id=eid if isinstance(eid, str) else None,
                state=item.get("state") if isinstance(item.get("state"), str) else None,
                affords=list(affords) if isinstance(affords, list) else None,
                props=item.get("props") if isinstance(item.get("props"), dict) else None,
                origin="story"))
        elif op == "set_state":
            eid, state = item.get("entity_id") or item.get("id"), item.get("state")
            if not (isinstance(eid, str) and isinstance(state, str) and state.strip()):
                continue
            out.append(WorldDelta(op="set_state", entity_id=eid, state=state.strip(),
                                  origin="story"))
    return out


# NPC 結構化 entity_delta 只准登記 fact / actor（NPC 不得新增 object/area/exit/真相）
NPC_ENTITY_KINDS = {FACT, ACTOR}
NPC_FACT_CONFIDENCE = "npc_claim"        # NPC 講的事是「主張」，非確證真相


def coerce_npc_entity_deltas(raw, *, npc_id: str, cap: int = STORY_DELTA_CAP) -> list:
    """把 NPC-chat 輸出的原始 entity_delta 清洗成安全的 WorldDelta 清單（只 fact/actor）。

    - 只准 kind ∈ {fact, actor}；object/area/exit 一律丟棄（NPC 不得新增地圖/出口/物件）。
    - fact entity **強制**帶 props.source=npc_id / props.confidence=npc_claim、origin=npc。
    - malformed item 丟棄(不污染)；非 list → []；超過 cap 截斷。
    - **不**處理 truth_id / reveal——NPC fact 寫進 WorldModel 不得自動 grant 真相。
    """
    out: list = []
    if not isinstance(raw, list):
        return out
    for item in raw:
        if len(out) >= cap:
            break
        item = normalize_entity_delta_dict(item)     # id→entity_id；衝突/非 dict → 丟棄
        if item is None:
            continue
        op = item.get("op")
        if op not in _STORY_ALLOWED_OPS:
            continue
        if op == "register":
            kind, label = item.get("kind"), item.get("label")
            if kind not in NPC_ENTITY_KINDS or not (isinstance(label, str) and label.strip()):
                continue
            props = dict(item.get("props") if isinstance(item.get("props"), dict) else {})
            props["source"] = npc_id                   # 來源 = 該 NPC
            props["confidence"] = NPC_FACT_CONFIDENCE   # NPC 主張，非確證
            if kind == FACT:
                props.setdefault("category", "npc_claim")
            eid = item.get("entity_id") or item.get("id")
            out.append(WorldDelta(
                op="register", kind=kind, label=label.strip(),
                entity_id=eid if isinstance(eid, str) else None,
                props=props, origin="npc"))
        elif op == "set_state":
            eid, state = item.get("entity_id") or item.get("id"), item.get("state")
            if not (isinstance(eid, str) and isinstance(state, str) and state.strip()):
                continue
            out.append(WorldDelta(op="set_state", entity_id=eid, state=state.strip(),
                                  origin="npc"))
    return out


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
    roles: list = field(default_factory=list)   # area roles（主題無關；safe_zone/site/entry/…）


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
        self.previous_area: str | None = None    # 上一個 current_area（供 spatial「返回上一區」route）
        # P1：dirty-version 計數（供 SpatialProjectionCache 失效判斷；runtime-only，不持久化）
        self._versions: dict[str, int] = {
            "world_version": 0, "area_version": 0, "exit_version": 0,
            "entity_version": 0, "fact_version": 0, "mode_version": 0}

    # ── P1：版本快照（任何結構變動都 bump world_version + 對應 category）──────────
    def _bump(self, *cats: str):
        self._versions["world_version"] += 1
        for c in cats:
            self._versions[c] = self._versions.get(c, 0) + 1

    def _bump_kind(self, kind: str):
        self._bump({AREA: "area_version", EXIT: "exit_version",
                    FACT: "fact_version"}.get(kind, "entity_version"))

    def version_snapshot(self) -> dict:
        """輕量版本快照：WorldModel 一變動就改值，供投影 cache 失效（不含 LLM、不持久化）。"""
        return dict(self._versions)

    # ── 登記 / 查詢 ──────────────────────────────────────────────────────────
    def register(self, kind: str, label: str, *, id: str | None = None,
                 state: str | None = None, props: dict | None = None,
                 affords: list | None = None, origin: str = "story",
                 roles: list | None = None) -> Entity:
        if kind not in KINDS:
            raise ValueError(f"unknown entity kind: {kind}")
        eid = id or f"{kind}.{slug(label)}"
        if eid in self.entities:                       # 已存在 → 不重複登記(可補 label/role)
            e = self.entities[eid]
            if label and not e.label:
                e.label = label
            added_role = False
            for r in (roles or []):                    # 合併 role（不覆蓋既有）
                if r not in e.roles:
                    e.roles.append(r); added_role = True
            if added_role:
                self._bump_kind(e.kind)
            return e
        e = Entity(id=eid, kind=kind, label=label,
                   state=state or _DEFAULT_STATE.get(kind, "known"),
                   props=dict(props or {}),
                   affords=list(affords if affords is not None else _DEFAULT_AFFORDS.get(kind, [])),
                   origin=origin, roles=list(roles or []))
        self.entities[eid] = e
        self._bump_kind(kind)
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
        if e.state != state:
            e.state = state
            self._bump_kind(e.kind)
        return True

    def inspect(self, ref: str) -> Entity | None:
        """玩家檢查某物件 → 標 inspected(世界記得他查過)。"""
        e = self.find(ref, kind=OBJECT)
        if e is not None and e.state != "inspected":
            e.state = "inspected"
            self._bump_kind(OBJECT)
        return e

    def take(self, ref: str) -> Entity | None:
        """玩家撿起/拿走某物件 → 標 taken + props.carried（沉澱進 inventory）。

        已 used 的物件不再回收為 taken（避免把用掉的東西復原）。回被改的 Entity（無則 None）。
        """
        e = self.find(ref, kind=OBJECT)
        if e is None or e.state == "used":
            return None
        if e.state != "taken":
            e.state = "taken"
            e.props["carried"] = True
            self._bump_kind(OBJECT)
        return e

    def tag_entity_area(self, eid: str, area_id: str) -> bool:
        """把物件/NPC 綁到所在區域（供 spatial visible/remote 判斷）。只在未綁定時設。"""
        e = self.entities.get(eid)
        if e is None or e.kind in (AREA, EXIT) or e.props.get("area") or not area_id:
            return False
        e.props["area"] = area_id
        self._bump_kind(e.kind)
        return True

    def set_label(self, eid: str, label: str) -> bool:
        """改某實體顯示名（current_area_label 屬投影輸出 → 經此 setter 才會 bump cache）。"""
        e = self.entities.get(eid)
        if e is None or not label or e.label == label:
            return False
        e.label = label
        self._bump_kind(e.kind)
        return True

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

    def register_exit(self, label: str, *, leads_to: str | None = None,
                      from_area: str | None = None, state: str = "known",
                      id: str | None = None, origin: str = "kernel") -> Entity:
        """登記一條 exit/route（WorldModel owns exits）。leads_to=目的地 area id；from_area=所在區域。"""
        props = {"leads_to": leads_to, "area": from_area or self.current_area}
        return self.register(EXIT, label, id=id, state=state, props=props, origin=origin)

    def set_exit_state(self, exit_ref: str, state: str) -> bool:
        e = self.find(exit_ref, kind=EXIT)
        if e is None:
            return False
        if e.state != state:
            e.state = state
            self._bump("exit_version")
        return True

    def move_through(self, exit_ref: str, *, negated: list | None = None) -> tuple[bool, str]:
        """穿過一條 exit/route。respects 否定 + exit 狀態（locked/blocked/used 不可通行）。

        回傳 (moved, reason)。reason ∈ moved_through / unknown_exit / negated / locked / blocked / used。
        """
        from core.narrative.negative_intent import is_negated
        e = self.find(exit_ref, kind=EXIT)
        if e is None:
            return (False, "unknown_exit")
        if negated and (is_negated(e.label, negated)
                        or is_negated(str(e.props.get("leads_to") or ""), negated)):
            return (False, "negated")
        if e.state not in _PASSABLE_EXIT_STATES:
            return (False, e.state)                    # locked / blocked / used → 不移動
        dest = e.props.get("leads_to")
        if dest:
            self.set_current_area(dest)
        e.state = "used"
        self._bump("exit_version")
        return (True, "moved_through")

    def exits_here(self) -> list[Entity]:
        """當前區域可見的 exit/route。"""
        return self.exits_from(self.current_area)

    def exits_from(self, area_id: str | None) -> list[Entity]:
        """從指定區域出發可見的 exit/route（props.area 未標 → 視為任意區域可見）。"""
        return [e for e in self.by_kind(EXIT)
                if e.props.get("area") in (None, area_id)]

    # ── Area roles（主題無關角色查詢）────────────────────────────────────────
    def set_area_role(self, area_id: str, role: str, *, exclusive: bool = False) -> bool:
        """給某 area 加上 role。exclusive=True → 先把該 role 從其他 area 移除（如 active_area）。"""
        e = self.entities.get(area_id)
        if e is None or e.kind != AREA:
            return False
        if exclusive:
            for other in self.by_kind(AREA):
                if other.id != area_id and role in other.roles:
                    other.roles.remove(role)
        if role not in e.roles:
            e.roles.append(role)
        self._bump("area_version")
        return True

    def areas_with_role(self, role: str) -> list[Entity]:
        return [e for e in self.by_kind(AREA) if role in (e.roles or [])]

    def area_with_role(self, role: str) -> Entity | None:
        a = self.areas_with_role(role)
        return a[0] if a else None

    def safe_zone_id(self) -> str:
        """安全區 id：優先 role=safe_zone；無 role 時 fallback 舊常數（相容），否則用主題無關預設。"""
        e = self.area_with_role(ROLE_SAFE_ZONE)
        if e is not None:
            return e.id
        if LEGACY_SAFE_ZONE_ID in self.entities:       # 舊存檔相容
            return LEGACY_SAFE_ZONE_ID
        return SAFE_ZONE_AREA_ID

    def site_area_id(self) -> str | None:
        """調查現場 id：優先 active_area（撤退前所在）→ site → entry。"""
        for role in (ROLE_ACTIVE_AREA, ROLE_SITE, ROLE_ENTRY):
            e = self.area_with_role(role)
            if e is not None:
                return e.id
        return None

    def entry_area_id(self) -> str | None:
        e = self.area_with_role(ROLE_ENTRY)
        return e.id if e is not None else None

    def is_safe_zone(self, area_id: str | None) -> bool:
        """是否為安全區：role=safe_zone 或（相容）預設/舊常數 id。"""
        if not area_id:
            return False
        e = self.entities.get(area_id)
        if e is not None and ROLE_SAFE_ZONE in (e.roles or []):
            return True
        return area_id in (SAFE_ZONE_AREA_ID, LEGACY_SAFE_ZONE_ID)

    def site_label(self, default: str = "現場") -> str:
        """調查現場顯示名（給 ReviewMode 文案；無則用通用詞）。"""
        sid = self.site_area_id()
        e = self.entities.get(sid) if sid else None
        return (e.label if e is not None and e.label else default) or default

    def withdraw_to_safe_zone(self) -> str:
        """撤退到結構性安全區（role=safe_zone；若無則用預設 id 建並標 role）;回傳該 area id。"""
        sid = self.safe_zone_id()
        if sid not in self.entities:
            self.register(AREA, SAFE_ZONE_LABEL, id=sid, roles=[ROLE_SAFE_ZONE], origin="kernel")
        else:
            self.set_area_role(sid, ROLE_SAFE_ZONE)
        self._set_current(sid)
        return sid

    def set_current_area(self, area_id: str, label: str | None = None) -> Entity:
        """kernel 場景變動 → 同步 current_area(沒有則登記)。"""
        e = self.entities.get(area_id) or self.register(
            AREA, label or area_id, id=area_id, origin="kernel")
        self._set_current(e.id)
        return e

    def _set_current(self, area_id: str):
        if self.current_area == area_id and (
                area_id in self.entities and self.entities[area_id].state == "current"):
            return                                       # 無變化 → 不 bump
        if self.current_area and self.current_area != area_id:
            self.previous_area = self.current_area       # 記上一區（供「返回上一區」route）
        for e in self.by_kind(AREA):
            if e.state == "current":
                e.state = "visited"
        self.current_area = area_id
        if area_id in self.entities:
            self.entities[area_id].state = "current"
        self._bump("area_version")

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
        d = delta if isinstance(delta, dict) else (
            delta.to_dict() if hasattr(delta, "to_dict") else None)
        d = normalize_entity_delta_dict(d)           # id→entity_id；衝突/非 dict → 丟棄不套用
        if d is None:
            return None
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

    # ── story/npc 結構化 entity_delta（kind-guarded，回傳被改的 entity id）─────────
    def apply_entity_delta(self, delta, *, allowed_kinds=None) -> Entity | None:
        """套用單一已清洗的 entity_delta，並 guard kind（story 不得碰 area/exit）。

        register：kind 須 ∈ allowed_kinds。set_state：目標實體 kind 須 ∈ allowed_kinds。
        回傳被新增/變更的 Entity（無效則 None）。
        """
        d = delta if isinstance(delta, dict) else (
            delta.to_dict() if hasattr(delta, "to_dict") else None)
        d = normalize_entity_delta_dict(d)           # id→entity_id；衝突/非 dict → 丟棄
        if d is None:
            return None
        op = d.get("op")
        if op == "register":
            if allowed_kinds is not None and d.get("kind") not in allowed_kinds:
                return None
            return self.apply(d)
        if op == "set_state":
            eid, state = d.get("entity_id"), d.get("state")
            e = self.entities.get(eid) if eid else None
            if e is None or not state:
                return None
            if allowed_kinds is not None and e.kind not in allowed_kinds:
                return None                            # 不准 story 改 area/exit 狀態
            self.set_state(eid, state)
            return e
        return None                                    # move_current 等 → story 無權

    def apply_story_deltas(self, deltas, *, allowed_kinds=STORY_ENTITY_KINDS) -> list:
        """套用一串 story entity_delta，回傳實際被新增/變更的 entity id（供 changed_entities）。"""
        changed: list = []
        for d in deltas or []:
            e = self.apply_entity_delta(d, allowed_kinds=allowed_kinds)
            if e is not None and e.id not in changed:
                changed.append(e.id)
        return changed

    # ── 「此處」的實體 / 可互動物 投影 ────────────────────────────────────────
    def entities_here(self) -> list[Entity]:
        """當前區域的實體（area 本身除外；有標定位置且不在此區域者排除）。"""
        out: list[Entity] = []
        for e in self.entities.values():
            if e.kind == AREA:
                continue
            loc = e.props.get("area")
            if loc and loc != self.current_area:
                continue
            out.append(e)
        return out

    def interactables_here(self) -> list[Entity]:
        """當前區域**可互動**的實體（物件/NPC/出口；已 taken/used 的物件排除）。"""
        out: list[Entity] = []
        for e in self.entities_here():
            if e.kind not in (OBJECT, ACTOR, EXIT):
                continue
            if e.kind == OBJECT and e.state in ("taken", "used"):
                continue
            out.append(e)
        return out

    # ── 持久化(存進 game_meta)────────────────────────────────────────────────
    def to_dict(self) -> dict:
        return {"current_area": self.current_area,
                "previous_area": self.previous_area,
                "entities": {eid: asdict(e) for eid, e in self.entities.items()}}

    @classmethod
    def from_dict(cls, data: dict | None) -> "WorldModel":
        m = cls()
        data = data or {}
        m.current_area = data.get("current_area")
        m.previous_area = data.get("previous_area")
        for eid, ed in (data.get("entities") or {}).items():
            m.entities[eid] = Entity(**ed)
        return m
