"""core.config.runtime — config-first Story prompt 來源 + static fallback（P4）。

來源優先序（ConfigPromptSource）：
  1. active config profile 的 story 編譯 prompt
  2. default config profile 的 story 編譯 prompt
  3. static built-in（skills/story/SKILL.md）

任一步失敗 / 缺必含行為 fragment / 空 prompt → 退下一級；全失敗 → static。**永不拋、永不 crash**。
契約：dev/CONTRACTS.md §十一 `ConfigPromptSource`；canonical = patch docs/08。
"""
from __future__ import annotations

import logging
from typing import Any, Optional

from core.agents.base import SkillLoader
from core.config.flags import config_center_enabled
from core.config.story_prompt import compose_story_prompt, validate_story_prompt

log = logging.getLogger("nightmare.config.runtime")


class ConfigPromptSource:
    """Story Agent 的 prompt 來源解析器（config-first，static fallback）。"""

    def __init__(self, store: Any, loader: Optional[SkillLoader] = None,
                 profile: Optional[str] = None, enabled: Optional[bool] = None) -> None:
        self.store = store
        self.loader = loader or SkillLoader()
        self.profile = profile
        self.enabled = config_center_enabled(store, profile) if enabled is None else enabled

    # ── static fallback ───────────────────────────────────────────────
    def _static(self, reason: str) -> tuple[str, dict]:
        try:
            text = self.loader.get("story")
        except Exception as e:                  # 連 SKILL.md 都讀不到（理論上不會）→ 空字串保命
            log.warning("static story SKILL load failed: %s", e)
            text = ""
        return text, {"source": "static", "reason": reason, "profile": None, "prompt_hash": None,
                      "enabled_fragments": []}

    # ── 解析 ──────────────────────────────────────────────────────────
    def story_system_prompt(self, runtime_variables: Optional[dict] = None) -> tuple[str, dict]:
        """回傳 (system_prompt, meta)。meta.source ∈ {config, static}。"""
        if not self.enabled:
            return self._static("flag_off")

        candidates = [("active", self.store.active_profile()),
                      ("default", self.store.default_profile())]
        seen: set[str] = set()
        for _label, prof in candidates:
            if prof in seen:
                continue
            seen.add(prof)
            try:
                compiled = compose_story_prompt(self.store, runtime_variables or {},
                                                profile=prof, strict=False)
                missing = validate_story_prompt(compiled)
                if missing or not compiled.compiled_prompt.strip():
                    log.warning("config story prompt incomplete (profile=%s, missing=%s); next source",
                                prof, missing)
                    continue
                log.info("story prompt source=config profile=%s hash=%s frags=%d",
                         prof, compiled.prompt_hash, len(compiled.enabled_fragments))
                return compiled.compiled_prompt, {
                    "source": "config", "profile": prof, "prompt_hash": compiled.prompt_hash,
                    "enabled_fragments": compiled.enabled_fragments, "warnings": compiled.warnings,
                    "model_settings": compiled.model_settings,
                }
            except Exception as e:
                log.warning("config story prompt failed (profile=%s): %s; next source", prof, e)
                continue
        return self._static("config_failed")
