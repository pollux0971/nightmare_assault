# 00 · 專題總覽

> **Nightmare Assault — LLM 驅動的無限恐怖文字冒險**
> **技術棧**：Python（後端 core）+ 網頁前端（HTML/CSS/JS）+ pywebview（包成桌面應用）+ OpenRouter API。後端維持 Python，前端用網頁技術，pywebview 橋接。
> **語言**：繁體中文
> **模式**：無限故事模式（Infinite Story Mode）
>
> **★ 最新定位（先讀）**：本作是**開放式恐怖探索遊戲**——驗收標準是「玩家做的事有沒有留下**可檢查的世界後果**」，
> 不是「有沒有推進主線」。方向見 `15-player-sovereignty.md`；落地機制（WorldModel 抽象實體 / 主題無關角色 /
> 撤離鎖 / WorldConsequence-vs-TruthEvidence 分流 / 空間投影）見 **`16-worldmodel.md`** 與 **`17-spatial-projection.md`**。
> 本檔以下為**原始設計理念**（仍成立，但「世界真相驅動」已由 WorldModel 補上一層可被指涉的實體機制）。

---

## 一、這個專題在做什麼

一款由多個 LLM agent 協同驅動的恐怖文字冒險遊戲。玩家輸入一個主題，系統生成一個有「真相」的恐怖世界，然後以**分鏡（beat）**為單位無限地把故事說下去。玩家在每個分鏡的決策點決定主角的行動（打字或選擇），故事據此延展。NPC 是會反思、會演化、會流露情感的角色，而非靜態道具。

與一般 AI 文字冒險（如 AI Dungeon）的差異：本作用**世界真相**而非**劇情骨架**驅動，因此同時拿到「玩家自由」與「劇情連貫」——這是整個設計的核心命題。

---

## 二、三大核心理念

### 1. 世界真相，不是劇情骨架（World Bible, not Plot Outline）

不可能三角：高自由度、劇情連貫、預定義骨架，三者最多同時要兩個。本作選擇「自由 + 連貫」，代價是放棄預定義劇情骨架，改用**世界真相**：定義「發生過什麼、威脅是什麼、什麼致命、有哪些秘密可被發現」，但**不**定義事件順序。玩家怎麼跑都自由，而他發現的每個碎片自動互相吻合——因為它們同源。

**雙層結構（揭露閘門）**：world bible 分兩層——`real_bible`（完整真相，只有 orchestrator 與 warden 能讀）與 `revealed_bible`（已揭露子集，story agent 只能讀這個）。orchestrator 是揭露閘門，每個 beat 檢查「進階揭露條件」是否達標，達標才把真相碎片從 real 搬到 revealed。如此 story agent 結構上看不到未揭露的真相，**不可能暴雷**——「該知道多少」從自律問題變成結構問題。揭露速度也成為控制故事長度與節奏的旋鈕（見難度系統）。「有目的地但隱晦」：終點存在，但靠揭露閘門讓玩家自己摸索。

### 2. 分鏡迴圈，停在決策點（Beat Loop, Stop at Decision）

故事不預先生成再刪除，而是 story agent **寫到主角即將抉擇就停筆**。一個分鏡 = 從劇情自然流動到主角必須做出有意義抉擇為止。長度不限，靠「決策功能」而非「字數」切割。可 streaming、不浪費 token、不需要監看 agent。

### 3. 活的世界（Living World）

NPC 不是靜態的。透過 **dreaming mode**（非同步反思），NPC 會在自己的「演化層」長出情緒、意圖、新的謊言、個人目標。世界真相是不可變的錨點；NPC 的演化層是可變的。聊天與劇情會反向影響 NPC，再透過 NPC 影響主線。

### 4. 逃生即推理（Escape-as-Inquiry）— 類型定性

威脅給壓力，資訊給出路。不是「先解謎再逃」或「先逃再解謎」，而是**為了逃而被迫解謎**：威脅在追 →（恐怖）得找到離開的規則 →（推理）規則藏在不敢去的地方 →（恐怖）……逃生需求逼著摸資訊，摸資訊的過程製造恐怖。恐怖與推理互為因果，而非並列。

定位（經確認）：**威脅是核心壓力來源，推理是主要出路**——逃避威脅與摸清資訊同等重要。參考座標：《返校》《煙火》一線，非純解謎也非純生存恐怖。

**關鍵結構**：結局條件 = 資訊拼圖的完成。「找到規則離開／找人／找關鍵道具」是同一件事的三種形式——都是 `revelation_pool` 的碎片，差別只在碎片型別（knowledge／item／person）。

---

## 三、核心不對稱原則：Warden 只管玩家

> **Warden（守門人）只對玩家施加限制，不約束 NPC。**

理由：保護「世界真相錨點」有兩道**各自管不同對象**的機制——

1. **程式碼權限邊界（管 NPC）**：dreaming/offstage-fate 不能寫 `world_truth` / `secret_core`，只能寫 `npc_evolving`。NPC 的錨點靠這道結構性保護鎖死，**不經過 warden**。
2. **玩家專用 Warden（管玩家）**：只約束玩家的破格行為（破格技能、違反致命規則、提早觸發結局）。

換言之：**玩家 → warden；NPC → 權限邊界 + dreaming 自我約束**。兩道機制不交叉——NPC 的 evolving 變更**不會**送 warden 檢查。只有玩家能真正破壞遊戲，所以只有玩家需要裁判。NPC 至多偶爾微矛盾——對「人心不可信」的恐怖題材，這多半是氣氛而非缺陷。

| 對象 | 錨點保護 | Warden 約束 | 自由度 |
|------|---------|------------|--------|
| **玩家** | — | 致命規則 + 結局條件 + 技能封頂 | 受裁判 |
| **NPC** | 權限邊界（碰不到 core） | 無 | 演化層自由生長 |

NPC 的一致性改由 dreaming 自己在 prompt 層自我約束（讀 ledger，被要求不矛盾於硬事實），是軟性繫繩，不是硬性閘門。

---

## 四、文件地圖

| 檔案 | 內容 |
|------|------|
| `00-overview.md` | 本檔。專題總覽、核心理念、詞彙表 |
| `01-architecture.md` | 系統架構、agent 編制、模型分層、權限模型、目錄結構 |
| `02-algorithms.md` | 核心演算法：beat 迴圈、切割、warden、dreaming、壓縮、存檔、結局 |
| `03-agent-context.md` | 三層記憶、Blackboard schema、各 agent 的 I/O 與 context 切片、token 預算 |
| `04-ui-ux.md` | UI/UX 設計原則與內容：互動規則、節奏、文案、版面內容（與技術無關） |
| `05-epic-list.md` | Epic 清單、開發階段、優先序、依賴、MVP 範圍、待決事項 |
| `06-frontend.md` | 前端技術規劃：網頁+pywebview 架構、串流渲染、深色恐怖主題（CSS）、畫面流程、跨電腦一致性 |
| `07-data-contracts.md` | 資料契約：API contract、各 agent Pydantic schema、串流 JSON 解析管線與 fallback、event 抽取、injection 防護 |
| `08-engineering.md` | 工程實作：SQLite schema、狀態 patch 與並行控制、錯誤恢復骨架、API key 安全、測試驗收指標 |
| `09-revision-notes.md` | 契約統一修訂重點 **+ 後續兩個補丁里程碑** |
| `10-progress-kernel.md` | **補丁一**：Narrative Progress Kernel——世界狀態權威、story 退化為 realizer、每步有後果、attractor 結局 |
| `11-config-center.md` | **補丁二**：Config Center——story prompt 模組化（fragment）、決定性組裝、config-first + static fallback、draft→preview→activate |
| `12-mvp-b.md` | **MVP-B 差異化**：技能封頂強化 / 一種結局序列 / 輕量 dreaming / 道具庫完整 / 防暴雷驗證報告（UB1–UB5） |
| `architecture.html` | **多代理架構視覺簡報**（瀏覽器開；← → 換頁），把本資料夾的結構視覺化 |
| `skills/` | 8 個 agent 的 prompt（SKILL.md）+ README 索引。對應 agent 已標注於各檔 frontmatter |

建議閱讀順序：00 → 01 → 02 → 03 → 04 → 05；補丁脈絡 09 → 10 → 11；想快速綜覽直接開 `architecture.html`。

---

## 五、詞彙表

| 術語 | 定義 |
|------|------|
| **Beat（分鏡）** | 故事的基本單位。從劇情流動到主角必須抉擇為止。長度不限。 |
| **World Bible（世界真相）** | 不可變的世界設定：真相、威脅、致命規則、可發現的碎片池、結局條件。 |
| **Anchor（錨點）** | World bible 中永不變動的部分（`world_truth`、NPC 的 `secret_core`）。 |
| **Evolving Layer（演化層）** | NPC 可變的狀態：情緒、意圖、已揭露層次、新謊言、個人目標。 |
| **Fact Ledger（事實帳本）** | 滾動摘要中的硬狀態，以具型別的二元組記錄。 |
| **Dreaming Mode** | 非同步的 NPC 反思 pass，更新演化層。compactor 的兄弟。 |
| **Warden（守門人）** | 只對玩家的裁判 agent：致命規則、結局條件、技能宣稱封頂。 |
| **Decision Point（決策點）** | 一個 beat 的結尾，主角必須抉擇之處。選項 + 自由輸入框。 |
| **三層記憶** | Hot（live context）/ Warm（ledger）/ Cold（封存可召回）。 |
| **Skill Claim（技能宣稱）** | 玩家即興宣稱能力，warden 加上侷限後成為事實。 |

---

## 六、已拍板的設計決定（決策日誌）

這些是經討論定案的決定，作為後續文件與實作的依據。

| # | 決定 | 影響的文件 |
|---|------|-----------|
| A1 | premise「有目的地但隱晦」：終點存在，靠揭露閘門讓玩家摸索 | 雙層 bible（02、03） |
| A2 | 三種輸入路徑：預設選項（直接用）／自訂格式輸入（過一次 Light 檢查格式與合理性）／完全自由打字（不檢查，直接給 story agent） | 04、02 |
| A3 | 雙層 world bible：`real_bible`（鎖）+ `revealed_bible`（story 可讀），orchestrator 當揭露閘門，含進階揭露條件 | 01、02、03 |
| A4 | 主題輸入暫不限制，setup 一律恐怖框架化（先測試可行性） | 02 setup |
| B5 | 聊天室：NPC 間可互動（不必推進主線）；故事內 NPC 聊天視為推進。退出時近期聊天濃縮成 3-4 句進 story hot context，完整紀錄入 cold | 02、03、04 |
| B6 | NPC 可死：`presence=dead`、停止 dreaming，保留 reflection_log 與已揭露資訊 | 02、03 |
| B7 | 純敘事判定生死為主 + 可選的敘事驅動 SAN 軟指標（非戰鬥數值） | 02、04 |
| B8 | 結局後揭露完整 world truth + 簡短復盤 → 回選單 | 02、05 |
| B9 | 難度 = 調揭露閘門鬆緊：真相碎片隱晦度 + dreaming 揭露速度（以 beat 為單位） | 02、05 |
| B10 | 單純續玩（狀態還原），不做離線做夢 | 05 |
| — | NPC 雙軸離場狀態：`presence`（present/absent/missing/dead）× `alignment`（allied/neutral/departed/hostile/dead）。離場 NPC 用**獨立的命運機制**（非在場 dreaming）：程式碼擲骰決定機遇/失蹤/屍體/敵對，LLM 寫血肉，每種命運從 revelation_pool 領碎片過揭露閘門 | 02、03、05 |
| 類型 | **逃生即推理**：威脅核心壓力、推理主要出路。結局條件 = 資訊拼圖完成 | 00、02 |
| 碎片三型 | revelation_pool 碎片分 knowledge／item／person，揭露後分別流向已知資訊／道具庫／NPC 位置 | 02、03 |
| 道具庫 | 共用道具庫，item 型碎片流入。`/inventory` 查看（不顯示是否關鍵道具）。held_by 可綁 NPC，命運跟著走 | 02、03、04 |
| 儲存 | 世界狀態快照涵蓋故事/聊天/dreaming/道具庫。隱藏命運存 SQLite 獨立 table（路線 A，不主動劇透） | 02、05 |
| 離場隱藏 | 離場 NPC dreaming/命運對玩家完全隱藏，重逢才揭曉 | 02 |
| 自訂格式 | 角色自訂四欄（姓名/個性/外觀/關係）+ Light 檢查 + 數個預設角色範本 | 04 |
| NPC 認知卡 | npc-chat 與 dreaming 共用：自我/環境/關係/專業四層，各有邊界。結構化餵入、指令極短、voice_sample 錨定個性 | 03 |
| NPC 個性隨機化 | 程式碼多軸擲骰（核心特質×說話節奏×情緒底色×怪癖），不花 LLM token | 03 |
| NPC self_aware | secret_core 拆出 self_aware bool：true 隱瞞說謊／false 真誠答錯（兩種恐怖） | 02、03 |
| NPC 職業 | 每個 NPC 有 profession，決定 knows_about；職業外問題不裝懂而用職業視角折射（常順帶給線索） | 03、04 |
| 技術棧 | 後端 Python（原計畫 Go，已改）；前端最初設想 customtkinter，後改**網頁（HTML/CSS/JS）+ pywebview**——因體驗命脈在文字動畫與排版，CSS 是母語，且跨電腦一致性最佳（字體內嵌+響應式），原生 GUI 反易跑版 | 01、06 |
| 前端 | 網頁技術（HTML/CSS/JS）+ pywebview 包成桌面應用。深色現代恐怖（CSS）。改用 web 因體驗命脈在文字動畫與排版，且 CSS 跨電腦一致性最佳（字體內嵌+響應式），原生 GUI 反易跑版 | 06 |
| 音訊 | 音樂分層**預生成**音軌庫（基本於序章備好、進階依預兆提前生成），播放時選軌+crossfade，絕不即時生成；環境音用素材庫；MVP 用預製素材先做機制 | 02、06 |
| audio_cue | beat_meta 加 audio_cue（normal/silence/sting/swell），story 可觸發特殊音訊（含驟然靜默） | 02、03 |
| 旁白型 beat | beat 可純敘事（CONTINUE 收尾、無決策）；連續旁白超 2-3 個程式碼強制給決策 | 02 |
| 場景系統 | 新增 scene_registry（current_location + known_locations + interactables），支撐 location_reached/屍體種植/item 位置（補評審 #2，進 MVP-A） | 03 |
| 串流 JSON 韌性 | 三級 repair（程式碼修復→LLM repair→fallback decision UI），narrative 已串流部分不回收（補 #4） | 07 |
| 並行控制 | 非同步 agent 只產 patch，主迴圈在安全點驗版本後 merge；同步路徑讀穩定快照（補 #5） | 08 |
| event 抽取 | 不信 story 自報的 revelations_touched，程式碼層抽結構化 events 佐證（補 #3，非新 agent） | 07 |
| injection 防護 | 玩家輸入包 `<player_action>` 標籤 + skill 守則，不犧牲自由度（補 #7） | 07、skills |
| API key | keyring（非混淆），JS 永不碰 key，路徑 JS→Python→OpenRouter（補 #8） | 06、08 |
| MVP 收斂 | 拆 MVP-A（核心循環跑穩 30 beat）/ MVP-B（差異化），audio/dreaming/命運/配圖移後（採納 #9） | 05 |
| 測試驗收 | 10 項指標，最關鍵：防暴雷、30 beat 連貫、JSON 穩定（補 #10） | 08 |
| 流程圖修正 | beat 迴圈補入 orchestrator（warden→orchestrator→story，修 #1） | 02 |
| **補丁① Progress Kernel** | 世界狀態推進由程式碼層 kernel 決定（非 LLM），story 退化為 realizer；每 beat 強制 ≥1 狀態 delta；結局用 attractor。修「開門後又問開門」「NPC 不進場」「行動沒後果」。預設 ON、失敗回退 legacy | 10、02 |
| **補丁② Config Center** | story（及其他 agent）prompt 拆成可配置 fragment；決定性組裝 + 穩定 prompt_hash + 零 LLM preview；config-first → static fallback；draft→preview→activate UI；每場 run 存 config 快照。預設 OFF、失敗退 SKILL.md | 11 |
| **MVP-B 差異化** | 技能宣稱封頂（接受但加侷限/謎題化）；一種結局序列（純敘述收尾 + 完整真相揭露 + 復盤）；輕量 dreaming（在場 NPC 演化，只寫 evolving）；道具庫完整（held_by/不洩漏 key item）；防暴雷對抗性驗證報告 | 12 |

| 新增的待決事項（取代部分舊項，見 05）：
- 離場 NPC 的 dreaming 是否對玩家完全隱藏（傾向：是，資訊差製造恐怖）
- 自訂格式輸入的「格式」具體長什麼樣（需設計輸入模板）
