// Package logger 提供遊戲的結構化日誌系統測試
// Story 10-4: 詳細日誌系統測試
package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLogLevel_String 測試日誌級別字符串轉換
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"DEBUG level", DEBUG, "DEBUG"},
		{"INFO level", INFO, "INFO"},
		{"WARN level", WARN, "WARN"},
		{"ERROR level", ERROR, "ERROR"},
		{"Unknown level", LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseLogLevel 測試日誌級別解析
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  LogLevel
		shouldErr bool
	}{
		{"debug lowercase", "debug", DEBUG, false},
		{"DEBUG uppercase", "DEBUG", DEBUG, false},
		{"info lowercase", "info", INFO, false},
		{"INFO uppercase", "INFO", INFO, false},
		{"warn lowercase", "warn", WARN, false},
		{"WARN uppercase", "WARN", WARN, false},
		{"warning lowercase", "warning", WARN, false},
		{"WARNING uppercase", "WARNING", WARN, false},
		{"error lowercase", "error", ERROR, false},
		{"ERROR uppercase", "ERROR", ERROR, false},
		{"invalid level", "invalid", INFO, true},
		{"empty string", "", INFO, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLogLevel(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseLogLevel(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLogLevel(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestNew_CreateLogger 測試創建日誌記錄器
func TestNew_CreateLogger(t *testing.T) {
	// 創建臨時目錄
	tempDir := t.TempDir()

	// 創建日誌記錄器
	logger, err := New(tempDir, INFO, 7)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// 檢查日誌記錄器屬性
	if logger.level != INFO {
		t.Errorf("logger.level = %v, want %v", logger.level, INFO)
	}
	if logger.logDir != tempDir {
		t.Errorf("logger.logDir = %v, want %v", logger.logDir, tempDir)
	}
	if logger.maxRetention != 7 {
		t.Errorf("logger.maxRetention = %v, want %v", logger.maxRetention, 7)
	}

	// 檢查日誌文件是否創建
	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(tempDir, "nightmare-"+today+".log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("log file not created: %s", logPath)
	}
}

// TestLogger_SetLevel 測試設置日誌級別
func TestLogger_SetLevel(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := New(tempDir, INFO, 7)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// 初始級別應該是 INFO
	if logger.GetLevel() != INFO {
		t.Errorf("initial level = %v, want %v", logger.GetLevel(), INFO)
	}

	// 設置為 DEBUG
	logger.SetLevel(DEBUG)
	if logger.GetLevel() != DEBUG {
		t.Errorf("after SetLevel(DEBUG) level = %v, want %v", logger.GetLevel(), DEBUG)
	}

	// 設置為 ERROR
	logger.SetLevel(ERROR)
	if logger.GetLevel() != ERROR {
		t.Errorf("after SetLevel(ERROR) level = %v, want %v", logger.GetLevel(), ERROR)
	}
}

// TestLogger_LogFiltering 測試日誌級別過濾
func TestLogger_LogFiltering(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := New(tempDir, WARN, 7)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// 記錄不同級別的日誌
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// 讀取日誌文件
	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(tempDir, "nightmare-"+today+".log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)

	// DEBUG 和 INFO 應該被過濾掉
	if strings.Contains(contentStr, "debug message") {
		t.Error("DEBUG message should be filtered out")
	}
	if strings.Contains(contentStr, "info message") {
		t.Error("INFO message should be filtered out")
	}

	// WARN 和 ERROR 應該被記錄
	if !strings.Contains(contentStr, "warn message") {
		t.Error("WARN message should be logged")
	}
	if !strings.Contains(contentStr, "error message") {
		t.Error("ERROR message should be logged")
	}
}

// TestLogger_JSONFormat 測試日誌 JSON 格式
func TestLogger_JSONFormat(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := New(tempDir, DEBUG, 7)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// 記錄一條帶上下文的日誌
	logger.Info("test message", map[string]interface{}{
		"user_id": 123,
		"action":  "test_action",
	})

	// 讀取日誌文件
	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(tempDir, "nightmare-"+today+".log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// 解析 JSON
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Fatal("no log entries found")
	}

	var entry LogEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// 檢查日誌條目
	if entry.Level != "INFO" {
		t.Errorf("entry.Level = %v, want %v", entry.Level, "INFO")
	}
	if entry.Message != "test message" {
		t.Errorf("entry.Message = %v, want %v", entry.Message, "test message")
	}
	if entry.Context["user_id"] != float64(123) {
		t.Errorf("entry.Context[user_id] = %v, want %v", entry.Context["user_id"], 123)
	}
	if entry.Context["action"] != "test_action" {
		t.Errorf("entry.Context[action] = %v, want %v", entry.Context["action"], "test_action")
	}
}

// TestLogger_MultipleContexts 測試合併多個上下文
func TestMergeContext(t *testing.T) {
	tests := []struct {
		name     string
		contexts []map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty contexts",
			contexts: []map[string]interface{}{},
			expected: nil,
		},
		{
			name: "single context",
			contexts: []map[string]interface{}{
				{"key1": "value1"},
			},
			expected: map[string]interface{}{"key1": "value1"},
		},
		{
			name: "multiple contexts",
			contexts: []map[string]interface{}{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			expected: map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
		{
			name: "overlapping keys",
			contexts: []map[string]interface{}{
				{"key1": "value1", "key2": "old"},
				{"key2": "new", "key3": "value3"},
			},
			expected: map[string]interface{}{"key1": "value1", "key2": "new", "key3": "value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeContext(tt.contexts...)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("mergeContext() = %v, want nil", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("mergeContext() length = %v, want %v", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("mergeContext()[%q] = %v, want %v", key, result[key], expectedValue)
				}
			}
		})
	}
}

// TestLogger_CleanOldLogs 測試日誌自動清理
func TestLogger_CleanOldLogs(t *testing.T) {
	tempDir := t.TempDir()

	// 創建一些舊的日誌文件
	oldDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	oldLogPath := filepath.Join(tempDir, "nightmare-"+oldDate+".log")
	if err := os.WriteFile(oldLogPath, []byte("old log"), 0644); err != nil {
		t.Fatalf("failed to create old log file: %v", err)
	}

	// 修改文件的修改時間為 10 天前
	oldTime := time.Now().AddDate(0, 0, -10)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to change file time: %v", err)
	}

	// 創建日誌記錄器，保留期為 7 天
	logger, err := New(tempDir, INFO, 7)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// 手動觸發清理
	logger.cleanOldLogs()

	// 檢查舊日誌文件是否被刪除
	if _, err := os.Stat(oldLogPath); !os.IsNotExist(err) {
		t.Error("old log file should be deleted")
	}

	// 檢查今天的日誌文件是否存在
	today := time.Now().Format("2006-01-02")
	todayLogPath := filepath.Join(tempDir, "nightmare-"+today+".log")
	if _, err := os.Stat(todayLogPath); os.IsNotExist(err) {
		t.Error("today's log file should exist")
	}
}

// TestGetLogDir 測試獲取日誌目錄
func TestGetLogDir(t *testing.T) {
	logDir, err := GetLogDir()
	if err != nil {
		t.Fatalf("GetLogDir() error = %v", err)
	}

	// 檢查路徑包含 .nightmare/logs
	if !strings.Contains(logDir, ".nightmare") || !strings.Contains(logDir, "logs") {
		t.Errorf("GetLogDir() = %v, should contain .nightmare/logs", logDir)
	}
}

// TestInitGlobal 測試全局日誌記錄器初始化
func TestInitGlobal(t *testing.T) {
	tempDir := t.TempDir()

	// 初始化全局日誌記錄器
	err := InitGlobal(tempDir, DEBUG, 7)
	if err != nil {
		t.Fatalf("InitGlobal() error = %v", err)
	}

	// 檢查全局日誌記錄器是否可用
	if GetGlobal() == nil {
		t.Error("GetGlobal() should return non-nil logger")
	}

	// 測試全局函數
	Debug("debug test")
	Info("info test")
	Warn("warn test")
	Error("error test")

	// 讀取日誌文件驗證
	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(tempDir, "nightmare-"+today+".log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "debug test") {
		t.Error("debug message not logged")
	}
	if !strings.Contains(contentStr, "info test") {
		t.Error("info message not logged")
	}
	if !strings.Contains(contentStr, "warn test") {
		t.Error("warn message not logged")
	}
	if !strings.Contains(contentStr, "error test") {
		t.Error("error message not logged")
	}
}

// TestLogger_Rotation 測試日誌輪轉
func TestLogger_Rotation(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := New(tempDir, INFO, 7)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	// 記錄一條日誌
	logger.Info("before rotation")

	// 模擬日期變更
	logger.currentDate = "2000-01-01"

	// 記錄另一條日誌，應該觸發輪轉
	logger.Info("after rotation")

	// 檢查今天的日誌文件是否存在
	today := time.Now().Format("2006-01-02")
	todayLogPath := filepath.Join(tempDir, "nightmare-"+today+".log")
	if _, err := os.Stat(todayLogPath); os.IsNotExist(err) {
		t.Error("today's log file should exist after rotation")
	}

	// 讀取今天的日誌文件
	content, err := os.ReadFile(todayLogPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "after rotation") {
		t.Error("log after rotation should be in new file")
	}
}
