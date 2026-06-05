---
id: NR0
stage: R
lane: D
title: RevelationBridge（線索 → 真相進度橋接）
status: done
worktree: main
depends_on: [U10, SK11, UB7]
contracts: [RevelationBridge, EvidenceEvent, RevealLedger, EvidenceMapping]
last_good_snapshot: NR0__post__20260605-022804
owner_session: session-24-opus
---

# NR0 · RevelationBridge — 把玩家可見線索接上官方真相進度

- **階段**：R（敘事控制 v0.2 · 揭露橋接）　**工線**：D（agent / 資料層）
- **依賴**：U10（orchestrator 揭露閘門 / revealed_bible）、SK11（kernel 每行動有後果 / events）、UB7（結局 masked recap）
- **契約**：RevelationBridge, EvidenceEvent, RevealLedger, EvidenceMapping（見 dev/CONTRACTS.md §十四）

## 目標 / 範圍
修補 v0.1 selfplay 暴露的核心斷鏈：玩家調查產生 kernel clues / npc-chat hints，但**沒有轉成 real_bible/revealed_bible 真相進度**，導致結局永遠 `0/X`。建一條受控橋接：

```text
ProgressKernel / NPCChat / Story 輸出 → EvidenceEvent → RevelationBridge → RevealManager → RevealedBible / RevealLedger → Ending recap
```

> 原則（docs/00）：**不是叫 Story Agent 多揭露**，而是由系統決定哪些線索算真相進度。

## 對應來源
patch `task_cards/P0_RevelationBridge.md`、`docs/02-revelation-bridge-spec.md`、`docs/08-integration-guide.md`、`reference_code/revelation_bridge.py`、`examples/evidence_mapping.example.json`、`examples/reveal_ledger.example.json`。

## 實作步驟
- 新增 `EvidenceEvent`（evidence_id / source / player_action / surface_text / truth_id / suggested_reveal_level / evidence_strength / scene_id / beat_number；`atmosphere_only` 旗標）模型（core/models.py 或 core/narrative/）。
- 建 `EvidenceMapping`（clue_id↔truth_id / minimum_action / grant_level / grant_strength）：對應目前 kernel scene graph 的線索（generated + static）。
- 新增 `RevelationBridge.process_evidence_events()`：吃 kernel 結果 + npc-chat 退出 + story 輸出 → 過映射 → 呼 RevealManager（§十二 RevealLadder，不可跳級、需 evidence）→ 寫 RevealLedger / revealed_bible。
- `RevealLedger`：每 truth_id 記當前階梯等級（hidden→hinted→observed→suspected→confirmed→actionable）+ 來源 evidence。
- recap renderer（refine UB7）改讀 RevealLedger：顯示 hinted/observed/suspected/confirmed **部分進度**，不再只算 confirmed。
- beat loop：每次 kernel 結果後、npc-chat 退出後呼叫 bridge（旁路，flag-gated；OFF 退回現況）。

## 護欄（docs/02 §rule）
- 不把每個氛圍細節都當真相線索（`atmosphere_only=true`）。
- 不可無 evidence 直跳 confirmed。
- 結構性防暴雷不變：bridge 只決定「等級」，content 仍只露已達等級的部分；story/npc 永不見 real_bible。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_revelation_bridge_nr0.py：檢查有意義線索（如警示紙條 432.7）過 bridge 後 `truth.signal_frequency >= hinted`；裝飾線索 `atmosphere_only` 不升級；無 evidence 不跳 confirmed。
- [ ] recap 在「玩家檢查過 ≥1 個 mapped 線索」時**不顯示 0/X**（至少 hinted）。
- [ ] flag OFF → 行為與現況一致；既有全套件綠；board --check 0 errors。

## 回滾備註
新增模型/服務 + flag-gated 接點；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR0 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR0 post --verify pass` → 回填 last_good_snapshot
