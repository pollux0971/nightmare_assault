---
id: NC3
stage: N
lane: D
title: Story Agent Downgrade（只執行 blueprint，不發明世界觀）
status: done
worktree: -
depends_on: [NC1, U13]
contracts: [StoryAgentDelta, PromptFragments]
last_good_snapshot: NC3__post__20260604-141942
owner_session: -
---

# NC3 · Story Agent Downgrade

- **階段**：N（敘事控制 · patch v0.1）　**工線**：D（Agent Logic）
- **依賴**：NC1, U13(story)　**契約**：StoryAgentDelta, PromptFragments

## 目標 / 範圍
降低 Story Agent 權限（原則 D：Story Agent 不是世界觀發明者）。story 只把 blueprint / obligations / allowed elements
寫成文字，**不得自行新增核心設定**；每 beat 只新增一個主要敘事資訊。

## 對應來源
patch `task_cards/P3_story_agent_downgrade.md`、`docs/09-story-agent-prompt-delta.md`、`prompts_reference/story_agent_delta.md`、`docs/02-agent-change-plan.md`。

## 實作步驟
- story context 加 `allowed_new_elements` / `forbidden_new_elements` / `beat_purpose` / `truth_reveal_limit` / `player_motive`。
- `skills/story/SKILL.md` 加「Story Agent Delta」段；移除/弱化鼓勵 checklist-stuffing 的措辭（與 UB6 序幕規則協調）。
- 可整合進 Config Center fragment（story.* 新增 `story.element_limit` / `story.beat_purpose`）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_story_downgrade.py：story context 帶 allowed/forbidden_new_elements + beat_purpose + truth_reveal_limit；SKILL.md 含 Delta 段；選項與 motive/clue/danger 相關（結構性檢查）。
- [ ] story 仍只讀 revealed（無 real_bible）；全回歸綠。

## 回滾備註
context 富化 + SKILL 規則（additive）；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC3 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC3 post --verify pass` → 回填 last_good_snapshot
