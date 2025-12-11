package engine

import (
	"context"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// MockSmartModelClient for testing
type MockSmartModelClient struct {
	response string
	err      error
}

func (m *MockSmartModelClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestGenerateOpeningDream(t *testing.T) {
	client := &MockSmartModelClient{
		response: "你站在一面巨大的鏡子前，鏡中的你卻做著相反的動作。當你伸出右手時，鏡中的你伸出了左手。你試圖靠近，但鏡子卻開始碎裂。每一片碎片中都映照著不同的你，每個你都在做著不同的事情。你聽到身後有腳步聲，回頭一看，卻什麼都沒有。當你再次轉向鏡子時，所有的你都消失了，只剩下一片漆黑。你感到一陣寒意，彷彿有什麼東西正在注視著你...",
	}

	dg := NewDreamGenerator(client)

	content, err := dg.GenerateOpeningDream(
		context.Background(),
		"廢棄醫院探險",
		"不能直視鏡子、必須保持安靜",
		"探險者",
	)

	if err != nil {
		t.Fatalf("GenerateOpeningDream() error = %v", err)
	}

	if len(content) < 200 {
		t.Errorf("Dream content too short: %d characters", len(content))
	}

	if content == "" {
		t.Error("Dream content should not be empty")
	}
}

func TestCreateDreamRecord(t *testing.T) {
	client := &MockSmartModelClient{}
	dg := NewDreamGenerator(client)

	context := game.DreamContext{
		PlayerHP:   100,
		PlayerSAN:  80,
		ChapterNum: 1,
		StoryTheme: "廢棄醫院",
	}

	record := dg.CreateDreamRecord(
		game.DreamTypeOpening,
		"Test dream content",
		"rule-1",
		context,
	)

	if record.Type != game.DreamTypeOpening {
		t.Errorf("Expected type %v, got %v", game.DreamTypeOpening, record.Type)
	}

	if record.Content != "Test dream content" {
		t.Errorf("Expected content 'Test dream content', got '%s'", record.Content)
	}

	if record.RelatedRuleID != "rule-1" {
		t.Errorf("Expected rule ID 'rule-1', got '%s'", record.RelatedRuleID)
	}

	if record.Context.ChapterNum != 1 {
		t.Errorf("Expected chapter 1, got %d", record.Context.ChapterNum)
	}
}

func TestGenerateChapterDream(t *testing.T) {
	client := &MockSmartModelClient{
		response: "你又夢到了那條走廊。這次，牆上的裂痕變得更大了，從中滲出暗紅色的液體。你聽到遠處傳來哭泣聲，但當你走近時，聲音卻突然停止了。",
	}

	dg := NewDreamGenerator(client)

	chapterContext := ChapterDreamContext{
		ChapterNum:   2,
		RecentEvents: "玩家進入了醫院的地下室",
		RuleHints:    "地下室有危險",
		PlayerSAN:    50,
		KnownClues:   []string{"血跡", "腳印"},
		HighStress:   true,
	}

	content, err := dg.GenerateChapterDream(
		context.Background(),
		DreamTypeNightmare,
		chapterContext,
	)

	if err != nil {
		t.Fatalf("GenerateChapterDream() error = %v", err)
	}

	if content == "" {
		t.Error("Chapter dream content should not be empty")
	}
}

func TestDetermineDreamProbability(t *testing.T) {
	tests := []struct {
		name             string
		context          ChapterDreamContext
		expectedType     ChapterDreamType
		expectedProbMin  float64
	}{
		{
			name: "High stress triggers nightmare",
			context: ChapterDreamContext{
				HighStress: true,
			},
			expectedType:    DreamTypeNightmare,
			expectedProbMin: 0.70,
		},
		{
			name: "Teammate death triggers grief",
			context: ChapterDreamContext{
				TeammateDeaths: true,
			},
			expectedType:    DreamTypeGrief,
			expectedProbMin: 0.80,
		},
		{
			name: "Low SAN triggers warning",
			context: ChapterDreamContext{
				PlayerSAN: 25,
			},
			expectedType:    DreamTypeWarning,
			expectedProbMin: 0.60,
		},
		{
			name: "Recent clue triggers hint",
			context: ChapterDreamContext{
				RecentClue: true,
				PlayerSAN:  50,
			},
			expectedType:    DreamTypeHint,
			expectedProbMin: 0.50,
		},
		{
			name: "Normal situation triggers random",
			context: ChapterDreamContext{
				PlayerSAN: 80,
			},
			expectedType:    DreamTypeRandom,
			expectedProbMin: 0.20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dreamType, probability := DetermineDreamProbability(tt.context)

			if dreamType != tt.expectedType {
				t.Errorf("Expected dream type %v, got %v", tt.expectedType, dreamType)
			}

			if probability < tt.expectedProbMin {
				t.Errorf("Expected probability >= %v, got %v", tt.expectedProbMin, probability)
			}
		})
	}
}

func TestGenerateAllDreamTypes(t *testing.T) {
	dreamTypes := []ChapterDreamType{
		DreamTypeNightmare,
		DreamTypeHint,
		DreamTypeGrief,
		DreamTypeWarning,
		DreamTypeRandom,
	}

	client := &MockSmartModelClient{
		response: "這是一段測試夢境內容，包含了足夠的文字來滿足最小長度要求。夢境中你看到了奇怪的景象，這些景象似乎在告訴你什麼重要的事情。",
	}

	dg := NewDreamGenerator(client)

	chapterContext := ChapterDreamContext{
		ChapterNum:    2,
		RecentEvents:  "玩家探索地下室",
		RuleHints:     "小心黑暗",
		PlayerSAN:     40,
		KnownClues:    []string{"血跡"},
		DeadTeammates: []string{"張三"},
	}

	for _, dreamType := range dreamTypes {
		t.Run(string(dreamType), func(t *testing.T) {
			content, err := dg.GenerateChapterDream(
				context.Background(),
				dreamType,
				chapterContext,
			)

			if err != nil {
				t.Fatalf("GenerateChapterDream(%v) error = %v", dreamType, err)
			}

			if content == "" {
				t.Errorf("Dream content should not be empty for type %v", dreamType)
			}

			if len(content) < 50 {
				t.Errorf("Dream content too short for type %v: %d characters", dreamType, len(content))
			}
		})
	}
}
