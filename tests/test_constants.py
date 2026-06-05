"""tests/test_constants.py — 驗證 core/constants.py 常數值（U02）。"""
from __future__ import annotations

import core.constants as C


class TestDelimiters:
    def test_delim_continue_value(self):
        assert C.DELIM_CONTINUE == "<<<CONTINUE>>>"

    def test_delim_decision_value(self):
        assert C.DELIM_DECISION == "<<<DECISION>>>"


class TestAllEvents:
    def test_all_events_contains_eight_entries(self):
        assert len(C.ALL_EVENTS) == 8

    def test_beat_completed_in_all_events(self):
        assert C.EVT_BEAT_COMPLETED in C.ALL_EVENTS

    def test_ending_triggered_in_all_events(self):
        assert C.EVT_ENDING_TRIGGERED in C.ALL_EVENTS

    def test_rule_violation_in_all_events(self):
        assert C.EVT_RULE_VIOLATION in C.ALL_EVENTS

    def test_skill_claimed_in_all_events(self):
        assert C.EVT_SKILL_CLAIMED in C.ALL_EVENTS

    def test_chatroom_opened_in_all_events(self):
        assert C.EVT_CHATROOM_OPENED in C.ALL_EVENTS

    def test_chatroom_closed_in_all_events(self):
        assert C.EVT_CHATROOM_CLOSED in C.ALL_EVENTS

    def test_npc_evolved_in_all_events(self):
        assert C.EVT_NPC_EVOLVED in C.ALL_EVENTS

    def test_context_threshold_in_all_events(self):
        assert C.EVT_CONTEXT_THRESHOLD in C.ALL_EVENTS

    def test_all_events_no_duplicates(self):
        assert len(C.ALL_EVENTS) == len(set(C.ALL_EVENTS))


class TestEventNameValues:
    def test_beat_completed_string(self):
        assert C.EVT_BEAT_COMPLETED == "BEAT_COMPLETED"

    def test_ending_triggered_string(self):
        assert C.EVT_ENDING_TRIGGERED == "ENDING_TRIGGERED"

    def test_rule_violation_string(self):
        assert C.EVT_RULE_VIOLATION == "RULE_VIOLATION"

    def test_skill_claimed_string(self):
        assert C.EVT_SKILL_CLAIMED == "SKILL_CLAIMED"

    def test_chatroom_opened_string(self):
        assert C.EVT_CHATROOM_OPENED == "CHATROOM_OPENED"

    def test_chatroom_closed_string(self):
        assert C.EVT_CHATROOM_CLOSED == "CHATROOM_CLOSED"

    def test_npc_evolved_string(self):
        assert C.EVT_NPC_EVOLVED == "NPC_EVOLVED"

    def test_context_threshold_string(self):
        assert C.EVT_CONTEXT_THRESHOLD == "CONTEXT_THRESHOLD"


class TestContextThresholds:
    def test_l1_value(self):
        assert C.CONTEXT_THRESHOLD_L1 == 0.70

    def test_l2_value(self):
        assert C.CONTEXT_THRESHOLD_L2 == 0.85

    def test_l3_value(self):
        assert C.CONTEXT_THRESHOLD_L3 == 0.95

    def test_thresholds_strictly_increasing(self):
        assert C.CONTEXT_THRESHOLD_L1 < C.CONTEXT_THRESHOLD_L2 < C.CONTEXT_THRESHOLD_L3


class TestOtherConstants:
    def test_summary_token_cap(self):
        assert C.SUMMARY_TOKEN_CAP == 1000

    def test_beat_window_size(self):
        assert C.BEAT_WINDOW_SIZE == 6

    def test_narration_only_max(self):
        assert C.NARRATION_ONLY_MAX == 3

    def test_model_tiers_keys(self):
        assert set(C.MODEL_TIERS.keys()) == {"heavy", "medium", "light"}

    def test_model_tiers_defaults_none(self):
        for v in C.MODEL_TIERS.values():
            assert v is None
