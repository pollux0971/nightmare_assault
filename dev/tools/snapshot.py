#!/usr/bin/env python3
"""純檔案快照工具 — Nightmare Assault 開發 OS 的回滾骨幹（零 git 依賴）。

快照 = 把來源程式碼子樹的時間戳副本複製進 dev/snapshots/<id>/payload/，
並寫 MANIFEST.json（含每檔 sha256）。回滾就是把某個快照覆蓋回工作區。
關機 / context 缺失後，全新 session 靠這個工具機械式還原到 known-good 狀態，
完全不依賴 git，也不依賴上一個 session 的記憶。

用法:
  snapshot.py snapshot <story-id> <phase> [--verify pass|fail|none] [--parent <id>] [--note "..."]
  snapshot.py restore  <snapshot-id> [--yes]
  snapshot.py list     [--story <story-id>]
  snapshot.py verify   <snapshot-id>
  snapshot.py latest-good <story-id>     # 印出該 story 最後一個 verify=pass 的快照 id

phase: pre（開工前還原點）| post（驗收通過的 known-good）| adhoc（臨時）
"""
from __future__ import annotations

import argparse
import datetime as _dt
import hashlib
import json
import shutil
import sys
from pathlib import Path

# ── 路徑 ──────────────────────────────────────────────────────────────────
ROOT = Path(__file__).resolve().parents[2]          # repo 根（dev/ 的上一層）
SNAP_DIR = ROOT / "dev" / "snapshots"
INDEX_MD = SNAP_DIR / "INDEX.md"

# 要快照的來源子樹（只複製存在者）。刻意排除執行期/密鑰檔。
TRACKED = [
    "core", "ui", "skills", "config", "story_templates", "tests", "data",
    "webview_app.py", "main.py", "pyproject.toml", "conftest.py",
]
EXCLUDE_DIRS = {"__pycache__", ".git", ".venv", "venv", "node_modules"}
EXCLUDE_FILES = {"config.json", ".env", "test_api.txt"}
EXCLUDE_SUFFIX = {".pyc", ".db", ".sqlite3"}
VALID_PHASES = {"pre", "post", "adhoc"}


def _ts() -> str:
    return _dt.datetime.now().strftime("%Y%m%d-%H%M%S")


def _sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for chunk in iter(lambda: fh.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


def _skip(path: Path) -> bool:
    if path.name in EXCLUDE_FILES or path.suffix in EXCLUDE_SUFFIX:
        return True
    return any(part in EXCLUDE_DIRS for part in path.parts)


def _iter_source_files():
    """yield (abs_path, rel_path_str) 為所有要快照的來源檔。"""
    for entry in TRACKED:
        p = ROOT / entry
        if not p.exists():
            continue
        if p.is_file():
            if not _skip(p):
                yield p, str(p.relative_to(ROOT))
        else:
            for f in p.rglob("*"):
                if f.is_file() and not _skip(f):
                    yield f, str(f.relative_to(ROOT))


# ── 指令：snapshot ─────────────────────────────────────────────────────────
def cmd_snapshot(args) -> int:
    story = args.story_id.strip()
    phase = args.phase.strip()
    if phase not in VALID_PHASES:
        print(f"ERROR: phase 必須是 {sorted(VALID_PHASES)}，得到 {phase!r}", file=sys.stderr)
        return 2
    snap_id = f"{story}__{phase}__{_ts()}"
    dest = SNAP_DIR / snap_id
    payload = dest / "payload"
    payload.mkdir(parents=True, exist_ok=False)

    files = []
    for abs_path, rel in _iter_source_files():
        target = payload / rel
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(abs_path, target)
        files.append({"path": rel, "sha256": _sha256(abs_path), "size": abs_path.stat().st_size})

    manifest = {
        "snapshot_id": snap_id,
        "story_id": story,
        "phase": phase,
        "timestamp": _dt.datetime.now().isoformat(timespec="seconds"),
        "verification": args.verify,          # pass | fail | none
        "parent": args.parent,
        "note": args.note,
        "root": str(ROOT),
        "file_count": len(files),
        "files": sorted(files, key=lambda x: x["path"]),
    }
    (dest / "MANIFEST.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    _rebuild_index()
    print(f"OK snapshot {snap_id}  ({len(files)} files, verify={args.verify})")
    print(snap_id)        # 末行印純 id，方便腳本擷取
    return 0


# ── 指令：restore ──────────────────────────────────────────────────────────
def cmd_restore(args) -> int:
    snap_id = args.snapshot_id.strip()
    dest = SNAP_DIR / snap_id
    payload = dest / "payload"
    if not (dest / "MANIFEST.json").exists():
        print(f"ERROR: 找不到快照 {snap_id}", file=sys.stderr)
        return 2
    manifest = json.loads((dest / "MANIFEST.json").read_text(encoding="utf-8"))
    if not args.yes:
        print(f"將用快照 {snap_id}（{manifest['file_count']} 檔，"
              f"phase={manifest['phase']}, verify={manifest['verification']}）覆蓋工作區。")
        print("這會刪除該快照涵蓋子樹中、快照後新增的檔。確認請加 --yes。")
        return 1

    # 1) 清掉被快照涵蓋的 TRACKED 子樹（達成乾淨還原，含刪除新增檔）
    for entry in TRACKED:
        p = ROOT / entry
        if not p.exists():
            continue
        if p.is_dir():
            for f in list(p.rglob("*")):
                if f.is_file() and not _skip(f):
                    f.unlink()
        elif not _skip(p):
            p.unlink()
    # 2) 從 payload 複製回去
    restored = 0
    for f in payload.rglob("*"):
        if f.is_file():
            rel = f.relative_to(payload)
            target = ROOT / rel
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(f, target)
            restored += 1
    # 3) 清掉因刪檔殘留的空目錄
    for entry in TRACKED:
        p = ROOT / entry
        if p.is_dir():
            for d in sorted(p.rglob("*"), reverse=True):
                if d.is_dir() and not any(d.iterdir()):
                    d.rmdir()
    print(f"OK restore {snap_id}  →  還原 {restored} 檔到工作區")
    return 0


# ── 指令：list ─────────────────────────────────────────────────────────────
def _all_manifests():
    out = []
    if not SNAP_DIR.exists():
        return out
    for d in sorted(SNAP_DIR.iterdir()):
        mf = d / "MANIFEST.json"
        if mf.is_file():
            try:
                out.append(json.loads(mf.read_text(encoding="utf-8")))
            except Exception:
                pass
    return out


def cmd_list(args) -> int:
    rows = _all_manifests()
    if args.story:
        rows = [m for m in rows if m.get("story_id") == args.story]
    if not rows:
        print("（無快照）")
        return 0
    print(f"{'snapshot_id':<48} {'phase':<6} {'verify':<6} {'files':<5} timestamp")
    print("-" * 92)
    for m in rows:
        print(f"{m['snapshot_id']:<48} {m.get('phase',''):<6} "
              f"{str(m.get('verification','')):<6} {m.get('file_count',0):<5} {m.get('timestamp','')}")
    return 0


# ── 指令：verify ───────────────────────────────────────────────────────────
def cmd_verify(args) -> int:
    snap_id = args.snapshot_id.strip()
    dest = SNAP_DIR / snap_id
    mf = dest / "MANIFEST.json"
    if not mf.is_file():
        print(f"ERROR: 找不到快照 {snap_id}", file=sys.stderr)
        return 2
    manifest = json.loads(mf.read_text(encoding="utf-8"))
    payload = dest / "payload"
    bad = []
    for rec in manifest["files"]:
        f = payload / rec["path"]
        if not f.is_file():
            bad.append((rec["path"], "missing"))
        elif _sha256(f) != rec["sha256"]:
            bad.append((rec["path"], "sha256-mismatch"))
    if bad:
        print(f"FAIL: 快照 {snap_id} 完整性校驗失敗（{len(bad)} 檔）")
        for path, why in bad[:20]:
            print(f"  - {path}: {why}")
        return 1
    print(f"OK verify {snap_id}  —  {manifest['file_count']} 檔完整")
    return 0


# ── 指令：latest-good ──────────────────────────────────────────────────────
def cmd_latest_good(args) -> int:
    rows = [m for m in _all_manifests()
            if m.get("story_id") == args.story_id and m.get("verification") == "pass"]
    if not rows:
        print("", end="")        # 空字串：表示沒有 known-good
        return 1
    rows.sort(key=lambda m: m["snapshot_id"])
    print(rows[-1]["snapshot_id"])
    return 0


# ── INDEX.md 重建 ──────────────────────────────────────────────────────────
def _rebuild_index() -> None:
    rows = _all_manifests()
    lines = [
        "# 快照索引（INDEX.md）",
        "",
        "> 由 `dev/tools/snapshot.py` 自動重建。回滾用 known-good 還原點清單。",
        "",
        "| snapshot_id | story | phase | verify | files | parent | timestamp |",
        "|---|---|---|---|---|---|---|",
    ]
    for m in rows:
        lines.append(
            f"| {m['snapshot_id']} | {m.get('story_id','')} | {m.get('phase','')} "
            f"| {m.get('verification','')} | {m.get('file_count',0)} "
            f"| {m.get('parent') or '-'} | {m.get('timestamp','')} |"
        )
    lines.append("")
    INDEX_MD.write_text("\n".join(lines), encoding="utf-8")


# ── CLI ────────────────────────────────────────────────────────────────────
def main(argv=None) -> int:
    SNAP_DIR.mkdir(parents=True, exist_ok=True)
    p = argparse.ArgumentParser(description="純檔案快照工具（回滾骨幹，零 git）")
    sub = p.add_subparsers(dest="cmd", required=True)

    s = sub.add_parser("snapshot", help="建立快照")
    s.add_argument("story_id")
    s.add_argument("phase", help="pre | post | adhoc")
    s.add_argument("--verify", choices=["pass", "fail", "none"], default="none")
    s.add_argument("--parent", default=None)
    s.add_argument("--note", default=None)
    s.set_defaults(func=cmd_snapshot)

    r = sub.add_parser("restore", help="把快照覆蓋回工作區")
    r.add_argument("snapshot_id")
    r.add_argument("--yes", action="store_true", help="確認執行（否則只預演）")
    r.set_defaults(func=cmd_restore)

    l = sub.add_parser("list", help="列出快照")
    l.add_argument("--story", default=None)
    l.set_defaults(func=cmd_list)

    v = sub.add_parser("verify", help="校驗快照完整性")
    v.add_argument("snapshot_id")
    v.set_defaults(func=cmd_verify)

    g = sub.add_parser("latest-good", help="印出該 story 最後一個 verify=pass 的快照 id")
    g.add_argument("story_id")
    g.set_defaults(func=cmd_latest_good)

    args = p.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
