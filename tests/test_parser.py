"""U08 StreamParser 測試（承重牆）— 覆蓋 07 §三所有情境。"""
from core.llm.parser import (
    StreamParser, NARRATIVE_CHUNK, CONTINUE_PAUSE, DECISION_READY,
)
from core.models import DecisionPoint

GOOD_JSON = ('{"situation_recap":"你看著門。","decision_type":"action",'
             '"suggested_options":[{"text":"開門","tone":"bold"}],'
             '"beat_meta":{"beat_number":3}}')


def _feed_all(p, tokens):
    evs = []
    for t in tokens:
        evs += p.feed(t)
    return evs


def test_normal_narrative_then_decision():
    p = StreamParser()
    evs = _feed_all(p, ["走廊很長，盡頭有光。", "<<<DECISION>>>", GOOD_JSON])
    types = [e.type for e in evs]
    assert NARRATIVE_CHUNK in types and DECISION_READY in types
    dp = p.finalize()
    assert isinstance(dp, DecisionPoint)
    assert dp.situation_recap == "你看著門。"
    assert dp.suggested_options[0].text == "開門"
    assert "走廊很長" in p.narrative


def test_delimiter_split_across_tokens():
    # <<<DECISION>>> 被拆成多個 token，仍須偵測到（B2）
    p = StreamParser()
    _feed_all(p, ["黑暗中。", "<<<", "DECI", "SION", ">>>", GOOD_JSON])
    assert p.saw_decision
    dp = p.finalize()
    assert dp.situation_recap == "你看著門。"
    assert "黑暗中。" in p.narrative


def test_l1_repair_trailing_comma_and_wrapper_text():
    broken = ('這是決策：{"situation_recap":"S","decision_type":"action",'
              '"suggested_options":[{"text":"逃","tone":"bold"},],'
              '"beat_meta":{"beat_number":1},}  以上。')
    p = StreamParser()
    _feed_all(p, ["敘事。", "<<<DECISION>>>", broken])
    dp = p.finalize()
    assert dp.situation_recap == "S" and not dp.is_narration_only


def test_l2_llm_repair_used_when_l1_fails():
    calls = {"n": 0}

    def repair(broken, hint):
        calls["n"] += 1
        return GOOD_JSON
    p = StreamParser(llm_repair=repair)
    _feed_all(p, ["x", "<<<DECISION>>>", "完全不是 json 的東西 @@@"])
    dp = p.finalize()
    assert calls["n"] == 1
    assert dp.situation_recap == "你看著門。"


def test_l3_fallback_when_all_repair_fails():
    p = StreamParser()  # 無 llm_repair
    _feed_all(p, ["x", "<<<DECISION>>>", "garbage not json"])
    dp = p.finalize()
    assert dp.is_narration_only is False
    assert len(dp.suggested_options) >= 1  # 通用但可玩
    assert dp.situation_recap


def test_continue_pause():
    p = StreamParser()
    evs = _feed_all(p, ["你聽見腳步聲。", "<<<CONTINUE>>>", "然後一切安靜。"])
    types = [e.type for e in evs]
    assert CONTINUE_PAUSE in types
    assert "你聽見腳步聲。" in p.narrative


def test_forgot_decision_narration_only():
    p = StreamParser()
    _feed_all(p, ["這是一段純旁白，沒有決策。"])
    dp = p.finalize(expect_narration=True)
    assert dp.is_narration_only is True
    assert "純旁白" in p.narrative


def test_forgot_decision_unexpected_falls_back():
    p = StreamParser()
    _feed_all(p, ["應該要有決策但沒有。"])
    dp = p.finalize(expect_narration=False)
    assert dp.is_narration_only is False
    assert len(dp.suggested_options) >= 1


def test_narrative_not_recalled_after_repair():
    p = StreamParser()
    _feed_all(p, ["重要敘事不可回收。", "<<<DECISION>>>", "garbage"])
    before = p.narrative
    p.finalize()  # 走 fallback
    assert p.narrative == before
    assert "重要敘事不可回收。" in p.narrative


def test_decision_tolerates_llm_enum_drift():
    """回歸：真實模型常吐非列舉 tone/pacing/audio_cue（analytical/direct/tense…），
    不得因此整個 DecisionPoint 被拒、掉 fallback（修『選項一直只有三個』）。"""
    import json as _json
    from core.llm.parser import StreamParser
    dj = _json.dumps({
        "situation_recap": "走廊盡頭有人影。", "decision_type": "investigate",  # 連 decision_type 也漂移
        "suggested_options": [
            {"text": "靠近", "tone": "analytical"}, {"text": "後退", "tone": "direct"},
            {"text": "衝過去", "tone": "forceful"}],
        "beat_meta": {"pacing": "tense", "audio_cue": "weird"}})
    p = StreamParser(beat_number=3)
    for tok in ["人影站著。", "<<<DECISION>>>", dj]:
        p.feed(tok)
    dp = p.finalize()
    assert "你站在原地" not in dp.situation_recap          # 沒掉 fallback
    assert [o.tone for o in dp.suggested_options] == ["cautious", "bold", "aggressive"]
    assert dp.decision_type == "action"                    # investigate → action
    assert dp.beat_meta.pacing == "rising" and dp.beat_meta.audio_cue == "normal"
