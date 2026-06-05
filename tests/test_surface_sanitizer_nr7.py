"""NR7 — SurfaceTextSanitizer 驗收測試（敘事控制 v0.2）。

驗收：CJK 中夾雜的洩漏 token（technical/protocol/COLLECT/inst）被移除；授權詞（432.7/17Hz）保留；
正常英文單字內部不被誤砍；內部協定標記/壞 fence 移除；契約授權母題不消毒。
"""
from __future__ import annotations

from core.narrative.sanitizer import (
    SurfaceTextSanitizer, sanitize_text, allowed_from_contract, DEFAULT_BLOCKED_TOKENS,
)
from core.narrative.models import NarrativeContract, ProtagonistMotive, MotifPalette


# ── CJK 夾雜洩漏 → 移除（reference 對齊）─────────────────────────────────────
def test_removes_cjk_mixed_leakage():
    s = SurfaceTextSanitizer()
    text = "辦公室裡靜得能聽見technical的呼吸聲。"
    assert "technical" in s.find_leaks(text)
    clean, leaks = s.sanitize(text)
    assert "technical" not in clean
    assert "technical" in leaks
    assert "辦公室裡靜得能聽見的呼吸聲。" == clean


# ── 多種洩漏 token + 內部標記 ───────────────────────────────────────────────
def test_removes_various_tokens_and_markers():
    s = SurfaceTextSanitizer()
    clean, leaks = s.sanitize("他低聲說protocol，然後COLLECT。```json殘留<<<DECISION>>>")
    assert all(t not in clean for t in ("protocol", "COLLECT", "```json", "<<<DECISION>>>"))
    assert "protocol" in leaks and "COLLECT" in leaks


# ── 授權 in-world 詞保留（432.7 / 17Hz）─────────────────────────────────────
def test_keeps_authorized_terms():
    s = SurfaceTextSanitizer()
    text = "儀器掉到 432.7 以下，蜂鳴維持在 17Hz。"
    clean, leaks = s.sanitize(text)
    assert "432.7" in clean and "17Hz" in clean
    assert leaks == []


# ── 不誤砍正常英文單字內部（inst → institution 不受影響）────────────────────
def test_word_boundary_no_false_positive():
    s = SurfaceTextSanitizer()
    # 'inst' 是 blocked，但 institution 內部不該被砍
    assert "inst" in DEFAULT_BLOCKED_TOKENS
    clean, leaks = s.sanitize("The institution stood there.")
    assert "institution" in clean
    assert "inst" not in leaks


# ── 契約授權母題不被消毒 ─────────────────────────────────────────────────────
def test_contract_allowed_motifs_preserved():
    nc = NarrativeContract(
        core_premise="x",
        protagonist_motive=ProtagonistMotive("a", "b", "c", "d"),
        central_question="?",
        motif_palette=MotifPalette(primary=["protocol"]))   # 此局 protocol 是 in-world 詞
    allowed = allowed_from_contract(nc)
    assert "protocol" in allowed
    out = sanitize_text("這是他們的protocol儀式。", allowed_terms=allowed)
    assert "protocol" in out                       # 被授權 → 保留


# ── 便捷函式 ─────────────────────────────────────────────────────────────────
def test_sanitize_text_helper():
    assert "technical" not in sanitize_text("聽見technical聲")
