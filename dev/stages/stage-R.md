# 階段 R · 敘事控制 v0.2（揭露橋接 Revelation Bridge）

- **進入條件**：階段 N（NC0–NC7）done、MVP-C（MC1–MC5）done。
- **診斷（為何有這個 patch）**：開 `ENABLE_NARRATIVE_CONTROL` 後開場收斂了，但 v0.1 selfplay 暴露更底層的整合斷鏈——**玩家調查無法轉成官方真相進度**（ProgressKernel clues / NPC-chat hints / real_bible 三層脫節 → 結局永遠 `0/X`）；另有 npc-chat 失控擴張世界觀、重複提問被氛圍敷衍、模糊逃脫渲染同乾淨逃脫、母題停滯、開場後動機淡掉、表層非故事洩漏。
- **目標**：把線「接通」而非加內容。核心橋接 `Evidence → Reveal Level → RevealedBible → Recap`，並把敘事控制延伸到 npc-chat，加答債 / 結局表層變體 / 兩段式逃脫 / 母題冷卻 / 動機心跳 / 表層消毒。
- **整合驗收（done 門檻）**：玩家檢查警示紙條 + 問 NPC 可疑頻率 + 查文件後**結局 recap 不得 0/X**（即使只是 hinted/suspected 也顯示部分發現）｜0 confirmed 逃脫渲染為 ambiguous_escape｜「我試圖離開」先出口候選不即結局｜重複提問付答債｜npc-chat 不發明未授權 lore｜母題不停滯｜表層無洩漏｜flag OFF 全程退回現況。

## 套用順序（patch apply order）

```text
NR0 → NR1 → NR2 → NR3 → NR4 → NR5 → NR6 → NR7
```

> 先 NR0（橋接地基）→ NR1/NR2（npc-chat 控制 + 答債）→ NR3/NR4（結局表層 + 逃脫提交）→ NR5/NR6（母題/動機）→ NR7（消毒收尾）。

## 本階段工單

| 工單 | 工線 | 內容 | depends_on |
|---|---|---|---|
| `NR0` | D | RevelationBridge（線索 → 真相進度橋接；RevealLedger；recap 部分進度） | U10, SK11, UB7 |
| `NR1` | D | NPCChat 敘事控制（同一敘事契約 + 證據橋接） | NR0, MC1, NC4 |
| `NR2` | D | Answer Debt（重複提問逼出部分答/具體線索） | NR1, U13 |
| `NR3` | D | EndingGate Surface Variant（模糊逃脫看得出來） | NC6, UB7 |
| `NR4` | D | Escape Commit Gate（兩段式逃脫；先出口候選再提交） | NR3, SK13 |
| `NR5` | F | Motif Cooldown（重複母題須演化/揭露/暫停） | NC5, NC3 |
| `NR6` | D | Motive Heartbeat（開場後動機每 2–3 beat 可見） | NR0, NC1 |
| `NR7` | F | SurfaceTextSanitizer（render 前清非故事洩漏） | U13, MC1 |

> 核心原則（docs/00）：**不靠叫 Story Agent 多揭露**，而是系統決定哪些線索算真相進度；MVP-B 需要更好的接線，不是更多內容（non-goals：不加新 lore / 新世界機制 / 更大故事圖）。
> 旁路紀律：受 `ENABLE_NARRATIVE_CONTROL`（必要時細分子旗標）控管，預設行為不變；新欄位皆 optional、不破壞既有存檔；**story / npc-chat 永不見 real_bible 不變**；refine（非取代）UB7 masked 結局與 §十二 RevealLadder。
> docs-first：canonical 在 `nightmare-assault-mvp-b-narrative-control-patch-v0.2/`（docs 00–10 / task_cards P0–P7 / reference_code / examples）；reference_code 只是範例，依現有結構改寫、不照抄。
> 契約見 dev/CONTRACTS.md §十四；執行機制見 dev/PARALLEL-PLAN.md §四。
