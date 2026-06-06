"""core.narrative.opening_pools — 開場變體池 + 權重 + medium→literals 映射（補丁 v0.8）。

把「開場素材」從具體名詞降級成**抽象類別 + 權重池**（patch docs/01 原則：用抽象欄位取代具體例子）。

- 過度重複的素材刻意降權（`missing_person` / `handwritten_note` / `missing_sibling`），
  讓 selector 自然把它們稀釋掉，而**不是**靠 prompt 提醒。
- `MEDIUM_LITERALS`：每個 message medium 對應的常見表層字串——
  既用於「被選中的 medium 用過後記進 cooldown」，也用於 gate 偵測「該 medium 不該出現的字串」。
- `first_interactable_type` 全部是**抽象方向**（terminal/door/container…），不綁任何主題名詞。

這些是**預設池**；之後可由 GUI（P7）覆寫權重，但 runtime 不依賴 GUI。
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Mapping


# ── 動機 archetype 權重（downweight：找人）────────────────────────────────────
MOTIVE_WEIGHTS: dict[str, float] = {
    "missing_person": 0.6,        # 過度重複的「找人」→ 降權
    "retrieve_object": 1.0,
    "verify_event": 1.2,
    "repair_system": 1.0,
    "escape_or_withdraw": 0.9,
    "investigate_signal": 1.2,
    "protect_someone": 0.9,
    "prove_innocence": 1.0,
    "recover_memory": 1.0,
    "deliver_or_hide": 0.8,
    "identify_entity": 0.9,
    "map_route": 0.8,
}

# ── 人物錨點權重（downweight：固定姓名式的失蹤手足；鼓勵非人物錨點）──────────────
ANCHOR_WEIGHTS: dict[str, float] = {
    "missing_sibling": 0.5,       # 「林晨」式固定姓名錨點 → 降權
    "former_colleague": 1.0,
    "unknown_sender": 1.1,
    "past_self": 1.0,
    "client": 0.9,
    "supervisor": 0.9,
    "patient_subject": 0.9,
    "recorded_voice": 1.0,
    "accused_person": 0.9,
    "no_person_anchor": 1.2,      # 鼓勵由事件/物件驅動，打破「每次都有個失蹤的人」
}

# ── 訊息載體權重（downweight：手寫紙條）+ 每 medium 的表層字串 ────────────────────
MEDIUM_WEIGHTS: dict[str, float] = {
    "handwritten_note": 0.4,      # 「紙條」→ 降權
    "voice_message": 1.0,
    "corrupted_log": 1.1,
    "access_record": 1.0,
    "radio_burst": 1.0,
    "cctv_frame": 1.0,
    "printed_receipt": 0.8,
    "device_status": 1.0,
    "body_mark": 0.8,
    "object_placement": 0.9,
    "environmental_trace": 1.0,
    "npc_claim": 0.9,
    "inventory_anomaly": 1.0,
    "schedule_entry": 0.9,
    "map_annotation": 0.9,
    "photo_artifact": 1.0,
    "terminal_prompt": 0.9,
    "emergency_broadcast": 0.9,
}

# 每個 message medium 的常見表層字串（記 cooldown / gate 偵測用）。
# handwritten_note 的字串就是 baseline 反覆出現的「紙條」家族。
MEDIUM_LITERALS: dict[str, list[str]] = {
    "handwritten_note": ["紙條", "纸条", "便條", "便条", "手寫留言", "手写留言",
                         "字條", "字条", "便籤", "便签"],
    "voice_message": ["語音留言", "錄音", "語音訊息"],
    "corrupted_log": ["損壞日誌", "錯誤紀錄", "系統日誌"],
    "access_record": ["門禁紀錄", "刷卡紀錄", "出入紀錄"],
    "radio_burst": ["無線電", "斷續訊號"],
    "cctv_frame": ["監視器畫面", "監視畫面"],
    "printed_receipt": ["熱感列印單", "收據"],
    "device_status": ["錯誤碼", "狀態燈", "儀器狀態"],
    "body_mark": ["手腕編號", "身上的標記"],
    "object_placement": ["被放在門口的東西"],
    "environmental_trace": ["濕腳印", "刮痕"],
    "npc_claim": ["有人聲稱"],
    "inventory_anomaly": ["身上多出的物件", "口袋裡多出的東西"],
    "schedule_entry": ["值班表", "行程表"],
    "map_annotation": ["地圖圈記", "地圖上的記號"],
    "photo_artifact": ["模糊照片", "反光裡多出的人"],
    "terminal_prompt": ["終端機殘留指令", "殘留指令"],
    "emergency_broadcast": ["廣播訊息", "緊急廣播"],
}

# ── 第一個可互動物方向（抽象類別，不綁主題名詞）─────────────────────────────────
INTERACTABLE_WEIGHTS: dict[str, float] = {
    "terminal_or_console": 1.0,
    "door_or_hatch": 1.0,
    "container_or_locker": 1.0,
    "control_panel": 0.9,
    "personal_device": 1.0,
    "fixed_screen_or_monitor": 0.9,
    "wall_marking_or_sign": 0.9,
    "intercom_or_speaker": 0.9,
    "window_or_aperture": 0.8,
    "discarded_object": 0.9,
    "switch_or_lever": 0.9,
    "body_or_remains": 0.7,
}

# ── 抽象 → 表層敘事提示（給 StoryAgent 的「該寫什麼類型」指引；無具體名詞）──────────
MOTIVE_GOAL_HINTS: dict[str, str] = {
    "missing_person": "確認某個與你有關的人此刻在哪裡、出了什麼事",
    "retrieve_object": "找回一件你必須帶走、卻不在該在位置的東西",
    "verify_event": "查清一件被記錄下來、卻說不通的事是否真的發生過",
    "repair_system": "讓某個失效的系統重新運作，否則情況會更糟",
    "escape_or_withdraw": "在情況惡化前，找到能安全離開這裡的路",
    "investigate_signal": "追查一個不該存在的訊號究竟從何而來",
    "protect_someone": "把某個處境危險的人帶到相對安全的地方",
    "prove_innocence": "證明某件被算在你頭上的事其實不是你做的",
    "recover_memory": "拼回一段你想不起來、卻關係重大的記憶",
    "deliver_or_hide": "把一樣東西送到該去的地方，或在被發現前把它藏好",
    "identify_entity": "弄清楚某個身分不明的存在到底是誰、是什麼",
    "map_route": "摸清這個地方的結構，替自己建立一條能依靠的路線",
}

ANCHOR_LABEL_HINTS: dict[str, str | None] = {
    "missing_sibling": "你失散的手足",
    "former_colleague": "你的前同事",
    "unknown_sender": "一個匿名的發訊者",
    "past_self": "過去的你自己",
    "client": "委託你來的人",
    "supervisor": "你的上司",
    "patient_subject": "你曾照管的受試者",
    "recorded_voice": "錄音裡的那個聲音",
    "accused_person": "被指控的那個人",
    "no_person_anchor": None,        # 沒有人物錨點：由事件/物件驅動
}


@dataclass
class VariationPools:
    """一組變體池（可被 GUI/config 覆寫；不傳則用模組預設）。"""
    motive_weights: dict[str, float] = field(default_factory=lambda: dict(MOTIVE_WEIGHTS))
    anchor_weights: dict[str, float] = field(default_factory=lambda: dict(ANCHOR_WEIGHTS))
    medium_weights: dict[str, float] = field(default_factory=lambda: dict(MEDIUM_WEIGHTS))
    interactable_weights: dict[str, float] = field(default_factory=lambda: dict(INTERACTABLE_WEIGHTS))
    medium_literals: dict[str, list[str]] = field(default_factory=lambda: {
        k: list(v) for k, v in MEDIUM_LITERALS.items()})

    def literals_for_medium(self, medium: str) -> list[str]:
        return list(self.medium_literals.get(medium, []))

    @classmethod
    def from_config(cls, data: Mapping | None) -> "VariationPools":
        """從 variation_pools.json 形態的 dict 載入（缺項回退預設）。"""
        pools = cls()
        if not data:
            return pools
        mot = data.get("motive_archetypes") or {}
        if mot:
            pools.motive_weights = {k: float((v or {}).get("weight", 1.0)) for k, v in mot.items()}
        anc = data.get("personal_anchors") or {}
        if anc:
            pools.anchor_weights = {k: float((v or {}).get("weight", 1.0)) for k, v in anc.items()}
        med = data.get("message_mediums") or {}
        if med:
            pools.medium_weights = {k: float((v or {}).get("weight", 1.0)) for k, v in med.items()}
            lits = {k: list((v or {}).get("literals", [])) for k, v in med.items() if (v or {}).get("literals")}
            if lits:
                merged = {k: list(v) for k, v in MEDIUM_LITERALS.items()}
                merged.update(lits)
                pools.medium_literals = merged
        inter = data.get("first_interactables") or {}
        if inter:
            pools.interactable_weights = {k: float((v or {}).get("weight", 1.0)) for k, v in inter.items()}
        return pools


def default_pools() -> VariationPools:
    return VariationPools()
