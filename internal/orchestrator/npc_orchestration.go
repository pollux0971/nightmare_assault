package orchestrator

import (
	"context"
	"fmt"
	"log"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.6: Real NPC Orchestration Implementation
// ==========================================================================

// RealNPCAgent wraps the agents.NPCAgent and provides orchestrator-level NPC generation
// This implementation generates complete NPCs with Show-Don't-Tell introductions
type RealNPCAgent struct {
	npcAgent *agents.NPCAgent
}

// NewRealNPCAgent creates a new RealNPCAgent
func NewRealNPCAgent(config agents.AgentConfig) *RealNPCAgent {
	return &RealNPCAgent{
		npcAgent: agents.NewNPCAgent(config),
	}
}

// GenerateProfiles generates NPC profiles with Show-Don't-Tell introductions
//
// Story 7.6 Implementation:
//   - Generates 2-4 NPCs based on difficulty (req.Count)
//   - Each NPC includes: Name, Skills, Inventory, Secret, Introduction
//   - Introduction follows Show-Don't-Tell principle
func (r *RealNPCAgent) GenerateProfiles(ctx context.Context, req NPCRequest) ([]*NPCProfile, error) {
	log.Printf("[RealNPCAgent] Generating %d NPCs for theme: %s", req.Count, req.Skeleton.MainTheme)

	profiles := make([]*NPCProfile, 0, req.Count)

	// Archetype rotation for diversity
	// Story 7.6: Generate 2-4 teammates (not all archetypes, focus on helpful ones)
	archetypes := []agents.NPCArchetype{
		agents.NPCArchetypeGuide,         // N-05: Helpful, provides direction
		agents.NPCArchetypeKnowledgeable, // N-02: Provides clues
		agents.NPCArchetypeNeutral,       // N-04: Provides contrast
		agents.NPCArchetypeSacrificial,   // N-01: Can die for drama
	}

	// Create story context from skeleton
	storyContext := agents.StoryContext{
		Theme: req.Skeleton.MainTheme,
		Scene: req.Skeleton.WorldView,
	}

	// Generate each NPC
	for i := 0; i < req.Count; i++ {
		// Select archetype (cycle through available ones)
		archetype := archetypes[i%len(archetypes)]

		// Step 1: Generate NPC instance
		generateReq := &agents.GenerateRequest{
			Archetype:    archetype,
			StoryContext: storyContext,
			GlobalSeeds:  []agents.GlobalSeedInfo{}, // Empty for now
			PlotStructure: agents.PlotStructure{
				TotalBeats: 20,
				Act1Range:  [2]int{1, 5},
				Act2Range:  [2]int{6, 15},
				Act3Range:  [2]int{16, 20},
			},
		}

		generateResp, err := r.npcAgent.InvokeGenerate(ctx, generateReq)
		if err != nil {
			log.Printf("[RealNPCAgent] Failed to generate NPC %d: %v", i+1, err)
			// Continue with next NPC instead of failing completely
			continue
		}

		npcInstance := generateResp.NPC

		// Step 2: Generate Show-Don't-Tell introduction
		introReq := &agents.IntroductionRequest{
			NPC:          npcInstance,
			StoryContext: storyContext,
		}

		introResp, err := r.npcAgent.InvokeIntroduction(ctx, introReq)
		if err != nil {
			log.Printf("[RealNPCAgent] Failed to generate introduction for NPC %d: %v", i+1, err)
			// Use a default introduction if generation fails
			introResp = &agents.IntroductionResponse{
				Introduction: fmt.Sprintf("%s站在那裡，默默地觀察著周圍的一切。", npcInstance.Name),
			}
		}

		// Store introduction in NPC instance
		npcInstance.Introduction = introResp.Introduction

		// Step 3: Convert to orchestrator NPCProfile
		profile := &NPCProfile{
			ID:          npcInstance.ID,
			Name:        npcInstance.Name,
			Description: npcInstance.Introduction, // Use Show-Don't-Tell introduction as description
		}

		profiles = append(profiles, profile)

		log.Printf("[RealNPCAgent] Generated NPC %d/%d: %s (%s)", i+1, req.Count, npcInstance.Name, archetype)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("failed to generate any NPCs")
	}

	log.Printf("[RealNPCAgent] Successfully generated %d NPCs", len(profiles))
	return profiles, nil
}

// GetNPCCountForDifficulty returns the number of NPCs to generate based on difficulty
//
// Story 7.6 AC #1: Phase 1 生成 2-4 名隊友
//   - Easy: 2 NPCs (simpler management)
//   - Normal: 3 NPCs (balanced)
//   - Hard: 3 NPCs (balanced)
//   - Hell: 4 NPCs (more complexity, more potential betrayal/death)
func GetNPCCountForDifficulty(difficulty string) int {
	switch difficulty {
	case "easy":
		return 2
	case "normal":
		return 3
	case "hard":
		return 3
	case "hell":
		return 4
	default:
		return 3 // Default to normal
	}
}
