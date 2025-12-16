package narration

import (
	"testing"
)

// TestPlanGlobalSeeds 測試 Global Seeds 規劃
func TestPlanGlobalSeeds(t *testing.T) {
	plotStructure := ThreeActStructure{
		Act1: Act{
			Name:      "Setup",
			BeatRange: [2]int{1, 2},
		},
		Act2: Act{
			Name:      "Confrontation",
			BeatRange: [2]int{3, 5},
		},
		Act3: Act{
			Name:      "Resolution",
			BeatRange: [2]int{6, 7},
		},
	}

	endings := []Ending{
		{ID: "true-ending", Name: "True Ending"},
		{ID: "good-ending", Name: "Good Ending"},
		{ID: "bad-ending", Name: "Bad Ending"},
	}

	seeds := PlanGlobalSeeds(plotStructure, endings, "medium")

	// 驗證 Seeds 數量（3-5 個）
	if len(seeds) < 3 || len(seeds) > 5 {
		t.Errorf("Expected 3-5 global seeds, got %d", len(seeds))
	}

	// 驗證每個 Seed
	for i, seed := range seeds {
		// 驗證 ID
		if seed.ID == "" {
			t.Errorf("Seed %d: missing ID", i)
		}

		// 驗證 Content
		if seed.Content == "" {
			t.Errorf("Seed %d: missing content", i)
		}

		// 驗證 ClueChain（必須有 3 層）
		if len(seed.ClueChain) != 3 {
			t.Errorf("Seed %d: expected 3 clues in chain, got %d", i, len(seed.ClueChain))
		}

		// 驗證 ClueChain 的 Tier
		for j, clue := range seed.ClueChain {
			expectedTier := j + 1
			if clue.Tier != expectedTier {
				t.Errorf("Seed %d, Clue %d: expected tier %d, got %d", i, j, expectedTier, clue.Tier)
			}

			if clue.ClueContent == "" {
				t.Errorf("Seed %d, Clue %d: missing clue content", i, j)
			}
		}

		// 驗證 LinkedEnding
		if seed.LinkedEnding == "" {
			t.Errorf("Seed %d: missing linked ending", i)
		}
	}
}

// TestPlanGlobalSeedsDistribution 測試 Seeds 在三幕中的分佈
func TestPlanGlobalSeedsDistribution(t *testing.T) {
	plotStructure := ThreeActStructure{
		Act1: Act{Name: "Setup", BeatRange: [2]int{1, 3}},
		Act2: Act{Name: "Confrontation", BeatRange: [2]int{4, 6}},
		Act3: Act{Name: "Resolution", BeatRange: [2]int{7, 8}},
	}

	endings := []Ending{{ID: "ending-1"}}

	seeds := PlanGlobalSeeds(plotStructure, endings, "medium")

	for i, seed := range seeds {
		// 驗證 Tier 1 線索在 Act1 範圍內
		tier1 := seed.ClueChain[0]
		if tier1.BeatRange[0] < plotStructure.Act1.BeatRange[0] ||
			tier1.BeatRange[1] > plotStructure.Act1.BeatRange[1] {
			t.Errorf("Seed %d: Tier 1 clue beat range %v not in Act1 range %v",
				i, tier1.BeatRange, plotStructure.Act1.BeatRange)
		}

		// 驗證 Tier 2 線索在 Act2 範圍內
		tier2 := seed.ClueChain[1]
		if tier2.BeatRange[0] < plotStructure.Act2.BeatRange[0] ||
			tier2.BeatRange[1] > plotStructure.Act2.BeatRange[1] {
			t.Errorf("Seed %d: Tier 2 clue beat range %v not in Act2 range %v",
				i, tier2.BeatRange, plotStructure.Act2.BeatRange)
		}

		// 驗證 Tier 3 線索在 Act3 範圍內
		tier3 := seed.ClueChain[2]
		if tier3.BeatRange[0] < plotStructure.Act3.BeatRange[0] ||
			tier3.BeatRange[1] > plotStructure.Act3.BeatRange[1] {
			t.Errorf("Seed %d: Tier 3 clue beat range %v not in Act3 range %v",
				i, tier3.BeatRange, plotStructure.Act3.BeatRange)
		}
	}
}

// TestPlanThreeActStructure 測試三幕結構規劃
func TestPlanThreeActStructure(t *testing.T) {
	tests := []struct {
		length        string
		expectedBeats int
		act1Percent   float64
		act2Percent   float64
		act3Percent   float64
	}{
		{"short", 4, 0.25, 0.50, 0.25},
		{"medium", 7, 0.25, 0.50, 0.25},
		{"long", 12, 0.25, 0.50, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.length, func(t *testing.T) {
			structure := PlanThreeActStructure(tt.length)

			totalBeats := structure.Act3.BeatRange[1]

			if totalBeats != tt.expectedBeats {
				t.Errorf("Expected %d beats for %s story, got %d",
					tt.expectedBeats, tt.length, totalBeats)
			}

			// 驗證 Act1
			act1Duration := structure.Act1.BeatRange[1] - structure.Act1.BeatRange[0] + 1
			expectedAct1 := int(float64(totalBeats) * tt.act1Percent)
			if act1Duration < expectedAct1-1 || act1Duration > expectedAct1+1 {
				t.Errorf("Act1 duration %d not close to expected %d (%.0f%%)",
					act1Duration, expectedAct1, tt.act1Percent*100)
			}

			// 驗證連續性
			if structure.Act2.BeatRange[0] != structure.Act1.BeatRange[1]+1 {
				t.Error("Act2 should start right after Act1")
			}

			if structure.Act3.BeatRange[0] != structure.Act2.BeatRange[1]+1 {
				t.Error("Act3 should start right after Act2")
			}

			// 驗證 Goals 存在
			if len(structure.Act1.Goals) == 0 {
				t.Error("Act1 should have goals")
			}
			if len(structure.Act2.Goals) == 0 {
				t.Error("Act2 should have goals")
			}
			if len(structure.Act3.Goals) == 0 {
				t.Error("Act3 should have goals")
			}
		})
	}
}

// TestPlanKeyPlotPoints 測試關鍵劇情點規劃
func TestPlanKeyPlotPoints(t *testing.T) {
	structure := ThreeActStructure{
		Act1: Act{Name: "Setup", BeatRange: [2]int{1, 3}},
		Act2: Act{Name: "Confrontation", BeatRange: [2]int{4, 9}},
		Act3: Act{Name: "Resolution", BeatRange: [2]int{10, 12}},
	}

	plotPoints := PlanKeyPlotPoints(structure)

	// 至少應該有 3-4 個關鍵劇情點
	if len(plotPoints) < 3 {
		t.Errorf("Expected at least 3 plot points, got %d", len(plotPoints))
	}

	// 驗證包含必要的劇情點
	requiredPoints := map[string]bool{
		"Inciting Incident": false,
		"Midpoint":          false,
		"Climax":            false,
	}

	for _, point := range plotPoints {
		if point.Name == "" {
			t.Error("Plot point missing name")
		}
		if point.Beat == 0 {
			t.Error("Plot point missing beat")
		}
		if point.Description == "" {
			t.Error("Plot point missing description")
		}

		if _, exists := requiredPoints[point.Name]; exists {
			requiredPoints[point.Name] = true
		}
	}

	for name, found := range requiredPoints {
		if !found {
			t.Errorf("Missing required plot point: %s", name)
		}
	}
}

// TestPlanEndings 測試多重結局規劃
func TestPlanEndings(t *testing.T) {
	numSeeds := 5

	endings := PlanEndings(numSeeds)

	// 至少應該有 2-3 個結局
	if len(endings) < 2 {
		t.Errorf("Expected at least 2 endings, got %d", len(endings))
	}

	// 驗證結局類型
	hasTrue := false
	hasGood := false
	hasBad := false

	for _, ending := range endings {
		if ending.ID == "" {
			t.Error("Ending missing ID")
		}
		if ending.Name == "" {
			t.Error("Ending missing name")
		}
		if ending.Description == "" {
			t.Error("Ending missing description")
		}

		// 驗證條件
		if ending.Condition.MinSeedPercentage < 0 || ending.Condition.MinSeedPercentage > 1 {
			t.Errorf("Invalid seed percentage: %f", ending.Condition.MinSeedPercentage)
		}

		// 檢查結局類型
		if ending.ID == "true-ending" {
			hasTrue = true
			if ending.Condition.MinSeedPercentage < 0.8 {
				t.Error("True ending should require >= 80% seeds")
			}
		}
		if ending.ID == "good-ending" {
			hasGood = true
		}
		if ending.ID == "bad-ending" {
			hasBad = true
		}
	}

	if !hasTrue {
		t.Error("Missing True Ending")
	}
	if !hasGood {
		t.Error("Missing Good Ending")
	}
	if !hasBad {
		t.Error("Missing Bad Ending")
	}
}
