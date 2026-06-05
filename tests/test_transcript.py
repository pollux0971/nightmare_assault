"""U20 故事逐字稿複製測試（headless）。"""
from webview_app import API
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus

from tests.test_progress_integration import FakeCaller


def _api():
    api = API(window=None)
    api._config = {"api_key": "x", "base_url": "y", "agent_models": {}, "timeout": 5}
    api._make_loop = lambda: BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(),
                                      run_id="t", use_kernel=True)
    return api


def test_transcript_empty_before_start():
    assert _api().get_transcript()["text"] == ""


def test_transcript_accumulates_opening_and_beats():
    api = _api()
    api._loop = api._make_loop()
    api._transcript = []
    api._run_start({"theme": "x", "npc_count": 1})
    api._run_step("我打開病房門", "free_text")
    api._run_step("我往前走", "free_text")
    text = api.get_transcript()["text"]
    assert text
    assert text.count("── 第") >= 3          # 開場 + 2 beat
    assert "你在病房醒來" in text             # 開場序列被收錄
    assert "〔" in text                       # 含決策收束句
    assert "往前走" in text or "開門" in text  # 含選項文字


def test_transcript_resets_on_new_game():
    api = _api()
    api._loop = api._make_loop()
    api._transcript = ["舊內容"]
    api.start_game({"theme": "x", "npc_count": 1})   # 重置（背景 thread 會清空）
    assert api._transcript == [] or "舊內容" not in "".join(api._transcript)
