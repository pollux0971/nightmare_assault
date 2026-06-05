"""core.narrative.negative_intent — 否定意圖守衛（NegativeIntentGuard，Player Sovereignty）。

玩家明確「拒絕」某件事時，explicit negative intent **優先於** keyword match：
- 「不結束本次調查」**不可**因為含子字串「結束本次調查」就被當成 campaign_end。
- 「不進 B 區」**不可**選 enter_B_area event。

零 LLM、規則版。對應 docs/player-sovereignty-principles.md + P0。
"""
from __future__ import annotations

import re

# 否定前綴（出現在某目標前 → 玩家拒絕該目標）
_NEG_PREFIX = ["不進", "不去", "不要進", "不想進", "先不進", "暫時不進", "別進",
               "不走", "不要走", "先不走", "不踏入", "不靠近", "不碰", "不開",
               "不想", "先不要", "暫時不", "不打算", "拒絕", "不願"]
# 明確「不結束本次調查」的否定（保護 campaign_end 不被誤觸）
_NEG_END = ["不結束", "不是要結束", "不想結束", "先不結束", "別結束", "還不結束",
            "不要結束", "不算結束", "沒有要結束", "不打算結束"]
# 表示「只是暫離 / 不離整局」的語意（強化非 ending）
_NOT_LEAVING_RUN = ["不結束本次調查", "不離開研究站", "只是先", "只是到外面", "整理線索",
                    "稍後再回", "等等再回", "之後再回來"]


def negates_ending(text: str) -> bool:
    """玩家是否明確表示「不要結束本次調查」。True → ExitResolver 不得判 campaign_end。"""
    t = text or ""
    return any(p in t for p in _NEG_END) or any(p in t for p in _NOT_LEAVING_RUN)


def negated_targets(text: str) -> list[str]:
    """抽出玩家明確拒絕的目標（如「不進 B 區」→ 'B 區' / 'B區'）。

    回傳被否定的目標字串清單（可能含空白變體），供 kernel 否決移動到該目標的事件。
    """
    t = text or ""
    out: list[str] = []
    for pre in _NEG_PREFIX:
        idx = 0
        while True:
            i = t.find(pre, idx)
            if i < 0:
                break
            tail = t[i + len(pre):]
            # 取否定詞後面的目標片段（到標點/連接詞為止）
            m = re.match(r"\s*([0-9A-Za-z一-鿿]{1,8})", tail)
            if m:
                target = m.group(1).strip()
                if target and target not in out:
                    out.append(target)
                    out.append(target.replace(" ", ""))   # 「B 區」「B區」都收
            idx = i + len(pre)
    # 去重保序
    seen, uniq = set(), []
    for x in out:
        if x and x not in seen:
            seen.add(x); uniq.append(x)
    return uniq


def is_negated(target: str, negated: list[str]) -> bool:
    """目標（如事件 destination 名）是否落在玩家否定清單內（雙向子字串比對）。"""
    target = (target or "").replace(" ", "")
    for n in negated or []:
        n = (n or "").replace(" ", "")
        if n and (n in target or target in n):
            return True
    return False
