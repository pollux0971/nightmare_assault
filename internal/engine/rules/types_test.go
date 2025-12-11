package rules

import (
	"encoding/json"
	"testing"
)

func TestRuleTypeString(t *testing.T) {
	tests := []struct {
		ruleType RuleType
		expected string
	}{
		{RuleTypeScenario, "場景規則"},
		{RuleTypeTime, "時間規則"},
		{RuleTypeBehavior, "行為規則"},
		{RuleTypeObject, "對象規則"},
		{RuleTypeStatus, "狀態規則"},
	}

	for _, tt := range tests {
		if got := tt.ruleType.String(); got != tt.expected {
			t.Errorf("RuleType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestRuleTypeEnglishName(t *testing.T) {
	tests := []struct {
		ruleType RuleType
		expected string
	}{
		{RuleTypeScenario, "Scenario"},
		{RuleTypeTime, "Time"},
		{RuleTypeBehavior, "Behavior"},
		{RuleTypeObject, "Object"},
		{RuleTypeStatus, "Status"},
	}

	for _, tt := range tests {
		if got := tt.ruleType.EnglishName(); got != tt.expected {
			t.Errorf("RuleType.EnglishName() = %v, want %v", got, tt.expected)
		}
	}
}

func TestRuleTypeJSONMarshal(t *testing.T) {
	tests := []struct {
		ruleType RuleType
		expected string
	}{
		{RuleTypeScenario, `"Scenario"`},
		{RuleTypeTime, `"Time"`},
		{RuleTypeBehavior, `"Behavior"`},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.ruleType)
		if err != nil {
			t.Errorf("Marshal failed: %v", err)
		}
		if string(data) != tt.expected {
			t.Errorf("Marshal = %s, want %s", data, tt.expected)
		}
	}
}

func TestRuleTypeJSONUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		expected RuleType
	}{
		{`"Scenario"`, RuleTypeScenario},
		{`"Time"`, RuleTypeTime},
		{`"Behavior"`, RuleTypeBehavior},
		{`"Object"`, RuleTypeObject},
		{`"Status"`, RuleTypeStatus},
	}

	for _, tt := range tests {
		var rt RuleType
		if err := json.Unmarshal([]byte(tt.input), &rt); err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}
		if rt != tt.expected {
			t.Errorf("Unmarshal = %v, want %v", rt, tt.expected)
		}
	}
}

func TestParseRuleType(t *testing.T) {
	tests := []struct {
		input    string
		expected RuleType
	}{
		{"Scenario", RuleTypeScenario},
		{"場景規則", RuleTypeScenario},
		{"scenario", RuleTypeScenario},
		{"Time", RuleTypeTime},
		{"Behavior", RuleTypeBehavior},
		{"unknown", RuleTypeBehavior}, // default
	}

	for _, tt := range tests {
		if got := ParseRuleType(tt.input); got != tt.expected {
			t.Errorf("ParseRuleType(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestAllRuleTypes(t *testing.T) {
	types := AllRuleTypes()
	if len(types) != 5 {
		t.Errorf("AllRuleTypes() returned %d types, want 5", len(types))
	}
}

func TestConsequenceTypeString(t *testing.T) {
	tests := []struct {
		conseq   ConsequenceType
		expected string
	}{
		{ConsequenceWarning, "警告"},
		{ConsequenceDamage, "傷害"},
		{ConsequenceInstantDeath, "即死"},
	}

	for _, tt := range tests {
		if got := tt.conseq.String(); got != tt.expected {
			t.Errorf("ConsequenceType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestConsequenceTypeJSONRoundtrip(t *testing.T) {
	tests := []ConsequenceType{
		ConsequenceWarning,
		ConsequenceDamage,
		ConsequenceInstantDeath,
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt)
		if err != nil {
			t.Errorf("Marshal failed: %v", err)
		}

		var ct ConsequenceType
		if err := json.Unmarshal(data, &ct); err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}

		if ct != tt {
			t.Errorf("Roundtrip = %v, want %v", ct, tt)
		}
	}
}

func TestNewRule(t *testing.T) {
	r := NewRule("test-rule-1", RuleTypeScenario)

	if r.ID != "test-rule-1" {
		t.Errorf("ID = %v, want test-rule-1", r.ID)
	}
	if r.Type != RuleTypeScenario {
		t.Errorf("Type = %v, want RuleTypeScenario", r.Type)
	}
	if r.Priority != 5 {
		t.Errorf("Priority = %v, want 5", r.Priority)
	}
	if !r.Active {
		t.Error("Active should be true by default")
	}
	if r.MaxViolations != 1 {
		t.Errorf("MaxViolations = %v, want 1", r.MaxViolations)
	}
}

func TestRuleRecordViolation(t *testing.T) {
	r := NewRule("test", RuleTypeBehavior)
	r.MaxViolations = 2

	// First violation - not exceeded
	if r.RecordViolation() {
		t.Error("First violation should not exceed max")
	}
	if r.Violations != 1 {
		t.Errorf("Violations = %d, want 1", r.Violations)
	}

	// Second violation - not exceeded
	if r.RecordViolation() {
		t.Error("Second violation should not exceed max")
	}

	// Third violation - exceeded
	if !r.RecordViolation() {
		t.Error("Third violation should exceed max")
	}
}

func TestRuleShouldWarn(t *testing.T) {
	r := NewRule("test", RuleTypeBehavior)
	r.MaxViolations = 1
	r.WarningText = "注意！"

	// No violations yet
	if !r.ShouldWarn() {
		t.Error("Should warn when no violations")
	}

	r.RecordViolation()
	// At max violations
	if !r.ShouldWarn() {
		t.Error("Should warn at max violations")
	}

	r.RecordViolation()
	// Exceeded max
	if r.ShouldWarn() {
		t.Error("Should not warn when exceeded")
	}
}

func TestRuleSetOperations(t *testing.T) {
	rs := NewRuleSet()

	if rs.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", rs.Count())
	}

	// Add rules
	r1 := NewRule("rule-1", RuleTypeScenario)
	r2 := NewRule("rule-2", RuleTypeBehavior)
	r3 := NewRule("rule-3", RuleTypeScenario)
	r3.Active = false

	rs.Add(r1)
	rs.Add(r2)
	rs.Add(r3)

	if rs.Count() != 3 {
		t.Errorf("Count = %d, want 3", rs.Count())
	}

	// Get by ID
	if rs.GetByID("rule-1") != r1 {
		t.Error("GetByID failed for rule-1")
	}
	if rs.GetByID("nonexistent") != nil {
		t.Error("GetByID should return nil for nonexistent")
	}

	// Get by type
	scenarios := rs.GetByType(RuleTypeScenario)
	if len(scenarios) != 2 {
		t.Errorf("GetByType(Scenario) = %d, want 2", len(scenarios))
	}

	// Get active
	active := rs.GetActive()
	if len(active) != 2 {
		t.Errorf("GetActive = %d, want 2", len(active))
	}

	// Count by type
	counts := rs.CountByType()
	if counts[RuleTypeScenario] != 2 {
		t.Errorf("CountByType[Scenario] = %d, want 2", counts[RuleTypeScenario])
	}
	if counts[RuleTypeBehavior] != 1 {
		t.Errorf("CountByType[Behavior] = %d, want 1", counts[RuleTypeBehavior])
	}
}

func TestRuleSetJSONRoundtrip(t *testing.T) {
	rs := NewRuleSet()

	r := NewRule("test-rule", RuleTypeTime)
	r.Trigger = Condition{Type: "time", Value: "midnight", Operator: "equals"}
	r.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 10, SANDamage: 20}
	r.Clues = []string{"時鐘指向十二點", "月光異常明亮"}
	r.WarningText = "感覺不太對勁..."

	rs.Add(r)

	// Serialize
	data, err := rs.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Deserialize
	rs2 := NewRuleSet()
	if err := rs2.FromJSON(data); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if rs2.Count() != 1 {
		t.Errorf("Roundtrip count = %d, want 1", rs2.Count())
	}

	r2 := rs2.GetByID("test-rule")
	if r2 == nil {
		t.Fatal("Rule not found after roundtrip")
	}
	if r2.Type != RuleTypeTime {
		t.Errorf("Type = %v, want RuleTypeTime", r2.Type)
	}
	if r2.Consequence.HPDamage != 10 {
		t.Errorf("HPDamage = %d, want 10", r2.Consequence.HPDamage)
	}
	if len(r2.Clues) != 2 {
		t.Errorf("Clues count = %d, want 2", len(r2.Clues))
	}
}

func TestRuleGetDiscovered(t *testing.T) {
	rs := NewRuleSet()

	r1 := NewRule("rule-1", RuleTypeScenario)
	r1.Discovered = true
	r2 := NewRule("rule-2", RuleTypeBehavior)
	r2.Discovered = false

	rs.Add(r1)
	rs.Add(r2)

	discovered := rs.GetDiscovered()
	if len(discovered) != 1 {
		t.Errorf("GetDiscovered = %d, want 1", len(discovered))
	}
	if discovered[0].ID != "rule-1" {
		t.Errorf("Discovered rule ID = %s, want rule-1", discovered[0].ID)
	}
}
