# 階段 P · 配置中心 / Story Agent 模組化（patch v1.1）

- **進入條件**：MVP-A（U00–U20）+ 階段 S（SK01–SK13）全 done。
- **目標**：讓 Story Agent 變**可配置、模組化**而不破壞 MVP-A——prompt 拆成 fragment 組裝、可預覽、可快照、可回滾；ProgressKernel 仍是 world-state 真相，story 只 realize；static prompt 永遠保留 fallback。docs-first、分階段、不一次大改。
- **整合驗收**：開門 anti-repetition 過、story 無 real_bible、prompt_hash 決定性、new clue 進 beat、free input 保留、每 run 存 config 快照、ON/OFF 都不崩。

## 套用順序（patch apply order）

```text
P0 → P1 → P2 → P3 → P4 → P6 → P5 → P7
```

> P5 UI 刻意排在 P6 之後（先快照/測試再可視化編輯）；P7（其他 agent）為最後一站「Later」。

## 本階段工單（序列為主，依工線並行）

| 工單 | 工線 | 內容 | depends_on |
|---|---|---|---|
| `P0` | A | 配置中心契約凍結（名稱/profile/fragment key/flags/fallback） | - |
| `P1` | B | 配置表（additive SQLite + 種子預設） | P0, U05 |
| `P2` | C | PromptComposer（決定性 compiled prompt + hash + preview） | P1 |
| `P3` | D | Story Agent fragment 化（kernel obedience + 反重複 + 自由選擇） | P2, U13 |
| `P4` | D | Runtime 整合（config-first + static fallback） | P3, SK04 |
| `P6` | F | Run 配置快照 + prompt 回歸測試 | P4 |
| `P5` | E | 配置 UI（draft→preview→activate） | P6, U18 |
| `P7` | D | 擴展至其他 agent（warden/orchestrator/compactor/setup） | P6, P5 |

> 範圍切分：**P0–P6 = MVP-A 收尾核心**（修好「開門後還問開門」的 prompt 面、讓 story 可配置可 debug）；**P5/P7 = 收尾 + 擴展尾段**（UI 與其他 agent 遷移）。
> docs-first：canonical 規格在 `nightmare-assault-story-doc-patch-batch-v1.1/`（docs + task_cards）；`reference/`、`prompts_reference/` 只是範例，不照抄。
> 鐵律：additive migration（不破壞既有存檔）｜PromptComposer 決定性（同輸入同 hash）｜story 永不見 real_bible｜前端不解析分隔符｜config 失敗一律退 static fallback。
> 契約見 dev/CONTRACTS.md §十一；執行機制見 dev/PARALLEL-PLAN.md §四；工線角色見 dev/WORKFLOW.md。
