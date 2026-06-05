# 階段 2 · LLM 基礎 + agent 外殼

- **進入條件**：階段1
- **目標**：client/SkillCaller/StreamParser/event 抽取就緒（先 mock，再接真 API）。
- **整合驗收（階段 done 門檻）**：LLM 可 call/stream/fallback+trace；parser 三級 repair 過；SkillCaller 載 SKILL。

## 本階段工單（依工線並行，序列依 depends_on）

| 工單 | 工線 | 標題 | depends_on |
|---|---|---|---|
| `U06` | C | OpenRouterClient + fallback + trace | U01, U02, U05 |
| `U07` | C | SkillCaller 基類 + SKILL.md 載入 | U01, U06 |
| `U08` | C | StreamParser 三級 repair（承重牆） | U01 |
| `U11` | D | Event 抽取（程式碼層，非 agent） | U01 |

> 同階段不同工線可並行（git worktree + 子 agent Sonnet）；承重牆 U08/U14 建議升 Opus。
> 執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
