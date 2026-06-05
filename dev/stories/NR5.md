---
id: NR5
stage: R
lane: F
title: Motif Cooldown（母題冷卻 — 重複母題須演化/揭露/暫停）
status: done
worktree: main
depends_on: [NC5, NC3]
contracts: [MotifCooldown, QualityGate]
last_good_snapshot: NR5__post__20260605-011527
owner_session: session-24-opus
---

# NR5 · Motif Cooldown

- **階段**：R　**工線**：F（品質 / 驗證）
- **依賴**：NC5（QualityGate）、NC3（Story Agent Downgrade / StoryContextBuilder）
- **契約**：MotifCooldown, QualityGate（見 §十四 / §十二）

## 目標 / 範圍
防止同一母題（stopped_clock / water_reflection / metal_scraping / distorted_npc_face …）反覆出現卻不演化，造成停滯。重複母題的下一次使用須揭露新資訊 / 改變狀態 / 變可行動 / 進冷卻。

## 對應來源
patch `task_cards/P5_MotifCooldown.md`、`docs/06-motif-cooldown-and-motive-heartbeat.md §motif`、`reference_code/motif_tracker.py`。

## 實作步驟
- 逐場景追蹤 used_motifs 次數（接 NR1 npc-chat used_motifs + story 輸出）。
- 加冷卻計數；超 `max_uses_per_scene` → StoryContextBuilder 封鎖或要求轉化該母題。
- QualityGate（NC5）flag 連續停滯母題（同母題連 3 beat 告警）。
- 範例階梯：掛鐘停 11:55 → 顫動 → **第三次須揭露 11:55 是林晨最後通話時間**（轉成 EvidenceEvent，接 NR0）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_motif_cooldown_nr5.py：同母題同場景出現 2 次後，第三次須揭露/改變狀態/被封鎖；QualityGate 對連 3 beat 同母題告警。
- [ ] flag OFF 退回現況；既有套件綠。

## 回滾備註
motif tracker + context 封鎖 + gate 告警；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR5 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR5 post --verify pass` → 回填 last_good_snapshot
