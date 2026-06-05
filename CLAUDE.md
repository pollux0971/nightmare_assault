# CLAUDE.md — Nightmare Assault 開發鐵律

> 本檔的指令 **OVERRIDE** 預設行為，任何在本專案工作的 session 必須嚴格遵守。
> 設計真相在 `nightmare-assault-design-fixed/`；施工方法在 `…/build/` + `nightmare-assault-parallel-dev-plan.md`；
> 我執行的開發作業系統在 `dev/`。四者關係見 `dev/SOURCES.md`。

---

## 0. 最高指令

**本專案只能照 `dev/` 作業系統開發。不得在工單流程外寫任何遊戲程式碼。**
所有實作都對應某個 `dev/stories/<工單id>.md`（U00–U19 / UB*）；沒有對應工單的程式碼不准寫——
若需要，先建工單（改 `dev/tools/gen_stories.py` 資料表或手建），再認領。

---

## 1. 每個 session 開場序列（強制，先做這個）

依 `dev/ROLLBACK.md §二` 執行，順序不可省：
1. 讀 `dev/STATUS.md`（唯一事實來源）+ `dev/journal/` 最後 1–2 筆。
2. `python3 dev/tools/board.py --check` → 有 DRIFT/ERROR 先修。
3. **先處理所有 `in_progress` 殘留**（驗收→補 post 標 done，或回滾標 todo），確認清空。
4. 才依 `dev/PARALLEL-PLAN.md` 認領下一個 todo（其 `depends_on` 須全 done）。

> 這一步保證關機 / context 缺失後不靠記憶也能正確接續。

---

## 2. 工單生命週期鐵律（見 WORKFLOW §五）

認領 → **取 pre 快照** → `in_progress` → 實作（**先測試骨架、mock-first**）→ **跑驗收** →
- pass → **取 post 快照（`--verify pass`）** → 回填 `last_good_snapshot` → `done` → 合併 feature→integration → `board.py` → journal。
- fail → **一律回滾**（`snapshot.py restore`）→ `todo` → 重做。**不准「先放著」帶髒狀態前進。**

進 `in_progress` 前**必須**先有 pre 快照；否則不准動程式碼。

---

## 3. Definition of Done（缺一不算完成）

1. 工單的「## 驗收（可執行）」指令存在且 **pass**。
2. **post 快照已建**且 `snapshot.py verify` 通過，`last_good_snapshot` 已回填。
3. `board.py` 已重算 STATUS.md（`--check` 印 OK 一致、0 errors），journal 已更新。
4. 未越界改 FROZEN 契約。

---

## 4. 契約紀律

- `dev/CONTRACTS.md` 是**契約索引（FROZEN）**，canonical 規格在 `design-fixed/build/CONTRACTS.md` + `07` + `08` + `01 §四/§五`。**唯讀**；要改走 `dev/CONTRACTS.md §九`（影響盤點 → **使用者同意** → 同步所有工線 → 重凍結 → journal）。
- 寫工單時把 canonical 相關段落貼給子 agent（Pydantic 類/簽名/常數/事件名）。
- worktree 合併衝突若落在契約檔 = 有人越界，停下走變更程序。

---

## 5. 工程鐵律（程式碼層強制，見 CONTRACTS / 07 / 08 / CHECKLIST）

- **權限**：story 程式碼層拿不到 `real_bible` 物件（C2，非靠 prompt）；dreaming/offstage 碰不到 `secret_core`（C6）；違規寫 anchor 拋 `PermissionError`。
- **防暴雷斷言（E2）**：掃 story 輸出，確認不含未 revealed fragment 的 content；每 beat 驗。
- **injection（C3）**：餵 story/npc-chat 的玩家輸入一律包 `<player_action>`，永不直拼 prompt。
- **前端不解析（D4）**：`<<<CONTINUE>>>/<<<DECISION>>>` 與 JSON repair 只在後端 `StreamParser`；前端只接 `NA.*` 事件。
- **warden fallback 順序（B9）**：本地硬規則 → LLM 語義 → 全失敗才保守正常推進（不誤殺）。
- **並行 patch（A4/A5）**：非同步只產 patch，安全點 merge；過期 base_version 拒/rebase；同步路徑只讀穩定快照。
- **快照存「當時」摘要（A3）**：回 beat 10 不帶 beat 30 摘要。
- **graceful degradation（B8/F8）**：除 setup（唯一不可降級），任一 agent 掛掉遊戲都要能續行。

---

## 6. 並行紀律（contract-first 模組邊界並行）

- 一工單 = 一 worktree = 一 session；工單 frontmatter `owner_session` 已佔用者不得重入。
- 工線 A–F（見 WORKFLOW §三）；分支 `main/develop/integration/feature/<lane>-*`，每日整合一次。
- **禁止**：UI 工線改 Models｜agent 工線改 SQLite schema｜parser 工線改前端事件名｜多人同改 CONTRACTS｜story 讀 real_bible。
- `dev/`（STATUS/stories/snapshots/journal）只在 main/integration 維護；feature worktree 只改自己工線程式碼。
- **開發模型策略**：主 agent（整合者，我）用 **Opus**；子 agent（各工線）預設 **Sonnet**，承重牆 **U08 StreamParser / U14 compactor** 可升 Opus。切換見 `dev/PARALLEL-PLAN.md` 抬頭。fan-out 用 Agent 工具帶 `model: "sonnet"`。

---

## 7. 回滾骨幹（不靠 git）

- 回滾一律用 `dev/snapshots/` 檔案快照（`dev/tools/snapshot.py`），**永不靠 git commit/reset**。git 只用於 worktree 隔離。
- 寧可回滾重做一個工單，也不要在不確定的半成品上繼續疊。

---

## 8. 設計即規格

- `design-fixed/skills/` 是**遊戲執行期 agent prompt**（被 `U07` SkillLoader 載入專案根 `skills/`），是被實作的對象，**不是開發工作流**。別與 `dev/` 混為一談。
- 實作引用 `nightmare-assault-design-fixed/` 章節與 `build/`，**不複製、不臆造**規格。範圍採 **MVP-A → MVP-B**（不做清單見 parallel-dev-plan §11）。
- 有歧義回查 `09-revision-notes.md`、`build/DESIGN-CHANGES.md`、`00 §六 決策日誌`；仍不明則問使用者。
- 寫完整的程式，不偷工減料；就算遇到 /compact 或關機，靠本作業系統續接。

---

## 9. 指令速查

```bash
python3 dev/tools/board.py [--check]                                  # 重算/驗證 STATUS（含前向依賴 lint）
python3 dev/tools/snapshot.py snapshot <工單id> pre|post [--verify pass]
python3 dev/tools/snapshot.py restore <snapshot-id> --yes
python3 dev/tools/snapshot.py latest-good <工單id>
python3 dev/tools/gen_stories.py                                      # 補建工單/階段骨架（不覆寫）
```

**文件地圖**：`dev/SOURCES.md`（四組文件歸屬）·`dev/PARALLEL-PLAN.md`（階段排程，「做階段/工單 N」依據）·
`dev/WORKFLOW.md`（方法論/工線/分支/生命週期）·`dev/ROLLBACK.md`（復原協定）·
`dev/CONTRACTS.md`（契約索引 FROZEN）·`dev/STATUS.md`（看板）·`dev/stages/`·`dev/stories/`·`dev/journal/`。

> 使用者說「做階段 N / 做工單 U0X」時：依 `dev/PARALLEL-PLAN.md §四` 執行（確認前置→開 worktree→子 agent 並行→整合 barrier→驗收）。
