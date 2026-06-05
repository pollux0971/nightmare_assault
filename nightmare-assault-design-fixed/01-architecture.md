# 01 · 系統架構

---

## 一、分層架構

```
┌─────────────────────────────────────────────────────────┐
│                  Agent Layer（LLM 決策）                  │
│                                                           │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐          │
│  │ setup  │ │ story  │ │ warden │ │ npc-chat │          │
│  └────────┘ └────────┘ └────────┘ └──────────┘          │
│  ┌──────────┐ ┌────────────┐                             │
│  │ dreaming │ │ compactor  │                             │
│  └──────────┘ └────────────┘                             │
├─────────────────────────────────────────────────────────┤
│                  State Layer（狀態管理）                  │
│  Blackboard · WorldBible · NPCRegistry · Ledger          │
│  BeatHistory · SnapshotStore · TokenBudget               │
├─────────────────────────────────────────────────────────┤
│                Persistence Layer（持久化）                │
│   SQLite（beat 快照 / 聊天封存 / 存檔）· JSON（設定）     │
└─────────────────────────────────────────────────────────┘
```

State Layer 不依賴 LLM，可獨立測試。Agent Layer 透過 Blackboard 讀寫狀態，不直接互相呼叫。

---

## 二、Agent 編制

| Agent | 模型分層 | 觸發頻率 | 職責 | 串流 |
|-------|---------|---------|------|------|
| **setup** | Heavy | 一次（開新局） | 生 real_bible + NPC bible + 主角 + 開場序列 | 否 |
| **orchestrator** | Light | 每個 beat | 揭露閘門：檢查條件，把碎片從 real 搬到 revealed（多數程式碼判） | 否 |
| **story** | Medium | 每個 beat | 核心迴圈，寫 beat，停在決策點（**只讀 revealed_bible**） | **是** |
| **warden** | Light | 每個 beat（玩家動作） | 致命規則 + 結局條件 + 技能宣稱封頂（**僅玩家**） | 否 |
| **npc-chat** | Light | 隨選（聊天室開啟時） | 玩家主動發起的多輪對話；NPC 間互動 | 是 |
| **dreaming** | Light | 非同步（每 K beat / 觸發） | **在場** NPC 反應式反思，更新演化層（無 warden 閘門） | 否 |
| **offstage-fate** | Light | 非同步（命運觸發才跑） | **離場** NPC 生成式命運（機遇/失蹤/屍體/敵對），程式碼擲骰+LLM 寫血肉 | 否 |
| **compactor** | Medium | 非同步（使用率門檻） | 滾動摘要 + ledger 維護 + 聊天封存濃縮 | 否 |

設計要點：
- 唯一每 beat 都跑且串流的生成 agent 是 **story**，它只讀 `revealed_bible`，結構上不可能暴雷。
- **orchestrator** 在 story 之前跑，決定本 beat 揭露什麼；多數揭露條件程式碼判定，省成本。
- **warden** 與 **story** 串聯（warden 先判玩家動作 → story 依指令生成）。
- **dreaming** 與 **compactor** 是回合外非同步（玩家讀字時跑，不卡互動）。

> **補丁更新（見 §七、`10-progress-kernel.md`）**：在 Progress Kernel 啟用下，**世界狀態怎麼推進由程式碼層 kernel 決定，story 退化為純 realizer**（只把已 committed 的結果寫成敘事）。warden 之後接的是 kernel 而非 orchestrator 自由流程；orchestrator 揭露閘門仍在（legacy 流程與揭露管線保留，kernel 失敗時回退）。

---

## 三、模型分層與 Fallback

| 層級 | 主模型 | Temperature | 用途 |
|------|--------|-------------|------|
| **Heavy** | claude-sonnet-4 | 0.6–0.7 | setup（真相密度高、一次性，值得用好模型） |
| **Medium** | gemini-flash | 0.7–0.8 | story、compactor（創意敘事 + 結構整合） |
| **Light** | gpt-4.1-mini | 0.3–0.6 | warden（低溫，精確裁判）、npc-chat、dreaming |

```
Fallback 鏈:
  Heavy:  claude-sonnet-4 → claude-3.5-sonnet → gemini-flash
  Medium: gemini-flash    → claude-haiku-3.5  → gpt-4.1-mini
  Light:  gpt-4.1-mini    → gemini-flash-lite → 本地邏輯（warden 規則可程式碼保底）
```

warden 的致命規則「關鍵詞精確匹配」與技能封頂的部分規則可在純 Python 邏輯保底，LLM 失效時不致全盤崩潰。

---

## 四、權限模型（誰能寫什麼）

這是並行安全與連貫性的基礎。錨點的不可變性靠**權限**強制，而非靠 warden。

| 寫入者 | 可寫入欄位 | 不可碰 |
|--------|-----------|--------|
| setup | 全部（初始化）→ 之後 `real_bible`、`secret_core` 鎖死 | — |
| orchestrator | `revealed_bible`（從 real 搬碎片）、`turn_context.newly_revealed` | `real_bible`（只讀）、`secret_core` |
| story | `BeatHistory`（追加）、`turn_context.narrative` | `real_bible`（**讀不到**）、`secret_core`、`npc_evolving` |
| warden | `turn_context.warden_verdict`、`Ledger`（技能二元組） | `real_bible`（只讀）、`npc_evolving` |
| npc-chat | `ChatLog`（追加） | 一切 anchor |
| **dreaming** | **`npc_evolving`（演化層）、`offstage_intent`** | **`real_bible`、`secret_core`（權限邊界）** |
| compactor | `RollingSummary`、`Ledger`（重組）、`ColdArchive`、`recent_chat_digest` | 一切 anchor |

關鍵：
- **story agent 讀不到 `real_bible`**——暴雷在結構上不可能，不靠自律。
- **dreaming 在程式碼層被禁止寫入錨點**——即使無 warden 檢查 NPC，核心真相仍鎖死。

---

## 五、SignalBus 事件

發布/訂閱模型，解耦 agent 與系統反應。

| 事件 | 發布者 | 訂閱反應 |
|------|--------|---------|
| `BEAT_COMPLETED` | story | 觸發快照、檢查壓縮使用率、檢查 dreaming 排程 |
| `ENDING_TRIGGERED` | warden | 停止 beat 迴圈、進入結局序列 |
| `RULE_VIOLATION` | warden | story 生成死亡 beat |
| `SKILL_CLAIMED` | warden | 寫入 ledger、UI 提示侷限 |
| `CHATROOM_OPENED` / `CLOSED` | 前端 | 啟動/關閉 npc-chat 子迴圈；關閉時觸發濃縮 |
| `NPC_EVOLVED` | dreaming | 標記該 NPC 演化層已更新（story 下個 beat 採用） |
| `CONTEXT_THRESHOLD` | compactor | 啟動對應等級壓縮 |

---

## 六、目錄結構

```
nightmare-assault/
├── core/                         # Python：後端邏輯 + orchestration
│   ├── orchestrator.py           # beat 迴圈、回合流程
│   ├── agents/                   # SkillCaller、各 agent 呼叫封裝
│   ├── state/                    # Blackboard、WorldBible、NPCRegistry、Ledger
│   ├── memory/                   # 三層記憶、滾動摘要、快照
│   ├── llm/                      # OpenRouter client、fallback、串流
│   ├── signal.py                 # SignalBus
│   └── persistence/              # SQLite、JSON
├── ui/                           # 網頁前端（HTML/CSS/JS），由 pywebview 載入
│   ├── index.html                # 單頁應用，各畫面為 section
│   ├── css/
│   │   ├── theme.css             # 深色現代恐怖主題（變數、配色）
│   │   └── animations.css        # 逐字浮現、淡入、明暗呼吸
│   ├── js/
│   │   ├── streaming.js          # 接收 token、分隔符狀態機、節奏控制
│   │   ├── views.js              # 畫面切換
│   │   ├── audio.js              # Web Audio 分層播放
│   │   └── api.js                # 呼叫 window.pywebview.api
│   └── assets/
│       ├── fonts/                # 內嵌襯線字體（.woff2，跨電腦一致）
│       └── audio/                # 預製環境音/音樂素材（MVP）
├── webview_app.py                # pywebview 入口：建視窗、暴露 API class
├── skills/                       # 各 agent 的 prompt（markdown，可熱重載）
│   ├── README.md
│   ├── setup/SKILL.md
│   ├── orchestrator/SKILL.md
│   ├── story/SKILL.md
│   ├── warden/SKILL.md
│   ├── npc-chat/SKILL.md
│   ├── dreaming/SKILL.md
│   ├── offstage-fate/SKILL.md
│   └── compactor/SKILL.md
├── story_templates/              # 故事模板（可選）
├── config/                       # YAML 設定、模型分層配置
├── storage/                      # 執行期：存檔、快照、聊天封存（SQLite）
├── main.py                       # 進入點：呼叫 webview_app 啟動桌面視窗
└── pyproject.toml                # Python 專案（含 pywebview 依賴）
```

**後端 Python + 網頁前端（pywebview 橋接）**：後端 core 是純 Python；前端是 HTML/CSS/JS，由 pywebview 載入並提供 JS↔Python 綁定（同程序，無需 HTTP/subprocess）。前端按鈕呼叫 `window.pywebview.api.xxx()` 直達後端；LLM 串流放 Python 背景 thread，逐 token 用 `evaluate_js` 推給前端渲染——見 06 前端文件。

`skills/` 把 prompt 從程式碼抽出成 markdown，可熱重載、可版本控制。不需要複雜的 skill 引擎——一個 loader 即可。

---

## 七、Narrative Progress Kernel（補丁一 · 旁路增量層）

> 完整說明見 `10-progress-kernel.md`。這裡只放架構定位。

在 Agent Layer 與 State Layer 之間插入一層 **Progress Kernel**：把「世界狀態怎麼推進」從 LLM 手上拿走，交給程式碼層決定，**story 退化為純 realizer**。

```
玩家行動 → warden → ProgressKernel.resolve_player_action（EventPatch + 義務 + forbidden）
        → ContextBuilder（story 最小 context，無 real_bible）
        → story（realize 已決定的結果，串流）
        → PatchValidator.apply（base_version + 強制 ≥1 progress_delta）→ commit
        → ProgressBridge.sync_to_blackboard（clues/inventory/npc/scene 進 snapshot）
        → 結局 attractor 檢查
```

| 元件 | 檔案 | 角色 |
|------|------|------|
| `ProgressKernel` | `core/progress_kernel.py` | 決定本 beat 推進的事件 |
| `PatchValidator` | `core/patch_validator.py` | 版本檢查 + 強制每步有 delta |
| `SceneGraphProvider` | `core/scene_graph.py` | 機會圖（static / 主題化 generated） |
| `ContextBuilder` | `core/progress_context.py` | story 最小 context（結構性防暴雷） |
| `ProgressBridge` | `core/progress_bridge.py` | GameState ↔ Blackboard 同步 |
| `Attractors` | `core/attractors.py` | 結局吸引子（非固定終點） |

**旁路紀律**：`ENABLE_PROGRESS_KERNEL` 預設 ON；`BeatLoop` 單一分流（`_step_kernel` vs `_step_legacy`），kernel/圖/驗證失敗 → 回退 legacy，不 crash。

---

## 八、Config Center（補丁二 · 旁路增量層）

> 完整說明見 `11-config-center.md`。這裡只放架構定位。

在 State Layer 旁加一層 **Config Center**：把 agent（先 story）的 prompt 拆成可配置 fragment，可組裝 / 預覽 / 快照 / 回滾，而 runtime 仍能用 static 預設啟動。

```
prompt_fragments（SQLite，additive）
  → PromptComposer（依 sort_order 組裝、穩定 prompt_hash、零 LLM preview）
  → ConfigPromptSource（active profile → default profile → static SKILL.md fallback）
  → SkillCaller.stream(system_override=…) → agent
```

| 元件 | 檔案 | 角色 |
|------|------|------|
| `ConfigStore` / 配置表 | `core/config/schema.py` | additive 配置表 + 種子 + CRUD |
| `PromptComposer` | `core/config/composer.py` | 決定性組裝 + prompt_hash + preview |
| `ConfigPromptSource` | `core/config/runtime.py` | config-first 來源 + static fallback |
| `FeatureFlags` | `core/config/flags.py` | `.env > active > default > hardcoded` |
| 配置 UI | `webview_app.py` + `ui/`（`dlg-config`） | draft → preview → activate |

**旁路紀律**：`ENABLE_CONFIG_CENTER` 預設 OFF（行為與補丁前一致）；config 失敗一律退 `skills/story/SKILL.md`，不崩 MVP-A。**story `include_real_bible=false` 硬規則不變**。
