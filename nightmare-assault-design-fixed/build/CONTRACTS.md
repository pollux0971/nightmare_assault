# CONTRACTS · 契約錨點

> **用途**：每次叫 AI 寫程式時，把本檔（或相關段落）貼在對話開頭。這些是**不可變接口**——AI 不准改名、改簽名、改結構。它失憶幾次都沒關係，只要每次都看到這份，寫出來的東西就接得起來。
> **規則**：若 AI 想改這裡的任何接口，必須先停下來問你，不能自作主張。

---

## 一、專案結構（檔案放哪，固定）

```
nightmare-assault/
├── core/
│   ├── models.py            # 所有 Pydantic 資料類（契約二）
│   ├── blackboard.py        # Blackboard 狀態容器 + 版本/patch
│   ├── signal.py            # SignalBus
│   ├── llm/
│   │   ├── client.py        # OpenRouterClient（契約三）
│   │   └── parser.py        # StoryParsePipeline（契約四）
│   ├── agents/
│   │   ├── base.py          # SkillCaller 基類
│   │   ├── setup.py orchestrator.py story.py warden.py
│   │   ├── dreaming.py offstage_fate.py compactor.py npc_chat.py
│   ├── memory/
│   │   ├── summary.py       # 滾動摘要 + ledger
│   │   └── snapshot.py      # 快照
│   ├── scene.py             # 場景系統
│   ├── orchestrator_loop.py # beat 主迴圈
│   └── persistence/db.py    # SQLite
├── ui/                      # index.html + css/ + js/ + assets/
├── webview_app.py           # pywebview 入口 + API class
├── skills/                  # SKILL.md（已存在，prompt 來源）
├── tests/                   # 每個工單的測試
├── config/
└── main.py
```

AI 寫任何模組，檔案位置照這張表，不要自創目錄。

---

## 二、核心資料類（Pydantic，名稱與欄位固定）

> 完整版見 design 的 07-data-contracts.md。這裡是**最常被引用、不准改名**的核心。

```python
from pydantic import BaseModel, Field
from typing import Literal

# ── 決策點（story 輸出的結構化部分）──
class Option(BaseModel):
    text: str
    tone: Literal["cautious","bold","evasive","aggressive"]

class BeatMeta(BaseModel):
    beat_number: int
    revelations_touched: list[str] = Field(default_factory=list)
    npcs_present: list[str] = Field(default_factory=list)
    pacing: Literal["calm","rising","peak"] = "calm"
    audio_cue: Literal["normal","silence","sting","swell"] = "normal"

class DecisionPoint(BaseModel):
    situation_recap: str
    decision_type: Literal["action","dialogue"]
    suggested_options: list[Option] = Field(default_factory=list)
    free_input_hint: str = "或描述你想做的事…"
    beat_meta: BeatMeta
    is_narration_only: bool = False

# ── warden ──
class WardenOutput(BaseModel):
    rule_violation: bool = False
    violated_rule: str | None = None
    ending_triggered: Literal["death_physical","death_mental","truth_revealed","escape","transformation"] | None = None
    ending_is_soft: bool = False
    skill_claim: str | None = None
    skill_verdict: Literal["allow","reject"] | None = None
    skill_limitation: str | None = None
    directive_to_story: str

# ── 碎片 / 場景 ──
class Revelation(BaseModel):
    id: str
    type: Literal["knowledge","item","person"]
    content: str
    reveal_condition: dict

class Interactable(BaseModel):
    id: str
    type: Literal["item","clue","corpse","door","npc_trace"]
    linked_fragment: str | None = None
    revealed: bool = False

class Location(BaseModel):
    id: str
    name: str
    description: str
    discovered: bool = False
    exits: list[str] = Field(default_factory=list)
    interactables: list[Interactable] = Field(default_factory=list)

# ── NPC ──
class NPCEvolving(BaseModel):
    emotional_state: dict = Field(default_factory=dict)
    relationship: dict = Field(default_factory=dict)
    intent: Literal["observe","befriend","betray","flee","manipulate"] = "observe"
    revealed_layers: list[str] = Field(default_factory=list)
    emergent_lies: list[str] = Field(default_factory=list)
    personal_arc: str = ""

class NPC(BaseModel):
    name: str
    profession: str
    personality: Literal["leader","nervous","analytical","optimistic","mysterious"]
    voice_sample: str
    public_face: str
    secret_core: str          # 不可變
    self_aware: bool
    appearance: str = ""
    presence: Literal["present","absent","missing","dead"] = "present"
    alignment: Literal["allied","neutral","departed","hostile","dead"] = "neutral"
    offstage_intent: str | None = None
    return_condition: str | None = None
    fate_pressure: float = 0.0
    carried_fragment: str | None = None
    evolving: NPCEvolving = Field(default_factory=NPCEvolving)
```

其餘 agent 輸出類（SetupOutput / OrchestratorOutput / DreamingOutput / OffstageFateOutput / CompactorOutput）見 07，簽名同樣固定。

---

## 三、LLM Client 介面（固定簽名）

```python
class OpenRouterClient:
    def __init__(self, config: dict): ...

    def call(self, agent: str, system: str, user: str,
             temperature: float, stream: bool = False) -> "LLMResult":
        """同步呼叫。stream=False 回完整文字；stream=True 回 generator。
           內含 fallback 鏈與 timeout。每次呼叫寫一筆 llm_traces。"""

    def stream(self, agent: str, system: str, user: str,
               temperature: float):
        """yield token 字串。供 story / npc-chat 串流用。"""

class LLMResult(BaseModel):
    text: str
    model_used: str
    input_tokens: int
    output_tokens: int
    latency_ms: int
    success: bool
    error: str | None = None
```

`agent` 參數決定用哪個模型分層 + 寫 trace 時的標記。fallback 鏈見 design 01 §三。

---

## 四、串流解析介面（固定）

```python
class StreamParser:
    """逐 token 餵入，偵測分隔符，分離 narrative 與 decision JSON。"""
    DELIM_CONTINUE = "<<<CONTINUE>>>"
    DELIM_DECISION = "<<<DECISION>>>"

    def feed(self, token: str) -> list["ParseEvent"]:
        """餵一個 token，回傳這個 token 觸發的事件（可能空）。
           事件型別：NARRATIVE_CHUNK / CONTINUE_PAUSE / DECISION_READY / NARRATION_END"""

    def finalize(self) -> DecisionPoint:
        """串流結束時呼叫。回傳解析+驗證後的 DecisionPoint。
           內含三級 repair（程式碼修復 → LLM repair → fallback UI）。見 07 §三。"""
```

分隔符是**字面常數**，前端 JS 與後端 parser 必須用同樣字串。分隔符可能被拆成多 token，parser 用滑動視窗比對，不可逐 token 等於判斷。

---

## 五、前後端 API（pywebview，固定方法名）

```python
class API:
    def check_config(self) -> dict
    def save_config(self, cfg: dict) -> dict
    def test_model(self, agent: str, model: str) -> dict
    def list_saves(self) -> list[dict]
    def start_game(self, opts: dict) -> dict
    def load_game(self, run_id: str) -> dict
    def submit_decision(self, text: str, input_path: str) -> dict
    def continue_narration(self) -> None
    def validate_custom_input(self, fields: dict) -> dict
    def get_inventory(self) -> list[dict]
    def get_status(self) -> dict
    def save_game_now(self, label: str) -> dict
    def get_game_state(self) -> dict
```

前端 JS 推送函式（後端用 `window.evaluate_js` 呼叫），名稱固定：
```javascript
NA.appendToken(tok)      // 純敘事文字，不含分隔符
NA.onContinue()          // 後端 parser 偵測到 CONTINUE 暫停
NA.onDecision(json)      // 後端已驗證的決策點
NA.onAudioCue(cue)       // 音訊事件
NA.onStatus(status)      // idle/setting_up/generating/awaiting_continue/awaiting_decision/saving/loading/error
NA.onError(error)        // 可恢復/不可恢復錯誤
NA.onBeatComplete()      // beat 結束
```

完整 contract 見 07 §一。

---

## 六、分隔符與常數（單一真相來源）

```python
DELIM_CONTINUE = "<<<CONTINUE>>>"
DELIM_DECISION = "<<<DECISION>>>"
MODEL_TIERS = {"heavy": ..., "medium": ..., "light": ...}  # 從 config 讀
SUMMARY_TOKEN_CAP = 1000
BEAT_WINDOW_SIZE = 6          # 保留最近幾個 beat 原文
NARRATION_ONLY_MAX = 3        # 連續旁白型上限，超過強制決策
CONTEXT_THRESHOLD_L1 = 0.70
CONTEXT_THRESHOLD_L2 = 0.85
CONTEXT_THRESHOLD_L3 = 0.95
```

這些常數放 `core/constants.py`，全專案 import，不准各處寫死不同值。
