"""core.llm.client — OpenRouter HTTP client，含 fallback 鏈與 llm_traces 記錄（U06）。

只從環境變數取 API key；不在任何地方硬編金鑰。
測試時透過 transport= 注入 httpx.MockTransport，不打真實網路。
"""
from __future__ import annotations

import hashlib
import json
import time
from typing import Generator, Optional

import httpx

from core.models import LLMResult


# ---------------------------------------------------------------------------
# 內部工具
# ---------------------------------------------------------------------------

def _prompt_hash(system: str, user: str) -> str:
    """回傳 system+user 合併後的 SHA-256 前 16 字。"""
    raw = f"{system}\n{user}".encode("utf-8")
    return hashlib.sha256(raw).hexdigest()[:16]


# ---------------------------------------------------------------------------
# OpenRouterClient
# ---------------------------------------------------------------------------

class OpenRouterClient:
    """同步 OpenRouter（OpenAI 相容）client，含 fallback 鏈與 trace 寫入。

    Args:
        config: 包含以下鍵的 dict：
            - api_key (str): OpenRouter API key
            - base_url (str): e.g. "https://openrouter.ai/api/v1"
            - agent_models (dict[str, list[str]]): {agent_name: [primary, fallback1, ...]}
            - timeout (float | int): 秒數，預設 60
        db: 可選 Database 實例；有則每次嘗試後寫一筆 llm_traces。
        transport: 可選 httpx.BaseTransport；傳入 MockTransport 供測試。
    """

    def __init__(
        self,
        config: dict,
        db=None,
        transport: Optional[httpx.BaseTransport] = None,
    ) -> None:
        self._api_key: str = config["api_key"]
        self._base_url: str = config["base_url"].rstrip("/")
        self._agent_models: dict[str, list[str]] = config.get("agent_models", {})
        self._timeout: float = float(config.get("timeout", 60))
        # max_tokens：預設 None＝不送（用模型上限／帳戶額度決定，可到 1M）。
        # 只有 config 明確設值時才送（想省成本/對齊額度上限時用）；它只是「可負擔上限」，仍按真實用量計費。
        mt = config.get("max_tokens")
        self._max_tokens = int(mt) if mt else None
        self._db = db

        # 建立共用 httpx.Client（transport 可注入，測試用）
        client_kwargs: dict = {
            "timeout": self._timeout,
            "headers": {
                "Authorization": f"Bearer {self._api_key}",
                "Content-Type": "application/json",
            },
        }
        if transport is not None:
            client_kwargs["transport"] = transport

        self._client = httpx.Client(**client_kwargs)

    def _body(self, model: str, system: str, user: str, temperature: float, stream: bool) -> dict:
        """組 chat/completions 請求 body；max_tokens 僅在 config 有設時才送。"""
        body = {
            "model": model,
            "messages": [
                {"role": "system", "content": system},
                {"role": "user", "content": user},
            ],
            "temperature": temperature,
            "stream": stream,
        }
        if self._max_tokens:
            body["max_tokens"] = self._max_tokens
        return body

    # ------------------------------------------------------------------
    # 公開 API
    # ------------------------------------------------------------------

    def call(
        self,
        agent: str,
        system: str,
        user: str,
        temperature: float,
        stream: bool = False,
    ) -> LLMResult:
        """同步呼叫，依 agent_models[agent] 清單依序嘗試（fallback 鏈）。

        每次嘗試都寫一筆 llm_traces（若 db 存在）。
        全部失敗則回 LLMResult(success=False, error=...)。
        """
        models = self._agent_models.get(agent, [])
        if not models:
            err = f"agent '{agent}' 無設定模型清單（agent_models）"
            self._write_trace(agent=agent, model=None, input_tokens=0,
                              output_tokens=0, latency_ms=0, success=False,
                              error=err, prompt_hash=_prompt_hash(system, user))
            return LLMResult(
                text="", model_used="", input_tokens=0, output_tokens=0,
                latency_ms=0, success=False, error=err,
            )

        last_error: str = ""
        p_hash = _prompt_hash(system, user)

        for model in models:
            t0 = time.monotonic()
            try:
                resp = self._client.post(
                    f"{self._base_url}/chat/completions",
                    json=self._body(model, system, user, temperature, stream),
                )
                latency_ms = int((time.monotonic() - t0) * 1000)

                if resp.status_code != 200:
                    last_error = (
                        f"HTTP {resp.status_code}: {resp.text[:200]}"
                    )
                    self._write_trace(
                        agent=agent, model=model, input_tokens=0,
                        output_tokens=0, latency_ms=latency_ms,
                        success=False, error=last_error, prompt_hash=p_hash,
                    )
                    continue  # fallback

                data = resp.json()
                text = data["choices"][0]["message"]["content"]
                usage = data.get("usage", {})
                input_tokens = usage.get("prompt_tokens", 0)
                output_tokens = usage.get("completion_tokens", 0)

                self._write_trace(
                    agent=agent, model=model, input_tokens=input_tokens,
                    output_tokens=output_tokens, latency_ms=latency_ms,
                    success=True, error=None, prompt_hash=p_hash,
                )
                return LLMResult(
                    text=text,
                    model_used=model,
                    input_tokens=input_tokens,
                    output_tokens=output_tokens,
                    latency_ms=latency_ms,
                    success=True,
                    error=None,
                )

            except httpx.TimeoutException as exc:
                latency_ms = int((time.monotonic() - t0) * 1000)
                last_error = f"Timeout: {exc}"
                self._write_trace(
                    agent=agent, model=model, input_tokens=0,
                    output_tokens=0, latency_ms=latency_ms,
                    success=False, error=last_error, prompt_hash=p_hash,
                )
                continue  # fallback

            except httpx.HTTPError as exc:
                latency_ms = int((time.monotonic() - t0) * 1000)
                last_error = f"HTTPError: {exc}"
                self._write_trace(
                    agent=agent, model=model, input_tokens=0,
                    output_tokens=0, latency_ms=latency_ms,
                    success=False, error=last_error, prompt_hash=p_hash,
                )
                continue  # fallback

        # 全部模型嘗試失敗
        return LLMResult(
            text="", model_used="", input_tokens=0, output_tokens=0,
            latency_ms=0, success=False,
            error=last_error or "所有模型均失敗",
        )

    def stream(
        self,
        agent: str,
        system: str,
        user: str,
        temperature: float,
    ) -> Generator[str, None, None]:
        """串流呼叫，yield token 字串。

        依 agent_models[agent] 清單依序嘗試（fallback 鏈）。
        每次嘗試都寫一筆 llm_traces（若 db 存在）。
        全部失敗則 raise RuntimeError。
        """
        models = self._agent_models.get(agent, [])
        if not models:
            raise RuntimeError(f"agent '{agent}' 無設定模型清單（agent_models）")

        p_hash = _prompt_hash(system, user)
        last_error: str = ""

        for model in models:
            t0 = time.monotonic()
            try:
                with self._client.stream(
                    "POST",
                    f"{self._base_url}/chat/completions",
                    json=self._body(model, system, user, temperature, stream=True),
                ) as resp:
                    if resp.status_code != 200:
                        latency_ms = int((time.monotonic() - t0) * 1000)
                        last_error = f"HTTP {resp.status_code}"
                        self._write_trace(
                            agent=agent, model=model, input_tokens=0,
                            output_tokens=0, latency_ms=latency_ms,
                            success=False, error=last_error, prompt_hash=p_hash,
                        )
                        continue  # fallback

                    yielded_any = False
                    for raw_line in resp.iter_lines():
                        line = raw_line.strip()
                        if not line:
                            continue
                        if line.startswith("data:"):
                            line = line[len("data:"):].strip()
                        if line == "[DONE]":
                            break
                        try:
                            chunk = json.loads(line)
                            delta = (
                                chunk.get("choices", [{}])[0]
                                .get("delta", {})
                                .get("content", "")
                            )
                            if delta:
                                yielded_any = True
                                yield delta
                        except (json.JSONDecodeError, IndexError, KeyError):
                            continue

                    latency_ms = int((time.monotonic() - t0) * 1000)
                    self._write_trace(
                        agent=agent, model=model, input_tokens=0,
                        output_tokens=0, latency_ms=latency_ms,
                        success=True, error=None, prompt_hash=p_hash,
                    )
                    return  # 成功串流完畢，退出 fallback 迴圈

            except httpx.TimeoutException as exc:
                latency_ms = int((time.monotonic() - t0) * 1000)
                last_error = f"Timeout: {exc}"
                self._write_trace(
                    agent=agent, model=model, input_tokens=0,
                    output_tokens=0, latency_ms=latency_ms,
                    success=False, error=last_error, prompt_hash=p_hash,
                )
                continue  # fallback

            except httpx.HTTPError as exc:
                latency_ms = int((time.monotonic() - t0) * 1000)
                last_error = f"HTTPError: {exc}"
                self._write_trace(
                    agent=agent, model=model, input_tokens=0,
                    output_tokens=0, latency_ms=latency_ms,
                    success=False, error=last_error, prompt_hash=p_hash,
                )
                continue  # fallback

        raise RuntimeError(last_error or "stream: 所有模型均失敗")

    # ------------------------------------------------------------------
    # 內部輔助
    # ------------------------------------------------------------------

    def _write_trace(
        self,
        agent: str,
        model: Optional[str],
        input_tokens: int,
        output_tokens: int,
        latency_ms: int,
        success: bool,
        error: Optional[str],
        prompt_hash: str,
    ) -> None:
        """若有 db，寫一筆 llm_traces；無 db 時靜默略過。"""
        if self._db is None:
            return
        self._db.write_llm_trace(
            run_id="__client__",
            beat_number=None,
            agent=agent,
            model=model,
            prompt_hash=prompt_hash,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            latency_ms=latency_ms,
            success=success,
            error=error,
        )
