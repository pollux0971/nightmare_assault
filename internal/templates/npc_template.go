package templates

// NPCArchetype represents the archetype category of an NPC
type NPCArchetype string

const (
	NPCArchetypeSurvivor NPCArchetype = "survivor" // 倖存者
	NPCArchetypeThreat   NPCArchetype = "threat"   // 威脅者
	NPCArchetypeNeutral  NPCArchetype = "neutral"  // 中立者
	NPCArchetypeHelper   NPCArchetype = "helper"   // 幫助者
	NPCArchetypeVictim   NPCArchetype = "victim"   // 受害者
	NPCArchetypeBetrayer NPCArchetype = "betrayer" // 背叛者
)

// KnowledgeLevel represents how much an NPC knows about the rules
type KnowledgeLevel string

const (
	KnowledgeLevelNone    KnowledgeLevel = "none"    // 完全不知情
	KnowledgeLevelPartial KnowledgeLevel = "partial" // 部分知情
	KnowledgeLevelFull    KnowledgeLevel = "full"    // 完全知情
)

// TrustLevel represents how trustworthy an NPC is
type TrustLevel string

const (
	TrustLevelUntrustworthy TrustLevel = "untrustworthy" // 不可信
	TrustLevelNeutral       TrustLevel = "neutral"       // 中立
	TrustLevelTrustworthy   TrustLevel = "trustworthy"   // 可信
)

// NPCState represents one state in an NPC's mental/behavioral transformation
type NPCState struct {
	Description       string   `yaml:"description"`                  // 狀態描述
	PersonalityTraits []string `yaml:"personality_traits"`           // 性格特質
	DialogueStyle     string   `yaml:"dialogue_style"`               // 對話風格
	BehaviorPatterns  []string `yaml:"behavior_patterns,omitempty"`  // 行為模式（可選）
}

// NPCTemplate represents an NPC archetype template
// Implements three-state transformation: 正常狀態 → 焦慮狀態 → 崩壞狀態
type NPCTemplate struct {
	ID               string         `yaml:"id"`                          // 唯一標識符
	Name             string         `yaml:"name"`                        // 原型名稱
	Archetype        NPCArchetype   `yaml:"archetype"`                   // 原型類別
	FunctionalRole   string         `yaml:"functional_role"`             // 功能定位
	NormalState      NPCState       `yaml:"normal_state"`                // 正常狀態
	AnxiousState     NPCState       `yaml:"anxious_state"`               // 焦慮狀態
	CorruptedState   NPCState       `yaml:"corrupted_state"`             // 崩壞狀態
	SpecialAbilities []string       `yaml:"special_abilities,omitempty"` // 特殊能力列表
	KnowledgeLevel   KnowledgeLevel `yaml:"knowledge_level"`             // 知識程度
	TrustLevel       TrustLevel     `yaml:"trust_level"`                 // 可信度
	BackgroundHints  []string       `yaml:"background_hints,omitempty"`  // 背景提示列表
	Tags             []string       `yaml:"tags,omitempty"`              // 標籤（用於分類和搜索）
	Description      string         `yaml:"description,omitempty"`       // 原型描述（可選）
}

// NPCTemplateCollection represents a collection of NPC templates from a YAML file
type NPCTemplateCollection struct {
	Version   string         `yaml:"version"`   // 模板版本
	NPCTypes  []*NPCTemplate `yaml:"npc_types"` // NPC 原型列表
}

// Validate checks if the NPC template is valid
func (nt *NPCTemplate) Validate() error {
	if nt.ID == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "npc ID cannot be empty",
		}
	}

	if nt.Name == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "npc name cannot be empty",
		}
	}

	// Validate Normal State (正常狀態)
	if err := nt.validateState(nt.NormalState, "normal_state"); err != nil {
		return err
	}

	// Validate Anxious State (焦慮狀態)
	if err := nt.validateState(nt.AnxiousState, "anxious_state"); err != nil {
		return err
	}

	// Validate Corrupted State (崩壞狀態)
	if err := nt.validateState(nt.CorruptedState, "corrupted_state"); err != nil {
		return err
	}

	return nil
}

// validateState validates a single NPC state
func (nt *NPCTemplate) validateState(state NPCState, stateName string) error {
	if state.Description == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: stateName + " description cannot be empty",
		}
	}

	if len(state.PersonalityTraits) == 0 {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: stateName + " personality_traits list cannot be empty",
		}
	}

	if state.DialogueStyle == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: stateName + " dialogue_style cannot be empty",
		}
	}

	return nil
}

// HasTag checks if the NPC has a specific tag
func (nt *NPCTemplate) HasTag(tag string) bool {
	for _, t := range nt.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetState returns the specified state (1=normal, 2=anxious, 3=corrupted)
func (nt *NPCTemplate) GetState(stateNumber int) *NPCState {
	switch stateNumber {
	case 1:
		return &nt.NormalState
	case 2:
		return &nt.AnxiousState
	case 3:
		return &nt.CorruptedState
	default:
		return nil
	}
}

// HasPersonalityTrait checks if any state has a specific personality trait
func (nt *NPCTemplate) HasPersonalityTrait(trait string) bool {
	states := []*NPCState{&nt.NormalState, &nt.AnxiousState, &nt.CorruptedState}
	for _, state := range states {
		for _, t := range state.PersonalityTraits {
			if t == trait {
				return true
			}
		}
	}
	return false
}

// HasAbility checks if the NPC has a specific special ability
func (nt *NPCTemplate) HasAbility(ability string) bool {
	for _, a := range nt.SpecialAbilities {
		if a == ability {
			return true
		}
	}
	return false
}

// GetPersonalityTraitsString returns all personality traits from all states
func (nt *NPCTemplate) GetPersonalityTraitsString() string {
	traits := make([]string, 0)
	states := []*NPCState{&nt.NormalState, &nt.AnxiousState, &nt.CorruptedState}
	for _, state := range states {
		traits = append(traits, state.PersonalityTraits...)
	}

	result := ""
	for i, trait := range traits {
		if i > 0 {
			result += ", "
		}
		result += trait
	}
	return result
}

// IsKnowledgeable checks if the NPC has partial or full knowledge
func (nt *NPCTemplate) IsKnowledgeable() bool {
	return nt.KnowledgeLevel == KnowledgeLevelPartial || nt.KnowledgeLevel == KnowledgeLevelFull
}

// IsTrustworthy checks if the NPC is trustworthy
func (nt *NPCTemplate) IsTrustworthy() bool {
	return nt.TrustLevel == TrustLevelTrustworthy
}

// GetAbilityCount returns the number of special abilities
func (nt *NPCTemplate) GetAbilityCount() int {
	return len(nt.SpecialAbilities)
}
