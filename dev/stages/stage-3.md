# 階段 3 · 核心 agent 真接

- **進入條件**：階段2
- **目標**：setup/orchestrator/warden 真實接入並各自可驗。
- **整合驗收（階段 done 門檻）**：setup 生雙層 bible+scene；orchestrator 條件揭露；warden 硬規則優先有效。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U09` | D | setup agent | U03, U04, U07 |
| `U10` | D | orchestrator 揭露閘門 | U03, U04, U07, U11 |
| `U12` | D | warden（本地硬規則優先） | U03, U07 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
