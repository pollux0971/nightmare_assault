# ROLLBACK — 關機 / context 缺失後的復原協定

> **目標**：一個**全新、零記憶**的 Claude session（或人），不靠任何上一個 session 的對話記憶，
> 只讀檔案就能（a）重建「做到哪」，（b）判斷未完成的 WIP 該續做還是回滾，（c）回到已知良好狀態。
>
> **回滾骨幹是 `dev/snapshots/` 檔案快照，不靠 git。** git 只用於 worktree 並行隔離；
> 即使 git 歷史壞掉，本協定仍能運作。
>
> 術語：本專案工作單位＝**工單（work order, U00–U19/UB*）**；下文「story」與「工單」同義，檔案仍在 `dev/stories/`。

---

## 一、為什麼能在零記憶下復原（三個不變式）

開發 OS 維持以下不變式，使狀態永遠落在檔案、不落在記憶：

1. **STATUS.md 可機械重算**：它由 `board.py` 從 `dev/stories/*.md` 的 frontmatter 重算，
   不是手寫。→ 任何漂移都能被 `board.py --check` 抓出。
2. **進 in_progress 必先有 pre 快照**；**done 必有 post 快照（`--verify pass`）**。
   → 每個 story 都有「開工前還原點」與「known-good 還原點」。
3. **每一步都寫 journal**（append-only）。→ 意圖與決策有書面軌跡。

---

## 二、開機復原程序（每個全新 session 開場**必跑**）

```
步驟 0：定位
  cd <repo 根>；確認 dev/ 存在。

步驟 1：讀狀態（不靠記憶）
  - 讀 dev/STATUS.md（唯一事實來源；頂部「⚠ 需注意」區就是上次的未竟事務）。
  - 讀 dev/journal/ 最後 1–2 個檔（最近做了什麼、為什麼）。
  - 讀 dev/CONTRACTS.md 頂部狀態（**已 FROZEN**；契約唯讀，改走 §九 程序）。

步驟 2：驗證看板無漂移（exit code 可驅動自動化：乾淨=0、漂移/錯誤=1）
  python3 dev/tools/board.py --check
  - 印「OK: 一致」→ 看板可信。
  - 印「DRIFT」→ 有人手改過 STATUS.md 或 story；以 stories/*.md 為準，
    跑 `python3 dev/tools/board.py` 重生 STATUS.md，再往下。
  - 有 ERROR → 先修：
      · 懸空 depends_on / 重複 id；
      · 「story 檔 X 解析不到 id」＝上次關機把某 story 檔寫壞/截斷，
        補好 frontmatter 或用 gen_stories 重建後再往下。

步驟 3：清理未完成的 WIP（核心）
  對每個 status=in_progress 的 story（= 上次關機時正在做、可能是髒 WIP）：

    a. 跑該 story 檔「## 驗收（可執行）」列出的指令。

    b. 若 PASS（其實已做完，只是沒收尾）：
         python3 dev/tools/snapshot.py snapshot <id> post --verify pass
         → 回填 story frontmatter 的 last_good_snapshot；status→done
         → python3 dev/tools/board.py
         → journal 記「復原：<id> 驗收通過，已補收尾」

    c. 若 FAIL 或 狀態不明（半成品）：【回滾】
         # 取還原點：優先該 story 自己的 pre 快照；
         # 若無，取所屬 工線 最後一個 known-good
         SID=$(python3 dev/tools/snapshot.py list --story <id> | grep ' pre ' | tail -1 | awk '{print $1}')
         python3 dev/tools/snapshot.py restore "$SID" --yes
         → 把 story frontmatter status→todo、worktree/owner_session→ -
         → python3 dev/tools/board.py
         → journal 記「回滾：<id> 驗收失敗，已還原 <SID>，將重做」
         → 之後從乾淨狀態重新認領該 story（見 WORKFLOW §五）

步驟 4：確認無 in_progress 殘留後
  - 才依 WORKFLOW §三 波次，認領下一個所有 depends_on 皆 done 的 todo。
```

---

## 三、決策樹（in_progress 的 story 該怎麼辦）

```
in_progress story
   │
   ├─ 驗收指令 PASS ─────────────→ 補 post 快照 + 標 done（完成收尾）
   │
   ├─ 驗收指令 FAIL ─────────────→ restore pre/known-good + 標 todo（回滾重做）
   │
   └─ 無法判斷 / 環境壞了 ────────→ 一律保守回滾到該 工線 最後 known-good，
                                    寧可重做一個 story，不要帶著髒狀態前進
```

> 原則：**寧可回滾重做一個 story，也不要在不確定的半成品上繼續疊。**
> 一個 story 通常數小時內可重建；髒狀態往後拖會污染整條 工線。

---

## 四、回滾的三種尺度

| 尺度 | 情境 | 動作 |
|---|---|---|
| **Story 級**（最常用） | 單一 story 做壞 / 中斷 | `restore <該 story pre 快照>`，該 story 重做 |
| **工線 級** | 某 工線 多個 story 連環受污染 | `restore <該 工線 最後 known-good post 快照>`，往後 story 全部重做 |
| **全域級** | 整體狀態錯亂 | 取所有 工線 中時間最早的共同 known-good，全部回該點（極少用） |

`latest-good` 取某 story 的 known-good：
```bash
python3 dev/tools/snapshot.py latest-good <story-id>
```

---

## 五、快照保留與健康

- 快照存 `dev/snapshots/<id>/`，含 `payload/`（來源副本）+ `MANIFEST.json`（每檔 sha256）。
- `INDEX.md` 自動重建，是 known-good 清單的人類可讀檢視。
- 定期 `snapshot.py verify <id>` 校驗完整性（sha256 比對；竄改/位元腐化會被抓出）。
- 快照刻意**不進 git**（見 `.gitignore`）——回滾不依賴版本控制。
- 清理：已 done 且其階段整段穩定後，舊的 pre 快照可手動刪；**保留每個 done 工單的最後一個 post（known-good）**。

### 快照與 worktree 的位置（並行期重要）
`snapshot.py` 的快照寫在「**它所在那棵樹**」的 `dev/snapshots/`（root = 腳本上兩層）。
- **階段 0–1（foundation 在 main）**：快照都在 main 的 `dev/snapshots/`，單一位置，最單純。
- **並行期（工單在某 worktree）**：pre/post 快照寫在**該 worktree** 的 `dev/snapshots/`。
  因此復原一個 in_progress 的 worktree story，要**先 `cd` 進它的 worktree**（路徑在 STATUS.md
  該 story 的 `worktree` 欄）再跑步驟 3。story 驗收 pass、合併回 main 後，那次 **merge 本身**
  就是該 story 在 main 上的 known-good；萬一整棵 worktree 遺失，退回 main 上一個已合併 story 的狀態重做。
- 規則：**在哪棵樹改碼，就在哪棵樹打快照**；STATUS.md 的 `worktree` 欄是找回它的索引。

---

## 六、防呆檢查清單（復原完成前自問）

- [ ] `board.py --check` 印「OK 一致」、0 errors。
- [ ] STATUS.md 頂部「⚠ 需注意」區為空（無殘留 in_progress/blocked，或已逐一處置）。
- [ ] 每個剛標 done 的 story 都有 `last_good_snapshot`，且 `snapshot.py verify` 通過。
- [ ] journal 已記錄本次復原的每個 restore / promote 決定。
- [ ] 才開始認領新 story。
