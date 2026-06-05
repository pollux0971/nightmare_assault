---
id: UB1
stage: B
lane: D
title: 技能宣稱封頂強化
status: done
worktree: -
depends_on: [U12]
contracts: [WardenOutput, Ledger]
last_good_snapshot: UB1__post__20260603-203259
owner_session: -
---

# UB1 · 技能宣稱封頂強化

- **階段**：B（MVP-B（A 穩後））　**工線**：D（Agent Logic（setup/orchestrator/warden/story/compactor/event））
- **依賴**：U12　**契約**：WardenOutput, Ledger

## 目標 / 範圍
技能宣稱封頂更完整，侷限具體且接劇情（能變謎題/線索），寫 (技能,侷限) ledger。

## 對應設計章節（引用，不複製；契約見 dev/CONTRACTS.md）
02 §六、CHECKLIST C10

## 實作步驟
- 強化封頂判斷
- 侷限生成接劇情
- ledger 二元組

## 驗收（可執行 — pass 才算 done）
- [ ] 誇張宣稱被封頂、侷限變謎題/線索（C10）；ledger 記錄；UI 提示侷限

## 回滾備註
依賴 U12；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot UB1 pre`
- 完成：驗收 pass → `snapshot.py snapshot UB1 post --verify pass` → 回填 last_good_snapshot
