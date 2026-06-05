---
id: UB5
stage: B
lane: B
title: 道具庫完整
status: done
worktree: -
depends_on: [U04, U05]
contracts: [SharedInventory]
last_good_snapshot: UB5__post__20260603-202742
owner_session: -
---

# UB5 · 道具庫完整

- **階段**：B（MVP-B（A 穩後））　**工線**：B（Core State / Persistence（models/blackboard/scene/db））
- **依賴**：U04, U05　**契約**：SharedInventory

## 目標 / 範圍
item 型碎片流入道具庫、held_by 綁 NPC、/inventory 查看、隨快照保存。

## 對應設計章節（引用，不複製；契約見 dev/CONTRACTS.md）
03 §二、00 道具庫決策

## 實作步驟
- item 結構 + 增刪查
- held_by 轉移
- 隨 beat 快照保存

## 驗收（可執行 — pass 才算 done）
- [ ] item 增刪查、held_by 轉移、不洩漏 is_key_item、隨快照保存

## 回滾備註
依賴 U04/U05；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot UB5 pre`
- 完成：驗收 pass → `snapshot.py snapshot UB5 post --verify pass` → 回填 last_good_snapshot
