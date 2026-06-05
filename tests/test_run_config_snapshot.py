"""P6 — Run 配置快照 驗收測試。

驗收：
- 新 run → run_config_snapshots 寫入 profile + config_json + compiled_prompt_hash + enabled_fragment_keys。
- 載入存檔優先用該 run snapshot（可重現：hash 可對應回 fragment 集）。
- ENABLE_RUN_CONFIG_SNAPSHOT=0 → 不寫。
"""
from __future__ import annotations

import json
import os

import pytest

from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.config.composer import PromptComposer
from core.models import SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput, CompactorOutput


STORY_JSON = ('{"situation_recap":"走廊。","decision_type":"action",'
              '"suggested_options":[{"text":"前進","tone":"cautious"}],'
              '"beat_meta":{"beat_number":0}}')


def _setup():
    return SetupOutput(
        real_bible={"world_truth": {"what_really_happened": "S", "the_threat_is": "走廊",
                                    "deadly_rule": "不可喝藥水"},
                    "hard_triggers": [], "ending_conditions": [], "revelation_pool": []},
        npc_registry=[NPCBible(name="醫生", profession="醫生", personality="mysterious",
                              voice_sample="你不該來。", public_face="冷靜",
                              secret_core="X", self_aware=True, appearance="白袍")],
        protagonist={"name": "林默", "starting_situation": "大廳"},
        scene_registry=SceneRegistry(current_location="hall",
                                     known_locations=[Location(id="hall", name="大廳",
                                                               description="陰暗", exits=["corridor"])]),
        opening_sequence=["你醒來。"],
    )


class Caller:
    def __init__(self):
        from core.agents.base import SkillLoader
        self.loader = SkillLoader()

    def call(self, agent, context, output_model=None, temperature=None):
        return {"setup": _setup(),
                "warden": WardenOutput(directive_to_story="繼續"),
                "compactor": CompactorOutput(compressed_summary="s", ledger_updates=[],
                                             archived_beats=[], preserved_foreshadowings=[],
                                             final_usage_estimate=0.1)}[agent]

    def stream(self, agent, context, temperature=None, system_override=None):
        for t in ["你在走廊。", "<<<DECISION>>>", STORY_JSON]:
            yield t


def _run(run_id="run-snap"):
    db = Database()
    loop = BeatLoop(Caller(), Blackboard(), db, run_id=run_id, use_kernel=False)
    loop.start({"theme": "x", "npc_count": 1})
    return loop, db


def test_new_run_stores_config_snapshot():
    loop, db = _run()
    snaps = loop.run_config_snapshots()
    assert len(snaps) >= 1
    story = [s for s in snaps if s["agent_name"] == "story"][0]
    assert story["profile_name"] == "mvp_a_safe"
    assert story["compiled_prompt_hash"]
    cfg = json.loads(story["config_json"])
    assert cfg["context_policy"]["include_real_bible"] == 0
    frags = json.loads(story["enabled_fragments_json"])
    assert "story.no_repetition" in frags and "story.kernel_obedience" in frags


def test_stored_hash_is_reproducible():
    """run 存的 hash 應等於以該 profile 重新 compose 的 hash（可重現/可反查）。"""
    loop, db = _run("run-repro")
    story = [s for s in loop.run_config_snapshots() if s["agent_name"] == "story"][0]
    fresh = PromptComposer(db.config_store()).preview("story", "mvp_a_safe", {})
    assert story["compiled_prompt_hash"] == fresh.prompt_hash


def test_snapshot_persists_across_reopen(tmp_path):
    path = str(tmp_path / "s.db")
    db = Database(path)
    loop = BeatLoop(Caller(), Blackboard(), db, run_id="run-persist", use_kernel=False)
    loop.start({"theme": "x", "npc_count": 1})
    db.close()
    db2 = Database(path)
    rows = db2.config_store().get_run_config_snapshots("run-persist")
    assert rows and rows[0]["compiled_prompt_hash"]      # 載入存檔可取回該 run 的 config 版本
    db2.close()


def test_snapshot_disabled_writes_nothing(monkeypatch):
    monkeypatch.setenv("ENABLE_RUN_CONFIG_SNAPSHOT", "false")
    loop, db = _run("run-off")
    assert loop.run_config_snapshots() == []
