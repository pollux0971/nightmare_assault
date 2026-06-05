"""tests/test_setup.py — U09 Setup Agent 測試。

覆蓋範圍：
- roll_personality_axes：決定性（固定 seed）、三軸皆在合法池內。
- run_setup（假 caller）：
    - blackboard.snapshot() 中 real_bible/npc_registry/scene_registry 非空、
      protagonist 有 name、opening_sequence 非空。
    - NPC 有 self_aware 欄位。
    - caller 拋例外 → run_setup 往上拋（B8 不降級）。
- 真 API smoke（無 OPENROUTER_API_KEY 時自動 skip）。
"""
from __future__ import annotations

import json
import os
import random

import pytest

from core.agents.base import SkillCaller, SkillLoader
from core.agents.setup import roll_personality_axes, run_setup
from core.blackboard import Blackboard
from core.models import (
    Interactable,
    LLMResult,
    Location,
    NPCBible,
    SceneRegistry,
    SetupOutput,
)


# ─────────────────────────────────────────────────────────────────────────────
# 合法池常數（與 setup.py 一致）
# ─────────────────────────────────────────────────────────────────────────────

_SPEECH_RHYTHM = {"簡短", "絮叨", "停頓多", "喜歡反問"}
_EMOTIONAL_BASE = {"壓抑", "焦躁", "疏離", "過度熱情"}
_QUIRK = {"搓手", "不看眼睛", "反覆確認", "輕聲自語", "咬指甲"}


# ─────────────────────────────────────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────────────────────────────────────

def _make_setup_output() -> SetupOutput:
    """建立一個合法的 SetupOutput，供假 caller 回傳。"""
    npc = NPCBible(
        name="陳醫師",
        profession="前精神科主任",
        personality="mysterious",
        voice_sample="你確定你看到的是真實的嗎？",
        public_face="好心的老醫師，試圖穩定大家情緒",
        secret_core="他是最初儀式的主謀，知道所有人的命運",
        self_aware=True,
        appearance="白髮蒼蒼，白袍整潔，眼神卻從不直視任何人",
    )
    npc2 = NPCBible(
        name="林護士",
        profession="病房護士",
        personality="nervous",
        voice_sample="我……我不知道，我只是照規定辦事。",
        public_face="盡責的護士，只是有些緊張",
        secret_core="她早已死亡，卻渾然不覺，不斷重複最後一班的輪班",
        self_aware=False,
        appearance="臉色蒼白，制服上有不知何時留下的污漬",
    )
    scene = SceneRegistry(
        current_location="ward_b3",
        known_locations=[
            Location(
                id="ward_b3",
                name="B3 病房",
                description="長廊盡頭的單人病房，燈光不穩定，窗外一片漆黑。",
                discovered=True,
                exits=["corridor"],
                interactables=[
                    Interactable(
                        id="rusted_drawer",
                        type="clue",
                        linked_fragment=None,
                        revealed=False,
                    )
                ],
            ),
            Location(
                id="corridor",
                name="走廊",
                description="消毒水味道極濃的長走廊，盡頭的大門緊鎖。",
                discovered=False,
                exits=["ward_b3", "nurses_station"],
                interactables=[],
            ),
            Location(
                id="nurses_station",
                name="護理站",
                description="玻璃後方一片黑暗，偶有鍵盤聲傳出。",
                discovered=False,
                exits=["corridor"],
                interactables=[],
            ),
        ],
    )
    return SetupOutput(
        real_bible={
            "world_truth": {
                "what_really_happened": "這棟醫院三十年前發生了一場秘密儀式，所有參與者皆死亡，但他們的意識被困在建築物中。",
                "the_threat_is": "被困的意識會侵蝕活人的記憶，讓他們忘記自己是誰，最終成為新的困靈。",
                "deadly_rule": "在完全忘記自己名字之前離開，否則將永久留在此地。",
            },
            "revelation_pool": [
                {
                    "id": "frag_001",
                    "type": "knowledge",
                    "content": "三十年前的儀式記錄",
                    "reveal_condition": {"min_beats": 3},
                }
            ],
            "ending_conditions": [
                {"type": "escape", "trigger": "持有鑰匙並到達出口", "prerequisites": ["has_key"]},
                {"type": "death_mental", "trigger": "完全失憶", "gate": "none"},
            ],
            "atmosphere": ["消毒水味", "不穩定的日光燈", "遠處傳來的呼吸聲"],
        },
        npc_registry=[npc, npc2],
        protagonist={
            "name": "王明",
            "starting_situation": "你是一名調查記者，為了尋找失蹤的妹妹而進入這棟廢棄醫院。",
        },
        scene_registry=scene,
        opening_sequence=[
            "雨水打在破碎的玻璃窗上，你站在這棟被遺棄的精神病院入口。",
            "手機的手電筒照出走廊深處，消毒水的氣味令人作嘔。",
            "遠處傳來一聲輕微的腳步聲——你並不孤單。你要怎麼做？",
        ],
    )


class FakeSetupClient:
    """假 LLM client，回傳預先定義的 SetupOutput JSON。"""

    def __init__(self, output: SetupOutput | None = None, raise_error: bool = False):
        self._output = output or _make_setup_output()
        self._raise_error = raise_error

    def call(self, agent, system, user, temperature, stream=False):
        if self._raise_error:
            raise RuntimeError("假 client 強制拋出例外（模擬 LLM 失敗）")
        return LLMResult(
            text=self._output.model_dump_json(),
            model_used="fake-model",
            input_tokens=10,
            output_tokens=50,
            latency_ms=100,
            success=True,
            error=None,
        )

    def stream(self, agent, system, user, temperature):
        yield ""


def _make_caller(
    tmp_path,
    output: SetupOutput | None = None,
    raise_error: bool = False,
) -> SkillCaller:
    """建立一個使用假 client 的 SkillCaller，SKILL.md 寫在 tmp_path 下。"""
    skills_dir = tmp_path / "skills"
    setup_dir = skills_dir / "setup"
    setup_dir.mkdir(parents=True, exist_ok=True)
    (setup_dir / "SKILL.md").write_text("你是世界建構者。", encoding="utf-8")

    loader = SkillLoader(str(skills_dir))
    client = FakeSetupClient(output=output, raise_error=raise_error)
    return SkillCaller(client, loader)


# ─────────────────────────────────────────────────────────────────────────────
# roll_personality_axes 測試
# ─────────────────────────────────────────────────────────────────────────────

def test_roll_personality_axes_deterministic():
    """固定 seed → 每次結果相同（決定性）。"""
    rng1 = random.Random(42)
    rng2 = random.Random(42)
    result1 = roll_personality_axes(rng1)
    result2 = roll_personality_axes(rng2)
    assert result1 == result2


def test_roll_personality_axes_keys():
    """回傳 dict 有三個正確的鍵。"""
    result = roll_personality_axes(random.Random(0))
    assert set(result.keys()) == {"speech_rhythm", "emotional_base", "quirk"}


def test_roll_personality_axes_values_in_pool():
    """每個軸的值都在合法池內。"""
    rng = random.Random(99)
    # 多骰幾次確認各軸都在池內
    for seed in range(20):
        r = random.Random(seed)
        result = roll_personality_axes(r)
        assert result["speech_rhythm"] in _SPEECH_RHYTHM, (
            f"speech_rhythm '{result['speech_rhythm']}' 不在池內"
        )
        assert result["emotional_base"] in _EMOTIONAL_BASE, (
            f"emotional_base '{result['emotional_base']}' 不在池內"
        )
        assert result["quirk"] in _QUIRK, (
            f"quirk '{result['quirk']}' 不在池內"
        )


def test_roll_personality_axes_no_rng_runs():
    """不傳 rng 時也能正常執行（使用 random 預設）。"""
    result = roll_personality_axes()
    assert "speech_rhythm" in result
    assert "emotional_base" in result
    assert "quirk" in result


# ─────────────────────────────────────────────────────────────────────────────
# run_setup（假 caller）測試
# ─────────────────────────────────────────────────────────────────────────────

def test_run_setup_blackboard_real_bible_nonempty(tmp_path):
    """run_setup 後 blackboard.real_bible 非空。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2, "protagonist_name": "王明", "tone": "恐怖"})
    snap = bb.snapshot()
    assert snap["real_bible"], "real_bible 不應為空"


def test_run_setup_blackboard_npc_registry_nonempty(tmp_path):
    """run_setup 後 blackboard.npc_registry 非空。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2})
    snap = bb.snapshot()
    assert snap["npc_registry"], "npc_registry 不應為空"


def test_run_setup_blackboard_scene_registry_nonempty(tmp_path):
    """run_setup 後 blackboard.scene_registry 非空且有 current_location。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2})
    snap = bb.snapshot()
    assert snap["scene_registry"], "scene_registry 不應為空"
    assert snap["scene_registry"]["current_location"], "scene_registry.current_location 不應為空"


def test_run_setup_blackboard_protagonist_has_name(tmp_path):
    """run_setup 後 blackboard.protagonist 有 name 欄位。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2, "protagonist_name": "王明"})
    snap = bb.snapshot()
    assert snap["protagonist"].get("name"), "protagonist 必須有 name 欄位"


def test_run_setup_returns_nonempty_opening_sequence(tmp_path):
    """run_setup 回傳非空的 opening_sequence。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    opening = run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2})
    assert opening, "opening_sequence 不應為空"
    assert isinstance(opening, list), "opening_sequence 應為 list"
    assert all(isinstance(s, str) for s in opening), "opening_sequence 元素應為 str"


def test_run_setup_npc_has_self_aware(tmp_path):
    """blackboard.npc_registry 中每個 NPC 都有 self_aware 欄位。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2})
    snap = bb.snapshot()
    for npc in snap["npc_registry"]:
        assert "self_aware" in npc, f"NPC {npc.get('name')} 缺少 self_aware 欄位"


def test_run_setup_revealed_bible_initialized(tmp_path):
    """run_setup 後 blackboard.revealed_bible 應被初始化為空結構（非 None）。"""
    caller = _make_caller(tmp_path)
    bb = Blackboard()
    run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2})
    snap = bb.snapshot()
    assert snap["revealed_bible"] is not None, "revealed_bible 不應為 None"
    assert "revealed_fragments" in snap["revealed_bible"], "revealed_bible 應有 revealed_fragments"
    assert "known_atmosphere" in snap["revealed_bible"], "revealed_bible 應有 known_atmosphere"
    assert snap["revealed_bible"]["revealed_fragments"] == [], "初始 revealed_fragments 應為空 list"


# ─────────────────────────────────────────────────────────────────────────────
# B8：不降級測試
# ─────────────────────────────────────────────────────────────────────────────

def test_run_setup_raises_on_caller_error(tmp_path):
    """caller 拋例外時 run_setup 直接往上拋（B8：不降級，不回傳空世界）。"""
    caller = _make_caller(tmp_path, raise_error=True)
    bb = Blackboard()
    with pytest.raises(Exception):
        run_setup(caller, bb, {"theme": "廢棄醫院", "npc_count": 2})
    # 驗證 blackboard 未被寫入（仍是初始空值）
    snap = bb.snapshot()
    assert not snap["real_bible"], "例外後 real_bible 應仍為空（未被寫入）"
    assert not snap["npc_registry"], "例外後 npc_registry 應仍為空（未被寫入）"


# ─────────────────────────────────────────────────────────────────────────────
# 真實 API smoke（無 OPENROUTER_API_KEY 時自動 skip）
# ─────────────────────────────────────────────────────────────────────────────

@pytest.mark.skipif(
    not os.environ.get("OPENROUTER_API_KEY"),
    reason="無 OPENROUTER_API_KEY 環境變數，跳過真實 API smoke 測試",
)
def test_real_smoke(tmp_path):
    """用真實 SkillCaller 執行一次 setup，驗證 SetupOutput 結構完整。"""
    from core.llm.client import OpenRouterClient

    model = os.environ.get("OPENROUTER_SMOKE_MODEL", "openai/gpt-4o-mini")
    real_client = OpenRouterClient({
        "api_key": os.environ["OPENROUTER_API_KEY"],
        "base_url": "https://openrouter.ai/api/v1",
        "agent_models": {"setup": [model]},
        "timeout": 120,
    })

    # 使用專案 skills/ 目錄下的真實 SKILL.md
    loader = SkillLoader("skills")
    caller = SkillCaller(real_client, loader)
    bb = Blackboard()

    opening = run_setup(caller, bb, {
        "theme": "廢棄醫院",
        "npc_count": 2,
        "protagonist_name": "李曉晴",
        "tone": "psychological horror",
    })

    snap = bb.snapshot()

    # 結構完整性驗證
    assert snap["real_bible"], "real_bible 不應為空"
    assert snap["npc_registry"], "npc_registry 不應為空"
    assert snap["scene_registry"], "scene_registry 不應為空"
    assert snap["protagonist"].get("name"), "protagonist 必須有 name"
    assert opening, "opening_sequence 不應為空"

    # NPC 結構驗證
    for npc in snap["npc_registry"]:
        assert "self_aware" in npc, f"NPC {npc.get('name')} 缺少 self_aware"
        assert "secret_core" in npc, f"NPC {npc.get('name')} 缺少 secret_core"
        assert "voice_sample" in npc, f"NPC {npc.get('name')} 缺少 voice_sample"

    # scene_registry 結構驗證
    assert snap["scene_registry"]["current_location"], "缺少 current_location"
    assert snap["scene_registry"]["known_locations"], "known_locations 不應為空"
