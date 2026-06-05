"""U07 SkillCaller + SkillLoader 測試。"""
import json
import time

import pytest

from core.agents.base import SkillCaller, SkillLoader
from core.models import LLMResult, WardenOutput


def _write_skill(root, agent, content):
    d = root / agent
    d.mkdir(parents=True, exist_ok=True)
    (d / "SKILL.md").write_text(content, encoding="utf-8")


class FakeClient:
    """假 LLM client：回傳預設文字。"""
    def __init__(self, text, success=True, error=None):
        self.text = text
        self.success = success
        self.error = error
        self.last = None

    def call(self, agent, system, user, temperature, stream=False):
        self.last = {"agent": agent, "system": system, "user": user, "temp": temperature}
        return LLMResult(text=self.text, model_used="fake", input_tokens=1,
                         output_tokens=1, latency_ms=1, success=self.success, error=self.error)

    def stream(self, agent, system, user, temperature):
        for tok in ["a", "b", "c"]:
            yield tok


def test_loader_reads_and_hot_reloads(tmp_path):
    _write_skill(tmp_path, "setup", "原始 prompt")
    loader = SkillLoader(str(tmp_path))
    assert loader.get("setup") == "原始 prompt"
    time.sleep(0.01)
    _write_skill(tmp_path, "setup", "更新後 prompt")
    assert loader.get("setup") == "更新後 prompt"   # 熱重載


def test_loader_missing_raises(tmp_path):
    with pytest.raises(FileNotFoundError):
        SkillLoader(str(tmp_path)).get("ghost")


def test_call_validates_output_model(tmp_path):
    _write_skill(tmp_path, "warden", "你是裁判。")
    out = {"directive_to_story": "繼續", "rule_violation": False}
    caller = SkillCaller(FakeClient(json.dumps(out)), SkillLoader(str(tmp_path)))
    result = caller.call("warden", {"player": "我推門"}, output_model=WardenOutput)
    assert isinstance(result, WardenOutput)
    assert result.directive_to_story == "繼續" and result.rule_violation is False


def test_call_builds_prompt_system_from_skill(tmp_path):
    _write_skill(tmp_path, "warden", "SKILL內容XYZ")
    fc = FakeClient(json.dumps({"directive_to_story": "ok"}))
    SkillCaller(fc, SkillLoader(str(tmp_path))).call("warden", {"k": "v"}, output_model=WardenOutput)
    assert fc.last["system"] == "SKILL內容XYZ"      # system=SKILL 內容
    assert "v" in fc.last["user"]                    # user=結構化 context


def test_call_without_model_returns_text(tmp_path):
    _write_skill(tmp_path, "story", "敘事者")
    caller = SkillCaller(FakeClient("一段敘事文字"), SkillLoader(str(tmp_path)))
    assert caller.call("story", {"beat": 1}) == "一段敘事文字"


def test_call_raises_on_client_failure(tmp_path):
    _write_skill(tmp_path, "warden", "x")
    caller = SkillCaller(FakeClient("", success=False, error="boom"), SkillLoader(str(tmp_path)))
    with pytest.raises(RuntimeError):
        caller.call("warden", {}, output_model=WardenOutput)


def test_call_raises_on_bad_json(tmp_path):
    _write_skill(tmp_path, "warden", "x")
    caller = SkillCaller(FakeClient("不是 json"), SkillLoader(str(tmp_path)))
    with pytest.raises(Exception):
        caller.call("warden", {}, output_model=WardenOutput)


def test_stream_yields_tokens(tmp_path):
    _write_skill(tmp_path, "story", "敘事者")
    caller = SkillCaller(FakeClient("x"), SkillLoader(str(tmp_path)))
    assert list(caller.stream("story", {"beat": 1})) == ["a", "b", "c"]


def test_temperature_override_and_default(tmp_path):
    _write_skill(tmp_path, "warden", "x")
    fc = FakeClient(json.dumps({"directive_to_story": "ok"}))
    caller = SkillCaller(fc, SkillLoader(str(tmp_path)), temperature_by_agent={"warden": 0.3})
    caller.call("warden", {}, output_model=WardenOutput)
    assert fc.last["temp"] == 0.3
    caller.call("warden", {}, output_model=WardenOutput, temperature=0.9)
    assert fc.last["temp"] == 0.9
