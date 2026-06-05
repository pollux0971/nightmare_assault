---
id: HB1
stage: H
lane: D
title: Story Evidence Extraction（調查無 reveal 變化 → 保底 evidence）
status: done
worktree: main
depends_on: [NR0, U13]
contracts: [StoryEvidenceExtractor, EvidenceEvent]
last_good_snapshot: HB1__post__20260605-102203
owner_session: session-24-opus
---

# HB1 · Story Evidence Extraction

- **階段**：H　**工線**：D（agent）
- **依賴**：NR0（RevelationBridge/帳本）、U13（story）
- **契約**：StoryEvidenceExtractor, EvidenceEvent（擴充）（見 §十五）

## 目標 / 範圍
玩家做了有意義調查、story narrative 也吐出具體新資訊，但本 beat reveal 無變化時，**保底**產一個 hinted EvidenceEvent，讓「檢查文件/儀器後 reveal_progress 不會完全沒反應」。

## 對應來源
patch `task_cards/P0_B1`、`docs/04`、`reference_code/evidence_extractor.py`、`tests_reference/test_story_evidence_extraction.py`。

## 實作步驟
- 新增 `core/narrative/evidence_extractor.py`：`StoryEvidenceExtractor(truth_keyword_index)`：`is_investigation(action)` + `has_concrete_info(narrative)` + `reveal_changed==False` → fallback hinted EvidenceEvent；map 不到 truth_id → `source=fallback, truth_id=null, debug_reason=...`。
- EvidenceEvent 擴充：加 `reveal_level / player_facing / debug_reason`（向後相容，預設值）。
- loop `_revelation_tick` 後：若本 beat 無 reveal 變化且為調查型 → 跑 extractor，把產出 evidence 也走 bridge；計 `unmapped_evidence_this_beat`。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_story_evidence_ha1.py（沿用 reference）：調查型 action + 具體 narrative + reveal_changed=False → 產 hinted evidence、命中 keyword → truth_id 對應；非調查/無具體資訊 → 不產。
- [ ] 既有套件綠；board --check 0 errors。

## 回滾備註
新增 extractor + loop 接點；EvidenceEvent 新欄位 optional；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HB1 pre｜完成：snapshot.py snapshot HB1 post --verify pass
