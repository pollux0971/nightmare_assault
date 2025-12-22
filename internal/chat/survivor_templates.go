package chat

// getSurvivorTemplates returns all dialogue templates for the Survivor archetype.
// Survivors are resilient, emotional, and adaptable. They've endured hardship and show it.
func getSurvivorTemplates() []DialogueTemplate {
	return []DialogueTemplate{
		// ==================== AGREE (High Trust) ====================
		{
			ID:        "surv_agree_01",
			Category:  CategoryAgree,
			Archetype: "Survivor",
			Content:   "對，我覺得你說得對。我們一起行動吧。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},
		{
			ID:        "surv_agree_02",
			Category:  CategoryAgree,
			Archetype: "Survivor",
			Content:   "好。這次我相信你。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(55),
			},
		},
		{
			ID:        "surv_agree_03",
			Category:  CategoryAgree,
			Archetype: "Survivor",
			Content:   "終於有人懂了。謝謝你。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(60),
				MaxFear:  intPtr(40),
			},
		},
		{
			ID:        "surv_agree_04",
			Category:  CategoryAgree,
			Archetype: "Survivor",
			Content:   "這...這也是我一直在想的。很高興你也這麼看。",
			Conditions: TemplateConditions{
				MinTrust: intPtr(50),
			},
		},

		// ==================== DISAGREE (Low Trust) ====================
		{
			ID:        "surv_disagree_01",
			Category:  CategoryDisagree,
			Archetype: "Survivor",
			Content:   "不。我能活到現在是因為相信直覺，而直覺說不行。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "surv_disagree_02",
			Category:  CategoryDisagree,
			Archetype: "Survivor",
			Content:   "我不這麼認為。我見過那樣想的人下場如何。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
			},
		},
		{
			ID:        "surv_disagree_03",
			Category:  CategoryDisagree,
			Archetype: "Survivor",
			Content:   "你錯了。我吃過苦頭才學會不冒這種險。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(45),
			},
		},
		{
			ID:        "surv_disagree_04",
			Category:  CategoryDisagree,
			Archetype: "Survivor",
			Content:   "這不可能。找別的選擇。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(30),
			},
		},

		// ==================== CONFUSED (Hallucination/Incoherent) ====================
		{
			ID:        "surv_confused_01",
			Category:  CategoryConfused,
			Archetype: "Survivor",
			Content:   "等等，什麼？這完全說不通。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_confused_02",
			Category:  CategoryConfused,
			Archetype: "Survivor",
			Content:   "我...我沒聽懂。能再說一次嗎？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_confused_03",
			Category:  CategoryConfused,
			Archetype: "Survivor",
			Content:   "這對不上啊。你到底想說什麼？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_confused_04",
			Category:  CategoryConfused,
			Archetype: "Survivor",
			Content:   "慢著。你還好嗎？聽起來怪怪的。",
			Conditions: TemplateConditions{
				MinStress: intPtr(40),
			},
		},

		// ==================== FEARFUL (High Fear) ====================
		{
			ID:        "surv_fearful_01",
			Category:  CategoryFearful,
			Archetype: "Survivor",
			Content:   "天啊，天啊。太多了。我撐不住...",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},
		{
			ID:        "surv_fearful_02",
			Category:  CategoryFearful,
			Archetype: "Survivor",
			Content:   "我已經走到這一步了，但...如果這就是盡頭呢？如果我們逃不出去呢？",
			Conditions: TemplateConditions{
				MinFear:   intPtr(65),
				MinStress: intPtr(50),
			},
		},
		{
			ID:        "surv_fearful_03",
			Category:  CategoryFearful,
			Archetype: "Survivor",
			Content:   "我很害怕。真的很害怕。拜託告訴我我們會沒事的。",
			Conditions: TemplateConditions{
				MinFear:   intPtr(70),
				MinStress: intPtr(60),
			},
		},
		{
			ID:        "surv_fearful_04",
			Category:  CategoryFearful,
			Archetype: "Survivor",
			Content:   "這感覺不對。一切都不對。我們得逃。",
			Conditions: TemplateConditions{
				MinFear: intPtr(60),
			},
		},

		// ==================== CURIOUS (Interested/Investigating) ====================
		{
			ID:        "surv_curious_01",
			Category:  CategoryCurious,
			Archetype: "Survivor",
			Content:   "真的嗎？多告訴我一些。也許能幫上忙。",
			Conditions: TemplateConditions{
				MaxFear: intPtr(50),
			},
		},
		{
			ID:        "surv_curious_02",
			Category:  CategoryCurious,
			Archetype: "Survivor",
			Content:   "我在聽。你還知道什麼？",
			Conditions: TemplateConditions{
				MinTrust: intPtr(40),
			},
		},
		{
			ID:        "surv_curious_03",
			Category:  CategoryCurious,
			Archetype: "Survivor",
			Content:   "這挺有意思的。繼續說，我想聽。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_curious_04",
			Category:  CategoryCurious,
			Archetype: "Survivor",
			Content:   "嗯。我從來沒這麼想過。是什麼讓你想到這個的？",
			Conditions: TemplateConditions{
				MinTrust: intPtr(45),
			},
		},

		// ==================== DEFENSIVE (Guarded/Suspicious) ====================
		{
			ID:        "surv_defensive_01",
			Category:  CategoryDefensive,
			Archetype: "Survivor",
			Content:   "我沒必要告訴你任何事。離我遠點。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(40),
			},
		},
		{
			ID:        "surv_defensive_02",
			Category:  CategoryDefensive,
			Archetype: "Survivor",
			Content:   "你為什麼問這個？你到底想從我這裡得到什麼？",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(35),
				MinFear:  intPtr(40),
			},
		},
		{
			ID:        "surv_defensive_03",
			Category:  CategoryDefensive,
			Archetype: "Survivor",
			Content:   "我已經分享夠多了。有些事我要留給自己。",
			Conditions: TemplateConditions{
				MaxTrust: intPtr(45),
			},
		},
		{
			ID:        "surv_defensive_04",
			Category:  CategoryDefensive,
			Archetype: "Survivor",
			Content:   "那是私事。別管了。",
			Conditions: TemplateConditions{
				MaxTrust:  intPtr(30),
				MinStress: intPtr(50),
			},
		},

		// ==================== NEUTRAL (Default responses) ====================
		{
			ID:        "surv_neutral_01",
			Category:  CategoryNeutral,
			Archetype: "Survivor",
			Content:   "好吧。我聽到了。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_neutral_02",
			Category:  CategoryNeutral,
			Archetype: "Survivor",
			Content:   "行。我們繼續走吧。",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_neutral_03",
			Category:  CategoryNeutral,
			Archetype: "Survivor",
			Content:   "知道了。接下來呢？",
			Conditions: TemplateConditions{},
		},
		{
			ID:        "surv_neutral_04",
			Category:  CategoryNeutral,
			Archetype: "Survivor",
			Content:   "好吧。我們該走了。",
			Conditions: TemplateConditions{},
		},
	}
}
