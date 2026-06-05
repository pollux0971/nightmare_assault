# 如何親自測試 Nightmare Assault（MVP-A）

> 後端核心 + 網頁前端已完成。這份教你在自己的機器上把它跑起來、玩一場。
> **API key 已預植**（`config/config.json`，取自你的 `test_api.txt`，已 gitignore，不會外洩）。

---

## 1. 安裝（一次）

```bash
cd /data/python/nightmare-assault
# 已有 .venv；補裝桌面殼（pywebview）。Linux 建議 Qt 後端（pip 可裝，免系統套件）：
.venv/bin/pip install "pywebview[qt]"
# 已裝的核心依賴：pydantic / httpx（LLM）；pytest（測試）
```

> 若 Qt 裝不起來，替代方案：GTK 後端 `sudo apt install python3-gi gir1.2-webkit2-4.1` 後 `pip install pywebview`。
> macOS/Windows 直接 `pip install pywebview`（用系統內建 WebView）。

## 2. 啟動

```bash
.venv/bin/python main.py
```

會開一個深色視窗。因為 key 已預植，會直接到**主選單**（若想換 key 走「設定」）。

## 3. 玩一場（建議流程）

1. 主選單 →「**新的夢魘**」。
2. 輸入**主題**（例如「廢棄的精神病院」「深海研究站」「末班列車」），主角名、同行人數。
3. 「**進入世界**」→ 序章儀式動畫（恐怖台詞浮現）→ 約 10–30 秒（setup 用較強模型生世界）。
4. 進**遊戲主畫面**：
   - 敘事**逐字浮現**，關鍵詞（死/血/屍/黑暗…）會泛血紅。
   - 敘述吐完 → 情境收束句淡入 → **選項逐個淡入**。
   - 也可在下方**自由打字**描述你想做的事（按 Enter 或送出）。
5. 每個決策推進一個分鏡；右上 🎒 看道具、💾 存檔、≡ 回選單。
6. 若觸發致命規則 → 結局畫面。

## 4. 重點觀察（MVP-A 六項驗收，你可親驗）

- **連續多 beat 不崩**：玩 10+ 個分鏡，敘事是否連貫。
- **防暴雷**：故事不會提早講出你還沒發現的真相（雙層 bible 結構性保證）。
- **JSON 穩定**：偶爾 LLM 格式跑掉時，仍會給出可玩的決策（後端三級 repair + fallback）。
- **自由輸入有回應**：打非選項的行動，故事會接。
- **存讀檔 / 道具**：💾 建存檔點、🎒 看道具庫。
- **換視窗大小不跑版**：拉大縮小視窗，版面自適應。

## 5. 疑難

- **一直轉/沒反應**：多半是某次 LLM 呼叫慢或失敗；狀態列會顯示「夢魘正在成形…」或錯誤。OpenRouter 偶有延遲，稍等或重試決策。
- **開局失敗**：通常是 key 或網路；確認 `config/config.json` 的 key 正確、能連 openrouter.ai。
- **想換模型**：編輯 `config/config.json` 的 `agent_models`（預設 setup=claude-3.5-sonnet、story=gemini-flash-1.5、其餘=gpt-4o-mini）。

## 6. 不用開視窗也能驗（無頭）

```bash
.venv/bin/python -m pytest -q          # 全部單元/整合測試（約 380+，含防暴雷/30beat/串流契約）
```

> MVP-A 不含：完整 dreaming、離場命運、聊天室、多結局、音訊、配圖、分支樹（那些在 MVP-B / 後續階段）。
