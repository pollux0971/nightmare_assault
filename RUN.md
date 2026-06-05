# 如何安裝與遊玩 Nightmare Assault

LLM 驅動的**無限恐怖文字冒險**：Python 後端 + 網頁前端（pywebview），透過 [OpenRouter](https://openrouter.ai) 呼叫大型語言模型即時生成世界、敘事、NPC 與結局。

> 需求：**Python 3.12+** ＋ 一把 **OpenRouter API key**（到 https://openrouter.ai 註冊取得，並儲值少量額度）。

---

## 1. 取得程式

```bash
git clone https://github.com/pollux0971/nightmare_assault.git
cd nightmare_assault
```

## 2. 安裝依賴

```bash
python3 -m venv .venv
.venv/bin/pip install -e .            # 核心依賴：pydantic / httpx / PyYAML / pywebview
.venv/bin/pip install -e ".[dev]"     # （可選）裝 pytest 跑測試
```

桌面視窗殼（pywebview）依平台：

- **Linux**：建議 Qt 後端 `.venv/bin/pip install "pywebview[qt]"`；
  或 GTK：`sudo apt install python3-gi gir1.2-webkit2-4.1` 後 `pip install pywebview`。
- **macOS / Windows**：`pip install pywebview` 即可（用系統內建 WebView）。

> 沒有桌面環境也能玩 —— 見下方第 6 節「**無頭 / 自動化遊玩**」。

## 3. 設定 API key

**第一次啟動會自動生成 `config/config.json` 框架（不含金鑰）。** 三種填法擇一：

1. **啟動後在遊戲「設定」畫面**貼上你的 key（最簡單，推薦）。
2. 或設環境變數：`export OPENROUTER_API_KEY=sk-or-v1-...`（Windows：`set OPENROUTER_API_KEY=...`）。
3. 或直接編輯 `config/config.json` 的 `"api_key"` 欄位。

> `config/config.json` 已被 `.gitignore` 排除，**你的金鑰不會被 commit / 上傳**。
> 模型清單可參考 `config/config.example.json`（已含建議的 8 個 agent 模型）；要換模型就改 `config.json` 的 `agent_models`。

## 4. 啟動

```bash
.venv/bin/python main.py
```

- 還沒填 key → 會停在**設定畫面**，貼上 key 即可。
- 已有 key → 直接進**主選單**。

## 5. 玩一場

1. 主選單 →「**新的夢魘**」。
2. 輸入**主題**（例：「廢棄的精神病院」「深海研究站」「末班列車」）、主角名、同行人數。
3. 「**進入世界**」→ 序章 →（setup 用較強模型生世界，約 10–30 秒）。
4. 進**遊戲主畫面**：
   - 敘事**逐字浮現**，關鍵詞（死/血/屍/黑暗…）泛血紅。
   - 敘述吐完 → 收束句淡入 → **選項**逐個淡入；也可在下方**自由打字**描述你想做的事。
   - 右上：🎒 道具、💬 與在場 NPC 對話、💾 存檔、≡ 回選單。
5. 認真**調查**（檢查紙條/文件/儀器、追問 NPC）會累積真相線索；觸發致命規則或抵達出口 → **結局**（真相揭曉 / 模糊逃脫…）。

## 6. 無頭驅動（CLI 自動化測試用，**非給玩家的玩法**）

> ⚠️ 這一節是給**開發 / QA / 自動化測試**用的 CLI 工具（`dev/tools/agent_play.py`），讓**程式或另一個 AI** 在沒有桌面視窗的情況下驅動遊戲、驗證機制與回歸。**這不是給人玩的方式**——真正要玩請走第 4–5 節的桌面視窗。

```bash
# 內建策略自我遊玩一場（煙霧測試 / 產逐字稿，非人工遊玩）
.venv/bin/python dev/tools/agent_play.py --auto --flag --max-beats 20 --jsonl-log run.jsonl

# 不需 API key 的介面冒煙（用假後端，給 CI 用）
.venv/bin/python dev/tools/agent_play.py --auto --no-llm --flag --max-beats 8

# 讓「另一個程式 / AI」用 JSON-over-stdio 協定自動測試（每行送一個動作，讀一行 observation）
.venv/bin/python dev/tools/agent_play.py --flag --jsonl-log run.jsonl --max-beats 20
```

旗標：`--flag` 開敘事控制（揭露橋接 / Player Sovereignty）｜`--no-llm` 不呼叫 LLM｜`--seed N` 固定隨機種子｜`--jsonl-log FILE` 落 observation/action/assertion（供自動化斷言）｜`--debug-reveal-truth` 除錯時才露完整真相。

## 7. 跑測試（可選）

```bash
.venv/bin/python -m pytest        # 全套件
```

## 疑難排解

- **視窗開不起來 / 沒桌面環境** → 用第 6 節的 `agent_play.py` 無頭跑。
- **`configured: false` 一直要 key** → 確認 key 已填、OpenRouter 帳號有額度。
- **某模型報錯（400 not a valid model）** → 到 `config.json` 把該 agent 的模型換成 OpenRouter 上存在的 slug。
- **額度用完（HTTP 402）** → 到 OpenRouter 儲值，或在 `config.json` 設較低 `max_tokens`。
- **一直轉/沒反應** → 多半是某次 LLM 呼叫慢或失敗；OpenRouter 偶有延遲，稍等或重試決策。

## 重點機制（想深入了解）

- **雙層 bible 防暴雷**：story / NPC 結構性看不到完整真相，只看「已揭露」的部分。
- **揭露橋接**：調查 → EvidenceEvent → reveal ladder（hidden→hinted→observed→suspected→confirmed）→ 結局復盤。
- **結局因果硬閘**：危險值達標不會直接致死；逃脫須先發現出口再明確提交；低真相逃脫顯示為「模糊逃脫」。

設計文件見 `nightmare-assault-design-fixed/`，開發作業系統見 `dev/`。
