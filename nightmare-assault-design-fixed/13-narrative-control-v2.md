# 13 · 敘事控制 v0.2 — 揭露橋接（Revelation Bridge）

> **實作對應**：開發工單 `NR0`–`NR7`（dev 階段 R）。**狀態：已落地、已測**（8/8 工單 done；679 passed；接續 §十二 敘事控制 v0.1 / 階段 N，及 MVP-C / 階段 C 之後）。
> **canonical 來源**：`nightmare-assault-mvp-b-narrative-control-patch-v0.2/`（docs 00–10 / task_cards P0–P7 / reference_code / examples）。
> **一句話**：v0.1 讓開場收斂了，v0.2 把「玩家調查」真正接上「官方真相進度」——讓結局不再永遠 `0/X`，並把敘事控制延伸到 npc-chat。

---

## 〇、為何有這個補丁（診斷）

開 `ENABLE_NARRATIVE_CONTROL=true` 後，selfplay 證實開場元素確實收斂（動機 + 一個核心異常，而非元素過載）。但同一場跑也暴露更底層的整合斷鏈：

```text
ProgressKernel clues  ─┐
NPC-chat hints        ─┼─ 三層各自獨立，沒有接通
real_bible / revealed_bible truth fragments ─┘
```

玩家可以認真調查十個 beat（檢查紙條、追問 NPC、查文件），結局復盤仍是 `0/7`——因為 kernel 自己追蹤的線索，跟 real_bible 的真相碎片是脫節的。連帶暴露另外幾個 MVP-B 系統問題：npc-chat 失控擴張世界觀、重複提問被氛圍敷衍、模糊逃脫渲染得同乾淨逃脫一樣、母題停滯、開場後動機淡掉、表層偶有非故事 token 洩漏。

**修補原則**：不靠叫 Story Agent「多揭露一點」，而是建一條**受控橋接**，由系統決定哪些線索算真相進度。

```text
Evidence → Reveal Level → RevealedBible → Recap
```

> **Non-goals**：不加新 lore、新世界機制、更大故事圖。MVP-B 需要更好的接線，不是更多內容。

---

## 一、NR0 · RevelationBridge（地基）

把每個對玩家有意義的線索事件統一成 `EvidenceEvent`，過映射表升級真相階梯，寫進帳本，最後讓結局復盤讀帳本。

```text
ProgressKernel / NPCChat / Story 輸出
→ EvidenceEvent
→ RevelationBridge.process_evidence_events()
→ RevealManager（§十二 RevealLadder：不可跳級、需 evidence）
→ RevealedBible / RevealLedger
→ Ending recap（顯示部分進度）
```

- **EvidenceEvent**：`evidence_id / source(kernel|npc_chat|story|inventory|document) / player_action / surface_text / truth_id / suggested_reveal_level / evidence_strength(0–1) / scene_id / beat_number`；裝飾性線索標 `atmosphere_only=true` 不入橋。
- **真相階梯**（取代二元「發現/未發現」）：`hidden → hinted → observed → suspected → confirmed → actionable`。
  - `hinted` 看到線索但無法解釋｜`observed` 觀察到可重現現象｜`suspected` 證據足以推因｜`confirmed` 找到權威證明｜`actionable` 知道該怎麼用。
- **EvidenceMapping**：`clue_id ↔ truth_id / minimum_action / grant_level / grant_strength`；無映射的玩家線索進 debug「unmapped」告警。
- **Recap 計部分進度**（refine UB7 masked）：`已確認 1 / 已懷疑 2 / 已觀察 1 / 仍未知 3`，精簡版「你發現了 4/7 條真相線索，其中 1 條已確認」。
- **護欄**：不把每個氛圍細節當真相線索；不可無 evidence 直跳 confirmed；結構性防暴雷不變（bridge 只決定等級，content 仍只露已達等級的部分，story/npc 永不見 real_bible）。

**驗收**：檢查「不要相信 432.7」警示紙條 → `truth.signal_frequency >= hinted`；只要檢查過 ≥1 個 mapped 線索，recap 不顯示 `0/X`。

---

## 二、NR1 · NPCChat 敘事控制

v0.1 只約束主 Story Agent；npc-chat 仍能無控擴張世界觀。讓 npc-chat 收同一份敘事契約。

- **輸入契約**：`motif_palette / forbidden_motifs / truth_reveal_ladder / current_reveal_levels / active_player_motive / answer_debt / context_budget`（沿用 MC1 認知卡白名單，不洩 real_bible/secret_core）。
- **NPCChatResponse**：`visible_reply / npc_emotion_delta / answer_status(answered|partial|evaded|refused) / evidence_events[] / new_lore_terms[] / used_motifs[] / quality_flags[]`。
- **NPC 可**：表達情緒 / 迴避 / 部分回答 / 說謊（只准映射到既有 truth 或動機）/ 透過受控 EvidenceEvent 露線索。
- **NPC 不可**：建新組織 / 協定 / 核心機制 / 怪物類型；不可 hidden 直跳 confirmed。
- **硬護欄**：出現 `motif_palette / allowed_new_terms` 外的新 lore → repair 重寫（移除新 lore、保留情緒與部分答、把有用提示轉 EvidenceEvent）。
- 退出時 evidence_events 餵 NR0 橋接；沿用 MC2 濃縮回 story hot context。

---

## 三、NR2 · Answer Debt（答債）

重複直問不該被氛圍敷衍。追蹤「已問問題 × 是否已用有用資訊償還」。

- **問題分類**：identity / mechanism / threat / location / action。
- **債等級**：`0 無 / 1 可迴避 / 2 須部分答 / 3 須具體線索或具理由拒答`。
- **規則**：debt≥2 時，下一相關回應**至少**含 partial answer / concrete clue / direction to evidence / explicit refusal with reason 之一。注入 story 與 npc-chat context。

> 範例（問兩次「432.7 是什麼？」）：壞＝「掛鐘滴答越來越響、瞳孔收縮、水窪震動」；好＝吳靜搖頭「它不是時間，是校準值。每次儀器掉到 432.7 以下，就有人開始聽見不屬於自己的記憶。」——沒全揭真相，但付了債。

---

## 四、NR3 + NR4 · 結局表層變體 + 兩段式逃脫

### NR3 — EndingGate Surface Variant
NC6 內部已會把 0 真相逃脫標 ambiguous，但玩家看到的文字與乾淨逃脫無異。新增**表層變體**（獨立於 ending_type）：

```text
clean_escape     有足夠真相/控制權地逃出
ambiguous_escape 身體逃出，但帶著不確定/污染/未解的身分威脅
failed_escape    抵達出口卻無法離開或付出重大代價
death            身體或認知失敗
truth_locked     存活但真相不足以解決核心問題
```

- 由 RevealLedger 決定 clean vs ambiguous：**0 confirmed 真相 → 不可 clean_escape**。
- ambiguous_escape renderer 呈現不確定收尾（「你走出去了。至少，你以為自己走出去了…手機螢幕亮起，林晨的舊號碼再次打來，通話時間停在你一直看見的那個時間。」）。debug 印 gate 理由。

### NR4 — Escape Commit Gate
「我試圖離開」不該即刻結局。改兩段式：

```text
attempt_escape → exit_candidate_found（出口發現 beat + 選項）→ player commit_escape → EndingGate.evaluate()
```

首次逃脫意圖轉成出口發現 + 選項（提交離開 / 繼續調查 / 處理威脅），玩家明確提交才結算，與 SK13 attractor 結局協調。

---

## 五、NR5 + NR6 · 母題冷卻 + 動機心跳

### NR5 — Motif Cooldown
逐場景追蹤 used_motifs；超 `max_uses_per_scene` 的下一次使用須**揭露新資訊 / 改變狀態 / 變可行動 / 進冷卻**；QualityGate flag 連 3 beat 同母題。

> 階梯範例：掛鐘停 11:55 → 顫動 → 第三次須揭露「11:55 是林晨最後通話時間」（轉 EvidenceEvent）。

### NR6 — Motive Heartbeat
追蹤距上次提及主角動機的 beat 數；逾 2–3 beat 加「動機提醒」義務，且須透過文件 / NPC 反應 / 道具 / 選項嵌入，**不重複同一句**。盡量讓心跳同時是 EvidenceEvent 或母題演化。

---

## 六、NR7 · SurfaceTextSanitizer

story / npc-chat render 前掃非故事洩漏（`technical / protocol / COLLECT / inst`、prompt artifact、壞 markdown fence、敘事內重複分隔符）。

- 先安全的決定性替換；不安全則短 repair prompt（**不改劇情/選項/evidence/JSON metadata**）。
- 授權 in-world 例外（`432.7 / 17Hz` / 明確 in-world 的 protocol，由 NarrativeContract 授權）；不全域刪英文。
- 只在後端（符合 D4：前端不解析）。

---

## 七、整合次序與 beat loop 接點

```text
Player action → ActionExtractor → Warden → ProgressKernel
→ EvidenceBridge(kernel_events) → RevealManager
→ StoryContextBuilder（含 motif cooldown + motive heartbeat + answer debt）
→ StoryAgent → SurfaceTextSanitizer → QualityGate → Snapshot
```

NPC-chat：`NPCChatContextBuilder → NPCChatAgent → SurfaceTextSanitizer → NPCChatQualityGate → EvidenceBridge(npc_evidence) → Compactor`。
結局：`attempt_escape → exit_candidate_found → commit_escape → EndingGate.evaluate → EndingRenderer.render_surface_variant`。

> **關鍵不變式**：任何對玩家重要的線索都必須變成一個 EvidenceEvent。

---

## 八、旁路紀律與驗收

- 受 `ENABLE_NARRATIVE_CONTROL`（必要時細分子旗標）控管，**預設行為不變**；新欄位皆 optional、不破壞既有存檔；**story / npc-chat 永不見 real_bible 不變**；refine（非取代）UB7 masked 結局與 §十二 RevealLadder。
- **整合驗收主線**：玩家檢查警示紙條 + 問 NPC 可疑頻率 + 查文件後，**結局 recap 不得 0/X**（即使只是 hinted/suspected 也顯示部分發現）｜0 confirmed 逃脫渲染為 ambiguous_escape｜「我試圖離開」先出口候選不即結局｜重複提問付答債｜npc-chat 不發明未授權 lore｜母題不停滯｜表層無洩漏｜flag OFF 全程退回現況。
- **觀測**：per-beat debug（active_motive / used_motifs / motif_cooldowns / evidence_events / reveal_updates / answer_debt / ending_candidate）；end-of-run summary（總 evidence / mapped vs unmapped / reveal 分佈 / 母題過用 / npc lore 違規 / ending gate 理由）；fail-fast 告警（檢查 3+ 線索但進度仍 0、npc new_lore_terms>2、同母題連 3 beat、attempt_escape 直接結局）。

> 契約索引見 `dev/CONTRACTS.md §十四`；開發階段索引見 `dev/stages/stage-R.md`。
