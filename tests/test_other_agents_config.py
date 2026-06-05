"""P7 — 擴展配置中心至其他 agent（warden/orchestrator/compactor/setup）驗收測試。

驗收：
- 四個 agent 各有 fragments + context policy；各 agent prompt_hash 進 run snapshot。
- real_bible 可見性規則不變（story=0、orchestrator/setup=1、warden/compactor=0）。
- 既有 agent 行為回歸不破（見全套件）。
"""
from __future__ import annotations

import pytest

from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.config.composer import PromptComposer
from core.agents.base import SkillLoader
from core.models import SetupOutput, SceneRegistry, Location, NPCBible, WardenOutput, CompactorOutput


OTHER_AGENTS = ["warden", "orchestrator", "compactor", "setup"]


def _store():
    return Database(":memory:").config_store()


@pytest.mark.parametrize("agent", OTHER_AGENTS)
def test_agent_has_fragments_and_policy(agent):
    store = _store()
    frags = store.get_bound_fragments(agent, "mvp_a_safe")
    assert len(frags) >= 2 and all(f["content"].strip() for f in frags)
    assert store.get_context_policy(agent, "mvp_a_safe") is not None
    assert store.get_agent_config(agent, "mvp_a_safe") is not None


@pytest.mark.parametrize("agent", OTHER_AGENTS)
def test_agent_prompt_composes(agent):
    compiled = PromptComposer(_store()).preview(agent, "mvp_a_safe", {})
    assert compiled.compiled_prompt.strip()
    assert compiled.prompt_hash
    assert compiled.enabled_fragments


def test_real_bible_visibility_rules_unchanged():
    store = _store()
    pol = lambda a: store.get_context_policy(a, "mvp_a_safe")["include_real_bible"]
    assert pol("story") == 0          # ★ story 永不見 real_bible
    assert pol("orchestrator") == 1   # orchestrator 可讀（條件揭露）
    assert pol("setup") == 1          # setup 建構 real_bible
    assert pol("warden") == 0
    assert pol("compactor") == 0


# ── run snapshot 含各 agent 的 prompt_hash ───────────────────────────────────
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
        self.loader = SkillLoader()

    def call(self, agent, context, output_model=None, temperature=None):
        return {"setup": _setup(), "warden": WardenOutput(directive_to_story="繼續"),
                "compactor": CompactorOutput(compressed_summary="s", ledger_updates=[],
                                             archived_beats=[], preserved_foreshadowings=[],
                                             final_usage_estimate=0.1)}[agent]

    def stream(self, agent, context, temperature=None, system_override=None):
        for t in ["你在走廊。", "<<<DECISION>>>", STORY_JSON]:
            yield t


def test_run_snapshot_covers_all_configured_agents():
    db = Database()
    loop = BeatLoop(Caller(), Blackboard(), db, run_id="run-p7", use_kernel=False)
    loop.start({"theme": "x", "npc_count": 1})
    snaps = {s["agent_name"]: s for s in loop.run_config_snapshots()}
    for agent in ["story"] + OTHER_AGENTS:
        assert agent in snaps, f"run snapshot 缺 {agent}"
        assert snaps[agent]["compiled_prompt_hash"]
