"""SignalBus — 全域發布／訂閱事件匯流排（U02）。

設計原則：
- 訂閱順序即呼叫順序。
- 同一 handler 重複訂閱同一事件不重覆加入。
- 一個 handler 拋出例外不中斷其餘 handler（例外隔離）。
- publish 不存在的事件名 no-op（不報錯）。
- once() 訂閱的 handler 觸發一次後自動退訂。
"""
from __future__ import annotations

import logging
from collections import defaultdict
from typing import Any, Callable

logger = logging.getLogger(__name__)

Handler = Callable[..., Any]


class SignalBus:
    """簡易發布／訂閱事件匯流排。"""

    def __init__(self) -> None:
        # event -> list of handlers（保持插入順序，list 去重由 subscribe 控制）
        self._handlers: dict[str, list[Handler]] = defaultdict(list)
        # once wrappers 的原始 handler 映射，用於退訂時查找包裝函式
        self._once_map: dict[str, dict[int, Handler]] = defaultdict(dict)

    # ── 訂閱 ──────────────────────────────────────────────────────────────

    def subscribe(self, event: str, handler: Handler) -> None:
        """訂閱事件。同一 handler 重複訂閱同一事件時忽略（冪等）。"""
        if handler not in self._handlers[event]:
            self._handlers[event].append(handler)

    def unsubscribe(self, event: str, handler: Handler) -> None:
        """退訂事件。若 handler 不在列表中則靜默忽略。"""
        handlers = self._handlers.get(event)
        if handlers and handler in handlers:
            handlers.remove(handler)
        # 若是 once-wrapped handler 也清掉映射
        once_map = self._once_map.get(event)
        if once_map:
            # 找到原始 handler 對應的 wrapper 並清除
            hid = id(handler)
            once_map.pop(hid, None)

    def once(self, event: str, handler: Handler) -> None:
        """訂閱事件，但只觸發一次，觸發後自動退訂。"""

        def _wrapper(*args: Any, **kwargs: Any) -> None:
            # 先退訂自己，再呼叫原始 handler（即使 handler 拋例外也已退訂）
            self._remove_once_wrapper(event, handler)
            handler(*args, **kwargs)

        # 記錄 wrapper，讓外部可以用原始 handler 來 unsubscribe
        self._once_map[event][id(handler)] = _wrapper
        if _wrapper not in self._handlers[event]:
            self._handlers[event].append(_wrapper)

    # ── 發布 ──────────────────────────────────────────────────────────────

    def publish(self, event: str, *args: Any, **kwargs: Any) -> None:
        """發布事件，依序呼叫所有訂閱者。

        - 不存在的事件 no-op。
        - 單一 handler 的例外被捕捉並記錄，不影響後續 handler。
        """
        handlers = self._handlers.get(event)
        if not handlers:
            return
        # 複製一份快照，避免 handler 在回呼中修改列表時產生迭代問題
        for h in list(handlers):
            try:
                h(*args, **kwargs)
            except Exception:
                logger.exception(
                    "SignalBus: handler %r raised an exception on event %r",
                    h,
                    event,
                )

    # ── 私有輔助 ──────────────────────────────────────────────────────────

    def _remove_once_wrapper(self, event: str, original_handler: Handler) -> None:
        """移除 once wrapper（在觸發前或觸發時呼叫）。"""
        once_map = self._once_map.get(event)
        if not once_map:
            return
        wrapper = once_map.pop(id(original_handler), None)
        if wrapper is not None:
            handlers = self._handlers.get(event)
            if handlers and wrapper in handlers:
                handlers.remove(wrapper)
