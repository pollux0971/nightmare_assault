"""UB5 — 道具庫完整 驗收測試。

驗收：item 增刪查、held_by 轉移、不洩漏 is_key_item、隨快照保存。
"""
from __future__ import annotations

import json
from dataclasses import asdict

from core.progress_models import GameState, EventPatch, PatchOp, InventoryItem
from core.patch_validator import PatchValidator
from core import progress_bridge as bridge
from core.persistence.db import Database


def _state():
    return GameState(version=1, beat_number=2, current_scene="ward", scene_phase="承")


def _apply(st, ops, delta=("scene_changed",)):
    """以當前 state.version 當 base_version 套一個 patch（版本化並行控制）。"""
    p = EventPatch(base_version=st.version, event_id="e", ops=list(ops), progress_delta=list(delta))
    return PatchValidator().apply(p, st)


def _op(op, path, value=None):
    return PatchOp(op=op, path=path, value=value)


# ── 增 / 查 ───────────────────────────────────────────────────────────────────
def test_item_add_and_fields():
    st = _apply(_state(), [_op("add", "inventory.rusty_key",
                {"name": "生鏽鑰匙", "description": "沾著暗紅", "is_key_item": True})])
    item = st.inventory["rusty_key"]
    assert item.name == "生鏽鑰匙" and item.is_key_item is True
    assert item.held_by is None and item.source_event == "e"


# ── held_by 轉移 ─────────────────────────────────────────────────────────────
def test_held_by_transfer():
    st = _apply(_state(), [_op("add", "inventory.amulet", {"name": "護身符"})])
    st = _apply(st, [_op("set", "inventory.amulet.held_by", "npc.night_nurse")])
    assert st.inventory["amulet"].held_by == "npc.night_nurse"


# ── 刪 ───────────────────────────────────────────────────────────────────────
def test_item_remove():
    st = _apply(_state(), [_op("add", "inventory.candle", {"name": "蠟燭"})])
    assert "candle" in st.inventory
    st = _apply(st, [_op("remove", "inventory.candle")])
    assert "candle" not in st.inventory


# ── 不洩漏 is_key_item（共用道具庫對外面）────────────────────────────────────
def test_shared_inventory_hides_is_key_item():
    st = _apply(_state(), [_op("add", "inventory.diary",
                {"name": "日記", "is_key_item": True, "held_by": "protagonist"})])

    class BB:
        scene_registry = {}; game_meta = {}; shared_inventory = {}
    bb = BB()
    bridge.sync_to_blackboard(st, bb)
    items = bb.shared_inventory["items"]
    assert items[0]["name"] == "日記" and items[0]["held_by"] == "protagonist"
    assert "is_key_item" not in items[0]                 # ★ 永不外露
    assert "is_key_item" not in json.dumps(bb.shared_inventory, ensure_ascii=False)


# ── 隨快照保存（含 is_key_item 內部欄位 round-trip）─────────────────────────
def test_inventory_persists_via_snapshot_dict():
    st = _apply(_state(), [_op("add", "inventory.key",
                {"name": "鑰匙", "is_key_item": True, "held_by": "npc.doc"})])
    st2 = bridge.from_snapshot_dict(bridge.to_snapshot_dict(st))
    it = st2.inventory["key"]
    assert it.is_key_item is True and it.held_by == "npc.doc"


def test_inventory_snapshot_table_roundtrip():
    db = Database(":memory:")
    items = [asdict(InventoryItem(id="k", name="鑰匙", description="", source_event="e",
                                  is_key_item=True, held_by="npc.x"))]
    db.save_inventory_snapshot("run1", 5, json.dumps(items, ensure_ascii=False))
    got = db.load_inventory_snapshot("run1", 5)
    assert got and got[0]["name"] == "鑰匙" and got[0]["is_key_item"] is True
    db.close()
