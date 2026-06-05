"""tests/test_db.py — SQLite 持久化層驗收測試（U05）。

全部使用 :memory: 資料庫，不依賴磁碟，可平行執行。
"""

import pytest
from core.persistence.db import Database


# ─────────────────────────────────────────────
# fixtures
# ─────────────────────────────────────────────

@pytest.fixture
def db():
    """每個測試拿到全新的 in-memory DB。"""
    database = Database(":memory:")
    yield database
    database.close()


# ─────────────────────────────────────────────
# 1. 初始化：schema_version 與 8 張表
# ─────────────────────────────────────────────

EXPECTED_TABLES = {
    "schema_meta",
    "runs",
    "beats",
    "npc_states",
    "inventory_snapshots",
    "chat_logs",
    "save_points",
    "llm_traces",
}


def test_schema_version(db):
    assert db.schema_version() == "1"


def test_all_tables_exist(db):
    rows = db._conn.execute(
        "SELECT name FROM sqlite_master WHERE type='table';"
    ).fetchall()
    tables = {r["name"] for r in rows}
    assert EXPECTED_TABLES.issubset(tables), (
        f"缺少表格: {EXPECTED_TABLES - tables}"
    )


def test_indexes_exist(db):
    rows = db._conn.execute(
        "SELECT name FROM sqlite_master WHERE type='index';"
    ).fetchall()
    index_names = {r["name"] for r in rows}
    expected = {
        "idx_beats_run_beat",
        "idx_npc_states_run_beat",
        "idx_inventory_run_beat",
        "idx_chat_logs_run_npc",
        "idx_llm_traces_run_beat",
    }
    assert expected.issubset(index_names), (
        f"缺少 index: {expected - index_names}"
    )


# ─────────────────────────────────────────────
# 2. create_run / list_runs
# ─────────────────────────────────────────────

def test_create_run_and_list(db):
    db.create_run("r001", title="黑暗莊園", theme="horror", difficulty="hard")
    runs = db.list_runs()
    assert len(runs) == 1
    r = runs[0]
    assert r["run_id"] == "r001"
    assert r["current_beat"] == 0
    assert r["updated_at"] is not None


def test_list_runs_multiple(db):
    for i in range(3):
        db.create_run(f"run-{i}", title=f"遊戲{i}")
    assert len(db.list_runs()) == 3


def test_list_runs_returns_required_keys(db):
    db.create_run("rx")
    r = db.list_runs()[0]
    for key in ("run_id", "current_beat", "current_location", "updated_at"):
        assert key in r, f"list_runs 缺少欄位 {key!r}"


# ─────────────────────────────────────────────
# 3. save_beat / load_beat
# ─────────────────────────────────────────────

def test_save_and_load_beat(db):
    db.create_run("r1")
    db.save_beat(
        run_id="r1",
        beat_number=5,
        narrative="你走進黑暗走廊",
        decision_json='{"choices": ["逃跑", "探索"]}',
        rolling_summary_snapshot="到目前為止你在走廊",
        blackboard_snapshot_json='{"hp": 80}',
        is_narration_only=False,
    )
    beat = db.load_beat("r1", 5)
    assert beat is not None
    assert beat["run_id"] == "r1"
    assert beat["beat_number"] == 5
    assert beat["narrative"] == "你走進黑暗走廊"
    assert beat["decision_json"] == '{"choices": ["逃跑", "探索"]}'
    assert beat["rolling_summary_snapshot"] == "到目前為止你在走廊"
    assert beat["blackboard_snapshot_json"] == '{"hp": 80}'
    assert beat["is_narration_only"] == 0


def test_load_beat_not_found(db):
    db.create_run("r2")
    assert db.load_beat("r2", 999) is None


def test_save_beat_upsert_overwrites(db):
    db.create_run("r3")
    db.save_beat("r3", 1, "第一次", None, "S1", None)
    db.save_beat("r3", 1, "第二次覆寫", None, "S1-updated", None)
    beat = db.load_beat("r3", 1)
    assert beat["narrative"] == "第二次覆寫"
    assert beat["rolling_summary_snapshot"] == "S1-updated"
    # 確認只有一筆
    count = db._conn.execute(
        "SELECT COUNT(*) FROM beats WHERE run_id='r3' AND beat_number=1;"
    ).fetchone()[0]
    assert count == 1


# ─────────────────────────────────────────────
# 4. A3 回溯不錯亂（核心驗收）
# ─────────────────────────────────────────────

def test_a3_rollback_isolation(db):
    """
    beat 10 的 rolling_summary_snapshot 不能被 beat 30 的寫入污染。
    load_beat(run, 10) 必須還是 'S10'，不是 'S30'。
    """
    db.create_run("run-a3")
    db.save_beat(
        run_id="run-a3",
        beat_number=10,
        narrative="beat ten narrative",
        decision_json=None,
        rolling_summary_snapshot="S10",
        blackboard_snapshot_json=None,
    )
    db.save_beat(
        run_id="run-a3",
        beat_number=30,
        narrative="beat thirty narrative",
        decision_json=None,
        rolling_summary_snapshot="S30",
        blackboard_snapshot_json=None,
    )
    beat10 = db.load_beat("run-a3", 10)
    beat30 = db.load_beat("run-a3", 30)

    assert beat10 is not None, "beat 10 應存在"
    assert beat30 is not None, "beat 30 應存在"
    assert beat10["rolling_summary_snapshot"] == "S10", (
        f"回溯錯亂！beat 10 的 snapshot 變成了 {beat10['rolling_summary_snapshot']!r}"
    )
    assert beat30["rolling_summary_snapshot"] == "S30"


def test_a3_upsert_does_not_bleed_to_other_beats(db):
    """對 beat 10 覆寫不影響 beat 20。"""
    db.create_run("run-a3b")
    db.save_beat("run-a3b", 10, "n10", None, "S10", None)
    db.save_beat("run-a3b", 20, "n20", None, "S20", None)
    # 再次覆寫 beat 10
    db.save_beat("run-a3b", 10, "n10-v2", None, "S10-v2", None)

    assert db.load_beat("run-a3b", 20)["rolling_summary_snapshot"] == "S20"
    assert db.load_beat("run-a3b", 10)["rolling_summary_snapshot"] == "S10-v2"


# ─────────────────────────────────────────────
# 5. save_points
# ─────────────────────────────────────────────

def test_add_and_list_save_points(db):
    db.create_run("rsp")
    db.add_save_point("rsp", 5, "大廳前")
    db.add_save_point("rsp", 15, "地窖入口")
    db.add_save_point("rsp", 25, "Boss 前")
    points = db.list_save_points("rsp")
    assert len(points) == 3
    assert points[0]["beat_number"] == 5
    assert points[0]["label"] == "大廳前"
    assert points[1]["beat_number"] == 15
    assert points[2]["beat_number"] == 25


def test_list_save_points_empty(db):
    db.create_run("rsp_empty")
    assert db.list_save_points("rsp_empty") == []


def test_list_save_points_only_own_run(db):
    db.create_run("run-A")
    db.create_run("run-B")
    db.add_save_point("run-A", 1, "A存檔")
    db.add_save_point("run-B", 2, "B存檔")
    assert len(db.list_save_points("run-A")) == 1
    assert db.list_save_points("run-A")[0]["label"] == "A存檔"


# ─────────────────────────────────────────────
# 6. llm_traces / llm_cost_summary
# ─────────────────────────────────────────────

def test_write_llm_trace_and_cost_summary(db):
    db.create_run("rllm")
    db.write_llm_trace("rllm", 1, "narrator", "gpt-4o", "abc123", 100, 200, 1500, True)
    db.write_llm_trace("rllm", 2, "npc_agent", "gpt-4o", "def456", 150, 300, 2000, True)
    db.write_llm_trace("rllm", 3, "narrator", "gpt-4o", "ghi789", 50, 80, 800, False, "timeout")

    summary = db.llm_cost_summary("rllm")
    assert summary["call_count"] == 3
    assert summary["total_input_tokens"] == 300   # 100+150+50
    assert summary["total_output_tokens"] == 580  # 200+300+80


def test_llm_cost_summary_empty_run(db):
    db.create_run("rllm_empty")
    summary = db.llm_cost_summary("rllm_empty")
    assert summary["call_count"] == 0
    assert summary["total_input_tokens"] == 0
    assert summary["total_output_tokens"] == 0


def test_llm_cost_summary_only_own_run(db):
    db.create_run("run-X")
    db.create_run("run-Y")
    db.write_llm_trace("run-X", 1, "a", "m", "h", 100, 200, 500, True)
    db.write_llm_trace("run-Y", 1, "a", "m", "h", 999, 888, 500, True)

    sx = db.llm_cost_summary("run-X")
    assert sx["total_input_tokens"] == 100
    sy = db.llm_cost_summary("run-Y")
    assert sy["total_input_tokens"] == 999


# ─────────────────────────────────────────────
# 7. 邊界：重複初始化（冪等）
# ─────────────────────────────────────────────

def test_reinit_is_idempotent(db):
    """對同一連線再次呼叫 _init_schema 不應報錯，schema_version 仍為 '1'。"""
    db._init_schema()
    db._init_schema()
    assert db.schema_version() == "1"


# ─────────────────────────────────────────────
# 8. context manager
# ─────────────────────────────────────────────

def test_context_manager():
    with Database(":memory:") as d:
        d.create_run("cm_run")
        assert len(d.list_runs()) == 1
    # close 之後連線已關，再次操作應 raise
    with pytest.raises(Exception):
        d.list_runs()
