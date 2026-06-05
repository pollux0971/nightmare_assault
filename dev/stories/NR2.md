---
id: NR2
stage: R
lane: D
title: Answer Debt（重複提問答債，逼出部分答/具體線索）
status: done
worktree: main
depends_on: [NR1, U13]
contracts: [AnswerDebt, NPCChatControl]
last_good_snapshot: NR2__post__20260605-010853
owner_session: session-24-opus
---

# NR2 · Answer Debt — 重複提問答債

- **階段**：R　**工線**：D（agent）
- **依賴**：NR1（NPCChat 控制 + context）、U13（story agent）
- **契約**：AnswerDebt, NPCChatControl（見 §十四）

## 目標 / 範圍
玩家反覆直問時，系統常以氛圍敷衍（卡關感）。引入答債：追蹤已問問題與是否已用有用資訊償還，debt≥2 時強制給部分答/具體線索/指向證據/具理由拒答。

## 對應來源
patch `task_cards/P2_AnswerDebt.md`、`docs/04-answer-debt-policy.md`、`reference_code/answer_debt.py`。

## 實作步驟
- 問題分類：identity / mechanism / threat / location / action（規則版分類器，含「432.7 是什麼」「林晨在哪」等樣式）。
- 依 topic 追蹤答債等級：0 無 / 1 可迴避 / 2 須部分答 / 3 須具體線索或具理由拒答。
- 把 answer_debt 注入 StoryAgent 與 NPC-chat context（接 NR1）。
- 強制：debt≥2 的下一相關回應**至少**含 partial answer / concrete clue / direction to evidence / explicit refusal with reason 之一。
- 償還的具體線索盡量轉成 EvidenceEvent（接 NR0），讓「逼問」也能推進真相進度。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_answer_debt_nr2.py：同問題問 2 次 → 第三次回應不得純氛圍，須含部分答/線索/具理由拒答；debt 計數依 topic 正確累加/重置。
- [ ] story 與 npc-chat 皆吃 answer_debt；flag OFF 退回現況；既有套件綠。

## 回滾備註
分類器 + debt ledger + context 欄位；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR2 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR2 post --verify pass` → 回填 last_good_snapshot
