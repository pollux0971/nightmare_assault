# 11 · Config Center / Story Agent 模組化（補丁）

> **補丁來源**：`nightmare-assault-story-doc-patch-batch-v1.1/`（docs-first）。
> **實作對應**：開發工單階段 P（`P0`–`P7`）。**狀態：已落地、已測**（`ENABLE_CONFIG_CENTER` 預設 OFF）。
> **核心一句**：把各 agent（先 story）的 prompt 拆成**可配置 fragment**，可組裝 / 可預覽 / 可快照 / 可回滾——而不破壞 MVP-A。
>
> **★ GUI 完成（v0.8，補丁十二）**：前端配置中心補成完整 **Agent Configuration Center**——**Agent Models 表**
> （每 agent primary/fallbacks/temp/max_tokens/enabled，可 Save + 每行 Test；**絕不顯示 api_key**）、
> **Prompt Blocks 表**（含 disabled binding，enabled 開關 / Edit / Save Draft / Activate；sort_order/Rollback 後端
> 無 API → disabled 標示）、**Compiled Preview** by agent/profile（零 LLM）、Test Prompt（後端未支援 → disabled）、
> dirty state + 未存提示 + toast。**只補前端 + API glue（`agent_models_overview`/`save_agent_models`/
> `list_prompt_blocks`/`set_fragment_enabled` + `schema.get_all_bindings`），不改 runtime。**

---

## 一、目標與原則

讓 Story Agent 變可配置、模組化，但 runtime 仍能用 static 預設啟動。設計原則（docs-first、分階段、不大改）：

- Story prompt 拆成 fragment；agent 的 model / temperature / context 預算 / prompt profile / feature flag 變可配置。
- **ProgressKernel 仍是世界狀態真相**（見 `10-progress-kernel.md`），story 仍是「committed 後果的 renderer」。
- 既有 story 流程**分階段**修改，不整包替換。
- **static prompt（`skills/story/SKILL.md`）永遠保留 fallback**；config 失敗一律退 static，不崩 MVP-A。

---

## 二、元件總覽

```
prompt_fragments（DB）
   │  依 agent_prompt_bindings.sort_order、只取 enabled
   ▼
PromptComposer.compose ──► compiled_prompt + 穩定 prompt_hash + enabled_fragments + model_settings + context_policy
   │                       （純函式、零 LLM；preview 不呼模型）
   ▼
ConfigPromptSource.story_system_prompt ──► active profile → default profile → static（SKILL.md）fallback
   │
   ▼
SkillCaller.stream(system_override=…) ──► story agent（config-first，失敗退 static）
```

| 元件 | 檔案 | 職責 |
|------|------|------|
| `ConfigSchema` | `core/config/schema.py` | additive 配置表（10 張）+ 種子 + `ConfigStore`（CRUD） |
| `PromptFragments` | `core/config/fragments.py` | fragment 文字庫（單一來源）+ binding/policy/profile/flag 種子常數 |
| `PromptComposer` | `core/config/composer.py` | fragment → 決定性 compiled prompt + 穩定 prompt_hash + preview + draft overrides |
| `story_prompt` | `core/config/story_prompt.py` | story 專用組裝 + 驗證 + 變數面對齊 |
| `FeatureFlags` | `core/config/flags.py` | flag 解析（`.env` > active profile DB > default DB > hardcoded） |
| `ConfigPromptSource` | `core/config/runtime.py` | config-first 來源解析 + static fallback（永不拋） |
| 配置 UI | `webview_app.py` + `ui/`（`dlg-config`） | draft → preview → activate 的可視化編輯 |

---

## 三、配置表（additive，不破壞既有存檔）

10 張表掛在既有 SQLite 連線上，**只 `CREATE IF NOT EXISTS` + `INSERT OR IGNORE`**（冪等、舊存檔可讀）：

`config_profiles` · `agent_configs` · `prompt_fragments` · `prompt_fragment_versions` · `agent_prompt_bindings` · `agent_context_policies` · `feature_flags` · `run_config_snapshots` · `prompt_test_cases` · `prompt_test_results`。

**Profile**（`config_profiles`）：`mvp_a_safe`（預設 active）/ `debug` / `creative` / `low_cost` / `strict_kernel`。

**Story fragment（8 件）** 與組裝順序：

| sort | fragment_key | 用途 |
|-----:|--------------|------|
| 10 | `story.role` | 敘事 renderer 角色 |
| 20 | `story.objective` | beat 生成目標 |
| 30 | `story.kernel_obedience` | 服從 ProgressKernel（不當世界裁判） |
| 40 | `story.no_repetition` | 不重複 forbidden_repeats（開門不再問） |
| 50 | `story.open_choice` | 保留自由輸入 |
| 60 | `story.context_policy` | 只用可見 context、不洩 real_bible |
| 70 | `story.output_format` | 分隔符 + decision JSON 契約 |
| 80 | `story.style_horror` | 風格（選用） |

---

## 四、PromptComposer 的決定性

`compose(agent, profile, runtime_variables)` →（`compiled_prompt`, `prompt_hash`, `enabled_fragments`, `model_settings`, `context_policy`）：

- 依 `sort_order` 排序、只取 `enabled` binding。
- 變數代入：`{{ name }}` 必填（缺 → strict 拋 / preview 標 `missing_required`）、`{{ name? }}` 選填（缺 → 可見 placeholder + warning）。
- **穩定 `prompt_hash`**：同 fragments + 同變數 → 同 hash；改 fragment / 改變數 → hash 變。
- **preview 零 LLM**：composer 不持有任何 client，是純函式。

> 用途：任一場 run 都能反查「哪個 prompt 版本造成這個行為」。

---

## 五、Runtime 整合（config-first + static fallback）

Story prompt 來源優先序（`ConfigPromptSource`）：

```
1. active config profile 的 story 編譯 prompt   （ENABLE_CONFIG_CENTER=true 時）
2. default config profile 的 story 編譯 prompt
3. static built-in（skills/story/SKILL.md）       ← 永遠保留
```

- flag 解析優先序：`.env override > active profile DB > default profile DB > hardcoded default`。
- config 載入 / 編譯失敗、或缺必含行為 fragment → log + 退下一級；全失敗 → static。**永不 crash MVP-A**。
- `SkillCaller.stream/call` 與 `run_story` 加 `system_override`：非 None 才用 config prompt，否則讀 SKILL.md（向後相容，未啟用時行為與補丁前完全一致）。

---

## 六、配置 UI（draft → preview → activate）

主選單「配置中心」開啟（`dlg-config`），對 Story Agent：

- **fragment 列表 + 編輯器**：看 / 改任一 fragment。
- **預覽**：即時看編譯後 prompt（`PromptComposer.preview`，**零 LLM、不花 token**）。
- **啟用**：編輯只進 **draft**（存 `prompt_fragment_versions`，**不動 active**）；按「啟用草稿」才提升為 active。
- **Feature flags 面板**：安全切 `ENABLE_CONFIG_CENTER` 等。
- 顯示 active profile + prompt hash。

> 安全規則：編輯中的 draft **不影響進行中的遊戲**，直到顯式 activate。

---

## 七、Run 快照 + prompt 回歸

- 每場新 run：對每個已配置 agent compile prompt → 算 prompt_hash → 存 `run_config_snapshots`（profile + config_json + compiled_prompt_hash + enabled_fragments，preview 零 LLM）。載入存檔優先用該 run 的 config 快照（可重現）。
- prompt 回歸測試集（`prompt_test_cases`）：**開門 anti-repetition、story 無 real_bible、prompt_hash 決定性、new clue 出現在 beat、free input 保留** 五項。

---

## 八、擴展至其他 agent（P7）

同一套機制套到 `warden` / `orchestrator` / `compactor` / `setup`（各有 fragments + context policy + agent_config + binding，prompt_hash 進 run snapshot）。**real_bible 可見性規則不變**：

| agent | `include_real_bible` |
|-------|:--------------------:|
| **story** | **false（硬規則 C2/E2）** |
| orchestrator | true（條件揭露用） |
| setup | true（建構 real_bible） |
| warden / compactor | false |

---

## 九、契約索引

開發契約見 `dev/CONTRACTS.md §十一`（`ConfigSchema` / `PromptFragments` / `PromptComposer` / `AgentContextPolicy` / `FeatureFlags` / `ConfigPromptSource` / `RunConfigSnapshot` / `ConfigUI`）。canonical 規格在 `nightmare-assault-story-doc-patch-batch-v1.1/`（docs + task_cards）與本專案 `core/config/`。
