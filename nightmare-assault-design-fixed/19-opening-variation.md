# 19 · 開場變體契約 — Opening Variation Contract（補丁十四 / v0.8）

> **承接**：`16-worldmodel.md`（世界記憶）/ `18-player-surface.md`（玩家體驗層）。本檔記錄**開場多樣性**——
> 讓每局開場的核心素材不再退化成 `紙條 / 林晨 / 找人`。
> **實作對應**：`core/narrative/opening_variation.py` · `opening_pools.py` · `opening_cooldown_store.py` ·
> `opening_variation_selector.py` · `opening_variation_gate.py` · `opening_variation_prompt.py` + loop 整合。
> **狀態**：全部落地、已測（**920 passed，flag OFF/ON 各一次**）。
> **一句話**：把開場具體名詞降級成「抽象類別 + 變體池 + cooldown + 契約強制」，LLM 只負責寫自然。

---

## 一、問題根因（不是文筆，是錨點）

baseline 觀察：`紙條` 幾乎每局出現、`林晨` 反覆出現、開場目標幾乎都是 `找人`。
根因不是模型文筆差，而是 **prompt examples 形成了高權重錨點**——文檔/prompt 反覆出現
「林晨留下紙條」「找人」之類，模型推斷這是「正確格式」，即使被要求創新也會回到高頻素材。

**原則**：不要用更多具體例子去修具體例子；要把具體例子降級成**抽象欄位 + 變體池**。
（不能只加一句「不要每次都用紙條」——那只是再加一個錨點。）

## 二、設計：契約決定素材，story 只執行

```text
Setup / OpeningDirector ─→ OpeningVariationContract（抽象素材）
                              │  StoryAgent 只能執行，不可重抽核心素材
                              ▼
CooldownLedger ── 近期 literal/archetype → forbidden
變體池 + 權重 ── 抽 motive / anchor / message_medium / first_interactable
                              ▼
StoryAgent 用表層敘事把素材寫自然（LLM）→ gate 檢查 → repair/fallback
```

| 抽象欄位 | 取代的具體錨點 | 池大小 |
|---|---|---|
| `motive_archetype`（動機） | 「找人」 | 12（missing_person 降權 0.6） |
| `personal_anchor_type`（人物錨點） | 固定姓名「林晨」 | 10（missing_sibling 降權 0.5、no_person_anchor 升權 1.2） |
| `message_medium`（第一則訊息載體） | 「紙條」 | 18（handwritten_note 降權 0.4） |
| `first_interactable_type`（第一個可互動物方向） | 常見錨定物 | 12（抽象類別，不綁主題名詞） |

> 契約**只決定開場用什麼素材**，刻意不含任何劇情走向/結局欄位——**不強制收束劇情**。

## 三、CooldownLedger（防近期重複）

- `recent_literals` / `recent_archetypes`：{素材 → last_used_run}，跨 run 持久化於 additive SQLite
  表 `opening_variation_state`（只 `CREATE IF NOT EXISTS`，舊存檔可讀；失敗退記憶體 ledger）。
- 冷卻窗：`current_run - last_used_run <= cooldown_runs` → 仍冷卻 → 轉 forbidden（literal 預設 3、archetype 2）。
- 被選中的 medium 用過後，其表層字串家族（如 handwritten_note→紙條/便條/字條）連同 motive/medium
  archetype 一起記進 ledger，**下一局**才看得到 cooldown。
- cooldown 把候選擋光 → `weighted_choice` 自動退回完整池並標 `cooldown_exhausted`（**不會無開場**）。

## 四、Selector（決定性、零 LLM）

- `build_contract(rng, current_run, ledger, pools)`：依權重抽 motive/medium/anchor（套 forbidden_archetypes）
  + interactable，組出契約；motive→`initial_goal`、anchor→`personal_anchor_label` 皆是**抽象提示**（無專有名詞）。
- 決定性 seed：`sha256(run_id:run_index)`（不依賴 PYTHONHASHSEED），存 `cooldown_debug.selector_seed`，可反查。
- 多次抽樣自然稀釋過度重複素材（測試：20 次開場 motive 不會全是 missing_person、至少 4 種）。

## 五、enforcement（gate → repair → fallback）

story 開場 output 進 `opening_variation_gate`：

| 違規類型 | 判定 |
|---|---|
| `forbidden_literal` | 出現 cooldown 擋下的具體字串 |
| `forbidden_archetype` | 用了被擋 archetype（missing_person→「失蹤/找人…」、handwritten_note→「紙條家族」） |
| `message_medium_mismatch` | 指定非手寫 medium 卻寫成紙條/便條/手寫留言 |

- 有違規 → **repair 一次**（補重寫指令再跑 story）；修好就採用。
- 仍違規 → **決定性 fallback opening**（程式組裝、保證不含 forbidden 素材）+ 乾淨 DecisionPoint。
- 違規/repair/fallback 寫進 observation `opening_variation_violation`，序幕暫存 beat_window 收斂成單一最終版。

## 六、observation（GUI / QA / AI 都吃這份）

```text
opening_variation           : motive_archetype / personal_anchor_type / message_medium /
                              first_interactable_type / forbidden_literals / forbidden_archetypes /
                              cooldown_applied / cooldown_exhausted / selector_seed
opening_variation_violation : has_violation / violations[] / repair_attempted / fallback_used
game_meta.opening_variation_contract : 完整契約（供 story/前端/QA；無 real_bible / hidden_truth）
```

## 七、邊界與 flag

- flag `ENABLE_OPENING_VARIATION_CONTRACT`（預設 OFF，獨立於 `ENABLE_NARRATIVE_CONTROL`）；OFF 時開場與補丁前完全一致。
- **非目標**：不改 WorldModel / TruthEvidenceGate / SpatialProjection / PlayerState、不新增固定故事內容、
  不做模板引擎、不強制收束劇情。
- 契約**不含 real_bible / hidden_truth**——素材是表層抽象類別，與防暴雷結構正交。

## 八、里程碑

```text
變體池      開場核心素材來自抽象類別，不是單一高頻名詞
Cooldown    近期用過的素材自動避開（跨 run 持久化）
契約        Setup 決定、Story 只執行、gate 強制
結果        紙條 / 林晨 / 找人 不再是預設答案
```
