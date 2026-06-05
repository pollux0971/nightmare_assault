"""core.config.flags — feature flag 解析（P4）。

優先序（固定，contract）：`.env override > active profile DB config > default profile DB config > hardcoded default`。
契約：dev/CONTRACTS.md §十一 `FeatureFlags`；canonical = patch docs/04、docs/08。
"""
from __future__ import annotations

import os
from typing import Any, Optional

# hardcoded 預設（最低優先；與 config/CONTRACT_FREEZE.md §四一致）
HARDCODED_DEFAULTS: dict[str, int] = {
    "ENABLE_CONFIG_CENTER": 0,
    "ENABLE_PROMPT_PREVIEW": 1,
    "ENABLE_RUN_CONFIG_SNAPSHOT": 1,
}

_FALSEY = ("false", "0", "no", "off", "")


def _truthy(v: Any) -> bool:
    return str(v).strip().lower() not in _FALSEY


def resolve_flag(name: str, store: Any = None, profile: Optional[str] = None) -> bool:
    """依固定優先序解析 flag。store 可為 None（無 DB 時退 env/hardcoded）。"""
    env = os.environ.get(name)
    if env is not None:
        return _truthy(env)
    if store is not None:
        try:
            dbv = store.get_flag(name, profile)   # 內部已 active→default profile 回退
            if dbv is not None:
                return bool(int(dbv))
        except Exception:
            pass
    return bool(HARDCODED_DEFAULTS.get(name, 0))


def config_center_enabled(store: Any = None, profile: Optional[str] = None) -> bool:
    return resolve_flag("ENABLE_CONFIG_CENTER", store, profile)


def run_config_snapshot_enabled(store: Any = None, profile: Optional[str] = None) -> bool:
    return resolve_flag("ENABLE_RUN_CONFIG_SNAPSHOT", store, profile)
