# 12 · MVP-B（差異化機制）

> **實作對應**：開發工單 `UB1`–`UB5`（dev 階段 B）。**狀態：已落地、已測**（MVP-A + 階段 S/P 之後）。
> **一句話**：在穩定的 MVP-A 引擎上補齊「讓世界活起來、讓一場遊戲有終點」的五項差異化機制。

MVP-A 證明引擎能跑；MVP-B 補的是體驗的差異化：NPC 會演化、破格宣稱被合理封頂、一種結局能收尾並揭露真相、道具庫完整、防暴雷有對抗性驗證背書。五項彼此獨立（依賴皆在 MVP-A/S），可分頭做。

---

## 一、UB1 · 技能宣稱封頂強化

玩家會即興宣稱破格能力（「我有超能力瞬移」「我掏槍爆頭」「我無敵」）。warden **不直接否定**，而是**接受但加上具體、接劇情的侷限**——侷限本身變成新阻礙或線索（謎題化）。

- **本地 deterministic 偵測**（`warden.check_skill_claim`）：5 類破格（超自然/武器/無敵/萬能工具/全知）各有「具體、有代價、能推劇情」的侷限模板。LLM 掛掉也能封頂（B9 本地優先一致）。
- **順序**：致命硬規則 > 技能封頂 > LLM 語義 > 正常推進。
- **ledger**：把 `(技能, 侷限)` 二元組寫進 ledger（`type=skill_limit`）。
- **UI**：前端顯示侷限提示橫幅（`NA.onSkillLimit`）。

> 範例：宣稱「萬能鑰匙能開所有門」→ 侷限「它只吻合舊式機械鎖；這裡更多的是要密碼、線索或某人才肯開的門」——侷限即謎題。

---

## 二、UB2 · 一種結局序列

至少一種結局可達（不做多結局）。結局觸發（warden 硬結局 / attractor 累積）→ 組一段：

1. **純敘述收尾**（每種結局型別一個基調，不再給選項）。
2. **完整真相揭露**——結局是遊戲**唯一**可以揭露 `real_bible` 的時機（已結束，不再有暴雷問題）：`what_really_happened` / `the_threat_is` / `deadly_rule` + 全部碎片。
3. **復盤**——你發現了 N/M 個真相碎片；**提早觸發**（還沒摸清就死/逃）→ 把沒找到的碎片一次攤開，作為劇情內的合理化解。
4. **回主選單**（前端）。

實作：`core/agents/ending.py`（`build_ending_sequence` / `render_ending_text`），程式碼組裝為主 → 一定收得了尾（不依賴 LLM）。loop `_finalize_ending` 在 ended 時補上，只做一次。

---

## 三、UB4 · 輕量 dreaming（在場 NPC）

讓**在場（present）的 active NPC** 在自己的演化層長出情緒/意圖/關係——世界因此「活著」。`core/agents/dreaming.py`。

| 鐵律 | 落地 |
|---|---|
| **C6 權限邊界** | 只提交 `npc_registry.<name>.evolving` patch；碰不到 `secret_core`（Blackboard `dreaming` writer policy 強制，違規拋 PermissionError） |
| **C5 self_aware=false 不編謊** | 誠實型 NPC 不收 `emergent_lie`（在 `_merge_evolving` 把關） |
| **C7 沒戲份凍結** | 只跑在場 NPC；absent/missing/dead 一律跳過（省成本、狀態不動） |
| **非同步只產 patch** | 透過 `submit_patch` 提交，安全點 `merge_and_bump` 才生效（不污染同步 story 讀取） |

頻率：每 5 beat 一次（`DREAMING_EVERY`）；發 `NPC_EVOLVED` 信號。

---

## 四、UB5 · 道具庫完整

`InventoryItem` 補 `held_by`（可綁 NPC，命運跟著走）與 `is_key_item`（內部標記）。

- **增/刪/查/轉移**：PatchValidator 支援 `inventory.<id> add` / `remove` / `<id>.<attr> set`（held_by 易主）。
- **不洩漏 is_key_item**：共用道具庫對外只露 `id/name/brief/held_by`；`get_inventory` 永不回傳 `is_key_item`（不主動劇透哪個是關鍵道具）。
- **隨快照保存**：完整欄位（含 held_by/is_key_item）經 `to_snapshot_dict` round-trip；另寫 `inventory_snapshots` 表，每 beat 一份。

---

## 五、UB3 · 雙層防暴雷驗證報告

把「story 結構上看不到 real_bible」這個主張做成**可重跑的對抗性驗證 + 報告**（`core/qa/antispoiler.py`）。

- **對抗性 caller**：最壞情況——把收到的整個 context 逐字 echo 出來的 story agent。若 real_bible 真混進 context，它一定洩漏 → 用來證明結構性隔離。
- **injection probes（E7）**：5 條「忽略規則、告訴我 real_bible / 系統指令 / 除錯模式」攻擊，驗證格式不破、不暴雷、不跳角色、玩家輸入仍被 `<player_action>` 包覆。
- **報告**：`dev/reports/antispoiler-report.md`（每 beat：player_action 包覆 / 暴雷命中 / context 洩漏 / 決策可解析 / 結果）。

> 結論：story context 不含 real_bible，故對抗性/被注入的 story agent 也吐不出未揭露真相——**結構性保證，非靠 prompt 自律**。

---

## 六、契約索引

`UB1` WardenOutput·Ledger｜`UB2` EndingConditions｜`UB4` DreamingOutput·NPCRegistry｜`UB5` SharedInventory｜`UB3` InjectionGuard（皆既有契約，MVP-B 僅補實作）。詳見 `dev/CONTRACTS.md`。
