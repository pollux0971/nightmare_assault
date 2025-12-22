package orchestrator

import (
	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TestEmotionDeltas contains predefined emotion changes for testing.
var TestEmotionDeltas = struct {
	TrustIncrease  manager.EmotionDelta
	TrustDecrease  manager.EmotionDelta
	FearIncrease   manager.EmotionDelta
	StressIncrease manager.EmotionDelta
	Hostile        manager.EmotionDelta
	Friendly       manager.EmotionDelta
}{
	TrustIncrease: manager.EmotionDelta{
		Trust:  10,
		Fear:   0,
		Stress: 0,
	},
	TrustDecrease: manager.EmotionDelta{
		Trust:  -15,
		Fear:   0,
		Stress: 0,
	},
	FearIncrease: manager.EmotionDelta{
		Trust:  0,
		Fear:   20,
		Stress: 10,
	},
	StressIncrease: manager.EmotionDelta{
		Trust:  0,
		Fear:   0,
		Stress: 15,
	},
	Hostile: manager.EmotionDelta{
		Trust:  -20,
		Fear:   5,
		Stress: 10,
	},
	Friendly: manager.EmotionDelta{
		Trust:  15,
		Fear:   -5,
		Stress: -10,
	},
}

// TestFacts contains predefined facts for knowledge propagation testing.
var TestFacts = struct {
	ShelterLocation knowledge.Fact
	MedicalSupplies knowledge.Fact
	DangerousArea   knowledge.Fact
	NPCSecret       knowledge.Fact
}{
	ShelterLocation: knowledge.Fact{
		ID:      "fact_shelter",
		Content: "避難所位於東翼的地下室",
		Source:  "npc_001",
	},
	MedicalSupplies: knowledge.Fact{
		ID:      "fact_medical",
		Content: "三樓藥品室還有醫療物資",
		Source:  "npc_002",
	},
	DangerousArea: knowledge.Fact{
		ID:      "fact_danger",
		Content: "西翼被封鎖了，那裡很危險",
		Source:  "npc_001",
	},
	NPCSecret: knowledge.Fact{
		ID:      "fact_secret",
		Content: "地下室有秘密實驗",
		Source:  "npc_001",
	},
}
