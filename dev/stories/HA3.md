---
id: HA3
stage: H
lane: F
title: Hidden Recap Masking（observation 不洩 hidden truth content）
status: done
worktree: main
depends_on: [NR0, UB7]
contracts: [HiddenRecapMask]
last_good_snapshot: HA3__post__20260605-101949
owner_session: session-24-opus
---

# HA3 · Hidden Recap Masking

- **階段**：H　**工線**：F（反劇透安全）
- **依賴**：NR0（RevealLedger/recap）、UB7（masked 結局）
- **契約**：HiddenRecapMask（見 §十五）

## 目標 / 範圍
玩家可見輸出（observation / API / agent_play 預設）**不得含 hidden truth content**。recap 對玩家只給 found（已達等級者）+ hidden_count + hidden_titles（遮罩標題）；full hidden recap 僅 explicit debug flag 可開。

## 對應來源
patch `task_cards/P0_A3`、`docs/07`。

## 實作步驟
- `recap_from_ledger`（或新 `public_recap`）：玩家面版本移除 hidden_list 的 content，只留 hidden_count + 遮罩標題（"未解的線索 #N"）。
- `agent_play`：observation 預設用 public recap；加 `--debug-reveal-truth` 才給 full（並標記）。
- 確認 ending masked render 已遵守（UB7）；observation 的 ending dict 不夾 hidden content。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_recap_mask_ha3.py：public recap JSON 序列化後**不含任何 hidden 碎片 content 子字串**；hidden_count 正確；debug flag 才露 full。
- [ ] agent_play observation（無 debug flag）grep 不到 hidden truth content；既有套件綠。

## 回滾備註
recap 遮罩 + flag；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HA3 pre｜完成：snapshot.py snapshot HA3 post --verify pass
