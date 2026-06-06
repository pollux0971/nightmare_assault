"""core.world.extractor — EntityExtractor（**fallback**）。

優先路徑:story / npc 直接輸出結構化 `entity_delta`(WorldModel.apply_deltas)。
本抽取器是 LLM 不配合(吐自由文字)時的退路,且**選擇性登記**——不把所有名詞都變實體,
只登記:可互動物件 / 可檢查事實 / NPC。地圖區域與出口由 kernel 圖負責,這裡不抽。

零 LLM、規則版。對應 15-player-sovereignty.md §二。
"""
from __future__ import annotations

import re

from core.world.model import WorldDelta, OBJECT, FACT, ACTOR, INSPECT, TAKE

# 通用「小型可互動物件」類別(主題無關;不是世界觀內容,是物件型別)
_OBJ_CATS = ["袖扣", "筆記本", "筆記", "照片", "鑰匙", "紙條", "便條", "文件", "員工證",
             "識別證", "錄音帶", "錄音", "卡片", "日誌", "手機", "磁帶", "光碟", "信封",
             "瓶子", "刀", "鎖", "膠帶", "地圖", "對講機", "工具"]
# 「敘事前景化、可拿取/可檢查」的線索詞(要有其一,才登記,避免過度登記)
_CUES = ["撿起", "拿起", "握著", "桌上", "地上", "抽屜", "看到一", "找到", "掉落", "遺留",
         "發現", "櫃裡", "口袋", "撿到", "翻出", "攤開", "夾著"]
# 可拿取類別(給 take affordance)
_TAKEABLE = {"袖扣", "鑰匙", "紙條", "便條", "卡片", "員工證", "識別證", "手機", "錄音帶",
             "磁帶", "光碟", "刀", "對講機"}


def extract_entities(narrative: str, *, npc_names: list | None = None) -> list[WorldDelta]:
    """從自由文字選擇性抽出 entity 登記 delta(objects / facts / actors)。"""
    t = narrative or ""
    out: list[WorldDelta] = []
    seen: set = set()

    # ── 物件:類別詞 + 前景化線索 才登記 ─────────────────────────────────────
    has_cue = any(c in t for c in _CUES)
    if has_cue:
        for cat in _OBJ_CATS:
            if cat not in t or cat in seen:
                continue
            seen.add(cat)
            # 試抓前置代號(如「WU 袖扣」),否則用類別詞當 label
            m = re.search(r"([A-Za-z0-9]{1,6})\s*的?\s*" + re.escape(cat), t)
            label = f"{m.group(1)} {cat}" if m else cat
            affords = [INSPECT, TAKE] if cat in _TAKEABLE else [INSPECT]
            out.append(WorldDelta(op="register", kind=OBJECT, label=label,
                                  affords=affords, origin="extractor"))

    # ── 可檢查事實:沿用 world_facts(扁平 → 收進實體模型)─────────────────────
    try:
        from core.narrative.world_facts import extract_world_facts
        for f in extract_world_facts(t, source="extractor", npc_names=npc_names):
            out.append(WorldDelta(op="register", kind=FACT, label=f.get("label", f["key"]),
                                  entity_id=f"fact.{f['key']}", props={"category": f.get("category")},
                                  origin="extractor"))
    except Exception:
        pass

    # ── NPC:已知名單中出現於文字 → 登記 actor(present)──────────────────────
    for name in (npc_names or []):
        if name and name in t and f"actor.{name}" not in seen:
            seen.add(f"actor.{name}")
            out.append(WorldDelta(op="register", kind=ACTOR, label=name,
                                  entity_id=f"actor.{name}", origin="extractor"))
    return out
