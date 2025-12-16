package templates

// SceneCategory represents the category of a scene template
type SceneCategory string

const (
	SceneCategoryBiological SceneCategory = "biological" // 血肉生物類
	SceneCategoryTemporal   SceneCategory = "temporal"   // 時間異常類
	SceneCategoryDigital    SceneCategory = "digital"    // 數位虛擬類
	SceneCategorySpatial    SceneCategory = "spatial"    // 空間扭曲類
)

// SceneStage represents one stage in a scene's horror progression
type SceneStage struct {
	Description string   `yaml:"description"` // 階段的恐怖描述
	Atmosphere  []string `yaml:"atmosphere"`  // 氛圍關鍵字列表
	CommonProps []string `yaml:"common_props,omitempty"` // 常見道具列表（可選）
	Hazards     []string `yaml:"hazards,omitempty"`      // 危險元素列表（可選）
}

// SceneTemplate represents a scene template for horror game scenes
// Implements three-stage horror progression: 日常違和 → 顯著異常 → 法則崩壞
type SceneTemplate struct {
	ID              string        `yaml:"id"`               // 唯一標識符
	Name            string        `yaml:"name"`             // 場景名稱
	Category        SceneCategory `yaml:"category"`         // 場景類別
	ApplicableAreas []string      `yaml:"applicable_areas"` // 適用場景（如：醫院、屠宰場）
	Stage1          SceneStage    `yaml:"stage1"`           // 階段1：日常違和
	Stage2          SceneStage    `yaml:"stage2"`           // 階段2：顯著異常
	Stage3          SceneStage    `yaml:"stage3"`           // 階段3：法則崩壞
	Tags            []string      `yaml:"tags"`             // 標籤（用於分類和搜索）
	Description     string        `yaml:"description"`      // 場景總體描述（可選）
}

// SceneTemplateCollection represents a collection of scene templates from a YAML file
type SceneTemplateCollection struct {
	Version  string           `yaml:"version"`  // 模板版本
	Category SceneCategory    `yaml:"category"` // 集合的類別
	Scenes   []*SceneTemplate `yaml:"scenes"`   // 場景列表
}

// Validate checks if the scene template is valid
func (st *SceneTemplate) Validate() error {
	if st.ID == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "scene ID cannot be empty",
		}
	}

	if st.Name == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: "scene name cannot be empty",
		}
	}

	// Validate Stage 1 (日常違和)
	if err := st.validateStage(st.Stage1, "stage1"); err != nil {
		return err
	}

	// Validate Stage 2 (顯著異常)
	if err := st.validateStage(st.Stage2, "stage2"); err != nil {
		return err
	}

	// Validate Stage 3 (法則崩壞)
	if err := st.validateStage(st.Stage3, "stage3"); err != nil {
		return err
	}

	return nil
}

// validateStage validates a single scene stage
func (st *SceneTemplate) validateStage(stage SceneStage, stageName string) error {
	if stage.Description == "" {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: stageName + " description cannot be empty",
		}
	}

	if len(stage.Atmosphere) == 0 {
		return &LoadError{
			ErrType: "VALIDATION_ERROR",
			Message: stageName + " atmosphere list cannot be empty",
		}
	}

	return nil
}

// HasTag checks if the scene has a specific tag
func (st *SceneTemplate) HasTag(tag string) bool {
	for _, t := range st.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetStage returns the specified stage (1, 2, or 3)
func (st *SceneTemplate) GetStage(stageNumber int) *SceneStage {
	switch stageNumber {
	case 1:
		return &st.Stage1
	case 2:
		return &st.Stage2
	case 3:
		return &st.Stage3
	default:
		return nil
	}
}

// HasAtmosphere checks if any stage has a specific atmosphere keyword
func (st *SceneTemplate) HasAtmosphere(keyword string) bool {
	stages := []*SceneStage{&st.Stage1, &st.Stage2, &st.Stage3}
	for _, stage := range stages {
		for _, a := range stage.Atmosphere {
			if a == keyword {
				return true
			}
		}
	}
	return false
}

// GetAtmosphereKeywords returns all atmosphere keywords from all stages
func (st *SceneTemplate) GetAtmosphereKeywords() string {
	keywords := make([]string, 0)
	stages := []*SceneStage{&st.Stage1, &st.Stage2, &st.Stage3}
	for _, stage := range stages {
		keywords = append(keywords, stage.Atmosphere...)
	}

	result := ""
	for i, keyword := range keywords {
		if i > 0 {
			result += ", "
		}
		result += keyword
	}
	return result
}

// CountElements returns the total number of props and hazards across all stages
func (st *SceneTemplate) CountElements() int {
	count := 0
	stages := []*SceneStage{&st.Stage1, &st.Stage2, &st.Stage3}
	for _, stage := range stages {
		count += len(stage.CommonProps) + len(stage.Hazards)
	}
	return count
}

// HasApplicableArea checks if the scene is applicable to a specific area
func (st *SceneTemplate) HasApplicableArea(area string) bool {
	for _, a := range st.ApplicableAreas {
		if a == area {
			return true
		}
	}
	return false
}
