# SOURCES — 文件來源與單一真相歸屬

> 本專案有四組文件，分工明確。**改東西前先確認你在改的是不是該層的單一真相。**

| 文件 | 角色 | 單一真相（canonical for…） |
|---|---|---|
| `nightmare-assault-design-fixed/` (00–09) | **設計真相（WHAT）** | 世界觀、演算法、UI/UX、資料契約(07)、工程(08)、修訂記錄(09) |
| `…design-fixed/build/` | **施工方法（HOW）** | 19 工單(BUILD-PLAN)、具體契約(CONTRACTS：Pydantic/簽名/常數/目錄)、CHECKLIST 驗收項 |
| `nightmare-assault-parallel-dev-plan.md` | **並行策略** | 6 工線、7 階段、git 分支、測試策略、MVP-A/B 範圍與不做清單 |
| `dev/` | **執行操作層** | 工單看板(STATUS)、檔案快照回滾、階段排程、我（Claude）逐工單施工的入口 |

> 舊版 `nightmare-assault-design/` 已被 `-fixed` 取代；舊的 E0–E17 stories/epics 歸檔在 `dev/_legacy/`（不進 board 掃描，僅留存歷史）。

---

## 歸屬規則（避免重複/漂移）

- **設計規格**（演算法、schema 細節、UI 內容）→ 改 `design-fixed/`，不在 dev/ 複製。
- **契約（介面/Pydantic/簽名/常數/事件/SQLite 表）** → canonical 在 `design-fixed/build/CONTRACTS.md` + `07` + `08` + `01 §四/§五`；`dev/CONTRACTS.md` 只是**索引**（列 id 指過去），改契約走 `dev/CONTRACTS.md §九` 程序。
- **工單內容/驗收** → canonical 在 `build/BUILD-PLAN` + `CHECKLIST`；`dev/stories/*.md` 是**可執行鏡像**（含 lane/stage/快照紀律），由 `dev/tools/gen_stories.py` 的資料表生成。要改工單範圍：改 `gen_stories.py` 資料表或直接編該工單檔。
- **並行/階段策略** → canonical 在 `parallel-dev-plan.md`；`dev/PARALLEL-PLAN.md` 是**執行映射**（工單×工線×階段 + 模型策略 + 執行機制）。
- **進度/回滾** → 只在 `dev/`：`STATUS.md`(看板)、`snapshots/`(回滾)、`journal/`(日誌)。build/README 建議的 PROGRESS.md 由 `dev/STATUS.md` 取代（更強：機械重算 + 快照回滾）。

## skills/ 的定位

`design-fixed/skills/`＝**遊戲執行期 agent prompt**（setup/story/warden…），是被實作的對象，由 `U07` SkillLoader 載入專案根 `skills/`（熱重載）。**不是開發工作流的一部分。**

## 一眼導覽

新 session 開機：`CLAUDE.md` → `dev/STATUS.md` → `dev/ROLLBACK.md`。
要施工：`dev/PARALLEL-PLAN.md`（做哪個工單）+ 該工單檔 + `dev/CONTRACTS.md` 指到的 canonical 段落。
