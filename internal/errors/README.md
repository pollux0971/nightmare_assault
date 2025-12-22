# 錯誤處理系統 (Error Handling System)

## 概述

這個套件提供了 Nightmare Assault 遊戲的統一錯誤處理系統，支援：
- 統一的錯誤類型分類
- 用戶友善的中文錯誤訊息
- 自動重試建議
- 錯誤上下文追蹤
- 與現有 API 客戶端的無縫整合

## 錯誤類型

### ErrorType 常量

```go
const (
    ErrorTypeUnknown              // 未知錯誤
    ErrorTypeNetwork              // 網路錯誤
    ErrorTypeAPI                  // API錯誤
    ErrorTypeAuth                 // 認證錯誤
    ErrorTypeRateLimit            // 請求限制
    ErrorTypeSaveCorrupt          // 存檔損壞
    ErrorTypeSaveNotFound         // 存檔不存在
    ErrorTypeConfig               // 配置錯誤
    ErrorTypeTimeout              // 連線逾時
    ErrorTypeServiceUnavailable   // 服務不可用
)
```

## 核心功能

### 1. GameError 結構

```go
type GameError struct {
    Type        ErrorType              // 錯誤類型
    Message     string                 // 技術訊息
    UserMessage string                 // 用戶友善訊息
    Suggestion  string                 // 建議操作
    Retryable   bool                   // 是否可重試
    Err         error                  // 原始錯誤
    Context     map[string]interface{} // 額外上下文
}
```

### 2. 創建錯誤

#### 網路錯誤
```go
err := errors.NewNetworkError("無法連接到伺服器", originalErr)
```

#### API 錯誤
```go
err := errors.NewAPIError("API 請求失敗", originalErr, "openai")
```

#### 認證錯誤
```go
err := errors.NewAuthError("API Key 無效", originalErr)
```

#### 速率限制錯誤
```go
err := errors.NewRateLimitError("請求過於頻繁", originalErr)
```

#### 存檔損壞錯誤
```go
err := errors.NewSaveCorruptError("校驗碼不符", originalErr, slotID)
```

#### 存檔不存在錯誤
```go
err := errors.NewSaveNotFoundError("檔案不存在", slotID)
```

#### 配置錯誤
```go
err := errors.NewConfigError("配置格式錯誤", originalErr)
```

#### 超時錯誤
```go
err := errors.NewTimeoutError("請求超時", originalErr)
```

#### 服務不可用錯誤
```go
err := errors.NewServiceUnavailableError("服務維護中", originalErr)
```

### 3. 錯誤適配器

#### 適配 API 錯誤
```go
// 將 API 客戶端錯誤轉換為 GameError
gameErr := errors.AdaptAPIError(apiErr, "openai")
```

#### 適配存檔錯誤
```go
// 將存檔錯誤轉換為 GameError
gameErr := errors.AdaptSaveError(saveErr, slotID)
```

### 4. 錯誤包裝

```go
// 包裝現有錯誤
gameErr := errors.WrapError(
    originalErr,
    errors.ErrorTypeNetwork,
    "網路連線失敗"
)
```

### 5. 錯誤查詢

#### 檢查是否可重試
```go
if errors.IsRetryable(err) {
    // 顯示重試選項
}
```

#### 取得用戶訊息
```go
userMsg := errors.GetUserMessage(err)
```

#### 取得建議
```go
suggestion := errors.GetSuggestion(err)
```

#### 取得完整訊息
```go
fullMsg := errors.GetFullMessage(err)  // 包含訊息和建議
```

### 6. 重試提示

```go
// 格式化重試提示
if errors.ShouldShowRetryOption(err) {
    prompt := errors.FormatRetryPrompt(err)
    // 顯示: "錯誤訊息\n\n建議\n\n是否要重試？(y/n)"
}
```

## 使用範例

### API 錯誤處理

```go
import (
    "github.com/nightmare-assault/nightmare-assault/internal/errors"
)

func callLLMAPI(provider string) error {
    // 模擬 API 呼叫
    apiErr := client.SendMessage(ctx, messages)
    if apiErr != nil {
        // 轉換為 GameError
        gameErr := errors.AdaptAPIError(apiErr, provider)

        // 顯示用戶友善訊息
        fmt.Println(gameErr.GetUserMessage())

        // 顯示建議
        if suggestion := gameErr.GetSuggestion(); suggestion != "" {
            fmt.Println(suggestion)
        }

        // 檢查是否可重試
        if gameErr.IsRetryable() {
            // 提供重試選項
            return askUserToRetry(gameErr)
        }

        return gameErr
    }
    return nil
}
```

### 存檔錯誤處理

```go
func loadSaveFile(slotID int) error {
    saveFile, err := savefile.LoadV2(saveDir, slotID)
    if err != nil {
        // 轉換為 GameError
        gameErr := errors.AdaptSaveError(err, slotID)

        // 顯示完整訊息（包含建議）
        fmt.Println(gameErr.GetFullMessage())

        // 根據錯誤類型採取行動
        switch gameErr.Type {
        case errors.ErrorTypeSaveNotFound:
            // 引導用戶選擇其他槽位或開始新遊戲
            return promptNewGameOrOtherSlot()
        case errors.ErrorTypeSaveCorrupt:
            // 建議開始新遊戲
            return promptStartNewGame()
        default:
            return gameErr
        }
    }
    return nil
}
```

### 統一錯誤顯示

```go
func displayError(err error) {
    if err == nil {
        return
    }

    // 取得用戶訊息
    userMsg := errors.GetUserMessage(err)
    fmt.Printf("❌ %s\n", userMsg)

    // 取得建議
    if suggestion := errors.GetSuggestion(err); suggestion != "" {
        fmt.Printf("\n💡 %s\n", suggestion)
    }

    // 檢查是否提供重試選項
    if errors.ShouldShowRetryOption(err) {
        fmt.Println("\n是否要重試？(y/n)")
    }
}
```

## Story 10-6 驗收標準實作

### AC1: 網路請求失敗

```go
// 當 LLM API 調用超時或錯誤時
err := client.SendMessage(ctx, messages)
if err != nil {
    gameErr := errors.AdaptAPIError(err, provider)

    // 顯示訊息: "無法連接到 LLM 服務，請檢查網路連接與 API Key 設定。"
    fmt.Println(gameErr.GetUserMessage())

    // 建議: "建議操作：\n• 輸入 /api 重新配置 API Key\n• ..."
    fmt.Println(gameErr.GetSuggestion())

    // 提供自動重試選項
    if gameErr.IsRetryable() {
        fmt.Println(errors.FormatRetryPrompt(gameErr))
    }
}
```

### AC2: 存檔損壞

```go
// 當讀檔失敗時
_, err := savefile.LoadV2(saveDir, slotID)
if err != nil {
    gameErr := errors.AdaptSaveError(err, slotID)

    // 顯示訊息: "存檔文件損壞或版本不兼容。"
    fmt.Println(gameErr.GetUserMessage())

    // 建議: "建議開始新遊戲或嘗試其他存檔槽。"
    fmt.Println(gameErr.GetSuggestion())

    // 返回主選單
    return returnToMainMenu()
}
```

## 測試覆蓋

- `types_test.go`: 核心錯誤類型測試
- `adapters_test.go`: 錯誤適配器測試
- `example_test.go`: 使用範例和文檔測試

執行測試:
```bash
go test ./internal/errors/... -v
go test ./internal/errors/... -cover
```

## 設計原則

1. **統一介面**: 所有錯誤都通過 GameError 提供一致的介面
2. **用戶友善**: 所有用戶訊息都使用繁體中文，避免技術術語
3. **可操作性**: 每個錯誤都提供具體的建議行動
4. **可重試性**: 明確標記哪些錯誤可以重試
5. **上下文保留**: 保留原始錯誤和相關上下文資訊
6. **向後兼容**: 與現有 API 客戶端錯誤系統無縫整合

## 錯誤訊息語氣指南

所有錯誤訊息應該：
- 使用繁體中文
- 避免責備用戶
- 提供清晰的下一步行動
- 保持遊戲的恐怖氛圍（適當時）
- 技術訊息用於日誌，用戶訊息用於顯示

## 未來擴展

可能的擴展方向：
- 錯誤統計和追蹤
- 錯誤本地化（多語言支援）
- 錯誤報告系統
- 自動重試策略
- 錯誤恢復建議
