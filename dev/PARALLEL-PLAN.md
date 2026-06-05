# PARALLEL-PLAN — 階段化執行排程（工單 × 工線）

> 把 MVP-A 的 20 個工單（U00–U19）+ MVP-B（UB1–UB5）依 `depends_on` 切成編號階段。
> **contract-first 模組邊界並行**：先凍結契約（見 `CONTRACTS.md`），再讓不同工線在邊界內並行。
> 使用者說「**做階段 N**」或「**做工單 U0X**」，我就讀 `dev/STATUS.md` 確認前置、依 §四執行。
> 來源：`nightmare-assault-parallel-dev-plan.md`（§4 工線 / §6 階段）+ `design-fixed/build/BUILD-PLAN.md`（工單）。方法論見 `WORKFLOW.md`。

> **執行模式**：序列為主（照工單 `depends_on`），同階段不同工線可並行（git worktree + 子 agent）。
> **模型策略**：主 agent（整合者，我）用 **Opus**；子 agent（各工線）預設 **Sonnet**；承重牆 **U08 StreamParser / U14 compactor** 建議升 Opus。

---

## 一、工線（parallel-dev-plan §4）

| 工線 | 角色 | 工單 |
|---|---|---|
| **A** | 整合者 / Tech Lead（契約、review、整合分支、E2E） | U00, U15（跨工單監看） |
| **B** | Core State / Persistence | U01, U03, U04, U05, UB5 |
| **C** | LLM Infra / Parser | U02, U06, U07, U08 |
| **D** | Agent Logic | U09, U10, U11, U12, U13, U14, UB1, UB2, UB4 |
| **E** | Frontend / pywebview | U16, U17, U18, U19 |
| **F** | QA / Test（貫穿） | 每工單驗收內含；UB3 防暴雷報告 |

> 一人 + AI：我當 A（Tech Lead），B/C/D/E/F 由子 agent 分頭做，每次都貼 `CONTRACTS.md` 相關段落。

---

## 二、階段排程（進入條件 + 整合驗收）

| 階段 | 工單（並行結構） | 進入條件 | 整合驗收（done 門檻） |
|---|---|---|---|
| **0** 契約凍結+地基 | `U00`(A) | 無 | 契約無矛盾；pytest 綠燈；import core OK ✅**已完成** |
| **1** 核心資料層（無 LLM） | B:`U01`→`U03`→`U04`/`U05`｜C:`U02` | 階段0 | models 可驗證；patch+version；SQLite 存假 beat 讀回；回 beat 10 不帶新摘要；signalbus 收發 |
| **2** LLM 基礎+agent 外殼 | C:`U06`→`U07`、`U08`｜D:`U11` | 階段1 | client call/stream/fallback+trace；parser 三級 repair 過；SkillCaller 載 SKILL；event 抽取可用 |
| **3** 核心 agent 真接 | D:`U09`、`U10`、`U12` | 階段2 | setup 生雙層 bible+scene；orchestrator 條件揭露；warden 硬規則優先（LLM 掛仍觸發） |
| **4** story+compactor | D:`U13`(防暴雷)、`U14`(30beat) | 階段3 | story 停決策點且不暴雷；compactor 30 假 beat 不爆、伏筆留、回溯摘要正確 |
| **5** beat 主迴圈 | A:`U15` | 階段4 | 連續多 beat 狀態一致；非同步 patch 不污染 story 讀取。**後端 MVP-A 核心通** |
| **6** 前端+MVP-A 打磨 | E:`U16`→`U17`→`U18`→`U19` | 階段5 | **MVP-A 驗收（見下）** |
| **B** MVP-B | D:`UB1`/`UB2`/`UB4`｜F:`UB3`｜B:`UB5` | MVP-A 驗收 | 技能封頂強化、一種結局、防暴雷報告、輕量 dreaming、道具庫完整 |
| **S** 穩定化補丁（Progress Kernel） | B/C/D/A:`SK01`→…→`SK13`（序列為主） | MVP-A done | world-state 由 kernel 決定、story 只 realize；開門不重複、每 beat ≥1 delta、attractor 結局。✅**已完成** |
| **P** 配置中心 / Story Agent 模組化 | A:`P0`｜B:`P1`｜C:`P2`｜D:`P3`/`P4`/`P7`｜F:`P6`｜E:`P5`（apply order P0→P1→P2→P3→P4→P6→P5→P7） | MVP-A + 階段S done | story prompt 可配置/模組化（fragment 組裝、preview、run 快照、回滾），不破壞 MVP-A；static prompt 永遠 fallback。詳見 `stage-P.md` |
| **N** 敘事控制（Narrative Control v0.1） | A:`NC0`｜D:`NC1`/`NC2`/`NC3`/`NC4`/`NC6`｜F:`NC5`｜E:`NC7`（apply order NC0→…→NC7） | MVP-B（UB1–UB7）done | 開場少元素建動機、真相分層、story 不發明世界觀、結局有因果門檻；flag OFF 退回現況。詳見 `stage-N.md` |
| **R** 敘事控制 v0.2（揭露橋接） | D:`NR0`/`NR1`/`NR2`/`NR3`/`NR4`/`NR6`｜F:`NR5`/`NR7`（apply order NR0→…→NR7） | 階段 N + MVP-C done | 調查→真相橋接（結局不再 0/X）、npc-chat 收同一敘事契約、答債、結局表層變體、兩段式逃脫、母題冷卻、動機心跳、表層消毒；flag OFF 退回現況。詳見 `stage-R.md` |
| **H** Runtime Hard-Gate（v0.3.1） | A:`HA1`｜D:`HA2`/`HB1`/`HB2`/`HC1`｜F:`HA3`/`HD1`/`HD2`/`HE1`（apply order HA1→HA2→HA3→HB1→HB2→HC1→HD1→HD2→HE1） | 階段 R done | 把半接的 Narrative Control 升級成真正 runtime hard-gate（pass/reject/repair once/fallback）：ended⇒無 options、danger≠death、observation 不洩 hidden、調查必轉 evidence、NPC 結構化受閘、品質/表層能 repair；flag OFF 退回現況。詳見 `stage-H.md` |

**MVP-A 驗收（六項，design 08 §五 / parallel-dev-plan §7 M5）**
30 beat 不崩 ｜ 防暴雷（story 不輸出未 revealed fragment）｜ JSON ≥95% 可解析/可 repair ｜
30 beat 後核心伏筆仍可引用 ｜ 回 beat 10 不帶 beat 30 摘要 ｜ 換解析度不跑版。

```
甘特（mock 支線可在早期並行；工單依賴為實線）
階段:  0✅   1            2            3          4         5        6
A      U00 ───────────────────────────────────────────────── U15
B          U01→U03→U04/U05 ······································ (UB5)
C          U02         U06→U07/U08
D                      U11          U09/U10/U12   U13/U14
E      (UI mock 支線 ·············································) U16→U17→U18→U19
F      (fixtures / 防暴雷 / 30beat / injection — 貫穿) ················
```

> mock-first：所有 LLM agent 先用 mock output 跑通流程，再接真 API（parallel-dev-plan 原則二）。
> 前端可用 mock API 先做（支線），但工單 U16–U19 的「真 API」依賴 U15。

---

## 三、工單 ↔ 階段 ↔ 工線 索引

- 階段0：`U00`(A) ✅
- 階段1：`U01`(B) `U02`(C) `U03`(B) `U04`(B) `U05`(B)
- 階段2：`U06`(C) `U07`(C) `U08`(C,承重牆) `U11`(D)
- 階段3：`U09`(D) `U10`(D) `U12`(D)
- 階段4：`U13`(D,防暴雷) `U14`(D,承重牆)
- 階段5：`U15`(A)
- 階段6：`U16`(E) `U17`(E) `U18`(E) `U19`(E)
- MVP-B：`UB1`(D) `UB2`(D) `UB3`(F) `UB4`(D) `UB5`(B)

- 階段S（穩定化補丁，MVP-A 後）：`SK01`–`SK13`(B/C/D/A) ✅
- 階段P（配置中心 / Story Agent 模組化，patch v1.1）：`P0`(A) `P1`(B) `P2`(C) `P3`(D) `P4`(D) `P6`(F) `P5`(E) `P7`(D)

- 階段N（敘事控制 v0.1，MVP-B 後）：`NC0`(A) `NC1`(D) `NC2`(D) `NC3`(D) `NC4`(D) `NC5`(F) `NC6`(D) `NC7`(E)
- 階段C（MVP-C，聊天室 + 離場命運）：`MC1`(D) `MC2`(D) `MC3`(E) `MC4`(D) `MC5`(D)
- 階段R（敘事控制 v0.2，揭露橋接，階段 N + MVP-C 後）：`NR0`(D) `NR1`(D) `NR2`(D) `NR3`(D) `NR4`(D) `NR5`(F) `NR6`(D) `NR7`(F)
- 階段H（Runtime Hard-Gate v0.3.1，階段 R 後）：`HA1`(A) `HA2`(D) `HA3`(F) `HB1`(D) `HB2`(D) `HC1`(D) `HD1`(F) `HD2`(F) `HE1`(F)

階段索引（含進入/驗收）：`dev/stages/stage-0.md` … `stage-6.md`、`stage-B.md`、`stage-S.md`、`stage-P.md`、`stage-N.md`、`stage-C.md`、`stage-R.md`、`stage-H.md`。

---

## 四、呼叫一個階段/工單時，我怎麼執行

```
使用者：「做階段 N」或「做工單 U0X」
 1. 讀 dev/STATUS.md → 確認該工單所有 depends_on 皆 done。
    CONTRACTS=FROZEN 才可動（已 FROZEN）。
 2. 為該階段每條要並行的工線開 worktree：
      git worktree add ../na-<lane> -b feature/<lane>-<工單>
 3. 每工線派一個背景子 agent（worktree 隔離，模型依模型策略：預設 Sonnet，U08/U14 升 Opus），交付指令包：
      · 該工單 frontmatter + 內文（目標/步驟/驗收）
      · CONTRACTS.md 相關段落（貼具體 canonical：07/08/build/CONTRACTS）
      · 對應 SKILL.md（若牽涉某 agent）
      · 鐵律：FROZEN 契約唯讀、權限模型、mock-first、防暴雷斷言、injection 包 <player_action>
      · 快照紀律：開工 snapshot pre；驗收 pass 後 snapshot post --verify pass
 4. 我（A）監看；每工線完成 → 跑驗收 → post 快照 → 回填 last_good_snapshot。
 5. 整合 barrier（每日一次/階段末）：feature → integration → 跑 smoke test → develop；
      合併衝突落在 FROZEN 契約檔 = 越界，停（走 CONTRACTS §九）。
      跑 dev/tools/board.py 更新 STATUS、board.py --check 必須 OK。
 6. 跑該階段「整合驗收」；全綠才宣告階段 done → 可呼叫下一階段；dev/journal/ 記一筆。
```

關機/中斷後：照 `ROLLBACK.md §二` 開機序列；in_progress 殘留先處置，再續呼叫。

---

## 五、不變式

- 任一工單 `depends_on` 的 stage ≤ 自身 stage（**無前向依賴**，`board.py --check` 自動驗）。
- 每個 `contracts` 對得到 `dev/CONTRACTS.md`。
- 進 in_progress 必先有 pre 快照；done 必有 post 快照（`--verify pass`）。
