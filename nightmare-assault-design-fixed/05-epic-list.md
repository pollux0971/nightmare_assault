# 05 · Epic List

> 依優先順序排列。每個 Epic 含核心功能，非完整 feature 清單。

---

## 開發階段總覽

> **MVP 收斂（採納評審 #9）**：原 Phase 1 功能偏大（含 audio/dreaming/ending/inventory…對專題太重）。
> 拆成 MVP-A（能玩的核心循環）與 MVP-B（展示差異化），先把 A 跑穩再做 B。

```
MVP-A — 能穩定玩的核心循環（先做這個，跑穩 30 beat）:
  E0 基礎設施 → E1 狀態與記憶（含 scene_registry）→ E2 setup+orchestrator+story
  → E3 beat 迴圈 → E3b compactor（承重牆）→ E4 warden → E6 存檔
  → E7 前端（網頁+pywebview，核心畫面）
  ⟶ 驗收：穩定玩 30 beat、防暴雷、JSON 可解析、存讀檔正確

MVP-B — 展示差異化（A 穩了再加）:
  E4 技能宣稱封頂 + E5 一種結局 + 雙層 bible 防暴雷驗證
  + NPC 簡易 evolving（輕量版 dreaming）+ 道具庫
  ⟶ 驗收：技能封頂、雙層防暴雷可證、一種結局可達

先不做（移到 Phase 2/3）:
  audio、完整 dreaming、npc-chat、offstage-fate、配圖、複雜動畫、完整回溯

Phase 2 — 活的世界:
  E8 聊天室 → E9 dreaming（完整）→ E9b 離場命運 → E10 三層記憶與召回
  → E11 回溯 → E7b 音訊（預製素材）

Phase 3 — 豐富化:
  E12 配圖 → E13 分支樹 → E14 多模板 → E15 難度 → E16 復盤統計 → E17 音樂生成
```

**收斂理由**：LLM 系統 debug 極耗時，audio/dreaming/命運這些雖迷人，但都依賴核心循環先穩。把它們堆進第一個里程碑會讓你遲遲無法驗證「雙層 bible + beat + compactor 能不能跑穩 30 beat」——而那才是專題的說服力來源。原估 51 天的 Phase 1，MVP-A 約 30 天、MVP-B 約 12 天，分兩個里程碑心理負擔小很多。

---

## Phase 1 — MVP-A 核心循環 + MVP-B 差異化

### E0 — 基礎設施 ⬛
建立其他 Epic 依賴的底層。
- Config（YAML + 環境變數）、結構化 Logger
- SignalBus（發布/訂閱）
- OpenRouter LLMClient（重試、Fallback、超時、**串流**）
- 3 模型分層配置
- SKILL.md 動態載入器（prompt 外部化、熱重載）
- JSON 解析與錯誤處理

**驗收**：能呼叫三層 LLM 並串流；SignalBus 收發正常；SKILL.md 可熱重載。

### E1 — 狀態與記憶 ⬛
管理所有狀態，不依賴 LLM。
- Blackboard（含權限模型：dreaming 碰不到 anchor）
- WorldBible / NPCRegistry（anchor vs evolving 分離；real vs revealed 雙層）
- **SceneRegistry（場景系統：current_location + known_locations + interactables）← 補 #2**
- Fact Ledger（二元組增刪查）
- **SharedInventory（共用道具庫，碎片 item 型流入）**
- RollingSummary（有上限）+ BeatWindow（滑動視窗）
- SnapshotStore（每 beat 快照，含當時摘要 + 道具庫 + 隱藏命運獨立 table）
- TokenBudget 追蹤

**驗收**：各狀態模組可獨立測試；快照可還原；權限邊界生效；道具庫增刪查正確。

### E2 — Setup + Orchestrator + Story Agent 🔴
接通核心 LLM 生成與雙層 bible。
- SkillCaller 框架（呼叫、解析、重試、Fallback）
- setup agent：主題 → real_bible（雙層）+ NPC bible + 主角 + 開場序列
- orchestrator：揭露閘門（多數程式碼判條件 + 少數 Light LLM），real → revealed 搬碎片
- story agent：**只讀 revealed_bible** → 生成 beat → 停在決策點
- 三種輸入路徑（選項 / 自訂格式檢查 / 自由打字）
- 輸出 schema 解析

**驗收**：輸入主題能生出雙層 bible；story 看不到未揭露真相；揭露條件達標才釋出碎片；三種輸入皆正確處理。

### E3 — Beat 迴圈與切割 🔴
- 主迴圈（warden → story → 決策點 → 快照）
- `<<<CONTINUE>>>` / `<<<DECISION>>>` 串流狀態機
- 400 字安全網
- 玩家輸入處理（選項 / 自由文字）
- BEAT_COMPLETED 信號 → 快照 + 壓縮檢查

**驗收**：可連續走 30+ beat 不崩；切割穩定；長 beat 正確分塊。

### E3b — Compactor 滾動壓縮 🔴（承重牆）
無限模式的命脈，MVP 不可省。
- compactor agent（非同步，玩家讀字時跑）
- 滾動摘要（散文主線，有上限）+ fact ledger（二元組）維護
- 滑動視窗（保留最近 5-8 beat 原文，更舊折進摘要）
- 使用率三級觸發（L1/L2/L3）
- 保護清單（伏筆、anchor、ledger 不刪）

**驗收**：30+ beat 後 context 不爆、摘要連貫、伏筆不遺失；摘要品質撐得起後續敘事（核心驗證點）。

### E4 — Warden（含技能宣稱）🔴
玩家專屬裁判。
- 致命規則檢查（關鍵詞 + 語義）
- 結局條件偵測（硬/軟 gate）
- 技能宣稱封頂（瞎掰 → 侷限 → ledger 二元組）
- 本地邏輯保底（規則精確匹配）

**驗收**：違規正確觸發死亡；技能宣稱被合理封頂且侷限融入劇情；軟結局正確 gate。

### E5 — 結局系統 🟠（MVP-B；MVP-A 不要求）
- ending_conditions 結構（終局狀態 + gate/前置）
- 結局序列生成（純敘述收尾 beat）
- 提早結束的劇情內化解
- 死亡/真相/逃脫三種結局

**MVP-B 驗收**：至少一種結局可達；提早觸發在劇情內合理化解，不生硬擋下。完整四種結局移到 Phase 2/3 後再補齊。

### E6 — 存檔與快照 🔴
- 自動每 beat 快照（含當時摘要 + 道具庫 + NPC 狀態）
- 隱藏命運存獨立 table（不主動劇透，路線 A）
- 具名存檔點 / 讀檔
- SQLite 持久化

**驗收**：存讀檔正確；還原到舊 beat 摘要狀態正確（不錯亂）；道具庫與聊天紀錄隨快照保存。

### E7 — 前端（網頁+pywebview）🔴
網頁前端 + pywebview 桌面殼，後端維持 Python。細部 epic 見 06-frontend.md 第十四節。
- F1 pywebview 骨架（視窗、JS↔Python 綁定、view 切換、CSS 主題）
- F2 串流渲染（Python 背景 thread + StreamParser → JS 接收已分類事件、逐字渲染 + 節奏）← 承重牆
- F3 遊戲主畫面（敘事區 + 決策按鈕 + 自由輸入，HTML/CSS）
- F4 前置畫面（啟動/設定/主選單/新局/載入）
- F6 存檔 UI
- F5 道具面板（modal，MVP-B 可完整化）
- F7 結局畫面（MVP-B：至少一種結局串流 + 真相揭露 + 復盤）

**驗收**：完整畫面流程走通；串流逐字節奏到位且 UI 不凍結；決策點正確呈現；道具面板可查看；深色恐怖主題與內嵌襯線字體生效；換解析度不跑版。

### E7b — 音訊與等待動畫 🟡（移至 Phase 2，採納 #9 收斂）
細部見 06-frontend 第十、十一節。
- F10 等待動畫：Pillow 氛圍動畫 + 恐怖台詞 + 序章儀式
- F11 音訊播放機制：分層播放（音樂+環境音）、crossfade、pacing 切換、心跳加速、audio_cue
- 音樂用**預製免費恐怖素材**（非生成），先把機制做對
- 動態音量（手動）

**驗收**：音樂隨 pacing 切換且 crossfade 平順；環境音循環、心跳隨緊張加快；silence/sting/swell 正確觸發；等待畫面是氛圍體驗而非 loading bar。
（音樂**生成**接入 = Phase 3 的 F12，技術風險最高，最後做。）

**MVP-A 驗收**：主題輸入 → 開場 → 連續 30+ beat；防暴雷；JSON 可解析且可 repair；存讀檔正確；context 不爆、敘事連貫。  
**MVP-B 驗收**：技能宣稱正確封頂；item 碎片入道具庫且可查看；至少一種結局可達；能展示雙層 bible 防暴雷驗證。

---

## Phase 2 — 活的世界

### E8 — 聊天室 🟡
- npc-chat agent 整合（獨立小 context）
- 聊天室 UI（訊息串、NPC 選擇、結束按鈕）
- 故事內聊天（beat 變體，dialogue 決策型）vs 聊天室（主動）區分
- story beat 召喚聊天室
- 退出濃縮三向分流

**驗收**：聊天室可跑 10+ 輪；NPC 語氣符合性格與演化層。

### E9 — Dreaming Mode（在場 NPC）🟡
- dreaming agent（非同步，無 warden 閘門，只寫 evolving）
- 情緒/關係/意圖/個人目標演化
- emergent_lie 自我約束（讀 ledger）
- reflection_log（觀察情感湧現）
- 只跑在場 active NPC（成本優化）
- NPC_EVOLVED → story 採用

**驗收**：在場 NPC 情緒隨劇情/聊天演化；聊天反向影響主線；沒戲份者凍結。

### E9b — 離場 NPC 命運機制 🟡
與 E9 不同 skill（生成式）。
- NPC 雙軸狀態（presence × alignment，含 missing）
- 離線命運 tick + fate_pressure 計量器（程式碼擲骰）
- 加權命運表（受 alignment 影響）：機遇/失蹤/屍體/敵對
- offstage-fate skill（LLM 寫血肉）
- 命運從 revelation_pool 領碎片，過揭露閘門（不暴雷）
- 屍體結局種 corpse_interactable 到 Scene
- 重逢/撞見的場景觸發

**驗收**：離場 NPC 朝命運收束；四種結局皆能觸發且各攜帶錨定碎片；屍體可被搜出線索；不暴雷。

### E10 — 三層記憶與召回 🟡
- Cold 封存（完整聊天紀錄、舊 beat、reflection_log → SQLite）
- 語義召回（依 NPC/主題從 cold 拉回 hot）
- Prompt caching（story 靜態前綴）

**驗收**：完整紀錄不遺失；召回正確；caching 降低 story 成本。

### E11 — 回溯 🟡
- 線性 undo（還原快照 + 截斷）
- 回溯 UI（選 beat、提示非決定性）
- 可選：(快照,選擇) → beat 快取

**驗收**：回溯狀態正確（含摘要）；玩家理解非決定性。

---

## Phase 3 — 豐富化

| Epic | 內容 | 優先 |
|------|------|------|
| E12 配圖 | gpt-image 由 `beat_meta.pacing` 驅動的分鏡示意圖 | 🔵 |
| E13 分支樹 | 回溯保留多分支，可探索不同路徑 | 🔵 |
| E14 多模板 | 第二個 world bible 模板（鬼屋/深海站），模板切換 | 🔵 |
| E15 難度 | 調揭露閘門鬆緊：碎片隱晦度 + dreaming 揭露速度（beat 為單位） | 🔵 |
| E16 復盤統計 | 存活 beat 數、發現碎片比例、關鍵決策時間軸、結局復盤 | 🔵 |
| E17 音樂生成接入 | 背景分層預生成（基本+進階音軌）、音軌庫管理，替換預製素材 | 🔵 |

---

## 工時與依賴

| Epic | 優先 | 預估 | 依賴 |
|------|------|------|------|
| E0 基礎設施 | ⬛ | 5 天 | 無 |
| E1 狀態與記憶 | ⬛ | 7 天 | E0 |
| E2 setup+story | 🔴 | 6 天 | E0,E1 |
| E3 beat 迴圈 | 🔴 | 5 天 | E2 |
| E3b compactor | 🔴 | 4 天 | E3 |
| E4 warden | 🔴 | 5 天 | E2,E3 |
| E5 結局 | 🟠 | 3 天 | E4；MVP-B |
| E6 存檔 | 🔴 | 3 天 | E1,E3 |
| E7 前端 網頁+pywebview | 🔴 | 9 天 | E3 |
| E7b 音訊與等待動畫 | 🟡 | 5 天 | E7；Phase 2 |
| E8 聊天室 | 🟡 | 5 天 | E2,E3 |
| E9 dreaming | 🟡 | 5 天 | E1,E8 |
| E9b 離場命運 | 🟡 | 4 天 | E9 |
| E10 三層記憶 | 🟡 | 4 天 | E1,E8 |
| E11 回溯 | 🟡 | 3 天 | E6 |
| E12–E16 | 🔵 | 各 3–7 天 | 視項目 |
| **MVP-A 合計** | | **~39 天** | 不含 E5、E7b、完整道具/結局 |
| **MVP-B 追加** | | **~12 天** | E5 + 技能/道具/NPC 輕量 evolving |
| **Phase 1+2 合計** | | **~68 天** | |

---

## 待決事項（更新後）

**已解決（見 00 決策日誌）**：互動主軸（A2 三輸入路徑）、SAN（B7 可選軟指標）、world bible 全藏半露（A3 雙層 bible + 半露鉤子）、難度（B9 = 揭露速度）、回溯（線性 MVP）。

**仍待決**：
1. **dreaming 頻率 K**：在場 NPC 每幾 beat 反思一次？離場 NPC 倍率？建議在場 K=5、離場 3K，跑起來再調。
2. **emergent_lie 自我約束強度**：完全放任（偶有微矛盾）vs 中等（讀 ledger 軟約束）。MVP 中等起。
3. ~~離場 NPC dreaming 是否隱藏~~ **已定：完全隱藏**（資訊差製造重逢恐怖），存檔採路線 A（獨立 table 不主動劇透）。
4. ~~自訂格式輸入模板~~ **已定**：姓名/個性/外觀/關係四欄 + Light 檢查，附數個預設角色範本（見 04）。
5. **聊天是否計入 beat number / 軟結局 gate 計數**：傾向不計入（聊天是側通道）。
6. **A2 主角固定度**：starting_situation 給身分但不鎖性格（鬆定性），離格行動 story 照接——需在 setup prompt 明確。
7. **命運表的具體權重數值**：機遇/失蹤/屍體/敵對在各 alignment 下的實際機率，需實測平衡。
