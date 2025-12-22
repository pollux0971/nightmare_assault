package orchestrator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.6: NPC Profile Tests
// ==========================================================================

func TestNewNPCProfile(t *testing.T) {
	instance := agents.NPCInstance{
		ID:          "NPC-001",
		Name:        "王小芳",
		Archetype:   agents.NPCArchetypeSacrificial,
		Personality: []string{"無助", "恐懼", "善良"},
		Appearance:  "瘦弱的身軀，蒼白的臉色，眼神中充滿恐懼與無助。衣著凌亂，手腕上有不明傷痕。她的雙手微微顫抖，似乎隨時都會崩潰。",
		Backstory:   "一個不幸捲入這場噩夢的無辜者，過去平凡的生活已經一去不復返。每天都在恐懼中度過，不知道下一刻會發生什麼。曾經試圖逃離卻總是失敗，身心都受到了極大的折磨。現在只剩下恐懼與絕望，不知道明天是否還能活著，只能在黑暗中無助地等待著未知的命運降臨。",
		Skills:      []string{"躲藏", "求救"},
		Inventory:   []string{"破舊的護士服", "沾血的繃帶"},
		Secret:      "她曾目睹醫院地下室的非法實驗，但因恐懼而選擇沉默，現在那些記憶不斷折磨著她。",
		Introduction: "王小芳蜷縮在角落，雙手死死抓著破舊的護士服，眼神不斷飄向門口。「救...救命...」聲音顫抖得幾乎聽不清，身體抖得像秋風中的落葉。",
		LinkedSeeds: []string{},
		DeathTiming: 15,
		Status:      agents.NPCStatusAlive,
	}

	profile := NewNPCProfile(instance)

	if profile.ID != instance.ID {
		t.Errorf("Expected ID %s, got %s", instance.ID, profile.ID)
	}
	if profile.Name != instance.Name {
		t.Errorf("Expected Name %s, got %s", instance.Name, profile.Name)
	}
	if profile.Archetype != instance.Archetype {
		t.Errorf("Expected Archetype %s, got %s", instance.Archetype, profile.Archetype)
	}
	if len(profile.Personality) != len(instance.Personality) {
		t.Errorf("Expected Personality length %d, got %d", len(instance.Personality), len(profile.Personality))
	}
	if profile.Appearance != instance.Appearance {
		t.Error("Appearance mismatch")
	}
	if profile.Backstory != instance.Backstory {
		t.Error("Backstory mismatch")
	}
}

func TestNPCProfile_ToNPCInstance(t *testing.T) {
	profile := &NPCProfile{
		ID:          "NPC-002",
		Name:        "李醫生",
		Archetype:   agents.NPCArchetypeKnowledgeable,
		Personality: []string{"神秘", "謹慎", "知情"},
		Appearance:  "深邃的眼神，神秘的氣質。穿著整潔但略顯陳舊的白袍，總是若有所思地看著遠方。眼角的皺紋透露出他經歷過的滄桑。",
		Backstory:   "似乎對這裡發生的一切有所了解，但總是欲言又止。曾經目睹過許多不可思議的事情，但每次想要說出真相時都會感到莫名的恐懼。過去的經歷讓他學會了保持沉默，因為知道太多的人往往活不長。那些試圖揭露真相的人都遭遇了不幸，這讓他更加謹慎，只在關鍵時刻才會透露一些隱晦的線索。",
		Skills:      []string{"觀察", "解謎"},
		Inventory:   []string{"神秘的筆記", "古老的鑰匙"},
		Secret:      "他知道這個醫院的核心秘密，但受到某種約束無法直接揭露，只能用暗示引導他人發現真相。",
		Introduction: "李醫生靠在牆邊，手中翻著神秘的筆記，頭也不抬：「你來晚了。」指尖劃過書頁上的符號，眼神深邃得像看穿了一切。",
		LinkedSeeds: []string{"GS-001", "GS-002"},
		DeathTiming: 0,
		Status:      agents.NPCStatusAlive,
	}

	instance := profile.ToNPCInstance()

	if instance.ID != profile.ID {
		t.Errorf("Expected ID %s, got %s", profile.ID, instance.ID)
	}
	if instance.Name != profile.Name {
		t.Errorf("Expected Name %s, got %s", profile.Name, instance.Name)
	}
	if instance.Archetype != profile.Archetype {
		t.Errorf("Expected Archetype %s, got %s", profile.Archetype, instance.Archetype)
	}
	if len(instance.LinkedSeeds) != len(profile.LinkedSeeds) {
		t.Errorf("Expected LinkedSeeds length %d, got %d", len(profile.LinkedSeeds), len(instance.LinkedSeeds))
	}
}

func TestNPCProfile_Validate(t *testing.T) {
	tests := []struct {
		name      string
		profile   *NPCProfile
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid profile",
			profile: &NPCProfile{
				ID:          "NPC-001",
				Name:        "測試角色",
				Archetype:   agents.NPCArchetypeSacrificial,
				Personality: []string{"無助", "恐懼", "善良"},
				Appearance:  "瘦弱的身軀，蒼白的臉色，眼神中充滿恐懼與無助。衣著凌亂，手腕上有不明傷痕。她的雙手微微顫抖，似乎隨時都會崩潰。",
				Backstory:   "一個不幸捲入這場噩夢的無辜者，過去平凡的生活已經一去不復返。每天都在恐懼中度過，不知道下一刻會發生什麼。曾經試圖逃離卻總是失敗，身心都受到了極大的折磨。現在只剩下恐懼與絕望，不知道明天是否還能活著，只能在黑暗中無助地等待著未知的命運降臨。",
				Skills:      []string{"躲藏"},
				Inventory:   []string{"破舊的護士服"},
				Secret:      "隱藏的秘密，不為人知的過去，深埋在心底的恐懼，無法對任何人訴說的經歷與痛苦。",
				Introduction: "她蜷縮在角落，雙手死死抓著破舊的護士服，眼神不斷飄向門口。「救...救命...」聲音顫抖得幾乎聽不清，身體抖得像秋風中的落葉。",
				LinkedSeeds: []string{},
				DeathTiming: 15,
				Status:      agents.NPCStatusAlive,
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			profile: &NPCProfile{
				ID:          "",
				Name:        "測試角色",
				Archetype:   agents.NPCArchetypeSacrificial,
				Personality: []string{"無助", "恐懼", "善良"},
			},
			wantError: true,
			errorMsg:  "ID cannot be empty",
		},
		{
			name: "Too few personality keywords",
			profile: &NPCProfile{
				ID:          "NPC-001",
				Name:        "測試角色",
				Archetype:   agents.NPCArchetypeSacrificial,
				Personality: []string{"無助"},
			},
			wantError: true,
			errorMsg:  "personality must have 3-5 keywords",
		},
		{
			name: "Too many linked seeds",
			profile: &NPCProfile{
				ID:          "NPC-001",
				Name:        "測試角色",
				Archetype:   agents.NPCArchetypeSacrificial,
				Personality: []string{"無助", "恐懼", "善良"},
				Appearance:  "瘦弱的身軀，蒼白的臉色，眼神中充滿恐懼與無助。衣著凌亂，手腕上有不明傷痕。她的雙手微微顫抖，似乎隨時都會崩潰。",
				Backstory:   "一個不幸捲入這場噩夢的無辜者，過去平凡的生活已經一去不復返。每天都在恐懼中度過，不知道下一刻會發生什麼。曾經試圖逃離卻總是失敗，身心都受到了極大的折磨。現在只剩下恐懼與絕望，不知道明天是否還能活著，只能在黑暗中無助地等待著未知的命運降臨。",
				Skills:      []string{"躲藏"},
				Inventory:   []string{"物品"},
				Secret:      "秘密內容，不為人知的過去，深埋在心底的恐懼，無法對任何人訴說的經歷與痛苦。",
				LinkedSeeds: []string{"GS-001", "GS-002", "GS-003"},
			},
			wantError: true,
			errorMsg:  "linked_seeds must have at most 2 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestNPCProfile_StatusMethods(t *testing.T) {
	profile := &NPCProfile{
		Status: agents.NPCStatusAlive,
	}

	if !profile.IsAlive() {
		t.Error("Expected IsAlive() to return true")
	}
	if profile.IsDead() {
		t.Error("Expected IsDead() to return false")
	}

	profile.MarkDead(20, "被怪物殺死")

	if profile.IsAlive() {
		t.Error("Expected IsAlive() to return false after MarkDead")
	}
	if !profile.IsDead() {
		t.Error("Expected IsDead() to return true after MarkDead")
	}
	if profile.DeathBeat != 20 {
		t.Errorf("Expected DeathBeat 20, got %d", profile.DeathBeat)
	}
	if profile.DeathReason != "被怪物殺死" {
		t.Errorf("Expected DeathReason '被怪物殺死', got '%s'", profile.DeathReason)
	}
	if profile.Status != agents.NPCStatusDead {
		t.Errorf("Expected Status %s, got %s", agents.NPCStatusDead, profile.Status)
	}
}

func TestNPCProfile_Serialization(t *testing.T) {
	profile := &NPCProfile{
		ID:          "NPC-001",
		Name:        "王小芳",
		Archetype:   agents.NPCArchetypeSacrificial,
		Personality: []string{"無助", "恐懼", "善良"},
		Appearance:  "瘦弱的身軀，蒼白的臉色，眼神中充滿恐懼與無助。衣著凌亂，手腕上有不明傷痕。她的雙手微微顫抖，似乎隨時都會崩潰。",
		Backstory:   "一個不幸捲入這場噩夢的無辜者，過去平凡的生活已經一去不復返。每天都在恐懼中度過，不知道下一刻會發生什麼。曾經試圖逃離卻總是失敗，身心都受到了極大的折磨。現在只剩下恐懼與絕望，不知道明天是否還能活著，只能在黑暗中無助地等待著未知的命運降臨。",
		Skills:      []string{"躲藏", "求救"},
		Inventory:   []string{"破舊的護士服", "沾血的繃帶"},
		Secret:      "她曾目睹醫院地下室的非法實驗，但因恐懼而選擇沉默，現在那些記憶不斷折磨著她。",
		Introduction: "王小芳蜷縮在角落，雙手死死抓著破舊的護士服，眼神不斷飄向門口。「救...救命...」聲音顫抖得幾乎聽不清，身體抖得像秋風中的落葉。",
		LinkedSeeds: []string{"GS-001"},
		DeathTiming: 15,
		Status:      agents.NPCStatusAlive,
	}

	// Test ToJSON
	jsonStr, err := profile.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify JSON is valid
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Test FromJSON
	deserialized, err := NPCProfileFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Verify all fields match
	if deserialized.ID != profile.ID {
		t.Errorf("Expected ID %s, got %s", profile.ID, deserialized.ID)
	}
	if deserialized.Name != profile.Name {
		t.Errorf("Expected Name %s, got %s", profile.Name, deserialized.Name)
	}
	if deserialized.Archetype != profile.Archetype {
		t.Errorf("Expected Archetype %s, got %s", profile.Archetype, deserialized.Archetype)
	}
	if len(deserialized.Personality) != len(profile.Personality) {
		t.Errorf("Expected Personality length %d, got %d", len(profile.Personality), len(deserialized.Personality))
	}
}

func TestNPCProfile_GetArchetypeInfo(t *testing.T) {
	profile := &NPCProfile{
		Archetype: agents.NPCArchetypeSacrificial,
	}

	name := profile.GetArchetypeName()
	if name == "" {
		t.Error("Expected non-empty archetype name")
	}
	if !strings.Contains(name, "犧牲者") {
		t.Errorf("Expected archetype name to contain '犧牲者', got '%s'", name)
	}

	desc := profile.GetArchetypeDescription()
	if desc == "" {
		t.Error("Expected non-empty archetype description")
	}
}
