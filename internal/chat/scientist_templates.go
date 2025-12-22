package chat

// getScientistTemplates returns all dialogue templates for the Scientist archetype.
// Scientists are analytical, logical, and prone to overthinking. They value evidence and reasoning.
func getScientistTemplates() []DialogueTemplate {
	return []DialogueTemplate{
		// ==================== AGREE (High Trust) ====================
		{
			ID:        "sci_agree_01",
			Category:  CategoryAgree,
			Archetype: "Scientist",
			Content:   "有道理，我們小心行事。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},
		{
			ID:        "sci_agree_02",
			Category:  CategoryAgree,
			Archetype: "Scientist",
			Content:   "我也這麼想。數據支持你的假設。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(60),
			},
		},
		{
			ID:        "sci_agree_03",
			Category:  CategoryAgree,
			Archetype: "Scientist",
			Content:   "沒錯，這與我的觀察相符。我們應該記錄下來。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(55),
				MaxFear:  intPtr(40),
			},
		},
		{
			ID:        "sci_agree_04",
			Category:  CategoryAgree,
			Archetype: "Scientist",
			Content:   "邏輯上說得通。我同意你的評估。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},

		// ==================== DISAGREE (Low Trust) ====================
		{
			ID:        "sci_disagree_01",
			Category:  CategoryDisagree,
			Archetype: "Scientist",
			Content:   "我認為這不正確。證據不支持這個說法。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "sci_disagree_02",
			Category:  CategoryDisagree,
			Archetype: "Scientist",
			Content:   "這與我的觀察矛盾。你確定嗎？",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(45),
			},
		},
		{
			ID:        "sci_disagree_03",
			Category:  CategoryDisagree,
			Archetype: "Scientist",
			Content:   "我持懷疑態度。你的推論有缺陷。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
			},
		},

		// ==================== CONFUSED (Hallucination/Incoherent) ====================
		{
			ID:        "sci_confused_01",
			Category:  CategoryConfused,
			Archetype: "Scientist",
			Content:   "等等，這說不通。你能再解釋一次嗎？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "sci_confused_02",
			Category:  CategoryConfused,
			Archetype: "Scientist",
			Content:   "我無法理解你的邏輯。你在說什麼？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "sci_confused_03",
			Category:  CategoryConfused,
			Archetype: "Scientist",
			Content:   "這...與數據不一致。你還好嗎？",
			Conditions: TemplateConditions{
				MinStress: intPtr(40),
			},
		},
		{
			ID:        "sci_confused_04",
			Category:  CategoryConfused,
			Archetype: "Scientist",
			Content:   "我一定遺漏了什麼。這對不上。",
			Conditions: TemplateConditions{},
		},

		// ==================== FEARFUL (High Fear) ====================
		{
			ID:        "sci_fearful_01",
			Category:  CategoryFearful,
			Archetype: "Scientist",
			Content:   "這...這完全錯了。我們得離開這裡。",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},
		{
			ID:        "sci_fearful_02",
			Category:  CategoryFearful,
			Archetype: "Scientist",
			Content:   "我無法再合理化這一切了。有什麼根本性的錯誤。",
			Conditions: TemplateConditions{
				MinFear:   intPtr(65),
				MinStress: intPtr(50),
			},
		},
		{
			ID:        "sci_fearful_03",
			Category:  CategoryFearful,
			Archetype: "Scientist",
			Content:   "我的手停不住顫抖。如果我們已經太遲了怎麼辦？",
			Conditions: TemplateConditions{
				MinFear:   intPtr(70),
				MinStress: intPtr(60),
			},
		},
		{
			ID:        "sci_fearful_04",
			Category:  CategoryFearful,
			Archetype: "Scientist",
			Content:   "數字已經不重要了。我們面臨真正的危險。",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},

		// ==================== CURIOUS (Interested/Investigating) ====================
		{
			ID:        "sci_curious_01",
			Category:  CategoryCurious,
			Archetype: "Scientist",
			Content:   "有趣。告訴我更多你觀察到的細節。",
			Conditions: TemplateConditions{
				MaxFear: intPtr(50),
			},
		},
		{
			ID:        "sci_curious_02",
			Category:  CategoryCurious,
			Archetype: "Scientist",
			Content:   "這是個引人入勝的假設。你有佐證嗎？",
			Conditions: TemplateConditions{
				MinTrust: intPtr(40),
				MaxFear:  intPtr(45),
			},
		},
		{
			ID:        "sci_curious_03",
			Category:  CategoryCurious,
			Archetype: "Scientist",
			Content:   "我需要更多數據。你還注意到什麼？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "sci_curious_04",
			Category:  CategoryCurious,
			Archetype: "Scientist",
			Content:   "這可能很重要。讓我想想其中的含義。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(45),
			},
		},

		// ==================== DEFENSIVE (Guarded/Suspicious) ====================
		{
			ID:        "sci_defensive_01",
			Category:  CategoryDefensive,
			Archetype: "Scientist",
			Content:   "我現在不方便透露那些資訊。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "sci_defensive_02",
			Category:  CategoryDefensive,
			Archetype: "Scientist",
			Content:   "你為什麼問我這個？你有什麼目的？",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
				MinFear:  intPtr(40),
			},
		},
		{
			ID:        "sci_defensive_03",
			Category:  CategoryDefensive,
			Archetype: "Scientist",
			Content:   "我已經說夠了。我們應該專注於當前的任務。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(45),
			},
		},
		{
			ID:        "sci_defensive_04",
			Category:  CategoryDefensive,
			Archetype: "Scientist",
			Content:   "這是個帶有陷阱的問題。我不認為該回答。",
			Conditions: TemplateConditions{
				MaxTrust:  intPtr(40),
				MinStress: intPtr(50),
			},
		},

		// ==================== NEUTRAL (Default responses) ====================
		{
			ID:        "sci_neutral_01",
			Category:  CategoryNeutral,
			Archetype: "Scientist",
			Content:   "我明白了。讓我考慮一下。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "sci_neutral_02",
			Category:  CategoryNeutral,
			Archetype: "Scientist",
			Content:   "知道了。我會把這個加入觀察記錄。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "sci_neutral_03",
			Category:  CategoryNeutral,
			Archetype: "Scientist",
			Content:   "了解。我們繼續前進吧。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "sci_neutral_04",
			Category:  CategoryNeutral,
			Archetype: "Scientist",
			Content:   "好的。保持專注。",
			Conditions: TemplateConditions{},
		},
	}
}
