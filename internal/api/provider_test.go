package api

import (
	"testing"
)

func TestBuiltinProviders(t *testing.T) {
	providers := BuiltinProviders()

	if len(providers) == 0 {
		t.Error("Expected at least one builtin provider")
	}

	// Check that we have providers from each category
	categories := make(map[string]int)
	for _, p := range providers {
		categories[p.Category]++
	}

	if categories["official"] == 0 {
		t.Error("Expected at least one official provider")
	}
	if categories["gateway"] == 0 {
		t.Error("Expected at least one gateway provider")
	}
	if categories["local"] == 0 {
		t.Error("Expected at least one local provider")
	}
}

func TestGetProviderInfo(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"openai", "OpenAI"},
		{"anthropic", "Anthropic"},
		{"google", "Google"},
		{"ollama", "Ollama"},
		{"custom", "自訂"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		info := GetProviderInfo(tt.id)
		if tt.expected == "" {
			if info != nil {
				t.Errorf("GetProviderInfo(%s): expected nil, got %v", tt.id, info)
			}
		} else {
			if info == nil {
				t.Errorf("GetProviderInfo(%s): expected info, got nil", tt.id)
			} else if info.Name != tt.expected {
				t.Errorf("GetProviderInfo(%s).Name: expected %s, got %s", tt.id, tt.expected, info.Name)
			}
		}
	}
}

func TestGetProvidersByCategory(t *testing.T) {
	official := GetProvidersByCategory("official")
	if len(official) == 0 {
		t.Error("Expected at least one official provider")
	}

	// Verify all returned providers have correct category
	for _, p := range official {
		if p.Category != "official" {
			t.Errorf("Provider %s has category %s, expected official", p.ID, p.Category)
		}
	}

	local := GetProvidersByCategory("local")
	if len(local) == 0 {
		t.Error("Expected at least one local provider")
	}

	// Check for specific providers
	foundOllama := false
	for _, p := range local {
		if p.ID == "ollama" {
			foundOllama = true
			break
		}
	}
	if !foundOllama {
		t.Error("Expected ollama in local providers")
	}
}

func TestProviderInfoFormat(t *testing.T) {
	// Check that OpenAI-compatible providers have FormatOpenAI
	openaiCompatible := []string{"openai", "mistral", "xai", "deepseek", "groq", "ollama"}
	for _, id := range openaiCompatible {
		info := GetProviderInfo(id)
		if info == nil {
			t.Errorf("Provider %s not found", id)
			continue
		}
		if info.Format != FormatOpenAI {
			t.Errorf("Provider %s: expected format %s, got %s", id, FormatOpenAI, info.Format)
		}
	}

	// Check Anthropic format
	anthropicInfo := GetProviderInfo("anthropic")
	if anthropicInfo == nil {
		t.Error("Anthropic provider not found")
	} else if anthropicInfo.Format != FormatAnthropic {
		t.Errorf("Anthropic: expected format %s, got %s", FormatAnthropic, anthropicInfo.Format)
	}

	// Check Google format
	googleInfo := GetProviderInfo("google")
	if googleInfo == nil {
		t.Error("Google provider not found")
	} else if googleInfo.Format != FormatGoogle {
		t.Errorf("Google: expected format %s, got %s", FormatGoogle, googleInfo.Format)
	}

	// Check Cohere format
	cohereInfo := GetProviderInfo("cohere")
	if cohereInfo == nil {
		t.Error("Cohere provider not found")
	} else if cohereInfo.Format != FormatCohere {
		t.Errorf("Cohere: expected format %s, got %s", FormatCohere, cohereInfo.Format)
	}
}
