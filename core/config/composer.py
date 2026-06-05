"""core.config.composer — PromptComposer（P2，決定性 prompt 組裝）。

把 fragments 依 sort_order 組成單一 compiled prompt：只取 enabled、變數代入、產**穩定 prompt_hash**。
**純函式、零 LLM**（preview 不呼任何模型）。契約：dev/CONTRACTS.md §十一 `PromptComposer`；canonical = patch docs/01 §P2、docs/08。

變數語法（fragment content 內）：
- `{{ name }}`  → 必填變數；缺 → strict 時拋 PromptCompositionError、preview 時記 missing_required。
- `{{ name? }}` → 選填變數；缺 → 代入可見 placeholder `[[opt:name]]` 並記 warning。
靜態 fragment（無 placeholder）不受影響，仍決定性組裝。
"""
from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass, field
from typing import Any, Optional

# {{ name }} 或 {{ name? }}；name 允許 . _ 數字字母
_VAR_RE = re.compile(r"\{\{\s*([A-Za-z_][\w.]*)(\?)?\s*\}\}")

_SEP = "\n\n"   # fragment 之間的固定分隔（決定性）


class PromptCompositionError(ValueError):
    """缺必填變數（strict 模式）時拋出。"""


@dataclass
class CompiledPrompt:
    compiled_prompt: str
    prompt_hash: str
    enabled_fragments: list[str]
    model_settings: dict = field(default_factory=dict)
    context_policy: Optional[dict] = None
    warnings: list[str] = field(default_factory=list)
    missing_required: list[str] = field(default_factory=list)
    variables_used: list[str] = field(default_factory=list)
    profile_name: str = ""
    agent_name: str = ""

    @property
    def ok(self) -> bool:
        """preview/test 是否通過（無缺必填變數）。"""
        return not self.missing_required


def _substitute(content: str, variables: dict, *, missing_req: list, missing_opt: list,
                used: set) -> str:
    """把單一 fragment 內的 {{var}} / {{var?}} 代入；記錄缺漏。"""
    def repl(m: re.Match) -> str:
        name, optional = m.group(1), bool(m.group(2))
        if name in variables and variables[name] is not None:
            used.add(name)
            return str(variables[name])
        if optional:
            if name not in missing_opt:
                missing_opt.append(name)
            return f"[[opt:{name}]]"
        if name not in missing_req:
            missing_req.append(name)
        return f"[[MISSING:{name}]]"
    return _VAR_RE.sub(repl, content)


class PromptComposer:
    """無狀態組裝器：吃 ConfigStore + runtime 變數，吐 CompiledPrompt。不持有任何 LLM client。"""

    def __init__(self, store: Any) -> None:
        self.store = store

    def compose(self, agent_name: str, profile_name: Optional[str] = None,
                runtime_variables: Optional[dict] = None, *, strict: bool = True,
                overrides: Optional[dict] = None) -> CompiledPrompt:
        """組裝 compiled prompt。strict=True 缺必填變數即拋；strict=False（preview）回填 missing_required。

        overrides（P5 draft 預覽）：{fragment_key: content} 暫時替換該 fragment 內容（**不寫 DB**），
        讓 UI 在 activate 前預覽 draft 效果而不影響 active run。
        """
        profile = profile_name or self.store.active_profile()
        variables = dict(runtime_variables or {})
        overrides = overrides or {}
        frags = self.store.get_bound_fragments(agent_name, profile)   # 已 enabled + 依 sort_order 排序

        missing_req: list[str] = []
        missing_opt: list[str] = []
        used: set[str] = set()
        parts: list[str] = []
        keys: list[str] = []
        hash_units: list[str] = []
        for fr in frags:
            content = overrides.get(fr["fragment_key"], fr["content"])
            sub = _substitute(content, variables,
                              missing_req=missing_req, missing_opt=missing_opt, used=used)
            parts.append(sub)
            keys.append(fr["fragment_key"])
            hash_units.append(f"{fr['fragment_key']}@{fr.get('version', 1)}\n{sub}")

        compiled = _SEP.join(parts)

        warnings: list[str] = []
        for name in missing_opt:
            warnings.append(f"選填變數缺失，已用 placeholder：{name}")
        if missing_req:
            warnings.append(f"必填變數缺失：{', '.join(missing_req)}")

        if strict and missing_req:
            raise PromptCompositionError(
                f"compose({agent_name}/{profile}) 缺必填變數：{', '.join(missing_req)}")

        # ── 決定性 prompt_hash：對 (agent, profile, 排序後 fragment[key@version + 代入內容]) 雜湊 ──
        raw = f"{agent_name}|{profile}|" + "\n--\n".join(hash_units)
        prompt_hash = hashlib.sha256(raw.encode("utf-8")).hexdigest()[:16]

        # model_settings / context_policy（供 runtime 整合 P4 / 快照 P6）
        cfg = self.store.get_agent_config(agent_name, profile) or {}
        model_settings = {
            "model": cfg.get("model"),
            "temperature": cfg.get("temperature"),
            "max_output_tokens": cfg.get("max_output_tokens"),
            "context_budget_tokens": cfg.get("context_budget_tokens"),
            "stream_enabled": cfg.get("stream_enabled"),
            "output_schema_name": cfg.get("output_schema_name"),
        }
        context_policy = self.store.get_context_policy(agent_name, profile)

        return CompiledPrompt(
            compiled_prompt=compiled,
            prompt_hash=prompt_hash,
            enabled_fragments=keys,
            model_settings=model_settings,
            context_policy=context_policy,
            warnings=warnings,
            missing_required=missing_req,
            variables_used=sorted(used),
            profile_name=profile,
            agent_name=agent_name,
        )

    def preview(self, agent_name: str, profile_name: Optional[str] = None,
                runtime_variables: Optional[dict] = None,
                overrides: Optional[dict] = None) -> CompiledPrompt:
        """預覽：永不呼 LLM、永不拋（缺必填變數以 .missing_required / .ok 呈現，由呼叫端判失敗）。

        overrides：見 compose（P5 draft 預覽，不寫 DB）。
        """
        return self.compose(agent_name, profile_name, runtime_variables, strict=False,
                            overrides=overrides)
