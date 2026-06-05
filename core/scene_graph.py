"""SceneGraphProvider（SK02）。

kernel 只依賴 provider 輸出的標準 graph，不綁死病房劇情。現用 StaticOpeningSceneGraphProvider
讀預定開場圖；GeneratedSceneGraphProvider（setup 依主題生成後續區域）留 Patch 2。
"""
from __future__ import annotations

import json
from abc import ABC, abstractmethod
from pathlib import Path

from core.graph_invariants import validate_graph_invariants
from core.progress_models import Obligation

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_GRAPH_PATH = ROOT / "data" / "opening_scene_graph.json"


class SceneGraphProvider(ABC):
    """提供 ProgressKernel 所需的標準 scene graph 與開局資訊。"""

    @abstractmethod
    def graph(self) -> dict: ...

    @abstractmethod
    def start_scene(self) -> str: ...

    def default_obligations(self) -> list[Obligation]:
        return []


class StaticOpeningSceneGraphProvider(SceneGraphProvider):
    """讀預定 data/opening_scene_graph.json（病房→走廊→抓痕→護士）。"""

    def __init__(self, path: str | Path = DEFAULT_GRAPH_PATH):
        self._path = Path(path)
        self._graph = json.loads(self._path.read_text(encoding="utf-8"))

    def graph(self) -> dict:
        return self._graph

    def start_scene(self) -> str:
        return self._graph.get("start_scene", "scene.ward_101")

    def default_obligations(self) -> list[Obligation]:
        return [
            Obligation(id="obl.leave_starting_room", kind="transition", description="Leave the starting ward."),
            Obligation(id="obl.seed_first_clue", kind="grant_clue", description="Seed the first clue."),
            Obligation(id="obl.introduce_first_npc", kind="spawn_npc", description="Introduce first NPC."),
        ]


# ── Patch 2：主題化場景圖（SK08）────────────────────────────────────────────
def _first(seq, default=None):
    for x in (seq or []):
        return x
    return default


def build_themed_graph(bb_snapshot: dict) -> dict:
    """用程式碼模板 + setup 的主題內容編譯出有效的推進圖（任何主題皆可）。

    結構固定（exit→second scene→meet npc→escalation，與開場圖同 schema），內容主題化；
    內容缺漏時 synthesize 預設，永遠回傳有效 graph。LLM 不生 raw graph（避免 schema 風險）。
    """
    sc = bb_snapshot.get("scene_registry") or {}
    locs = sc.get("known_locations") if isinstance(sc, dict) else []
    locs = [l for l in (locs or []) if isinstance(l, dict)]
    npcs = [n for n in (bb_snapshot.get("npc_registry") or []) if isinstance(n, dict)]
    rb = bb_snapshot.get("real_bible") or {}
    atmosphere = ""
    wt = rb.get("world_truth") if isinstance(rb, dict) else None
    if isinstance(wt, dict):
        atmosphere = wt.get("the_threat_is") or wt.get("what_really_happened") or ""

    # 起始場景
    start_id = (sc.get("current_location") if isinstance(sc, dict) else None) \
        or (locs[0]["id"] if locs and locs[0].get("id") else None) or "scene.start"
    start_name = next((l.get("name") for l in locs if l.get("id") == start_id), None) \
        or (locs[0].get("name") if locs else None) or "起始之地"

    # 第二、第三場景（不同於起始；缺則 synthesize）
    others = [l for l in locs if l.get("id") and l.get("id") != start_id]
    mid_id = others[0]["id"] if others else "scene.beyond"
    mid_name = (others[0].get("name") if others else None) or "更深處"
    deep_id = others[1]["id"] if len(others) > 1 else "scene.depths"
    deep_name = (others[1].get("name") if len(others) > 1 else None) or "最深處"

    npc0 = _first(npcs)
    npc_name = (npc0.get("name") if npc0 else None) or "一個身影"
    npc_id = "npc." + str(npc_name)

    atmo = (atmosphere[:50] + "…") if atmosphere else "這裡留下了不該存在的痕跡。"

    # NR0：把 kernel 探索線索綁定到 real_bible 的真相碎片（clue 自帶 truth_id + evidence_strength）。
    # 強度階梯：愈深入的線索強度愈高；clue.core 為決定性證據（足以直接 confirmed）。
    _pool = [f for f in (rb.get("revelation_pool") or [])
             if isinstance(f, dict) and f.get("id")]

    def _truth_for(slot, strength, max_level="actionable"):
        if not _pool:
            return None
        f = _pool[min(slot, len(_pool) - 1)]
        return (f["id"], strength, max_level)

    _CLUE_TRUTH = {
        "clue.first_sign": _truth_for(0, 0.30),
        "clue.start_detail": _truth_for(1, 0.45),
        "clue.descent": _truth_for(2, 0.65),
        "clue.core": (_pool[-1]["id"], 1.60, "actionable") if _pool else None,
    }

    def _clue(cid, title, content, tags):
        value = {"title": title, "content": content, "tags": tags}
        t = _CLUE_TRUTH.get(cid)
        if t:
            value["truth_id"], value["evidence_strength"], value["max_level"] = t
        return {"op": "add", "path": f"clues.{cid}", "value": value}

    def _exit(to_id):
        return [{"op": "set", "path": "current_scene", "value": to_id},
                {"op": "set", "path": "scene_phase", "value": "beginning"}]

    events = [
        # ── 起始場景：≥2 出口（exit_start/exit_alt）+ 搜查（多解法/多 intent）──
        {"id": "event.exit_start", "scene_id": start_id,
         "intent_tags": ["open", "move", "inspect"],
         "preconditions": ["not_recent:event.exit_start"],
         "effects": _exit(mid_id) + [_clue("clue.first_sign", "第一個異樣", atmo, ["opening"])],
         "progress_delta": ["event_resolved", "location_changed", "new_clue_added"],
         "narrative_obligations": [
             f"必須描寫玩家離開「{start_name}」、進入「{mid_name}」。",
             "必須自然描寫一個新線索。下一個選項不得再問是否離開同一出口。"],
         "forbidden_after": ["event.exit_start", "ask_exit_start"],
         "grants_clues": ["clue.first_sign"],
         "satisfies": ["obl.leave_starting_room", "obl.seed_first_clue"], "max_repeat": 1},

        {"id": "event.exit_alt", "scene_id": start_id,
         "intent_tags": ["move", "inspect", "free"],
         "preconditions": ["not_recent:event.exit_alt"],
         "effects": _exit(mid_id) + [{"op": "inc", "path": "danger_level", "value": 1}],
         "progress_delta": ["event_resolved", "location_changed", "danger_level_changed"],
         "narrative_obligations": [
             f"玩家以另一條較險的路徑離開「{start_name}」進入「{mid_name}」，並付出代價（壓力升高）。"],
         "forbidden_after": ["event.exit_alt", "ask_exit_alt"],
         "satisfies": ["obl.leave_starting_room"], "max_repeat": 1},

        {"id": "event.search_start", "scene_id": start_id,
         "intent_tags": ["inspect", "search", "free"],
         "preconditions": ["clue_missing:clue.start_detail"],
         "effects": [_clue("clue.start_detail", "被忽略的細節",
                           "在「" + start_name + "」更仔細看，才注意到的東西。", ["detail"])],
         "progress_delta": ["new_clue_added"],
         "narrative_obligations": ["玩家在原地找到一條新線索，而非重複描述同一場景。"],
         "forbidden_after": [], "grants_clues": ["clue.start_detail"], "max_repeat": 1},

        # ── 第二場景：NPC + 下潛出口 + 升壓（≥2 事件、≥1 出口）──
        {"id": "event.meet_npc", "scene_id": mid_id,
         "intent_tags": ["inspect", "move", "free", "call"],
         "preconditions": [f"npc_not_visible:{npc_id}"],
         "effects": [{"op": "set", "path": f"npcs.{npc_id}.visible", "value": True},
                     {"op": "set", "path": f"npcs.{npc_id}.current_scene", "value": mid_id},
                     {"op": "set", "path": f"npcs.{npc_id}.entry_mode", "value": "seen_at_distance"}],
         "progress_delta": ["npc_spawned"],
         "narrative_obligations": [
             f"必須讓「{npc_name}」以遠距離、曖昧、不完全可信的方式出現在「{mid_name}」，"
             "並給玩家新的互動方向（詢問/跟隨/躲避/觀察）。"],
         "forbidden_after": ["event.meet_npc"], "spawns_npcs": [npc_id],
         "satisfies": ["obl.introduce_first_npc"], "max_repeat": 1},

        {"id": "event.descend", "scene_id": mid_id,
         "intent_tags": ["move", "open"],
         "preconditions": ["not_recent:event.descend"],
         "effects": _exit(deep_id) + [_clue("clue.descent", "向下的痕跡",
                                           "通往「" + deep_name + "」的路上留下的東西。", ["depth"])],
         "progress_delta": ["event_resolved", "location_changed", "new_clue_added"],
         "narrative_obligations": [f"玩家深入到「{deep_name}」，環境更壓迫。"],
         "forbidden_after": ["event.descend", "ask_descend"],
         "grants_clues": ["clue.descent"], "max_repeat": 1},

        {"id": "event.mid_escalate", "scene_id": mid_id,
         "intent_tags": ["free", "move", "inspect"],
         "preconditions": ["not_recent:event.mid_escalate"],
         "effects": [{"op": "inc", "path": "danger_level", "value": 1}],
         "progress_delta": ["danger_level_changed"],
         "narrative_obligations": [f"「{mid_name}」遠處壓力升級（聲響/逼近/異變），不得原地重複。"],
         "forbidden_after": [], "max_repeat": 2},

        # ── 最深處：核心線索（真相）+ 退回出口 + 升壓（≥2 事件、≥1 出口）──
        {"id": "event.deep_find", "scene_id": deep_id,
         "intent_tags": ["inspect", "search", "free"],
         "preconditions": ["clue_missing:clue.core"],
         "effects": [_clue("clue.core", "核心真相的碎片",
                           "在「" + deep_name + "」深處，最不該看到的東西。", ["truth"])],
         "progress_delta": ["new_clue_added", "truth_fragment_revealed"],
         "narrative_obligations": ["玩家逼近核心真相，但只給碎片，不是全貌。"],
         "forbidden_after": [], "grants_clues": ["clue.core"], "max_repeat": 1},

        {"id": "event.deep_retreat", "scene_id": deep_id,
         "intent_tags": ["move", "free"],
         "preconditions": ["not_recent:event.deep_retreat"],
         "effects": _exit(mid_id),
         "progress_delta": ["event_resolved", "location_changed"],
         "narrative_obligations": [f"玩家退回「{mid_name}」，但帶著新的不安。"],
         "forbidden_after": [], "max_repeat": 9},

        {"id": "event.deep_escalate", "scene_id": deep_id,
         "intent_tags": ["free", "inspect"],
         "preconditions": ["not_recent:event.deep_escalate"],
         "effects": [{"op": "inc", "path": "danger_level", "value": 2}],
         "progress_delta": ["danger_level_changed"],
         "narrative_obligations": ["危險主動逼近，壓力顯著升高。"],
         "forbidden_after": [], "max_repeat": 4},
    ]
    graph = {"version": 1, "description": f"themed graph for {start_name}",
             "start_scene": start_id, "events": events}
    validate_graph_invariants(graph, start_scene=start_id)   # 強制多出口/多解法/多 intent
    return graph


class GeneratedSceneGraphProvider(SceneGraphProvider):
    """從 setup 輸出（blackboard snapshot）編譯主題化推進圖。內容主題化、結構模板化。"""

    def __init__(self, bb_snapshot: dict):
        self._graph = build_themed_graph(bb_snapshot)

    def graph(self) -> dict:
        return self._graph

    def start_scene(self) -> str:
        return self._graph.get("start_scene", "scene.start")

    def default_obligations(self) -> list[Obligation]:
        return [
            Obligation(id="obl.leave_starting_room", kind="transition", description="Leave the starting scene."),
            Obligation(id="obl.seed_first_clue", kind="grant_clue", description="Seed the first clue."),
            Obligation(id="obl.introduce_first_npc", kind="spawn_npc", description="Introduce first NPC."),
        ]
