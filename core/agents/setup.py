"""core.agents.setup — Setup Agent（U09）。

職責：
1. roll_personality_axes()：免 LLM token 的個性程式碼擲骰，從固定池各抽一軸。
2. run_setup()：組合擲骰結果 + opts → 呼 SkillCaller → 寫 Blackboard → 回傳 opening_sequence。

B8 原則：setup 唯一不可降級——caller 拋例外時直接往上拋，不吞掉給空世界。
"""
from __future__ import annotations

import random
from typing import Any

from core.agents.base import SkillCaller
from core.blackboard import Blackboard
from core.models import SetupOutput


# ─────────────────────────────────────────────────────────────────────────────
# 個性軸定義
# ─────────────────────────────────────────────────────────────────────────────

_SPEECH_RHYTHM: list[str] = ["簡短", "絮叨", "停頓多", "喜歡反問"]
_EMOTIONAL_BASE: list[str] = ["壓抑", "焦躁", "疏離", "過度熱情"]
_QUIRK: list[str] = ["搓手", "不看眼睛", "反覆確認", "輕聲自語", "咬指甲"]


def roll_personality_axes(rng: random.Random | None = None) -> dict:
    """從固定池各抽一軸，回傳個性骰結果 dict。

    Args:
        rng: 可選的 random.Random 實例；未傳則使用 random 模組預設（非決定性）。

    Returns:
        dict 含三個鍵：speech_rhythm、emotional_base、quirk。
    """
    r = rng if rng is not None else random
    return {
        "speech_rhythm": r.choice(_SPEECH_RHYTHM),
        "emotional_base": r.choice(_EMOTIONAL_BASE),
        "quirk": r.choice(_QUIRK),
    }


# ─────────────────────────────────────────────────────────────────────────────
# run_setup
# ─────────────────────────────────────────────────────────────────────────────

def run_setup(
    caller: SkillCaller,
    blackboard: Blackboard,
    opts: dict,
    rng: random.Random | None = None,
) -> list[str]:
    """執行 Setup Agent：擲骰 → 呼 LLM → 寫 Blackboard → 回傳 opening_sequence。

    Args:
        caller:     SkillCaller 實例（含 LLM client + SkillLoader）。
        blackboard: 中央狀態容器。
        opts:       玩家設定，例如 {theme, npc_count, protagonist_name, tone}。
        rng:        可選的 random.Random，供測試產生決定性骰結果。

    Returns:
        opening_sequence（list[str]），直接傳給前端顯示。

    Raises:
        任何 caller.call() 拋出的例外都直接往上拋（B8：不降級）。
    """
    npc_count: int = opts.get("npc_count", 3)

    # 1. 為每個 NPC 先擲骰個性軸（免 LLM token）
    npc_personalities: list[dict[str, Any]] = [
        {"npc_index": i, "personality_axes": roll_personality_axes(rng)}
        for i in range(npc_count)
    ]

    # 2. 組 context（opts + 骰出來的軸，讓 LLM 依軸生 voice_sample）
    context: dict[str, Any] = {
        **opts,
        "npc_personality_axes": npc_personalities,
    }

    # 3. 呼叫 LLM（B8：不捕捉例外，直接往上拋）
    setup_output: SetupOutput = caller.call(
        "setup",
        context,
        output_model=SetupOutput,
    )

    # 4. 寫入 Blackboard
    #    writer='setup' 擁有最高權限，可寫全部欄位（含 real_bible 錨點）

    # 4a. real_bible（錨點，鎖死後不可由其他 agent 改寫）
    blackboard.write("setup", "real_bible", setup_output.real_bible)

    # 4b. npc_registry：轉為 list[dict]（model_dump），保留全部欄位
    npc_list: list[dict] = [npc.model_dump() for npc in setup_output.npc_registry]
    blackboard.write("setup", "npc_registry", npc_list)

    # 4c. protagonist
    blackboard.write("setup", "protagonist", setup_output.protagonist)

    # 4d. scene_registry：SceneRegistry Pydantic → dict
    blackboard.write("setup", "scene_registry", setup_output.scene_registry.model_dump())

    # 4e. revealed_bible：初始化為空結構（玩家尚未發現任何碎片）
    blackboard.write("setup", "revealed_bible", {
        "revealed_fragments": [],
        "known_atmosphere": [],
        "truth_progress": {},          # NR0：各真相碎片的揭露等級（hinted/observed/…）
    })

    # 5. 回傳 opening_sequence
    return setup_output.opening_sequence
