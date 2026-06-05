# 階段 5 · beat 主迴圈

- **進入條件**：階段4
- **目標**：把 warden→orchestrator→story→快照→安全點 merge→compactor 串成迴圈。
- **整合驗收（階段 done 門檻）**：連續多 beat 狀態一致；非同步 patch 不污染 story 讀取。後端 MVP-A 核心通。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U15` | A | beat 主迴圈 + 安全點 merge | U05, U10, U12, U13, U14 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
