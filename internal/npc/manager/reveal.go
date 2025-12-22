package manager

import "strings"

// ==========================================================================
// Story 1.6: Trait Revelation Logic
// ==========================================================================

// shouldReveal checks if a trait should progress to the next revelation stage.
// It evaluates all triggers using AND logic (all conditions must be met).
//
// AC #4: Checks triggering conditions and returns whether to advance revelation status
func shouldReveal(trait *TraitFull, state *NPCRuntimeState, context RevealContext) bool {
	// If no triggers defined, cannot reveal
	if len(trait.Triggers) == 0 {
		return false
	}

	// All triggers must be satisfied (AND logic)
	for _, trigger := range trait.Triggers {
		if !evaluateTrigger(trigger, state, context) {
			return false
		}
	}

	return true
}

// evaluateTrigger checks if a single trigger condition is satisfied.
func evaluateTrigger(trigger TraitTrigger, state *NPCRuntimeState, context RevealContext) bool {
	switch trigger.Type {
	case TrustLevel:
		return compareValues(state.Emotion.Trust, trigger.Threshold, trigger.Comparator)
	case FearLevel:
		return compareValues(state.Emotion.Fear, trigger.Threshold, trigger.Comparator)
	case StressLevel:
		return compareValues(state.Emotion.Stress, trigger.Threshold, trigger.Comparator)
	case InteractionCount:
		return compareValues(context.InteractionCount, trigger.Threshold, trigger.Comparator)
	case TimeBased:
		return compareValues(context.CurrentBeat, trigger.Threshold, trigger.Comparator)
	case Event:
		return containsEvent(context.RecentEvents, trigger.EventName)
	default:
		return false
	}
}

// compareValues compares two integer values based on the comparator.
// Supported comparators: ">=", "<=", "==", ">", "<"
func compareValues(actual, threshold int, comparator string) bool {
	switch comparator {
	case ">=":
		return actual >= threshold
	case "<=":
		return actual <= threshold
	case "==":
		return actual == threshold
	case ">":
		return actual > threshold
	case "<":
		return actual < threshold
	default:
		return false
	}
}

// containsEvent checks if an event name exists in the recent events list.
func containsEvent(events []string, eventName string) bool {
	for _, event := range events {
		if strings.EqualFold(event, eventName) {
			return true
		}
	}
	return false
}

// RevealTrait progresses a trait through the revelation stages.
// Story 8.1: Enhanced to support multi-phase progression: Hidden -> HintPhase1 -> HintPhase2 -> Revealed
// It updates the trait's status in the NPCRuntimeState.TraitStates map.
//
// AC #5: Implements the trait revelation state machine
func (s *NPCRuntimeState) RevealTrait(traitID string) {
	currentStatus, exists := s.TraitStates[traitID]
	if !exists {
		// If trait status doesn't exist, initialize it to Hidden
		currentStatus = Hidden.String()
	}

	// Story 8.1: Progress through four stages: Hidden -> HintPhase1 -> HintPhase2 -> Revealed
	switch TraitStatus(parseTraitStatus(currentStatus)) {
	case Hidden:
		s.TraitStates[traitID] = HintPhase1.String()
	case HintPhase1:
		s.TraitStates[traitID] = HintPhase2.String()
	case HintPhase2:
		s.TraitStates[traitID] = Revealed.String()
	case Revealed:
		// Already revealed, no further progression
		// Keep it as revealed
	}
}

// parseTraitStatus converts a string status to TraitStatus.
func parseTraitStatus(status string) TraitStatus {
	switch status {
	case "hidden":
		return Hidden
	case "hinting", "hint_phase_1": // Legacy compatibility
		return HintPhase1
	case "hint_phase_2":
		return HintPhase2
	case "revealed":
		return Revealed
	default:
		return Hidden
	}
}

// GetHintingTraits returns all traits that are currently in the "hinting" status.
// Story 8.1: Enhanced to include both HintPhase1 and HintPhase2 traits
// Returns the full TraitFull objects, not just hint strings.
//
// AC #6: Returns list of traits in Hinting state (includes both phases)
func (s *NPCRuntimeState) GetHintingTraits(profile *NPCProfile) []TraitFull {
	var hintingTraits []TraitFull

	for _, trait := range profile.Traits {
		status, exists := s.TraitStates[trait.ID]
		if exists && (status == HintPhase1.String() || status == HintPhase2.String()) {
			// Convert basic trait to full trait for compatibility
			// In a real system, profile.Traits would already be TraitFull
			fullTrait := FromBasicTrait(trait)
			fullTrait.Status = parseTraitStatus(status)
			hintingTraits = append(hintingTraits, fullTrait)
		}
	}

	return hintingTraits
}

// GetRevealedTraits returns all traits that have been fully revealed.
// Returns the full TraitFull objects.
func (s *NPCRuntimeState) GetRevealedTraits(profile *NPCProfile) []TraitFull {
	var revealedTraits []TraitFull

	for _, trait := range profile.Traits {
		status, exists := s.TraitStates[trait.ID]
		if exists && status == Revealed.String() {
			// Convert basic trait to full trait for compatibility
			fullTrait := FromBasicTrait(trait)
			fullTrait.Status = Revealed
			revealedTraits = append(revealedTraits, fullTrait)
		}
	}

	return revealedTraits
}

// GetTraitStatus returns the current status of a specific trait.
func (s *NPCRuntimeState) GetTraitStatus(traitID string) TraitStatus {
	status, exists := s.TraitStates[traitID]
	if !exists {
		return Hidden
	}
	return parseTraitStatus(status)
}

// InitializeTraitStatus initializes a trait to Hidden status if not already tracked.
func (s *NPCRuntimeState) InitializeTraitStatus(traitID string) {
	if _, exists := s.TraitStates[traitID]; !exists {
		s.TraitStates[traitID] = Hidden.String()
	}
}

// CheckAndRevealTraits evaluates all traits and progresses those that meet their trigger conditions.
// This is a convenience method that combines shouldReveal and RevealTrait.
//
// Returns a list of trait IDs that were progressed to the next stage.
func CheckAndRevealTraits(traits []TraitFull, state *NPCRuntimeState, context RevealContext) []string {
	var progressedTraits []string

	for i := range traits {
		trait := &traits[i]

		// Initialize trait status if not present
		state.InitializeTraitStatus(trait.ID)

		// Get current status
		currentStatus := state.GetTraitStatus(trait.ID)

		// Only check if not already fully revealed
		if currentStatus != Revealed {
			if shouldReveal(trait, state, context) {
				state.RevealTrait(trait.ID)
				progressedTraits = append(progressedTraits, trait.ID)
			}
		}
	}

	return progressedTraits
}

// ==========================================================================
// Story 8.1: Progressive Trait Revelation Optimization
// ==========================================================================

// GetTraitsInPhase returns all traits currently in a specific revelation phase.
// This allows for phase-specific hint retrieval.
//
// Story 8.1 AC1: Support for multi-phase trait status queries
func (s *NPCRuntimeState) GetTraitsInPhase(profile *NPCProfile, phase TraitStatus) []TraitFull {
	var traits []TraitFull

	for _, trait := range profile.Traits {
		status, exists := s.TraitStates[trait.ID]
		if exists && status == phase.String() {
			fullTrait := FromBasicTrait(trait)
			fullTrait.Status = phase
			traits = append(traits, fullTrait)
		}
	}

	return traits
}

// CalculatePhaseTransitionScore calculates a score (0-100) indicating whether a trait
// should transition to the next revelation phase.
//
// Story 8.1 AC4: Multi-factor phase transition scoring
//
// Scoring factors:
// - Trust level (0-40 points): Higher trust accelerates revelation
// - Interaction count (0-30 points): More interactions reveal more
// - Time/beat progression (0-30 points): Natural progression over time
//
// A score of 70+ generally triggers phase transition, but this should be combined
// with shouldReveal() trigger checking for final decision.
//
// Parameters:
//   - state: Current NPC runtime state (for emotion/trust)
//   - interactionCount: Number of relevant interactions with this trait
//   - currentBeat: Current game beat/time
//   - lastCheckBeat: Last beat when trait reveal was checked
//   - revealTier: Trait difficulty (1=easy, 2=medium, 3=hard)
//
// Returns: Score from 0-100
func CalculatePhaseTransitionScore(
	state *NPCRuntimeState,
	interactionCount int,
	currentBeat int,
	lastCheckBeat int,
	revealTier int,
) int {
	score := 0

	// Factor 1: Trust level (0-40 points)
	// High trust (80+): 40 points
	// Medium trust (50-79): 25 points
	// Low trust (20-49): 10 points
	// Very low trust (<20): 0 points
	trustScore := 0
	switch {
	case state.Emotion.Trust >= 80:
		trustScore = 40
	case state.Emotion.Trust >= 50:
		trustScore = 25
	case state.Emotion.Trust >= 20:
		trustScore = 10
	default:
		trustScore = 0
	}
	score += trustScore

	// Factor 2: Interaction count (0-30 points)
	// Easy traits (tier 1): 10 interactions = full points
	// Medium traits (tier 2): 15 interactions = full points
	// Hard traits (tier 3): 20 interactions = full points
	maxInteractions := 10 + (revealTier-1)*5
	interactionScore := (interactionCount * 30) / maxInteractions
	if interactionScore > 30 {
		interactionScore = 30
	}
	score += interactionScore

	// Factor 3: Time/beat progression (0-30 points)
	// Awards points for time passage since last check
	// This ensures traits eventually reveal even with low interaction
	beatsPassed := currentBeat - lastCheckBeat
	timeScore := beatsPassed / 5 // 1 point per 5 beats
	if timeScore > 30 {
		timeScore = 30
	}
	score += timeScore

	return score
}

// AccelerateTraitRevelation checks if player interaction should accelerate trait revelation.
// This is called when specific dialogue or actions indicate heightened interest in the trait.
//
// Story 8.1 AC3: Interaction-based acceleration mechanism
//
// Returns true if the trait should immediately progress to the next phase.
//
// Acceleration occurs when:
// - Trust level is high (70+) AND interaction shows deep engagement
// - Special trigger events occur (e.g., player asks directly about the trait)
// - Emotional breakthrough moments (fear drops significantly, indicating openness)
//
// Parameters:
//   - trait: The trait being evaluated
//   - state: Current NPC runtime state
//   - context: Revelation context with game state
//   - playerAction: Description of player's action/dialogue (for semantic matching)
//
// Returns: true if trait should immediately progress
func AccelerateTraitRevelation(
	trait *TraitFull,
	state *NPCRuntimeState,
	context RevealContext,
	playerAction string,
) bool {
	// Condition 1: High trust (70+) enables acceleration
	if state.Emotion.Trust < 70 {
		return false
	}

	// Condition 2: Check if player action is relevant to this trait
	// This is a simple substring check; real implementation could use
	// semantic similarity or LLM-based matching
	isRelevantAction := len(playerAction) > 0 &&
		(containsKeywords(playerAction, trait.Content) ||
			containsKeywords(playerAction, trait.ID))

	if !isRelevantAction {
		return false
	}

	// Condition 3: Emotional breakthrough (Fear decreased recently)
	// This would require tracking previous fear levels, which we'll skip for now
	// Future enhancement: state.PreviousEmotion.Fear - state.Emotion.Fear > 20

	// Condition 4: Minimum interaction threshold
	// Don't accelerate if there have been too few interactions
	if context.InteractionCount < 3 {
		return false
	}

	return true
}

// containsKeywords checks if text contains any keywords from the reference string.
// This is a simple heuristic for relevance matching.
//
// The function splits the reference into words and checks if any significant word
// (length > 3) appears in the text. This provides flexible matching for trait relevance.
func containsKeywords(text string, reference string) bool {
	if reference == "" {
		return true // Empty reference always matches
	}

	textLower := strings.ToLower(text)
	referenceLower := strings.ToLower(reference)

	// First try exact substring match (most specific)
	if strings.Contains(textLower, referenceLower) {
		return true
	}

	// Fall back to keyword matching: split reference into words and check each
	words := strings.FieldsFunc(referenceLower, func(r rune) bool {
		// Split on whitespace and common punctuation
		return r == ' ' || r == ',' || r == '.' || r == '(' || r == ')' ||
		       r == '[' || r == ']' || r == '{' || r == '}' || r == ';' ||
		       r == ':' || r == '!' || r == '?'
	})

	for _, word := range words {
		// Only consider words longer than 3 characters (skip "of", "the", etc.)
		if len(word) > 3 && strings.Contains(textLower, word) {
			return true
		}
	}

	return false
}

// CheckAndProgressTraits evaluates all traits and progresses those ready for next phase.
// Story 8.1 version with multi-phase support and phase transition scoring.
//
// This is an enhanced version of CheckAndRevealTraits that uses the new
// CalculatePhaseTransitionScore algorithm for more nuanced progression.
//
// Story 8.1 AC4: Natural phase transition based on multi-factor scoring
//
// Returns: Map from trait ID to new status (for traits that progressed)
func CheckAndProgressTraits(
	traits []TraitFull,
	state *NPCRuntimeState,
	context RevealContext,
) map[string]string {
	progressedTraits := make(map[string]string)

	for i := range traits {
		trait := &traits[i]

		// Initialize trait status if not present
		state.InitializeTraitStatus(trait.ID)

		// Get current status
		currentStatus := state.GetTraitStatus(trait.ID)

		// Only check if not already fully revealed
		if currentStatus == Revealed {
			continue
		}

		// Calculate phase transition score
		score := CalculatePhaseTransitionScore(
			state,
			trait.InteractionCount,
			context.CurrentBeat,
			trait.LastRevealCheck,
			trait.RevealTier,
		)

		// Threshold for progression: 70+ score
		shouldProgress := score >= 70

		// Also check traditional trigger-based reveal
		if !shouldProgress && shouldReveal(trait, state, context) {
			shouldProgress = true
		}

		// Progress if conditions met
		if shouldProgress {
			// Store old status for comparison
			oldStatus := state.TraitStates[trait.ID]

			// Progress to next phase
			state.RevealTrait(trait.ID)

			// Track the progression
			newStatus := state.TraitStates[trait.ID]
			if newStatus != oldStatus {
				progressedTraits[trait.ID] = newStatus

				// Update last check time
				trait.LastRevealCheck = context.CurrentBeat
			}
		}
	}

	return progressedTraits
}
