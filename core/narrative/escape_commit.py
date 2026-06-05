"""core.narrative.escape_commit — 兩段式逃脫（NR4，敘事控制 v0.2）。

「我試圖離開」不該即刻觸發結局。改成：
    attempt_escape → exit_candidate_found（出口發現 beat + 選項）→ commit_escape → 結算結局。

首次逃脫意圖 → 轉成出口發現（不結局）；玩家明確提交才讓結局結算。
零 LLM、規則版。對應 dev/CONTRACTS.md §十四（EscapeCommitGate）。
"""
from __future__ import annotations

# 逃脫意圖關鍵詞（與 ending_gate._ESCAPE_WORDS 對齊）
_ESCAPE_WORDS = ["逃", "離開", "出口", "出去", "逃出", "逃離", "escape", "衝出", "跑出", "走出去"]
# 明確「提交離開」的關鍵詞（比單純逃脫更堅決）
_COMMIT_WORDS = ["確定離開", "就走", "離開吧", "我要走", "現在離開", "頭也不回",
                 "下定決心", "毫不猶豫", "決定離開", "穿過門", "踏出去", "走出去不回頭",
                 "commit", "提交離開"]


def is_escape_intent(text: str) -> bool:
    return any(w in (text or "") for w in _ESCAPE_WORDS)


def is_explicit_commit(text: str) -> bool:
    return any(w in (text or "") for w in _COMMIT_WORDS)


# 出口發現 beat 要交給 story 的義務（描述出口 + 給「離開/繼續調查」選擇）
EXIT_CANDIDATE_OBLIGATION = (
    "玩家想離開。描述他們**找到了一條可能的出口**（一扇門、一道樓梯、一個破口），"
    "但離開與否是重大抉擇。明確提供選項：下定決心離開（commit）、或留下繼續調查／處理威脅。"
    "**不要**在這個 beat 就讓玩家離開或結束遊戲。")


class EscapeCommitGate:
    """判斷逃脫流程要進到哪一步。狀態 exit_candidate_found 由 loop 持有。"""

    def decide(self, player_decision: str, *, exit_candidate_found: bool) -> str:
        """回傳 'await_commit' | 'commit' | 'none'。

        - 尚未發現出口 + 有逃脫意圖（非明確提交）→ 'await_commit'（產出口發現 beat，不結局）。
        - 已發現出口 + （明確提交 或 再次逃脫意圖）→ 'commit'（可結算結局）。
        - 一開始就明確提交且已找到出口 → 'commit'。
        - 其餘 → 'none'。
        """
        intent = is_escape_intent(player_decision)
        commit = is_explicit_commit(player_decision)
        if not exit_candidate_found:
            if intent and not commit:
                return "await_commit"
            if commit:
                return "await_commit"           # 連出口都還沒找到 → 先找出口，仍不即結局
            return "none"
        # 已找到出口
        if intent or commit:
            return "commit"
        return "none"
