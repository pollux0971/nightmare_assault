# 15 · Player Sovereignty — 從「故事收束型」轉向「開放式恐怖探索型」

> **實作對應**：`docs/player-sovereignty-principles.md`（原則）+ `core/narrative/exit_resolver.py` /
> `negative_intent.py` / `world_facts.py` + loop WorldProgress + `dev/tools/agent_play.py` 觀測。
> **狀態**：P0 已落地、已測（740 passed，flag OFF/ON 各一次）；WorldModel（下節）為**規劃中**的下一層。
> **一句話**：**你要強制的是「世界要回應玩家」，不是「劇情要走向主線」。**

---

## 〇、定位轉變（最重要）

Nightmare Assault 是**開放式恐怖探索遊戲**，不是分支劇情遊戲。

```text
舊核心驗收：玩家有沒有推進主線真相？
新核心驗收：玩家做的事，有沒有留下「可檢查的世界後果」？
```

之前把「逃不出去 / 罐頭 fallback / 強制 ending / 0/7 像失敗畫面」當成各自的 bug 去修；
真正的根因是**系統假設「故事要收束」**，把玩家的自由行動誤解成「偏離劇情」。

### Player Sovereignty 七原則（所有 agent / gate 的共同前提）

1. 玩家可拒絕主線：探索、撤退、繞路、破壞、躲藏、誤判、放棄調查。
2. 系統不得把玩家硬拉回唯一主線。
3. 系統只保證「行動有後果」，不保證「玩家接近真相」。
4. 真相揭露是探索的結果，不是系統強制派發。
5. 結局只由**玩家明確確認**或**不可逆後果**觸發，不得由敘事節奏 / 吸引子自動收束。
6. 玩家說「離開」預設不直接 ending；不確定就**問**（ExitResolver）。
7. `0/X` 真相不是失敗，而是**低資訊結局**（不渲染成 fail screen）。

> Narrative Control 的職責收斂為三件：①防系統矛盾 ②防行動沒回饋 ③防真相/結局/NPC lore 失控。
> **不**負責：強迫進主線 / 強迫揭露 / 強迫進結局 / 強迫對話 / 把故事拉回某條 plot path。

---

## 一、P0 已落地（runtime 行為，不加故事內容）

### ExitResolver（離開意圖，偏向「問」不偏「猜」）
`core/narrative/exit_resolver.py`：玩家說離開 → 分意圖
`area_transition / temporary_retreat / safe_zone_reached / return_to_motive`（皆**續行不結局**）、
`run_ending`（= campaign_end，**唯一**進 EndingGate）、語意不明 → **ExitOffer 四選一**（永遠含「結束本次調查」這個不被困的出口）。
flag ON 時 **attractor 不再自動收束**；結局只來自玩家明確結束或 warden 硬觸發（不可逆）。

### NegativeIntentGuard（explicit 否定優先於 keyword）
`core/narrative/negative_intent.py`：
- `negates_ending`：「**不**結束本次調查」不得因含子字串「結束本次調查」被誤判成 ending。
- `negated_targets`：「不進 B 區」→ kernel 過濾「移動到 B 區」的候選事件。

### WorldStateFact（NPC/story 有用資訊 → 可檢查事實，不必是 truth）
`core/narrative/world_facts.py`：`known_exit_locked / generator_needed / machine_room_known /
<npc>_confirmed_present`。**沒有 truth reveal 但有 world_fact 的 beat，也算有進展。**

### WorldProgress 觀測（讓 AI 在遊玩中就抓到問題）
observation 新增 `world_state`（scene/danger/clue/item）、`world_progress`（current_area / known_areas /
changed_exits_this_beat / new_world_facts_this_beat / investigation_state / available_next）、
`deltas`（**had_consequence** 為核心）+ 停滯計數；內建斷言 **`WorldResponds`**（行動無後果直接 fail）/
`NotStuckInScene` / `DangerProducesThreat`。

### 0/X = 低資訊結局
`render_ending_text`：0 發現 → 「你對這裡發生的事，幾乎一無所知」，不再列 ？？？ 責備玩家。

---

## 二、下一層（規劃中）：WorldModel — 抽象的「物件 / 實體」機制

實機 selfplay 暴露的問題（系統把玩家往深處推、世界記不住敘述過的袖扣/筆記本、「退到外面」沒真的移動、
NPC 講的話蒸發）**不是各自的 bug，而是同一個缺失機制的表現：沒有一個「世界裡可被指涉的東西」的模型。**

之前的 `world_facts`（扁平 string→fact）與字串比對否定，是這個缺口的**權宜替身**。正解是建一個
**主題無關**的實體模型，把移動、物件、出口、事實、NPC、否定、撤退、「世界記得」全部收成它的投影：

```text
Entity: id / kind(object|area|exit|actor|fact) / label / state / props / affords / origin
狀態機（抽象，不含主題內容）：
  object: unseen→noticed→inspected→taken/used
  exit  : unknown→known→{available|locked|blocked}→used
  area  : unknown→known→visited→current
  actor : unknown→present/absent→talked
  fact  : asserted
```

**核心轉變：玩家行動 = 對某個 Entity 套用一個 affordance**（不再猜關鍵詞 / 比字串）。

| 現有問題 | 在實體模型裡的自然解 |
|---|---|
| stay-put 違規 | 「只在原地」= 無 `move_to` 目標 → 不 fire 移動；「不進 B 區」= **拒絕 `move_to(area.B)`** |
| 世界記不住 | story 敘述袖扣 → 登記 `object.cufflink(noticed)`；之後可 `inspect` → 世界記得 |
| 退到外面沒移動 | 「外面」是 `area.outside` 實體 → `withdraw_to` 真的改 current_area |
| NPC 資訊蒸發 | NPC 一句 → `fact`（+ 可選 `area(known_unvisited)`）實體 |
| recap 標題空白 | 真相碎片就是 `fact` 實體、自帶 label |
| had_consequence 用猜的 | 套 affordance **必然**改某實體 state → 「有後果」變**結構保證** |

`world_progress` 觀測 = 這個模型的投影；`available_next` = 當前區域實體的 affordances。
填充：kernel 圖種 area/exit；EntityExtractor 掃 story/NPC 輸出登記實體（機制抽象，主題內容 runtime 由 LLM 給）；
理想上 story/NPC 回**結構化 entity_delta**，抽取只是 fallback。

> 建置路徑：先立 `core/world/model.py`（Entity + 狀態機 + affordance + store + intent→affordance resolver），
> 再把 kernel 移動 / ExitResolver / NegativeIntentGuard / world_facts / observation **全部改成它的投影**——
> 不是再補特例，而是把特例收進同一個抽象機制。

---

## 三、玩法分級（皆為有效玩法，非好壞）

```text
0/X：玩家逃了，但一無所知        ← 低資訊結局（合法）
2/X：玩家知道一點，但可能誤判
4/X：玩家帶著部分真相離開
X/X：玩家理解全貌，能做高代價選擇
```

錯的不是 `0/X`，錯的是**「玩家明明調查了，卻 0/X」**（調查無回饋）。

> 玩家可以不追真相，但世界必須記得他做過什麼。
> 結局不是劇情節奏的產物，而是玩家選擇或不可逆後果的產物。
