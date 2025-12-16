package templates

// RuleCategory represents the category of a rule template
type RuleCategory string

const (
	RuleCategorySensory RuleCategory = "sensory" // 感官認知
	RuleCategorySpatial RuleCategory = "spatial" // 空間邏輯
	RuleCategorySocial  RuleCategory = "social"  // 社交互動
)

// RuleDifficulty represents the difficulty level of a rule
type RuleDifficulty string

const (
	RuleDifficultyEasy   RuleDifficulty = "easy"   // 簡單
	RuleDifficultyMedium RuleDifficulty = "medium" // 中等
	RuleDifficultyHard   RuleDifficulty = "hard"   // 困難
	RuleDifficultyHell   RuleDifficulty = "hell"   // 地獄
)

// Punishment represents the consequences of violating a rule
type Punishment struct {
	SANDamage int    `yaml:"san_damage"` // SAN 損失
	HPDamage  int    `yaml:"hp_damage"`  // HP 損失
	Effect    string `yaml:"effect"`     // 效果描述
}

// RuleTemplate represents a rule template for horror game rules
type RuleTemplate struct {
	ID            string         `yaml:"id"`              // 唯一標識符
	Name          string         `yaml:"name"`            // 規則名稱
	Category      RuleCategory   `yaml:"category"`        // 規則類別
	Difficulty    RuleDifficulty `yaml:"difficulty"`      // 難度等級
	TriggerMedium string         `yaml:"trigger_medium"`  // 正確觸發媒介
	FalseClue     string         `yaml:"false_clue"`      // 錯誤線索/煙霧彈
	SurvivalRule  string         `yaml:"survival_rule"`   // 生存規則
	Punishment    Punishment     `yaml:"punishment"`      // 懲罰
	ClueHints     []string       `yaml:"clue_hints"`      // 線索提示列表
	Tags          []string       `yaml:"tags"`            // 標籤（用於分類和搜索）
}

// RuleTemplateCollection represents a collection of rule templates from a YAML file
type RuleTemplateCollection struct {
	Version  string          `yaml:"version"`  // 模板版本
	Category RuleCategory    `yaml:"category"` // 集合的類別
	Rules    []*RuleTemplate `yaml:"rules"`    // 規則列表
}

// Validate checks if the rule template is valid
func (rt *RuleTemplate) Validate() error {
	if rt.ID == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "rule ID cannot be empty",
		}
	}

	if rt.Name == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "rule name cannot be empty",
		}
	}

	if rt.TriggerMedium == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "trigger_medium cannot be empty",
		}
	}

	if rt.SurvivalRule == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "survival_rule cannot be empty",
		}
	}

	return nil
}

// GetDifficultyLevel returns the numeric difficulty level (1-4)
func (rt *RuleTemplate) GetDifficultyLevel() int {
	switch rt.Difficulty {
	case RuleDifficultyEasy:
		return 1
	case RuleDifficultyMedium:
		return 2
	case RuleDifficultyHard:
		return 3
	case RuleDifficultyHell:
		return 4
	default:
		return 1
	}
}

// HasTag checks if the rule has a specific tag
func (rt *RuleTemplate) HasTag(tag string) bool {
	for _, t := range rt.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetTotalDamage returns the total potential damage (HP + SAN)
func (rt *RuleTemplate) GetTotalDamage() int {
	return rt.Punishment.HPDamage + rt.Punishment.SANDamage
}
