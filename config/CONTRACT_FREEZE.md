# CONTRACT_FREEZE — 配置中心 / Story Agent 模組化（patch v1.1，階段 P · P0）

> **FROZEN（P0 凍結）。** 動任何 runtime 前先凍結名稱，止住介面漂移。
> canonical 規格在 `nightmare-assault-story-doc-patch-batch-v1.1/`（docs/04 表、docs/05 fragment、docs/08 整合契約）。
> 契約索引在 `dev/CONTRACTS.md §十一`。本檔只凍結「名稱與政策」，**不含 runtime 程式碼**。
> 要改任何凍結名稱：走 `dev/CONTRACTS.md §九` 變更程序（影響盤點 → 使用者同意 → 同步 → 重凍結 → journal）。

---

## 一、Agent 名稱（固定 5 個）

`setup` ｜ `story` ｜ `warden` ｜ `orchestrator` ｜ `compactor`

> 與 `skills/{agent}/SKILL.md`、`core/agents/{agent}.py`、llm_traces.agent 欄一致。`npc-chat`/`dreaming`/`offstage-fate`
> 為輔助 skill，不在配置中心首批管理範圍（P7 之後再議）。

## 二、Profile 名稱（固定 5 個）

| profile | 用途 |
|---|---|
| `mvp_a_safe` | MVP-A 穩定預設（**is_active 預設**） |
| `debug` | 多 log、低創意 |
| `creative` | 高自由度、高 temperature |
| `low_cost` | 便宜模型、緊縮 context |
| `strict_kernel` | 最強 ProgressKernel 服從 |

## 三、Prompt Fragment key 命名

命名規則：`<agent>.<purpose>`（小寫、點分隔）。

**Story Agent 八件**（sort_order）：

| order | fragment_key | 必須 | 用途 |
|---:|---|:---:|---|
| 10 | `story.role` | ✓ | 敘事 renderer 角色 |
| 20 | `story.objective` | ✓ | beat 生成目標 |
| 30 | `story.kernel_obedience` | ✓ | 服從 ProgressKernel（不當世界裁判） |
| 40 | `story.no_repetition` | ✓ | 不重複 forbidden_repeats（開門不再問） |
| 50 | `story.open_choice` | ✓ | 保留自由輸入 |
| 60 | `story.context_policy` | ✓ | 只用可見 context、不洩 real_bible |
| 70 | `story.output_format` | ✓ | 分隔符 + decision JSON 契約 |
| 80 | `story.style_horror` | ○ | 風格（恐怖、緊湊） |

> 其他 agent fragment（P7）：`warden.role` / `warden.local_rule_first` / `orchestrator.role` /
> `orchestrator.no_over_reveal` / `compactor.role` / `compactor.fact_ledger_policy` / `setup.role` /
> `setup.opportunity_graph_policy`（文字見 patch docs/05）。

fragment `status` 狀態機：`draft → published → active → archived`。

## 四、Feature flags（名稱 + 預設）

| flag | 預設 | 階段 | 意義 |
|---|---|---|---|
| `ENABLE_PROGRESS_KERNEL` | true | 既有/SK04 | 用 ProgressKernel 決定推進 |
| `ENABLE_CONFIG_CENTER` | false | P4 開 | story prompt 走配置來源（否則 static-only） |
| `ENABLE_PROMPT_PREVIEW` | true | P2 | 允許不呼 LLM 預覽編譯 prompt |
| `ENABLE_RUN_CONFIG_SNAPSHOT` | true | P6 | 每場 run 存 config/prompt-hash 快照 |
| `ENABLE_THEMED_GRAPH` | true | 既有/SK09 | 主題化機會圖 |
| `ENABLE_SPARSE_FALLBACK` | true | 既有/SK11 | 處理越界行動 |

**flag 解析優先序（固定）**：`.env override > active profile DB config > default profile DB config > hardcoded default`。

## 五、Prompt fallback 政策（固定）

story prompt 來源優先序：

```text
1. active config profile 的 story 編譯 prompt（ENABLE_CONFIG_CENTER=true 時）
2. default config profile 的 story 編譯 prompt
3. static built-in story prompt（skills/story/SKILL.md，永遠保留）
```

- config 載入/編譯失敗 → log error + 退 **static fallback**，遊戲續行，標 run warning。**絕不因配置失敗 crash MVP-A。**
- 載入存檔：優先用該 run 已存的 config 快照（可重現）；遷移到當前 profile 須顯式。

## 六、硬規則（不可破壞）

1. **story `include_real_bible = false`**（context policy 層、程式碼強制，C2/E2）。
2. orchestrator 可讀 real_bible；setup 唯一不可降級（B8）。
3. 前端不解析分隔符/JSON（D4）；玩家輸入包 `<player_action>`（C3）。
4. ProgressKernel 仍是 world-state 真相，story 只 realize。
5. 配置表 migration **additive only**：既有表不改名不刪、舊存檔可讀。
6. PromptComposer **決定性**：同 fragments + 同變數 → 同 compiled prompt 與 prompt_hash；preview 零 LLM。
