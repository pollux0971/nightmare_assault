---
id: NC7
stage: N
lane: E
title: Config Center Hooks（敘事控制參數可配置）
status: done
worktree: -
depends_on: [NC6, P1]
contracts: [ConfigSchema, NarrativeContract]
last_good_snapshot: NC7__post__20260604-142759
owner_session: -
---

# NC7 · Config Center Hooks

- **階段**：N（敘事控制 · patch v0.1）　**工線**：E（Frontend / pywebview）
- **依賴**：NC6, P1(配置表)　**契約**：ConfigSchema, NarrativeContract
- **備註**：optional（risk=optional）；apply order 最後一站。

## 目標 / 範圍
把敘事控制參數接到配置中心，未來可視化調整：opening_budget / reveal_policy / forbidden_motifs /
story_agent_element_limit / ending_gate_thresholds。**safe profile 為預設**。

## 對應來源
patch `task_cards/P7_config_hooks.md`、`docs/01-stage-roadmap §P7`、`docs/11-context-budget-policy.md`。

## 實作步驟
- 把上述參數接進 config table（沿用階段 P 的 `agent_configs`/`feature_flags`/新 narrative_control 設定列）。
- forbidden_motifs 可接 prompt fragment 或 config；其餘接 config。
- 配置中心 UI 可查看/調整（draft→preview→activate 沿用 P5）；預設仍用 safe profile。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_narrative_config_hooks.py：opening_budget/reveal_policy/forbidden_motifs/element_limit/ending_thresholds 可由 config 讀寫；safe profile 為預設；未設時走安全值。
- [ ] JS 過；前端可查看/調整；全回歸綠。

## 回滾備註
additive config 接點 + 前端（draft 不影響 active）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC7 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC7 post --verify pass` → 回填 last_good_snapshot
