"""P4 — Runtime 整合（config-first + static fallback）驗收測試。

驗收：
- ENABLE_CONFIG_CENTER=true → story 用 config 編譯 prompt；active 缺 → 退 default；全失敗 → 退 static 仍跑完 beat。
- flag 優先序 .env > active profile > default > hardcoded。
- 30-beat smoke ON/OFF 皆不崩。
"""
from __future__ import annotations

import json

import pytest

from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.config.runtime import ConfigPromptSource
from core.config.flags import resolve_flag, config_center_enabled
from core.agents.base import SkillLoader
from core.models import SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput, CompactorOutput


STORY_JSON = ('{"situation_recap":"前方走廊。","decision_type":"action",'
              '"suggested_options":[{"text":"往走廊深處走","tone":"cautious"}],'
              '"beat_meta":{"beat_number":0}}')


def _setup():
    return SetupOutput(
        real_bible={"world_truth": {"what_really_happened": "機密SECRET", "the_threat_is": "走廊",
                                    "deadly_rule": "不可喝藥水"},
                    "hard_triggers": [], "ending_conditions": [], "revelation_pool": []},
        npc_registry=[NPCBible(name="醫生", profession="醫生", personality="mysterious",
                              voice_sample="你不該來。", public_face="冷靜",
                              secret_core="兇手SECRET", self_aware=True, appearance="白袍")],
        protagonist={"name": "林默", "starting_situation": "大廳醒來"},
        scene_registry=SceneRegistry(current_location="hall",
                                     known_locations=[Location(id="hall", name="大廳",
                                                               description="陰暗", exits=["corridor"])]),
        opening_sequence=["你在大廳醒來。"],
    )


class RecordingCaller:
    """legacy 流程假 caller；記錄每次 story stream 收到的 system_override。"""
    def __init__(self):
        self.loader = SkillLoader()
        self.warden_out = WardenOutput(directive_to_story="繼續")
        self.compactor_out = CompactorOutput(compressed_summary="仍在醫院。", ledger_updates=[],
                                             archived_beats=[], preserved_foreshadowings=[],
                                             final_usage_estimate=0.2)
        self.systems: list = []

    def call(self, agent, context, output_model=None, temperature=None):
        return {"setup": _setup(), "warden": self.warden_out, "compactor": self.compactor_out}[agent]

    def stream(self, agent, context, temperature=None, system_override=None):
        assert agent == "story"
        self.systems.append(system_override)
        for t in ["你站在走廊。", "<<<DECISION>>>", STORY_JSON]:
            yield t


def _loop(db, caller):
    return BeatLoop(caller, Blackboard(), db, run_id="run-p4", use_kernel=False)


def _db_with_flag(value: int) -> Database:
    db = Database()
    db.config_store().set_flag("ENABLE_CONFIG_CENTER", value)
    return db


# ── flag 優先序 ──────────────────────────────────────────────────────────────
def test_flag_precedence_env_over_db(monkeypatch):
    store = _db_with_flag(1).config_store()           # DB 說 on
    monkeypatch.setenv("ENABLE_CONFIG_CENTER", "false")
    assert resolve_flag("ENABLE_CONFIG_CENTER", store) is False   # env 覆寫 DB
    monkeypatch.setenv("ENABLE_CONFIG_CENTER", "true")
    assert resolve_flag("ENABLE_CONFIG_CENTER", store) is True


def test_flag_db_over_hardcoded():
    assert config_center_enabled(_db_with_flag(1).config_store()) is True     # DB on
    assert config_center_enabled(_db_with_flag(0).config_store()) is False    # DB off → hardcoded off
    assert config_center_enabled(None) is False                              # 無 store → hardcoded off


# ── config-first 來源 ────────────────────────────────────────────────────────
def test_config_off_is_static_no_override():
    db = Database()                                   # flag 預設 off
    caller = RecordingCaller()
    loop = _loop(db, caller)
    assert loop._prompt_source is None
    loop.start({"theme": "x", "npc_count": 1})
    assert caller.systems and all(s is None for s in caller.systems)   # 未傳 override → SkillCaller 用 static


def test_config_on_uses_composed_prompt():
    db = _db_with_flag(1)
    caller = RecordingCaller()
    loop = _loop(db, caller)
    assert loop._prompt_source is not None
    loop.start({"theme": "x", "npc_count": 1})
    sys_used = caller.systems[-1]
    assert sys_used is not None and "不是世界狀態裁判" in sys_used   # 來自組裝 fragment
    assert loop._last_prompt_meta["source"] == "config"
    assert loop._last_prompt_meta["prompt_hash"]


def test_active_missing_falls_back_to_default_profile():
    db = _db_with_flag(1)
    store = db.config_store()
    store.set_active_profile("creative")              # creative 無 story binding
    ps = ConfigPromptSource(store, SkillLoader())
    system, meta = ps.story_system_prompt({})
    assert meta["source"] == "config"
    assert meta["profile"] == store.default_profile()  # 退到 default(mvp_a_safe)
    assert "不是世界狀態裁判" in system


def test_full_config_failure_falls_back_to_static():
    db = _db_with_flag(1)
    store = db.config_store()
    # 砍掉 default profile 的所有 story binding → active 與 default 都組不出 → 退 static
    store.connection.execute("DELETE FROM agent_prompt_bindings WHERE agent_name='story';")
    store.connection.commit()
    ps = ConfigPromptSource(store, SkillLoader())
    system, meta = ps.story_system_prompt({})
    assert meta["source"] == "static"
    assert system == SkillLoader().get("story")        # 等於 SKILL.md static fallback


def test_compose_exception_does_not_crash(monkeypatch):
    """compose 全程拋例外 → 仍回 static，不 crash。"""
    db = _db_with_flag(1)
    ps = ConfigPromptSource(db.config_store(), SkillLoader())
    import core.config.runtime as rt
    monkeypatch.setattr(rt, "compose_story_prompt",
                        lambda *a, **k: (_ for _ in ()).throw(RuntimeError("boom")))
    system, meta = ps.story_system_prompt({})
    assert meta["source"] == "static" and system


# ── 30-beat smoke ON/OFF ─────────────────────────────────────────────────────
@pytest.mark.parametrize("flag", [0, 1])
def test_30beat_smoke_no_crash(flag):
    db = _db_with_flag(flag)
    caller = RecordingCaller()
    loop = _loop(db, caller)
    loop.start({"theme": "x", "npc_count": 1})
    for i in range(30):
        out = loop.step(f"我往前走 {i}")
        assert out["narrative"]
    assert loop.beat_number >= 30                       # 連續 30 beat 不崩
    if flag:
        assert any(s is not None for s in caller.systems)
