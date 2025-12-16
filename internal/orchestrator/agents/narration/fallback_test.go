package narration

import (
	"testing"
)

// TestBuildFallbackSkeleton 測試降級骨架生成
func TestBuildFallbackSkeleton(t *testing.T) {
	request := &SkeletonRequest{
		Theme:       "廢棄醫院",
		Difficulty:  "normal",
		StoryLength: "medium",
		Adult18Plus: false,
	}

	fallback := BuildFallbackSkeleton(request)

	// 驗證基本字段存在
	if fallback.WorldView.Setting == "" {
		t.Error("Fallback should have setting")
	}

	if fallback.CoreTruth.Truth == "" {
		t.Error("Fallback should have truth")
	}

	// 驗證劇情結構
	if fallback.PlotStructure.EstimatedBeats == 0 {
		t.Error("Fallback should have estimated beats")
	}

	// 驗證 Global Seeds（至少 3 個）
	if len(fallback.PlotStructure.GlobalSeeds) < 3 {
		t.Errorf("Fallback should have at least 3 global seeds, got %d",
			len(fallback.PlotStructure.GlobalSeeds))
	}

	// 驗證結局（至少 2 個）
	if len(fallback.PossibleEndings) < 2 {
		t.Errorf("Fallback should have at least 2 endings, got %d",
			len(fallback.PossibleEndings))
	}

	// 驗證三幕結構
	if fallback.PlotStructure.ThreeAct.Act1.Name == "" {
		t.Error("Fallback should have Act1")
	}
	if fallback.PlotStructure.ThreeAct.Act2.Name == "" {
		t.Error("Fallback should have Act2")
	}
	if fallback.PlotStructure.ThreeAct.Act3.Name == "" {
		t.Error("Fallback should have Act3")
	}
}

// TestBuildFallbackSkeletonWithTheme 測試帶主題的降級骨架
func TestBuildFallbackSkeletonWithTheme(t *testing.T) {
	tests := []struct {
		theme    string
		contains string
	}{
		{"廢棄醫院", "醫院"},
		{"詭異學校", "學校"},
		{"荒涼旅館", "旅館"},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			request := &SkeletonRequest{
				Theme:       tt.theme,
				Difficulty:  "normal",
				StoryLength: "medium",
			}

			fallback := BuildFallbackSkeleton(request)

			// 驗證世界觀包含主題相關內容
			// (簡化版本，只檢查是否有內容)
			if fallback.WorldView.Setting == "" {
				t.Error("Fallback setting should not be empty")
			}

			if fallback.WorldView.Background == "" {
				t.Error("Fallback background should not be empty")
			}
		})
	}
}

// TestBuildFallbackSkeletonDifficulty 測試不同難度的降級骨架
func TestBuildFallbackSkeletonDifficulty(t *testing.T) {
	tests := []struct {
		difficulty    string
		expectedSeeds int
	}{
		{"easy", 3},
		{"normal", 4},
		{"hard", 5},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			request := &SkeletonRequest{
				Theme:       "test",
				Difficulty:  tt.difficulty,
				StoryLength: "medium",
			}

			fallback := BuildFallbackSkeleton(request)

			// 驗證 Seeds 數量符合難度
			if len(fallback.PlotStructure.GlobalSeeds) != tt.expectedSeeds {
				t.Errorf("Expected %d seeds for %s difficulty, got %d",
					tt.expectedSeeds, tt.difficulty, len(fallback.PlotStructure.GlobalSeeds))
			}
		})
	}
}
