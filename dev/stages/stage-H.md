# 階段 H · Runtime Hard-Gate（v0.3.1）

- **進入條件**：階段 R（NR0–NR7）done。
- **緣由**：把已存在但「半接」的 Narrative Control 模組從 monitor 升級成真正 runtime **hard-gate**——每個 gate 都有行為（pass / reject / repair once / fallback），不只是 log。修補 NR 階段的軟點（NPC 結構化 evidence、品質只 log、結局帶 options、danger→death、observation 洩 hidden、調查偶爾 0 反應）。
- **目標**：①`ended=true` 不得再有 options；②調查必轉 EvidenceEvent 推進 reveal/revealed_bible；③NPC-chat 受 Narrative Contract / Reveal Ladder / Gate 約束（結構化）；④QualityGate/Sanitizer 能 repair/reject/retry；⑤API/agent_play 不洩 hidden truth。
- **整合驗收（done 門檻）**：`not (ended and options)`｜danger-only 不致死｜observation 無 hidden content｜調查後 reveal 不永遠 0/X｜NPC 結構化且不繞 ladder｜QualityGate 失敗會 repair-once/fallback｜sanitizer 覆蓋全玩家面；flag OFF 全程退回現況。

## 套用順序（APPLY_ORDER：先安全線，再體驗線）

```text
Batch A（結局/劇透安全）  HA1 → HA2 → HA3
Batch B（調查→Evidence→Reveal） HB1 → HB2
Batch C（NPC runtime gate）     HC1
Batch D（品質/表層 repair）     HD1 → HD2
Batch E（可觀測/回歸）          HE1
```

## 本階段工單

| 工單 | 工線 | 內容 | depends_on |
|---|---|---|---|
| `HA1` | A | Ending Observation Invariant（ended ⇒ options=[]） | U15, NR4 |
| `HA2` | D | Death Causality Guard（danger ≠ death） | U12, SK13 |
| `HA3` | F | Hidden Recap Masking（observation 不洩 hidden content） | NR0, UB7 |
| `HB1` | D | Story Evidence Extraction（調查無 reveal 變化 → 保底 evidence） | NR0, U13 |
| `HB2` | D | RevelationBridge Unified Inputs（多來源 + reveal_updates 可觀測） | NR0, HB1 |
| `HC1` | D | NPCChat Structured Gate（結構化 + 閘門 + repair + bridge） | NR1, HB2 |
| `HD1` | F | QualityGate Repair Once（check→repair once→fallback） | NC5, NR5 |
| `HD2` | F | Surface Sanitizer All Outputs（全玩家面 + path/IP/權限） | NR7 |
| `HE1` | F | Observation Debug Fields（debug 欄位 + NPC 分層，不洩 hidden） | HB2, HA1 |

> 紀律（APPLY_ORDER §強制禁止）：danger 當直接死亡 ✗；普通 beat 同帶 ending+options ✗；observation 洩 full hidden truth ✗；QualityGate 只 log 放行 ✗；NPC 繞 Reveal Ladder ✗。
> 旁路：受 `ENABLE_NARRATIVE_CONTROL` 控管，預設行為不變；新欄位 optional；story/npc 永不見 real_bible 不變。
> beat pipeline 短期不大重構（docs/02）：在 `_step_kernel` 內整理 private method，MVP-B 穩定後再抽 BeatPipeline class。
> docs-first：canonical 在 `nightmare-assault-runtime-hardgate-patch-v0.3.1/`（reference_code 只是範例、不照抄）。契約見 dev/CONTRACTS.md §十五。
