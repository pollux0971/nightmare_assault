"""core.narrative.npc_prose_facts — NPC 散文 fact 抽取（fallback，範圍極小）。

真 LLM 的 NPC 常常**不吐** structured `entity_delta`，只給散文回覆。本抽取器是退路：
把 NPC `visible_reply` 裡**明確、可用、非真相**的資訊（位置/鎖死出口/需先做某事）抽成
WorldModel **fact** entity——保留自然語意 label（「通訊設備在B2機房」，不是 `machine_room_known`）。

嚴格邊界（白名單式，只抽三類明確主張）：
  - 只輸出 kind=fact（不碰 object/area/exit；經 coerce_npc_entity_deltas 再保險）。
  - 每次最多 1–2 個；malformed/ambiguous 一律放棄。
  - 不抽氣氛/比喻/幻覺；不推 reveal；不改 current_area（後兩者由 bridge 路徑保證）。
  - 只在 structured entity_delta 缺席/為空時，由 loop 啟用（見 _bridge_npc_entity_delta）。

零 LLM、規則版。對應 15-player-sovereignty.md（NPC 散文 fact fallback）。
"""
from __future__ import annotations

import re

# 分句：中英標點 + 換行（逗號也切，方便隔離單一主張）
_SPLIT = re.compile(r"[。．.！!？?；;，,、\n]+")
# 去掉開頭的連接詞/填充語（保留主張本體）
_LEAD = re.compile(r"^(但|不過|而|至於|然後|另外|其實|我想|我覺得|聽說|據說|可能|也許|大概|應該|反正|總之|你知道嗎?|對了)+")

# 氣氛/比喻/幻覺 → 一律不抽（requirement 11；安全網）
_ATMOSPHERE = ("像", "彷彿", "似乎", "宛如", "好像", "幽靈", "鬼", "低語", "呢喃", "陰影",
               "影子", "扭曲", "尖叫", "夢魘", "腥", "霧", "蠕動", "爬行", "嗡鳴", "雪花",
               "噪點", "顫抖", "滲血", "血", "詭異", "不祥", "寒意")

# ── 三類「明確可用主張」的判定詞 ─────────────────────────────────────────────
_LOC_VERB = ("在", "位於", "放在", "擺在", "設在", "就在", "收在", "藏在")
_LOC_PLACE = ("機房", "控制室", "配電室", "發電機房", "醫療室", "實驗室", "倉庫", "機艙",
              "檔案室", "儲藏室", "值班室", "地下室", "b1", "b2", "b3", "b4")
_LOC_THING = ("設備", "發電機", "通訊", "電源", "控制台", "無線電", "對講機", "鑰匙", "鑰卡",
              "紀錄", "記錄", "文件", "開關", "閥", "主機", "伺服器", "儀器", "工具", "藥",
              "醫療", "電池", "保險絲", "面板", "終端", "卡片", "鑰")

_EXIT_NOUN = ("門", "閘門", "出口", "梯", "通道", "電梯", "樓梯", "閘口", "艙門", "防火門", "防火梯")
_EXIT_LOCKED = ("鎖死", "鎖住", "鎖了", "封死", "封住", "封閉", "堵死", "堵住", "打不開",
                "出不去", "過不去", "不能走", "無法通行", "走不通", "卡死", "卡住", "故障")

_ACT_MODAL = ("要先", "必須先", "得先", "需要先", "得要", "一定要", "你得", "你要", "務必", "只能先")
_ACT_VERB = ("重啟", "啟動", "打開", "接通", "修復", "修好", "輸入", "關閉", "切斷", "找到",
             "拿到", "供電", "送電", "解鎖", "開啟", "恢復", "繞過", "重新啟動")

LOCATION_CLAIM = "location_claim"
LOCKED_EXIT_CLAIM = "locked_exit_claim"
ACTION_REQUIRED_CLAIM = "action_required_claim"


def _classify(clause_lc: str) -> str | None:
    """回傳該子句的粗分類 tag，或 None（不是明確可用主張）。"""
    if any(a in clause_lc for a in _ATMOSPHERE):
        return None                                    # 氣氛/比喻/幻覺 → 放棄
    # location_claim：thing + 定位動詞 + 地點
    if (any(v in clause_lc for v in _LOC_VERB)
            and any(p in clause_lc for p in _LOC_PLACE)
            and any(t in clause_lc for t in _LOC_THING)):
        return LOCATION_CLAIM
    # locked_exit_claim：出口名詞 + 鎖死/封死（是「主張」，不登記 exit entity）
    if (any(n in clause_lc for n in _EXIT_NOUN)
            and any(s in clause_lc for s in _EXIT_LOCKED)):
        return LOCKED_EXIT_CLAIM
    # action_required_claim：必須先 + 動作
    if (any(m in clause_lc for m in _ACT_MODAL)
            and any(v in clause_lc for v in _ACT_VERB)):
        return ACTION_REQUIRED_CLAIM
    return None


def _clean_label(clause: str) -> str:
    s = _LEAD.sub("", (clause or "").strip()).strip("　 \t\"'「」『』（）()")
    return s


def _label_ok(label: str) -> bool:
    # 太短（無資訊）或太長（敘事/含糊）→ 放棄；含問號（提問非主張）→ 放棄
    if not (3 <= len(label) <= 40):
        return False
    if "？" in label or "?" in label:
        return False
    return True


def extract_npc_prose_facts(reply: str, *, npc_id: str, cap: int = 2) -> list:
    """從 NPC 散文回覆抽 ≤cap 個 fact（WorldDelta，已帶 source/confidence/origin）。

    僅三類明確主張；malformed/ambiguous/氣氛一律放棄。回 coerce 過的 WorldDelta 清單。
    """
    from core.world.model import coerce_npc_entity_deltas, FACT
    t = reply or ""
    if not t.strip():
        return []
    candidates: list = []
    seen: set = set()
    for raw_clause in _SPLIT.split(t):
        clause = raw_clause.strip()
        if not clause or len(clause) > 40:             # 過長 = 敘事/含糊 → 跳過
            continue
        tag = _classify(clause.lower())
        if tag is None:
            continue
        label = _clean_label(clause)
        if not _label_ok(label):
            continue
        key = label.replace(" ", "")
        if key in seen:
            continue
        seen.add(key)
        candidates.append({"op": "register", "kind": FACT, "label": label,
                           "props": {"tags": [tag]}})
        if len(candidates) >= cap:
            break
    # 經 NPC coerce：注入 source=npc_id / confidence=npc_claim、origin=npc，並再夾 cap
    return coerce_npc_entity_deltas(candidates, npc_id=npc_id, cap=cap)
