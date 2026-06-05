"""core.narrative.sanitizer — 表層文字消毒（NR7，敘事控制 v0.2）。

story / npc-chat render 前掃非故事洩漏（technical/protocol/COLLECT/inst、prompt artifact、
壞 markdown fence、敘事內重複分隔符），破壞沉浸。先安全的決定性替換；保留契約授權的 in-world 詞。

只在後端跑（符合 D4：前端不解析）。零 LLM。對應 dev/CONTRACTS.md §十四（SurfaceTextSanitizer）。
"""
from __future__ import annotations

import re

# 預設不允許的非故事 token（ASCII；以「非字母邊界」比對，避免誤砍正常英文單字內部）
DEFAULT_BLOCKED_TOKENS = {
    "technical", "protocol", "COLLECT", "inst", "JSON", "json",
    "system", "assistant", "user", "prompt", "token", "schema", "core", "access",
}
# 中文技術/權限洩漏詞（HD2）：替換成中性詞
BLOCKED_CJK = {"權限": "記錄", "存取權": "記錄", "系統提示": "提示", "存取紀錄": "記錄"}
# 契約可授權的 in-world 詞（不消毒）；數字/頻率類預設保留
_DEFAULT_ALLOWED = {"432.7", "17Hz", "17hz"}
# 後端內部協定標記（絕不該出現在敘事表層）
_INTERNAL_MARKERS = ["<<<CONTINUE>>>", "<<<DECISION>>>", "```json", "```"]
# HD2：嚴重污染（路徑 / IP）——觸發 repair，不只刪詞
_PATH_RE = re.compile(r"/(?:usr|var|home|data|tmp|etc|opt|bin|root)/[^\s，。！？、）」』]+")
_IP_RE = re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b")
SEVERE_HITS = {"unix_path", "ip_address"}


class SurfaceTextSanitizer:
    def __init__(self, blocked_tokens: set[str] | None = None,
                 allowed_terms: set[str] | None = None):
        self.blocked_tokens = set(blocked_tokens or DEFAULT_BLOCKED_TOKENS)
        self.allowed_terms = set(_DEFAULT_ALLOWED) | set(allowed_terms or set())

    def _pattern(self, token: str) -> re.Pattern:
        # ASCII token：前後不可緊接英文字母（容許 CJK/標點邊界，如「聽見technical的」）
        return re.compile(rf"(?<![A-Za-z]){re.escape(token)}(?![A-Za-z])")

    def find_leaks(self, text: str) -> list[str]:
        text = text or ""
        leaks: list[str] = []
        for token in self.blocked_tokens:
            if token in self.allowed_terms:
                continue
            if self._pattern(token).search(text):
                leaks.append(token)
        for cjk in BLOCKED_CJK:
            if cjk in text:
                leaks.append(cjk)
        for mk in _INTERNAL_MARKERS:
            if mk in text:
                leaks.append(mk)
        if _PATH_RE.search(text):
            leaks.append("unix_path")
        if _IP_RE.search(text):
            leaks.append("ip_address")
        return leaks

    def scan(self, text: str) -> list[str]:
        """回傳命中類別（含 unix_path / ip_address；供 QualityGate 判斷是否嚴重污染）。"""
        return self.find_leaks(text)

    def sanitize(self, text: str) -> tuple[str, list[str]]:
        """回傳 (clean_text, leaks_found)。決定性移除洩漏 token / 路徑 / IP / 內部標記、收斂空白。"""
        text = text or ""
        leaks = self.find_leaks(text)
        clean = text
        # 嚴重污染（路徑 / IP）先決定性替換成中性詞
        clean = _PATH_RE.sub("某段技術紀錄", clean)
        clean = _IP_RE.sub("某組編號", clean)
        for token in self.blocked_tokens:
            if token in self.allowed_terms:
                continue
            clean = self._pattern(token).sub("", clean)
        for cjk, repl in BLOCKED_CJK.items():
            clean = clean.replace(cjk, repl)
        for mk in _INTERNAL_MARKERS:
            clean = clean.replace(mk, "")
        clean = re.sub(r"[ \t]{2,}", " ", clean)
        clean = re.sub(r"\n{3,}", "\n\n", clean)
        clean = re.sub(r"\s+([，。、！？」』])", r"\1", clean)
        return clean.strip(), leaks

    def sanitize_options(self, options: list) -> list[str]:
        """逐項消毒選項文字（HD2）。"""
        return [self.sanitize(o)[0] for o in (options or [])]

    def has_severe(self, leaks: list[str]) -> bool:
        """是否含嚴重污染（路徑 / IP）——觸發 repair，不只刪詞。"""
        return any(h in SEVERE_HITS for h in (leaks or []))

    def needs_repair(self, leaks: list[str]) -> bool:
        """嚴重污染（路徑 / IP）→ 需 repair；一般 token 決定性移除即可。"""
        return self.has_severe(leaks)


REPAIR_INSTRUCTION = (
    "移除文中的非故事 artifact 詞元（工具/系統/協定名、JSON 標記、prompt 殘留），"
    "不要改動劇情事實、選項、evidence 事件或任何 JSON metadata。")


def sanitize_text(text: str, allowed_terms: set[str] | None = None) -> str:
    """便捷函式：回傳消毒後文字（丟棄 leaks 清單）。"""
    return SurfaceTextSanitizer(allowed_terms=allowed_terms).sanitize(text)[0]


def allowed_from_contract(narrative_contract) -> set[str]:
    """從 NarrativeContract 取授權 in-world 詞（母題即視為授權，不被消毒）。"""
    allowed: set[str] = set()
    palette = getattr(narrative_contract, "motif_palette", None)
    if palette is not None:
        for grp in ("primary", "secondary"):
            allowed |= set(getattr(palette, grp, []) or [])
    return allowed
