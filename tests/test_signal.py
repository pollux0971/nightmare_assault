"""tests/test_signal.py — 驗證 core/signal.py SignalBus 行為（U02）。"""
from __future__ import annotations

import pytest

from core.signal import SignalBus


@pytest.fixture
def bus() -> SignalBus:
    return SignalBus()


class TestSubscribeAndPublish:
    def test_handler_called_on_publish(self, bus: SignalBus):
        calls: list[tuple] = []
        bus.subscribe("EVT_A", lambda *a, **kw: calls.append((a, kw)))
        bus.publish("EVT_A", 1, 2, key="val")
        assert len(calls) == 1
        assert calls[0] == ((1, 2), {"key": "val"})

    def test_duplicate_subscribe_ignored(self, bus: SignalBus):
        calls: list[int] = []

        def handler():
            calls.append(1)

        bus.subscribe("EVT", handler)
        bus.subscribe("EVT", handler)  # 重複，應忽略
        bus.publish("EVT")
        assert len(calls) == 1

    def test_multiple_handlers_all_called(self, bus: SignalBus):
        order: list[str] = []
        bus.subscribe("EVT", lambda: order.append("first"))
        bus.subscribe("EVT", lambda: order.append("second"))
        bus.subscribe("EVT", lambda: order.append("third"))
        bus.publish("EVT")
        assert order == ["first", "second", "third"]

    def test_publish_no_subscribers_noop(self, bus: SignalBus):
        # 不應拋出任何例外
        bus.publish("UNKNOWN_EVENT")

    def test_publish_passes_args_and_kwargs(self, bus: SignalBus):
        received: list = []

        def handler(x, y, z=None):
            received.append((x, y, z))

        bus.subscribe("EVT", handler)
        bus.publish("EVT", 10, 20, z=30)
        assert received == [(10, 20, 30)]


class TestUnsubscribe:
    def test_unsubscribe_stops_handler(self, bus: SignalBus):
        calls: list[int] = []

        def handler():
            calls.append(1)

        bus.subscribe("EVT", handler)
        bus.publish("EVT")
        assert len(calls) == 1

        bus.unsubscribe("EVT", handler)
        bus.publish("EVT")
        assert len(calls) == 1  # 退訂後不再被呼叫

    def test_unsubscribe_nonexistent_handler_noop(self, bus: SignalBus):
        # 退訂不存在的 handler 不應拋出例外
        bus.unsubscribe("EVT", lambda: None)

    def test_unsubscribe_nonexistent_event_noop(self, bus: SignalBus):
        bus.unsubscribe("DOES_NOT_EXIST", lambda: None)

    def test_unsubscribe_only_target_handler(self, bus: SignalBus):
        calls: list[str] = []

        def h1():
            calls.append("h1")

        def h2():
            calls.append("h2")

        bus.subscribe("EVT", h1)
        bus.subscribe("EVT", h2)
        bus.unsubscribe("EVT", h1)
        bus.publish("EVT")
        assert calls == ["h2"]


class TestOnce:
    def test_once_fires_exactly_once(self, bus: SignalBus):
        calls: list[int] = []
        bus.once("EVT", lambda: calls.append(1))
        bus.publish("EVT")
        bus.publish("EVT")
        bus.publish("EVT")
        assert len(calls) == 1

    def test_once_receives_args(self, bus: SignalBus):
        received: list = []
        bus.once("EVT", lambda v: received.append(v))
        bus.publish("EVT", 42)
        assert received == [42]

    def test_once_and_subscribe_coexist(self, bus: SignalBus):
        permanent: list[int] = []
        one_shot: list[int] = []

        bus.subscribe("EVT", lambda: permanent.append(1))
        bus.once("EVT", lambda: one_shot.append(1))

        bus.publish("EVT")  # 兩個都觸發
        bus.publish("EVT")  # 只有 permanent 觸發

        assert len(permanent) == 2
        assert len(one_shot) == 1


class TestExceptionIsolation:
    def test_exception_does_not_stop_subsequent_handlers(self, bus: SignalBus):
        calls: list[str] = []

        def bad_handler():
            raise RuntimeError("intentional error")

        def good_handler():
            calls.append("good")

        bus.subscribe("EVT", bad_handler)
        bus.subscribe("EVT", good_handler)
        bus.publish("EVT")  # bad_handler 拋例外，good_handler 仍應被呼叫

        assert calls == ["good"]

    def test_multiple_exceptions_all_others_still_called(self, bus: SignalBus):
        calls: list[str] = []

        def err1():
            raise ValueError("err1")

        def ok():
            calls.append("ok")

        def err2():
            raise TypeError("err2")

        bus.subscribe("EVT", err1)
        bus.subscribe("EVT", ok)
        bus.subscribe("EVT", err2)
        bus.publish("EVT")

        assert calls == ["ok"]

    def test_once_handler_exception_still_unsubscribes(self, bus: SignalBus):
        """once handler 拋例外後，第二次 publish 不應再呼叫它。"""
        calls: list[int] = []

        def bad_once():
            calls.append(1)
            raise RuntimeError("once error")

        bus.once("EVT", bad_once)
        bus.publish("EVT")  # 觸發並拋例外，但應已自動退訂
        bus.publish("EVT")  # 不應再呼叫

        assert len(calls) == 1


class TestPublishUnknownEvent:
    def test_publish_unknown_event_noop(self, bus: SignalBus):
        """publish 不存在的事件不應拋出例外。"""
        bus.publish("THIS_EVENT_WAS_NEVER_SUBSCRIBED")

    def test_publish_after_all_unsubscribed(self, bus: SignalBus):
        calls: list[int] = []

        def h():
            calls.append(1)

        bus.subscribe("EVT", h)
        bus.unsubscribe("EVT", h)
        bus.publish("EVT")  # no-op
        assert calls == []
