---
id: NR1
stage: R
lane: D
title: NPCChat 敘事控制（套用同一敘事契約 + 證據橋接）
status: done
worktree: main
depends_on: [NR0, MC1, NC4]
contracts: [NPCChatControl, NarrativeContract, RevealLadder]
last_good_snapshot: NR1__post__20260605-010650
owner_session: session-24-opus
---

# NR1 · NPCChat 敘事控制

- **階段**：R　**工線**：D（agent）
- **依賴**：NR0（RevelationBridge）、MC1（npc-chat agent）、NC4（RevealLadder）
- **契約**：NPCChatControl, NarrativeContract, RevealLadder（見 §十四 / §十二）

## 目標 / 範圍
v0.1 只約束了主 Story Agent，**npc-chat 仍可無控擴張世界觀**（新組織/協定/機制/怪物、真相跳級），造成母題過載與暴雷風險。讓 npc-chat 收同一份敘事契約並把有用提示轉成受控 EvidenceEvent。

## 對應來源
patch `task_cards/P1_NPCChatNarrativeControl.md`、`docs/03-npc-chat-narrative-control.md`、`reference_code/npc_chat_control.py`、`prompts_reference/npc_chat_delta.md`、`examples/npc_chat_control_context.example.json`。

## 實作步驟
- `NPCChatContextBuilder`：注入 motif_palette / forbidden_motifs / truth_reveal_ladder / current_reveal_levels / active_player_motive / answer_debt / context_budget（不洩 real_bible/secret_core，沿用 MC1 認知卡白名單）。
- npc-chat 輸出改為 `NPCChatResponse`：visible_reply / npc_emotion_delta / answer_status(answered|partial|evaded|refused) / evidence_events[] / new_lore_terms[] / used_motifs[] / quality_flags[]。
- NPC 可：表達情緒 / 迴避 / 部分回答 / 說謊（只准映射到既有 truth 或動機）/ 透過受控 EvidenceEvent 露線索。
- NPC 不可：建新組織/協定/核心機制/怪物類型；不可由 hidden 直跳 confirmed。
- 硬護欄：出現 motif_palette / allowed_new_terms 之外的新 lore term → 走 repair（重寫移除新 lore、保留情緒與部分答、把有用提示轉 EvidenceEvent）。
- 退出時：evidence_events 餵 NR0 RevelationBridge；沿用 MC2 濃縮回 story hot context。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_npc_chat_control_nr1.py：npc-chat **無法**引入 motif_palette/reveal_ladder 外的重大新世界概念（越界→repair / quality_flag）；可提供有用 hint 但不發明無關機制。
- [ ] 玩家相關提問可得部分答或 evidence_event，且 evidence 進 RevealLedger（接 NR0）。
- [ ] 結構性防暴雷：context 無 real_bible/secret_core；flag OFF 退回 MC1 行為；既有套件綠。

## 回滾備註
context builder + 輸出契約擴充 + repair；新欄位 optional；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR1 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR1 post --verify pass` → 回填 last_good_snapshot
