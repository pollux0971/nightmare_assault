---
id: NC5
stage: N
lane: F
title: Quality Gate（輸出前檢查元素過載 + repair/fallback）
status: done
worktree: -
depends_on: [NC3, NC4]
contracts: [QualityGate]
last_good_snapshot: NC5__post__20260604-142339
owner_session: -
---

# NC5 · Quality Gate

- **階段**：N（敘事控制 · patch v0.1）　**工線**：F（QA / Test）
- **依賴**：NC3, NC4　**契約**：QualityGate

## 目標 / 範圍
在 Story Agent 輸出前/後做**規則版**品質檢查 `NarrativeQualityGate`：元素數量、動機是否清楚、forbidden motifs、
reveal jump、選項是否有意義。違規 → repair prompt 或 fallback。

## 對應來源
patch `task_cards/P5_quality_gate.md`、`docs/06-quality-gate-spec.md`、`reference_code/quality_gate.py`、`tests_reference/test_reference_quality_gate.py`。

## 實作步驟
- 實作 QualityGate（純規則，零 LLM）：count new lore terms / motive present / forbidden motifs / reveal jump / options 有意義。
- 違規 → 組 repair prompt 重生一次；仍違規 → fallback（降級為安全 beat，不崩）。
- 接進 beat 收尾（與 StreamParser/quality 串接，旁路、flag 控制）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_quality_gate.py：元素過載/缺動機/forbidden motif/reveal jump/無意義選項各被偵測；違規輸出可被 repair 或 fallback；合格輸出放行（零 LLM 規則檢查）。
- [ ] gate 失敗不崩遊戲（graceful）；全回歸綠。

## 回滾備註
旁路 gate（flag 控制）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC5 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC5 post --verify pass` → 回填 last_good_snapshot
