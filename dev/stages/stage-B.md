# 階段 B · MVP-B（A 穩後）

- **進入條件**：MVP-A 驗收通過
- **目標**：技能封頂強化、一種結局、防暴雷報告、輕量 dreaming、道具庫完整。
- **整合驗收（階段 done 門檻）**：差異化展示成立。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `UB1` | D | 技能宣稱封頂強化 | U12 |
| `UB2` | D | 一種結局序列 | U12, U13 |
| `UB3` | F | 雙層防暴雷驗證報告 | U13 |
| `UB4` | D | 輕量 dreaming（在場 NPC） | U03, U07 |
| `UB5` | B | 道具庫完整 | U04, U05 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
