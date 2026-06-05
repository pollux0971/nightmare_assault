---
id: HC1
stage: H
lane: D
title: NPCChat Structured Gate（結構化輸出 + 閘門 + repair + bridge）
status: done
worktree: main
depends_on: [NR1, HB2]
contracts: [NPCChatResponse, EvidenceEvent]
last_good_snapshot: HC1__post__20260605-102618
owner_session: session-24-opus
---

# HC1 · NPCChat Structured Gate

- **階段**：H　**工線**：D（agent）
- **依賴**：NR1（npc-chat 控制 context）、HB2（統一 bridge）
- **契約**：NPCChatResponse（結構化）, EvidenceEvent（見 §十五）

## 目標 / 範圍
run_npc_chat 從「純文字裸輸出」改為**結構化** NPCChatResponse，並接 NPCChatControlGate：違規 new lore → repair once → 仍違規用 safe_fallback_reply；evidence_events 經 cap_to_ceiling → RevelationBridge（不再只靠 keyword scan）。修補 NR1 的軟點（NPC 走同一 bridge 但不靠掃字）。

## 對應來源
patch `task_cards/P0_C1`、`docs/05`、`reference_code/npc_chat_gate.py`、`prompts_reference/npc_chat_structured_prompt.md`。

## 實作步驟
- 定義 `NPCChatResponse`（reply/answer_status/evidence_events/new_lore_terms/used_truth_ids/blocked_or_uncertain_claims）。
- run_npc_chat（flag on）：要 LLM 回結構化（StreamParser/JSON）→ 解析；解析失敗 fallback 純文字（向後相容）。
- 過 `NPCChatControlGate.validate` → invalid `repair once` → 仍 invalid `safe_fallback_reply`。
- evidence_events → cap_evidence_to_ceiling → `loop.bridge_npc_evidence`（改吃結構化 evidence，不再只關鍵詞掃）。
- flag OFF / 無 contract → 與 MC1 純文字行為一致。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_npc_structured_hc1.py：合法結構化回覆通過、evidence 推進 ledger；違規 new lore → repair；repair 仍違規 → safe fallback（actionable、不新增 lore）；解析失敗 → 純文字 fallback 不崩。
- [ ] 既有 npc-chat 測試綠；board --check 0 errors。

## 回滾備註
結構化輸出 + gate + repair；解析失敗保底純文字；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HC1 pre｜完成：snapshot.py snapshot HC1 post --verify pass
