// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// DebriefSection represents the different sections of the debrief view.
type DebriefSection int

const (
	// SectionSummary is the death summary section.
	SectionSummary DebriefSection = iota
	// SectionRules is the triggered rules section.
	SectionRules
	// SectionClues is the missed clues section.
	SectionClues
	// SectionDecisions is the key decisions section.
	SectionDecisions
	// SectionOptions is the action options section.
	SectionOptions
)

// DebriefAction represents actions the player can take after debrief.
type DebriefAction int

const (
	// DebriefActionRollback returns to the last checkpoint (easy mode only).
	DebriefActionRollback DebriefAction = iota
	// DebriefActionNewGame starts a new game.
	DebriefActionNewGame
	// DebriefActionMenu returns to the main menu.
	DebriefActionMenu
)

// DebriefSelectMsg is sent when a debrief option is selected.
type DebriefSelectMsg struct {
	Action DebriefAction
}

// DebriefModel represents the death debrief view.
type DebriefModel struct {
	data           *game.DebriefData
	width          int
	height         int
	currentSection DebriefSection
	expandedRules  map[int]bool
	expandedClues  map[int]bool
	selectedOption int
	scrollOffset   int
	maxScroll      int
	options        []DebriefAction
}

// NewDebriefModel creates a new debrief view model.
func NewDebriefModel(data *game.DebriefData) DebriefModel {
	// Build options based on difficulty
	options := []DebriefAction{}
	if data != nil && data.CanRollback() {
		options = append(options, DebriefActionRollback)
	}
	options = append(options, DebriefActionNewGame, DebriefActionMenu)

	return DebriefModel{
		data:           data,
		currentSection: SectionSummary,
		expandedRules:  make(map[int]bool),
		expandedClues:  make(map[int]bool),
		selectedOption: 0,
		options:        options,
	}
}

// Init initializes the debrief view.
func (m DebriefModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the debrief view.
func (m DebriefModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.calculateMaxScroll()

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentSection == SectionOptions {
				if m.selectedOption > 0 {
					m.selectedOption--
				} else {
					// Go back to decisions section
					m.currentSection = SectionDecisions
				}
			} else if m.currentSection > SectionSummary {
				m.currentSection--
			}

		case "down", "j":
			if m.currentSection == SectionOptions {
				if m.selectedOption < len(m.options)-1 {
					m.selectedOption++
				}
			} else if m.currentSection < SectionOptions {
				m.currentSection++
				if m.currentSection == SectionOptions {
					m.selectedOption = 0
				}
			}

		case "tab":
			// Cycle through sections
			if m.currentSection < SectionOptions {
				m.currentSection++
				if m.currentSection == SectionOptions {
					m.selectedOption = 0
				}
			} else {
				m.currentSection = SectionSummary
			}

		case "shift+tab":
			// Cycle backwards through sections
			if m.currentSection > SectionSummary {
				m.currentSection--
			} else {
				m.currentSection = SectionOptions
			}

		case "enter", " ":
			if m.currentSection == SectionOptions {
				return m, m.selectOption()
			}
			// Toggle expansion for rules/clues
			if m.currentSection == SectionRules {
				m.toggleRuleExpansion(m.getSelectedRuleIndex())
			} else if m.currentSection == SectionClues {
				m.toggleClueExpansion(m.getSelectedClueIndex())
			}

		case "e":
			// Expand all in current section
			if m.currentSection == SectionRules {
				for i := range m.data.TriggeredRules {
					m.expandedRules[i] = true
				}
			} else if m.currentSection == SectionClues {
				for i := range m.data.AllClues {
					m.expandedClues[i] = true
				}
			}

		case "c":
			// Collapse all in current section
			if m.currentSection == SectionRules {
				m.expandedRules = make(map[int]bool)
			} else if m.currentSection == SectionClues {
				m.expandedClues = make(map[int]bool)
			}

		case "q", "esc":
			return m, func() tea.Msg {
				return DebriefSelectMsg{Action: DebriefActionMenu}
			}

		case "pgup":
			m.scrollOffset -= 5
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}

		case "pgdown":
			m.scrollOffset += 5
			if m.scrollOffset > m.maxScroll {
				m.scrollOffset = m.maxScroll
			}
		}
	}

	return m, nil
}

// selectOption returns a command for the selected action.
func (m DebriefModel) selectOption() tea.Cmd {
	if m.selectedOption >= 0 && m.selectedOption < len(m.options) {
		action := m.options[m.selectedOption]
		return func() tea.Msg {
			return DebriefSelectMsg{Action: action}
		}
	}
	return nil
}

// toggleRuleExpansion toggles the expansion state of a rule.
func (m *DebriefModel) toggleRuleExpansion(index int) {
	if index >= 0 {
		m.expandedRules[index] = !m.expandedRules[index]
	}
}

// toggleClueExpansion toggles the expansion state of a clue.
func (m *DebriefModel) toggleClueExpansion(index int) {
	if index >= 0 {
		m.expandedClues[index] = !m.expandedClues[index]
	}
}

// getSelectedRuleIndex returns the currently focused rule index.
func (m DebriefModel) getSelectedRuleIndex() int {
	// For now, return 0; could add more sophisticated selection
	if m.data != nil && len(m.data.TriggeredRules) > 0 {
		return 0
	}
	return -1
}

// getSelectedClueIndex returns the currently focused clue index.
func (m DebriefModel) getSelectedClueIndex() int {
	if m.data != nil && len(m.data.GetMissedClues()) > 0 {
		return 0
	}
	return -1
}

// calculateMaxScroll calculates the maximum scroll offset.
func (m *DebriefModel) calculateMaxScroll() {
	// Calculate based on content height vs view height
	contentHeight := m.estimateContentHeight()
	viewHeight := m.height - 6 // Account for header/footer
	if contentHeight > viewHeight {
		m.maxScroll = contentHeight - viewHeight
	} else {
		m.maxScroll = 0
	}
}

// estimateContentHeight estimates the total content height.
func (m DebriefModel) estimateContentHeight() int {
	height := 10 // Base height for summary
	if m.data != nil {
		height += len(m.data.TriggeredRules) * 8
		height += len(m.data.GetMissedClues()) * 4
		height += len(m.data.GetSignificantDecisions()) * 3
	}
	return height
}

// View renders the debrief screen.
func (m DebriefModel) View() string {
	if m.data == nil {
		return m.renderNoData()
	}

	theme := themes.GetManager().GetCurrentTheme()

	// Container style
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(lipgloss.Color("#1a1a2e"))

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width).
		MarginTop(1).
		MarginBottom(1)

	title := titleStyle.Render("═══ 死亡覆盤 ═══")

	// Build content sections
	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	// Section 1: Death Summary
	content.WriteString(m.renderSummarySection())
	content.WriteString("\n")

	// Section 2: Triggered Rules
	content.WriteString(m.renderRulesSection())
	content.WriteString("\n")

	// Section 3: Missed Clues
	content.WriteString(m.renderCluesSection())
	content.WriteString("\n")

	// Section 4: Key Decisions
	content.WriteString(m.renderDecisionsSection())
	content.WriteString("\n")

	// Section 5: Options
	content.WriteString(m.renderOptionsSection())

	// Footer
	footer := m.renderFooter()

	// Combine
	mainContent := content.String()

	// Apply vertical centering/scrolling
	contentLines := strings.Split(mainContent, "\n")
	viewableLines := m.height - 4 // Account for footer
	if len(contentLines) > viewableLines && m.scrollOffset < len(contentLines)-viewableLines {
		start := m.scrollOffset
		end := start + viewableLines
		if end > len(contentLines) {
			end = len(contentLines)
		}
		contentLines = contentLines[start:end]
	}

	finalContent := strings.Join(contentLines, "\n") + "\n" + footer

	return containerStyle.Render(finalContent)
}

// renderNoData renders a fallback when no debrief data is available.
func (m DebriefModel) renderNoData() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(lipgloss.Color("#888888"))

	return style.Render("沒有可用的覆盤資料")
}

// renderSummarySection renders the death summary section.
func (m DebriefModel) renderSummarySection() string {
	theme := themes.GetManager().GetCurrentTheme()
	isActive := m.currentSection == SectionSummary

	headerStyle := m.getSectionHeaderStyle(isActive)
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Width(min(m.width-4, 80)).
		PaddingLeft(2)

	var b strings.Builder
	b.WriteString(headerStyle.Render("【死因摘要】"))
	b.WriteString("\n")

	summary := m.data.GetDeathSummary()
	b.WriteString(contentStyle.Render(summary))

	// Add death details
	if m.data.DeathInfo != nil {
		detailStyle := lipgloss.NewStyle().
			Foreground(theme.Colors.Secondary).
			PaddingLeft(2)

		details := fmt.Sprintf("\n章節：%d  HP：%d  SAN：%d",
			m.data.DeathInfo.Chapter,
			m.data.DeathInfo.FinalHP,
			m.data.DeathInfo.FinalSAN)
		b.WriteString(detailStyle.Render(details))
	}

	return b.String()
}

// renderRulesSection renders the triggered rules section.
func (m DebriefModel) renderRulesSection() string {
	isActive := m.currentSection == SectionRules
	headerStyle := m.getSectionHeaderStyle(isActive)

	var b strings.Builder
	b.WriteString(headerStyle.Render("【觸發的規則】"))
	b.WriteString("\n")

	if len(m.data.TriggeredRules) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			PaddingLeft(2).
			Italic(true)
		b.WriteString(emptyStyle.Render("（沒有觸發任何規則）"))
		return b.String()
	}

	for i, rule := range m.data.TriggeredRules {
		b.WriteString(m.renderRuleItem(i, rule, m.expandedRules[i]))
		b.WriteString("\n")
	}

	return b.String()
}

// renderRuleItem renders a single rule item.
func (m DebriefModel) renderRuleItem(index int, rule *game.RuleReveal, expanded bool) string {
	theme := themes.GetManager().GetCurrentTheme()

	// Determine icon based on consequence
	icon := "▸"
	if expanded {
		icon = "▾"
	}

	// Rule type badge
	badgeStyle := lipgloss.NewStyle().
		Background(theme.Colors.Secondary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	typeStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Error).
		Bold(true)

	ruleHeader := fmt.Sprintf("%s [%d] %s：%s",
		icon, index+1,
		badgeStyle.Render(rule.RuleType),
		typeStyle.Render(rule.TriggerCondition))

	headerStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	var b strings.Builder
	b.WriteString(headerStyle.Render(ruleHeader))

	if expanded {
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			PaddingLeft(6)

		// Consequence
		b.WriteString("\n")
		b.WriteString(detailStyle.Render(fmt.Sprintf("後果：%s", rule.ConsequenceType)))

		// Discovered clues
		if len(rule.DiscoveredClues) > 0 {
			b.WriteString("\n")
			discoveredStyle := lipgloss.NewStyle().
				Foreground(theme.Colors.Success).
				PaddingLeft(6)
			b.WriteString(discoveredStyle.Render("✓ 已發現的線索："))
			for _, clue := range rule.DiscoveredClues {
				b.WriteString("\n")
				b.WriteString(detailStyle.Render("  • " + clue))
			}
		}

		// Missed clues
		if len(rule.MissedClues) > 0 {
			b.WriteString("\n")
			missedStyle := lipgloss.NewStyle().
				Foreground(theme.Colors.Error).
				PaddingLeft(6)
			b.WriteString(missedStyle.Render("✗ 錯過的線索："))
			for _, clue := range rule.MissedClues {
				b.WriteString("\n")
				b.WriteString(detailStyle.Render("  • " + clue))
			}
		}

		// Explanation
		if rule.Explanation != "" {
			b.WriteString("\n")
			explainStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Italic(true).
				PaddingLeft(6)
			b.WriteString(explainStyle.Render("💡 " + rule.Explanation))
		}
	}

	return b.String()
}

// renderCluesSection renders the missed clues section.
func (m DebriefModel) renderCluesSection() string {
	isActive := m.currentSection == SectionClues
	headerStyle := m.getSectionHeaderStyle(isActive)

	var b strings.Builder
	b.WriteString(headerStyle.Render("【錯過的線索】"))
	b.WriteString("\n")

	missedClues := m.data.GetMissedClues()
	if len(missedClues) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			PaddingLeft(2).
			Italic(true)
		b.WriteString(emptyStyle.Render("（沒有錯過任何線索）"))
		return b.String()
	}

	for i, clue := range missedClues {
		b.WriteString(m.renderClueItem(i, clue, m.expandedClues[i]))
		b.WriteString("\n")
	}

	return b.String()
}

// renderClueItem renders a single clue item.
func (m DebriefModel) renderClueItem(index int, clue *game.ClueInfo, expanded bool) string {
	theme := themes.GetManager().GetCurrentTheme()

	icon := "▸"
	if expanded {
		icon = "▾"
	}

	clueStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Warning).
		PaddingLeft(2)

	var b strings.Builder
	header := fmt.Sprintf("%s [第%d章] %s", icon, clue.Chapter, clue.Content)
	b.WriteString(clueStyle.Render(header))

	if expanded && clue.Context != "" {
		contextStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			PaddingLeft(6)
		b.WriteString("\n")
		b.WriteString(contextStyle.Render("上下文：" + clue.Context))
	}

	return b.String()
}

// renderDecisionsSection renders the key decisions section.
func (m DebriefModel) renderDecisionsSection() string {
	isActive := m.currentSection == SectionDecisions
	headerStyle := m.getSectionHeaderStyle(isActive)

	var b strings.Builder
	b.WriteString(headerStyle.Render("【關鍵決策點】"))
	b.WriteString("\n")

	decisions := m.data.GetSignificantDecisions()
	if len(decisions) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			PaddingLeft(2).
			Italic(true)
		b.WriteString(emptyStyle.Render("（沒有重大決策記錄）"))
		return b.String()
	}

	theme := themes.GetManager().GetCurrentTheme()
	for _, decision := range decisions {
		decisionStyle := lipgloss.NewStyle().
			PaddingLeft(2)

		// Mark hallucination decisions
		prefix := "•"
		if decision.IsHallucination {
			prefix = "👁"
		}

		text := fmt.Sprintf("%s [第%d章] 選擇了「%s」", prefix, decision.Chapter, decision.SelectedText)
		b.WriteString(decisionStyle.Render(text))

		if decision.IsHallucination {
			hallucinationStyle := lipgloss.NewStyle().
				Foreground(theme.Colors.Error).
				PaddingLeft(6)
			b.WriteString("\n")
			b.WriteString(hallucinationStyle.Render("⚠ 這是一個幻覺選項"))
		}

		if decision.Consequence != "" {
			consequenceStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Italic(true).
				PaddingLeft(6)
			b.WriteString("\n")
			b.WriteString(consequenceStyle.Render("→ " + decision.Consequence))
		}
		b.WriteString("\n")
	}

	// Hallucination summary
	if len(m.data.HallucinationLogs) > 0 {
		b.WriteString("\n")
		summaryStyle := lipgloss.NewStyle().
			Foreground(theme.Colors.Warning).
			PaddingLeft(2)
		b.WriteString(summaryStyle.Render(
			fmt.Sprintf("共選擇了 %d 個幻覺選項", len(m.data.HallucinationLogs))))
	}

	return b.String()
}

// renderOptionsSection renders the action options section.
func (m DebriefModel) renderOptionsSection() string {
	isActive := m.currentSection == SectionOptions
	theme := themes.GetManager().GetCurrentTheme()

	var b strings.Builder

	// Separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Align(lipgloss.Center).
		Width(m.width)
	b.WriteString(separatorStyle.Render("─────────────────"))
	b.WriteString("\n\n")

	optionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		PaddingLeft(4)

	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Bold(true).
		PaddingLeft(2)

	for i, action := range m.options {
		text := m.getActionText(action)
		if isActive && i == m.selectedOption {
			b.WriteString(selectedStyle.Render("> " + text))
		} else {
			b.WriteString(optionStyle.Render("  " + text))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// getActionText returns the display text for an action.
func (m DebriefModel) getActionText(action DebriefAction) string {
	switch action {
	case DebriefActionRollback:
		if cp := m.data.GetLatestCheckpoint(); cp != nil {
			return fmt.Sprintf("回溯重試（第%d章）", cp.Chapter)
		}
		return "回溯重試"
	case DebriefActionNewGame:
		return "開始新遊戲"
	case DebriefActionMenu:
		return "返回主選單"
	default:
		return "未知選項"
	}
}

// getSectionHeaderStyle returns the style for a section header.
func (m DebriefModel) getSectionHeaderStyle(isActive bool) lipgloss.Style {
	theme := themes.GetManager().GetCurrentTheme()

	if isActive {
		return lipgloss.NewStyle().
			Foreground(theme.Colors.Accent).
			Bold(true).
			PaddingLeft(1).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.Colors.Accent)
	}

	return lipgloss.NewStyle().
		Foreground(theme.Colors.Primary).
		PaddingLeft(2)
}

// renderFooter renders the footer with controls.
func (m DebriefModel) renderFooter() string {
	theme := themes.GetManager().GetCurrentTheme()

	footerStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Secondary).
		Align(lipgloss.Center).
		Width(m.width)

	controls := "↑↓ 導航 | Tab 切換區塊 | Enter 展開/選擇 | E 全部展開 | C 全部收合 | Esc 返回"

	return footerStyle.Render(controls)
}

// SetData updates the debrief data.
func (m *DebriefModel) SetData(data *game.DebriefData) {
	m.data = data
	// Rebuild options
	m.options = []DebriefAction{}
	if data != nil && data.CanRollback() {
		m.options = append(m.options, DebriefActionRollback)
	}
	m.options = append(m.options, DebriefActionNewGame, DebriefActionMenu)
	m.selectedOption = 0
}

// SetSize sets the view dimensions.
func (m *DebriefModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.calculateMaxScroll()
}

// GetCurrentSection returns the current section.
func (m DebriefModel) GetCurrentSection() DebriefSection {
	return m.currentSection
}

// GetSelectedOption returns the selected option index.
func (m DebriefModel) GetSelectedOption() int {
	return m.selectedOption
}
