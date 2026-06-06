#!/usr/bin/env python3
"""opening_variation_selfplay — 開場變體重複率量測（真 LLM）。

用途：
  P0 baseline：flag OFF，跑 N 場「只取開場 narrative」，量 紙條 / 林晨 / 找人(missing_person)
              的出現率與 message medium 分布。
  after：flag ON，OpeningVariationContract 生效後再量一次，對照下降幅度。

跑法：
  python3 dev/tools/opening_variation_selfplay.py --runs 6                 # baseline（flag OFF）
  python3 dev/tools/opening_variation_selfplay.py --runs 6 --contract      # after（flag ON）

只跑開場（loop.start），不跑後續 beat，省 token。
"""
from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

# 量測關鍵詞（baseline 重複症狀）——繁簡都收，避免簡體輸出被漏算。
PAPER_LITERALS = ("紙條", "纸条", "便條", "便条", "手寫留言", "手写留言", "字條", "字条", "便籤", "便签")
NAME_LITERAL = "林晨"
MISSING_PERSON_CUES = ("失蹤", "失踪", "找人", "尋找", "寻找", "尋人", "寻人",
                       "下落不明", "失聯", "失联", "不見了", "不见了")
HANDNOTE_CUES = PAPER_LITERALS

# 各主題輪換，避免主題本身錨定（baseline 也用同一組以公平對照）
THEMES = [
    ("午夜的廢棄海事研究站", "周凱"),
    ("斷電的地下醫院檔案層", "蘇明"),
    ("暴雨夜的山區氣象站", "李航"),
    ("停運的舊地鐵維修段", "陳逸"),
    ("孤島上的無人燈塔", "林default不要用"),  # 故意不放人名做動機，測 anchor 多樣性
    ("廢棄太空站的休眠艙區", "Vega"),
    ("深夜仍亮燈的生物實驗大樓", "韓哲"),
    ("被洪水圍困的水壩控制室", "鄭薇"),
]


def _contains_any(text: str, needles) -> bool:
    return any(n and n in text for n in needles)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--runs", type=int, default=6)
    ap.add_argument("--contract", action="store_true",
                    help="啟用 OpeningVariationContract（after 模式）")
    ap.add_argument("--config", default=str(ROOT / "config" / "config.json"))
    ap.add_argument("--out", default=None, help="把結果 JSON 寫到檔案")
    args = ap.parse_args()

    # 開場序幕走 kernel intro beat；contract 整合掛在 narrative control 之外，獨立 flag。
    os.environ["ENABLE_NARRATIVE_CONTROL"] = "true"
    os.environ["ENABLE_OPENING_VARIATION_CONTRACT"] = "true" if args.contract else "false"

    sys.path.insert(0, str(ROOT / "dev" / "tools"))
    import tempfile
    from dev.tools.agent_play import _make_caller
    from core.orchestrator_loop import BeatLoop
    from core.blackboard import Blackboard
    from core.persistence.db import Database
    from core.signal import SignalBus

    caller, note = _make_caller(args.config, no_llm=False)
    mode = "AFTER(contract=ON)" if args.contract else "BASELINE(contract=OFF)"
    print(f"[{mode}] caller={note} runs={args.runs}")

    # 共用一個 DB 檔 + 每局不同 run_id → cooldown ledger 跨 6 局累積、seed 各不同。
    shared_db_path = tempfile.mktemp(suffix=".db")
    shared_db = Database(shared_db_path)

    def make_loop(i):
        return BeatLoop(caller, Blackboard(), shared_db, SignalBus(),
                        run_id=f"ov-{i}", use_kernel=True)

    records = []
    paper = name = missing = 0
    medium_dist: dict = {}
    motive_dist: dict = {}
    for i in range(args.runs):
        theme, proto = THEMES[i % len(THEMES)]
        loop = make_loop(i)
        res = loop.start({"theme": theme, "protagonist_name": proto, "npc_count": 2})
        narrative = res.get("narrative") or ""
        ov = (res.get("opening_variation") or {}) if isinstance(res, dict) else {}
        has_paper = _contains_any(narrative, PAPER_LITERALS)
        has_name = NAME_LITERAL in narrative
        has_missing = _contains_any(narrative, MISSING_PERSON_CUES)
        paper += int(has_paper)
        name += int(has_name)
        missing += int(has_missing)
        med = ov.get("message_medium", "(n/a)")
        mot = ov.get("motive_archetype", "(n/a)")
        medium_dist[med] = medium_dist.get(med, 0) + 1
        motive_dist[mot] = motive_dist.get(mot, 0) + 1
        viol = (res.get("opening_variation_violation") or {}) if isinstance(res, dict) else {}
        rec = {
            "run": i, "theme": theme, "paper": has_paper, "name": has_name,
            "missing_person": has_missing, "message_medium": med, "motive": mot,
            "violation": bool(viol.get("has_violation")),
            "fallback_used": bool(viol.get("fallback_used")),
            "head": narrative[:90].replace("\n", " "),
        }
        records.append(rec)
        flags = []
        if has_paper: flags.append("紙條")
        if has_name: flags.append("林晨")
        if has_missing: flags.append("missing_person")
        print(f"  [{i}] {theme[:14]:<14} medium={med:<18} motive={mot:<18} "
              f"{'⚠'+','.join(flags) if flags else 'ok'}"
              f"{'  VIOL' if rec['violation'] else ''}{'  FALLBACK' if rec['fallback_used'] else ''}")
        print(f"       {rec['head']}")

    n = args.runs
    summary = {
        "mode": mode, "runs": n,
        "paper_rate": round(paper / n, 3),
        "name_rate": round(name / n, 3),
        "missing_person_rate": round(missing / n, 3),
        "message_medium_dist": medium_dist,
        "motive_dist": motive_dist,
        "records": records,
    }
    print("\n=== SUMMARY ===")
    print(f"  紙條 rate          : {summary['paper_rate']}")
    print(f"  林晨 rate          : {summary['name_rate']}")
    print(f"  missing_person rate: {summary['missing_person_rate']}")
    print(f"  message_medium dist: {medium_dist}")
    print(f"  motive dist        : {motive_dist}")
    if args.out:
        Path(args.out).write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"  written → {args.out}")


if __name__ == "__main__":
    main()
