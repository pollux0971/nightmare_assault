---
id: HD2
stage: H
lane: F
title: Surface Sanitizer All Outputs（覆蓋全玩家面 + path/IP/權限）
status: done
worktree: main
depends_on: [NR7]
contracts: [SurfaceSanitizer]
last_good_snapshot: HD2__post__20260605-103125
owner_session: session-24-opus
---

# HD2 · Surface Sanitizer All Outputs

- **階段**：H　**工線**：F（表層安全）
- **依賴**：NR7（SurfaceTextSanitizer）
- **契約**：SurfaceSanitizer（擴充）（見 §十五 / §十四）

## 目標 / 範圍
把消毒擴展到**所有玩家面輸出**（narrative / options / situation_recap / NPC reply / ending rendered_text / observation JSON），並新增 Unix path / IP / 權限/存取權/系統提示 / core 偵測；嚴重污染觸發 repair 而非只刪詞。

## 對應來源
patch `task_cards/P1_D2`、`docs/07`、`reference_code/surface_sanitizer.py`、`tests_reference/test_surface_sanitizer_runtime.py`。

## 實作步驟
- 擴充 `core/narrative/sanitizer.py`：加 PATH_RE（/usr|var|home|data|tmp|etc）、IP_RE、權限/存取權/系統提示/core；`sanitize_options(list)`；`scan` 回 hits 類別。
- loop：options（dp.suggested_options 文字）、situation_recap 也過消毒；ending rendered_text 過消毒。
- npc reply（HC1 路徑）過消毒。
- 嚴重污染（path/IP/權限）→ 標記供 QualityGate repair（接 HD1），不只刪詞。
- in-world 授權詞例外不變。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_sanitizer_runtime_hd2.py（沿用 reference）：含 /usr/local、IP、technical → scan 命中 unix_path/ip_address、sanitize 後皆移除；options 列表逐項消毒；授權詞保留。
- [ ] 既有 sanitizer 測試綠；board --check 0 errors。

## 回滾備註
sanitizer 擴充 + 多輸出接點；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HD2 pre｜完成：snapshot.py snapshot HD2 post --verify pass
