---
id: MC5
stage: C
lane: D
title: 離場命運隱藏 + 重逢揭曉
status: done
worktree: -
depends_on: [MC4, U10]
contracts: [OffstageFate, OrchestratorOutput]
last_good_snapshot: MC5__post__20260604-150331
owner_session: -
---

# MC5 · 離場命運隱藏 + 重逢揭曉

- **階段**：C（MVP-C）　**工線**：D（Agent Logic）
- **依賴**：MC4, U10(揭露閘門)　**契約**：OffstageFate, OrchestratorOutput

## 目標 / 範圍
離場命運**對玩家完全隱藏**——命運敘事/碎片存獨立隱藏紀錄（不進 story context、不洩漏）。
NPC 重新 present（機遇/敵對歸來）或玩家搜屍（corpse）時，才把 carried_fragment 過揭露閘門揭曉。

## 對應來源
00 §六「離場隱藏」、`skills/offstage-fate`（hidden_from_player）、02 §三（揭露閘門）、CHECKLIST C7。

## 實作步驟
- 隱藏儲存：離場 fate_narrative/carried_fragment 存 SQLite 獨立 table（路線 A，不主動劇透；不進 story snapshot 可見層）。
- 重逢/搜屍觸發：NPC present 或 corpse 被 touch → orchestrator 把 carried_fragment 從 hidden 搬 revealed（過揭露閘門）→ story 才看得到。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_reunion_reveal_mc5.py：離場期間 story context / get_* 看不到命運敘事或碎片（隱藏斷言）；NPC 重新 present → carried_fragment 進 revealed；搜屍 → 碎片揭曉；隱藏紀錄存讀一致。
- [ ] 不主動劇透（路線 A）；全回歸綠。

## 回滾備註
additive 隱藏 table + 揭露接點；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot MC5 pre`
- 完成：驗收 pass → `snapshot.py snapshot MC5 post --verify pass` → 回填 last_good_snapshot
