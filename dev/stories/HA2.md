---
id: HA2
stage: H
lane: D
title: Death Causality Guard（danger 達標 ≠ death）
status: done
worktree: main
depends_on: [U12, SK13]
contracts: [EndingCausalityGate]
last_good_snapshot: HA2__post__20260605-101807
owner_session: session-24-opus
---

# HA2 · Death Causality Guard

- **階段**：H　**工線**：D（gate / attractor）
- **依賴**：U12（warden）、SK13（attractor ending）
- **契約**：EndingCausalityGate（見 §十五）

## 目標 / 範圍
`death_physical` 不得僅因 danger threshold 觸發。新增 EndingCausalityGate：死亡必須有 warden hard_trigger==death 或 progress.death_cause_event 或明確致命行為；danger 達標只能降級 danger_warning / injury / failed_escape。

## 對應來源
patch `task_cards/P0_A2`、`docs/03`、`reference_code/ending_gate.py`、`tests_reference/test_death_gate_nr.py`。

## 實作步驟
- 新增 `core/narrative/ending_causality.py`：`EndingCandidate` + `EndingCausalityGate.check_death(candidate, warden_result, progress_result, state)` → GateResult（allowed / downgrade_to）。
- 在 loop 觸發 death 結局前（warden/attractor）過 gate；danger-only → downgrade，不 ended death。
- 旁路：flag-gated（ENABLE_NARRATIVE_CONTROL）；flag OFF 行為不變。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_death_gate_ha2.py：danger 超標但無死亡事件 → gate 不允許 death_physical、downgrade_to=danger_warning；有 hard_trigger/death_cause → 允許。
- [ ] 既有全套件綠；board --check 0 errors。

## 回滾備註
新增 gate 模組 + flag-gated 接點；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HA2 pre｜完成：snapshot.py snapshot HA2 post --verify pass
