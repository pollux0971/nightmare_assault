"""core.narrative.contract — 從 setup 輸出組裝 NarrativeContract（NC1）。

setup 不直接寫長開場；本模組從 real_bible + protagonist 程式碼組裝**敘事生成契約**，
供 OpeningDirector / Story Agent 讀取（取代「直接讀完整世界觀」）。

防暴雷：契約裡 `core_premise` 保有內部前提（可能含真相），但**只有 OpeningDirector 篩出的 surface 元素**
會流到 story（見 NC2）；story 仍永不直接讀 real_bible。
"""
from __future__ import annotations

from typing import Any

from core.narrative.models import (
    NarrativeContract, ProtagonistMotive, MotifPalette, OpeningBudget,
)

# 開場避免過度堆疊的母題（forbidden_or_limited 預設）；可由 config 覆寫（NC7）。
_DEFAULT_LIMITED_MOTIFS = ["血字警告", "塗掉臉的照片", "憑空消失的腳印", "夢裡的聲音", "菌絲"]


def build_narrative_contract(blackboard: Any, config: dict | None = None) -> NarrativeContract:
    """從 blackboard 的 real_bible/protagonist 組裝 NarrativeContract。

    config（NC7，optional）：可覆寫 opening budget / forbidden_motifs / reveal_limit（safe 預設）。
    """
    snap = blackboard.snapshot() if hasattr(blackboard, "snapshot") else {}
    real = snap.get("real_bible") or {}
    world = real.get("world_truth") or {}
    proto = snap.get("protagonist") or {}

    name = proto.get("name") or "你"
    situation = (proto.get("starting_situation") or "").strip()

    # 動機：私人失去 + 此刻目標 + 情感賭注 + 第一個證據物件（開場一定保留）
    personal_loss = situation or "你牽掛的人在這裡消失了"
    motive = ProtagonistMotive(
        personal_loss=personal_loss,
        immediate_goal="找到他／找出這裡發生了什麼，並活著離開",
        emotional_stake="你不能再失去這個人",
        first_proof=f"{name} 帶在身上、與失蹤者直接相關的物件（如證件/紙條/遺物）",
    )

    # 核心問題（非劇透鉤子）：盡量用威脅/前提包裝成一個懸念
    threat = world.get("the_threat_is") or ""
    central_question = "你看見的，是現實，還是被什麼重寫過的記憶？" if threat else "這裡到底發生過什麼？"

    cfg = config or {}
    palette = MotifPalette(
        primary=list((real.get("atmosphere") or [])[:3]) or ["潮濕", "鏽蝕", "低頻嗡鳴"],
        secondary=[],
        forbidden_or_limited=list(cfg.get("forbidden_motifs") or _DEFAULT_LIMITED_MOTIFS),
    )
    budget = OpeningBudget()                 # 預設：主要新元素 ≤3、新 lore ≤3、≤900 字
    if cfg:
        from core.narrative.config import apply_to_budget
        apply_to_budget(budget, cfg)

    return NarrativeContract(
        core_premise=world.get("what_really_happened", "") or "（前提未設定）",
        protagonist_motive=motive,
        central_question=central_question,
        motif_palette=palette,
        opening_budget=budget,
        opening_reveal_limit=cfg.get("opening_reveal_limit", "hinted"),   # 開場最多 hinted
    )


def store_contract(blackboard: Any, contract: NarrativeContract) -> None:
    """把契約存進 blackboard.game_meta（optional 欄位，不破壞既有存檔）。"""
    from dataclasses import asdict
    meta = dict(getattr(blackboard, "game_meta", {}) or {})
    meta["narrative_contract"] = asdict(contract)
    blackboard.game_meta = meta
