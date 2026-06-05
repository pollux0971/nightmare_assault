---
id: NC2
stage: N
lane: D
title: Opening Director（限制開場元素數量）
status: done
worktree: -
depends_on: [NC1, UB6]
contracts: [OpeningBlueprint, NarrativeContract]
last_good_snapshot: NC2__post__20260604-141724
owner_session: -
---

# NC2 · Opening Director

- **階段**：N（敘事控制 · patch v0.1）　**工線**：D（Agent Logic）
- **依賴**：NC1, UB6(序幕鉤子)　**契約**：OpeningBlueprint, NarrativeContract

## 目標 / 範圍
解決開場元素過多（原則 A：開頭四件事就夠）。新增 `OpeningDirector`，從 NarrativeContract 挑**少量**開場元素，
輸出 `OpeningBlueprint`。**refine（非取代）UB6**：UB6 的開場義務保留，但元素數量由 budget 控制、揭露上限到 hinted。

## 對應來源
patch `task_cards/P2_opening_director.md`、`docs/04-opening-director-spec.md`、`reference_code/opening_director.py`、`examples/opening_blueprint.example.json`。

## 實作步驟
- 建 OpeningBlueprint schema：allowed_elements / blocked_elements / motive_evidence / opening_reveal_limit / first_choice_policy。
- OpeningDirector 從 contract 挑 allowed_elements（≤ opening_budget）、產 blocked_elements、設 truth seed reveal level（≤ hinted）。
- 接進 `_kernel_intro_beat`：把 blueprint（只給表層、無 real_bible）餵 story；UB6 的 5 seed 改由 director 依 budget 篩選。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_opening_director.py：開場主要新元素 ≤3、新 lore terms ≤3；至少 1 個玩家動機 + 1 個可行動線索；開場 reveal level ≤ hinted（**不得 confirmed**）；blueprint 不含 real_bible / hidden_truth（防暴雷）。
- [ ] flag OFF 時走 UB6 原路；flag ON 走 director；全回歸綠。

## 回滾備註
旁路 director（flag 控制）；UB6 路徑保留；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC2 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC2 post --verify pass` → 回填 last_good_snapshot
