"""Blackboard — 中央狀態容器（U03）。

設計原則：
- 所有 agent 只透過 Blackboard 讀寫，不直接操作各自本地狀態。
- 錨點（real_bible / secret_core）由權限表強制，違規寫入拋 PermissionError。
- 非同步 agent 只提交 patch（submit_patch），安全點（merge_and_bump）才實際生效。
- snapshot() 回傳深拷貝，pending patch 完全不可見（A5 隔離）。
"""
from __future__ import annotations

import copy
from typing import Any


# ─────────────────────────────────────────────────────────────────────────────
# 權限表
# ─────────────────────────────────────────────────────────────────────────────

# 錨點欄位：任何觸及這些路徑的寫入，除 setup 以外全部禁止
_ANCHOR_ROOTS: frozenset[str] = frozenset({"real_bible"})
_ANCHOR_SUBKEY: str = "secret_core"  # 任何路徑中帶有此 subkey 也算錨點


def _is_anchor(target: str) -> bool:
    """判斷 target 路徑是否為錨點（real_bible 根 或 任何 secret_core 子路徑）。"""
    parts = target.split(".")
    root = parts[0]
    if root in _ANCHOR_ROOTS:
        return True
    # 任何路徑節點含 secret_core
    if _ANCHOR_SUBKEY in parts:
        return True
    return False


def _touches_npc_evolving(target: str) -> bool:
    """判斷 target 是否觸及 npc_registry.<name>.evolving。"""
    parts = target.split(".")
    # 格式：npc_registry.<name>.evolving[.*]
    if len(parts) >= 3 and parts[0] == "npc_registry" and parts[2] == "evolving":
        return True
    return False


# writer → 允許的根欄位集合（None 表示全部允許）
# 以及是否禁止寫 npc_evolving
_WRITER_POLICY: dict[str, dict[str, Any]] = {
    "setup": {
        # 初始化：可寫全部
        "allowed_roots": None,  # None = 無限制
        "deny_npc_evolving": False,
        "deny_anchor": False,
    },
    "orchestrator": {
        "allowed_roots": {"revealed_bible", "turn_context"},
        "deny_npc_evolving": False,
        "deny_anchor": True,
    },
    "story": {
        "allowed_roots": {"beat_window", "turn_context"},
        "deny_npc_evolving": True,
        "deny_anchor": True,
    },
    "warden": {
        "allowed_roots": {"turn_context", "ledger"},
        "deny_npc_evolving": True,
        "deny_anchor": True,
    },
    "npc_chat": {
        "allowed_roots": {"chat_log"},
        "deny_npc_evolving": False,
        "deny_anchor": True,
    },
    "dreaming": {
        # 可寫 npc_registry.*.evolving / npc_registry.*.offstage_intent
        "allowed_roots": {"npc_registry"},
        "deny_npc_evolving": False,
        "deny_anchor": True,
    },
    "offstage_fate": {
        # npc_registry.*.presence/alignment/carried_fragment/offstage_intent, scene_registry
        "allowed_roots": {"npc_registry", "scene_registry"},
        "deny_npc_evolving": False,
        "deny_anchor": True,
    },
    "compactor": {
        "allowed_roots": {"rolling_summary", "ledger", "recent_chat_digest"},
        "deny_npc_evolving": False,
        "deny_anchor": True,
    },
}


def can_write(writer: str, target: str) -> bool:
    """判定 writer 是否允許對 target 路徑寫入。

    Returns True 表示允許，False 表示禁止。
    未知 writer 一律禁止。
    """
    policy = _WRITER_POLICY.get(writer)
    if policy is None:
        return False

    # 1. 錨點檢查：setup 豁免，其餘一律禁止
    if _is_anchor(target):
        if policy["deny_anchor"]:
            return False
        # setup 不 deny_anchor，視為允許
        return True

    # 2. npc_evolving 檢查
    if policy["deny_npc_evolving"] and _touches_npc_evolving(target):
        return False

    # 3. 根欄位允許清單（None = 全部允許）
    allowed_roots = policy["allowed_roots"]
    if allowed_roots is None:
        return True

    root = target.split(".")[0]
    return root in allowed_roots


# ─────────────────────────────────────────────────────────────────────────────
# Blackboard
# ─────────────────────────────────────────────────────────────────────────────

# Blackboard 持有的所有 top-level 欄位名
_ALL_FIELDS: frozenset[str] = frozenset({
    "real_bible",
    "revealed_bible",
    "npc_registry",
    "protagonist",
    "shared_inventory",
    "rolling_summary",
    "ledger",
    "recent_chat_digest",
    "beat_window",
    "turn_context",
    "scene_registry",
    "game_meta",
    "version",
    "beat_number",
    # npc_chat 寫入的欄位（chat_log 非 top-level state，但允許掛在 blackboard 上）
    "chat_log",
})


class Blackboard:
    """中央狀態容器。

    所有 agent 只透過 write() / submit_patch() 寫入，snapshot() 讀取。
    非同步 patch 在 merge_and_bump() 安全點才生效。
    """

    def __init__(self) -> None:
        # ── 狀態欄位（初始空值）──────────────────────────────────────────
        self.real_bible: dict = {}
        self.revealed_bible: dict = {}
        self.npc_registry: list = []
        self.protagonist: dict = {}
        self.shared_inventory: dict = {}
        self.rolling_summary: str = ""
        self.ledger: list = []
        self.recent_chat_digest: str | None = None
        self.beat_window: list = []
        self.turn_context: dict = {}
        self.scene_registry: dict | None = None
        self.game_meta: dict = {}
        self.chat_log: list = []

        # ── 版本控制 ──────────────────────────────────────────────────────
        self.version: int = 0
        self.beat_number: int = 0

        # ── Pending patch 佇列（非同步 agent 暫存）────────────────────────
        self._pending: list[dict] = []

    # ── 同步寫入 ──────────────────────────────────────────────────────────

    def write(self, writer: str, target: str, value: Any) -> None:
        """同步寫入（setup / orchestrator / story / warden 在自己回合呼叫）。

        先通過 can_write 檢查；不通過拋 PermissionError。
        target 為 top-level 欄位名（亦可為點分路徑，但此處只處理根欄位直接賦值）。
        若 beat_window 且 writer==story → append 語意（value 應為單一 beat dict/str）。
        """
        if not can_write(writer, target):
            raise PermissionError(
                f"Writer '{writer}' is not allowed to write target '{target}'"
            )
        self._apply(target, value, writer)

    def _apply(self, target: str, value: Any, writer: str | None = None) -> None:
        """實際將 value 寫入對應欄位。

        支援：
        - top-level 欄位直接賦值
        - beat_window：writer==story 時 append（其他 writer 直接賦值）
        - npc_registry 子路徑：npc_registry.<name>.<attr> 點分路徑賦值
        - turn_context 子路徑：turn_context.<key> 點分路徑
        """
        parts = target.split(".")

        if len(parts) == 1:
            root = parts[0]
            # beat_window 的 story append 語意
            if root == "beat_window" and writer == "story":
                self.beat_window.append(value)
            else:
                setattr(self, root, value)
            return

        # 點分路徑：目前支援 npc_registry.<name>.<attr> 與 turn_context.<key>
        root = parts[0]

        if root == "turn_context":
            # turn_context.<key>
            key = ".".join(parts[1:])
            self.turn_context[key] = value
            return

        if root == "npc_registry" and len(parts) >= 3:
            # npc_registry.<name>.<attr>[.*]
            npc_name = parts[1]
            attr_path = parts[2:]
            # 找到對應 NPC
            target_npc = None
            for npc in self.npc_registry:
                npc_name_val = npc.get("name") if isinstance(npc, dict) else getattr(npc, "name", None)
                if npc_name_val == npc_name:
                    target_npc = npc
                    break
            if target_npc is None:
                raise KeyError(f"NPC '{npc_name}' not found in npc_registry")
            # 賦值到 NPC 的 attr_path
            _set_nested(target_npc, attr_path, value)
            return

        # 其他點分路徑：根欄位 dict 逐層賦值
        obj = getattr(self, root)
        _set_nested(obj, parts[1:], value)


    # ── Patch 相關 ────────────────────────────────────────────────────────

    def submit_patch(self, patch: dict) -> None:
        """非同步 agent 提交 patch，不立即生效。

        patch schema: {base_version: int, writer: str, target: str, value: Any}
        """
        required_keys = {"base_version", "writer", "target", "value"}
        missing = required_keys - patch.keys()
        if missing:
            raise ValueError(f"patch 缺少欄位：{missing}")
        self._pending.append(copy.deepcopy(patch))

    def collect_pending(self) -> list:
        """回傳目前 pending patch 的快照（不清空）。"""
        return list(self._pending)

    def snapshot(self) -> dict:
        """回傳目前狀態的穩定深拷貝，pending patch 完全不可見（A5 隔離）。"""
        return {
            "real_bible": copy.deepcopy(self.real_bible),
            "revealed_bible": copy.deepcopy(self.revealed_bible),
            "npc_registry": copy.deepcopy(self.npc_registry),
            "protagonist": copy.deepcopy(self.protagonist),
            "shared_inventory": copy.deepcopy(self.shared_inventory),
            "rolling_summary": self.rolling_summary,
            "ledger": copy.deepcopy(self.ledger),
            "recent_chat_digest": self.recent_chat_digest,
            "beat_window": copy.deepcopy(self.beat_window),
            "turn_context": copy.deepcopy(self.turn_context),
            "scene_registry": copy.deepcopy(self.scene_registry),
            "game_meta": copy.deepcopy(self.game_meta),
            "chat_log": copy.deepcopy(self.chat_log),
            "version": self.version,
            "beat_number": self.beat_number,
        }

    def merge_and_bump(self) -> dict:
        """安全點：套用所有 base_version == self.version 的 pending patch。

        - 版本匹配：套用（過 can_write，違規拋 PermissionError）
        - 版本不匹配：丟棄（視為過期）
        - 完成後 self.version += 1，清空 pending。
        - 回傳摘要 dict：{applied: int, discarded: int, version_after: int}
        """
        applied = 0
        discarded = 0
        errors = []

        for patch in list(self._pending):
            base_version = patch["base_version"]
            writer = patch["writer"]
            target = patch["target"]
            value = patch["value"]

            if base_version != self.version:
                # 過期 patch → 丟棄
                discarded += 1
                continue

            # 版本匹配 → 嘗試套用
            if not can_write(writer, target):
                # 違規寫入：拋出 PermissionError（patch 不套用）
                self._pending = []
                raise PermissionError(
                    f"Patch writer '{writer}' is not allowed to write target '{target}'"
                )

            try:
                self._apply(target, value, writer)
                applied += 1
            except Exception as e:
                errors.append({"patch": patch, "error": str(e)})
                discarded += 1

        self._pending = []
        self.version += 1

        result: dict = {
            "applied": applied,
            "discarded": discarded,
            "version_after": self.version,
        }
        if errors:
            result["errors"] = errors
        return result


# ─────────────────────────────────────────────────────────────────────────────
# 輔助函式
# ─────────────────────────────────────────────────────────────────────────────

def _set_nested(obj: Any, path: list[str], value: Any) -> None:
    """在 dict 或物件上按 path 逐層設值（最後一層賦值）。"""
    for key in path[:-1]:
        if isinstance(obj, dict):
            if key not in obj:
                obj[key] = {}
            obj = obj[key]
        else:
            obj = getattr(obj, key)
    last_key = path[-1]
    if isinstance(obj, dict):
        obj[last_key] = value
    else:
        setattr(obj, last_key, value)
