"""core.narrative.action_intent — ActionIntent 分類（WorldConsequence vs TruthEvidence Split）。

開放式探索的核心區分：玩家「找路 / 整理 / 引用 NPC fact / 一般檢查」是**世界後果**，
不該推進 truth reveal；只有**明確的真相調查**（研究紀錄、解讀異常數據…）或合法 structured
evidence_events 才推 reveal。本模組把玩家輸入分成七類，供 TruthEvidenceGate 判斷。

零 LLM、規則版。對應 15-player-sovereignty.md（WorldConsequence vs TruthEvidence）。
"""
from __future__ import annotations

WORLD_NAVIGATION = "world_navigation"        # 找路 / 移動 / 往某方向
WORLD_REVIEW = "world_review"                # 整理 / 盤點 / 回顧已知
OBJECT_INSPECTION = "object_inspection"      # 檢查某物件（只改物件狀態）
TRUTH_INVESTIGATION = "truth_investigation"  # 研究紀錄 / 解讀數據 / 追查真相（唯一推 reveal）
NPC_FACT_QUERY = "npc_fact_query"            # 引用 NPC 說的事去做某事（只寫 WorldModel fact）
CAMPAIGN_END = "campaign_end"                # 結束本次調查
UNKNOWN = "unknown"

# ── 判定詞（白名單；truth_investigation 需「明確研究真相」的動作）────────────────
# 真相調查：研究/解讀/分析「紀錄、數據、頻率、檔案、屍體…」或明確追查真相
_TRUTH = ("研究", "分析", "解讀", "破解", "推敲", "研判", "驗屍", "解剖", "比對",
          "實驗紀錄", "實驗記錄", "研究紀錄", "研究記錄", "異常頻率", "異常數據",
          "數據", "檔案", "日誌", "錄音", "病歷", "監控紀錄", "追查真相", "查清真相",
          "弄清真相", "查明真相", "挖出真相", "真相是什麼", "到底發生什麼", "深入研究",
          "解開謎", "破解謎", "搞清楚發生", "查清楚到底")
# 引用 NPC 說的事（去找路 / 去找東西）——只寫 fact，不推 reveal
_NPC_FACT = ("根據他說", "根據她說", "根據他講", "根據她講", "根據npc", "他說的", "她說的",
             "他提到", "她提到", "按照他說", "依他說", "依她說", "據他說", "據她說",
             "循著他說", "照他說", "他講的線索", "她講的線索", "他剛說", "她剛說",
             "根據剛才他", "根據剛才她", "他說的線索", "她說的線索")
# 整理 / 盤點 / 回顧（review）
_REVIEW = ("整理線索", "整理筆記", "盤點", "回顧", "梳理", "整理已知", "整理資訊",
           "整理思緒", "重新理", "整理一下", "不新增調查")
# 找路 / 移動 / 方向
_NAV = ("移動", "前往", "走向", "找路", "找出口", "方向", "繞到", "繞過去", "摸向",
        "朝", "往", "走到", "走去", "穿過", "沿著", "出發", "去找路", "去那邊",
        "回到", "進入", "返回", "下樓", "上樓", "走廊", "通道")
# 一般檢查（只改物件狀態）
_INSPECT = ("檢查", "查看", "檢視", "端詳", "翻看", "看看", "掀開", "翻找", "摸一摸")

# 明確「不碰真相」的措辭 → 本 beat 一律 block reveal
_NO_TRUTH = ("不碰真相", "不追真相", "不查真相", "不想知道真相", "不深入", "不再深入",
             "不新增調查", "不再調查", "暫不調查", "先不調查", "只找路", "只移動",
             "只整理", "只是找路", "純粹找路", "只想找路", "只要找路", "別碰真相")


def no_truth_intent(text: str) -> bool:
    """玩家明確表示「不碰真相 / 不新增調查 / 只找路 / 只整理」→ 本 beat block reveal。"""
    return any(p in (text or "") for p in _NO_TRUTH)


def classify_action(text: str, *, exploration_mode: str = "active_exploration",
                    exit_affordance=None) -> str:
    """把玩家輸入分成七類 ActionIntent。

    優先序：campaign_end > review（含 review_mode）> truth_investigation（且非 no-truth）
    > npc_fact_query > world_navigation > object_inspection > unknown。
    """
    t = text or ""
    from core.narrative.exit_resolver import END_CAMPAIGN
    from core.narrative.exploration_mode import is_review_locked

    if exit_affordance is not None and getattr(exit_affordance, "affordance", None) == END_CAMPAIGN:
        return CAMPAIGN_END
    # review_mode 或明確整理措辭 → world_review（即使含其他詞，整理優先）
    if is_review_locked(exploration_mode) or any(p in t for p in _REVIEW):
        return WORLD_REVIEW
    # 明確「不碰真相」→ 不可判 truth_investigation（落到 nav/inspect/unknown，gate 也會擋）
    nt = no_truth_intent(t)
    if not nt and any(p in t for p in _TRUTH):
        return TRUTH_INVESTIGATION
    # 引用 NPC 說的事（在 nav 之前——「根據他說的線索找路」應記為 npc_fact_query）
    if any(p in t for p in _NPC_FACT):
        return NPC_FACT_QUERY
    if any(p in t for p in _NAV):
        return WORLD_NAVIGATION
    if any(p in t for p in _INSPECT):
        return OBJECT_INSPECTION
    return UNKNOWN
