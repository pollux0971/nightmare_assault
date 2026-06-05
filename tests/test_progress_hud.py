"""SK07 HUD 進度資訊推送測試（kernel 模式）。"""
from webview_app import API
from core.orchestrator_loop import BeatLoop
from core.blackboard import Blackboard
from core.persistence.db import Database
from core.signal import SignalBus

from tests.test_progress_integration import FakeCaller


class FakeWindow:
    def __init__(self):
        self.calls = []

    def evaluate_js(self, js):
        self.calls.append(js)


def test_kernel_pushes_progress_info_to_hud():
    win = FakeWindow()
    api = API(window=win)
    api._config = {"api_key": "x", "base_url": "y", "agent_models": {}, "timeout": 5}
    api._make_loop = lambda: BeatLoop(FakeCaller(), Blackboard(), Database(), SignalBus(),
                                      run_id="t", use_kernel=True)
    api._loop = api._make_loop()
    api._run_start({"theme": "x", "npc_count": 1})
    win.calls.clear()
    api._run_step("我打開病房門", "free_text")
    joined = "\n".join(win.calls)
    assert "NA.onProgressInfo" in joined
    assert "event.exit_start" in joined          # HUD 顯示推進事件
    assert "推進判定" in joined                      # kernel 階段 label
