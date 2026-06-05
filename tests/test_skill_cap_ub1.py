"""UB1 — 技能宣稱封頂強化 驗收測試。

驗收：誇張宣稱被封頂、侷限具體接劇情（變謎題/線索）、ledger 記錄、不誤封正常行動。
"""
from __future__ import annotations

from core.agents.warden import check_skill_claim, run_warden
from core.blackboard import Blackboard


# ── 本地封頂偵測 ─────────────────────────────────────────────────────────────
def test_supernatural_claim_capped_with_limitation():
    v = check_skill_claim("我使用瞬間移動穿到走廊另一頭")
    assert v is not None
    assert v.skill_verdict == "allow"          # 接受但…
    assert v.skill_limitation and len(v.skill_limitation) > 10   # …有具體侷限
    assert v.rule_violation is False           # 不是死亡
    assert "侷限" in v.directive_to_story        # 指示 story 把侷限寫成阻礙/線索


def test_weapon_claim_capped():
    v = check_skill_claim("我掏出槍對它爆頭")
    assert v is not None and "槍" in v.skill_limitation + v.skill_claim


def test_invincible_claim_capped():
    v = check_skill_claim("我是無敵的，刀槍不入")
    assert v is not None and v.skill_limitation


def test_normal_action_not_capped():
    """正常行動不該被誤判為破格宣稱（不誤殺/不誤封）。"""
    assert check_skill_claim("我推開那扇門") is None
    assert check_skill_claim("我檢查牆上的血跡") is None
    assert check_skill_claim("我問護士這裡發生什麼事") is None


# ── 與 run_warden 整合：本地優先（LLM 掛也能封頂）──────────────────────────
def test_run_warden_caps_locally_without_llm():
    bb = Blackboard()
    bb.write("setup", "real_bible", {"deadly_rule": "不可喝藥水", "hard_triggers": []})
    v = run_warden("我用魔法把整棟樓夷平", bb, caller=None)   # 無 caller（LLM 不可用）
    assert v.skill_claim and v.skill_limitation              # 仍封頂
    assert v.rule_violation is False


def test_lethal_rule_takes_priority_over_skill():
    """致命硬規則優先於技能封頂（B9 順序）。"""
    bb = Blackboard()
    bb.write("setup", "real_bible", {"deadly_rule": "不可喝藥水", "hard_triggers": ["喝藥水"]})
    v = run_warden("我有超能力，然後喝藥水", bb, caller=None)
    assert v.rule_violation is True and v.ending_triggered == "death_physical"


# ── ledger 記錄（透過 loop 的 _record_skill_claim）────────────────────────────
def test_skill_claim_written_to_ledger():
    from core.orchestrator_loop import BeatLoop
    from core.persistence.db import Database
    loop = BeatLoop(caller=None, blackboard=Blackboard(), db=Database(), run_id="t", use_kernel=False)
    v = check_skill_claim("我隔空念力捏碎它的頭")
    loop._record_skill_claim(v)
    led = loop.bb.ledger
    assert any(e.get("type") == "skill_limit" and "侷限" in e.get("content", "") for e in led)
