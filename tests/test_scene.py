"""
tests/test_scene.py — pytest suite for core/scene.py (工單 U04：場景系統).
"""

import pytest

from core.models import Interactable, Location, SceneRegistry
from core.scene import SceneManager, validate_unique_location_ids


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

def make_registry() -> SceneRegistry:
    """Build a minimal scene graph with three locations and exits.

    Layout:
        foyer  --[hallway]--> hallway  --[library]--> library
               --[library]-->                       (dead-end)
        foyer  <--[foyer]-- hallway
    """
    foyer = Location(
        id="foyer",
        name="入口大廳",
        description="昏暗的入口大廳，空氣中瀰漫著腐爛氣息。",
        discovered=True,         # starting location — already discovered
        exits=["hallway"],
        interactables=[
            Interactable(id="candle", type="item"),
            Interactable(id="blood_clue", type="clue"),
        ],
    )
    hallway = Location(
        id="hallway",
        name="長廊",
        description="延伸至黑暗深處的長廊。",
        discovered=False,
        exits=["foyer", "library"],
        interactables=[],
    )
    library = Location(
        id="library",
        name="圖書室",
        description="塵封書架，地板上散落著碎紙。",
        discovered=False,
        exits=["hallway"],
        interactables=[],
    )
    return SceneRegistry(
        current_location="foyer",
        known_locations=[foyer, hallway, library],
    )


@pytest.fixture
def manager() -> SceneManager:
    return SceneManager(make_registry())


# ---------------------------------------------------------------------------
# current() / get()
# ---------------------------------------------------------------------------

def test_current_returns_starting_location(manager: SceneManager) -> None:
    loc = manager.current()
    assert loc.id == "foyer"
    assert loc.name == "入口大廳"


def test_get_existing_location(manager: SceneManager) -> None:
    loc = manager.get("library")
    assert loc.id == "library"


def test_get_unknown_location_raises_key_error(manager: SceneManager) -> None:
    with pytest.raises(KeyError):
        manager.get("nonexistent_room")


# ---------------------------------------------------------------------------
# move()
# ---------------------------------------------------------------------------

def test_move_along_valid_exit_returns_destination(manager: SceneManager) -> None:
    dest = manager.move("hallway")
    assert dest.id == "hallway"


def test_move_updates_current_location(manager: SceneManager) -> None:
    manager.move("hallway")
    assert manager.current().id == "hallway"


def test_move_sets_destination_discovered(manager: SceneManager) -> None:
    assert manager.get("hallway").discovered is False
    manager.move("hallway")
    assert manager.get("hallway").discovered is True


def test_move_to_non_exit_raises_value_error(manager: SceneManager) -> None:
    # library is NOT a direct exit from foyer
    with pytest.raises(ValueError, match="library"):
        manager.move("library")


def test_move_chain_across_multiple_rooms(manager: SceneManager) -> None:
    manager.move("hallway")
    manager.move("library")
    assert manager.current().id == "library"
    assert manager.get("library").discovered is True


def test_move_back_uses_reverse_exit(manager: SceneManager) -> None:
    manager.move("hallway")
    manager.move("foyer")   # hallway → foyer exit exists
    assert manager.current().id == "foyer"


# ---------------------------------------------------------------------------
# location_reached()
# ---------------------------------------------------------------------------

def test_starting_location_reached(manager: SceneManager) -> None:
    # foyer starts with discovered=True
    assert manager.location_reached("foyer") is True


def test_unvisited_location_not_reached(manager: SceneManager) -> None:
    assert manager.location_reached("hallway") is False


def test_location_reached_after_move(manager: SceneManager) -> None:
    manager.move("hallway")
    assert manager.location_reached("hallway") is True


def test_location_reached_unknown_id_returns_false(manager: SceneManager) -> None:
    assert manager.location_reached("ghost_wing") is False


# ---------------------------------------------------------------------------
# plant_interactable()
# ---------------------------------------------------------------------------

def test_plant_interactable_appears_in_location(manager: SceneManager) -> None:
    corpse = Interactable(id="victim_01", type="corpse")
    manager.plant_interactable("hallway", corpse)
    ids = [it.id for it in manager.get("hallway").interactables]
    assert "victim_01" in ids


def test_plant_interactable_duplicate_id_raises_value_error(
    manager: SceneManager,
) -> None:
    corpse_a = Interactable(id="victim_01", type="corpse")
    corpse_b = Interactable(id="victim_01", type="corpse")  # same id
    manager.plant_interactable("hallway", corpse_a)
    with pytest.raises(ValueError, match="victim_01"):
        manager.plant_interactable("hallway", corpse_b)


def test_plant_interactable_unknown_location_raises_key_error(
    manager: SceneManager,
) -> None:
    item = Interactable(id="mysterious_box", type="item")
    with pytest.raises(KeyError):
        manager.plant_interactable("attic", item)


def test_plant_multiple_different_ids_allowed(manager: SceneManager) -> None:
    manager.plant_interactable("library", Interactable(id="note_01", type="clue"))
    manager.plant_interactable("library", Interactable(id="door_01", type="door"))
    ids = [it.id for it in manager.get("library").interactables]
    assert "note_01" in ids
    assert "door_01" in ids


# ---------------------------------------------------------------------------
# reveal_interactable() / revealed_interactables()
# ---------------------------------------------------------------------------

def test_reveal_interactable_sets_flag(manager: SceneManager) -> None:
    assert manager.get("foyer").interactables[0].revealed is False
    manager.reveal_interactable("foyer", "candle")
    assert manager.get("foyer").interactables[0].revealed is True


def test_revealed_interactables_only_returns_revealed(manager: SceneManager) -> None:
    manager.reveal_interactable("foyer", "candle")
    revealed = manager.revealed_interactables("foyer")
    assert len(revealed) == 1
    assert revealed[0].id == "candle"
    assert all(it.revealed for it in revealed)


def test_revealed_interactables_empty_when_none_revealed(
    manager: SceneManager,
) -> None:
    result = manager.revealed_interactables("foyer")
    assert result == []


def test_reveal_interactable_unknown_interactable_raises_key_error(
    manager: SceneManager,
) -> None:
    with pytest.raises(KeyError):
        manager.reveal_interactable("foyer", "nonexistent_thing")


def test_reveal_interactable_unknown_location_raises_key_error(
    manager: SceneManager,
) -> None:
    with pytest.raises(KeyError):
        manager.reveal_interactable("secret_vault", "hidden_item")


def test_plant_and_reveal_corpse_type(manager: SceneManager) -> None:
    """植入 corpse 型 interactable 並 reveal（結局情境）。"""
    corpse = Interactable(
        id="npc_body_01",
        type="corpse",
        linked_fragment="fragment_ending_death",
        revealed=False,
    )
    manager.plant_interactable("library", corpse)
    # Before reveal
    assert manager.revealed_interactables("library") == []
    manager.reveal_interactable("library", "npc_body_01")
    revealed = manager.revealed_interactables("library")
    assert len(revealed) == 1
    assert revealed[0].type == "corpse"
    assert revealed[0].linked_fragment == "fragment_ending_death"
    assert revealed[0].revealed is True


# ---------------------------------------------------------------------------
# validate_unique_location_ids() — CHECKLIST A6
# ---------------------------------------------------------------------------

def test_validate_unique_ids_passes_for_valid_registry(
    manager: SceneManager,
) -> None:
    # Should not raise
    validate_unique_location_ids(manager._registry)


def test_validate_unique_ids_raises_on_duplicate() -> None:
    loc_a = Location(
        id="duplicate_room",
        name="Room A",
        description="First",
        exits=[],
    )
    loc_b = Location(
        id="duplicate_room",
        name="Room B",
        description="Second",
        exits=[],
    )
    registry = SceneRegistry(
        current_location="duplicate_room",
        known_locations=[loc_a, loc_b],
    )
    with pytest.raises(ValueError, match="duplicate_room"):
        validate_unique_location_ids(registry)


def test_validate_unique_ids_raises_on_multiple_duplicates() -> None:
    locs = [
        Location(id="room_x", name="X1", description=".", exits=[]),
        Location(id="room_x", name="X2", description=".", exits=[]),
        Location(id="room_y", name="Y1", description=".", exits=[]),
        Location(id="room_y", name="Y2", description=".", exits=[]),
        Location(id="room_z", name="Z",  description=".", exits=[]),
    ]
    registry = SceneRegistry(current_location="room_x", known_locations=locs)
    with pytest.raises(ValueError):
        validate_unique_location_ids(registry)


def test_validate_unique_ids_empty_registry_passes() -> None:
    registry = SceneRegistry(current_location="nowhere", known_locations=[])
    validate_unique_location_ids(registry)  # must not raise


def test_validate_unique_ids_single_location_passes() -> None:
    registry = SceneRegistry(
        current_location="only_room",
        known_locations=[
            Location(id="only_room", name="Only", description=".", exits=[])
        ],
    )
    validate_unique_location_ids(registry)  # must not raise
