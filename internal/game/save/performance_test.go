package save

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSavePerformance500ms(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Create a reasonably large save data
	gameState := createLargeSaveData()

	start := time.Now()
	err := manager.Save(1, gameState)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// NFR01: Save operation must complete within 500ms
	if elapsed > 500*time.Millisecond {
		t.Errorf("Save operation took %v, expected < 500ms", elapsed)
	}

	t.Logf("Save completed in %v", elapsed)
}

func TestLoadPerformance500ms(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Create and save large data first
	gameState := createLargeSaveData()
	if err := manager.Save(1, gameState); err != nil {
		t.Fatalf("Setup save failed: %v", err)
	}

	start := time.Now()
	loaded, err := manager.Load(1)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("Loaded data is nil")
	}

	// NFR01: Load operation must complete within 500ms
	if elapsed > 500*time.Millisecond {
		t.Errorf("Load operation took %v, expected < 500ms", elapsed)
	}

	t.Logf("Load completed in %v", elapsed)
}

func TestConcurrentSaveSafety(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	const numGoroutines = 5
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Try to save concurrently to the same slot
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			gameState := NewSaveData()
			gameState.Player.HP = iteration * 10
			if err := manager.Save(1, gameState); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent save error: %v", err)
	}

	// Verify final state is consistent
	loaded, err := manager.Load(1)
	if err != nil {
		t.Fatalf("Load after concurrent saves failed: %v", err)
	}

	// HP should be one of the values (0, 10, 20, 30, 40)
	validHPs := map[int]bool{0: true, 10: true, 20: true, 30: true, 40: true}
	if !validHPs[loaded.Player.HP] {
		t.Errorf("Unexpected HP value after concurrent saves: %d", loaded.Player.HP)
	}
}

func TestConcurrentLoadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Create save first
	gameState := NewSaveData()
	gameState.Player.HP = 100
	if err := manager.Save(1, gameState); err != nil {
		t.Fatalf("Setup save failed: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make(chan *SaveData, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Try to load concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loaded, err := manager.Load(1)
			if err != nil {
				errors <- err
				return
			}
			results <- loaded
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent load error: %v", err)
	}

	// All loads should return consistent data
	for loaded := range results {
		if loaded.Player.HP != 100 {
			t.Errorf("Inconsistent load: expected HP 100, got %d", loaded.Player.HP)
		}
	}
}

func TestLargeSaveDataSize(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Create large save data
	gameState := createLargeSaveData()

	// Add extra context to push size
	gameState.Context.GameBible = string(make([]byte, 100*1024)) // 100KB of context

	err := manager.Save(1, gameState)
	if err != nil {
		t.Fatalf("Large save failed: %v", err)
	}

	// Load should still work
	loaded, err := manager.Load(1)
	if err != nil {
		t.Fatalf("Load large save failed: %v", err)
	}

	if len(loaded.Context.GameBible) != 100*1024 {
		t.Errorf("GameBible size mismatch after load")
	}
}

func BenchmarkSave(b *testing.B) {
	tmpDir := b.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)
	gameState := createLargeSaveData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Save(1, gameState)
	}
}

func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")
	manager := NewSaveManager(saveDir)

	// Setup: create save file
	gameState := createLargeSaveData()
	manager.Save(1, gameState)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Load(1)
	}
}

// Helper function to create reasonably large save data
func createLargeSaveData() *SaveData {
	save := NewSaveData()
	save.Metadata.PlayTime = 36000 // 10 hours
	save.Game.CurrentChapter = 5
	save.Game.ChapterProgress = 0.75

	// Add inventory items
	for i := 0; i < 20; i++ {
		save.Player.Inventory = append(save.Player.Inventory, Item{
			Name:        "item_" + string(rune('A'+i)),
			Description: "A mysterious item found in the darkness...",
		})
	}

	// Add known clues
	for i := 0; i < 50; i++ {
		save.Player.KnownClues = append(save.Player.KnownClues, "clue_discovered_in_chapter_"+string(rune('0'+i%10)))
	}

	// Add triggered rules
	for i := 0; i < 30; i++ {
		save.Game.TriggeredRules = append(save.Game.TriggeredRules, "rule_"+string(rune('0'+i/10))+string(rune('0'+i%10)))
	}

	// Add teammates
	teammates := []string{"Alice", "Bob", "Charlie", "David"}
	for _, name := range teammates {
		save.Teammates = append(save.Teammates, TeammateState{
			Name:         name,
			Alive:        true,
			HP:           80,
			Location:     "basement",
			Relationship: 50,
			Items:        []Item{{Name: "flashlight", Description: "A dim light in the darkness"}},
		})
	}

	// Add story context
	save.Context.RecentSummary = "The group has been exploring the abandoned hospital for hours. " +
		"Strange noises echo through the corridors. Alice found a mysterious key. " +
		"Bob discovered bloodstains on the walls. The tension is palpable."
	save.Context.CurrentScene = "You stand in a dimly lit corridor. The walls are covered in peeling wallpaper. " +
		"A faint light flickers at the end of the hallway. You hear footsteps behind you."
	save.Context.GameBible = "This is a horror game set in an abandoned hospital. The players must uncover the truth " +
		"while avoiding deadly traps and supernatural entities. Hidden rules govern the game world."

	return save
}
