# 階段 S · 穩定化補丁（Narrative Progress Kernel）

- **進入條件**：MVP-A（U00–U19）done
- **目標**：不讓 LLM 決定推進——ProgressKernel+PatchValidator 決定，story 只 realize。修好『開門後還問開門』、NPC 不進場、線索/背包不持久。
- **整合驗收**：T1 開門不重複、T3 每 beat ≥1 delta、T4 NPC 出場、30-beat 指標達標、ON/OFF 都不崩、防暴雷不破。

## 工單（序列為主）

| 工單 | 工線 | 內容 | depends_on |
|---|---|---|---|
| `SK01` | B | progress models + PatchValidator | - |
| `SK02` | B | SceneGraphProvider + ProgressKernel + graph | SK01 |
| `SK03` | D | ContextBuilder + GameState↔Blackboard bridge | SK01,SK02 |
| `SK04` | A | beat loop 整合（flag/分流/fallback/log）| SK03 |
| `SK05` | B | clue/inventory/NPC 持久化 + debug API | SK04 |
| `SK06` | C | 30-beat 回歸 + 指標 | SK04 |
| `SK07` | C | ON/OFF/fallback 測試 + HUD 打磨 | SK05,SK06 |

> ENABLE_PROGRESS_KERNEL 預設 ON；失敗回退舊流程。GeneratedSceneGraphProvider 留 Patch 2。
