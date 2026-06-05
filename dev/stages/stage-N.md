# 階段 N · 敘事控制（Narrative Control patch v0.1）

- **進入條件**：MVP-B（UB1–UB7）done。
- **目標**：補一個**敘事控制層**——開場只放少量高價值元素（建立動機，不展示世界觀）、真相分層揭露、Story Agent 不發明世界觀、結局有因果門檻。不大重構 beat loop。
- **整合驗收**：開場主要新元素 ≤3 / 至少 1 動機 + 1 可行動線索 / 開場最多 hinted / 真相不跳級 / story 不新增未授權元素 / 0/8 不可 clean escape / 全程 flag OFF 可退回現況。

## 套用順序（patch apply order）

```text
NC0 → NC1 → NC2 → NC3 → NC4 → NC5 → NC6 → NC7
```

> 先 NC0–NC3（契約/開場/降權），再 NC4–NC6（揭露/品質/結局），NC7（config）為 optional 收尾。

## 本階段工單

| 工單 | 工線 | 內容 | depends_on |
|---|---|---|---|
| `NC0` | A | 契約凍結 + 旁路接點（flag 預設 OFF） | - |
| `NC1` | D | Narrative Contract（setup 產生敘事生成契約） | NC0, U09 |
| `NC2` | D | Opening Director（限制開場元素 ≤budget；refine UB6） | NC1, UB6 |
| `NC3` | D | Story Agent Downgrade（只執行 blueprint） | NC1, U13 |
| `NC4` | D | Reveal Ladder（真相分層，不跳級） | NC1, U10 |
| `NC5` | F | Quality Gate（輸出前檢查 + repair/fallback） | NC3, NC4 |
| `NC6` | D | Ending Gate（因果門檻；0/8 不可 clean escape；refine UB7） | NC4, UB7 |
| `NC7` | E | Config Center Hooks（參數可配置，optional） | NC6, P1 |

> 核心原則（docs/00）：A 開頭四件事就夠｜B 每 beat 一個主要敘事目的｜C 真相有階梯不跳級｜D Story Agent 不發明世界觀。
> 旁路紀律：`ENABLE_NARRATIVE_CONTROL` 預設 OFF；舊流程可 fallback；新欄位 optional；**story 永不見 real_bible 不變**。
> docs-first：canonical 在 `nightmare-assault-mvp-b-narrative-control-patch-v0.1/`（docs/task_cards/reference_code/examples）；reference_code 只是範例，依現有結構改寫、不照抄。
> 契約見 dev/CONTRACTS.md §十二；執行機制見 dev/PARALLEL-PLAN.md §四。
