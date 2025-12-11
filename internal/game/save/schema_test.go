package save

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSaveDataStructure(t *testing.T) {
	// Test that SaveData can be created with all required fields
	now := time.Now()
	save := SaveData{
		Version: CurrentVersion,
		Metadata: Metadata{
			SavedAt:     now,
			PlayTime:    3600,
			Difficulty:  "normal",
			StoryLength: "medium",
		},
		Player: PlayerState{
			HP:         100,
			SAN:        80,
			Location:   "abandoned_hospital",
			Inventory:  []Item{{Name: "flashlight", Description: "A dim flashlight"}},
			KnownClues: []string{"blood_trail", "broken_window"},
		},
		Game: GameState{
			CurrentChapter:  2,
			ChapterProgress: 0.5,
			TriggeredRules:  []string{"rule_001"},
			DiscoveredRules: []string{},
		},
		Teammates: []TeammateState{
			{
				Name:         "Alice",
				Alive:        true,
				HP:           80,
				Location:     "abandoned_hospital",
				Items:        []Item{{Name: "knife", Description: "A rusty knife"}},
				Relationship: 50,
			},
		},
		Context: StoryContext{
			RecentSummary: "The group entered the abandoned hospital...",
			CurrentScene:  "You stand in a dark corridor...",
			GameBible:     "Horror setting, psychological elements...",
		},
	}

	if save.Version != CurrentVersion {
		t.Errorf("Expected version %d, got %d", CurrentVersion, save.Version)
	}

	if save.Player.HP != 100 {
		t.Errorf("Expected player HP 100, got %d", save.Player.HP)
	}

	if len(save.Teammates) != 1 {
		t.Errorf("Expected 1 teammate, got %d", len(save.Teammates))
	}
}

func TestSaveDataJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	save := SaveData{
		Version: CurrentVersion,
		Metadata: Metadata{
			SavedAt:     now,
			PlayTime:    1800,
			Difficulty:  "hard",
			StoryLength: "long",
		},
		Player: PlayerState{
			HP:         50,
			SAN:        30,
			Location:   "basement",
			Inventory:  []Item{},
			KnownClues: []string{"secret_door"},
		},
		Game: GameState{
			CurrentChapter:  3,
			ChapterProgress: 0.8,
			TriggeredRules:  []string{"rule_002", "rule_003"},
			DiscoveredRules: []string{"rule_001"},
		},
		Teammates: []TeammateState{},
		Context: StoryContext{
			RecentSummary: "Summary...",
			CurrentScene:  "Scene...",
			GameBible:     "Bible...",
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatalf("Failed to marshal SaveData: %v", err)
	}

	// Deserialize from JSON
	var loaded SaveData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal SaveData: %v", err)
	}

	// Verify all fields
	if loaded.Version != save.Version {
		t.Errorf("Version mismatch: expected %d, got %d", save.Version, loaded.Version)
	}

	if !loaded.Metadata.SavedAt.Equal(save.Metadata.SavedAt) {
		t.Errorf("SavedAt mismatch: expected %v, got %v", save.Metadata.SavedAt, loaded.Metadata.SavedAt)
	}

	if loaded.Player.HP != save.Player.HP {
		t.Errorf("Player HP mismatch: expected %d, got %d", save.Player.HP, loaded.Player.HP)
	}

	if loaded.Player.SAN != save.Player.SAN {
		t.Errorf("Player SAN mismatch: expected %d, got %d", save.Player.SAN, loaded.Player.SAN)
	}

	if loaded.Game.CurrentChapter != save.Game.CurrentChapter {
		t.Errorf("CurrentChapter mismatch: expected %d, got %d", save.Game.CurrentChapter, loaded.Game.CurrentChapter)
	}

	if len(loaded.Game.TriggeredRules) != len(save.Game.TriggeredRules) {
		t.Errorf("TriggeredRules count mismatch: expected %d, got %d", len(save.Game.TriggeredRules), len(loaded.Game.TriggeredRules))
	}
}

func TestMetadataFields(t *testing.T) {
	meta := Metadata{
		SavedAt:     time.Now(),
		PlayTime:    7200,
		Difficulty:  "nightmare",
		StoryLength: "short",
	}

	if meta.PlayTime != 7200 {
		t.Errorf("Expected PlayTime 7200, got %d", meta.PlayTime)
	}

	if meta.Difficulty != "nightmare" {
		t.Errorf("Expected Difficulty nightmare, got %s", meta.Difficulty)
	}
}

func TestPlayerStateFields(t *testing.T) {
	player := PlayerState{
		HP:         75,
		SAN:        60,
		Location:   "rooftop",
		Inventory:  []Item{{Name: "key", Description: "A brass key"}},
		KnownClues: []string{"clue1", "clue2"},
	}

	if player.HP != 75 {
		t.Errorf("Expected HP 75, got %d", player.HP)
	}

	if len(player.Inventory) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(player.Inventory))
	}

	if player.Inventory[0].Name != "key" {
		t.Errorf("Expected item name 'key', got '%s'", player.Inventory[0].Name)
	}
}

func TestGameStateFields(t *testing.T) {
	game := GameState{
		CurrentChapter:  5,
		ChapterProgress: 0.25,
		TriggeredRules:  []string{"r1", "r2", "r3"},
		DiscoveredRules: []string{"r1"},
	}

	if game.CurrentChapter != 5 {
		t.Errorf("Expected CurrentChapter 5, got %d", game.CurrentChapter)
	}

	if game.ChapterProgress != 0.25 {
		t.Errorf("Expected ChapterProgress 0.25, got %f", game.ChapterProgress)
	}

	if len(game.TriggeredRules) != 3 {
		t.Errorf("Expected 3 triggered rules, got %d", len(game.TriggeredRules))
	}
}

func TestTeammateStateFields(t *testing.T) {
	teammate := TeammateState{
		Name:         "Bob",
		Alive:        false,
		HP:           0,
		Location:     "morgue",
		Items:        []Item{},
		Relationship: -10,
	}

	if teammate.Name != "Bob" {
		t.Errorf("Expected name 'Bob', got '%s'", teammate.Name)
	}

	if teammate.Alive {
		t.Error("Expected teammate to be dead")
	}

	if teammate.Relationship != -10 {
		t.Errorf("Expected relationship -10, got %d", teammate.Relationship)
	}
}

func TestStoryContextFields(t *testing.T) {
	ctx := StoryContext{
		RecentSummary: "The story so far...",
		CurrentScene:  "You are in a dark room...",
		GameBible:     "Core game rules and lore...",
	}

	if ctx.RecentSummary != "The story so far..." {
		t.Errorf("RecentSummary mismatch")
	}

	if ctx.CurrentScene != "You are in a dark room..." {
		t.Errorf("CurrentScene mismatch")
	}
}

func TestItemStructure(t *testing.T) {
	item := Item{
		Name:        "medical_kit",
		Description: "A first aid kit with bandages and antiseptic",
	}

	if item.Name != "medical_kit" {
		t.Errorf("Expected name 'medical_kit', got '%s'", item.Name)
	}

	// Test JSON serialization of Item
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal Item: %v", err)
	}

	var loaded Item
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal Item: %v", err)
	}

	if loaded.Name != item.Name {
		t.Errorf("Item name mismatch after serialization")
	}
}

func TestNewSaveData(t *testing.T) {
	save := NewSaveData()

	if save.Version != CurrentVersion {
		t.Errorf("Expected version %d, got %d", CurrentVersion, save.Version)
	}

	if save.Player.HP != 100 {
		t.Errorf("Expected default HP 100, got %d", save.Player.HP)
	}

	if save.Player.SAN != 100 {
		t.Errorf("Expected default SAN 100, got %d", save.Player.SAN)
	}

	if save.Teammates == nil {
		t.Error("Expected Teammates to be initialized")
	}

	if save.Player.Inventory == nil {
		t.Error("Expected Inventory to be initialized")
	}
}
