---
id: NC0
stage: N
lane: A
title: 契約凍結 + 旁路接點（feature flag，預設 OFF）
status: done
worktree: -
depends_on: []
contracts: [NarrativeContract, RevealLadder]
last_good_snapshot: NC0__post__20260604-141235
owner_session: -
---

# NC0 · 契約凍結 + 旁路接點

- **階段**：N（敘事控制 · patch v0.1）　**工線**：A（整合者 / Tech Lead）
- **依賴**：無（進入條件：MVP-B UB1–UB7 done）　**契約**：NarrativeContract, RevealLadder

## 目標 / 範圍
不動 runtime 大架構，只新增旁路規格與接點。找出入口、標記餵 story 的資料、加 feature flag（不改預設行為）。

## 對應來源
patch `task_cards/P0_contract_freeze.md`、`docs/01-stage-roadmap §P0`、`docs/08-integration-guide`、`CLAUDE_CODE_START_HERE.md`。

## 實作步驟
- 盤點入口：`start_game` / beat loop（`core/orchestrator_loop.py`）/ story_agent_call（`run_story`）/ ending_check（`_finalize_ending`/attractor）。
- 標記哪些資料會傳給 Story Agent（對照 ContextBuilder / build_opening_context）。
- 加 flag `ENABLE_NARRATIVE_CONTROL`（預設 OFF；`.env > config`）；建 docs 中 Narrative Contract schema 對照（dev/CONTRACTS.md §十二）。
- 既有 beat loop API / ProgressKernel 輸出 / Story 輸出格式凍結；新欄位一律 optional。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_narrative_flags.py：`ENABLE_NARRATIVE_CONTROL` 預設 OFF → 行為與現況一致；可由 env/config 開關。
- [ ] 既有全套件綠（未改預設行為）；新增欄位皆 optional、不破壞既有存檔；board --check 契約解析。

## 回滾備註
純旁路規格 + flag（預設 OFF）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC0 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC0 post --verify pass` → 回填 last_good_snapshot
