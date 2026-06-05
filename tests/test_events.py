"""U11 Event 抽取測試。"""
from core.events import event_extract, merge_with_self_report


def test_searched_location_with_known():
    evs = event_extract("我搜索地下室的角落", "", known_locations=["地下室"])
    assert {"type": "searched_location", "target": "地下室"} in evs


def test_questioned_npc_with_known():
    evs = event_extract("我質問張醫生關於那晚的事", "", known_npcs=["張醫生"])
    assert {"type": "questioned_npc", "npc": "張醫生"} in evs


def test_picked_item_with_known():
    evs = event_extract("我撿起地上的鑰匙", "", known_items=["鑰匙"])
    assert {"type": "picked_item", "item": "鑰匙"} in evs


def test_reached_location_from_story_narration():
    evs = event_extract("往前走", "你抵達頂樓，風很大。", known_locations=["頂樓"])
    assert {"type": "reached_location", "location": "頂樓"} in evs


def test_heuristic_without_known_lists():
    evs = event_extract("搜索儲藏室", "")
    assert any(e["type"] == "searched_location" and e.get("target") == "儲藏室" for e in evs)


def test_no_extraction_returns_empty():
    assert event_extract("我靜靜站著思考", "四周一片寂靜。") == []


def test_dedup():
    evs = event_extract("我搜索地下室，再搜索地下室", "", known_locations=["地下室"])
    assert evs.count({"type": "searched_location", "target": "地下室"}) == 1


def test_merge_union_and_code_authoritative():
    extracted = [{"type": "searched_location", "target": "地下室"}]
    merged = merge_with_self_report(extracted, ["frag_known"])
    assert "frag_known" in merged["touched"] and "地下室" in merged["touched"]
    assert merged["authoritative"] == "code_extract"
    assert merged["events"] == extracted  # 程式碼抽取被保留（衝突信程式碼）


def test_merge_empty_self_report():
    extracted = [{"type": "picked_item", "item": "鑰匙"}]
    merged = merge_with_self_report(extracted, None)
    assert merged["touched"] == ["鑰匙"]
