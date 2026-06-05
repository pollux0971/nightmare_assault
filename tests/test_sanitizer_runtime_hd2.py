"""HD2 — Surface Sanitizer All Outputs 驗收測試（Runtime Hard-Gate v0.3.1）。

驗收：Unix path / IP / technical 等被 scan 命中並消毒；嚴重污染（path/IP）needs_repair=True；
sanitize_options 逐項消毒；中文權限詞替換；授權詞保留。
"""
from __future__ import annotations

from core.narrative.sanitizer import SurfaceTextSanitizer, sanitize_text


# ── reference 對齊：path / IP / technical ───────────────────────────────────
def test_removes_technical_path_ip():
    s = SurfaceTextSanitizer()
    text = "你看到 /usr/local/core/access.log 和一組 IP 192.168.1.1，technical 的呼吸聲傳來。"
    hits = s.scan(text)
    assert "unix_path" in hits and "ip_address" in hits
    out, leaks = s.sanitize(text)
    assert "/usr/local" not in out and "192.168" not in out and "technical" not in out
    assert s.has_severe(leaks) and s.needs_repair(leaks)   # 嚴重污染 → 需 repair


# ── 中文權限詞替換 ───────────────────────────────────────────────────────────
def test_cjk_permission_words():
    s = SurfaceTextSanitizer()
    out, leaks = s.sanitize("門上寫著需要存取權與系統提示。")
    assert "存取權" not in out and "系統提示" not in out
    assert "權限" not in out


# ── sanitize_options 逐項 ────────────────────────────────────────────────────
def test_sanitize_options():
    s = SurfaceTextSanitizer()
    opts = ["走向 /etc/passwd", "查看 protocol 面板", "離開"]
    out = s.sanitize_options(opts)
    assert "/etc/passwd" not in out[0]
    assert "protocol" not in out[1]
    assert out[2] == "離開"


# ── 授權 in-world 詞保留（432.7）；非嚴重污染不需 repair ─────────────────────
def test_authorized_terms_and_no_false_repair():
    s = SurfaceTextSanitizer()
    out, leaks = s.sanitize("儀器讀數停在 432.7。")
    assert "432.7" in out
    assert not s.needs_repair(leaks)
    # technical 是一般 token（非嚴重）→ 移除但不需 repair
    _, leaks2 = s.sanitize("聽見technical聲")
    assert not s.needs_repair(leaks2)


# ── 便捷函式 ─────────────────────────────────────────────────────────────────
def test_helper():
    assert "/var/log" not in sanitize_text("檔案在 /var/log/syslog 裡")


# ── loop 整合：選項也被消毒 ─────────────────────────────────────────────────
def test_loop_sanitizes_options(monkeypatch):
    import core.constants as C
    monkeypatch.setattr(C, "ENABLE_NARRATIVE_CONTROL", True)
    from core.orchestrator_loop import BeatLoop
    from core.models import DecisionPoint, Option
    loop = BeatLoop.__new__(BeatLoop)
    loop._narrative_contract = object()
    dp = DecisionPoint(situation_recap="走廊 /usr/bin 盡頭。", decision_type="action",
                       suggested_options=[Option(text="檢查 protocol 面板", tone="cautious")],
                       free_input_hint="", beat_meta={"beat_number": 1})
    out = loop._sanitize_decision_point(dp)
    assert "protocol" not in out.suggested_options[0].text
    assert "/usr/bin" not in out.situation_recap
