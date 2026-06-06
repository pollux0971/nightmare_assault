"""開場變體契約測試（補丁 v0.8）。

覆蓋 P1 schema / P2 cooldown / P3 selector / gate，以及 MASTER_CLAUDE_CODE_PROMPT 指定的必測項：
  - forbidden_literals 含「紙條」→ opening 不得出現紙條或常見別稱。
  - forbidden_literals 含「林晨」→ opening 不得出現林晨。
  - forbidden_archetypes 含 missing_person → 不得生成找人開場。
  - message_medium=corrupted_log → 不得輸出 handwritten note。
  - 多次 opening 不應全部是 missing_person。
  - Contract 不得強制收束劇情。
"""
from __future__ import annotations

from random import Random

import pytest

from core.narrative.opening_pools import VariationPools, default_pools
from core.narrative.opening_variation import (
    CooldownLedger,
    MessageMedium,
    MotiveArchetype,
    OpeningVariationContract,
    PersonalAnchorType,
    violates_forbidden_literals,
    weighted_choice,
)
from core.narrative.opening_variation_gate import check_opening_output
from core.narrative.opening_variation_selector import (
    build_contract,
    generate_opening_contract,
    usage_for_contract,
)


# ── P1 schema ────────────────────────────────────────────────────────────────

def test_enums_cover_taxonomy():
    assert len(list(MotiveArchetype)) == 12
    assert len(list(PersonalAnchorType)) == 10
    assert len(list(MessageMedium)) == 18
    assert MotiveArchetype.MISSING_PERSON.value == "missing_person"


def test_contract_roundtrip_dict():
    c = OpeningVariationContract(
        motive_archetype="verify_event", personal_anchor_type="no_person_anchor",
        message_medium="access_record", initial_goal="x", first_interactable_type="terminal_or_console")
    d = c.to_dict()
    assert d["motive_archetype"] == "verify_event"
    c2 = OpeningVariationContract.from_dict({**d, "unexpected_field": 1})
    assert c2 == c


def test_contract_has_no_plot_convergence_fields():
    """Contract 不得強制收束劇情：只有開場素材欄位，沒有任何 ending/forced/resolution 欄位。"""
    fields = set(OpeningVariationContract.__dataclass_fields__)  # type: ignore[attr-defined]
    forbidden = {"ending", "ending_type", "forced_outcome", "resolution",
                 "climax", "final_truth", "must_reach", "convergence"}
    assert fields.isdisjoint(forbidden)


# ── gate ─────────────────────────────────────────────────────────────────────

def test_forbidden_literal_paper_is_blocked():
    v = check_opening_output("門縫裡夾著一張紙條。", forbidden_literals=["紙條"],
                             forbidden_archetypes=[], expected_message_medium="corrupted_log")
    assert any(x.type == "forbidden_literal" for x in v)


def test_forbidden_literal_paper_alias_blocked_via_medium():
    """corrupted_log 卻寫「便條/字條」也算 medium mismatch（別名群）。"""
    for alias in ("便條", "字條", "手寫留言"):
        v = check_opening_output(f"桌上有一張{alias}。", forbidden_literals=[],
                                 forbidden_archetypes=[], expected_message_medium="corrupted_log")
        assert any(x.type == "message_medium_mismatch" for x in v), alias


def test_forbidden_literal_linchen_is_blocked():
    assert violates_forbidden_literals("林晨曾經來過。", ["林晨"]) == ["林晨"]
    v = check_opening_output("林晨曾經來過。", forbidden_literals=["林晨"], forbidden_archetypes=[])
    assert any(x.value == "林晨" for x in v)


def test_missing_person_archetype_blocked():
    v = check_opening_output("你來這裡是為了尋找失蹤的人。", forbidden_literals=[],
                             forbidden_archetypes=["missing_person"])
    assert any(x.value == "missing_person" for x in v)


def test_message_medium_corrupted_log_not_note():
    v = check_opening_output("你看到一張手寫留言。", forbidden_literals=[],
                             forbidden_archetypes=[], expected_message_medium="corrupted_log")
    assert any(x.type == "message_medium_mismatch" for x in v)


def test_gate_catches_simplified_chinese():
    """真實模型常輸出簡體；gate 必須同時抓繁簡（實測 deepseek 簡體輸出曾繞過）。"""
    # 簡體「纸条」+ corrupted_log medium → message_medium_mismatch
    v = check_opening_output("门缝里夹着一张纸条。", forbidden_literals=[],
                             forbidden_archetypes=[], expected_message_medium="corrupted_log")
    assert any(x.type == "message_medium_mismatch" for x in v)
    # 簡體「失踪/寻找」+ missing_person 被擋 → forbidden_archetype
    v2 = check_opening_output("你来这里是为了寻找失踪的妹妹。", forbidden_literals=[],
                              forbidden_archetypes=["missing_person"])
    assert any(x.value == "missing_person" for x in v2)


def test_clean_opening_has_no_violation():
    v = check_opening_output("終端機殘留著一段錯誤紀錄，時間戳對不上。",
                             forbidden_literals=["紙條", "林晨"],
                             forbidden_archetypes=["missing_person", "handwritten_note"],
                             expected_message_medium="corrupted_log")
    assert v == []


# ── P2 cooldown ──────────────────────────────────────────────────────────────

def test_cooldown_forbids_recent_archetype():
    ledger = CooldownLedger(recent_archetypes={"missing_person": 10}, archetype_cooldown_runs=2)
    assert "missing_person" in ledger.forbidden_archetypes(12)
    assert "missing_person" not in ledger.forbidden_archetypes(13)   # 窗外


def test_cooldown_forbids_recent_literal():
    ledger = CooldownLedger(recent_literals={"紙條": 10}, literal_cooldown_runs=3)
    assert "紙條" in ledger.forbidden_literals(11)
    assert "紙條" in ledger.forbidden_literals(13)
    assert "紙條" not in ledger.forbidden_literals(14)


def test_cooldown_record_and_roundtrip():
    ledger = CooldownLedger()
    ledger.record(5, literals=["紙條"], archetypes=["missing_person"])
    d = ledger.to_dict()
    back = CooldownLedger.from_dict(d)
    assert back.recent_literals["紙條"] == 5
    assert back.recent_archetypes["missing_person"] == 5


# ── P3 selector ──────────────────────────────────────────────────────────────

def test_weighted_choice_respects_forbidden():
    rng = Random(0)
    key, exhausted = weighted_choice(rng, {"a": 100, "b": 1}, {"a"})
    assert key == "b" and not exhausted


def test_weighted_choice_exhausted_falls_back():
    rng = Random(0)
    key, exhausted = weighted_choice(rng, {"a": 1}, {"a"})
    assert key == "a" and exhausted


def test_selector_respects_forbidden_missing_person():
    ledger = CooldownLedger(recent_archetypes={"missing_person": 10}, archetype_cooldown_runs=3)
    c = build_contract(rng=Random(1), current_run=11, ledger=ledger,
                       pools=VariationPools(motive_weights={"missing_person": 100, "verify_event": 1},
                                            medium_weights={"corrupted_log": 1},
                                            anchor_weights={"no_person_anchor": 1},
                                            interactable_weights={"terminal_or_console": 1}))
    assert c.motive_archetype != "missing_person"
    assert "missing_person" in c.forbidden_archetypes


def test_selector_deterministic_same_seed():
    led = CooldownLedger()
    a = build_contract(rng=Random(42), current_run=1, ledger=led, selector_seed=42)
    b = build_contract(rng=Random(42), current_run=1, ledger=CooldownLedger(), selector_seed=42)
    assert a.motive_archetype == b.motive_archetype
    assert a.message_medium == b.message_medium
    assert a.personal_anchor_type == b.personal_anchor_type


def test_multiple_openings_not_all_missing_person():
    """多次 opening 不應全部是 missing_person（預設池已降權 + 抽樣多樣）。"""
    motives = []
    for i in range(20):
        c = build_contract(rng=Random(i), current_run=i, ledger=CooldownLedger(),
                           selector_seed=i)
        motives.append(c.motive_archetype)
    assert motives.count("missing_person") < len(motives)
    assert len(set(motives)) >= 4                    # 至少 4 種不同動機


def test_contract_initial_goal_has_no_proper_nouns():
    """initial_goal/objective 是抽象提示，不含 紙條/林晨 等具體錨點。"""
    for i in range(12):
        c = build_contract(rng=Random(i), current_run=i, ledger=CooldownLedger())
        text = (c.initial_goal or "") + (c.opening_objective_sentence or "")
        assert "林晨" not in text and "紙條" not in text


def test_usage_for_contract_records_medium_literals():
    pools = default_pools()
    c = build_contract(rng=Random(3), current_run=1, ledger=CooldownLedger(),
                       pools=VariationPools(motive_weights={"verify_event": 1},
                                            medium_weights={"handwritten_note": 1},
                                            anchor_weights={"no_person_anchor": 1},
                                            interactable_weights={"terminal_or_console": 1}))
    literals, archetypes = usage_for_contract(c, pools)
    assert "紙條" in literals                          # 用了手寫便條 → 紙條家族進 cooldown
    assert "handwritten_note" in archetypes and "verify_event" in archetypes


# ── persistence (P2 store) ───────────────────────────────────────────────────

def test_generate_opening_contract_persists_cooldown(tmp_path):
    from core.persistence.db import Database
    db = Database(str(tmp_path / "ov.db"))
    # 第一局：用了某 medium → 其字串進 ledger。
    c1 = generate_opening_contract(db, "run-1")
    lit1, _ = usage_for_contract(c1)
    # 第二局：剛用過的 literal 應出現在 forbidden_literals。
    c2 = generate_opening_contract(db, "run-2")
    if lit1:                                          # 該 medium 有字串才會被擋
        assert any(l in c2.forbidden_literals for l in lit1)
    db.close()


def test_store_reset(tmp_path):
    from core.narrative.opening_cooldown_store import OpeningVariationStore
    from core.persistence.db import Database
    db = Database(str(tmp_path / "ov2.db"))
    store = OpeningVariationStore(db.connection)
    assert store.next_run_index() == 1
    assert store.next_run_index() == 2
    store.reset()
    assert store.next_run_index() == 1
    db.close()


# ── P5/P6 prompt + enforcement helpers ───────────────────────────────────────

def _contract(**kw) -> OpeningVariationContract:
    base = dict(motive_archetype="investigate_signal", personal_anchor_type="past_self",
                message_medium="corrupted_log", initial_goal="追查一個不該存在的訊號",
                first_interactable_type="terminal_or_console",
                personal_anchor_label="過去的你自己",
                forbidden_literals=["紙條", "林晨"],
                forbidden_archetypes=["missing_person", "handwritten_note"])
    base.update(kw)
    return OpeningVariationContract(**base)


def test_apply_contract_to_context_injects_and_keeps_safe():
    from core.narrative.opening_variation_prompt import apply_contract_to_context
    ctx = {"narrative_obligations": ["既有義務"], "instruction": "base"}
    out = apply_contract_to_context(ctx, _contract())
    assert out["opening_variation_contract"]["message_medium"] == "corrupted_log"
    assert any("禁止" in o or "契約" in o for o in out["narrative_obligations"])
    assert ctx["narrative_obligations"] == ["既有義務"]   # 原 ctx 不被改（回新 dict）
    assert "紙條" not in out["instruction"]                # 指令本身不洩漏 forbidden 字串為示範


def test_fallback_opening_is_clean():
    from core.narrative.opening_variation_gate import check_opening_output
    from core.narrative.opening_variation_prompt import fallback_opening
    c = _contract()
    narrative, dp = fallback_opening(c)
    v = check_opening_output(narrative, forbidden_literals=c.forbidden_literals,
                             forbidden_archetypes=c.forbidden_archetypes,
                             expected_message_medium=c.message_medium)
    assert v == []                                        # fallback 保證乾淨
    assert dp.suggested_options and not dp.is_narration_only


def test_fallback_avoids_forbidden_medium_surface():
    """即使 medium=handwritten_note 但 紙條 仍在 cooldown，fallback 也不能用紙條家族字串。"""
    from core.narrative.opening_variation_gate import check_opening_output
    from core.narrative.opening_variation_prompt import fallback_opening
    c = _contract(message_medium="handwritten_note", forbidden_literals=["紙條", "便條", "字條"],
                  forbidden_archetypes=[])
    narrative, _ = fallback_opening(c)
    assert check_opening_output(narrative, forbidden_literals=c.forbidden_literals,
                                forbidden_archetypes=[], expected_message_medium=None) == []


# ── 整合：透過 BeatLoop.start（flag ON）驗 gate → repair → fallback ─────────────

class _OVCaller:
    """可控 story 輸出的假 caller：story 依序消耗 narratives；setup/warden 給定值。"""
    def __init__(self, story_narratives):
        from core.models import SetupOutput, SceneRegistry, Location, NPCBible
        self._story = list(story_narratives)
        self._i = 0
        self.setup_out = SetupOutput(
            real_bible={"world_truth": {"what_really_happened": "x", "the_threat_is": "y",
                                        "deadly_rule": "z"},
                        "hard_triggers": [], "ending_conditions": [],
                        "revelation_pool": [{"id": "f1", "title": "t", "content": "c"}]},
            npc_registry=[NPCBible(name="A", profession="p", personality="mysterious",
                                   voice_sample="…", public_face="f", secret_core="s",
                                   self_aware=True, appearance="")],
            protagonist={"name": "我", "starting_situation": "醒來"},
            scene_registry=SceneRegistry(current_location="hall",
                                         known_locations=[Location(id="hall", name="大廳", description="")]),
            opening_sequence=["你醒來。"])

    def call(self, agent, context, output_model=None, temperature=None, **kw):
        from core.models import WardenOutput
        if agent == "setup":
            return self.setup_out
        if agent == "warden":
            return WardenOutput(directive_to_story="繼續")
        raise AssertionError(agent)

    def stream(self, agent, context, temperature=None, **kw):
        assert agent == "story"
        narr = self._story[min(self._i, len(self._story) - 1)]
        self._i += 1
        story_json = ('{"situation_recap":"r","decision_type":"action",'
                      '"suggested_options":[{"text":"往前走","tone":"bold"}],'
                      '"beat_meta":{"beat_number":0}}')
        for t in [narr, "<<<DECISION>>>", story_json]:
            yield t


def _ov_loop(caller):
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database
    from core.signal import SignalBus
    return BeatLoop(caller, Blackboard(), Database(), SignalBus(), run_id="ov", use_kernel=True)


def _force_contract(monkeypatch, contract):
    import core.narrative.opening_variation_selector as sel
    monkeypatch.setattr(sel, "generate_opening_contract", lambda db, run_id, **kw: contract)


def test_flag_off_no_contract_in_game_meta(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    monkeypatch.setattr(C, "ENABLE_OPENING_VARIATION_CONTRACT", False)
    loop = _ov_loop(_OVCaller(["走廊很暗，終端機還亮著。"]))
    res = loop.start({"theme": "t", "npc_count": 1})
    assert "opening_variation_contract" not in (loop.bb.game_meta or {})
    assert "opening_variation" not in res


def test_flag_on_contract_in_game_meta_and_obs(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    monkeypatch.setattr(C, "ENABLE_OPENING_VARIATION_CONTRACT", True)
    _force_contract(monkeypatch, _contract())
    loop = _ov_loop(_OVCaller(["走廊很暗，終端機殘留著一段錯誤紀錄。"]))
    res = loop.start({"theme": "t", "npc_count": 1})
    gm = loop.bb.game_meta or {}
    assert gm.get("opening_variation_contract", {}).get("message_medium") == "corrupted_log"
    assert res.get("opening_variation", {}).get("motive_archetype") == "investigate_signal"
    assert not res.get("opening_variation_violation", {}).get("fallback_used")


def test_flag_on_violation_repaired(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    monkeypatch.setattr(C, "ENABLE_OPENING_VARIATION_CONTRACT", True)
    _force_contract(monkeypatch, _contract())
    # 第一版含紙條（違規）；重寫版乾淨 → repair 成功、不 fallback。
    loop = _ov_loop(_OVCaller(["門縫夾著一張紙條。", "終端機殘留著一段錯誤紀錄。"]))
    res = loop.start({"theme": "t", "npc_count": 1})
    viol = res.get("opening_variation_violation", {})
    assert viol.get("repair_attempted") and not viol.get("fallback_used")
    assert not viol.get("has_violation")
    assert "紙條" not in res["narrative"]


def test_flag_on_violation_falls_back(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    monkeypatch.setattr(C, "ENABLE_OPENING_VARIATION_CONTRACT", True)
    _force_contract(monkeypatch, _contract())
    # 兩版都含紙條 → repair 仍違規 → 決定性 fallback（乾淨）。
    loop = _ov_loop(_OVCaller(["夾著紙條。", "還是紙條。"]))
    res = loop.start({"theme": "t", "npc_count": 1})
    viol = res.get("opening_variation_violation", {})
    assert viol.get("repair_attempted") and viol.get("fallback_used")
    assert not viol.get("has_violation")                  # fallback 後乾淨
    assert "紙條" not in res["narrative"]
    assert (loop.bb.game_meta or {}).get("opening_variation_violation", {}).get("fallback_used")
    # repair/fallback 期間的暫存 beat_window 已收斂成單一 beat0 最終版
    bw0 = [b for b in (loop.bb.beat_window or [])
           if isinstance(b, dict) and b.get("beat_number") == 0]
    assert len(bw0) == 1 and "紙條" not in bw0[0]["narrative"]
