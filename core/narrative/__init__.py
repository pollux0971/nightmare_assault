"""core.narrative — 敘事控制層（階段 N，patch v0.1）。

旁路控制層（feature-flag `ENABLE_NARRATIVE_CONTROL` 預設 OFF）：開場只放少量高價值元素建立動機、
真相分層揭露、Story Agent 不發明世界觀、結局有因果門檻。不大重構 beat loop；新欄位皆 optional；
story 永不見 real_bible 不變。契約見 dev/CONTRACTS.md §十二。

子模組：
- models          — NarrativeContract / OpeningBlueprint / TruthSeed / QualityGateResult（NC0/NC1）
- opening_director — 從 contract 挑 ≤budget 開場元素 → OpeningBlueprint（NC2）
- reveal_manager  — 真相分層 hidden→…→actionable，不跳級、evidence-gated（NC4）
- quality_gate    — 規則版輸出檢查 + repair/fallback（NC5）
- ending_gate     — 結局因果門檻；0/8 不可 clean escape（NC6）
"""
from __future__ import annotations
