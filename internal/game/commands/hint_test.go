package commands

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/hint"
)

func TestHintCommandName(t *testing.T) {
	cmd := NewHintCommand()
	if cmd.Name() != "hint" {
		t.Errorf("Name() = %s, want 'hint'", cmd.Name())
	}
}

func TestHintCommandHelp(t *testing.T) {
	cmd := NewHintCommand()
	help := cmd.Help()
	if !strings.Contains(help, "SAN") {
		t.Error("Help should mention SAN")
	}
	if !strings.Contains(help, "/hint") {
		t.Error("Help should mention /hint")
	}
}

func TestHintCommandHellModeDenied(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyHell
	}

	result, err := cmd.Execute(nil)

	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if !strings.Contains(result, "沒有人能幫助你") {
		t.Errorf("Hell mode should deny hints, got: %s", result)
	}
	if cmd.IsConfirmMode() {
		t.Error("Should not enter confirm mode in Hell mode")
	}
}

func TestHintCommandHintStateDisabled(t *testing.T) {
	cmd := NewHintCommand()
	cmd.HintState = hint.NewHintState(game.DifficultyHell)
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyHard // Even if difficulty says Hard
	}

	result, err := cmd.Execute(nil)

	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if !strings.Contains(result, "沒有人能幫助你") {
		t.Errorf("Disabled state should deny hints, got: %s", result)
	}
}

func TestHintCommandInsufficientSAN(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 5 // Less than 10
	}
	cmd.GetChapter = func() int {
		return 1
	}

	result, err := cmd.Execute(nil)

	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if !strings.Contains(result, "理智不足") {
		t.Errorf("Should show insufficient SAN message, got: %s", result)
	}
	if !strings.Contains(result, "需要 10 SAN") {
		t.Error("Should show required SAN amount")
	}
	if cmd.IsConfirmMode() {
		t.Error("Should not enter confirm mode with insufficient SAN")
	}
}

func TestHintCommandConfirmationPrompt(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}

	result, err := cmd.Execute(nil)

	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if !strings.Contains(result, "花費 10 SAN") {
		t.Errorf("Should show SAN cost, got: %s", result)
	}
	if !strings.Contains(result, "y") || !strings.Contains(result, "n") {
		t.Error("Should show confirmation options")
	}
	if !cmd.IsConfirmMode() {
		t.Error("Should enter confirm mode")
	}
}

func TestHintCommandConfirmYes(t *testing.T) {
	cmd := NewHintCommand()
	sanDeducted := 0
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		sanDeducted = amount
		return 50 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "廢棄醫院"
	}

	// First execute to enter confirm mode
	_, _ = cmd.Execute(nil)

	// Confirm with "y"
	result, err := cmd.Execute([]string{"y"})

	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if sanDeducted != 10 {
		t.Errorf("Should deduct 10 SAN, deducted %d", sanDeducted)
	}
	if !strings.Contains(result, "消耗 10 SAN") {
		t.Errorf("Should show SAN consumption, got: %s", result)
	}
	if cmd.IsConfirmMode() {
		t.Error("Should exit confirm mode after confirmation")
	}
}

func TestHintCommandConfirmNo(t *testing.T) {
	cmd := NewHintCommand()
	sanDeducted := false
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		sanDeducted = true
		return 50
	}

	// Enter confirm mode
	_, _ = cmd.Execute(nil)

	// Cancel with "n"
	result, err := cmd.Execute([]string{"n"})

	if err != nil {
		t.Errorf("Should not return error: %v", err)
	}
	if sanDeducted {
		t.Error("Should not deduct SAN on cancel")
	}
	if !strings.Contains(result, "取消") {
		t.Errorf("Should show cancellation message, got: %s", result)
	}
	if cmd.IsConfirmMode() {
		t.Error("Should exit confirm mode after cancellation")
	}
}

func TestHintCommandConfirmEmpty(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}

	// Enter confirm mode
	_, _ = cmd.Execute(nil)

	// Empty args = cancel
	result, _ := cmd.Execute(nil)

	if !strings.Contains(result, "取消") {
		t.Errorf("Empty args should cancel, got: %s", result)
	}
}

func TestHintCommandCostIncrement(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 100
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		return 100 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "場景"
	}

	// First hint
	_, _ = cmd.Execute(nil)
	_, _ = cmd.Execute([]string{"y"})

	// Second hint - still 10
	result, _ := cmd.Execute(nil)
	if !strings.Contains(result, "花費 10 SAN") {
		t.Errorf("Second hint should still cost 10, got: %s", result)
	}
	_, _ = cmd.Execute([]string{"y"})

	// Third hint - now 15
	result, _ = cmd.Execute(nil)
	if !strings.Contains(result, "花費 15 SAN") {
		t.Errorf("Third hint should cost 15, got: %s", result)
	}
}

func TestHintCommandUsageCountDisplay(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 100
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		return 100 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "場景"
	}

	// Use one hint
	_, _ = cmd.Execute(nil)
	_, _ = cmd.Execute([]string{"y"})

	// Second hint should show usage
	result, _ := cmd.Execute(nil)
	if !strings.Contains(result, "第 2 次提示") {
		t.Errorf("Should show hint count, got: %s", result)
	}
}

func TestHintCommandWithLLMHint(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		return 50 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "廢棄醫院"
	}
	cmd.GenerateLLMHint = func(prompt string) (string, error) {
		return "這是 LLM 生成的提示文字", nil
	}

	// Enter confirm and execute
	_, _ = cmd.Execute(nil)
	result, _ := cmd.Execute([]string{"y"})

	if !strings.Contains(result, "這是 LLM 生成的提示文字") {
		t.Errorf("Should contain LLM hint, got: %s", result)
	}
}

func TestHintCommandWithLLMFailure(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		return 50 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "廢棄醫院"
	}
	cmd.GenerateLLMHint = func(prompt string) (string, error) {
		return "", nil // Empty = failure
	}

	// Enter confirm and execute
	_, _ = cmd.Execute(nil)
	result, _ := cmd.Execute([]string{"y"})

	// Should still have a hint (fallback)
	if result == "" {
		t.Error("Should provide fallback hint on LLM failure")
	}
	if !strings.Contains(result, "消耗 10 SAN") {
		t.Error("Should still show SAN consumption")
	}
}

func TestHintCommandCancelConfirmation(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}

	// Enter confirm mode
	_, _ = cmd.Execute(nil)
	if !cmd.IsConfirmMode() {
		t.Error("Should be in confirm mode")
	}

	// Cancel manually
	cmd.CancelConfirmation()

	if cmd.IsConfirmMode() {
		t.Error("Should not be in confirm mode after cancel")
	}
	if cmd.GetPendingCost() != 0 {
		t.Error("Pending cost should be 0 after cancel")
	}
}

func TestHintCommandSetHintState(t *testing.T) {
	cmd := NewHintCommand()
	state := hint.NewHintState(game.DifficultyEasy)

	cmd.SetHintState(state)

	if cmd.HintState != state {
		t.Error("HintState not set correctly")
	}
}

func TestHintCommandHintTypeDisplay(t *testing.T) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}
	cmd.DeductSAN = func(amount int) int {
		return 50 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "場景"
	}
	cmd.GetTurnsSinceMove = func() int {
		return 6 // Stuck
	}

	// Execute
	_, _ = cmd.Execute(nil)
	result, _ := cmd.Execute([]string{"y"})

	// Should show direction hint type (since stuck)
	if !strings.Contains(result, "探索方向") {
		t.Errorf("Stuck player should get direction hint, got: %s", result)
	}
}

func TestHintCommandDifferentChapters(t *testing.T) {
	cmd := NewHintCommand()
	chapter := 1
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 100
	}
	cmd.GetChapter = func() int {
		return chapter
	}
	cmd.DeductSAN = func(amount int) int {
		return 100 - amount
	}
	cmd.GetCurrentScene = func() string {
		return "場景"
	}

	// Use hints in chapter 1
	_, _ = cmd.Execute(nil)
	_, _ = cmd.Execute([]string{"y"})
	_, _ = cmd.Execute(nil)
	_, _ = cmd.Execute([]string{"y"})

	// Change to chapter 2
	chapter = 2

	// New chapter, cost should reset to 10
	result, _ := cmd.Execute(nil)
	if !strings.Contains(result, "花費 10 SAN") {
		t.Errorf("New chapter should reset cost, got: %s", result)
	}
}

func TestNewHintCommand(t *testing.T) {
	cmd := NewHintCommand()

	if cmd == nil {
		t.Fatal("NewHintCommand should not return nil")
	}
	if cmd.Generator == nil {
		t.Error("Generator should not be nil")
	}
}

// Benchmark
func BenchmarkHintCommandExecute(b *testing.B) {
	cmd := NewHintCommand()
	cmd.GetDifficulty = func() game.DifficultyLevel {
		return game.DifficultyEasy
	}
	cmd.GetSAN = func() int {
		return 50
	}
	cmd.GetChapter = func() int {
		return 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.confirmMode = false
		_, _ = cmd.Execute(nil)
	}
}
