"""core.narrative.world_facts — 可檢查的世界狀態事實（WorldStateFact，Player Sovereignty P0）。

NPC / story 給出的**有用資訊**不一定進 truth reveal，但必須能寫入可檢查的 world_state_fact
（如 known_exit_locked / generator_needed / machine_room_known / <npc>_confirmed_present）。
這樣即使沒有 truth reveal，玩家的行動仍留下「可檢查的世界後果」。

零 LLM、規則版。對應 docs/player-sovereignty-principles.md + P0。
"""
from __future__ import annotations

# (主題關鍵詞, 觸發詞, fact_key, category)；category ∈ exit|mechanism|location|hazard|presence
_FACT_PATTERNS = [
    (["出口", "大門", "防爆門", "閘門"], ["鎖死", "鎖住", "打不開", "封住", "上鎖", "鎖了"],
     "known_exit_locked", "exit"),
    (["發電機", "電力", "供電", "電源"], ["重啟", "啟動", "修復", "需要", "沒電", "斷電", "恢復"],
     "generator_needed", "mechanism"),
    (["機房", "控制室", "通訊設備", "無線電", "通訊"], ["在", "位於", "通往", "去", "往"],
     "machine_room_known", "location"),
    (["鑰匙", "門禁卡", "密碼", "通行證"], ["需要", "要", "缺", "找", "拿到"],
     "access_needed", "mechanism"),
    (["北門", "後門", "側門"], ["別", "不要", "危險", "千萬", "避開"],
     "side_door_hazard", "hazard"),
]


def extract_world_facts(text: str, *, source: str = "story",
                        npc_names: list | None = None) -> list[dict]:
    """從文字抽出 world_state_fact。回傳 [{key, label, category, source}]（去重）。

    npc_names：已知 NPC 名單；偵測「<名字> 來過/在場/出現/見過」→ <name>_confirmed_present。
    """
    t = text or ""
    out: list[dict] = []

    def _add(key, label, category):
        if not any(f["key"] == key for f in out):
            out.append({"key": key, "label": label, "category": category, "source": source})

    for topics, triggers, key, category in _FACT_PATTERNS:
        if any(tp in t for tp in topics) and any(tg in t for tg in triggers):
            topic = next(tp for tp in topics if tp in t)
            _add(key, f"{topic}相關：{key}", category)

    _PRESENCE = ["來過", "在場", "出現過", "見過", "待過", "去過", "經過這"]
    for name in (npc_names or []):
        if name and name in t and any(v in t for v in _PRESENCE):
            _add(f"{name}_confirmed_present", f"{name} 確認曾在此出現", "presence")
    return out


def add_world_facts(blackboard, facts: list[dict], *, beat: int | None = None) -> list[str]:
    """把新事實寫進 game_meta['world_facts']（已存在則不重複）。回傳本次**新增**的 key。"""
    try:
        store = dict((getattr(blackboard, "game_meta", {}) or {}).get("world_facts") or {})
        added: list[str] = []
        for f in facts or []:
            k = f.get("key")
            if k and k not in store:
                store[k] = {"label": f.get("label", k), "category": f.get("category", "general"),
                            "source": f.get("source", "story"), "beat": beat}
                added.append(k)
        if added:
            blackboard.game_meta = {**blackboard.game_meta, "world_facts": store}
        return added
    except Exception:
        return []


def get_world_facts(blackboard) -> dict:
    return dict((getattr(blackboard, "game_meta", {}) or {}).get("world_facts") or {})
