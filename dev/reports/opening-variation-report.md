# Opening Variation Contract（補丁 v0.8）— 階段報告

> 對應 `nightmare-assault-opening-variation-contract-patch-v0.8/`。
> 目標：消除 prompt 錨點導致的 `紙條 / 林晨 / 找人` 反覆出現——用抽象變體池 + cooldown + 契約強制，
> 不靠「不要重複」的 prompt 提醒。

## 結論

- **全測 928 passed / 3 skipped，flag OFF 與 ON 各一次。**（新增 29 條開場變體測試）
- **config safe**：`config/config.json` 未追蹤、diff 無真 key。
- **真 LLM 量測（deepseek-chat-v3，各 6 局，同 6 主題）**：開場核心素材從「幾乎單一」變成「每局不同」。

| 指標 | BEFORE（flag OFF） | AFTER（flag ON，修簡體 gate 後） |
|---|---|---|
| message_medium 分布 | 全部 `(n/a)`——story 自由選，預設退化成紙條/照片/訊息 | **6 局 6 種不同**（body_mark / npc_claim / schedule_entry / terminal_prompt / inventory_anomaly / printed_receipt），**0 handwritten_note** |
| motive_archetype 分布 | 全部 `(n/a)`——missing_person 口味約 50% | **6 局 6 種不同**（recover_memory / retrieve_object / prove_innocence / deliver_or_hide / investigate_signal / repair_system），**指派 missing_person 0/6** |
| 紙條 出現率（敘事偵測，繁簡都收） | 0.167 | after-1 0.167 →（修簡體 gate）**after-2 0.0** |
| missing_person 出現率（敘事偵測） | **0.50** | **0.333**（且皆為背景，非指派動機，見下） |

> **after-1 → after-2**：契約是決定性的（run_id + seed 固定），兩次 after-run 抽到的 medium/motive 完全相同；
> 差別只在 gate——after-1 漏掉 run1 的簡體 `纸条`，**修簡體 gate 後 after-2 紙條率歸 0**（repair 生效）。
>
> **missing_person 殘留 0.333 的性質**：after-2 的 run2/run4 敘事提到「失踪站长 / 前同事林昭失踪」，但這兩局的
> **指派動機分別是 `prove_innocence` / `investigate_signal`**，失蹤者只是 `former_colleague` 錨點的背景，不是玩家目標。
> 這是**設計上正確**的——本補丁要消除的是「每局都把『找失蹤的人』當預設**動機**」（已達成：指派 0/6 vs baseline 50%），
> 而非禁止敘事出現「失蹤」一詞。`missing_person` 的語意只有在它被近期當過動機（進 cooldown）時才被 gate 擋。
>
> BEFORE 的 personal_anchor 幾乎都是「一個有名字的失蹤年輕女性」（林小夏 / 林筱…）——這正是 `林晨` 錨點的同類復發。
> AFTER 的 anchor 由契約抽象指派（past_self / former_colleague / no_person_anchor…），且名字由本局世界觀長出、不復發。

## 過程中抓到並修掉的真 bug

**簡體中文繞過 gate**：第一次 after-run 雖然契約已把 medium/motive 完全多樣化，但 run1 的**最終敘事**仍漏出
`纸条`（簡體）。根因：`opening_variation_gate` 的 forbidden cue 只放繁體（`紙條`），而 deepseek 輸出簡體，
gate 漏判 → 沒觸發 repair。**修正**：gate 的 `_HANDNOTE_CUES` / `_MISSING_PERSON_CUES` 與 pools 的
handwritten_note literals 改為**繁簡並收**；新增簡體偵測測試。修正後 gate 會對簡體輸出觸發 repair/fallback。

## 各階段

| 階段 | 內容 | 狀態 |
|---|---|---|
| P0 baseline | 6 局開場真 LLM 量測，確認 紙條/找人/失蹤的人 反覆 | ✅ missing_person 50%、紙條 17%、anchor 多為具名失蹤少女 |
| P1 schema | enums + OpeningVariationContract + CooldownLedger + weighted_choice + 變體池 + gate | ✅ |
| P2 cooldown | CooldownLedger（recent→forbidden）+ additive SQLite 跨 run 持久化 + run 計數 | ✅ |
| P3 selector | 池+cooldown→契約，決定性 seed，fallback，抽象 goal/anchor | ✅ |
| P4 整合 | `ENABLE_OPENING_VARIATION_CONTRACT`（OFF default）；start 產契約存 game_meta，graceful | ✅ |
| P5 prompt | 契約注入序幕 ctx + `skills/story/SKILL.md` 去錨定章節（synced design-fixed） | ✅ |
| P6 obs/tests | gate→repair→fallback + observation；29 條測試；beat_window 收斂；簡體 gate 修正 | ✅ |

## 非目標（已遵守）

不改 WorldModel / TruthEvidenceGate / SpatialProjection / PlayerState；不新增固定故事內容；不做模板引擎；
**不強制收束劇情**（契約只決定開場素材，無任何 ending/走向欄位——已用測試 `test_contract_has_no_plot_convergence_fields` 鎖住）。

## P7（GUI variation pools/weights 編輯）

依 APPLY_ORDER「之後再做」，本輪不做。
