# 10 · Narrative Progress Kernel（穩定化補丁）

> **補丁來源**：`nightmare-assault-mvp-a-stabilization-patch/`（+ 後續主題化/敘事規劃強化）。
> **實作對應**：開發工單階段 S（`SK01`–`SK13`）。**狀態：已落地、已測**。
> **核心一句**：把「世界狀態怎麼推進」從 LLM 手上拿走，交給程式碼層的 kernel；**story agent 退化成純 realizer**。

---

## 一、為什麼需要這個補丁

原始設計讓 LLM agent（orchestrator + story）在自由流程裡自行決定世界狀態。實測暴露三個結構性問題：

| 症狀 | 根因 |
|------|------|
| **開門後又問要不要開門** | 沒有權威的「已完成事件」紀錄，story 每 beat 重新詮釋世界 |
| **NPC 不進場、線索/背包不持久** | 狀態變更散落在敘事裡，沒有被結構化提交 |
| **行動沒有後果** | LLM 可能生出「什麼都沒改變」的 beat |

補丁的設計原則（沿用 00 §一「世界真相驅動」）：

> **強制「狀態推進」，而非「固定劇情推進」**。每個 beat 至少改變一項世界狀態（場景 / 線索 / 道具 / NPC / 威脅 / 結局吸引力），但玩家走法仍自由。

---

## 二、元件總覽

```
玩家行動
   │
   ▼
ProgressKernel.resolve_player_action ──► EventPatch（事件 + 義務 + forbidden_repeats + new_clues/items/npcs）
   │            （intent 正規化 → 候選事件選擇 → 評分 → 落地 dummy 節點）
   ▼
ContextBuilder.build_story_context ────► story 最小 context（只含 revealed + 義務，無 real_bible）
   │
   ▼
story（realizer）串流敘事
   │
   ▼
PatchValidator.apply ──────────────────► 驗 base_version + 強制 ≥1 progress_delta → commit GameState
   │
   ▼
ProgressBridge.sync_to_blackboard ─────► clues/inventory/npc/scene → 進 snapshot + story 可見層
```

| 元件 | 檔案 | 職責 |
|------|------|------|
| `ProgressModels` | `core/progress_models.py` | `GameState` / `EventPatch` / `EventCandidate` / `Obligation` / `LedgerEntry` / `InventoryItem` / `NPCPresence` / `PatchOp` / `SceneState` / `ProgressResult`（純 dataclass） |
| `PatchValidator` | `core/patch_validator.py` | base_version 檢查、**≥1 progress_delta 否則拒絕**、forbidden、套用 ops（scene/clue/inventory/npc） |
| `SceneGraphProvider` | `core/scene_graph.py` | 機會圖介面；`StaticOpeningSceneGraphProvider`（讀 `data/opening_scene_graph.json`）、`GeneratedSceneGraphProvider`（依 setup 主題生成） |
| `ProgressKernel` | `core/progress_kernel.py` | `resolve_player_action`：intent normalize → candidate select → score → dummy nodes → `EventPatch` |
| `ContextBuilder` | `core/progress_context.py` | 組 story 最小 context（committed_event / obligations / forbidden_repeats / new_clues / spawned_npcs + revealed_bible，**無 real_bible**） |
| `ProgressBridge` | `core/progress_bridge.py` | `GameState ↔ Blackboard` 同步（clues/inventory/npc/scene → snapshot + story） |
| `Attractors` | `core/attractors.py` | 結局吸引子：依累積狀態算各結局拉力，越門檻才觸發（非固定終點） |
| `GraphInvariants` | `core/graph_invariants.py` | 機會圖不變式 INV1–5（每場景 ≥2 事件、起始 ≥2 出口、離場 ≥2 解法…） |

---

## 三、story agent 的角色轉變

| | 補丁前 | 補丁後 |
|--|--------|--------|
| story 的身份 | 世界狀態的**裁判 + 敘事者** | **純 realizer（敘事者）** |
| 「門開了嗎」誰決定 | story 自己詮釋 | **kernel 決定**，story 只描寫結果 |
| 反重複 | 靠 prompt 自律 | kernel 給 `forbidden_repeats`，story 不得再提供 |

`skills/story/SKILL.md` 在最高優先區塊加入 **Narrative Progress Contract**：story 收到 `committed_event` / `narrative_obligations` / `forbidden_repeats` / `new_clues` / `spawned_npcs`，必須照辦、不得重複已解決事件、必反映新場景。這份 prompt 規則在補丁二（配置中心）被外部化為可配置 fragment（見 `11-config-center.md`），行為等價。

**範例（開門回歸）**

```
forbidden_repeats = [ask_open_door_101]
✗ 補丁前： 「你要打開那扇門嗎？」          ← 又問一次
✓ 補丁後： 「門開了，走廊的冷光把病房切成兩半。你可以：往走廊深處走 / 檢查門框抓痕 / 回頭確認病房。」
```

---

## 四、每個行動都有後果（PatchValidator）

`PatchValidator.apply(patch, state)` 的硬約束：

1. **base_version 檢查**：patch 的 base_version 必須等於當前 GameState version（並行安全），否則拒絕/rebase。
2. **≥1 progress_delta**：patch 必須帶至少一項狀態變化，否則拒絕。progress_delta 可以是場景變更、線索新增、道具增刪、NPC 狀態變更、威脅升降、新機會解鎖、結局吸引力變化任一。
3. **forbidden 套用**：把已解決事件加入 `forbidden_repeats`，下個 beat 不得重複。
4. **稀疏 fallback 強化**：玩家越界（圖上沒有對應事件）時，`ProgressKernel` 仍以 context-aware 的 `_make_fallback_event` 生出有 delta 的事件（缺線索→種線索 / 缺 NPC→留痕跡 / 否則升壓），保證越界輸入也有後果。

---

## 五、結局是吸引子，不是固定終點（Attractors）

`core/attractors.py`：每 beat 依累積狀態算各結局的「拉力」，某結局拉力越過門檻才觸發。

| 結局吸引子 | 拉力來源 | 門檻常數（`core/constants.py`） |
|------------|----------|--------------------------------|
| **death** | danger_level 累積 | `DANGER_DEATH_THRESHOLD` |
| **truth** | 揭露的真相碎片數 | `TRUTH_FRAGMENT_THRESHOLD` |
| **escape** | 線索數（摸清出路）+ 存活 + 有核心 | `ESCAPE_CLUE_THRESHOLD` |

結局與 warden 的硬規則結局**並存**：warden 結局 `via=warden`，吸引子結局 `via=attractor`。

---

## 六、旁路與降級（不破壞既有契約）

- **Feature flag `ENABLE_PROGRESS_KERNEL` 預設 ON**（`core/constants.py`，env 可關）。
- `BeatLoop` 是**單一分流點**（不做 flag 分支地獄）：`_step_kernel`（kernel 流程）vs `_step_legacy`（原 LLM 自由流程，原封保留）。
- kernel / 機會圖 / validation 任一失敗 → **log + 回退 legacy 流程，不 crash**（B8 graceful degradation）。
- `submit_decision` 仍由玩家輸入啟動、story 仍 stream、每 beat 仍快照（新增 clue/inventory/npc 進 snapshot）。
- scene graph provider 鏈：explicit → generated（主題化）→ static → legacy。

---

## 七、soft lookahead（可重算，非既定）

`ProgressResult.soft_lookahead`：每 beat 重算「落地後場景的其他可能事件」，餵給 story context 並**標示為非既定**（story 不得當成已發生）。不持久於 GameState——是每 beat 的軟預覽，讓敘事有方向感而不鎖死劇情。

---

## 八、契約索引

開發契約見 `dev/CONTRACTS.md §十`（`ProgressModels` / `PatchValidator` / `SceneGraphProvider` / `ProgressKernel` / `ProgressContext` / `ProgressBridge`）。本檔為設計說明，canonical 規格在 `nightmare-assault-mvp-a-stabilization-patch/` 與本專案 `core/progress_*.py`。
