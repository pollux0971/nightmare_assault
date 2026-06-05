"""NR2 — Answer Debt 驗收測試（敘事控制 v0.2）。

驗收：同問題問 2 次 → requires_payoff；分類正確；partial/answered 重置、refused 遞減；
payoff 義務文字產生；非問題不記債。
"""
from __future__ import annotations

from core.narrative.answer_debt import (
    AnswerDebtTracker, classify_question, payoff_obligation,
)


# ── 重複提問 → 須償還（reference 對齊）─────────────────────────────────────
def test_repeated_question_requires_payoff():
    t = AnswerDebtTracker()
    key = classify_question("432.7 是什麼？")
    assert key and key.startswith("mechanism_question")
    t.register_question(key)
    assert not t.requires_payoff(key)
    t.register_question(key)
    assert t.requires_payoff(key)               # 問第二次 → debt 2 → 須付


# ── 分類：各 category / topic ───────────────────────────────────────────────
def test_classification_categories():
    assert classify_question("林晨在哪裡？").startswith("location_question")
    assert "linchen" in classify_question("林晨在哪裡？")
    assert classify_question("他們是誰？").startswith("identity_question")
    assert classify_question("這裡危險嗎？").startswith("threat_question")
    assert classify_question("我接下來該怎麼辦？").startswith("action_question")
    assert "signal" in classify_question("整點報時是什麼意思？")


# ── 非問題不記債 ─────────────────────────────────────────────────────────────
def test_non_question_returns_none():
    assert classify_question("我往前走。") is None
    assert classify_question("") is None


# ── 償還重置 / 拒答遞減 ─────────────────────────────────────────────────────
def test_answer_resets_or_decrements():
    t = AnswerDebtTracker()
    k = "mechanism_question:signal"
    t.register_question(k); t.register_question(k); t.register_question(k)
    assert t.level(k) == 3
    t.register_answer(k, "refused")
    assert t.level(k) == 2                       # 拒答只遞減
    t.register_answer(k, "partial")
    assert t.level(k) == 0                       # 部分答 → 清零
    assert not t.requires_payoff(k)


# ── payoff 義務文字 ─────────────────────────────────────────────────────────
def test_payoff_obligation_text():
    msg = payoff_obligation("mechanism_question:signal", 2)
    assert "部分答案" in msg and "氛圍" in msg


# ── round-trip（存 game_meta）────────────────────────────────────────────────
def test_roundtrip():
    t = AnswerDebtTracker()
    t.register_question("x:y"); t.register_question("x:y")
    r = AnswerDebtTracker.from_dict(t.to_dict())
    assert r.requires_payoff("x:y")
