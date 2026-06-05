# 階段 6 · 前端 + MVP-A 打磨

- **進入條件**：階段5
- **目標**：pywebview+串流渲染+主畫面+存檔/道具，端到端打通並打磨。
- **整合驗收（階段 done 門檻）**：**MVP-A 驗收**：30 beat 不崩、防暴雷、JSON≥95% 可 repair、回溯正確、不跑版。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U16` | E | pywebview 骨架 + API | U15 |
| `U17` | E | 串流渲染（前端承重牆） | U08, U16 |
| `U18` | E | 主畫面 + 前置畫面（keyring） | U16, U17 |
| `U19` | E | 存檔 UI + 道具面板 | U05, U18 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
