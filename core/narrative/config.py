"""core.narrative.config — 敘事控制參數可配置（NC7，optional）。

把 opening_budget / reveal_policy / forbidden_motifs / element_limit / ending_thresholds 接到配置中心，
**safe profile 為預設**：未設時一律走安全值（與各模組硬預設一致）。數值型由 config 中心 feature_flags 讀，
列表/策略型由 overrides（UI/JSON）提供。
"""
from __future__ import annotations

from typing import Any, Optional

# 安全預設（與 NC1–NC6 各模組硬預設一致）
SAFE_DEFAULTS: dict = {
    "opening_max_named_objects": 3,
    "opening_max_new_lore_terms": 3,
    "opening_max_chars": 900,
    "opening_reveal_limit": "hinted",      # reveal policy
    "element_limit": 1,                    # story_agent_element_limit
    "forbidden_motifs": ["血字警告", "塗掉臉的照片", "憑空消失的腳印", "夢裡的聲音", "菌絲"],
    "ending_clean_requires_truth": True,   # ending_gate_thresholds：0/8 不可 clean
}

# 數值型參數 → config 中心 feature_flags 的 key（可由 UI/JSON 調整）
_FLAG_KEYS = {
    "opening_max_named_objects": "NC_OPENING_MAX_OBJECTS",
    "opening_max_new_lore_terms": "NC_OPENING_MAX_LORE",
    "opening_max_chars": "NC_OPENING_MAX_CHARS",
    "element_limit": "NC_ELEMENT_LIMIT",
}


def get_narrative_config(store: Any = None, profile: Optional[str] = None,
                         overrides: Optional[dict] = None) -> dict:
    """回傳敘事控制設定：safe 預設 ← config 中心數值覆寫 ← overrides（UI/JSON）。"""
    cfg = dict(SAFE_DEFAULTS)
    if store is not None:
        for key, flag in _FLAG_KEYS.items():
            try:
                v = store.get_flag(flag, profile)
            except Exception:
                v = None
            if v is not None:
                cfg[key] = int(v)
    if overrides:
        cfg.update({k: v for k, v in overrides.items() if k in SAFE_DEFAULTS})
    return cfg


def apply_to_budget(budget: Any, cfg: dict) -> None:
    """把 config 套到 OpeningBudget（就地）。"""
    budget.max_named_objects = cfg.get("opening_max_named_objects", budget.max_named_objects)
    budget.max_new_lore_terms = cfg.get("opening_max_new_lore_terms", budget.max_new_lore_terms)
    budget.max_opening_chars = cfg.get("opening_max_chars", budget.max_opening_chars)
