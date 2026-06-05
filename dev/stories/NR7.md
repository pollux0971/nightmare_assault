---
id: NR7
stage: R
lane: F
title: SurfaceTextSanitizer（render 前清非故事洩漏）
status: done
worktree: main
depends_on: [U13, MC1]
contracts: [SurfaceTextSanitizer]
last_good_snapshot: NR7__post__20260605-011800
owner_session: session-24-opus
---

# NR7 · SurfaceTextSanitizer

- **階段**：R　**工線**：F（品質 / 驗證）
- **依賴**：U13（story agent / StreamParser 輸出）、MC1（npc-chat 輸出）
- **契約**：SurfaceTextSanitizer（見 §十四）

## 目標 / 範圍
story / npc-chat 偶爾洩漏非故事 token（technical / protocol / COLLECT / inst、prompt artifact、壞 markdown fence、敘事內重複分隔符），破壞沉浸。render 前掃掉，但保留契約授權的 in-world 詞。

## 對應來源
patch `task_cards/P7_SurfaceTextSanitizer.md`、`docs/07-surface-text-sanitizer.md`、`reference_code/surface_sanitizer.py`。

## 實作步驟
- 建 disallowed token 清單 + 授權 in-world 例外（432.7、17Hz、明確 in-world 的 protocol，由 NarrativeContract 授權）。
- 在 story / npc-chat 輸出之後、render 之前跑 sanitizer（後端，符合 D4：前端不解析）。
- 修補策略：先安全的決定性替換；不安全則短 repair prompt（移除非故事 artifact，**不改劇情/選項/evidence/JSON metadata**）。
- 不全域刪英文（CJK 中夾雜的洩漏才處理，如「靜得能聽見technical的呼吸聲」→ 移除 technical）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_surface_sanitizer_nr7.py：含 `technical/protocol/COLLECT/inst` 洩漏的輸出 → 被移除/修補；授權詞（432.7/17Hz）保留；劇情/選項/evidence 不被改。
- [ ] sanitizer 只在後端；flag OFF（或預設）退回現況；既有套件綠。

## 回滾備註
sanitizer pass + repair fallback；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR7 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR7 post --verify pass` → 回填 last_good_snapshot
