"""NPC First Contact Gate（patch v0.7 P2）—— 未介紹 NPC 首次對話須先 onboarding，不當資訊端點。"""
from __future__ import annotations

import core.constants as C
from core.agents.npc_chat import build_npc_chat_context, run_npc_chat_structured
from core.world.actor_profile import get_npc_profile, UNINTRODUCED, INTRODUCED


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


class _FakeChatCaller:
    """最小 npc-chat caller：回固定純文字回覆（不吐 JSON/entity_delta）。"""
    def __init__(self, reply):
        self.reply = reply

    def call(self, agent, ctx, **kw):
        return self.reply


def _npc_name(loop):
    return (loop.bb.snapshot().get("npc_registry") or [{}])[0].get("name")


# ── context：unintroduced → 帶 first_contact + npc_profile ─────────────────────
def test_context_has_first_contact_when_unintroduced(monkeypatch):
    loop = _started_loop(monkeypatch)
    name = _npc_name(loop)
    assert get_npc_profile(loop.bb, name).intro_state == UNINTRODUCED
    ctx = build_npc_chat_context(loop.bb, name, "通訊設備在哪？")
    assert "npc_profile" in ctx and ctx["npc_profile"]["personality_description"]
    assert "first_contact" in ctx
    fc = ctx["first_contact"]
    assert fc["is_first_contact"] and fc["player_question"] == "通訊設備在哪？"
    assert any("位置" in m or "做什麼" in m for m in fc["must_include"])


def test_context_no_first_contact_after_introduced(monkeypatch):
    loop = _started_loop(monkeypatch)
    name = _npc_name(loop)
    from core.world.actor_profile import mark_introduced
    mark_introduced(loop.bb, name)
    ctx = build_npc_chat_context(loop.bb, name, "你還好嗎？")
    assert "first_contact" not in ctx                    # 已介紹 → 不再注入首接觸
    assert "npc_profile" in ctx                          # 但個性語氣恆給


# ── run：首次對話成功 → intro_state 轉 introduced；不推 reveal ─────────────────
def test_first_chat_marks_introduced_and_no_reveal(monkeypatch):
    loop = _started_loop(monkeypatch)
    name = _npc_name(loop)
    led = loop._reveal_ledger
    before = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    caller = _FakeChatCaller(
        "那人沒立刻回答，先看了看你手裡的燈。「你不是值班的。」他把工具塞回口袋，"
        "「通訊設備？機房方向，但我不建議你一個人過去。」")
    assert get_npc_profile(loop.bb, name).intro_state == UNINTRODUCED
    resp = run_npc_chat_structured(caller, loop.bb, name, "通訊設備在哪？")
    assert resp.visible_reply                             # 有回覆
    assert get_npc_profile(loop.bb, name).intro_state == INTRODUCED   # 轉 introduced
    after = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    assert before == after                               # 首次接觸不推 reveal


def test_known_role_has_no_hidden_truth(monkeypatch):
    loop = _started_loop(monkeypatch)
    name = _npc_name(loop)
    p = get_npc_profile(loop.bb, name)
    # known_role 是表層職業，不得含 secret_core 內容（FakeCaller setup 的 NPC）
    npc = next(n for n in loop.bb.snapshot()["npc_registry"] if n.get("name") == name)
    secret = (npc.get("secret_core") or "")
    if secret:
        assert secret not in p.known_role and secret not in p.surface_motive
