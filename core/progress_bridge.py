"""GameState ↔ Blackboard 橋（SK03）。

kernel 的 GameState 是推進的事實來源；bridge 把 clues/inventory/npc/scene 同步進既有 Blackboard，
讓 snapshot、get_inventory、debug、story context 都拿得到，而不必重寫整個狀態層（integration-guide §3）。
"""
from __future__ import annotations

from dataclasses import asdict

from core.progress_models import (
    GameState, InventoryItem, LedgerEntry, NPCPresence, Obligation, SceneState,
)
from core.scene_graph import SceneGraphProvider


def init_game_state(provider: SceneGraphProvider) -> GameState:
    """依 provider 開局：start_scene + 預設 obligations。"""
    start = provider.start_scene()
    return GameState(
        version=1, beat_number=0,
        current_scene=start, scene_phase="beginning",
        scenes={start: SceneState(id=start)},
        open_obligations=list(provider.default_obligations()),
    )


def to_snapshot_dict(gs: GameState) -> dict:
    """JSON-able 序列化（set → list）供 snapshot 持久化。"""
    d = asdict(gs)
    d["forbidden_repeats"] = sorted(gs.forbidden_repeats)
    return d


def from_snapshot_dict(d: dict) -> GameState:
    """從 snapshot dict 還原 GameState（回溯/讀檔用）。"""
    return GameState(
        version=d["version"], beat_number=d["beat_number"],
        current_scene=d["current_scene"], scene_phase=d["scene_phase"],
        scenes={k: SceneState(**v) for k, v in d.get("scenes", {}).items()},
        event_status=dict(d.get("event_status", {})),
        event_counts=dict(d.get("event_counts", {})),
        recent_events=list(d.get("recent_events", [])),
        forbidden_repeats=set(d.get("forbidden_repeats", [])),
        open_obligations=[Obligation(**o) for o in d.get("open_obligations", [])],
        clues={k: LedgerEntry(**v) for k, v in d.get("clues", {}).items()},
        inventory={k: InventoryItem(**v) for k, v in d.get("inventory", {}).items()},
        npcs={k: NPCPresence(**v) for k, v in d.get("npcs", {}).items()},
        danger_level=d.get("danger_level", 0),
    )


def sync_to_blackboard(gs: GameState, blackboard) -> None:
    """把 GameState 的可見面同步進 Blackboard（系統級寫入，供 snapshot/UI/story）。

    - scene_registry.current_location ← current_scene
    - shared_inventory.items ← inventory（既有 get_inventory 讀此；不洩漏 is_key_item）
    - game_meta.{clues, visible_npcs, scene_phase, danger_level, beat_number, progress}
    """
    sr = dict(blackboard.scene_registry) if isinstance(blackboard.scene_registry, dict) else {}
    sr["current_location"] = gs.current_scene
    blackboard.scene_registry = sr

    # 共用道具庫：對外只露 id/name/brief/held_by，**永不露 is_key_item**（不主動劇透哪個是關鍵道具）。
    blackboard.shared_inventory = {
        "items": [{"id": i.id, "name": i.name, "brief": i.description, "held_by": i.held_by}
                  for i in gs.inventory.values()]
    }

    blackboard.game_meta = {
        **(blackboard.game_meta if isinstance(blackboard.game_meta, dict) else {}),
        "clues": [{"id": c.id, "title": c.title, "content": c.content} for c in gs.clues.values()],
        "visible_npcs": [
            {"id": n.id, "name": n.name, "scene": n.current_scene, "entry_mode": n.entry_mode}
            for n in gs.npcs.values() if n.visible
        ],
        "scene_phase": gs.scene_phase,
        "danger_level": gs.danger_level,
        "beat_number": gs.beat_number,
        "progress_kernel": True,
    }


def debug_state(gs: GameState) -> dict:
    """get_debug_state 用：當前進度狀態檢視。"""
    return {
        "scene": gs.current_scene,
        "phase": gs.scene_phase,
        "beat_number": gs.beat_number,
        "danger_level": gs.danger_level,
        "recent_events": gs.recent_events[-5:],
        "forbidden_repeats": sorted(gs.forbidden_repeats),
        "open_obligations": [o.id for o in gs.open_obligations if o.status == "open"],
        "clues": [{"id": c.id, "title": c.title} for c in gs.clues.values()],
        "inventory": [{"id": i.id, "name": i.name} for i in gs.inventory.values()],
        "visible_npcs": [n.id for n in gs.npcs.values() if n.visible],
    }
