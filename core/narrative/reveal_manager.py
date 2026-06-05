"""core.narrative.reveal_manager — 全域揭露上限助手（NC4）。

NC3/NC4 用這兩個純函式算「餵 story 的全域揭露上限」（依累積 evidence 漸進、一次升一階）。

> 注意：每個真相**碎片**的揭露進度（per-truth ladder）由 `core.narrative.revelation`
> 的 `RevelationBridge` / `RevealLedger`（strength 驅動）負責，那才是 revealed_bible 的權威來源。
> 本模組只管「全域 context ceiling」這一個標量，兩者各司其職、不重複。
"""
from __future__ import annotations

from core.narrative.models import REVEAL_ORDER       # 單一來源，不重複定義


# ── 全域 context 揭露上限（依累積 evidence 漸進，不跳級）─────────────────────
def allowed_reveal_for(evidence_count: int) -> str:
    """依玩家累積 evidence（如線索數）給「目標」揭露上限。"""
    if evidence_count <= 0:
        return "hinted"
    if evidence_count == 1:
        return "observed"
    if evidence_count <= 3:
        return "suspected"
    if evidence_count <= 5:
        return "confirmed"
    return "actionable"


def next_level_no_skip(current: str, target: str) -> str:
    """從 current 朝 target 前進**最多一階**（不跳級）。"""
    ci = REVEAL_ORDER.index(current) if current in REVEAL_ORDER else 0
    ti = REVEAL_ORDER.index(target) if target in REVEAL_ORDER else ci
    if ti <= ci:
        return current
    return REVEAL_ORDER[ci + 1]
