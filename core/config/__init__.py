"""core.config — 配置中心 / Story Agent 模組化（階段 P，patch v1.1）。

旁路增量層：story prompt 由 fragment 組裝、可配置/可預覽/可快照/可回滾；
static prompt（skills/story/SKILL.md）永遠保留 fallback。對應契約 dev/CONTRACTS.md §十一。

子模組：
- fragments  — 預設 fragment 文字庫 + profile/binding/policy/flag 種子常數（canonical 對齊 patch docs/05）。
- schema     — additive 配置表 DDL + 種子 + ConfigStore（CRUD）。
- composer   — PromptComposer（決定性編譯 + prompt_hash + preview，P2）。
- runtime    — config-first prompt 來源解析 + static fallback（P4）。
"""
from __future__ import annotations
