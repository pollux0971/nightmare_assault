"""U06 OpenRouterClient 測試 — 全用 httpx.MockTransport，不打真網路。

真實 API smoke 測試在末尾，無 OPENROUTER_API_KEY 時自動 skip。
"""
import json
import os

import httpx
import pytest

from core.llm.client import OpenRouterClient
from core.persistence.db import Database


def _ok(text="hi", pt=3, ct=2):
    return httpx.Response(200, json={
        "choices": [{"message": {"content": text}, "delta": {"content": text}}],
        "usage": {"prompt_tokens": pt, "completion_tokens": ct},
    })


def _make(handler, db=None, agent_models=None):
    transport = httpx.MockTransport(handler)
    config = {
        "api_key": "test-key",
        "base_url": "https://example/api/v1",
        "agent_models": agent_models or {"story": ["m1", "m2"]},
        "timeout": 5,
    }
    return OpenRouterClient(config, db=db, transport=transport)


def test_success():
    c = _make(lambda req: _ok("hello"))
    r = c.call("story", "sys", "usr", 0.7)
    assert r.success and r.text == "hello" and r.model_used == "m1"
    assert r.input_tokens == 3 and r.output_tokens == 2


def test_fallback_on_500():
    def h(req):
        m = json.loads(req.content)["model"]
        return httpx.Response(500, text="boom") if m == "m1" else _ok("ok2")
    r = _make(h).call("story", "s", "u", 0.7)
    assert r.success and r.model_used == "m2"


def test_all_models_fail():
    r = _make(lambda req: httpx.Response(500, text="boom")).call("story", "s", "u", 0.7)
    assert not r.success and r.error


def test_timeout_triggers_fallback():
    seen = []

    def h(req):
        m = json.loads(req.content)["model"]
        seen.append(m)
        if m == "m1":
            raise httpx.TimeoutException("slow")
        return _ok("ok2")
    r = _make(h).call("story", "s", "u", 0.7)
    assert r.success and r.model_used == "m2" and "m1" in seen


def test_no_models_configured():
    r = _make(lambda req: _ok(), agent_models={}).call("ghost", "s", "u", 0.7)
    assert not r.success


def test_trace_written_to_db():
    db = Database()
    _make(lambda req: _ok(pt=5, ct=4), db=db).call("story", "s", "u", 0.7)
    summary = db.llm_cost_summary("__client__")
    assert summary["call_count"] >= 1
    assert summary["total_input_tokens"] == 5 and summary["total_output_tokens"] == 4


def test_stream_yields_tokens():
    sse = (b'data: {"choices":[{"delta":{"content":"He"}}]}\n\n'
           b'data: {"choices":[{"delta":{"content":"llo"}}]}\n\n'
           b'data: [DONE]\n\n')
    c = _make(lambda req: httpx.Response(200, content=sse))
    toks = list(c.stream("story", "s", "u", 0.7))
    assert "".join(toks) == "Hello"


# ── 真實 API smoke（預設 skip）──────────────────────────────────────────────
@pytest.mark.skipif(not os.environ.get("OPENROUTER_API_KEY"),
                    reason="無 OPENROUTER_API_KEY 環境變數")
def test_real_smoke():
    model = os.environ.get("OPENROUTER_SMOKE_MODEL", "openai/gpt-4o-mini")
    c = OpenRouterClient({
        "api_key": os.environ["OPENROUTER_API_KEY"],
        "base_url": "https://openrouter.ai/api/v1",
        "agent_models": {"smoke": [model]},
        "timeout": 30,
    })
    r = c.call("smoke", "You are a test.", "Reply with exactly: OK", 0.0)
    assert r.success, f"real call failed: {r.error}"
    assert r.text.strip()
