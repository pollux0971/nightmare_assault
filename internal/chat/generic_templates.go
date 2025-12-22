package chat

// getGenericTemplates returns universal dialogue templates that work for any archetype.
// These serve as fallbacks when archetype-specific templates don't match.
func getGenericTemplates() []DialogueTemplate {
	return []DialogueTemplate{
		// ==================== AGREE (High Trust) ====================
		{
			ID:        "gen_agree_01",
			Category:  CategoryAgree,
			Archetype: "Any",
			Content:   "我同意。我們就這麼做吧。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},
		{
			ID:        "gen_agree_02",
			Category:  CategoryAgree,
			Archetype: "Any",
			Content:   "你說得對。我支持你。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},
		{
			ID:        "gen_agree_03",
			Category:  CategoryAgree,
			Archetype: "Any",
			Content:   "聽起來合理。我們繼續吧。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(45),
			},
		},

		// ==================== DISAGREE (Low Trust) ====================
		{
			ID:        "gen_disagree_01",
			Category:  CategoryDisagree,
			Archetype: "Any",
			Content:   "我不認為這是個好主意。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "gen_disagree_02",
			Category:  CategoryDisagree,
			Archetype: "Any",
			Content:   "我不同意。我們應該重新考慮。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "gen_disagree_03",
			Category:  CategoryDisagree,
			Archetype: "Any",
			Content:   "不，這行不通。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
			},
		},

		// ==================== CONFUSED (Hallucination/Incoherent) ====================
		{
			ID:        "gen_confused_01",
			Category:  CategoryConfused,
			Archetype: "Any",
			Content:   "什麼？我不明白。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_confused_02",
			Category:  CategoryConfused,
			Archetype: "Any",
			Content:   "這對我來說沒有意義。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_confused_03",
			Category:  CategoryConfused,
			Archetype: "Any",
			Content:   "我搞混了。你能說清楚嗎？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_confused_04",
			Category:  CategoryConfused,
			Archetype: "Any",
			Content:   "等等，你在說什麼？",
			Conditions: TemplateConditions{},
		},

		// ==================== FEARFUL (High Fear) ====================
		{
			ID:        "gen_fearful_01",
			Category:  CategoryFearful,
			Archetype: "Any",
			Content:   "我害怕。這不對勁。",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},
		{
			ID:        "gen_fearful_02",
			Category:  CategoryFearful,
			Archetype: "Any",
			Content:   "有什麼不對。我能感覺到。",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},
		{
			ID:        "gen_fearful_03",
			Category:  CategoryFearful,
			Archetype: "Any",
			Content:   "這很糟糕。真的很糟糕。",
			Conditions: TemplateConditions{
				MinFear:   intPtr(65),
				MinStress: intPtr(50),
			},
		},

		// ==================== CURIOUS (Interested/Investigating) ====================
		{
			ID:        "gen_curious_01",
			Category:  CategoryCurious,
			Archetype: "Any",
			Content:   "多告訴我一些。我有興趣。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_curious_02",
			Category:  CategoryCurious,
			Archetype: "Any",
			Content:   "這很有趣。繼續說。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_curious_03",
			Category:  CategoryCurious,
			Archetype: "Any",
			Content:   "你還能告訴我什麼？",
			Conditions: TemplateConditions{
				MinTrust: intPtr(40),
			},
		},

		// ==================== DEFENSIVE (Guarded/Suspicious) ====================
		{
			ID:        "gen_defensive_01",
			Category:  CategoryDefensive,
			Archetype: "Any",
			Content:   "我不太想談這個。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "gen_defensive_02",
			Category:  CategoryDefensive,
			Archetype: "Any",
			Content:   "你為什麼想知道？",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "gen_defensive_03",
			Category:  CategoryDefensive,
			Archetype: "Any",
			Content:   "這不關你的事。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(30),
			},
		},

		// ==================== NEUTRAL (Default responses - ULTIMATE FALLBACK) ====================
		{
			ID:        "gen_neutral_01",
			Category:  CategoryNeutral,
			Archetype: "Any",
			Content:   "我明白了。讓我想想。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_neutral_02",
			Category:  CategoryNeutral,
			Archetype: "Any",
			Content:   "了解。我們繼續。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_neutral_03",
			Category:  CategoryNeutral,
			Archetype: "Any",
			Content:   "好的。接下來呢？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_neutral_04",
			Category:  CategoryNeutral,
			Archetype: "Any",
			Content:   "好吧。我們繼續前進。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_neutral_05",
			Category:  CategoryNeutral,
			Archetype: "Any",
			Content:   "我聽到了。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "gen_neutral_06",
			Category:  CategoryNeutral,
			Archetype: "Any",
			Content:   "知道了。",
			Conditions: TemplateConditions{},
		},
	}
}
