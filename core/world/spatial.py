"""core.world.spatial — Spatial WorldModel Projection（確定性、唯讀、可快取）。

對應 nightmare-assault-spatial-worldmodel-planning-patch-v0.6（P0–P4）。

核心原則（docs/00-executive-decision）：
  WorldModel 是唯一事實來源；Spatial Projection 是它的**確定性投影（衍生快取）**，
  **不**呼叫 LLM、**不**新增 area/exit/fact/object、**不**做幾何/pathfinding。
  遊戲迴圈永遠不等投影；昂貴的 mental_map_text 用確定性模板（async LLM 為選用，遲到就用 fallback）。

投影只回答空間問題：current_area / routes_from_here / blocked_routes / safe_retreat_routes /
visible_entities / known_remote_entities / mental_map_text。
"""
from __future__ import annotations

from dataclasses import dataclass, field

from core.world.model import AREA, EXIT, FACT, OBJECT, ACTOR

# exit 狀態分類（spec 04）：可通行 vs 受阻
PASSABLE_EXIT_STATES = {"known", "available", "used"}
BLOCKED_EXIT_STATES = {"locked", "blocked", "unknown", "unsafe", "jammed"}

# P4 觀測預算（docs/06）
DEFAULT_LIMITS = {
    "visible_entities": 20, "known_remote_entities": 20,
    "routes_from_here": 12, "blocked_routes": 12, "mental_map_text": 800,
    "spatial_summary": 600,
}

# Spatial UX：玩家/QA 可讀摘要
SUMMARY_MAX_CHARS = 600
SPATIAL_SUMMARY_SOURCE = "deterministic_projection"
# 受阻狀態 → 玩家面中文
_BLOCKED_STATE_ZH = {"locked": "鎖住", "blocked": "受阻", "unknown": "情況不明",
                     "unsafe": "不安全", "jammed": "卡死"}


@dataclass
class SpatialRoute:
    exit_id: str
    label: str
    from_area: str | None
    to_area: str | None
    state: str
    requires: list = field(default_factory=list)
    roles: list = field(default_factory=list)

    def to_dict(self) -> dict:
        d = {"exit_id": self.exit_id, "label": self.label, "to_area": self.to_area,
             "state": self.state}
        if self.requires:
            d["requires"] = list(self.requires)
        return d


@dataclass
class SpatialEntityView:
    id: str
    kind: str
    label: str
    state: str
    roles: list = field(default_factory=list)
    area: str | None = None

    def to_dict(self) -> dict:
        return {"id": self.id, "kind": self.kind, "label": self.label, "state": self.state}


@dataclass
class SpatialProjection:
    current_area: str | None
    current_area_label: str
    current_area_roles: list
    routes_from_here: list
    blocked_routes: list
    safe_retreat_routes: list
    visible_entities: list
    known_remote_entities: list
    mental_map_text: str
    versions: dict
    counts: dict = field(default_factory=dict)
    truncated: dict = field(default_factory=dict)
    player_summary: str = ""               # Spatial UX：玩家/QA 可讀短摘要（確定性）
    player_summary_truncated: bool = False

    def to_debug_dict(self) -> dict:
        """observation.spatial_debug 形狀（docs/06 example）。"""
        return {
            "current_area": self.current_area,
            "current_area_label": self.current_area_label,
            "current_area_roles": list(self.current_area_roles),
            "routes_from_here": [r.to_dict() for r in self.routes_from_here],
            "blocked_routes": [r.to_dict() for r in self.blocked_routes],
            "safe_retreat_routes": [r.to_dict() for r in self.safe_retreat_routes],
            "visible_entities": [e.to_dict() for e in self.visible_entities],
            "known_remote_entities": [e.to_dict() for e in self.known_remote_entities],
            "mental_map_text": self.mental_map_text,
            # Spatial UX：玩家/QA 面板摘要（不餵 story；deterministic）
            "spatial_summary": self.player_summary,
            "spatial_summary_truncated": self.player_summary_truncated,
            "spatial_summary_source": SPATIAL_SUMMARY_SOURCE,
            "counts": dict(self.counts),
            "truncated": dict(self.truncated),
            "versions": dict(self.versions),
        }


def _exit_from_to(e) -> tuple:
    """本專案 exit：props.area=from_area，props.leads_to=to_area（相容 to_area/from_area 別名）。"""
    from_a = e.props.get("area") or e.props.get("from_area")
    to_a = e.props.get("leads_to") or e.props.get("to_area")
    return from_a, to_a


def build_spatial_projection(world, *, limits: dict | None = None,
                             exploration_mode: str = "active_exploration",
                             focus_id: str | None = None) -> SpatialProjection:
    """從 WorldModel 建**確定性唯讀**投影（不改 WorldModel、不呼叫 LLM）。

    focus_id：當前焦點實體 id。已 taken/used 的物件一律不算地面 visible（已進 inventory），
    **除非當前 focus 正在檢視它**（剛撿起/正在端詳）——此時仍顯示在眼前。
    """
    lim = {**DEFAULT_LIMITS, **(limits or {})}
    cur = world.current_area
    cur_e = world.get(cur) if cur else None
    cur_label = (getattr(cur_e, "label", None) or cur or "未知之地")
    cur_roles = list(getattr(cur_e, "roles", []) or [])

    # ── routes / blocked / safe_retreat（只看本區域出發的 exit）──────────────────
    passable, blocked, safe_retreat = [], [], []
    for e in world.exits_from(cur):
        from_a, to_a = _exit_from_to(e)
        route = SpatialRoute(exit_id=e.id, label=e.label, from_area=from_a, to_area=to_a,
                             state=e.state, requires=list(e.props.get("requires", []) or []),
                             roles=list(getattr(e, "roles", []) or []))
        if e.state in PASSABLE_EXIT_STATES:
            passable.append(route)
            if to_a and world.is_safe_zone(to_a):       # 只有「可通行且通往安全區」才算退路
                safe_retreat.append(route)
        else:                                            # 受阻（含 unknown/unsafe/jammed）
            blocked.append(route)

    # ── visible（此區） vs known_remote（已知但不在眼前）──────────────────────────
    visible, remote = [], []
    for ent in world.entities.values():
        if ent.kind in (AREA, EXIT):
            continue
        area = ent.props.get("area")
        view = SpatialEntityView(id=ent.id, kind=ent.kind, label=ent.label, state=ent.state,
                                 roles=list(getattr(ent, "roles", []) or []), area=area)
        if ent.kind == FACT:
            remote.append(view)                          # fact 是「知道」非「看見」→ 一律 remote
        elif ent.kind == OBJECT and ent.state in ("taken", "used") and ent.id != focus_id:
            continue                                     # 已拿走/用掉 → 進 inventory，不算地面 visible
                                                         # （除非正是當前 focus：剛撿起/正在端詳 → 仍可見）
        elif area in (None, cur):
            visible.append(view)                         # 物件/NPC 綁在本區（或未綁定）→ 可見
        else:
            remote.append(view)                          # 綁在別區 → 已知但不在眼前

    counts = {"visible_entities": len(visible), "known_remote_entities": len(remote),
              "routes_from_here": len(passable), "blocked_routes": len(blocked)}
    truncated = {
        "visible_entities": len(visible) > lim["visible_entities"],
        "known_remote_entities": len(remote) > lim["known_remote_entities"],
        "routes_from_here": len(passable) > lim["routes_from_here"],
        "blocked_routes": len(blocked) > lim["blocked_routes"],
    }
    visible = visible[:lim["visible_entities"]]
    remote = remote[:lim["known_remote_entities"]]
    passable = passable[:lim["routes_from_here"]]
    blocked = blocked[:lim["blocked_routes"]]
    safe_retreat = safe_retreat[:lim["routes_from_here"]]

    mental = deterministic_mental_map_text(cur_label, passable, blocked, visible, remote)
    if len(mental) > lim["mental_map_text"]:
        mental = mental[:lim["mental_map_text"]]

    proj = SpatialProjection(
        current_area=cur, current_area_label=cur_label, current_area_roles=cur_roles,
        routes_from_here=passable, blocked_routes=blocked, safe_retreat_routes=safe_retreat,
        visible_entities=visible, known_remote_entities=remote, mental_map_text=mental,
        versions=world.version_snapshot(), counts=counts, truncated=truncated)
    proj.player_summary, proj.player_summary_truncated = player_facing_spatial_summary(
        proj, max_chars=lim.get("spatial_summary", SUMMARY_MAX_CHARS))
    return proj


def player_facing_spatial_summary(projection: "SpatialProjection", *,
                                  max_chars: int = SUMMARY_MAX_CHARS) -> tuple:
    """由 SpatialProjection **確定性**生成玩家/QA 可讀短摘要（**不呼叫 LLM**、不餵 story）。

    分段（只列非空）：目前位置 / 可走路線 / 被阻擋路線 / 安全撤退路線 / 眼前可互動物 /
    已知但不在眼前。回 (text, truncated)；truncated = 任一清單被預算截斷 或 文字超出字數上限。
    只用投影內容，不新增任何地圖元素。
    """
    P = projection
    lines = [f"目前位置：{P.current_area_label or '未知之地'}"]
    if P.routes_from_here:
        lines.append("可走路線：" + "、".join(r.label for r in P.routes_from_here))
    else:
        lines.append("可走路線：（沒有明顯可走的出口）")
    if P.blocked_routes:
        lines.append("被阻擋路線：" + "、".join(
            f"{r.label}（{_BLOCKED_STATE_ZH.get(r.state, r.state)}）" for r in P.blocked_routes))
    if P.safe_retreat_routes:
        lines.append("安全撤退路線：" + "、".join(r.label for r in P.safe_retreat_routes))
    # 眼前可互動物：可見的物件/NPC（排除已被拿走/用掉的物件）
    interactable = [e for e in P.visible_entities
                    if not (e.kind == OBJECT and e.state in ("taken", "used"))]
    if interactable:
        lines.append("眼前可互動物：" + "、".join(e.label for e in interactable))
    if P.known_remote_entities:
        lines.append("已知但不在眼前：" + "、".join(e.label for e in P.known_remote_entities))

    text = "\n".join(lines)
    truncated = any(P.truncated.values()) if isinstance(P.truncated, dict) else False
    if len(text) > max_chars:
        text = text[:max_chars].rstrip() + "…"
        truncated = True
    return text, truncated


def deterministic_mental_map_text(current_area_label, routes, blocked, visible, remote) -> str:
    """確定性 mental_map_text 模板（docs/05）——只用投影內容，不新增任何名詞/路線/真相。"""
    parts = [f"你目前位於：{current_area_label}。"]
    if routes:
        parts.append("可前往：" + "、".join(r.label for r in routes) + "。")
    if blocked:
        parts.append("受阻路線：" + "、".join(r.label for r in blocked) + "。")
    if visible:
        parts.append("眼前可見：" + "、".join(e.label for e in visible) + "。")
    if remote:
        parts.append("你知道但不在眼前：" + "、".join(e.label for e in remote[:8]) + "。")
    return "\n".join(parts)


# ── P1：dirty-version cache（不變則回快取；WorldModel 一變動就重算）────────────
class SpatialProjectionCache:
    """以 (current_area, version_snapshot, profile) 為 key 的單槽快取（確定性投影專用）。"""

    def __init__(self):
        self._key = None
        self._value: SpatialProjection | None = None
        self.hits = 0
        self.misses = 0

    def make_key(self, world, profile: str = "default") -> tuple:
        v = world.version_snapshot()
        return (world.current_area, profile,
                v.get("world_version", 0), v.get("area_version", 0),
                v.get("exit_version", 0), v.get("entity_version", 0),
                v.get("fact_version", 0), v.get("mode_version", 0))

    def get_or_build(self, world, builder, profile: str = "default") -> SpatialProjection:
        key = self.make_key(world, profile=profile)
        if self._key == key and self._value is not None:
            self.hits += 1
            return self._value
        self.misses += 1
        self._value = builder(world)
        self._key = key
        return self._value


# ── P4（選用）：async mental-map worker——非阻塞地用 LLM 潤飾 mental_map_text ─────
def projection_label_whitelist(projection: SpatialProjection) -> set:
    """投影內所有可被 mental_map 引用的合法標籤（當前區域名 + 路線 + 實體）。

    供呼叫端做「潤飾文不得引入投影外具名路線/實體」的更嚴格比對。
    """
    known = {(projection.current_area_label or "").replace(" ", "")}
    for r in (projection.routes_from_here + projection.blocked_routes
              + projection.safe_retreat_routes):
        known.add((r.label or "").replace(" ", ""))
    for e in (projection.visible_entities + projection.known_remote_entities):
        known.add((e.label or "").replace(" ", ""))
    known.discard("")
    return known


def validate_mental_map_summary(text: str, projection: SpatialProjection,
                                *, max_chars: int = 800) -> bool:
    """LLM 潤飾結果驗證（docs/05）。

    目前強制：非空 + 不超過字數預算（worker 失敗/超長 → 呼叫端回退確定性文字）。
    呼叫端如需更嚴格的「不得引入投影外具名路線/實體」，用 `projection_label_whitelist(projection)`
    自行比對；通用 prose 的具名項抽取（NER）不在本 MVP 範圍——MentalMapWorker 預設未接線。
    """
    return bool(text) and len(text) <= max_chars


class MentalMapWorker:
    """背景潤飾 mental_map_text 的 worker（**遊戲迴圈永不 await**）。

    - request_refresh：投一個 job（佇列滿就丟棄，不拖慢遊戲）。
    - get_text：回快取潤飾文；沒有就回傳呼叫端給的 fallback（確定性文字）。
    - worker **不寫 WorldModel**、潤飾失敗/超時 → 用 deterministic fallback。
    """

    def __init__(self, summarizer, *, max_chars: int = 800, queue_size: int = 16):
        import queue
        import threading
        self._summarizer = summarizer          # callable(projection_json, deterministic_text) -> str
        self._max_chars = max_chars
        self._jobs = queue.Queue(maxsize=queue_size)
        self._cache: dict[str, str] = {}
        self._stop = False
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()

    def request_refresh(self, key: str, projection: SpatialProjection):
        import queue
        try:
            self._jobs.put_nowait((key, projection))
        except queue.Full:
            pass                                # 寧可丟工作也不拖慢遊戲

    def get_text(self, key: str, fallback: str) -> str:
        return self._cache.get(key, fallback)

    def stop(self):
        self._stop = True

    def _run(self):
        while not self._stop:
            try:
                key, projection = self._jobs.get(timeout=0.5)
            except Exception:
                continue
            fallback = projection.mental_map_text
            try:
                text = self._summarizer(projection.to_debug_dict(), fallback)
                self._cache[key] = text if validate_mental_map_summary(
                    text, projection, max_chars=self._max_chars) else fallback
            except Exception:
                self._cache[key] = fallback
