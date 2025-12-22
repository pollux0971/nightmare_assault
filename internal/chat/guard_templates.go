package chat

// getGuardTemplates returns all dialogue templates for the Guard archetype.
// Guards are protective, disciplined, and direct. They prioritize security and order.
func getGuardTemplates() []DialogueTemplate {
	return []DialogueTemplate{
		// ==================== AGREE (High Trust) ====================
		{
			ID:        "guard_agree_01",
			Category:  CategoryAgree,
			Archetype: "Guard",
			Content:   "收到。我支持你。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},
		{
			ID:        "guard_agree_02",
			Category:  CategoryAgree,
			Archetype: "Guard",
			Content:   "計畫可行。我們行動吧。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(55),
			},
		},
		{
			ID:        "guard_agree_03",
			Category:  CategoryAgree,
			Archetype: "Guard",
			Content:   "你掩護我，我掩護你。開始行動。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(65),
				MaxFear:  intPtr(40),
			},
		},
		{
			ID:        "guard_agree_04",
			Category:  CategoryAgree,
			Archetype: "Guard",
			Content:   "同意。我相信你的判斷。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(60),
			},
		},

		// ==================== DISAGREE (Low Trust) ====================
		{
			ID:        "guard_disagree_01",
			Category:  CategoryDisagree,
			Archetype: "Guard",
			Content:   "否定。那不符合程序。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "guard_disagree_02",
			Category:  CategoryDisagree,
			Archetype: "Guard",
			Content:   "我不喜歡這樣。風險太高了。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
			},
		},
		{
			ID:        "guard_disagree_03",
			Category:  CategoryDisagree,
			Archetype: "Guard",
			Content:   "不行。找別的辦法。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(30),
			},
		},
		{
			ID:        "guard_disagree_04",
			Category:  CategoryDisagree,
			Archetype: "Guard",
			Content:   "這是戰術錯誤。我們需要重新考慮。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(45),
			},
		},

		// ==================== CONFUSED (Hallucination/Incoherent) ====================
		{
			ID:        "guard_confused_01",
			Category:  CategoryConfused,
			Archetype: "Guard",
			Content:   "再說一次？我沒聽懂。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_confused_02",
			Category:  CategoryConfused,
			Archetype: "Guard",
			Content:   "等等，什麼？再說一遍。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_confused_03",
			Category:  CategoryConfused,
			Archetype: "Guard",
			Content:   "這在戰術上說不通。解釋清楚。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_confused_04",
			Category:  CategoryConfused,
			Archetype: "Guard",
			Content:   "我沒聽明白。你確定這樣說嗎？",
			Conditions: TemplateConditions{
				MinStress: intPtr(40),
			},
		},

		// ==================== FEARFUL (High Fear) ====================
		{
			ID:        "guard_fearful_01",
			Category:  CategoryFearful,
			Archetype: "Guard",
			Content:   "我...我不喜歡這種感覺。這裡有問題。",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},
		{
			ID:        "guard_fearful_02",
			Category:  CategoryFearful,
			Archetype: "Guard",
			Content:   "情況很糟。真的很糟。我們需要支援。",
			Conditions: TemplateConditions{
				MinFear:   intPtr(65),
				MinStress: intPtr(50),
			},
		},
		{
			ID:        "guard_fearful_03",
			Category:  CategoryFearful,
			Archetype: "Guard",
			Content:   "我經歷過戰鬥，但這個...這個不一樣。",
			Conditions: TemplateConditions{
				MinFear:   intPtr(70),
				MinStress: intPtr(60),
			},
		},
		{
			ID:        "guard_fearful_04",
			Category:  CategoryFearful,
			Archetype: "Guard",
			Content:   "我受的訓練沒有為這個做準備。我們的處境很危險。",
			Conditions: TemplateConditions{
				MinFear: intPtr(65),
			},
		},

		// ==================== CURIOUS (Interested/Investigating) ====================
		{
			ID:        "guard_curious_01",
			Category:  CategoryCurious,
			Archetype: "Guard",
			Content:   "多說一點。我需要這方面的情報。",
			Conditions: TemplateConditions{
				MaxFear: intPtr(50),
			},
		},
		{
			ID:        "guard_curious_02",
			Category:  CategoryCurious,
			Archetype: "Guard",
			Content:   "有趣。你還觀察到什麼？",
			Conditions: TemplateConditions{
				MinTrust: intPtr(40),
			},
		},
		{
			ID:        "guard_curious_03",
			Category:  CategoryCurious,
			Archetype: "Guard",
			Content:   "我在聽。報告完整狀況。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_curious_04",
			Category:  CategoryCurious,
			Archetype: "Guard",
			Content:   "這可能有用。繼續說。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(45),
			},
		},

		// ==================== DEFENSIVE (Guarded/Suspicious) ====================
		{
			ID:        "guard_defensive_01",
			Category:  CategoryDefensive,
			Archetype: "Guard",
			Content:   "那是機密。我不能討論。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "guard_defensive_02",
			Category:  CategoryDefensive,
			Archetype: "Guard",
			Content:   "你為什麼需要知道這個？說明你的權限等級。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
				MinFear:  intPtr(40),
			},
		},
		{
			ID:        "guard_defensive_03",
			Category:  CategoryDefensive,
			Archetype: "Guard",
			Content:   "我不回答問題。我只提問。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(30),
			},
		},
		{
			ID:        "guard_defensive_04",
			Category:  CategoryDefensive,
			Archetype: "Guard",
			Content:   "退後。你越界了。",
			Conditions: TemplateConditions{
				MaxTrust:  intPtr(35),
				MinStress: intPtr(50),
			},
		},

		// ==================== NEUTRAL (Default responses) ====================
		{
			ID:        "guard_neutral_01",
			Category:  CategoryNeutral,
			Archetype: "Guard",
			Content:   "收到。保持警戒。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_neutral_02",
			Category:  CategoryNeutral,
			Archetype: "Guard",
			Content:   "了解。繼續前進。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_neutral_03",
			Category:  CategoryNeutral,
			Archetype: "Guard",
			Content:   "確認。保持位置。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "guard_neutral_04",
			Category:  CategoryNeutral,
			Archetype: "Guard",
			Content:   "知道了。注意周圍。",
			Conditions: TemplateConditions{},
		},
	}
}
