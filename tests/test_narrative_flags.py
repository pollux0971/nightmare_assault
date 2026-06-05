"""NC0 — 契約凍結 + 旁路 flag 驗收測試。

驗收：ENABLE_NARRATIVE_CONTROL 預設 OFF（行為不變）；可由 env 開關；narrative 模型可 import；新欄位皆 optional。
"""
from __future__ import annotations

import importlib
import os


def test_flag_default_off(monkeypatch):
    monkeypatch.delenv("ENABLE_NARRATIVE_CONTROL", raising=False)
    import core.constants as c
    importlib.reload(c)
    assert c.ENABLE_NARRATIVE_CONTROL is False           # 預設 OFF → 行為與現況一致


def test_flag_env_toggle(monkeypatch):
    import core.constants as c
    for val, expect in [("true", True), ("1", True), ("on", True),
                        ("false", False), ("0", False), ("", False)]:
        monkeypatch.setenv("ENABLE_NARRATIVE_CONTROL", val)
        importlib.reload(c)
        assert c.ENABLE_NARRATIVE_CONTROL is expect, val
    monkeypatch.delenv("ENABLE_NARRATIVE_CONTROL", raising=False)
    importlib.reload(c)


def test_narrative_models_importable():
    from core.narrative.models import (
        NarrativeContract, ProtagonistMotive, MotifPalette, OpeningBudget,
        OpeningBlueprint, TruthSeed, QualityGateResult, RevealLevel, REVEAL_ORDER,
    )
    nc = NarrativeContract(
        core_premise="廢棄療養院的記憶實驗",
        protagonist_motive=ProtagonistMotive("弟弟失蹤", "找到他", "愧疚", "弟弟的學生證"),
        central_question="林晨還是林晨嗎？",
    )
    assert nc.opening_reveal_limit == "hinted"            # 預設開場最多 hinted
    assert nc.opening_budget.max_named_objects == 3
    assert REVEAL_ORDER[1] == "hinted" and REVEAL_ORDER[-1] == "actionable"
