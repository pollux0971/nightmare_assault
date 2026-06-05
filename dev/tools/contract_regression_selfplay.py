#!/usr/bin/env python3
"""contract_regression_selfplay — WorldConsequence vs TruthEvidence Split 契約回歸。

逐一驅動每個 ActionIntent 類別，量測 reveal_progress delta + TruthEvidenceGate 決定，
斷言七條契約：
  C1 world_navigation       不推 reveal
  C2 world_review           不推 reveal
  C3 npc_fact_query         不推 reveal
  C4 object_inspection      不推 reveal（除非物件 truth-bearing → gate 例外驗證）
  C5 truth_investigation    仍可推 reveal
  C6 structured evidence    仍可推 reveal
  C7 review_mode            不產生未記帳 fact

阻擋類（C1–C4,C7）走真 LLM loop.step（阻擋與否與 LLM 無關，確定性）；
放行類（C5,C6）注入 truth-mapped clue / 結構化 evidence 以確定性驗證「確實推進」。

用法：.venv/bin/python dev/tools/contract_regression_selfplay.py
"""
from __future__ import annotations

import json
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))


def main(argv=None):
    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    from dev.tools.agent_play import _make_caller, _new_loop, _reveal_progress
    from core.narrative.truth_evidence_gate import TruthEvidenceGate
    from core.narrative.action_intent import OBJECT_INSPECTION
    from core.world.model import FACT

    caller, note = _make_caller(str(ROOT / "config" / "config.json"), no_llm=False)
    loop = _new_loop(caller)
    P = print
    P(f"[caller={note}  flag=ON]")
    loop.start({"theme": "午夜的廢棄海事研究站（弟弟林晨失蹤）",
                "protagonist_name": "周凱", "npc_count": 2})

    results = []   # (cid, desc, ok, detail)

    def _rp(loop):
        return _reveal_progress(loop)

    def _reveal_sum(rp):
        return rp["hinted"] + rp["observed"] + rp["suspected"] + rp["confirmed"]

    def step_check(cid, desc, action, expect_class=None):
        before = _rp(loop)
        out = loop.step(action)
        after = _rp(loop)
        delta = _reveal_sum(after) - _reveal_sum(before)
        ac = out.get("action_class")
        gate = out.get("reveal_gate_allowed")
        ok = (delta == 0) and (gate is False)
        if expect_class is not None:
            ok = ok and (ac == expect_class)
        detail = (f"action_class={ac} gate_allowed={gate} reason={out.get('reveal_gate_block_reason')} "
                  f"reveal_delta={delta} blocked={out.get('blocked_reveal_candidates')}")
        results.append((cid, desc, ok, detail))
        P(f"\n[{cid}] {desc}\n    {detail}\n    → {'PASS' if ok else 'FAIL'}")
        return out

    # ── C1 world_navigation 不推 reveal ──────────────────────────────────────
    step_check("C1", "world_navigation 不推 reveal",
               "只移動到B2通訊室方向", expect_class="world_navigation")

    # ── C3 npc_fact_query 不推 reveal ────────────────────────────────────────
    step_check("C3", "npc_fact_query 不推 reveal",
               "我根據他說的機房線索找路，去那個方向", expect_class="npc_fact_query")

    # ── C4 object_inspection 不推 reveal（loop）+ truth-bearing 例外（gate 單元）──
    step_check("C4a", "object_inspection 不推 reveal",
               "我檢查桌上的東西", expect_class="object_inspection")
    g = TruthEvidenceGate()
    tb_allow = g.evaluate(OBJECT_INSPECTION, truth_bearing=True)[0]
    tb_block = g.evaluate(OBJECT_INSPECTION, truth_bearing=False)[0]
    ok4b = (tb_allow is True) and (tb_block is False)
    results.append(("C4b", "object_inspection 例外：truth-bearing 物件才 allow",
                    ok4b, f"truth_bearing→{tb_allow}  plain→{tb_block}"))
    P(f"\n[C4b] truth-bearing 例外：allow={tb_allow} plain={tb_block} → {'PASS' if ok4b else 'FAIL'}")

    # ── C5 truth_investigation 仍可推 reveal（注入 truth-mapped clue 確定性驗證）─
    from core.progress_models import LedgerEntry
    tid = next(t for t in loop._reveal_ledger.truths if t not in loop._core_truth_ids)
    lvl_before = loop._reveal_ledger.truths[tid].level
    loop._game_state.clues["clue.contract_c5"] = LedgerEntry(
        id="clue.contract_c5", title="實驗紀錄", content="", source_event="contract_regression",
        first_seen_beat=loop._game_state.beat_number, truth_id=tid,
        evidence_strength=0.95, max_level="observed")
    out5 = loop.step("我坐下來仔細研究這些實驗紀錄，比對異常頻率的數據")
    lvl_after = loop._reveal_ledger.truths[tid].level
    ok5 = (out5.get("reveal_gate_allowed") is True) and (lvl_after != lvl_before)
    det5 = (f"action_class={out5.get('action_class')} gate_allowed={out5.get('reveal_gate_allowed')} "
            f"truth[{tid[:14]}] {lvl_before}→{lvl_after}")
    results.append(("C5", "truth_investigation 仍可推 reveal", ok5, det5))
    P(f"\n[C5] truth_investigation 推 reveal\n    {det5}\n    → {'PASS' if ok5 else 'FAIL'}")

    # ── C6 structured evidence_events 仍可推 reveal（NPC 結構化通道）────────────
    from core.narrative.npc_chat_control import NPCChatResponse
    loop._update_npc_truth_refs()
    allowed = ((loop.bb.game_meta or {}).get("npc_allowed_truth_refs") or {}).get("allowed_truth_refs", [])
    ok6, det6 = False, "no allowed_truth_refs（無法驗證 structured 路徑）"
    if allowed:
        ref = allowed[0]
        tid6 = ref["truth_id"]
        lvl6_before = loop._reveal_ledger.truths[tid6].level
        resp = NPCChatResponse(
            visible_reply="（壓低聲音）我親眼看到的，不會錯。", answer_status="actionable",
            evidence_events=[{"evidence_id": "ev.contract_c6", "truth_id": tid6,
                              "evidence_strength": 0.95, "max_level": ref.get("max_level", "hinted"),
                              "surface_text": "NPC 提供的結構化證據"}])
        updates = loop.bridge_npc_evidence(resp, npc_id="魏博明")
        lvl6_after = loop._reveal_ledger.truths[tid6].level
        ok6 = bool(updates) or (lvl6_after != lvl6_before)
        det6 = f"truth[{tid6[:14]}] {lvl6_before}→{lvl6_after} updates={updates}"
    results.append(("C6", "structured evidence_events 仍可推 reveal", ok6, det6))
    P(f"\n[C6] structured evidence 推 reveal\n    {det6}\n    → {'PASS' if ok6 else 'FAIL'}")

    # ── C2 + C7 review_mode 不推 reveal、不產生未記帳 fact（最後做，會上鎖）─────
    before = _rp(loop)
    facts_before = {e.id for e in loop._world.by_kind(FACT)}
    out2 = loop.step("我退到外面，在安全區整理已知資訊，不新增調查，不碰真相")
    after = _rp(loop)
    delta2 = _reveal_sum(after) - _reveal_sum(before)
    facts_after = {e.id for e in loop._world.by_kind(FACT)}
    new_facts = facts_after - facts_before
    ok2 = (delta2 == 0) and (out2.get("reveal_gate_allowed") is False)
    ok7 = (len(new_facts) == 0)
    _mode2 = (out2.get("world_progress") or {}).get("investigation_state", "?")
    results.append(("C2", "world_review / review_mode 不推 reveal", ok2,
                    f"mode={_mode2} gate_allowed={out2.get('reveal_gate_allowed')} "
                    f"reveal_delta={delta2}"))
    results.append(("C7", "review_mode 不產生未記帳 fact", ok7,
                    f"new_fact_entities={list(new_facts)}"))
    P(f"\n[C2] review_mode 不推 reveal → {'PASS' if ok2 else 'FAIL'}")
    P(f"[C7] review_mode 不產生未記帳 fact → {'PASS' if ok7 else 'FAIL'}")

    # ── 總結 ─────────────────────────────────────────────────────────────────
    P("\n" + "#" * 72)
    P("【Contract Regression 總結】")
    allok = True
    for cid, desc, ok, detail in results:
        allok = allok and ok
        P(f"  {'✅' if ok else '❌'} {cid:4} {desc}")
    P("#" * 72)
    P("RESULT: " + ("ALL PASS ✅" if allok else "SOME FAILED ❌"))
    out_path = ROOT / "dev" / "reports" / "contract-regression.json"
    out_path.write_text(json.dumps(
        [{"id": c, "desc": d, "pass": o, "detail": x} for c, d, o, x in results],
        ensure_ascii=False, indent=2), encoding="utf-8")
    P(f"[json → {out_path}]")
    return 0 if allok else 1


if __name__ == "__main__":
    sys.exit(main())
