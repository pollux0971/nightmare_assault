// Package engine provides the story generation engine for Nightmare Assault.
package engine

import (
	"math/rand"
	"time"
)

// FallbackStory represents a pre-written emergency story.
type FallbackStory struct {
	Theme       string
	Content     string
	Choices     []string
}

// FallbackStories provides emergency fallback stories when API fails.
var FallbackStories = []FallbackStory{
	{
		Theme: "hospital",
		Content: `黑暗籠罩著這座廢棄的醫院。你手中的手電筒發出微弱的光芒，勉強照亮前方幾公尺的走廊。

空氣中瀰漫著消毒水和某種腐敗的氣味。牆壁上的油漆已經斑駁脫落，露出底下灰色的水泥。

你來這裡是有原因的——你的朋友三天前在這裡失蹤了。警方說這只是一座普通的廢棄建築，但你知道事情沒那麼簡單。

走廊盡頭，你看到一扇門微微打開著，裡面透出一絲詭異的光芒。同時，你的左側傳來輕微的腳步聲。

---

**選擇：**
1. 走向那扇微開的門
2. 循著腳步聲去探查
3. 先在這裡搜索一下有用的物品`,
		Choices: []string{
			"走向那扇微開的門",
			"循著腳步聲去探查",
			"先在這裡搜索一下有用的物品",
		},
	},
	{
		Theme: "mansion",
		Content: `暴風雨的夜晚，你不得不在這座古老的洋館尋求庇護。

推開沉重的橡木大門，你踏入了一個彷彿時間靜止的世界。水晶吊燈在微風中輕輕搖曳，發出清脆的碰撞聲。大廳的牆上掛滿了油畫，畫中人物的眼睛似乎在注視著你。

「歡迎。」一個沙啞的聲音從樓梯上方傳來。你抬頭看去，卻什麼也看不見——只有無盡的黑暗。

桌上有一封信，信封上寫著你的名字。這不可能……你從未來過這裡。

---

**選擇：**
1. 打開那封信
2. 朝樓梯上方喊話
3. 轉身離開這個詭異的地方`,
		Choices: []string{
			"打開那封信",
			"朝樓梯上方喊話",
			"轉身離開這個詭異的地方",
		},
	},
	{
		Theme: "subway",
		Content: `最後一班列車已經開走了，但你還困在這座地下車站。

月台上的燈光忽明忽暗，投下一片片跳動的陰影。售票機螢幕顯示著亂碼，發出刺耳的電子雜音。

你確定自己看到有人走進了那條標示「禁止進入」的隧道。那個身影……看起來像是你失蹤的妹妹。

但她三年前就已經……

隧道深處傳來列車的轟鳴聲，但時刻表上顯示已經沒有車了。地面開始震動。

---

**選擇：**
1. 追進隧道
2. 試著找其他出口
3. 躲到月台的柱子後面觀察`,
		Choices: []string{
			"追進隧道",
			"試著找其他出口",
			"躲到月台的柱子後面觀察",
		},
	},
	{
		Theme: "forest",
		Content: `你迷失在這片不該存在的森林裡。

GPS顯示你仍在城市中心，但四周只有無盡的樹木。月光穿過枝葉的縫隙，在地上投下詭異的圖案——那些圖案看起來像是某種古老的符文。

遠處，你看到一盞燈火。可能是房子，也可能是……

樹上的烏鴉突然齊聲鳴叫，然後一齊飛走。森林陷入令人窒息的寂靜。你感覺到有什麼東西在樹叢中移動，越來越近。

---

**選擇：**
1. 朝燈火走去
2. 爬上樹躲避
3. 蹲下來研究那些符文`,
		Choices: []string{
			"朝燈火走去",
			"爬上樹躲避",
			"蹲下來研究那些符文",
		},
	},
	{
		Theme: "default",
		Content: `你睜開眼睛，發現自己身處一個完全陌生的地方。

黑暗籠罩著四周，只有微弱的光線從某個方向滲透進來。空氣冰冷，帶著一股霉味。你的頭很痛，完全想不起來自己是怎麼來到這裡的。

你試著移動，發現手腳都能自由活動。口袋裡摸到一支手機，但螢幕碎了，只能發出微弱的光。

遠處傳來一個聲音——像是有人在哭泣，又像是在低聲呢喃。

---

**選擇：**
1. 朝光源的方向走去
2. 朝聲音的來源移動
3. 先待在原地仔細觀察環境`,
		Choices: []string{
			"朝光源的方向走去",
			"朝聲音的來源移動",
			"先待在原地仔細觀察環境",
		},
	},
}

// GetFallbackStory returns a fallback story matching the theme or default.
func GetFallbackStory(theme string) *FallbackStory {
	theme = toLower(theme)

	// Try to match theme keywords
	for i := range FallbackStories {
		if FallbackStories[i].Theme != "default" && contains(theme, FallbackStories[i].Theme) {
			return &FallbackStories[i]
		}
	}

	// Check for Chinese keywords
	themeKeywords := map[string]string{
		"醫院": "hospital",
		"病院": "hospital",
		"洋館": "mansion",
		"豪宅": "mansion",
		"古堡": "mansion",
		"地鐵": "subway",
		"車站": "subway",
		"隧道": "subway",
		"森林": "forest",
		"樹林": "forest",
	}

	for keyword, fallbackTheme := range themeKeywords {
		if containsRune(theme, keyword) {
			for i := range FallbackStories {
				if FallbackStories[i].Theme == fallbackTheme {
					return &FallbackStories[i]
				}
			}
		}
	}

	// Return random non-default or default
	rand.Seed(time.Now().UnixNano())
	nonDefault := make([]*FallbackStory, 0)
	var defaultStory *FallbackStory

	for i := range FallbackStories {
		if FallbackStories[i].Theme == "default" {
			defaultStory = &FallbackStories[i]
		} else {
			nonDefault = append(nonDefault, &FallbackStories[i])
		}
	}

	if len(nonDefault) > 0 {
		return nonDefault[rand.Intn(len(nonDefault))]
	}
	return defaultStory
}

func containsRune(s, substr string) bool {
	return len(s) >= len(substr) && findRuneSubstring(s, substr) >= 0
}

func findRuneSubstring(s, substr string) int {
	sRunes := []rune(s)
	subRunes := []rune(substr)

	for i := 0; i <= len(sRunes)-len(subRunes); i++ {
		match := true
		for j := 0; j < len(subRunes); j++ {
			if sRunes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
