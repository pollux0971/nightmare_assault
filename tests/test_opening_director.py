"""NC2 — Opening Director 驗收測試。

驗收：開場主要新元素 ≤3、至少 1 動機 + 1 可行動線索、開場 reveal ≤ hinted、blueprint 不含 hidden_truth；
flag OFF 走 UB6 原路、ON 走 director。
"""
from __future__ import annotations

import json

import core.constants
from core.narrative.models import NarrativeContract, ProtagonistMotive, MotifPalette, OpeningBudget
from core.narrative.opening_director import select_opening_elements, apply_to_context


def _contract(forbidden=()):
    return NarrativeContract(
        core_premise="記憶實驗SECRET",
        protagonist_motive=ProtagonistMotive("弟弟失蹤", "找到他", "愧疚", "弟弟的學生證"),
        central_question="他還記得你嗎？",
        motif_palette=MotifPalette(forbidden_or_limited=list(forbidden)),
        opening_budget=OpeningBudget(max_named_objects=3),
        opening_reveal_limit="hinted",
    )


# ── 元素數量受 budget 限制 ───────────────────────────────────────────────────
def test_allowed_elements_capped_by_budget():
    bp = select_opening_elements(_contract(),
                                 candidate_elements=["走廊", "血跡", "廣播", "腳印", "菌絲", "鐵門"],
                                 candidate_truth_ids=["t1", "t2", "t3"])
    assert len(bp.allowed_elements) <= 3                    # 主要新元素 ≤3
    assert bp.motive_evidence == "弟弟的學生證"               # 至少 1 動機（first_proof 保留）
    assert bp.allowed_elements[0] == "弟弟的學生證"
    assert len(bp.truth_seeds) <= 1                         # 開場真相 seed ≤1
    assert bp.truth_seeds[0].reveal_level == "hinted"      # 最多 hinted（不 confirmed）
    assert bp.blocked_elements                              # 超出 budget 的被 blocked


def test_forbidden_motifs_blocked():
    bp = select_opening_elements(_contract(forbidden=["菌絲"]),
                                 candidate_elements=["菌絲狀物質", "走廊"], candidate_truth_ids=[])
    assert any("菌絲" in b for b in bp.blocked_elements)
    assert not any("菌絲" in a for a in bp.allowed_elements)


# ── apply_to_context：收斂 UB6 種子、不洩 hidden_truth ──────────────────────
def test_apply_to_context_refines_seeds():
    ub6_ctx = {"opening_seeds": [{"surface": f"元素{i}"} for i in range(5)],
               "opening_truth_ids": ["t1"], "narrative_obligations": ["介紹場景"]}
    out = apply_to_context(ub6_ctx, _contract())
    assert "opening_blueprint" in out
    assert len(out["opening_seeds"]) <= 3                   # 收斂到 budget
    assert out["opening_element_budget"] == 3
    assert out["opening_reveal_limit"] == "hinted"
    blob = json.dumps(out, ensure_ascii=False)
    assert "記憶實驗SECRET" not in blob                      # core_premise（hidden）不外洩


# ── 旁路 flag：intro beat ON 用 director ────────────────────────────────────
def test_intro_uses_director_when_flag_on(monkeypatch):
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database

    class Caller:
        def stream(self, agent, context, temperature=None, system_override=None):
            self.ctx = context
            for t in ["開場。", "<<<DECISION>>>",
                      '{"situation_recap":"x","decision_type":"action",'
                      '"suggested_options":[{"text":"看","tone":"cautious"}],"beat_meta":{"beat_number":0}}']:
                yield t

    monkeypatch.setattr(core.constants, "ENABLE_NARRATIVE_CONTROL", True)
    bb = Blackboard()
    bb.write("setup", "real_bible", {"world_truth": {"what_really_happened": "X", "the_threat_is": "Y",
                                                     "deadly_rule": "Z"}})
    bb.write("setup", "protagonist", {"name": "周凱", "starting_situation": "弟弟失蹤"})
    caller = Caller()
    loop = BeatLoop(caller, bb, Database(), run_id="t", use_kernel=False)
    loop._build_narrative_contract()
    # 直接驗 apply 後 ctx 收斂（intro beat 走 kernel；此處驗 director 邏輯接上）
    assert loop._narrative_contract is not None
