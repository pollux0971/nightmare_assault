"""tests/test_warden.py — U12 Warden Agent 驗收測試（B9 設計）。

測試涵蓋：
  B9-1  本地硬規則 + LLM 全失敗 → 仍觸發 rule_violation（關鍵不變式）
  B9-2  安全動作 + 正常 caller → rule_violation=False
  B9-3  技能宣稱：caller 回 skill_verdict="reject" → run_warden 透傳結果
  B9-4  未命中硬規則 + caller 拋例外 → 保守「正常推進」、rule_violation=False
  B9-5  硬結局 trigger 命中 → ending_triggered 正確、不需 LLM
  B9-6  正則 hard_trigger 比對
  B9-7  check_hard_rule 全未命中回 None
  B9-8  無 caller 且無硬規則命中 → 保守推進
  B9-9  有 gate 的 ending_condition 不被本地規則觸發（留給 LLM）
"""
from __future__ import annotations

import json

import pytest

from core.agents.warden import check_hard_rule, run_warden
from core.blackboard import Blackboard
from core.models import LLMResult, WardenOutput


# ─────────────────────────────────────────────────────────────────────────────
# 輔助 fixtures / fakes
# ─────────────────────────────────────────────────────────────────────────────

def _make_blackboard(real_bible: dict) -> Blackboard:
    """建一個裝載指定 real_bible 的 Blackboard。"""
    bb = Blackboard()
    # 用 setup 寫入（唯一有權限寫 real_bible 的 writer）
    bb.write("setup", "real_bible", real_bible)
    return bb


def _real_bible(
    deadly_rule: str = "午夜後不能回頭",
    hard_triggers: list[str] | None = None,
    ending_conditions: list[dict] | None = None,
) -> dict:
    return {
        "deadly_rule": deadly_rule,
        "hard_triggers": hard_triggers if hard_triggers is not None else ["回頭", "看後面"],
        "ending_conditions": ending_conditions if ending_conditions is not None else [],
    }


class _FakeClient:
    """假 LLM client，回傳指定 WardenOutput JSON 或模擬失敗。"""

    def __init__(self, output: WardenOutput | None = None, *, fail: bool = False):
        self._output = output
        self._fail = fail
        self.called = False

    def call(self, agent, system, user, temperature, stream=False):
        self.called = True
        if self._fail:
            return LLMResult(
                text="",
                model_used="fake",
                input_tokens=0,
                output_tokens=0,
                latency_ms=0,
                success=False,
                error="模擬 LLM 失敗",
            )
        text = self._output.model_dump_json() if self._output else json.dumps({"directive_to_story": "正常推進"})
        return LLMResult(
            text=text,
            model_used="fake",
            input_tokens=1,
            output_tokens=1,
            latency_ms=1,
            success=True,
        )


class _FakeCaller:
    """假 SkillCaller，直接回傳或拋例外。"""

    def __init__(self, output: WardenOutput | None = None, *, raise_exc: bool = False):
        self._output = output
        self._raise = raise_exc
        self.called = False

    def call(self, agent: str, context: dict, output_model=None, temperature=None):
        self.called = True
        if self._raise:
            raise RuntimeError("模擬 LLM 全失敗（網路 / 解析 / 驗證錯誤）")
        return self._output or WardenOutput(
            rule_violation=False,
            directive_to_story="正常推進",
        )


# ─────────────────────────────────────────────────────────────────────────────
# check_hard_rule 單元測試
# ─────────────────────────────────────────────────────────────────────────────

class TestCheckHardRule:
    def test_returns_none_when_no_match(self):
        """未命中任何 hard_trigger 或硬結局 → 回 None。"""
        rb = _real_bible(hard_triggers=["回頭"])
        result = check_hard_rule("我走向大門", rb)
        assert result is None

    def test_hard_trigger_keyword_match(self):
        """關鍵詞命中 hard_trigger → rule_violation=True, ending=death_physical。"""
        rb = _real_bible(hard_triggers=["回頭", "看後面"])
        result = check_hard_rule("我轉身回頭看", rb)
        assert result is not None
        assert result.rule_violation is True
        assert result.ending_triggered == "death_physical"
        assert result.violated_rule == rb["deadly_rule"]

    def test_hard_trigger_case_insensitive(self):
        """大小寫不敏感。"""
        rb = _real_bible(hard_triggers=["DRINK", "喝下"])
        result = check_hard_rule("我喝下那瓶藥水", rb)
        assert result is not None
        assert result.rule_violation is True

    def test_hard_trigger_regex(self):
        """B9-6：trigger 為正則表示式時正確比對。"""
        rb = _real_bible(hard_triggers=[r"喝.*藥水", "回頭"])
        result = check_hard_rule("我慢慢地喝掉那瓶神秘藥水", rb)
        assert result is not None
        assert result.rule_violation is True
        assert result.ending_triggered == "death_physical"

    def test_hard_ending_condition_no_gate(self):
        """B9-5：硬結局 ending_condition（無 gate）命中 → ending_triggered 正確。"""
        rb = _real_bible(
            hard_triggers=[],
            ending_conditions=[
                {"type": "death_mental", "trigger": "接受現實", "gate": None},
            ],
        )
        result = check_hard_rule("我決定接受現實，放棄抵抗", rb)
        assert result is not None
        assert result.ending_triggered == "death_mental"
        assert result.ending_is_soft is False

    def test_ending_condition_with_gate_skipped(self):
        """B9-9：有 gate 的 ending_condition 不被本地規則觸發（留給 LLM）。"""
        rb = _real_bible(
            hard_triggers=[],
            ending_conditions=[
                {
                    "type": "truth_revealed",
                    "trigger": "揭開面具",
                    "gate": {"min_beats": 5, "min_revelations": 3},
                }
            ],
        )
        result = check_hard_rule("我揭開那個人的面具", rb)
        # 有 gate → 本地規則不觸發
        assert result is None

    def test_empty_real_bible(self):
        """real_bible 為空 dict → 回 None，不拋例外。"""
        result = check_hard_rule("任何動作", {})
        assert result is None

    def test_no_triggers_returns_none(self):
        """B9-7：hard_triggers 為空列表 → 回 None。"""
        rb = _real_bible(hard_triggers=[])
        result = check_hard_rule("我回頭看了一眼", rb)
        assert result is None


# ─────────────────────────────────────────────────────────────────────────────
# run_warden 整合測試
# ─────────────────────────────────────────────────────────────────────────────

class TestRunWarden:

    # ── B9-1：關鍵不變式 ────────────────────────────────────────────────────

    def test_b9_1_hard_rule_wins_even_if_llm_fails(self):
        """B9-1（最關鍵）：player_decision 含 hard_trigger，caller 設為拋例外（模擬 LLM 全失敗）
        → run_warden 仍回 rule_violation=True（本地硬規則優先觸發）。"""
        bb = _make_blackboard(_real_bible(hard_triggers=["喝下藥水", "飲用藥水"]))
        # caller 設為一定拋例外
        failing_caller = _FakeCaller(raise_exc=True)

        result = run_warden("我喝下藥水，感覺有些奇怪", bb, caller=failing_caller)

        # 硬規則必須觸發，即使 LLM 完全失敗
        assert result.rule_violation is True
        assert result.ending_triggered == "death_physical"
        assert result.violated_rule is not None
        # caller 不應被呼叫（因硬規則已在 Step 1 命中）
        assert failing_caller.called is False

    def test_b9_1_hard_trigger_regex_llm_fails(self):
        """B9-1 變體：正則 hard_trigger 命中 + LLM 失敗 → 硬規則觸發。"""
        bb = _make_blackboard(_real_bible(hard_triggers=[r"飲.*液體"]))
        failing_caller = _FakeCaller(raise_exc=True)

        result = run_warden("我緩緩飲下那瓶詭異液體", bb, caller=failing_caller)

        assert result.rule_violation is True
        assert result.ending_triggered == "death_physical"

    # ── B9-2：安全動作 + 正常 caller ────────────────────────────────────────

    def test_b9_2_safe_action_normal_caller(self):
        """B9-2：安全動作 + 正常 caller 回不違規 → rule_violation=False。"""
        bb = _make_blackboard(_real_bible(hard_triggers=["回頭"]))
        safe_output = WardenOutput(rule_violation=False, directive_to_story="正常推進")
        caller = _FakeCaller(output=safe_output)

        result = run_warden("我小心翼翼地往前走", bb, caller=caller)

        assert result.rule_violation is False
        assert caller.called is True

    def test_b9_2_caller_result_passed_through(self):
        """B9-2 延伸：caller 回的完整結果原樣傳回。"""
        bb = _make_blackboard(_real_bible(hard_triggers=[]))
        llm_output = WardenOutput(
            rule_violation=False,
            directive_to_story="正常推進，但角色感到不安",
            ending_triggered=None,
        )
        caller = _FakeCaller(output=llm_output)

        result = run_warden("我檢查壁爐旁的照片", bb, caller=caller)

        assert result.directive_to_story == "正常推進，但角色感到不安"
        assert result.rule_violation is False

    # ── B9-3：技能宣稱封頂 ──────────────────────────────────────────────────

    def test_b9_3_skill_claim_reject(self):
        """B9-3：caller 回 skill_verdict="reject" → run_warden 透傳含封頂結果。"""
        bb = _make_blackboard(_real_bible(hard_triggers=[]))
        skill_output = WardenOutput(
            rule_violation=False,
            skill_claim="我是鎖匠，能撬開這電子封印",
            skill_verdict="reject",
            skill_limitation="你的機械技術在電子封印面前毫無用武之地",
            directive_to_story="技能宣稱被拒絕：電子封印超出鎖匠能力範圍",
        )
        caller = _FakeCaller(output=skill_output)

        result = run_warden("我是鎖匠，能撬開這電子封印", bb, caller=caller)

        assert result.skill_verdict == "reject"
        assert result.skill_claim == "我是鎖匠，能撬開這電子封印"
        assert result.skill_limitation is not None
        assert result.rule_violation is False

    def test_b9_3_skill_claim_allow(self):
        """B9-3 允許技能：caller 回 skill_verdict="allow" → 透傳。"""
        bb = _make_blackboard(_real_bible(hard_triggers=[]))
        skill_output = WardenOutput(
            rule_violation=False,
            skill_claim="我曾是醫生，能判斷這藥物成分",
            skill_verdict="allow",
            skill_limitation="你的醫學知識可判斷基本成分，但古老儀式藥物超出你的知識範疇",
            directive_to_story="技能已認可侷限：醫學知識有效但範圍有限",
        )
        caller = _FakeCaller(output=skill_output)

        result = run_warden("我曾是醫生，能判斷這藥物成分", bb, caller=caller)

        assert result.skill_verdict == "allow"
        assert result.skill_limitation is not None

    # ── B9-4：未命中硬規則 + caller 拋例外 → 保守推進 ─────────────────────

    def test_b9_4_safe_action_caller_fails_conservative(self):
        """B9-4：未命中硬規則 + caller 拋例外 → 保守「正常推進」、rule_violation=False。"""
        bb = _make_blackboard(_real_bible(hard_triggers=["回頭"]))
        failing_caller = _FakeCaller(raise_exc=True)

        result = run_warden("我查看書架上的書", bb, caller=failing_caller)

        assert result.rule_violation is False
        assert "正常推進" in result.directive_to_story

    def test_b9_4_no_caller_safe_action(self):
        """B9-8：無 caller 且未命中硬規則 → 保守推進。"""
        bb = _make_blackboard(_real_bible(hard_triggers=["回頭"]))

        result = run_warden("我慢慢打開抽屜", bb, caller=None)

        assert result.rule_violation is False
        assert "正常推進" in result.directive_to_story

    # ── B9-5：硬結局 trigger 命中 ───────────────────────────────────────────

    def test_b9_5_hard_ending_triggered(self):
        """B9-5：硬結局 trigger 命中 → ending_triggered 正確、不需 LLM。"""
        rb = _real_bible(
            hard_triggers=[],
            ending_conditions=[
                {"type": "death_physical", "trigger": "跳入深淵", "gate": None},
            ],
        )
        bb = _make_blackboard(rb)
        # caller 設為拋例外以確認不需要 LLM
        failing_caller = _FakeCaller(raise_exc=True)

        result = run_warden("我決定跳入深淵", bb, caller=failing_caller)

        assert result.ending_triggered == "death_physical"
        assert failing_caller.called is False  # LLM 根本沒被呼叫

    def test_b9_5_death_mental_ending(self):
        """B9-5 變體：death_mental 硬結局觸發。"""
        rb = _real_bible(
            hard_triggers=[],
            ending_conditions=[
                {"type": "death_mental", "trigger": "接受黑暗"},
            ],
        )
        bb = _make_blackboard(rb)

        result = run_warden("我決定接受黑暗，不再抵抗", bb, caller=None)

        assert result.ending_triggered == "death_mental"
        assert result.ending_is_soft is False

    # ── 邊界 / 其他 ──────────────────────────────────────────────────────────

    def test_blackboard_empty_real_bible_no_crash(self):
        """real_bible 為空時 run_warden 不拋例外，降到保守推進。"""
        bb = Blackboard()  # real_bible = {}
        result = run_warden("我做任何事", bb, caller=None)
        assert result.rule_violation is False

    def test_hard_rule_takes_priority_over_llm_positive(self):
        """即使 caller 會回「不違規」，硬規則命中時仍應觸發 rule_violation。"""
        bb = _make_blackboard(_real_bible(hard_triggers=["回頭"]))
        # caller 明確回不違規
        safe_caller = _FakeCaller(
            output=WardenOutput(rule_violation=False, directive_to_story="正常推進")
        )

        result = run_warden("我猛地回頭", bb, caller=safe_caller)

        assert result.rule_violation is True
        assert result.ending_triggered == "death_physical"
        # 由於硬規則在 Step 1 命中，caller 根本不該被呼叫
        assert safe_caller.called is False

    def test_soft_ending_with_gate_goes_to_llm(self):
        """有 gate 的軟結局：本地規則不觸發，交由 caller（LLM）判斷。"""
        rb = _real_bible(
            hard_triggers=[],
            ending_conditions=[
                {
                    "type": "escape",
                    "trigger": "衝出大門",
                    "gate": {"min_beats": 10},
                }
            ],
        )
        bb = _make_blackboard(rb)
        llm_output = WardenOutput(
            rule_violation=False,
            ending_triggered="escape",
            ending_is_soft=True,
            directive_to_story="結局序列：escape（但門被鎖鏈纏住，暫緩）",
        )
        caller = _FakeCaller(output=llm_output)

        result = run_warden("我衝出大門", bb, caller=caller)

        # 本地規則未觸發（有 gate），LLM 回了軟結局
        assert caller.called is True
        assert result.ending_is_soft is True

    def test_directive_contains_death_on_hard_trigger(self):
        """硬規則命中時，directive_to_story 應包含死亡相關指令文字。"""
        bb = _make_blackboard(_real_bible(hard_triggers=["凝視深淵"]))
        result = run_warden("我凝視深淵太久了", bb, caller=None)

        assert result.rule_violation is True
        assert "死亡" in result.directive_to_story or "death" in result.directive_to_story.lower()
