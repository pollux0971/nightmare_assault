# 階段 4 · story + compactor

- **進入條件**：階段3
- **目標**：story 串流防暴雷；compactor 撐 30 beat。
- **整合驗收（階段 done 門檻）**：story 停決策點、不暴雷；compactor 30 假 beat 不爆、伏筆留、回溯摘要正確。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U13` | D | story agent + 串流（防暴雷） | U07, U08, U10 |
| `U14` | D | compactor + 30 beat（承重牆） | U03, U07 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
