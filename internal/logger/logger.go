// Package logger 提供遊戲的結構化日誌系統
// Story 10-4: 詳細日誌系統
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 定義日誌級別
type LogLevel int

const (
	// DEBUG 調試級別 - 詳細的調試信息
	DEBUG LogLevel = iota
	// INFO 信息級別 - 一般信息
	INFO
	// WARN 警告級別 - 警告信息
	WARN
	// ERROR 錯誤級別 - 錯誤信息
	ERROR
)

// String 返回日誌級別的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel 從字符串解析日誌級別
func ParseLogLevel(s string) (LogLevel, error) {
	switch s {
	case "debug", "DEBUG":
		return DEBUG, nil
	case "info", "INFO":
		return INFO, nil
	case "warn", "WARN", "warning", "WARNING":
		return WARN, nil
	case "error", "ERROR":
		return ERROR, nil
	default:
		return INFO, fmt.Errorf("無效的日誌級別: %s", s)
	}
}

// LogEntry 定義單個日誌條目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Logger 結構化日誌記錄器
type Logger struct {
	mu            sync.Mutex
	level         LogLevel
	writer        io.Writer
	file          *os.File
	currentDate   string
	logDir        string
	maxRetention  int  // 日誌保留天數
	debugMode     bool // Story 10-8: Debug mode flag
	debugFile     *os.File // Story 10-8: Separate debug log file
	errorFile     *os.File // Story 10-8: Separate error log file
}

var (
	globalLogger *Logger
	once         sync.Once
)

// New 創建新的日誌記錄器
func New(logDir string, level LogLevel, maxRetention int) (*Logger, error) {
	// 確保日誌目錄存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("創建日誌目錄失敗: %w", err)
	}

	logger := &Logger{
		level:        level,
		logDir:       logDir,
		maxRetention: maxRetention,
		debugMode:    false, // Story 10-8: Default to non-debug mode
	}

	// 初始化日誌文件
	if err := logger.rotateLogFile(); err != nil {
		return nil, err
	}

	// Story 10-8: Open error log file (always active)
	if err := logger.openErrorLogFile(); err != nil {
		return nil, err
	}

	// 清理舊日誌
	go logger.cleanOldLogs()

	return logger, nil
}

// InitGlobal 初始化全局日誌記錄器
func InitGlobal(logDir string, level LogLevel, maxRetention int) error {
	var err error
	once.Do(func() {
		globalLogger, err = New(logDir, level, maxRetention)
	})
	return err
}

// GetGlobal 獲取全局日誌記錄器
func GetGlobal() *Logger {
	return globalLogger
}

// SetGlobal 設置全局日誌記錄器
func SetGlobal(logger *Logger) {
	globalLogger = logger
}

// rotateLogFile 輪轉日誌文件
func (l *Logger) rotateLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	// 如果是同一天，不需要輪轉
	if l.currentDate == today && l.file != nil {
		return nil
	}

	// 關閉舊文件
	if l.file != nil {
		l.file.Close()
	}

	// 創建新文件
	logPath := filepath.Join(l.logDir, fmt.Sprintf("nightmare-%s.log", today))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打開日誌文件失敗: %w", err)
	}

	l.file = file
	l.writer = file
	l.currentDate = today

	return nil
}

// cleanOldLogs 清理過期的日誌文件
func (l *Logger) cleanOldLogs() {
	if l.maxRetention <= 0 {
		return
	}

	files, err := os.ReadDir(l.logDir)
	if err != nil {
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -l.maxRetention)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 檢查文件名格式: nightmare-YYYY-MM-DD.log
		name := file.Name()
		if len(name) < 11 || name[:10] != "nightmare-" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// 檢查文件修改時間
		if info.ModTime().Before(cutoffDate) {
			logPath := filepath.Join(l.logDir, name)
			os.Remove(logPath)
		}
	}
}

// log 寫入日誌條目
// Story 10-8: Routes logs to appropriate files based on debug mode and log level
func (l *Logger) log(level LogLevel, message string, context map[string]interface{}) {
	// 檢查日誌級別
	if level < l.level {
		return
	}

	// 檢查是否需要輪轉日誌
	l.rotateLogFile()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   message,
		Context:   context,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Story 10-8 AC1: In debug mode, write all logs to debug.log
	if l.debugMode && l.debugFile != nil {
		l.debugFile.Write(data)
		l.debugFile.Write([]byte("\n"))
	}

	// Story 10-8 AC2: Always write errors and warnings to error.log
	if (level == ERROR || level == WARN) && l.errorFile != nil {
		l.errorFile.Write(data)
		l.errorFile.Write([]byte("\n"))
	}

	// Also write to main log file (for backward compatibility)
	if l.writer != nil {
		l.writer.Write(data)
		l.writer.Write([]byte("\n"))
	}
}

// Debug 記錄調試級別日誌
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	l.log(DEBUG, message, ctx)
}

// Info 記錄信息級別日誌
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	l.log(INFO, message, ctx)
}

// Warn 記錄警告級別日誌
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	l.log(WARN, message, ctx)
}

// Error 記錄錯誤級別日誌
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	l.log(ERROR, message, ctx)
}

// SetLevel 設置日誌級別
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 獲取當前日誌級別
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Close 關閉日誌記錄器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error

	if l.file != nil {
		if err := l.file.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Story 10-8: Close debug and error log files
	if l.debugFile != nil {
		if err := l.debugFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if l.errorFile != nil {
		if err := l.errorFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// mergeContext 合併上下文
func mergeContext(contexts ...map[string]interface{}) map[string]interface{} {
	if len(contexts) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	for _, ctx := range contexts {
		for k, v := range ctx {
			result[k] = v
		}
	}
	return result
}

// 全局便捷函數

// Debug 全局調試日誌
func Debug(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(message, context...)
	}
}

// Info 全局信息日誌
func Info(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(message, context...)
	}
}

// Warn 全局警告日誌
func Warn(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(message, context...)
	}
}

// Error 全局錯誤日誌
func Error(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(message, context...)
	}
}

// SetLevel 設置全局日誌級別
func SetLevel(level LogLevel) {
	if globalLogger != nil {
		globalLogger.SetLevel(level)
	}
}

// GetLevel 獲取全局日誌級別
func GetLevel() LogLevel {
	if globalLogger != nil {
		return globalLogger.GetLevel()
	}
	return INFO
}

// GetLogDir 獲取日誌目錄
func GetLogDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".nightmare", "logs"), nil
}

// SetDebugMode enables or disables debug mode.
// Story 10-8 AC1: When enabled, logs to debug.log with detailed information
// Story 10-8 AC2: When disabled, only logs errors and warnings to error.log
func (l *Logger) SetDebugMode(enabled bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.debugMode == enabled {
		return nil // No change needed
	}

	l.debugMode = enabled

	if enabled {
		// Open debug log file
		if err := l.openDebugLogFile(); err != nil {
			return err
		}
		// Set log level to DEBUG to capture all logs
		l.level = DEBUG
	} else {
		// Close debug log file
		if l.debugFile != nil {
			l.debugFile.Close()
			l.debugFile = nil
		}
		// Set log level to WARN to only capture warnings and errors
		l.level = WARN
	}

	return nil
}

// IsDebugMode returns whether debug mode is enabled.
func (l *Logger) IsDebugMode() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.debugMode
}

// openDebugLogFile opens the debug log file.
// Story 10-8 AC1: Debug logs go to ~/.nightmare/logs/debug.log
func (l *Logger) openDebugLogFile() error {
	debugPath := filepath.Join(l.logDir, "debug.log")
	file, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打開 debug 日誌文件失敗: %w", err)
	}
	l.debugFile = file
	return nil
}

// openErrorLogFile opens the error log file.
// Story 10-8 AC2: Error logs go to ~/.nightmare/logs/error.log
func (l *Logger) openErrorLogFile() error {
	errorPath := filepath.Join(l.logDir, "error.log")
	file, err := os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打開 error 日誌文件失敗: %w", err)
	}
	l.errorFile = file
	return nil
}

// SetDebugMode sets debug mode for the global logger.
func SetDebugMode(enabled bool) error {
	if globalLogger != nil {
		return globalLogger.SetDebugMode(enabled)
	}
	return nil
}

// IsDebugMode returns whether debug mode is enabled for the global logger.
func IsDebugMode() bool {
	if globalLogger != nil {
		return globalLogger.IsDebugMode()
	}
	return false
}
