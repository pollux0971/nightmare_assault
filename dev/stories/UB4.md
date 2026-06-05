---
id: UB4
stage: B
lane: D
title: 輕量 dreaming（在場 NPC）
status: done
worktree: -
depends_on: [U03, U07]
contracts: [DreamingOutput, NPCRegistry]
last_good_snapshot: UB4__post__20260603-203728
owner_session: -
---

# UB4 · 輕量 dreaming（在場 NPC）

- **階段**：B（MVP-B（A 穩後））　**工線**：D（Agent Logic（setup/orchestrator/warden/story/compactor/event））
- **依賴**：U03, U07　**契約**：DreamingOutput, NPCRegistry

## 目標 / 範圍
在場 active NPC 每 5 beat 更新 evolving；只寫 npc_evolving；self_aware=false 不編謊；只跑在場。非同步只產 patch。

## 對應設計章節（引用，不複製；契約見 dev/CONTRACTS.md）
02 §七、07 §二、CHECKLIST C5/C6/C7、05 待決2

## 實作步驟
- 每 5 beat 在場 active 更新 evolving
- 權限只寫 evolving（C6）
- self_aware=false 不產 emergent_lie（C5）

## 驗收（可執行 — pass 才算 done）
- [ ] 在場 NPC 情緒演化、非同步只產 patch、沒戲份凍結（C7）；self_aware=false 不編謊（C5）

## 回滾備註
依賴 U03/U07；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot UB4 pre`
- 完成：驗收 pass → `snapshot.py snapshot UB4 post --verify pass` → 回填 last_good_snapshot
