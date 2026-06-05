"""core.config.fragments — 預設 prompt fragment 文字庫 + 配置種子常數（P1）。

canonical 文字對齊 `nightmare-assault-story-doc-patch-batch-v1.1/docs/05-prompt-module-library.md`。
這裡是 fragment 內容與預設綁定/政策/profile/flag 的**單一來源**；P1 種子、P2 composer、P3 story、
P6 回歸測試都引用這裡，不在各處硬抄。
"""
from __future__ import annotations

# ── Agent / profile 名稱（P0 凍結，見 config/CONTRACT_FREEZE.md）───────────────
AGENTS = ("setup", "story", "warden", "orchestrator", "compactor")

PROFILES = (
    # (profile_name, description, is_active)
    ("mvp_a_safe", "MVP-A 穩定預設（config-first 的 default）", 1),
    ("debug", "多 log、低創意", 0),
    ("creative", "高自由度、高 temperature", 0),
    ("low_cost", "便宜模型、緊縮 context", 0),
    ("strict_kernel", "最強 ProgressKernel 服從", 0),
)

DEFAULT_PROFILE = "mvp_a_safe"   # config-first 的 default fallback profile

# ── Story Agent fragment 文字（docs/05 verbatim）──────────────────────────────
# 每筆：(fragment_key, title, category, content, description, enabled_by_default)
STORY_FRAGMENTS = [
    (
        "story.role", "Story 角色", "role",
        "你是互動恐怖文字遊戲的敘事生成 agent。\n"
        "你負責把系統已提交的世界狀態變化寫成自然、有張力、可互動的敘事。",
        "定義 Story Agent 角色：敘事 renderer，不是世界裁判。", 1,
    ),
    (
        "story.objective", "Story 目標", "objective",
        "你的目標是產生一個完整但不拖沓的 beat。\n"
        "每個 beat 必須反映目前世界狀態，完成系統指定的敘事義務，並在最後提供下一步選擇與自由輸入。",
        "定義 beat 生成目標。", 1,
    ),
    (
        "story.kernel_obedience", "服從 ProgressKernel", "rules",
        "你不是世界狀態裁判。\n"
        "場景變化、事件是否完成、NPC 是否出現、線索是否取得、道具是否加入背包，"
        "全部由系統提供的 state patches 與 narrative_obligations 決定。\n\n"
        "你必須遵守 narrative_obligations，不得否定、延後或改寫它們。",
        "強制服從 kernel obligations。", 1,
    ),
    (
        "story.no_repetition", "禁止重複已解決事件", "guardrail",
        "禁止重複已完成事件。\n"
        "如果 forbidden_repeats 中包含某個行動、事件、問題或目標，"
        "你不得再次把它寫成玩家需要做的選擇。\n\n"
        "例如：\n"
        "- 門已經打開，不得再次詢問玩家是否要打開同一扇門。\n"
        "- 玩家已經進入走廊，不得無故把玩家寫回原本病房。\n"
        "- 已經獲得的線索不得再次被當成首次發現。",
        "反重複不變式（開門不再問）。", 1,
    ),
    (
        "story.open_choice", "保留自由選擇", "rules",
        "你必須保留玩家自由選擇。\n"
        "可以提供推薦選項，但不得把玩家鎖死在唯一主線。\n"
        "每個 beat 的選項後必須保留自由輸入可能性。\n\n"
        "推薦選項應該根據目前新狀態生成，而不是回到上一個已完成問題。",
        "保留開放式玩家選擇。", 1,
    ),
    (
        "story.context_policy", "上下文政策", "context",
        "你只能使用系統提供的可見上下文。\n"
        "不得要求或推測 hidden real_bible。\n"
        "不得揭露未被 revealed_bible 或 narrative_obligations 提供的真相。\n"
        "不得為了補劇情而加入未提交的重大世界狀態變化。",
        "防 context 爆炸與隱藏資訊外洩（story 永不見 real_bible）。", 1,
    ),
    (
        "story.output_format", "輸出格式", "schema",
        "輸出必須包含兩部分：敘事文字與決策 JSON。\n\n"
        "格式：\n\n"
        "敘事文字\n\n"
        "<<<DECISION>>>\n"
        "{\n"
        '  "choices": [...],\n'
        '  "free_input_enabled": true,\n'
        '  "beat_meta": {\n'
        '    "progress_delta": {...},\n'
        '    "visible_npcs": [...],\n'
        '    "new_clues": [...],\n'
        '    "new_items": [...]\n'
        "  }\n"
        "}",
        "定義分隔符與 decision JSON 契約。", 1,
    ),
    (
        "story.style_horror", "恐怖風格", "style",
        "風格：壓迫、清晰、節奏緊湊。\n"
        "避免過度鋪陳。每段文字都應該推進場景、情緒、線索或風險。",
        "控制散文語氣（選用）。", 1,
    ),
]

# story fragment 組裝順序（docs/04 §agent_prompt_bindings）
STORY_BINDING_ORDER = [
    ("story.role", 10),
    ("story.objective", 20),
    ("story.kernel_obedience", 30),
    ("story.no_repetition", 40),
    ("story.open_choice", 50),
    ("story.context_policy", 60),
    ("story.output_format", 70),
    ("story.style_horror", 80),
]

# ── Story agent_config 預設（docs/04 §agent_configs 建議值）────────────────────
STORY_AGENT_CONFIG = dict(
    agent_name="story",
    enabled=1,
    model=None,                       # None → 用既有 client 的 tier 預設
    temperature=0.75,                 # docs/04：0.65–0.8
    max_output_tokens=1200,
    context_budget_tokens=6000,       # docs/04：5000–8000
    prompt_version="v1",
    output_schema_name="StoryBeatOutput",
    stream_enabled=1,
)

# ── Story agent_context_policy 預設（docs/04；story include_real_bible=false 硬規則）──
STORY_CONTEXT_POLICY = dict(
    agent_name="story",
    max_recent_beats=2,
    max_relevant_clues=5,
    max_relevant_items=8,
    max_visible_npcs=4,
    include_full_history=0,
    include_full_revealed_bible=1,
    include_real_bible=0,             # ★ 硬規則：story 永不見 real_bible（C2/E2）
    include_debug_trace=0,
    retrieval_strategy="relevant_only",
)

# ── 其他 agent fragment（P7；docs/05 verbatim）───────────────────────────────
# agent -> [(fragment_key, title, category, content, description, enabled_by_default)]
OTHER_AGENT_FRAGMENTS: dict[str, list] = {
    "warden": [
        ("warden.role", "Warden 角色", "role",
         "你是玩家行動可行性與風險判定 agent。\n"
         "你只判定玩家行動是否破格、是否觸發致命規則、是否需要代價，不負責創作完整劇情。",
         "定義 warden 角色：可行性/風險裁決。", 1),
        ("warden.local_rule_first", "本地規則優先", "guardrail",
         "本地 deterministic rule 的結果優先於 LLM 語義判斷。\n"
         "若本地規則已判定致命或禁止，不得覆寫為安全。",
         "本地硬規則優先於 LLM（B9）。", 1),
    ],
    "orchestrator": [
        ("orchestrator.role", "Orchestrator 角色", "role",
         "你負責判斷目前狀態是否應該揭露新的真相碎片。\n"
         "你可以讀 real_bible，但你只能把符合條件的最小碎片加入 revealed_bible。",
         "定義 orchestrator 角色：條件揭露。", 1),
        ("orchestrator.no_over_reveal", "不過度揭露", "guardrail",
         "不得一次揭露過多真相。\n"
         "每次只揭露與玩家當前線索、場景、NPC 或事件直接相關的資訊。",
         "限制每次揭露量。", 1),
    ],
    "compactor": [
        ("compactor.role", "Compactor 角色", "role",
         "你負責壓縮歷史上下文，保留已發生事實、未解伏筆、重要 NPC 關係、玩家承諾與世界狀態。",
         "定義 compactor 角色：壓縮+保護伏筆。", 1),
        ("compactor.fact_ledger_policy", "fact ledger 政策", "rules",
         "fact ledger 只記錄已確定事實，不記錄猜測。\n"
         "猜測與推論應放入 open_threads 或 player_hypotheses。",
         "ledger 只記事實。", 1),
    ],
    "setup": [
        ("setup.role", "Setup 角色", "role",
         "你負責根據主題建立初始世界設定、real_bible、revealed_bible 起點、NPC 種子、場景種子與開場事件。",
         "定義 setup 角色：建初始世界。", 1),
        ("setup.opportunity_graph_policy", "機會圖政策", "rules",
         "你不得生成唯一主線。\n"
         "你應生成 opportunity graph：場景、可互動物件、線索、NPC、威脅、轉場機會與 ending attractors。",
         "生成機會圖而非單一主線。", 1),
    ],
}

# agent -> [(fragment_key, sort_order)]
OTHER_AGENT_BINDINGS: dict[str, list] = {
    agent: [(f[0], 10 + 10 * i) for i, f in enumerate(frags)]
    for agent, frags in OTHER_AGENT_FRAGMENTS.items()
}

# 其他 agent 的 agent_config（docs/04 建議值）。stream 一律 false（非串流 agent）。
OTHER_AGENT_CONFIGS: dict[str, dict] = {
    "warden": dict(temperature=0.2, max_output_tokens=600, context_budget_tokens=3000,
                   output_schema_name="WardenOutput"),
    "orchestrator": dict(temperature=0.4, max_output_tokens=800, context_budget_tokens=4000,
                         output_schema_name="OrchestratorOutput"),
    "compactor": dict(temperature=0.3, max_output_tokens=1000, context_budget_tokens=6000,
                      output_schema_name="CompactorOutput"),
    "setup": dict(temperature=0.8, max_output_tokens=2000, context_budget_tokens=6000,
                  output_schema_name="SetupOutput"),
}

# 其他 agent 的 context policy（include_real_bible：orchestrator/setup=1，warden/compactor=0；story 永遠 0）。
OTHER_AGENT_CONTEXT_POLICIES: dict[str, dict] = {
    "warden": dict(max_recent_beats=2, max_relevant_clues=5, max_relevant_items=5, max_visible_npcs=4,
                   include_full_history=0, include_full_revealed_bible=1, include_real_bible=0,
                   include_debug_trace=0, retrieval_strategy="relevant_only"),
    "orchestrator": dict(max_recent_beats=3, max_relevant_clues=8, max_relevant_items=5, max_visible_npcs=5,
                         include_full_history=0, include_full_revealed_bible=1, include_real_bible=1,
                         include_debug_trace=0, retrieval_strategy="relevant_only"),
    "compactor": dict(max_recent_beats=6, max_relevant_clues=10, max_relevant_items=10, max_visible_npcs=6,
                      include_full_history=1, include_full_revealed_bible=1, include_real_bible=0,
                      include_debug_trace=0, retrieval_strategy="recent_first"),
    "setup": dict(max_recent_beats=0, max_relevant_clues=0, max_relevant_items=0, max_visible_npcs=0,
                  include_full_history=0, include_full_revealed_bible=0, include_real_bible=1,
                  include_debug_trace=0, retrieval_strategy="full"),
}

# ── Feature flags 種子（docs/04；profile 層預設值，.env 可覆寫）────────────────
FEATURE_FLAGS = [
    # (flag_name, default_value)
    ("ENABLE_CONFIG_CENTER", 0),
    ("ENABLE_PROMPT_PREVIEW", 1),
    ("ENABLE_RUN_CONFIG_SNAPSHOT", 1),
]
