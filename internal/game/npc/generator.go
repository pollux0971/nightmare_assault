package npc

import (
	"fmt"
	"math/rand/v2"
)

// ArchetypeTemplate defines the template for an NPC archetype
type ArchetypeTemplate struct {
	CoreTraits       []string
	SkillSuggestions []string
	BehaviorPatterns []string
	SpeechStyle      string
	FearResponse     string
}

// GetArchetypeTemplate returns the template for a given archetype
func GetArchetypeTemplate(archetype NPCArchetype) ArchetypeTemplate {
	templates := map[NPCArchetype]ArchetypeTemplate{
		ArchetypeVictim: {
			CoreTraits:       []string{"fearful", "dependent", "emotional"},
			SkillSuggestions: []string{"crying", "screaming", "hiding"},
			BehaviorPatterns: []string{"panics easily", "seeks protection", "freezes in danger"},
			SpeechStyle:      "trembling, high-pitched, frequent pauses",
			FearResponse:     "immediate panic, crying, paralysis",
		},
		ArchetypeUnreliable: {
			CoreTraits:       []string{"secretive", "unpredictable", "suspicious"},
			SkillSuggestions: []string{"deception", "hiding information", "misdirection"},
			BehaviorPatterns: []string{"lies about whereabouts", "hides discoveries", "acts strange"},
			SpeechStyle:      "evasive, vague, contradictory",
			FearResponse:     "denial, deflection, hiding",
		},
		ArchetypeLogic: {
			CoreTraits:       []string{"analytical", "calm", "rational"},
			SkillSuggestions: []string{"analysis", "pattern recognition", "deduction"},
			BehaviorPatterns: []string{"examines evidence", "proposes theories", "stays calm"},
			SpeechStyle:      "clear, measured, technical",
			FearResponse:     "rationalization, systematic thinking",
		},
		ArchetypeIntuition: {
			CoreTraits:       []string{"perceptive", "sensitive", "wary"},
			SkillSuggestions: []string{"danger sense", "reading atmosphere", "premonition"},
			BehaviorPatterns: []string{"senses danger early", "trusts gut feelings", "warns others"},
			SpeechStyle:      "uncertain, feeling-based, warning",
			FearResponse:     "heightened awareness, flight instinct",
		},
		ArchetypeInformer: {
			CoreTraits:       []string{"knowledgeable", "helpful", "expositor"},
			SkillSuggestions: []string{"lore knowledge", "clue provision", "background info"},
			BehaviorPatterns: []string{"shares background", "explains history", "provides context"},
			SpeechStyle:      "informative, storytelling, detailed",
			FearResponse:     "reveals more information under stress",
		},
		ArchetypePossessed: {
			CoreTraits:       []string{"influenced", "unstable", "dangerous"},
			SkillSuggestions: []string{"manipulation", "betrayal", "violence"},
			BehaviorPatterns: []string{"acts against group", "shows signs of possession", "becomes hostile"},
			SpeechStyle:      "shifting, occasionally inhuman, threatening",
			FearResponse:     "aggression, possession manifestation",
		},
	}

	return templates[archetype]
}

// GenerateTeammates generates teammates based on story length and difficulty
func GenerateTeammates(storyLength, difficulty string) []*Teammate {
	var count int

	// Determine teammate count based on story length
	switch storyLength {
	case "short":
		count = 1
	case "medium", "long":
		count = 2 + rand.IntN(2) // 2-3 teammates
	default:
		count = 2
	}

	teammates := make([]*Teammate, 0, count)
	usedArchetypes := make(map[NPCArchetype]bool)

	// Available archetypes
	archetypes := []NPCArchetype{
		ArchetypeVictim,
		ArchetypeUnreliable,
		ArchetypeLogic,
		ArchetypeIntuition,
		ArchetypeInformer,
		ArchetypePossessed,
	}

	// Generate teammates with diverse archetypes
	for i := 0; i < count; i++ {
		// Select archetype (prefer unused ones)
		var selectedArchetype NPCArchetype
		attempts := 0
		for {
			selectedArchetype = archetypes[rand.IntN(len(archetypes))]
			// Allow duplicate only if we've tried many times or all archetypes used
			if !usedArchetypes[selectedArchetype] || attempts > 10 || len(usedArchetypes) == len(archetypes) {
				break
			}
			attempts++
		}
		usedArchetypes[selectedArchetype] = true

		// Generate teammate
		id := fmt.Sprintf("tm-%03d", i+1)
		name := generateName(i)
		template := GetArchetypeTemplate(selectedArchetype)

		teammate := NewTeammate(id, name, selectedArchetype)
		teammate.Personality = PersonalityTraits{
			CoreTraits:       template.CoreTraits,
			BehaviorPatterns: template.BehaviorPatterns,
			SpeechStyle:      template.SpeechStyle,
			FearResponse:     template.FearResponse,
		}
		teammate.Skills = template.SkillSuggestions
		teammate.Background = generateBackground(selectedArchetype)

		teammates = append(teammates, teammate)
	}

	return teammates
}

// generateName generates a placeholder name for a teammate
func generateName(index int) string {
	names := []string{
		"李明", "王芳", "張偉", "劉娜", "陳強",
		"楊靜", "趙軍", "周敏", "吳剛", "鄭婷",
	}
	if index < len(names) {
		return names[index]
	}
	return fmt.Sprintf("隊友-%d", index+1)
}

// generateBackground generates a placeholder background for an archetype
func generateBackground(archetype NPCArchetype) string {
	backgrounds := map[NPCArchetype]string{
		ArchetypeVictim:     "容易受到驚嚇的普通人，缺乏應對危機的經驗",
		ArchetypeUnreliable: "行為詭異，似乎隱藏著某些秘密",
		ArchetypeLogic:      "理性冷靜，善於分析和推理",
		ArchetypeIntuition:  "直覺敏銳，能感知到危險的存在",
		ArchetypeInformer:   "對這個地方的歷史有所了解",
		ArchetypePossessed:  "行為越來越不正常，似乎受到某種影響",
	}
	return backgrounds[archetype]
}
