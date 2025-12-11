package engine

import (
	"testing"
)

func TestFallbackStories_NotEmpty(t *testing.T) {
	if len(FallbackStories) == 0 {
		t.Fatal("FallbackStories should not be empty")
	}
}

func TestFallbackStories_HaveContent(t *testing.T) {
	for i, story := range FallbackStories {
		if story.Theme == "" {
			t.Errorf("Story %d has empty theme", i)
		}
		if story.Content == "" {
			t.Errorf("Story %d (%s) has empty content", i, story.Theme)
		}
		if len(story.Choices) == 0 {
			t.Errorf("Story %d (%s) has no choices", i, story.Theme)
		}
		if len(story.Choices) < 2 {
			t.Errorf("Story %d (%s) has fewer than 2 choices", i, story.Theme)
		}
	}
}

func TestFallbackStories_HasDefault(t *testing.T) {
	hasDefault := false
	for _, story := range FallbackStories {
		if story.Theme == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Error("FallbackStories should have a 'default' theme")
	}
}

func TestGetFallbackStory_MatchesTheme(t *testing.T) {
	tests := []struct {
		theme    string
		expected string
	}{
		{"hospital", "hospital"},
		{"abandoned hospital", "hospital"},
		{"haunted mansion", "mansion"},
		{"subway station", "subway"},
		{"dark forest", "forest"},
	}

	for _, tt := range tests {
		story := GetFallbackStory(tt.theme)
		if story == nil {
			t.Errorf("GetFallbackStory(%q) returned nil", tt.theme)
			continue
		}
		if story.Theme != tt.expected {
			t.Errorf("GetFallbackStory(%q) theme = %q, want %q", tt.theme, story.Theme, tt.expected)
		}
	}
}

func TestGetFallbackStory_ChineseKeywords(t *testing.T) {
	tests := []struct {
		theme    string
		expected string
	}{
		{"廢棄醫院", "hospital"},
		{"詛咒洋館", "mansion"},
		{"末日地鐵站", "subway"},
		{"黑暗森林", "forest"},
	}

	for _, tt := range tests {
		story := GetFallbackStory(tt.theme)
		if story == nil {
			t.Errorf("GetFallbackStory(%q) returned nil", tt.theme)
			continue
		}
		if story.Theme != tt.expected {
			t.Errorf("GetFallbackStory(%q) theme = %q, want %q", tt.theme, story.Theme, tt.expected)
		}
	}
}

func TestGetFallbackStory_UnknownTheme(t *testing.T) {
	story := GetFallbackStory("completely random theme xyz")

	// Should return some story (either random or default)
	if story == nil {
		t.Error("GetFallbackStory should never return nil")
	}

	// Should have content
	if story.Content == "" {
		t.Error("Returned story should have content")
	}
}

func TestGetFallbackStory_EmptyTheme(t *testing.T) {
	story := GetFallbackStory("")

	if story == nil {
		t.Error("GetFallbackStory('') should not return nil")
	}
}

func TestContainsRune(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"你好世界", "好", true},
		{"你好世界", "世界", true},
		{"hello", "ell", true},
		{"你好", "世界", false},
		{"", "a", false},
		{"a", "", true},
	}

	for _, tt := range tests {
		if got := containsRune(tt.s, tt.substr); got != tt.expected {
			t.Errorf("containsRune(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.expected)
		}
	}
}

func TestFindRuneSubstring(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected int
	}{
		{"你好世界", "好", 1},
		{"你好世界", "世界", 2},
		{"hello", "ell", 1},
		{"你好", "世界", -1},
	}

	for _, tt := range tests {
		if got := findRuneSubstring(tt.s, tt.substr); got != tt.expected {
			t.Errorf("findRuneSubstring(%q, %q) = %d, want %d", tt.s, tt.substr, got, tt.expected)
		}
	}
}

func TestFallbackStory_ContentQuality(t *testing.T) {
	for _, story := range FallbackStories {
		// Content should be reasonable length
		if len(story.Content) < 100 {
			t.Errorf("Story %s content too short (%d chars)", story.Theme, len(story.Content))
		}

		// Choices should have reasonable length
		for i, choice := range story.Choices {
			if len(choice) < 3 {
				t.Errorf("Story %s choice %d too short: %q", story.Theme, i, choice)
			}
		}
	}
}
