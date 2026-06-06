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

# ── 洩漏的內部識別碼 / 壞分隔符 / 黏在中文上的拉丁殘片（決定性移除）──────────────
# 壞掉或殘留的協定分隔符（如 `<<<CONTINAlUE>>>`——被模型寫壞，StreamParser 認不出而留在敘事）。
_DELIM_FRAGMENT_RE = re.compile(r"<<<[^<>]{0,24}>>>")
# snake_case 內部識別碼（含底線；如 crumpled_paper / origine_verification）——絕非正常散文。
_SNAKE_ID_RE = re.compile(r"(?<![A-Za-z0-9])[A-Za-z][A-Za-z0-9]*_[A-Za-z0-9_]+(?![A-Za-z0-9])")
# 點分小寫識別碼（如 scene.beyond / object.foo）——前後段各 ≥2/≥3 字母，避免誤砍 a.m./i.e./句末大寫。
_DOT_ID_RE = re.compile(
    r"(?<![A-Za-z0-9.])[a-z][a-z0-9]+\.[a-z][a-z0-9]{2,}(?:\.[a-z0-9]+)*(?![A-Za-z0-9])")
# 黏在中文字旁、無空白的小寫拉丁長詞（≥4；如「然homme」「app口袋」例外因 app<4）——多為亂碼殘片。
_CJK_GLUED_LATIN_RE = re.compile(r"(?<=[一-鿿])[a-z]{4,}|[a-z]{4,}(?=[一-鿿])")


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
        # 壞分隔符 / 內部識別碼 / 黏 CJK 拉丁殘片（placeholder 洩漏，補丁後）；
        # 識別碼/拉丁殘片若被契約授權為 in-world 詞（allowed_terms）則不算洩漏。
        if _DELIM_FRAGMENT_RE.search(text):
            leaks.append("delim_fragment")
        if self._has_disallowed(_SNAKE_ID_RE, text) or self._has_disallowed(_DOT_ID_RE, text):
            leaks.append("internal_identifier")
        if self._has_disallowed(_CJK_GLUED_LATIN_RE, text):
            leaks.append("glued_latin")
        return leaks

    def _has_disallowed(self, rx: re.Pattern, text: str) -> bool:
        return any(m.group(0) not in self.allowed_terms for m in rx.finditer(text))

    def _strip_keep_allowed(self, rx: re.Pattern, text: str) -> str:
        """移除 rx 命中，但保留 allowed_terms 授權的 in-world 詞。"""
        return rx.sub(lambda m: m.group(0) if m.group(0) in self.allowed_terms else "", text)

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
        # 壞分隔符碎片 → 先清（須在識別碼之前，避免 <<<x_y>>> 被半砍）
        clean = _DELIM_FRAGMENT_RE.sub("", clean)
        # 內部識別碼（snake_case / 點分小寫）+ 黏 CJK 拉丁殘片 → 移除（保留 allowed_terms 授權詞）
        clean = self._strip_keep_allowed(_SNAKE_ID_RE, clean)
        clean = self._strip_keep_allowed(_DOT_ID_RE, clean)
        clean = self._strip_keep_allowed(_CJK_GLUED_LATIN_RE, clean)
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
