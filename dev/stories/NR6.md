---
id: NR6
stage: R
lane: D
title: Motive Heartbeat（動機心跳 — 開場後仍記得為何進來）
status: done
worktree: main
depends_on: [NR0, NC1]
contracts: [MotiveHeartbeat, NarrativeContract]
last_good_snapshot: NR6__post__20260605-011608
owner_session: session-24-opus
---

# NR6 · Motive Heartbeat

- **階段**：R　**工線**：D（agent）
- **依賴**：NR0（EvidenceEvent / RevelationBridge）、NC1（NarrativeContract 的 protagonist_motive）
- **契約**：MotiveHeartbeat, NarrativeContract（見 §十四 / §十二）

## 目標 / 範圍
v0.1 selfplay：開場動機（找弟弟/林晨）清楚，但**開場後動機淡掉**。維持主角動機每 2–3 beat 可見，且不重複同一句——嵌入文件 / NPC 反應 / 道具 / 選項。

## 對應來源
patch `task_cards/P6_MotiveHeartbeat.md`、`docs/06-motif-cooldown-and-motive-heartbeat.md §motive`。

## 實作步驟
- 追蹤距上次提及 active_player_motive 的 beat 數。
- 逾 2–3 beat → 在 StoryContextBuilder 加「動機心跳」義務。
- 支援 subtle 提醒管道：文件上出現林晨名字 / NPC 對名字反應 / 線索暗示曾經過此 / 選項問「追林晨還是逃」。
- 不重複同一句；盡量讓心跳同時是 EvidenceEvent 或母題演化（接 NR0/NR5）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_motive_heartbeat_nr6.py：連 3 beat 無動機提及 → 下一 context 含 motive heartbeat 義務；提醒管道多樣（不同 beat 不同句/管道）。
- [ ] flag OFF 退回現況；既有套件綠。

## 回滾備註
motive 追蹤 + context 義務；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR6 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR6 post --verify pass` → 回填 last_good_snapshot
