---
id: MC1
stage: C
lane: D
title: NPC-Chat agent + chat_logs 持久化
status: done
worktree: -
depends_on: [U07, U05]
contracts: [NPCChat, InjectionGuard]
last_good_snapshot: MC1__post__20260604-145409
owner_session: -
---

# MC1 · NPC-Chat agent + 持久化

- **階段**：C（MVP-C）　**工線**：D（Agent Logic）
- **依賴**：U07(SkillCaller), U05(db)　**契約**：NPCChat, InjectionGuard

## 目標 / 範圍
讓玩家能對**在場 NPC** 多輪對話。`core/agents/npc_chat.py`：組「認知卡」（公開面 + evolving + voice_sample，
**絕不放 real_bible/secret_core**，結構性防暴雷同 story）→ 呼 npc-chat agent → 回覆；玩家輸入包 `<player_action>`；
每則對話寫 ChatLog（追加）+ SQLite `chat_logs`（cold）。

## 對應來源
`skills/npc-chat/SKILL.md`（認知卡 + 四條邊界 + 職業折射）、00 §六 B5、CHECKLIST C2/C3。

## 實作步驟
- `core/agents/npc_chat.py`：`build_npc_chat_context(blackboard, npc_name, player_message, history)`（只投影公開面 + evolving + voice_sample + recent，**無 real_bible**）；`run_npc_chat(caller, ...)` → 回覆文字。
- self_aware：誠實型真心答錯不暗示、會說謊型可隱瞞（靠 prompt + 投影）。
- db：新增 `add_chat_log(run_id, npc, beat, role, content)` / `load_chat_logs(run_id, npc)`。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_npc_chat_mc1.py：build context 不含 real_bible/secret_core（防暴雷斷言）；多輪對話 round-trip；玩家輸入包 `<player_action>`；chat_logs 存讀一致。
- [ ] mock caller：回覆非空、符合認知卡投影；全回歸綠。

## 回滾備註
新增 agent + db 方法（additive）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot MC1 pre`
- 完成：驗收 pass → `snapshot.py snapshot MC1 post --verify pass` → 回填 last_good_snapshot
