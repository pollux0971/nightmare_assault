---
id: UB6
stage: B
lane: D
title: 序幕鉤子 + 真相種子（Opening Hook & Truth Seed）
status: done
worktree: -
depends_on: [U09, U13]
contracts: [OpeningSeeds, SetupOutput, InjectionGuard]
last_good_snapshot: UB6__post__20260604-130013
owner_session: -
---

# UB6 · 序幕鉤子 + 真相種子

- **階段**：B（MVP-B 差異化）　**工線**：D（Agent Logic）
- **依賴**：U09(setup), U13(story)　**契約**：OpeningSeeds, SetupOutput, InjectionGuard

## 目標 / 範圍
解決「開場缺一個夠強的核心疑問」：序幕不只進入場景，而是先丟出一個「不對勁的核心」。
新增 **Opening Truth Seed Layer**——在開場放 2–4 個**真假混合**的異常片段（不暴雷），讓玩家第一分鐘就被勾住。

## 對應來源
使用者設計回饋（patch）：`nightmare-assault-design-fixed/12-mvp-b.md §一–九`。
seed 五型：True / False / Imagery / Personal / Mechanical。開場義務 5 條。長度政策 600–900 字。

## 實作步驟
- `core/agents/opening.py`：
  - `build_opening_seeds(blackboard)`：從 real_bible（world_truth + revelation_pool + protagonist）程式碼組裝 seed；
    每個 seed = {id, type, **surface**(可餵 story), **hidden_truth**(留 real_bible，不餵 story), opening_obligation}。
  - `OPENING_OBLIGATIONS`（5 條）+ `OPENING_LEN_MIN/MAX`(600/900)。
  - `build_opening_context(blackboard, base_ctx)`：把 **surface + obligation** 注入開場 context（**絕不含 hidden_truth**）。
- loop `_kernel_intro_beat`（與 legacy start）用 build_opening_context 富化開場 context。
- `skills/story/SKILL.md` 加「開場序幕規則」段（動機鉤子/真假混合異常/身份鉤子/表層想像/停在第一選擇；禁止直接解釋完整真相）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_opening_ub6.py：build_opening_seeds 產 ≥2 seed 且涵蓋 personal/mechanical/true 類型；build_opening_context 含 5 條 opening_obligations + surface seeds；**context 不含任何 hidden_truth / real_bible（防暴雷斷言）**；長度政策常數存在。
- [ ] story SKILL.md 含開場規則；JS/載入無誤；全回歸綠燈。

## 回滾備註
opening 為開場 context 富化 + SKILL 規則（additive）；hidden_truth 結構性不外洩；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot UB6 pre`
- 完成：驗收 pass → `snapshot.py snapshot UB6 post --verify pass` → 回填 last_good_snapshot
