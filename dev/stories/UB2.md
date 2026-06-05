---
id: UB2
stage: B
lane: D
title: 一種結局序列
status: done
worktree: -
depends_on: [U12, U13]
contracts: [EndingConditions]
last_good_snapshot: UB2__post__20260603-204153
owner_session: -
---

# UB2 · 一種結局序列

- **階段**：B（MVP-B（A 穩後））　**工線**：D（Agent Logic（setup/orchestrator/warden/story/compactor/event））
- **依賴**：U12, U13　**契約**：EndingConditions

## 目標 / 範圍
至少一種結局可達（純敘述收尾 beat），不做多結局；提早觸發劇情內化解；揭露完整 truth + 復盤。

## 對應設計章節（引用，不複製；契約見 dev/CONTRACTS.md）
02 §（結局）、00 §六 B8、build/BUILD MVP-B

## 實作步驟
- ending_conditions + 偵測
- 收尾 beat 生成
- 提早觸發內化解

## 驗收（可執行 — pass 才算 done）
- [ ] 觸發後純敘述收尾、揭露完整 real truth、回選單；提早觸發劇情內合理化解

## 回滾備註
依賴 U12/U13；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot UB2 pre`
- 完成：驗收 pass → `snapshot.py snapshot UB2 post --verify pass` → 回填 last_good_snapshot
