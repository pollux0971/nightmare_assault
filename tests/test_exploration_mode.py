"""ExplorationMode / ReviewMode Lock（撤離鎖）—— 撤退整理時停止自動調查推進。

核心修補：撤到安全區後 current_area 必須**持久**（不被 kernel 推回站內），且 review 模式
不發 reveal、不新增 object、提供 return_inside/review_notes/inspect_inventory/end_campaign。
"""
from __future__ import annotations

import core.constants as C
from core.world.model import SAFE_ZONE_AREA_ID, OBJECT
from core.narrative.exploration_mode import (
    resolve_mode, is_review_locked, wants_notes, build_review_decision_point,
    ACTIVE_EXPLORATION, REVIEW_MODE, TEMPORARY_RETREAT, CAMPAIGN_END_REQUESTED,
    REVIEW_AFFORDANCES, RETURN_INSIDE, REVIEW_NOTES, INSPECT_INVENTORY, END_CAMPAIGN_ACTION,
)
from core.narrative.exit_resolver import resolve_exit_affordance


def _started_loop(monkeypatch):
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from tests.test_narrative_v2_integration_nr import _loop
    loop = _loop(); loop.start({"theme": "x", "npc_count": 1})
    return loop


# ── resolve_mode：措辭 → 模式 ────────────────────────────────────────────────
def test_resolve_mode_phrases():
    ed = lambda t: resolve_exit_affordance(t)
    # 退到外面 → temporary_retreat
    assert resolve_mode("我退到外面喘口氣", ACTIVE_EXPLORATION, ed("我退到外面喘口氣")) == TEMPORARY_RETREAT
    # 整理線索 → review_mode
    assert resolve_mode("我在這裡整理線索", ACTIVE_EXPLORATION, ed("我在這裡整理線索")) == REVIEW_MODE
    # 「不回研究站 / 不碰真相」→ review_mode（不可因含「回研究站」子字串而誤判 active）
    assert resolve_mode("先退到外面，不回研究站，不碰真相",
                        ACTIVE_EXPLORATION, ed("先退到外面，不回研究站")) == REVIEW_MODE
    # 明確結束 → campaign_end_requested
    assert resolve_mode("我結束本次調查，接受結果", ACTIVE_EXPLORATION,
                        ed("我結束本次調查，接受結果")) == CAMPAIGN_END_REQUESTED
    # 回去研究站 → 解鎖 active
    assert resolve_mode("我回去研究站，重新進入調查", REVIEW_MODE,
                        ed("我回去研究站")) == ACTIVE_EXPLORATION
    assert resolve_mode("重新進入研究站", REVIEW_MODE, ed("重新進入研究站")) == ACTIVE_EXPLORATION


def test_resolve_mode_sticky():
    # 已在 review，玩家輸入中性句（沒明確再入/結束）→ 維持鎖定（不自動恢復調查）
    assert resolve_mode("我看著手裡的紙", REVIEW_MODE, resolve_exit_affordance("我看著手裡的紙")) == REVIEW_MODE
    assert resolve_mode("我發呆了一會", TEMPORARY_RETREAT,
                        resolve_exit_affordance("我發呆了一會")) == TEMPORARY_RETREAT
    # active 維持 active
    assert resolve_mode("我往前走", ACTIVE_EXPLORATION, resolve_exit_affordance("我往前走")) == ACTIVE_EXPLORATION


def test_is_review_locked():
    assert is_review_locked(REVIEW_MODE) and is_review_locked(TEMPORARY_RETREAT)
    assert not is_review_locked(ACTIVE_EXPLORATION) and not is_review_locked(CAMPAIGN_END_REQUESTED)
    assert wants_notes("根據已知線索整理筆記") and not wants_notes("我往前走")


# ── build_review_decision_point：四個 canonical 選項 ─────────────────────────
def test_build_review_decision_point_options():
    from core.models import DecisionPoint, BeatMeta
    base = DecisionPoint(situation_recap="原本的處境", decision_type="action",
                         beat_meta=BeatMeta(beat_number=3))
    dp = build_review_decision_point(base, notes_text="你理了一遍線索。")
    texts = [o.text for o in dp.suggested_options]
    assert len(texts) == 4
    assert any("回到研究站" in t for t in texts)       # return_inside
    assert any("整理筆記" in t for t in texts)         # review_notes
    assert any("隨身" in t for t in texts)             # inspect_inventory
    assert any("結束本次調查" in t for t in texts)     # end_campaign
    assert "你理了一遍線索" in dp.situation_recap


# ══ loop 整合：撤離鎖 ═════════════════════════════════════════════════════════
def test_withdraw_then_current_area_persists_durability(monkeypatch):
    """核心修補：撤退後 current_area 跨 beat 持久，不被 kernel 推回站內。"""
    loop = _started_loop(monkeypatch)
    out1 = loop.step("先退到外面整理線索，不結束本次調查")
    assert not out1.get("ended")
    assert loop._world.current_area == SAFE_ZONE_AREA_ID
    assert is_review_locked(loop._exploration_mode)
    # 下一 beat：中性/整理句（未明確再入）→ current_area **仍**為安全區（durability）
    out2 = loop.step("我站在外面，把手裡的線索重新理一遍，暫時不進去")
    assert not out2.get("ended")
    assert loop._world.current_area == SAFE_ZONE_AREA_ID
    assert loop.world_progress(out2["decision_point"])["current_area"] == SAFE_ZONE_AREA_ID
    # 再一 beat 仍黏著
    out3 = loop.step("我繼續發呆，整理思緒")
    assert loop._world.current_area == SAFE_ZONE_AREA_ID


def test_reenter_restores_active_exploration(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop.step("先退到外面整理線索，不結束本次調查")
    assert is_review_locked(loop._exploration_mode)
    out = loop.step("我回去研究站，重新進入調查")
    assert loop._exploration_mode == ACTIVE_EXPLORATION
    assert not is_review_locked(loop._exploration_mode)
    assert not out.get("ended")


def test_review_mode_blocks_reveal_updates(monkeypatch):
    loop = _started_loop(monkeypatch)
    led = loop._reveal_ledger
    loop.step("先退到外面整理線索，不結束本次調查")   # 進 review
    before = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    out = loop.step("我根據已知線索整理筆記，不深入")
    after = {tid: t.level for tid, t in (getattr(led, "truths", {}) or {}).items()}
    assert before == after                              # review 模式不推進任何 reveal
    assert out["world_progress"]["investigation_state"] == REVIEW_MODE


def test_review_mode_blocks_new_object_entity(monkeypatch):
    loop = _started_loop(monkeypatch)
    loop.step("先退到外面整理線索，不結束本次調查")   # 進 review（review_locked）
    n_obj = len(loop._world.by_kind(OBJECT))
    # review 模式下，即使敘事提到可拾取物件，也**不**新增 object entity
    loop._world_model_tick("我環顧四周", "桌上撿起一張員工證，旁邊有把鑰匙。",
                           None, review_locked=True)
    assert len(loop._world.by_kind(OBJECT)) == n_obj    # 沒有新增
    # 但明確檢查**已知**物件仍可（只動既有 entity）
    if n_obj:
        known = loop._world.by_kind(OBJECT)[0]
        loop._world_model_tick(f"我檢查那個{known.label}", "你端詳它。", None, review_locked=True)
        assert loop._world.find(known.label).state == "inspected"


def test_review_mode_available_next_has_canonical_actions(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    avail = out["world_progress"]["available_next"]
    for a in (RETURN_INSIDE, REVIEW_NOTES, INSPECT_INVENTORY, END_CAMPAIGN_ACTION):
        assert a in avail
    assert set(REVIEW_AFFORDANCES).issubset(set(avail))


def test_review_dp_offers_return_and_end(monkeypatch):
    loop = _started_loop(monkeypatch)
    out = loop.step("先退到外面整理線索，不結束本次調查")
    texts = [o for o in out["world_progress"]["available_next"]]
    dp = out["decision_point"]
    opt_texts = [getattr(o, "text", "") for o in (dp.suggested_options or [])]
    assert any("回到研究站" in t for t in opt_texts)
    assert any("結束本次調查" in t for t in opt_texts)


def test_review_mode_entities_here_only_safe_zone(monkeypatch):
    loop = _started_loop(monkeypatch)
    # 先在 active 登記一個站內 object（無 area 綁定）
    loop._world_model_tick("我看桌上", "桌上撿起一張員工證。", None)
    assert loop._world.find("員工證", kind=OBJECT) is not None
    # 撤退進 review → entities_here 只顯示安全區的 entity（站內 object 不顯示）
    out = loop.step("先退到外面整理線索，不結束本次調查")
    wm = out["world_progress"]["world_model"]
    here_labels = [e["label"] for e in wm["entities_here"]]
    assert "員工證" not in here_labels                  # 站內物件不在安全區 entities_here
