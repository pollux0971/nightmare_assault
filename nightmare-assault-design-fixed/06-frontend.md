# 06 · 前端（網頁 + pywebview）

> 前端技術規劃。設計原則與互動內容見 04；本檔講「用網頁技術怎麼實作」。
> 後端維持 Python（core 一行不改）；前端用 HTML/CSS/JS，由 pywebview 包成桌面應用。

---

## 一、為什麼是網頁而非原生 GUI

這個遊戲的體驗命脈在**文字呈現的質感**：逐字浮現、關鍵詞漸變血紅、句間呼吸停頓、選項淡入、畫面明暗呼吸，加上之後的圖片與精緻動畫。這些是**動畫與排版**的活——CSS/JS 的母語，卻是原生 GUI（customtkinter/tkinter）的逆水行舟。

更關鍵的是**跨電腦一致性**（不跑版）：

| | 原生 GUI | 網頁技術 |
|---|---------|---------|
| 字體 | 吃系統已裝字體（別人沒裝就歪） | `@font-face` 內嵌，任何電腦同款字體 |
| 布局 | 寫死像素，換解析度/DPI 易歪 | rem/vw/flexbox/grid 自適應流動 |
| 縮放 | 受系統 DPI 設定影響 | CSS 可控、可鎖窗口最小尺寸 |
| 動畫 | 幾乎要逐幀 hack | CSS animation/transition 原生 |
| 圖片 | 勉強 | 原生強項 |

**網頁技術正是為「同一頁面在任何裝置都不歪」而生的**——這恰好解決跨電腦跑版的擔憂，而原生 GUI 反而更容易因系統字體/DPI 差異而歪。

---

## 二、整體架構：pywebview 橋接

```
┌─────────────────────────────────────────────────┐
│  桌面視窗（pywebview）                            │
│  ┌───────────────────────────────────────────┐  │
│  │  WebView（系統內建引擎）                    │  │
│  │  - Windows: Edge WebView2                  │  │
│  │  - macOS:   WebKit                         │  │
│  │  - Linux:   WebKitGTK                      │  │
│  │                                            │  │
│  │  前端：HTML + CSS + JS                      │  │
│  │  （文字畫布、決策、動畫、音訊播放）         │  │
│  └──────────────┬────────────────────────────┘  │
│                 │ pywebview 的 JS↔Python 綁定      │
│  ┌──────────────┴────────────────────────────┐  │
│  │  Python 後端（你的 core，一行不改）         │  │
│  │  agents · Blackboard · 記憶 · 持久化        │  │
│  └────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
        單一程序，pywebview 把兩邊綁在一起
```

pywebview 提供 `window.pywebview.api.xxx()` 讓 JS 直接呼叫 Python 函式；Python 可用 `window.evaluate_js()` 推資料給前端。比 Go↔Python 那座跨程序橋簡單得多——同程序、有現成綁定。

```python
# 後端暴露 API 給前端
class API:
    def start_game(self, theme, npc_count, ...): ...
    def submit_decision(self, text): ...        # 回傳 / 觸發串流
    def get_inventory(self): ...
    def save_game(self): ...

webview.create_window("Nightmare Assault", "ui/index.html", js_api=API())
webview.start()
```

---

## 三、串流：UI 不凍結（仍是核心）

即使換成網頁，「LLM 串流會阻塞數秒」的問題還在——但 web 的處理比原生 GUI 自然得多。

```
方案：後端串流 → Python StreamParser 分類事件 → JS 即時渲染

後端（背景執行緒跑 LLM，不卡 pywebview 主迴圈）:
  parser = StreamParser()
  for raw_token in story_stream():
      for event in parser.feed(raw_token):
          if event.type == "NARRATIVE_CHUNK":
              window.evaluate_js(f"NA.appendToken({json.dumps(event.text)})")
          elif event.type == "CONTINUE_PAUSE":
              window.evaluate_js("NA.onContinue()")
  decision = parser.finalize()  # 已完成 JSON repair + Pydantic 驗證
  window.evaluate_js(f"NA.onDecision({decision.model_dump_json()})")
  window.evaluate_js("NA.onBeatComplete()")

前端 JS:
  NA.appendToken = (text) => streamWithRhythm(text);  // 只渲染純文字
  NA.onContinue = () => showContinueButton();
  NA.onDecision = (decision) => renderDecision(decision);
  NA.onStatus = (status) => updateStatus(status);
  NA.onError = (err) => showRecoverableError(err);
```

**分隔符與 JSON 解析只在 Python 後端做**。JS 不解析 `<<<CONTINUE>>>` / `<<<DECISION>>>`，也不修 JSON；前端只接收後端已分類、已驗證的事件。節奏控制仍在 JS 端（比 after() 更自然）：用 `setTimeout`/`requestAnimationFrame` 控制吐字速度、關鍵詞放慢、句末停頓。CSS transition 處理淡入。

注意：LLM 呼叫仍要放 Python 的背景 thread，避免阻塞 pywebview 的主迴圈。

---

## 四、視覺風格：深色現代恐怖（CSS）

CSS 做這套風格輕鬆愉快，全是它的母語。

```css
:root {
  --bg:        #0A0A0C;   /* 近黑，非純黑，留空間感 */
  --bg-sub:    #14141A;   /* 卡片、輸入框 */
  --text:      #C8C8D0;   /* 灰白主文字，非純白 */
  --blood:     #8B1A1A;   /* 血紅點綴：hover/關鍵詞/危險，克制使用 */
  --border:    #3A3A45;
}

/* 襯線字體內嵌（跨電腦一致的關鍵） */
@font-face {
  font-family: 'NotoSerifTC';
  src: url('assets/fonts/NotoSerifCJKtc-Regular.woff2') format('woff2');
}
body { font-family: 'NotoSerifTC', serif; background: var(--bg); color: var(--text); }
```

你要的每個效果都是 CSS 幾行：

| 效果 | CSS/JS 手段 |
|------|------------|
| 逐字浮現 | JS 逐 token append + opacity transition |
| 關鍵詞血紅 | `<span class="blood">` + color transition |
| 句末停頓 | JS setTimeout 控制下個字延遲 |
| 選項淡入 | CSS `@keyframes fadeIn` + 逐個 animation-delay |
| 明暗呼吸 | CSS animation 在背景層做 brightness 循環 |
| 配圖淡入 | `<img>` + opacity transition（位置預留，後接） |

血紅是**點綴非主色**：整體陰沉灰黑，血紅只在關鍵時刻刺出。

---

## 五、畫面流程（單頁應用，切換 view）

不必多頁面，用一個 `index.html` + JS 切換顯示區塊（或輕量路由）。

```
啟動
 → 啟動畫面（標題 + 氛圍）
 → [首次/無設定] 設定畫面（API Key + 三模型 + 各 agent 搭配 + 圖片/音樂模型 + 連線測試）
 → 主選單（新遊戲 / 繼續 / 設定）
     ├─ 新遊戲 → 新局設定（主題/主角/NPC 數/自訂或預設角色/難度）→ 載入（setup 跑）→ 遊戲主畫面
     ├─ 繼續   → 存檔選擇 → 遊戲主畫面
     └─ 設定
 → 遊戲主畫面（核心迴圈）
     └─ 觸發結局 → 結局畫面（結局串流 + 完整真相揭露 + 復盤）→ 回主選單
```

每個畫面是一個 `<section>`，JS 控制 show/hide + 過場淡入。MVP-A 可先不做結局畫面；結局畫面屬 MVP-B，當至少一種結局序列可達後再接上。

---

## 六、遊戲主畫面版面（HTML 結構）

```html
<div id="game">
  <header>📍場景名 · 第N分鏡   <button>🎒</button><button>💾</button><button>≡</button></header>
  <div id="scene-image"><!-- 配圖預留，先空 --></div>
  <div id="narrative"><!-- 逐字浮現的敘事，可捲動回看 --></div>
  <div id="decision">
    <p id="recap">情境收束句</p>
    <div id="options"><!-- 動態 button，敘述吐完才淡入 --></div>
    <div id="free-input">
      <input placeholder="或描述你想做的事…"/><button>送出</button>
    </div>
  </div>
</div>
```

| 區域 | 元素 | 說明 |
|------|------|------|
| 頂部列 | `<header>` | 場景、beat 編號、道具/存檔/選單 |
| 配圖預留 | `#scene-image` | `<img>`，先空，後接 |
| 敘事區 | `#narrative` | 逐字 append，overflow 捲動看歷史 |
| 決策區 | `#decision` | 選項逐個淡入；危險選項 class blood |
| 自由輸入 | `#free-input` | 永遠可用，Enter 送出 |

子視窗用 modal（`<dialog>` 或覆蓋層）：道具面板、聊天室（Phase 2）、存檔/選單。

---

## 七、一個 Beat 的前端時序

```
玩家點選項 / 打字送出
  │
  ├─[JS] disable 決策區 + 輸入框（防連點），顯示敘事性 spinner 文案
  ├─[JS] window.pywebview.api.submit_decision(text)
  │
  ├─[Python 背景 thread] warden→orchestrator→story，串流 token
  │       raw token → Python StreamParser
  │       narrative event → window.evaluate_js("NA.appendToken(...)")
  │       continue event  → window.evaluate_js("NA.onContinue()")
  │       decision event  → window.evaluate_js("NA.onDecision(validatedJson)")
  │
  ├─[JS] 已分類事件處理：
  │       NA.appendToken → streamWithRhythm（30-50字/秒，關鍵詞放慢+blood class）
  │       NA.onContinue  → 浮現「繼續 ▼」，點擊後呼叫 continue_narration()
  │       NA.onDecision  → 呈現後端已驗證的決策 JSON
  │
  ├─[JS] 決策呈現：
  │       敘述吐完 → 500ms 停頓 → recap 淡入
  │       選項逐個淡入（animation-delay 300ms 遞增）
  │       解鎖輸入框、聚焦
  │
  └─ 等玩家 → 迴圈

  ⇣ 同時（玩家讀字時）：compactor/dreaming/offstage-fate 在 Python 背景 thread 跑
```

---

## 八、跨電腦不跑版的三個保證

呼應你的核心擔憂：

1. **字體內嵌**：思源宋體 `.woff2` 打包進 `assets/fonts/`，`@font-face` 載入。不靠對方系統有沒有裝。
2. **相對單位 + flexbox/grid**：不寫死像素，版面自適應解析度與窗口大小，不會「這台好好的那台擠出來」。
3. **pywebview 鎖窗口**：設最小尺寸避免壓到變形；各平台用系統 webview，行為可預期。

---

## 九、音訊整合（前端側，JS）

音訊系統演算法見 02 第十五節。前端負責播放混音，web 的音訊 API 很成熟。

```
JS 音訊層（Web Audio API 或 Howler.js）:
  音樂層:   依 beat 的 pacing 切換音軌 + crossfade（淡入淡出）
  環境音層: 循環疊加（滴水/腳步/嗡鳴/心跳）；心跳隨 pacing 加快
  事件層:   audio_cue 觸發（silence 驟靜 / sting 驚嚇 / swell 漸強）

動態音量:
  角落音量控制（呼應「害怕就調小」提示）
  進階：偵測久未互動 → 自動略降音樂喘息
```

**MVP 音訊**：用預製免費恐怖 ambient 素材，先把分層播放、crossfade、pacing 切換、心跳加速的**機制**做對。生成音軌（Phase 3）是後面替換進來的。

Web Audio API 的 crossfade、即時音量、多軌混音都是原生能力——比原生 GUI 接音訊庫順暢。

---

## 十、等待動畫（CSS，氛圍式）

web 做動畫毫無障礙，但恐怖遊戲的等待仍應是**氛圍延伸非花俏旋轉**：

```
- 恐怖台詞緩慢浮現又淡去（CSS opacity transition）
  「當你凝視深淵，深淵也在凝視你」
- 背景極輕微明暗呼吸（CSS animation brightness 循環）
- 血紅一點在黑暗中明滅
- 體貼提示穿插：「如果害怕，可以把音樂聲調小」
```

兩個關鍵等待時機：
1. **序章生成**（setup + 基本音軌，較久）：做成「進入世界」的儀式——漸強環境音 + 緩慢開場白 + 恐怖台詞。第一印象。
2. **每 beat 生成前**（數秒）：接住上一個選擇的敘事性 spinner 文案。

---

## 十一、API 配置畫面

第一次啟動檢查 `config/config.json` 是否存在且有效。無 → 配置；有 → 直接開始。

```
配置欄位（存 config/config.json）:
  api_key:        OpenRouter API Key（遮蔽顯示）
  model_base_url: 模型 API URL
  agent_models:   8 個 agent 各搭配的模型
                  setup/orchestrator/story/warden/npc-chat/dreaming/offstage-fate/compactor
  image_model:    圖片生成模型（先存著，功能後接）
  music_model:    音樂生成模型（先存著，MVP 用預製素材）

每個 agent 模型旁「測試這個搭配」按鈕:
  呼叫 Python 背景打最小請求，回報成功/失敗
  避免玩到一半某 agent 掛掉
```

API Key 不用「混淆」當安全方案：開發版用 `.env`，桌面版用系統 keyring。請求路徑 `JS → pywebview API → Python → OpenRouter`，**key 永不到前端**（絕不 JS 直連 OpenRouter）。細節見 08 第四節。

---

## 十二、前端狀態同步與錯誤呈現

前端以後端 `get_game_state()` / `NA.onStatus(status)` 為唯一狀態來源，不靠按鈕本地猜測。

```javascript
const GameState = {
  IDLE: "idle",
  SETTING_UP: "setting_up",
  GENERATING: "generating",
  AWAITING_CONTINUE: "awaiting_continue",
  AWAITING_DECISION: "awaiting_decision",
  SAVING: "saving",
  LOADING: "loading",
  ERROR: "error"
};
```

基本規則：
- `generating / setting_up / saving / loading`：所有決策輸入 disabled，防止重入。
- `awaiting_continue`：只顯示「繼續」，呼叫 `continue_narration()`。
- `awaiting_decision`：顯示選項與自由輸入。
- `error`：顯示可恢復錯誤；若後端已降級成 fallback decision UI，前端回到 `awaiting_decision`。

`NA.onError(error)` 不應直接讓遊戲崩潰；優先顯示「系統短暫失真」風格的可玩提示，除非是 API key / setup 失敗等不可恢復錯誤。

---

## 十三、技術選型與打包

```
前端:
  純 HTML/CSS/JS（不必上框架，遊戲前端結構簡單）
  若想要元件化，可選輕量的 Alpine.js / petite-vue（非必須）
  音訊: Web Audio API 或 Howler.js
  動畫: 純 CSS（animation/transition）

橋接:
  pywebview（pip install pywebview）

後端:
  你的 Python core（不變）

打包:
  pywebview 官方建議 PyInstaller
  前端資源（html/css/js/字體/音訊）打包進去
  各平台需對應 webview runtime（Win 的 WebView2 通常已內建）
```

---

## 十四、前端 Epic（取代舊 customtkinter 版，併入 05）

| Epic | 內容 | 階段 |
|------|------|------|
| F1 pywebview 骨架 | 視窗、JS↔Python 綁定、view 切換、CSS 主題 | MVP |
| F2 串流渲染 | Python 背景 thread + StreamParser → JS 接收已分類事件、逐字渲染 + 節奏 | MVP-A（承重牆） |
| F3 遊戲主畫面 | 敘事區、決策區、自由輸入、頂部列（HTML/CSS） | MVP |
| F4 前置畫面 | 啟動/設定/主選單/新局/載入 | MVP |
| F5 道具面板 | modal 道具庫 | MVP |
| F6 存檔 UI | 存檔選擇 | MVP |
| F7 結局畫面 | 結局串流 + 真相揭露 + 復盤 | MVP-B |
| F10 等待動畫 | CSS 氛圍動畫 + 恐怖台詞 + 序章儀式 | Phase 2 |
| F11 音訊播放 | Web Audio 分層 + crossfade + pacing + 環境音 + audio_cue | Phase 2（預製素材） |
| F8 聊天室 UI | modal 多輪對話 | Phase 2 |
| F13 動態音量 | 手動 + 久未互動自動降 | Phase 2 |
| F9 配圖 | scene-image 顯示生成圖 | Phase 3 |
| F12 音樂生成接入 | 背景分層預生成、音軌庫管理 | Phase 3 |

F2 是前端承重牆，但它只負責「渲染與節奏」，不負責解析 LLM 原始格式。分隔符偵測、JSON repair、DecisionPoint 驗證由後端 `StreamParser` 統一處理；前端與後端 story/parser 聯調時，只測 `NA.appendToken/onContinue/onDecision/onStatus/onError`。
