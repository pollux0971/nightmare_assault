# WORKFLOW — 工作流（contract-first 模組邊界並行 + 工單生命週期）

> 方法論。**逐階段可執行排程在 `dev/PARALLEL-PLAN.md`**；契約在 `dev/CONTRACTS.md`；回滾在 `dev/ROLLBACK.md`。
> 唯一事實來源 `dev/STATUS.md`（由 `board.py` 從 `dev/stories/*.md` 工單重算）。
> 並行隔離用 git worktree；**回滾一律用 `dev/snapshots/` 檔案快照，不靠 git。**
> 來源：`nightmare-assault-parallel-dev-plan.md`。

---

## 一、核心：不是「功能並行」，是「模組邊界並行」

本專案風險不在 UI/CRUD，而在：多 agent 資料契約一致、LLM 輸出可穩定解析、Blackboard/Snapshot/Compactor 不污染狀態、前後端串流同步、story 只讀 revealed。

→ 所以採 **contract-first**：先凍結契約（`CONTRACTS.md` 已 FROZEN），再讓工線在邊界內並行；
必須有一位**架構整合者（Tech Lead＝主 Opus agent）**負責契約、review、merge、整合測試。其他工線只能在契約內實作，不可改共用介面。

---

## 二、五原則（parallel-dev-plan §3）

1. **先凍結契約，再寫功能**：`core/models.py`、`core/constants.py`、`StreamParser` 事件、`OpenRouterClient`、`API` 方法名、`NA.*` 事件名、SQLite 表名與必要欄位——進工單 1 後不得任意改（走 CONTRACTS §九）。
2. **所有 LLM agent 先 mock，再接真 API**：先用 mock output 確認流程能通，再接真模型。
3. **前端不解析 LLM 原始字串**：只接 `NA.appendToken/onContinue/onDecision/onStatus/onError/onBeatComplete`；`<<<DECISION>>>` 與 JSON repair 全在後端 `StreamParser`。
4. **非同步 agent 只產 patch**：compactor/dreaming/offstage-fate 不直接寫主 Blackboard，只產 pending patch，beat 安全點統一 merge。
5. **每天至少一次整合測試**：feature → integration 每日合一次，跑 smoke test，避免「各自說好了卻接不起來」。

---

## 三、工線編制（parallel-dev-plan §4）

| 工線 | 負責 |
|---|---|
| **A** 整合者/Tech Lead | 維護 CONTRACTS、審 PR 是否改壞契約、integration branch、跨模組衝突、E2E。**不可省略**（即使一人開發也要保留此角色，避免自己亂改所有模組） |
| **B** Core State/Persistence | models / blackboard / scene / db / snapshot / save·load |
| **C** LLM Infra/Parser | client / parser / fallback / timeout / trace / JSON repair / SkillCaller base |
| **D** Agent Logic | setup / orchestrator / warden / story / event extraction / compactor |
| **E** Frontend/pywebview | webview_app / ui / mock API / streaming rendering / state·error UI |
| **F** QA/Test | fixtures / mock LLM outputs / parser·snapshot 測試 / 30 beat 模擬 / 防暴雷 / injection |

詳細工線↔工單映射見 `PARALLEL-PLAN.md §一/§三`。

---

## 四、Git 分支策略（parallel-dev-plan §8）

```
main          穩定可 demo 版本
develop        開發整合分支
integration    每日整合測試分支
feature/core-* / feature/llm-* / feature/agents-* / feature/ui-* / feature/test-*
```
合併規則：feature → integration（每日一次）→ 測試過 → develop → 穩定 demo → main。
改 CONTRACTS 須 Tech Lead（A）批准；PR 附測試結果或失敗原因。

**禁止事項**：UI 分支改 Pydantic models｜agent 分支改 SQLite schema｜parser 分支改前端事件名｜多人同改 `CONTRACTS.md`｜story 讀 real_bible。

> `git worktree add` 需先有初始 commit。`dev/`（STATUS/stories/snapshots/journal）只在 main/integration 維護；feature worktree 只改自己工線的程式碼。

---

## 五、工單生命週期狀態機

```
todo
  │ 認領：確認 depends_on 全 done → 填 owner_session/worktree、status→in_progress
  │       取【pre 快照】snapshot.py snapshot <id> pre → board.py → journal
  ▼
in_progress ──實作（先寫測試骨架再寫實作；mock-first）──→ 跑「## 驗收（可執行）」
  │
  ├─ PASS → snapshot post --verify pass → 回填 last_good_snapshot → status→done
  │         → 合併 feature→integration（每日）→ board.py → journal
  ├─ FAIL/中斷/關機 → 由 ROLLBACK.md 協定處理（restore 或續做）
  └─ 受阻（缺依賴/待決）→ status→blocked（記原因）
```

**Definition of Done（缺一不算）**：驗收指令存在且 pass｜post 快照已建且 `verify` 通過、`last_good_snapshot` 回填｜`board.py --check` OK｜未越界改 FROZEN 契約。

---

## 六、待決參數（實作時記實測值於 journal）

- dreaming 頻率（在場每 5 beat，UB4）｜emergent_lie 約束強度（中等）｜命運表權重（Phase 後）｜
  JSON repair 成功率門檻（≥95%，U08）｜中文 token 估算（CHECKLIST F7）｜回溯非決定性提示（F6）。

---

## 七、指令速查

```bash
python3 dev/tools/board.py [--check]                                   # 重算/驗證 STATUS（含前向依賴 lint）
python3 dev/tools/snapshot.py snapshot <工單id> pre|post [--verify pass]
python3 dev/tools/snapshot.py restore <snapshot-id> --yes
python3 dev/tools/snapshot.py latest-good <工單id>
python3 dev/tools/gen_stories.py                                       # 補建工單/階段骨架（不覆寫）
```
