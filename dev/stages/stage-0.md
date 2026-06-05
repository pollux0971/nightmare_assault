# 階段 0 · 契約凍結 + 地基

- **進入條件**：無
- **目標**：確認設計與 CONTRACTS 一致、建專案地基，讓後續工單可測。
- **整合驗收（階段 done 門檻）**：契約無矛盾；pytest 綠燈；import core OK。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U00` | A | 契約統一檢查 + 專案地基 | - |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
