# 03 · Agent Context 規劃

---

## 一、三層記憶

```mermaid
flowchart LR
    subgraph HOT[Hot · live context]
        H1[當前 beat]
        H2[最近 5-8 beat 原文]
        H3[玩家決定 verbatim]
    end
    subgraph WARM[Warm · 永遠可用]
        W1[fact ledger 二元組]
        W2[滾動摘要散文主線]
        W3[NPC 演化層當前值]
    end
    subgraph COLD[Cold · 按需召回]
        C1[完整聊天紀錄]
        C2[舊 beat 原文]
        C3[reflection_log 封存]
    end
    HOT -.滑出視窗.-> WARM
    WARM -.溢出/封存.-> COLD
    COLD -.語義召回.-> HOT
```

| 層 | 內容 | 在 prompt？ | 儲存 |
|----|------|-----------|------|
| **Hot** | 當前 beat、近期視窗、玩家決定 | 是 | 記憶體 |
| **Warm** | ledger 二元組、滾動摘要、NPC 演化層 | 是，永遠可用 | Blackboard |
| **Cold** | 完整聊天紀錄、舊 beat、reflection_log | 否，召回才進 | SQLite |

聊天室退出三向分流：完整紀錄 → cold；事實 → warm ledger；一句散文 → 滾動摘要。

---

## 二、Blackboard Schema

```yaml
Blackboard:
  # ── 不可變錨點（setup 後鎖死）──
  real_bible:                           # 完整真相，僅 orchestrator + warden 可讀
    world_truth: { what_really_happened, the_threat_is, deadly_rule }
    revelation_pool:                    # 5-8 碎片，含揭露條件
      - { id, type, content, reveal_condition }   # type: knowledge|item|person
    ending_conditions: [{type, trigger, gate|prerequisites}]
    atmosphere: [str]

  revealed_bible:                       # 已揭露子集，story agent 只能讀這個
    revealed_fragments: [{id, content, how_to_reveal}]
    known_atmosphere: [str]

  npc_registry:
    - name: str
      personality: enum [leader, nervous, analytical, optimistic, mysterious]
      personality_axes:                 # 程式碼擲骰組合，免費隨機化
        speech_rhythm: enum [簡短, 絮叨, 停頓多, 喜歡反問]
        emotional_base: enum [壓抑, 焦躁, 疏離, 過度熱情]
        quirk: str                      # 從怪癖池抽（搓手/不看眼睛/反覆確認…）
      profession: str                   # 職業，決定 knows_about 範圍
      voice_sample: str                 # 一句範例台詞，錨定語氣（非形容詞）
      secret_core: str                  # 客觀真相，不可變，錨定 world_truth
      self_aware: bool                  # NPC 自己知不知道這個真相
                                        #  true→隱瞞/說謊  false→真誠答錯（更恐怖）
      public_face: str
      presence: enum [present, absent, missing, dead]      # 物理在場
      alignment: enum [allied, neutral, departed, hostile, dead]  # 關係立場
      offstage_intent: str | null        # 離場支線（dreaming 寫）
      return_condition: str | null
      fate_pressure: float               # 離場命運計量器，達標擲骰
      carried_fragment: str | null       # 命運領到、待交付的碎片 id
      evolving:                         # 可變，dreaming 寫，無 warden
        emotional_state: {emotion, intensity}
        relationship: {trust, suspicion, affinity}
        intent: enum [observe, befriend, betray, flee, manipulate]
        revealed_layers: [str]
        emergent_lies: [str]
        personal_arc: str

  protagonist: { name, starting_situation, claimed_skills: [(技能,侷限)] }

  scene_registry:                       # 場景系統（location_reached、屍體種植、NPC 位置依賴它）
    current_location: str               # 主角當前所在 location id
    known_locations:
      - id: str
        name: str
        description: str
        discovered: bool                # 玩家是否已抵達/發現
        exits: [str]                    # 可前往的 location id
        interactables:
          - id: str
            type: enum [item, clue, corpse, door, npc_trace]
            linked_fragment: str | null # 綁定的 revelation_pool 碎片 id
            revealed: bool              # 玩家是否已注意到
    # NPC 位置 = npc.presence + 此處的 location 對照（NPC 在哪由其狀態 + 場景共同決定）

  shared_inventory:                     # 共用道具庫
    items:
      - { id, name, brief, acquired_beat, held_by, is_key_item, hidden_clue }

  # ── Warm 層 ──
  rolling_summary: str                  # 散文，有上限
  ledger: [(type, content)]             # 二元組
  recent_chat_digest: str | null        # 剛離開聊天室時的 3-4 句濃縮

  # ── Hot 層 ──
  beat_window: [Beat]                   # 最近 5-8
  turn_context:                         # 每 beat 重置
    player_decision: str                # verbatim
    input_path: enum [option, custom_checked, free_text]  # 三種輸入路徑
    warden_verdict: dict
    newly_revealed: [fragment]          # orchestrator 本 beat 揭露的
    narrative_output: dict
    pending_npc_evolutions: [dict]

  # ── 系統 ──
  game_meta: { beat_number, mode, difficulty, context_usage_pct }
```

---

## 三、各 Agent 的 Context 規格

每個 agent：讀哪些切片、可寫哪些欄位、輸入/輸出 token 預算。

### setup（Heavy，一次性）
- 讀：玩家輸入（主題、NPC 數、主角名、tone 關鍵字）
- 寫：整個 real_bible、npc_registry、protagonist、opening 序列
- Token：in ~600 / out ~2500（一次性，可大方）

### orchestrator（Light，每 beat，揭露閘門）
- 讀：real_bible.revelation_pool（未揭露碎片 + 揭露條件）+ game_meta.{beat_number, difficulty} + turn_context（玩家觸及的碎片、所在場景）+ npc trust
- 寫：revealed_bible（從 real 搬碎片）、turn_context.newly_revealed
- Token：in ~700 / out ~150（多數揭露條件程式碼判，僅少數需 Light LLM 判「玩家是否實質觸及某碎片」）
- **是唯一能讀 real_bible 又餵給 revealed 的橋；story 永遠讀不到 real**

### story（Medium，每 beat，串流）
- 讀：revealed_bible（已揭露子集，靜態部分**可快取**）+ rolling_summary + ledger + beat_window + player_decision + 在場 NPC 的 evolving + 近期聊天摘要 + warden directive + orchestrator newly_revealed
- 寫：BeatHistory（追加）、turn_context.narrative_output
- Token：in ~4000（恆定）/ out ~600–1500（長度不限，分塊）

### warden（Light，每 beat，僅玩家動作）
- 讀：player_decision + real_bible.{deadly_rule, ending_conditions} + ledger.{技能} + game_meta.beat_number（warden 可讀 real，判結局/規則需要完整真相）
- 寫：turn_context.warden_verdict、ledger（技能二元組）
- Token：in ~800 / out ~200（精簡，多用 enum）
- **不讀 NPC 演化提議**

### npc-chat（Light，隨選）
- 讀：該 NPC 的**認知卡**（見下）+ rolling_summary + 該聊天視窗
- 寫：ChatLog（追加）
- Token：in ~1200（獨立小 context）/ out ~300

### NPC 認知卡（npc-chat 與 dreaming 共用的餵入結構）

讓 NPC「了解自我/環境/關係/專業」但 call 仍精簡。核心：**結構化餵 context，指令極短；個性用 voice_sample 錨定，非形容詞指令。**

```yaml
npc_cognition_card:
  # 自我層
  name, personality, profession, voice_sample
  current_situation: str       # 「被困地下室，剛和玩家爭執」
  self_aware: bool             # 知不知道自己的 secret_core
  # 環境層（只給在場期間）
  recent_events: str           # 他在場的最近幾 beat 精簡摘要
  what_just_happened: str
  # 關係層（只給自己的感受）
  relationship: {對主角, 對在場其他NPC}
  # 專業層
  knows_about: [str]           # 職業 + 已揭露資訊讓他知道的事

指令（極短）:
  「以 {name} 的身分回應。只談你知道的事。
   職業外的問題不裝懂，用你的視角回應或轉移。」
```

**認知邊界（讓 NPC 不像笨蛋也不像全知神）**：
- 自我：知道 `public_face`，但秘密由 `self_aware` 決定。卡片**不含 secret_core 內容**——self_aware=true 才另給「你知道 X 但想隱瞞」，false 則不給（真誠答錯，更恐怖）。
- 環境：只知在場期間發生的事，離場期間一無所知。
- 關係：只知自己的感受，不讀他人內心。
- 專業：職業外問題不裝懂。

**個性隨機化（程式碼擲骰，不花 LLM token）**：
```
setup 時程式碼從多軸抽組合:
  核心特質 5 種 × 說話節奏 4 種 × 情緒底色 4 種 × 怪癖池 N
  → 幾乎不重複的獨特個性，免費生成、可控、不千篇一律
voice_sample 由 setup LLM 順帶生成（依抽到的組合）
```

**職業折射（「不答專業外」的正確實作）**：
不是拒答（像系統錯誤），而是用職業視角折射——工程師「我只懂機械，但這線路被動過手腳」（轉懂的領域 + 給線索）、護士迴避露不安、警衛轉移到規則暗示。職業與行為的矛盾（訪客卻懂設備）本身就是推理線索。

### dreaming（Light，非同步，無 warden）
- 讀：近期 beat + 該 NPC 的聊天封存片段（cold 召回）+ 該 NPC 全狀態 + ledger（自我約束用）
- 寫：npc_registry[].evolving（**權限邊界：碰不到 secret_core**）
- Token：in ~1500 / out ~400，**每個在場 active NPC 各一次**
- **僅在場 NPC**（反應式）

### offstage-fate（Light，非同步，離場 NPC 專用）
- 觸發：程式碼命運 tick 翻出結局後才呼叫（非每次 tick）
- 讀：結局類型（程式碼擲定）+ secret_core + 領到的 revelation 碎片 + offstage_intent
- 寫：npc_registry[].{presence, alignment, carried_fragment, offstage_intent}、Scene（屍體結局種 corpse_interactable）
- Token：in ~1000 / out ~400，**頻率遠低於在場 dreaming**（命運觸發才跑）
- **與 dreaming 不同 skill**（生成式 vs 反應式）

### compactor（Medium，非同步）
- 讀：beat_window（將滑出者）+ rolling_summary + 保護清單
- 寫：rolling_summary、ledger（重組）、ColdArchive
- Token：in ~3500 / out ~600

---

## 四、Prompt Caching 策略

每 beat 都跑的 story agent，其輸入中 **revealed_bible 的已揭露部分相對穩定（~800 tokens）**。開 prompt caching 後快取部分只算 0.1× 成本。

```
story prompt 結構（快取友善排序）:
  [可快取前綴] system prompt + revealed_bible 已揭露穩定部分
  [變動部分]   rolling_summary + ledger + beat_window + player_decision + NPC evolving + directive
```

注意快取 5 分鐘過期——適合玩家每 30–60 秒互動一次的節奏。

---

## 五、Token 預算與成本估算

單一典型 beat（story + warden）：

```
story（Medium @ ~$0.10/$0.40 per M）:
  in 4000 × 0.10/M + out 1000 × 0.40/M ≈ $0.0008
warden（Light @ ~$0.40/$1.60 per M）:
  in 800 × 0.40/M + out 200 × 1.60/M ≈ $0.0006
每 beat 互動成本 ≈ $0.0014
```

非同步 pass（不卡玩家，但計入成本）：

```
compactor（每 ~5 beat 一次，Medium）≈ $0.0006
dreaming（每 K beat，每 active NPC 一次，Light）:
  單次 ≈ in 1500 × 0.40/M + out 400 × 1.60/M ≈ $0.0012
  ⚠ 成本來源：5 NPC × 每 5 beat = 每 5 beat 多 5 次呼叫
  → 優化：只跑 active NPC（凍結沒戲份者）
```

**一場遊戲（50 beat、3 active NPC、dreaming 每 5 beat）粗估**：

```
互動: 50 × $0.0014                  = $0.070
compactor: 10 × $0.0006             = $0.006
dreaming: (50/5) × 3 × $0.0012      = $0.036
聊天室（若用，另計）
一場 ≈ $0.11（約 3.5 NTD）
```

**關鍵發現**：
1. dreaming 是最大的隱藏成本（隨 active NPC 數線性成長）→ 凍結沒戲份的 NPC 是必要優化。
2. story 的 input 恆定 ~4000 tokens → prompt caching 對它最有價值。
3. setup 一次性，可放心用 Heavy 把真相密度做高。
