---
id: UB7
stage: B
lane: D
title: 結局揭露政策（masked：只露已發現碎片）
status: done
worktree: -
depends_on: [UB2]
contracts: [EndingConditions]
last_good_snapshot: UB7__post__20260604-130407
owner_session: -
---

# UB7 · 結局揭露政策（masked）

- **階段**：B（MVP-B 差異化）　**工線**：D（Agent Logic）
- **依賴**：UB2（結局序列）　**契約**：EndingConditions

## 目標 / 範圍
UB2 的結局一次把所有碎片全文攤開（適合 debug，不適合正式體驗）。改為 **masked 揭露政策**：
死亡/結局只揭露玩家**已發現**的碎片全文；**未發現**的只顯示遮罩標題 + `？？？`，並加重玩鉤子。
保留 `full` 模式供 debug。預設 masked。

## 對應來源
使用者設計回饋（patch）：`nightmare-assault-design-fixed/12-mvp-b.md §十`。

## 實作步驟
- `core/agents/ending.py`：
  - 碎片補可遮罩 `title`（無則由 content 衍生短標題）。
  - `render_ending_text(ending, mode="masked")`：masked → 「已確認」(已發現全文) + 「未確認」(遮罩標題 + ？？？) + 重玩鉤子「有些答案你還沒走到它面前。」；`full` → 維持 UB2 全揭（debug）。
  - 預設 mode=masked。`build_ending_sequence` 的資料（found/missed）不變。
- 前端 `onEnding` 用 masked 呈現（已確認/未確認分區）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_ending_mask_ub7.py：masked → 已發現碎片露全文、未發現只露遮罩標題且**不露其全文** + 含重玩鉤子；full → 全揭（含未發現全文）；預設 masked。
- [ ] 既有 UB2 測試調整為對應新政策（render 全揭改 mode="full"）；JS 過；全回歸綠燈。

## 回滾備註
純 render 政策 + 前端分區（資料層不變）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot UB7 pre`
- 完成：驗收 pass → `snapshot.py snapshot UB7 post --verify pass` → 回填 last_good_snapshot
