# 17 · Spatial WorldModel Projection — 確定性空間投影 + 玩家面摘要

> **承接**：`16-worldmodel.md`（WorldModel 是唯一事實來源）。本檔記錄它的**空間投影層**。
> **實作對應**：`core/world/spatial.py` + WorldModel `version_snapshot` / `exits_from` / `is_safe_zone` +
> loop `spatial_debug()` + agent_play observation。
> **狀態**：P0–P4 已落地、已測（`test_spatial_projection.py` 20 + `test_spatial_summary.py` 11；全測 **847 passed**）；
> P5（餵 story/review agent context）**未接**（刻意）。
> **一句話**：**WorldModel 是事實來源；Spatial Projection 是它的確定性衍生快取——不呼叫 LLM、遊戲迴圈永不等它。**

---

## 〇、決策：能不能跟 WorldModel 並行？

能——但**只能是衍生投影層，不能變成 LLM 生成的地圖**。

```text
玩家行動
→ WorldModel 同步套用權威 delta（< 1–5 ms）
→ 快速空間投影同步從 WorldModel 算出（< 1–10 ms，無 LLM）
→ 沉重的 mental-map 文字用快取 / 確定性模板（async LLM 潤飾為選用，遲到就用 fallback）
→ story/NPC 收到的是最新有效投影，不是阻塞的生成結果
```

> **不要**：每個 beat 叫 LLM 重畫世界地圖再等它——又慢又不一致又會幻覺。

---

## 一、投影契約（P0，唯讀）— `build_spatial_projection(world)`

確定性、**唯讀**（不改 WorldModel、不改 version），回 `SpatialProjection`：

```text
current_area / current_area_label / current_area_roles
routes_from_here        本區可通行的出口（state ∈ known/available/used）
blocked_routes          本區受阻的出口（locked/blocked/unknown/unsafe/jammed）
safe_retreat_routes     可通行且通往 role=safe_zone 的路線
visible_entities         本區可見的物件/NPC（綁定 area==current 或未綁定；fact 一律不算可見）
known_remote_entities    已知但不在眼前（別區物件 + 所有 fact）
mental_map_text          確定性模板（只用投影內容，不捏造路線/實體/真相）
versions / counts / truncated
```

- **visible vs remote**：物件/NPC 在 `_world_model_tick` 被綁到登記時的 current_area（`tag_entity_area`，只認新登記）。
  撤到安全區後，站內物件 area≠current → 自動轉 `known_remote`（**安全區不顯示站內物件為眼前物**）。
- fact 是「知道」不是「看見」→ 永遠 remote。

---

## 二、Dirty-version Cache（P1）— 不變則不重算

- WorldModel 暴露 `version_snapshot()`：`world_version / area_version / exit_version / entity_version /
  fact_version / mode_version`；任何結構變動（register/set_state/inspect/exit/role/current/tag）都 bump。
- `SpatialProjectionCache` 以 `(current_area, profile, 各 version)` 為 key；profile=`exploration_mode`
  （模式變動也失效）。WorldModel 不變 → 回上次同一物件（O(1)）；一變動 → 重算。

---

## 三、Mental Map 文字（P3）＋ Async Worker（P4，選用）

- **確定性 fallback 永遠可用**：`deterministic_mental_map_text` 只用投影內容，**不新增任何地圖元素**。
- **`MentalMapWorker`（選用、預設未接線）**：daemon thread + 有界 queue（滿就丟工作）；遊戲迴圈**永不 await**；
  潤飾失敗/超時/驗證不過 → 用 fallback。`validate_mental_map_summary` 目前強制非空＋字數上限；
  `projection_label_whitelist` 提供「不得引入投影外具名項」白名單供未來嚴格比對（通用 prose NER 不在 MVP 範圍）。

---

## 四、觀測預算（P4）

`visible_entities / known_remote_entities / routes / blocked` 各有上限 + `truncated` 旗標 + `counts` 總數；
`mental_map_text` 字數上限。大 WorldModel 不會撐爆 observation；QA 一眼看出有沒有被截斷。

---

## 五、Spatial UX — 玩家/QA 可讀摘要（`player_facing_spatial_summary`）

由 `SpatialProjection` **確定性**生成（不呼叫 LLM、不餵 story），分段（只列非空）：

```text
目前位置：{current_area_label}
可走路線：{routes}（無 → 明示「沒有明顯可走的出口」，不捏造死路）
被阻擋路線：{label（中文狀態，如 鎖住）}
安全撤退路線：{safe_retreat}
眼前可互動物：{visible 物件/NPC，排除 taken/used}
已知但不在眼前：{known_remote 物件 + facts}
```

- observation 新增（top-level 與 `spatial_debug` 內皆有）：`spatial_summary` /
  `spatial_summary_truncated` / `spatial_summary_source="deterministic_projection"`。
- truncated＝清單被預算截斷 **或** 文字超出字數上限。agent_play `--auto` 印〔spatial〕面板（debug/UX 輔助，不取代 narrative）。
- review_mode 也產生摘要供玩家面板用，但**不餵 story agent**（P5 未接）。

---

## 六、UX selfplay 評估摘要（real LLM）

`dev/tools/spatial_ux_selfplay.py`：AI 玩家讀 summary 決策，逐 beat 評估 6 指標。實測：

| 維度 | 結果 |
|---|---|
| 定位（看得懂在哪） | ✅ 6/6 |
| visible↔remote 分流 | ✅✅ 撤退後物件正確轉「已知但不在眼前」 |
| 無幻影元素 | ✅✅ 零 phantom |
| 長度/與 narrative 重複 | ✅ 短（avg ~40 字）、重疊 ~0.1，不重複 |
| **可走/被阻路線** | ⚠️ **目前全空**——WorldModel 尚未登記 exit 實體（kernel 場景轉移只更新 current_area） |

> **已知缺口**（非本層 bug）：**routes/blocked 全空**，因為 kernel 移動玩家時沒在 WorldModel 登記 exit。
> 補「exit 登記」比進 P5 更值得——那是「導航維度從空轉變有用」的前提。**最大價值目前在 observation/UX，不在 story。**

---

## 七、邊界（本層**不做**）

不呼叫 LLM（同步路徑）·不 mutate WorldModel ·不擴 scene graph ·不做 2D/3D 幾何 / pathfinding ·
不改 TruthEvidenceGate / reveal ·不收斂 world_facts ·**不接 P5**（不餵 story/review agent context）。
