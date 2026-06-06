"""補丁：修 Full UX Selfplay 暴露的兩個 bug。

#3 WorldDelta id：core/world/extractor.py 用 `id=` 建 WorldDelta（欄位實為 entity_id）→ 已知 NPC 名
   出現在敘事且 story 無結構化 delta 時，fallback extract_entities 會拋 TypeError → world model tick skipped。
#6 placeholder 洩漏：story 把內部識別碼（crumpled_paper / origine_verification / scene.beyond）、
   壞分隔符（<<<CONTINAlUE>>>）、黏中文的拉丁殘片（然homme）寫進敘事 → SurfaceTextSanitizer 未覆蓋。
"""
from __future__ import annotations

from core.narrative.sanitizer import SurfaceTextSanitizer, sanitize_text
from core.world.extractor import extract_entities
from core.world.model import ACTOR, WorldDelta, WorldModel


# ── #3 WorldDelta id（extractor）─────────────────────────────────────────────

def test_extract_entities_actor_uses_entity_id_no_crash():
    """已知 NPC 名出現在敘事 → 不得拋 'unexpected keyword argument id'。"""
    deltas = extract_entities("走廊盡頭，林守一正背對著你整理工具。", npc_names=["林守一"])
    actor = [d for d in deltas if d.kind == ACTOR]
    assert actor, "應登記 actor delta"
    assert actor[0].entity_id == "actor.林守一"   # 用 entity_id，不是 id
    # WorldDelta 沒有 id 欄位（回歸守門）
    assert not hasattr(WorldDelta(op="register", kind=ACTOR, label="x"), "id")


def test_extract_entities_actor_applies_to_world_model():
    """fallback 路徑（tick 用）apply 到 WorldModel 不拋、actor 有登記。"""
    w = WorldModel()
    w.apply_deltas(extract_entities("謝明德從陰影中走出。", npc_names=["謝明德"]))
    actors = w.by_kind(ACTOR)
    assert any(a.id == "actor.謝明德" for a in actors)


# ── #6 placeholder / 壞分隔符 / 黏拉丁殘片（sanitizer）────────────────────────

def test_sanitizer_strips_snake_case_identifier():
    clean, leaks = SurfaceTextSanitizer().sanitize("你下意識去crumpled_paper口袋裡摸那張門禁卡")
    assert "crumpled_paper" not in clean
    assert "internal_identifier" in leaks
    assert clean == "你下意識去口袋裡摸那張門禁卡"


def test_sanitizer_strips_underscore_and_dotted_ids():
    assert "origine_verification" not in sanitize_text("終端機顯示 origine_verification 失敗")
    assert "scene.beyond" not in sanitize_text("通往 scene.beyond 的門")


def test_sanitizer_strips_corrupted_delimiter():
    clean, leaks = SurfaceTextSanitizer().sanitize("<<<CONTINAlUE>>> 雨水順著你的衣領滑入脊背")
    assert "<<<" not in clean and ">>>" not in clean and "CONTINAlUE" not in clean
    assert "delim_fragment" in leaks
    assert clean.strip().startswith("雨水")


def test_sanitizer_strips_cjk_glued_latin():
    clean, leaks = SurfaceTextSanitizer().sanitize("他低聲說了句然homme，然後消失")
    assert "homme" not in clean
    assert "glued_latin" in leaks


def test_sanitizer_preserves_legit_ascii_and_numbers():
    """不得誤砍：頻率/門牌/縮寫/單位/短碼。"""
    src = "頻率穩定在432.7赫茲，門牌B-12，凌晨3點a.m.還亮著17Hz的燈"
    clean, leaks = SurfaceTextSanitizer().sanitize(src)
    assert clean == src
    assert leaks == []


def test_sanitizer_preserves_plain_chinese():
    src = "正常的中文敘事，沒有任何洩漏，門在你面前緩緩打開。"
    assert sanitize_text(src) == src
