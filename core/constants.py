"""全專案單一真相常數模組（U02）。

所有其他模組應 import 這裡的常數，不可在各處寫死相同值。
"""
from __future__ import annotations

DELIM_CONTINUE = "<<<CONTINUE>>>"
DELIM_DECISION = "<<<DECISION>>>"

# 三層模型分層；實際模型字串由 config 讀，這裡給結構與預設鍵
MODEL_TIERS: dict[str, None] = {"heavy": None, "medium": None, "light": None}

SUMMARY_TOKEN_CAP = 1000      # 滾動摘要 token 上限
BEAT_WINDOW_SIZE = 6          # 保留最近幾個 beat 原文
NARRATION_ONLY_MAX = 3        # 連續旁白型上限，超過強制決策
CONTEXT_THRESHOLD_L1 = 0.70
CONTEXT_THRESHOLD_L2 = 0.85
CONTEXT_THRESHOLD_L3 = 0.95

# ── Narrative Progress Kernel（SK04）──
# 預設 ON；只有明確 ENABLE_PROGRESS_KERNEL=false（不分大小寫）才走舊 LLM 自由流程。
import os as _os
ENABLE_PROGRESS_KERNEL = _os.environ.get("ENABLE_PROGRESS_KERNEL", "true").strip().lower() not in ("false", "0", "no")

# ── Narrative Control（階段 N，patch v0.1）──
# 預設 OFF（旁路控制層）；ENABLE_NARRATIVE_CONTROL=true 才啟用 opening director / reveal ladder / gates。
ENABLE_NARRATIVE_CONTROL = _os.environ.get("ENABLE_NARRATIVE_CONTROL", "false").strip().lower() in ("true", "1", "yes", "on")

# Attractor-based ending（SK13）：結局是吸引子，不是固定終點節點
DANGER_DEATH_THRESHOLD = 5        # danger_level 累積到此 → 死亡吸引子
TRUTH_FRAGMENT_THRESHOLD = 2      # 揭露的真相碎片數到此 → 真相吸引子
ESCAPE_CLUE_THRESHOLD = 5         # 線索數（摸清出路）到此 + 存活 + 有核心 → 逃脫吸引子

# SignalBus 事件名（單一來源，供 signal.py 與訂閱者 import）
EVT_BEAT_COMPLETED = "BEAT_COMPLETED"
EVT_ENDING_TRIGGERED = "ENDING_TRIGGERED"
EVT_RULE_VIOLATION = "RULE_VIOLATION"
EVT_SKILL_CLAIMED = "SKILL_CLAIMED"
EVT_CHATROOM_OPENED = "CHATROOM_OPENED"
EVT_CHATROOM_CLOSED = "CHATROOM_CLOSED"
EVT_NPC_EVOLVED = "NPC_EVOLVED"
EVT_CONTEXT_THRESHOLD = "CONTEXT_THRESHOLD"

ALL_EVENTS = [
    EVT_BEAT_COMPLETED, EVT_ENDING_TRIGGERED, EVT_RULE_VIOLATION, EVT_SKILL_CLAIMED,
    EVT_CHATROOM_OPENED, EVT_CHATROOM_CLOSED, EVT_NPC_EVOLVED, EVT_CONTEXT_THRESHOLD,
]
