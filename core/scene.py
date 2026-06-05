"""
core/scene.py — Scene management for Nightmare Assault.
Provides SceneManager (operates on a SceneRegistry) and a registry-level
validate_unique_location_ids helper.
"""

from __future__ import annotations

from core.models import Interactable, Location, SceneRegistry


def validate_unique_location_ids(registry: SceneRegistry) -> None:
    """Verify that every Location.id in *registry* is globally unique.

    Raises
    ------
    ValueError
        If any id appears more than once.
    """
    seen: set[str] = set()
    duplicates: set[str] = set()
    for loc in registry.known_locations:
        if loc.id in seen:
            duplicates.add(loc.id)
        seen.add(loc.id)
    if duplicates:
        raise ValueError(
            f"Duplicate location id(s) detected in registry: {sorted(duplicates)}"
        )


class SceneManager:
    """Manages scene state by mutating a :class:`~core.models.SceneRegistry`.

    All mutations are applied in-place to the Pydantic objects stored in the
    registry so that the same registry instance can be serialised at any time.
    """

    def __init__(self, registry: SceneRegistry) -> None:
        self._registry = registry

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _location_map(self) -> dict[str, Location]:
        """Return a {id: Location} mapping built from known_locations."""
        return {loc.id: loc for loc in self._registry.known_locations}

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def current(self) -> Location:
        """Return the Location object for the current position."""
        loc_map = self._location_map()
        current_id = self._registry.current_location
        if current_id not in loc_map:
            raise KeyError(
                f"current_location '{current_id}' not found in known_locations"
            )
        return loc_map[current_id]

    def get(self, location_id: str) -> Location:
        """Return the Location with *location_id*.

        Raises
        ------
        KeyError
            If *location_id* is not in known_locations.
        """
        loc_map = self._location_map()
        if location_id not in loc_map:
            raise KeyError(f"Location '{location_id}' not found in registry")
        return loc_map[location_id]

    def move(self, to_id: str) -> Location:
        """Move the player to *to_id*.

        Rules
        -----
        * *to_id* must appear in the current location's ``exits`` list.
        * On success: ``current_location`` is updated, the destination's
          ``discovered`` flag is set to ``True``, and the destination
          Location is returned.

        Raises
        ------
        ValueError
            If *to_id* is not in the current location's exits.
        KeyError
            If *to_id* does not exist in known_locations.
        """
        here = self.current()
        if to_id not in here.exits:
            raise ValueError(
                f"Cannot move to '{to_id}' from '{here.id}': "
                f"not in exits {here.exits}"
            )
        destination = self.get(to_id)  # raises KeyError if missing
        self._registry.current_location = to_id
        destination.discovered = True
        return destination

    def location_reached(self, location_id: str) -> bool:
        """Return ``True`` if *location_id* has been visited (``discovered=True``).

        Returns ``False`` (not raises) if the location is unknown — an
        undiscovered location is definitionally unreached.
        """
        try:
            loc = self.get(location_id)
        except KeyError:
            return False
        return loc.discovered

    def plant_interactable(self, location_id: str, item: Interactable) -> None:
        """Add *item* to the interactables of the Location with *location_id*.

        Raises
        ------
        KeyError
            If *location_id* is not found.
        ValueError
            If an Interactable with the same ``id`` already exists in that
            location.
        """
        loc = self.get(location_id)
        existing_ids = {it.id for it in loc.interactables}
        if item.id in existing_ids:
            raise ValueError(
                f"Interactable '{item.id}' already exists in location '{location_id}'"
            )
        loc.interactables.append(item)

    def reveal_interactable(self, location_id: str, interactable_id: str) -> None:
        """Set ``revealed=True`` on the specified Interactable.

        Raises
        ------
        KeyError
            If *location_id* is not found, or if *interactable_id* is not
            found within that location's interactables.
        """
        loc = self.get(location_id)
        for it in loc.interactables:
            if it.id == interactable_id:
                it.revealed = True
                return
        raise KeyError(
            f"Interactable '{interactable_id}' not found in location '{location_id}'"
        )

    def revealed_interactables(self, location_id: str) -> list[Interactable]:
        """Return all Interactables in *location_id* that have ``revealed=True``.

        Raises
        ------
        KeyError
            If *location_id* is not found.
        """
        loc = self.get(location_id)
        return [it for it in loc.interactables if it.revealed]
