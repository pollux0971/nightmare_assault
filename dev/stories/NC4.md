---
id: NC4
stage: N
lane: D
title: Reveal Ladder（真相分層揭露，不可跳級）
status: done
worktree: -
depends_on: [NC1, U10]
contracts: [RevealLadder]
last_good_snapshot: NC4__post__20260604-142123
owner_session: -
---

# NC4 · Reveal Ladder

- **階段**：N（敘事控制 · patch v0.1）　**工線**：D（Agent Logic）
- **依賴**：NC1, U10(orchestrator)　**契約**：RevealLadder

## 目標 / 範圍
真相揭露分層（原則 C）。新增 `RevealManager`：每個 TruthFragment 有等級
`hidden→hinted→observed→suspected→confirmed→actionable`；不可跳級、需 evidence 才升級、開場最多 hinted。
story context 只傳 allowed reveal level 的內容。

## 對應來源
patch `task_cards/P4_reveal_ladder.md`、`docs/05-reveal-ladder-policy.md`、`reference_code/reveal_manager.py`、`examples/reveal_ladder.example.json`。

## 實作步驟
- TruthFragment 加 reveal_level；RevealManager 維護各碎片等級。
- 每 beat 依玩家取得的 evidence 判斷可否升級（一次一階，不跳級）；與 orchestrator 揭露閘門協調。
- ContextBuilder / OpeningDirector 只把 ≤ allowed level 的碎片內容餵 story。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_reveal_ladder.py：hidden 不可直接 confirmed（一次升一階）；開場最多 hinted；無 evidence 不升級；超過 allowed level 的內容不進 story context。
- [ ] 與既有 orchestrator 揭露不衝突；全回歸綠。

## 回滾備註
旁路 RevealManager（flag 控制）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC4 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC4 post --verify pass` → 回填 last_good_snapshot
