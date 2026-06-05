# 階段 1 · 核心資料層（無 LLM）

- **進入條件**：階段0
- **目標**：models/blackboard/scene/sqlite/constants/signalbus 用假資料測穩，打地基。
- **整合驗收（階段 done 門檻）**：models 可驗證；patch+version 可運作；SQLite 存假 beat 讀回；signalbus 收發。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U01` | B | Pydantic 資料類（core/models.py） | U00 |
| `U02` | C | 常數 + SignalBus | U00 |
| `U03` | B | Blackboard + 版本/patch（並行控制） | U01 |
| `U04` | B | 場景系統（SceneRegistry） | U01, U03 |
| `U05` | B | SQLite 持久化 | U01, U03 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
