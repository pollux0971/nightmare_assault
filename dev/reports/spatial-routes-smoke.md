# Spatial Routes Projection — Focused Smoke（commit `25ab123`）

> 真 LLM（deepseek-chat-v3-0324），flag NARRATIVE_CONTROL + OPENING_VARIATION ON。**只驗證，未改 code。**
> 主題：午夜的廢棄海事研究站（多相連艙區）。逐 beat 資料：`spatial-routes-smoke.jsonl`。
> 零 `world model tick skipped`、零 traceback、零「沒有明顯可走的出口」。

## 結論：可走路線真的有東西了 ✅

整場移動 entry → 主走廊 → 研究實驗室 → （撤退）安全區，每一拍 `可走路線` 都列出**可用的結構性 route**，
不再是「沒有明顯可走的出口」。

## 逐 beat（spatial_summary 的「可走路線」）

| beat | current_area | previous | mode | 可走路線 |
|---|---|---|---|---|
| 開場 | maintenance_entry | — | — | 暫退到安全區整理 |
| b2 | main_corridor | maintenance_entry | active | **返回上一個區域（維修入口艙）**、暫退到安全區整理 |
| b3 | main_corridor | maintenance_entry | active | 返回上一個區域（維修入口艙）、暫退到安全區整理 |
| b4 | research_lab | main_corridor | active | 返回上一個區域（主走廊）、**返回現場（維修入口艙）**、暫退到安全區整理 |
| b5 | **area.safe_zone** | research_lab | **review_mode** | 返回上一個區域（研究實驗室） |
| b6 | area.safe_zone | research_lab | review_mode | 返回上一個區域（研究實驗室） |
| b7 | area.safe_zone | research_lab | review_mode | 返回上一個區域（研究實驗室） |

## 對照 6 項檢查

| # | 檢查 | 結果 | 證據 |
|---|---|---|---|
| 1 | 移動到新區域 → 「返回上一個區域」 | ✅ | b2（→維修入口艙）、b4（→主走廊）、b5（→研究實驗室）皆顯示 `route.return_previous`，標籤帶正確的上一區名 |
| 2 | active area → 「暫退到安全區整理」 | ✅ | 開場 / b2 / b3 / b4 皆有 `route.withdraw_safe`（→area.safe_zone） |
| 3 | safe_zone → 「返回現場」/ return_to_site | ✅（含標籤註記） | b4（在 research_lab）明確顯示 `route.return_site`「返回現場（維修入口艙）」；b5–b7 在 safe_zone 仍有回到調查現場的 route，但因該現場＝上一區（research_lab 被標 active_area）而 **dedup 成「返回上一個區域」**——能力在、字面標籤不同（見下註） |
| 4 | campaign_exit 只在合理位置 | ✅ | 本場無 `campaign_exit` role 的 area → campaign_exit route **從未出現**（只在 role 存在時提供） |
| 5 | locked/blocked 只在 blocked，不混入可走 | ✅（未現場觸發） | 本場未出現任何 locked/blocked exit；所有可走 route 皆 state=available 的結構性 route。locked→blocked 由單元測試 `test_locked_exit_blocked_not_available` 決定性保證 |
| 6 | 無可推導 route 才顯示「沒有明顯可走的出口」 | ✅ | 全場每拍都有 ≥1 結構性 route → **從未**顯示該 marker |

## 附帶觀察

- **撤退確實生效**：b5「往出口方向退開危險」→ current_area 真的切到 `area.safe_zone` **且** mode 翻成
  `review_mode`。本場 UX 既知 #6（撤退不翻 mode）**未重現**——撤退→safe_zone+review 正常運作。
  撤離鎖具黏著性：b6/b7 的「往外探索」未脫離 safe_zone（review 持續），符合 durability 設計。
- `area.safe_zone` 是 role fallback id：玩家**實際撤退時**由 loop 建成真 area（投影本身未建 entity，符合唯讀）。

## 誠實註記（檢查 3 的標籤 dedup）

在 safe_zone（b5–b7），「回到調查現場」這條 route **存在**（to_area=research_lab，即撤退前的 active area），
但因為該現場恰好＝previous_area，dedup 規則保留了 `return_previous`，故字面顯示「返回上一個區域（研究實驗室）」
而非「返回現場」。能力與目標都正確（玩家能回到現場），只是 safe_zone 情境下的**標籤可更貼切**——
若想讓 safe_zone 一律顯示「返回現場」字樣，可在 dedup 時優先採 `return_site` 標籤（屬 polish，非 bug，
本輪未改 code）。對照 b4（現場≠上一區）已明確顯示「返回現場（維修入口艙）」，證明 return_site 機制本身正常。

## 一句話

Spatial Routes Projection 在真 LLM 遊玩中**確實改善**了 spatial_summary：移動後出現「返回上一個區域」、
active area 出現「暫退到安全區整理」、深入後同時出現「返回現場」、safe_zone 提供回到現場的 route；
campaign_exit 僅在 role 存在時出現；全程不再有「沒有明顯可走的出口」。唯一可 polish 處是 safe_zone 下
return_site 與 return_previous 同目標時的標籤選擇。
