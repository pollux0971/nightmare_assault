---
id: NC1
stage: N
lane: D
title: Narrative Contract（setup 產生敘事生成契約）
status: done
worktree: -
depends_on: [NC0, U09]
contracts: [NarrativeContract]
last_good_snapshot: NC1__post__20260604-141515
owner_session: -
---

# NC1 · Narrative Contract

- **階段**：N（敘事控制 · patch v0.1）　**工線**：D（Agent Logic）
- **依賴**：NC0, U09(setup)　**契約**：NarrativeContract

## 目標 / 範圍
讓 setup 不只產世界觀，而是產**故事生成契約** `NarrativeContract`：protagonist_motive / central_question /
motif_palette / opening_budget / reveal_ladder / forbidden_or_limited_motifs / ending_attractors。
setup 只產契約與候選元素，**不直接生成長開場**。

## 對應來源
patch `task_cards/P1_narrative_contract.md`、`docs/03-narrative-contract-spec.md`、`reference_code/models.py`、`examples/narrative_contract.example.json`。

## 實作步驟
- 依 reference_code/models.py 在 core 新增 NarrativeContract 系列（ProtagonistMotive/MotifPalette/OpeningBudget/RevealLevel）。
- setup 輸出時組裝 NarrativeContract（程式碼/LLM 皆可，先 mock-first）；存進 blackboard（optional 欄位）。
- Story Agent / OpeningDirector 改讀 NarrativeContract，不直接讀完整 setup 世界觀。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_narrative_contract.py：每次 setup output 都有明確 protagonist_motive 與 central_question；opening_budget 可被後續模組讀取；reveal_ladder/forbidden_motifs 結構正確。
- [ ] NarrativeContract 為 optional 欄位、不破壞既有 SetupOutput；全回歸綠。

## 回滾備註
additive setup 輸出 + 新模型；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC1 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC1 post --verify pass` → 回填 last_good_snapshot
