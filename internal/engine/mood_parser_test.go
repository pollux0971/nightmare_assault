package engine

import (
	"testing"
)

func TestParseMood_ExplorationMood(t *testing.T) {
	text := "你推開沉重的木門，走進一個黑暗的走廊。[MOOD:exploration]"
	mood := ParseMood(text)

	if mood != MoodExploration {
		t.Errorf("ParseMood() = %v, expected MoodExploration", mood)
	}
}

func TestParseMood_TensionMood(t *testing.T) {
	text := "遠處傳來急促的腳步聲，越來越近！[MOOD:tension]"
	mood := ParseMood(text)

	if mood != MoodTension {
		t.Errorf("ParseMood() = %v, expected MoodTension", mood)
	}
}

func TestParseMood_SafeMood(t *testing.T) {
	text := "你終於找到了一個安全的房間。[MOOD:safe]"
	mood := ParseMood(text)

	if mood != MoodSafe {
		t.Errorf("ParseMood() = %v, expected MoodSafe", mood)
	}
}

func TestParseMood_HorrorMood(t *testing.T) {
	text := "門後站著一個恐怖的身影！[MOOD:horror]"
	mood := ParseMood(text)

	if mood != MoodHorror {
		t.Errorf("ParseMood() = %v, expected MoodHorror", mood)
	}
}

func TestParseMood_MysteryMood(t *testing.T) {
	text := "這個謎題需要仔細思考。[MOOD:mystery]"
	mood := ParseMood(text)

	if mood != MoodMystery {
		t.Errorf("ParseMood() = %v, expected MoodMystery", mood)
	}
}

func TestParseMood_EndingMood(t *testing.T) {
	text := "你的生命走到了盡頭。[MOOD:ending]"
	mood := ParseMood(text)

	if mood != MoodEnding {
		t.Errorf("ParseMood() = %v, expected MoodEnding", mood)
	}
}

func TestParseMood_NoMoodTag(t *testing.T) {
	text := "這段文字沒有 mood 標記。"
	mood := ParseMood(text)

	if mood != MoodExploration {
		t.Errorf("ParseMood() = %v, expected default MoodExploration", mood)
	}
}

func TestParseMood_UnknownMood(t *testing.T) {
	text := "未知的 mood 類型。[MOOD:unknown]"
	mood := ParseMood(text)

	if mood != MoodExploration {
		t.Errorf("ParseMood() = %v, expected default MoodExploration for unknown mood", mood)
	}
}

func TestParseMood_CaseInsensitive(t *testing.T) {
	tests := []struct {
		text     string
		expected MoodType
	}{
		{"[MOOD:TENSION]", MoodTension},
		{"[MOOD:Tension]", MoodTension},
		{"[MOOD:TeNsIoN]", MoodTension},
		{"[MOOD:SAFE]", MoodSafe},
		{"[MOOD:Safe]", MoodSafe},
	}

	for _, tt := range tests {
		mood := ParseMood(tt.text)
		if mood != tt.expected {
			t.Errorf("ParseMood(%q) = %v, expected %v", tt.text, mood, tt.expected)
		}
	}
}

func TestParseMood_MultipleTagsUsesFirst(t *testing.T) {
	text := "[MOOD:tension] 一些文字 [MOOD:safe]"
	mood := ParseMood(text)

	if mood != MoodTension {
		t.Errorf("ParseMood() = %v, expected MoodTension (first tag)", mood)
	}
}

func TestParseMood_TagInMiddleOfText(t *testing.T) {
	text := "前半部分文字 [MOOD:horror] 後半部分文字"
	mood := ParseMood(text)

	if mood != MoodHorror {
		t.Errorf("ParseMood() = %v, expected MoodHorror", mood)
	}
}

func TestParseMood_EmptyString(t *testing.T) {
	text := ""
	mood := ParseMood(text)

	if mood != MoodExploration {
		t.Errorf("ParseMood() = %v, expected default MoodExploration for empty string", mood)
	}
}

func TestMoodType_String(t *testing.T) {
	tests := []struct {
		mood     MoodType
		expected string
	}{
		{MoodExploration, "exploration"},
		{MoodTension, "tension"},
		{MoodSafe, "safe"},
		{MoodHorror, "horror"},
		{MoodMystery, "mystery"},
		{MoodEnding, "ending"},
	}

	for _, tt := range tests {
		result := tt.mood.String()
		if result != tt.expected {
			t.Errorf("MoodType.String() = %q, expected %q", result, tt.expected)
		}
	}
}
