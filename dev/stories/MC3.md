---
id: MC3
stage: C
lane: E
title: 聊天室 UI + API
status: done
worktree: -
depends_on: [MC1, U18]
contracts: [NPCChat, JsApi]
last_good_snapshot: MC3__post__20260604-145821
owner_session: -
---

# MC3 · 聊天室 UI + API

- **階段**：C（MVP-C）　**工線**：E（Frontend / pywebview）
- **依賴**：MC1, U18(前端)　**契約**：NPCChat, JsApi

## 目標 / 範圍
前端聊天室：列在場 NPC、開聊天 modal、多輪對話、退出。webview API：`list_present_npcs` / `open_chatroom` /
`send_chat` / `close_chatroom`；對話以 NA 事件推前端。

## 對應來源
06 §（聊天室 UI）、07 §一（API/JsApi）、`skills/npc-chat`。

## 實作步驟
- webview API：list_present_npcs（從 blackboard 在場 NPC）/ open_chatroom(npc) / send_chat(npc, text) → 回 NPC 回覆 / close_chatroom(npc)（觸發 MC2 濃縮）。
- 前端：header 💬 鈕 + dlg-chat（NPC 列表 + 對話視窗 + 輸入框 + 退出）；不洩漏 is_key_item/secret。
- 玩家輸入包 `<player_action>`（後端 MC1 已處理）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_chat_api_mc3.py：list_present_npcs 只列在場；open/send/close round-trip；send_chat 回覆非空；close 觸發濃縮；不洩漏 secret。
- [ ] JS 語法過；前端 💬 開聊天、可多輪、可退出；全回歸綠。

## 回滾備註
純前端 + API（additive）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot MC3 pre`
- 完成：驗收 pass → `snapshot.py snapshot MC3 post --verify pass` → 回填 last_good_snapshot
