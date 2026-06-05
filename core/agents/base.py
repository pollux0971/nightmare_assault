"""core.agents.base — SkillCaller 基類 + SKILL.md 載入器（U07）。

把 prompt 從程式碼抽出成 skills/{agent}/SKILL.md：載入→組 prompt（system=SKILL 內容、
user=結構化 context）→呼 client→用對應 Pydantic 類驗證輸出。SKILL.md 可熱重載（依 mtime）。
"""
from __future__ import annotations

import json
import re
from pathlib import Path


def extract_json(text: str) -> dict:
    """容錯 JSON 擷取：真實 LLM 常把 JSON 包在 ```json 圍欄或前後散文裡。

    去 markdown 圍欄、取第一個 { 到最後一個 }、修智慧引號與尾逗號，再 json.loads。
    找不到合法 JSON 則拋（由呼叫端決定重試）。供所有結構化 agent 輸出共用。
    """
    s = (text or "").strip()
    if s.startswith("```"):
        s = re.sub(r"^```[a-zA-Z]*\n?", "", s)
        s = re.sub(r"\n?```\s*$", "", s).strip()
    i, j = s.find("{"), s.rfind("}")
    if i != -1 and j != -1 and j > i:
        s = s[i:j + 1]
    s = s.replace("“", '"').replace("”", '"').replace("‘", "'").replace("’", "'")
    s = re.sub(r",\s*([}\]])", r"\1", s)
    return json.loads(s)


class SkillLoader:
    """載入 skills/{agent}/SKILL.md，mtime 變更時自動重載（熱重載，F3）。"""

    def __init__(self, skills_dir: str = "skills"):
        self._dir = Path(skills_dir)
        self._cache: dict[str, tuple[float, str]] = {}   # agent -> (mtime, content)

    def path_for(self, agent: str) -> Path:
        return self._dir / agent / "SKILL.md"

    def get(self, agent: str) -> str:
        p = self.path_for(agent)
        if not p.is_file():
            raise FileNotFoundError(f"找不到 SKILL.md: {p}")
        mtime = p.stat().st_mtime
        cached = self._cache.get(agent)
        if cached is None or cached[0] != mtime:
            content = p.read_text(encoding="utf-8")
            self._cache[agent] = (mtime, content)
            return content
        return cached[1]


class SkillCaller:
    """統一 agent 呼叫：載 SKILL.md → 組 prompt → 呼 client → Pydantic 驗證。"""

    def __init__(self, client, loader: SkillLoader, temperature_by_agent: dict | None = None):
        self.client = client
        self.loader = loader
        self.temperature_by_agent = temperature_by_agent or {}

    @staticmethod
    def _format_context(context: dict) -> str:
        """把結構化 context 序列化成 user 訊息（指令極短，靠結構化餵入）。"""
        return json.dumps(context, ensure_ascii=False, indent=2)

    def _temp(self, agent: str, override) -> float:
        if override is not None:
            return override
        return self.temperature_by_agent.get(agent, 0.7)

    def call(self, agent: str, context: dict, output_model=None, temperature=None,
             max_repair: int = 1, system_override: str | None = None):
        """非串流呼叫。output_model 給定則回驗證後的 Pydantic 物件，否則回原始文字。

        結構化輸出採 **retry-with-repair**（08 §三）：真實 LLM 偶爾違反 schema（如把列舉
        欄位填成非允許值），驗證失敗時把錯誤回饋給模型再試一次。仍失敗才拋。

        system_override（配置中心 P4）：非 None 時用它當 system prompt，否則讀 SKILL.md（static fallback）。
        """
        system = system_override if system_override is not None else self.loader.get(agent)
        user = self._format_context(context)
        temp = self._temp(agent, temperature)
        result = self.client.call(agent, system, user, temp)
        if not getattr(result, "success", False):
            raise RuntimeError(f"LLM 呼叫失敗（{agent}）：{getattr(result, 'error', None)}")
        if output_model is None:
            return result.text

        for attempt in range(max_repair + 1):
            try:
                return output_model.model_validate(extract_json(result.text))
            except Exception as e:
                if attempt >= max_repair:
                    raise
                repair_user = (
                    user + "\n\n[修正要求] 你上次的輸出無法解析或不符結構，錯誤："
                    + str(e)[:400]
                    + "\n請只回傳合法 JSON；所有列舉(Literal)欄位必須用允許值之一，不要加任何說明文字或 markdown 圍欄。"
                )
                result = self.client.call(agent, system, repair_user, temp)
                if not getattr(result, "success", False):
                    raise RuntimeError(
                        f"LLM repair 呼叫失敗（{agent}）：{getattr(result, 'error', None)}") from e

    def stream(self, agent: str, context: dict, temperature=None, system_override: str | None = None):
        """串流呼叫，回傳 token generator（供 story / npc-chat；解析交 StreamParser）。

        system_override（配置中心 P4）：非 None 時用它當 system prompt，否則讀 SKILL.md（static fallback）。
        """
        system = system_override if system_override is not None else self.loader.get(agent)
        user = self._format_context(context)
        return self.client.stream(agent, system, user, self._temp(agent, temperature))
