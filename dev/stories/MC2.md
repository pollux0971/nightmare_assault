---
id: MC2
stage: C
lane: D
title: 聊天退出濃縮（3–4 句進 story hot context）
status: done
worktree: -
depends_on: [MC1, U14]
contracts: [NPCChat, RollingSummary]
last_good_snapshot: MC2__post__20260604-145547
owner_session: -
---

# MC2 · 聊天退出濃縮

- **階段**：C（MVP-C）　**工線**：D（Agent Logic）
- **依賴**：MC1, U14(compactor/summary)　**契約**：NPCChat, RollingSummary

## 目標 / 範圍
聊天室退出時三向分流（00 B5）：完整紀錄 → cold（SQLite，MC1 已存）；事實 → warm ledger；
**近期聊天濃縮成 3–4 句 → story hot context**（recent_chat_digest / rolling_summary）。發 CHATROOM_OPENED/CLOSED 信號。

## 對應來源
00 §六 B5、03 §一（三層記憶/聊天三向分流）、`skills/compactor`、CHECKLIST E。

## 實作步驟
- `condense_chat(caller, npc_name, history)`：把該場聊天濃縮成 3–4 句摘要（可 LLM 或規則 fallback）。
- 寫進 blackboard.recent_chat_digest（story 下個 beat 可見）；關鍵事實可進 ledger。
- 聊天開/關發 `EVT_CHATROOM_OPENED` / `EVT_CHATROOM_CLOSED`。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_chat_condense_mc2.py：退出 → recent_chat_digest 含 3–4 句濃縮、非空；完整紀錄仍在 chat_logs；CHATROOM_CLOSED 發出；caller 失敗有規則 fallback（不崩）。
- [ ] story 下個 beat context 看得到 digest；全回歸綠。

## 回滾備註
additive 濃縮 + 信號；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot MC2 pre`
- 完成：驗收 pass → `snapshot.py snapshot MC2 post --verify pass` → 回填 last_good_snapshot
