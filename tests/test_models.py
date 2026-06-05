import pytest
from pydantic import ValidationError

from core.models import (
    Option,
    BeatMeta,
    DecisionPoint,
    WorldTruth,
    Revelation,
    Interactable,
    Location,
    SceneRegistry,
    NPCEvolving,
    NPC,
    NPCBible,
    SetupOutput,
    FragmentReveal,
    OrchestratorOutput,
    WardenOutput,
    DreamingOutput,
    OffstageFateOutput,
    LedgerFact,
    CompactorOutput,
    LLMResult,
)


# ---------------------------------------------------------------------------
# Option
# ---------------------------------------------------------------------------

class TestOption:
    def test_valid(self):
        o = Option(text="Stay calm", tone="cautious")
        assert o.text == "Stay calm"
        assert o.tone == "cautious"

    def test_all_tones(self):
        for tone in ("cautious", "bold", "evasive", "aggressive"):
            o = Option(text="x", tone=tone)
            assert o.tone == tone

    def test_invalid_tone(self):
        with pytest.raises(ValidationError):
            Option(text="x", tone="weird")

    def test_missing_text(self):
        with pytest.raises(ValidationError):
            Option(tone="bold")


# ---------------------------------------------------------------------------
# BeatMeta
# ---------------------------------------------------------------------------

class TestBeatMeta:
    def test_valid_minimal(self):
        bm = BeatMeta(beat_number=1)
        assert bm.beat_number == 1
        assert bm.pacing == "calm"
        assert bm.audio_cue == "normal"
        assert bm.revelations_touched == []
        assert bm.npcs_present == []

    def test_valid_full(self):
        bm = BeatMeta(
            beat_number=3,
            revelations_touched=["rev_01"],
            npcs_present=["alice"],
            pacing="peak",
            audio_cue="sting",
        )
        assert bm.pacing == "peak"
        assert bm.audio_cue == "sting"

    def test_invalid_pacing(self):
        with pytest.raises(ValidationError):
            BeatMeta(beat_number=1, pacing="slow")

    def test_invalid_audio_cue(self):
        with pytest.raises(ValidationError):
            BeatMeta(beat_number=1, audio_cue="loud")

    def test_default_factory_no_sharing(self):
        """Two separate BeatMeta instances must NOT share the same list objects."""
        bm1 = BeatMeta(beat_number=1)
        bm2 = BeatMeta(beat_number=2)
        bm1.revelations_touched.append("rev_x")
        assert bm2.revelations_touched == [], (
            "default_factory should give each instance its own list"
        )
        bm1.npcs_present.append("npc_x")
        assert bm2.npcs_present == []


# ---------------------------------------------------------------------------
# DecisionPoint
# ---------------------------------------------------------------------------

class TestDecisionPoint:
    def _beat(self):
        return BeatMeta(beat_number=1)

    def test_valid(self):
        dp = DecisionPoint(
            situation_recap="You are in a dark room.",
            decision_type="action",
            beat_meta=self._beat(),
        )
        assert dp.decision_type == "action"
        assert dp.is_narration_only is False
        assert dp.free_input_hint == "或描述你想做的事…"

    def test_missing_beat_meta(self):
        with pytest.raises(ValidationError):
            DecisionPoint(
                situation_recap="x",
                decision_type="dialogue",
            )

    def test_invalid_decision_type(self):
        with pytest.raises(ValidationError):
            DecisionPoint(
                situation_recap="x",
                decision_type="think",
                beat_meta=self._beat(),
            )

    def test_nested_options(self):
        dp = DecisionPoint(
            situation_recap="x",
            decision_type="dialogue",
            suggested_options=[{"text": "Run", "tone": "bold"}],
            beat_meta=self._beat(),
        )
        assert isinstance(dp.suggested_options[0], Option)
        assert dp.suggested_options[0].tone == "bold"


# ---------------------------------------------------------------------------
# WorldTruth
# ---------------------------------------------------------------------------

class TestWorldTruth:
    def test_valid(self):
        wt = WorldTruth(
            what_really_happened="A cover-up.",
            the_threat_is="The corporation.",
            deadly_rule="Don't trust anyone.",
        )
        assert wt.deadly_rule == "Don't trust anyone."

    def test_missing_field(self):
        with pytest.raises(ValidationError):
            WorldTruth(what_really_happened="x", the_threat_is="y")


# ---------------------------------------------------------------------------
# Revelation
# ---------------------------------------------------------------------------

class TestRevelation:
    def test_valid(self):
        r = Revelation(id="rev_01", type="knowledge", content="The key is hidden.")
        assert r.reveal_condition == {}

    def test_invalid_type(self):
        with pytest.raises(ValidationError):
            Revelation(id="r", type="secret", content="x")

    def test_default_factory_dict(self):
        r1 = Revelation(id="r1", type="item", content="x")
        r2 = Revelation(id="r2", type="item", content="y")
        r1.reveal_condition["k"] = "v"
        assert r2.reveal_condition == {}


# ---------------------------------------------------------------------------
# Interactable
# ---------------------------------------------------------------------------

class TestInteractable:
    def test_valid(self):
        i = Interactable(id="i01", type="clue")
        assert i.linked_fragment is None
        assert i.revealed is False

    def test_invalid_type(self):
        with pytest.raises(ValidationError):
            Interactable(id="i", type="button")

    def test_with_fragment(self):
        i = Interactable(id="i02", type="item", linked_fragment="frag_01", revealed=True)
        assert i.linked_fragment == "frag_01"
        assert i.revealed is True


# ---------------------------------------------------------------------------
# Location
# ---------------------------------------------------------------------------

class TestLocation:
    def test_valid_minimal(self):
        loc = Location(id="loc_01", name="Hall", description="A dark hall.")
        assert loc.discovered is False
        assert loc.exits == []
        assert loc.interactables == []

    def test_nested_interactables_from_dict(self):
        loc = Location(
            id="loc_02",
            name="Lab",
            description="Sterile.",
            interactables=[
                {"id": "i01", "type": "clue"},
                {"id": "i02", "type": "corpse", "linked_fragment": "frag_x", "revealed": True},
            ],
        )
        assert len(loc.interactables) == 2
        assert isinstance(loc.interactables[0], Interactable)
        assert loc.interactables[1].linked_fragment == "frag_x"

    def test_default_factory_no_sharing(self):
        l1 = Location(id="l1", name="A", description="a")
        l2 = Location(id="l2", name="B", description="b")
        l1.exits.append("north")
        assert l2.exits == []


# ---------------------------------------------------------------------------
# SceneRegistry
# ---------------------------------------------------------------------------

class TestSceneRegistry:
    def test_valid(self):
        sr = SceneRegistry(current_location="loc_01")
        assert sr.known_locations == []

    def test_nested_locations_from_dict(self):
        sr = SceneRegistry(
            current_location="loc_01",
            known_locations=[
                {"id": "loc_01", "name": "Hall", "description": "Dark."},
            ],
        )
        assert isinstance(sr.known_locations[0], Location)


# ---------------------------------------------------------------------------
# NPCEvolving
# ---------------------------------------------------------------------------

class TestNPCEvolving:
    def test_valid_defaults(self):
        e = NPCEvolving()
        assert e.intent == "observe"
        assert e.emotional_state == {}
        assert e.relationship == {}
        assert e.revealed_layers == []
        assert e.emergent_lies == []
        assert e.personal_arc == ""

    def test_invalid_intent(self):
        with pytest.raises(ValidationError):
            NPCEvolving(intent="attack")

    def test_default_factory_no_sharing(self):
        e1 = NPCEvolving()
        e2 = NPCEvolving()
        e1.emergent_lies.append("lie_a")
        assert e2.emergent_lies == []


# ---------------------------------------------------------------------------
# NPC
# ---------------------------------------------------------------------------

class TestNPC:
    def _base(self):
        return dict(
            name="Alice",
            profession="Scientist",
            personality="analytical",
            voice_sample="Calm and measured.",
            public_face="Diligent researcher.",
            secret_core="Knows the truth.",
            self_aware=True,
        )

    def test_valid(self):
        npc = NPC(**self._base())
        assert npc.presence == "present"
        assert npc.alignment == "neutral"
        assert npc.fate_pressure == 0.0
        assert isinstance(npc.evolving, NPCEvolving)

    def test_invalid_personality(self):
        d = self._base()
        d["personality"] = "coward"
        with pytest.raises(ValidationError):
            NPC(**d)

    def test_invalid_presence(self):
        d = self._base()
        d["presence"] = "hidden"
        with pytest.raises(ValidationError):
            NPC(**d)

    def test_evolving_from_dict(self):
        d = self._base()
        d["evolving"] = {"intent": "betray", "personal_arc": "Seeks revenge."}
        npc = NPC(**d)
        assert npc.evolving.intent == "betray"
        assert npc.evolving.personal_arc == "Seeks revenge."


# ---------------------------------------------------------------------------
# NPCBible
# ---------------------------------------------------------------------------

class TestNPCBible:
    def test_valid(self):
        nb = NPCBible(
            name="Bob",
            profession="Guard",
            personality="nervous",
            voice_sample="Stutters.",
            public_face="Loyal soldier.",
            secret_core="Planning to defect.",
            self_aware=False,
        )
        assert nb.appearance == ""


# ---------------------------------------------------------------------------
# SetupOutput
# ---------------------------------------------------------------------------

class TestSetupOutput:
    def test_valid_with_nested_scene_registry(self):
        so = SetupOutput(
            scene_registry={"current_location": "loc_01"},
        )
        assert isinstance(so.scene_registry, SceneRegistry)
        assert so.npc_registry == []
        assert so.opening_sequence == []

    def test_npc_registry_from_list_of_dicts(self):
        so = SetupOutput(
            scene_registry={"current_location": "loc_01"},
            npc_registry=[
                {
                    "name": "Alice",
                    "profession": "Doc",
                    "personality": "leader",
                    "voice_sample": "Firm.",
                    "public_face": "Boss.",
                    "secret_core": "Traitor.",
                    "self_aware": True,
                }
            ],
        )
        assert isinstance(so.npc_registry[0], NPCBible)

    def test_missing_scene_registry(self):
        with pytest.raises(ValidationError):
            SetupOutput()


# ---------------------------------------------------------------------------
# FragmentReveal
# ---------------------------------------------------------------------------

class TestFragmentReveal:
    def test_valid(self):
        fr = FragmentReveal(id="frag_01", how_to_reveal="Show the letter.")
        assert fr.id == "frag_01"


# ---------------------------------------------------------------------------
# OrchestratorOutput
# ---------------------------------------------------------------------------

class TestOrchestratorOutput:
    def test_valid_defaults(self):
        oo = OrchestratorOutput()
        assert oo.fragments_to_reveal == []
        assert oo.reasoning == ""

    def test_with_fragments(self):
        oo = OrchestratorOutput(
            fragments_to_reveal=[{"id": "f1", "how_to_reveal": "Talk to Alice."}],
            reasoning="Player progressed far enough.",
        )
        assert isinstance(oo.fragments_to_reveal[0], FragmentReveal)


# ---------------------------------------------------------------------------
# WardenOutput
# ---------------------------------------------------------------------------

class TestWardenOutput:
    def test_valid_minimal(self):
        wo = WardenOutput(directive_to_story="Continue normally.")
        assert wo.rule_violation is False
        assert wo.ending_triggered is None
        assert wo.skill_verdict is None

    def test_valid_with_ending(self):
        wo = WardenOutput(
            directive_to_story="End it.",
            rule_violation=True,
            violated_rule="Spoke the forbidden name.",
            ending_triggered="death_mental",
            ending_is_soft=False,
        )
        assert wo.ending_triggered == "death_mental"

    def test_invalid_ending_triggered(self):
        with pytest.raises(ValidationError):
            WardenOutput(directive_to_story="x", ending_triggered="surrender")

    def test_invalid_skill_verdict(self):
        with pytest.raises(ValidationError):
            WardenOutput(directive_to_story="x", skill_verdict="maybe")

    def test_missing_directive(self):
        with pytest.raises(ValidationError):
            WardenOutput()


# ---------------------------------------------------------------------------
# DreamingOutput
# ---------------------------------------------------------------------------

class TestDreamingOutput:
    def test_valid(self):
        do = DreamingOutput(intent_update="befriend")
        assert do.revealed_layer is None
        assert do.emergent_lie is None
        assert do.reflection_log == ""

    def test_invalid_intent_update(self):
        with pytest.raises(ValidationError):
            DreamingOutput(intent_update="attack")

    def test_missing_intent_update(self):
        with pytest.raises(ValidationError):
            DreamingOutput()


# ---------------------------------------------------------------------------
# OffstageFateOutput
# ---------------------------------------------------------------------------

class TestOffstageFateOutput:
    def test_valid(self):
        ofo = OffstageFateOutput(
            fate_type="missing",
            fate_narrative="Vanished overnight.",
            fragment_delivery="Left a note.",
        )
        assert ofo.scene_seed is None
        assert ofo.reunion_hook == ""

    def test_invalid_fate_type(self):
        with pytest.raises(ValidationError):
            OffstageFateOutput(
                fate_type="wandering",
                fate_narrative="x",
                fragment_delivery="y",
            )

    def test_missing_required(self):
        with pytest.raises(ValidationError):
            OffstageFateOutput(fate_type="corpse")


# ---------------------------------------------------------------------------
# LedgerFact
# ---------------------------------------------------------------------------

class TestLedgerFact:
    def test_valid(self):
        lf = LedgerFact(type="clue", content="The safe code is 1234.")
        assert lf.type == "clue"


# ---------------------------------------------------------------------------
# CompactorOutput
# ---------------------------------------------------------------------------

class TestCompactorOutput:
    def test_valid(self):
        co = CompactorOutput(
            compressed_summary="Three beats summarized.",
            final_usage_estimate=0.42,
        )
        assert co.ledger_updates == []
        assert co.archived_beats == []
        assert co.preserved_foreshadowings == []

    def test_with_ledger_facts(self):
        co = CompactorOutput(
            compressed_summary="x",
            final_usage_estimate=0.1,
            ledger_updates=[{"type": "npc", "content": "Alice is suspicious."}],
        )
        assert isinstance(co.ledger_updates[0], LedgerFact)

    def test_missing_required(self):
        with pytest.raises(ValidationError):
            CompactorOutput(compressed_summary="x")


# ---------------------------------------------------------------------------
# LLMResult
# ---------------------------------------------------------------------------

class TestLLMResult:
    def test_valid_success(self):
        lr = LLMResult(
            text="Hello.",
            model_used="gpt-4o",
            input_tokens=10,
            output_tokens=5,
            latency_ms=200,
            success=True,
        )
        assert lr.error is None

    def test_valid_failure(self):
        lr = LLMResult(
            text="",
            model_used="gpt-4o",
            input_tokens=0,
            output_tokens=0,
            latency_ms=50,
            success=False,
            error="Timeout",
        )
        assert lr.error == "Timeout"

    def test_missing_required(self):
        with pytest.raises(ValidationError):
            LLMResult(text="x", model_used="m", input_tokens=1, output_tokens=1)
