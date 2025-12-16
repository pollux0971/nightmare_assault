package parallel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

// PreloadCache stores pregenerated results from speculative preloading
// This is populated by game_setup.go when user is on confirmation screen
var globalPreloadCache *PreloadCache

// PreloadCache stores fast-generated content
type PreloadCache struct {
	rules      *rules.RuleSet
	teammates  []*npc.Teammate
	configHash string
	generated  time.Time
}

// SetPreloadCache sets the global preload cache (called by game_setup.go)
func SetPreloadCache(cache *PreloadCache) {
	globalPreloadCache = cache
}

// GetPreloadCache returns the current preload cache
func GetPreloadCache() *PreloadCache {
	return globalPreloadCache
}

// ClearPreloadCache clears the global preload cache
func ClearPreloadCache() {
	globalPreloadCache = nil
}

// BuildGenerationTasks creates all generation tasks for game startup
func BuildGenerationTasks(
	gameConfig *game.GameConfig,
	smartProvider api.Provider,
) ([]*Task, error) {
	tasks := []*Task{}

	// Check if we have preloaded content
	configHash := gameConfig.Hash()
	var preloadedRules *rules.RuleSet
	var preloadedTeammates []*npc.Teammate

	if globalPreloadCache != nil && globalPreloadCache.configHash == configHash {
		preloadedRules = globalPreloadCache.rules
		preloadedTeammates = globalPreloadCache.teammates
	}

	// Task 1: Rules Generation (fast, no dependencies)
	if preloadedRules != nil {
		// Use preloaded results (instant)
		rulesTask := NewTask(TaskIDRules, "生成遊戲規則", PriorityHigh, 1*time.Second, func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			return preloadedRules, nil
		})
		tasks = append(tasks, rulesTask)
	} else {
		// Generate dynamically
		rulesTask := NewTask(TaskIDRules, "生成遊戲規則", PriorityHigh, 5*time.Second, func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			generator := rules.NewGenerator()
			ruleSet := generator.GenerateRules(gameConfig.Difficulty)
			return ruleSet, nil
		})
		tasks = append(tasks, rulesTask)
	}

	// Task 2: Teammates Generation (fast, no dependencies)
	if preloadedTeammates != nil {
		// Use preloaded results (instant)
		teammatesTask := NewTask(TaskIDTeammates, "生成隊友", PriorityLow, 1*time.Second, func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			return preloadedTeammates, nil
		})
		tasks = append(tasks, teammatesTask)
	} else {
		// Generate dynamically
		teammatesTask := NewTask(TaskIDTeammates, "生成隊友", PriorityLow, 5*time.Second, func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			length := gameConfig.Length.String()
			difficulty := gameConfig.Difficulty.String()
			teammates := npc.GenerateTeammates(length, difficulty)
			return teammates, nil
		})
		tasks = append(tasks, teammatesTask)
	}

	// Task 3: Opening Story (slow, depends on Rules + Teammates)
	storyTask := NewTask(TaskIDStory, "生成開場故事", PriorityCritical, 45*time.Second, func(ctx context.Context, deps DependencyResults) (interface{}, error) {
		// Get dependencies (not used in current implementation, but needed for dependency ordering)
		_, ok := deps[TaskIDRules].(*rules.RuleSet)
		if !ok {
			return nil, fmt.Errorf("rules dependency not found or invalid type")
		}

		_, ok = deps[TaskIDTeammates].([]*npc.Teammate)
		if !ok {
			return nil, fmt.Errorf("teammates dependency not found or invalid type")
		}

		// Create story engine
		engineConfig := engine.DefaultEngineConfig()
		engineConfig.Provider = smartProvider
		engineConfig.GameConfig = gameConfig
		storyEngine := engine.NewStoryEngine(engineConfig)

		// Create streaming callback for progress reporting
		streamCallback := func(chunk string) {
			// Streaming callback (content accumulates)
		}

		progressCallback := func(progress int, state engine.EstimationState) {
			// Progress callback from story engine
		}

		// Generate opening story
		result, err := storyEngine.GenerateOpening(ctx, streamCallback, progressCallback)
		if err != nil {
			return nil, fmt.Errorf("failed to generate opening story: %w", err)
		}

		return result, nil
	}).WithDependencies(TaskIDRules, TaskIDTeammates)

	tasks = append(tasks, storyTask)

	// Task 4: Opening Dream (slow, depends on Rules)
	dreamTask := NewTask(TaskIDDream, "生成開場夢境", PriorityMedium, 20*time.Second, func(ctx context.Context, deps DependencyResults) (interface{}, error) {
		// Get dependencies
		ruleSet, ok := deps[TaskIDRules].(*rules.RuleSet)
		if !ok {
			return nil, fmt.Errorf("rules dependency not found or invalid type")
		}

		// Build rules summary for dream
		rulesSummary := buildRulesSummary(ruleSet)

		// Create adapter for SmartModelClient
		client := &providerAdapter{provider: smartProvider}

		// Create dream generator
		dreamGen := engine.NewDreamGenerator(client)

		// Generate opening dream
		content, err := dreamGen.GenerateOpeningDream(
			ctx,
			gameConfig.Theme,
			rulesSummary,
			"探險者", // Default player role
		)

		if err != nil {
			return nil, fmt.Errorf("failed to generate opening dream: %w", err)
		}

		return content, nil
	}).WithDependencies(TaskIDRules)

	tasks = append(tasks, dreamTask)

	return tasks, nil
}

// providerAdapter adapts api.Provider to engine.SmartModelClient
type providerAdapter struct {
	provider api.Provider
}

// GenerateText implements engine.SmartModelClient interface
func (a *providerAdapter) GenerateText(ctx context.Context, prompt string) (string, error) {
	messages := []api.Message{
		{Role: "user", Content: prompt},
	}

	resp, err := a.provider.SendMessage(ctx, messages)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// buildRulesSummary creates a summary of rules for dream generation
func buildRulesSummary(ruleSet *rules.RuleSet) string {
	if ruleSet == nil {
		return "未知規則"
	}

	// Use CountByType since GetAll doesn't exist
	typeCount := ruleSet.CountByType()
	if len(typeCount) == 0 {
		return "未知規則"
	}

	var parts []string
	for ruleType, count := range typeCount {
		parts = append(parts, fmt.Sprintf("%d 條%s規則", count, ruleType))
	}

	return strings.Join(parts, "、")
}
