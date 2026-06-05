"""core.agents.opening — Opening Hook & Truth Seed Layer（UB6）。

序幕的任務不是「進入場景」，而是先讓玩家知道「這個故事有一個不對勁的核心」。
本層在開場放入 2–4 個**真假混合**的種子（True/False/Imagery/Personal/Mechanical）——
但**只把 surface（表層提示/義務）餵給 story；hidden_truth 留在 real_bible，永不餵 story**
（結構性防暴雷，C2/E2 一致）。story 依義務把種子寫成有鉤子的序幕，但不解釋完整真相。

設計來源：`nightmare-assault-design-fixed/12-mvp-b.md §一–九`（使用者回饋）。
"""
from __future__ import annotations

from typing import Any

# 開場長度政策（軟性，給 story 的提示；不做硬性截斷）
OPENING_LEN_MIN = 600
OPENING_LEN_MAX = 900

# 開場序幕 5 條義務（story 必須遵守）
OPENING_OBLIGATIONS = [
    "開場序幕不得只是地點描述：先建立一個角色動機鉤子（主角為何而來、在找誰）。",
    "加入一個真假混合的異常資訊，玩家一時無法判斷真假。",
    "加入一個與主角身份相關的恐怖鉤子。",
    "加入一個表層超自然想像，但不解釋成因（之後可在真相揭露時被回收）。",
    "最後才停在第一個可行動選擇；**禁止在開場直接解釋完整真相**，允許誤導但誤導須能被後續回收。",
]


def build_opening_seeds(blackboard: Any) -> list[dict]:
    """從 real_bible/protagonist 程式碼組裝序幕種子。

    每個 seed：{id, type, truth_ratio, surface(可餵 story), hidden_truth(留 real_bible), opening_obligation}。
    surface/opening_obligation 是「該寫什麼類型的鉤子」的指引（不含真相內容）；hidden_truth 是它背後的真相（不餵 story）。
    """
    snap = blackboard.snapshot() if hasattr(blackboard, "snapshot") else {}
    real = snap.get("real_bible") or {}
    world = real.get("world_truth") or {}
    pool = real.get("revelation_pool") or []
    proto = snap.get("protagonist") or {}
    who = proto.get("name") or "主角"
    situation = proto.get("starting_situation") or ""

    seeds: list[dict] = [
        {"id": "seed.personal", "type": "personal", "truth_ratio": 1.0,
         "surface": f"建立 {who} 來此的私人動機（與失蹤/牽掛的人直接相關）。{situation}".strip(),
         "hidden_truth": world.get("what_really_happened", ""),
         "opening_obligation": "用一個私人鉤子開場（主角在找誰、為何而來）。"},
        {"id": "seed.mechanical", "type": "mechanical", "truth_ratio": 0.5,
         "surface": "放一個暗示『這裡有某種不可違反的規則或危險』的警告，但不解釋原因。",
         "hidden_truth": world.get("deadly_rule", ""),
         "opening_obligation": "埋一個機制暗示（規則/危險），不說破。"},
        {"id": "seed.true", "type": "true", "truth_ratio": 0.6,
         "surface": "放一個真的但不完整的異常資訊，指向更深的核心真相，讓玩家想追下去。",
         "hidden_truth": world.get("the_threat_is", ""),
         "opening_obligation": "給一個真假難辨、指向主線的不完整線索。"},
        {"id": "seed.false", "type": "false", "truth_ratio": 0.35,
         "surface": "放一個會誤導玩家的表層解釋（例如以為只是普通失蹤或單純鬧鬼）。",
         "hidden_truth": "此表層解釋是錯的——真相另有其事，後續回收。",
         "opening_obligation": "放一個之後會被推翻的表層誤導。"},
        {"id": "seed.imagery", "type": "imagery", "truth_ratio": 0.2,
         "surface": "用表層想像營造畫面（牆像在呼吸／紅光像血管／腳印突然消失／廣播像夢裡的聲音），"
                    "讓玩家可能誤判為超自然，但不解釋成因。",
         "hidden_truth": world.get("the_threat_is", ""),
         "opening_obligation": "加一個會在真相揭露時被回收的表層超自然想像。"},
    ]
    # 標記實際有多少未揭露碎片（給 story「水面下還有東西」的張力，不露內容）
    for s in seeds:
        s["hidden_fragment_count"] = len(pool)
    return seeds


def surface_seeds(seeds: list[dict]) -> list[dict]:
    """只取可餵 story 的表層面（**剝除 hidden_truth**）。"""
    return [{"id": s["id"], "type": s["type"], "truth_ratio": s.get("truth_ratio"),
             "surface": s["surface"], "obligation": s["opening_obligation"]}
            for s in seeds]


def build_opening_context(blackboard: Any, base_ctx: dict) -> dict:
    """把開場義務 + 表層種子注入序幕 context（**絕不含 hidden_truth / real_bible**）。"""
    ctx = dict(base_ctx or {})
    seeds = build_opening_seeds(blackboard)
    ctx["opening_seeds"] = surface_seeds(seeds)                 # 只給表層
    ctx["opening_obligations"] = list(OPENING_OBLIGATIONS)
    ctx["opening_length_policy"] = {"min_chars": OPENING_LEN_MIN, "max_chars": OPENING_LEN_MAX}
    # 併進 narrative_obligations，讓 story 一定看到
    ob = list(ctx.get("narrative_obligations") or [])
    ctx["narrative_obligations"] = list(OPENING_OBLIGATIONS) + ob
    ctx["instruction"] = (
        (ctx.get("instruction", "") + " ") +
        f"這是序幕：依 opening_obligations 與 opening_seeds 寫一個有核心疑問的開場"
        f"（約 {OPENING_LEN_MIN}–{OPENING_LEN_MAX} 字），最後停在第一個選擇。不要解釋完整真相。"
    ).strip()
    return ctx
