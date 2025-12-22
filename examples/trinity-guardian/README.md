# Trinity & Guardian 代碼範例

本目錄包含 Trinity & Guardian 系統的可運行代碼範例。

## 範例列表

### 1. basic-usage.go - 基礎用法

展示最基本的 Trinity 使用方式。

**功能**:
- 創建 Trinity Router
- 使用不同 Agent（JudgeAgent、NarrationAgent、DreamAgent）
- 自動 Tier 路由
- 基本錯誤處理
- 獲取 Metrics 摘要

**運行方式**:
```bash
export ANTHROPIC_API_KEY="your-api-key"
cd examples/trinity-guardian
go run basic-usage.go
```

**預期輸出**:
```
=== Trinity Basic Usage Example ===

Step 1: Creating Trinity Router configuration...
Step 2: Creating Trinity Router...
✓ Trinity Router created successfully

Example 1: Using JudgeAgent (Thinking Tier)
-------------------------------------------
Agent: JudgeAgent
Tier: Thinking (Opus 4.5)
Response: [Response content...]

...
```

---

### 2. full-integration.go - 完整集成

展示 Trinity + Guardian 的完整集成，包括玩家保護、性能監控等。

**功能**:
- Trinity LLM Client 創建（高階 API）
- Guardian 玩家保護集成
- 遊戲狀態模擬
- 動態 Tier 升級（當 Guardian 觸發保護時）
- 完整的 Metrics 收集和報告
- Fallback 機制測試

**運行方式**:
```bash
export ANTHROPIC_API_KEY="your-api-key"
go run full-integration.go
```

**預期輸出**:
```
=== Trinity + Guardian Full Integration Example ===

Step 1: Initializing Trinity Router...
Step 2: Creating Trinity LLM Client...
✓ Trinity LLM Client created

Step 3: Initializing Experience Guardian...
✓ Experience Guardian initialized

Step 4: Creating mock game state...
✓ Game state created (HP: 15, SAN: 25)

Step 5: Running game loop simulation...
=======================================

--- Turn 1 ---
⚠️  Guardian Protection Activated: low HP/SAN streak: 1 turns
✓ Upgraded NarrationAgent to Thinking tier

...
```

---

### 3. custom-config.go - 自訂配置

展示高級配置和動態調整功能。

**功能**:
- 多 Provider 配置（Anthropic + OpenAI）
- 自訂 Agent Tier 覆寫
- 自訂 Retry 策略
- 動態 Tier 調整（基於性能分析）
- 請求級別的 Tier 覆寫
- 性能分析和優化建議

**運行方式**:
```bash
export ANTHROPIC_API_KEY="your-anthropic-key"
export OPENAI_API_KEY="your-openai-key"  # 可選
go run custom-config.go
```

**預期輸出**:
```
=== Trinity Custom Configuration Example ===

Step 1: Creating custom multi-provider configuration...
✓ Custom configuration created
  Thinking: claude-opus-4-20250514 (anthropic)
  Reactive: claude-3-5-sonnet-20241022 (anthropic)
  Rapid: gpt-4o-mini (openai)

Step 2: Creating Trinity LLM Client...
✓ Trinity LLM Client created

Step 3: Custom Agent Tier Overrides
------------------------------------
NarrationAgent:
  Default Tier: Reactive
  Custom Tier: Thinking
  Reason: Improve story quality

...
```

---

## 依賴項

所有範例需要以下依賴：

```go
github.com/nightmare-assault/nightmare-assault/internal/trinity
github.com/nightmare-assault/nightmare-assault/internal/guardian
github.com/nightmare-assault/nightmare-assault/internal/api/client
github.com/nightmare-assault/nightmare-assault/internal/engine
```

## 環境變數

### 必需

- `ANTHROPIC_API_KEY` - Anthropic API 金鑰

### 可選

- `OPENAI_API_KEY` - OpenAI API 金鑰（僅 custom-config.go 使用）
- `OPENROUTER_API_KEY` - OpenRouter API 金鑰（可用於多 Provider 配置）

## 設置步驟

1. **安裝依賴**:
   ```bash
   cd /path/to/nightmare-assault
   go mod download
   ```

2. **設置 API Key**:
   ```bash
   export ANTHROPIC_API_KEY="sk-ant-api03-..."
   ```

3. **運行範例**:
   ```bash
   cd examples/trinity-guardian
   go run basic-usage.go
   ```

## 範例對比

| 特性 | basic-usage.go | full-integration.go | custom-config.go |
|------|----------------|---------------------|------------------|
| Trinity Router | ✅ | ✅ | ✅ |
| Trinity LLM Client | ❌ | ✅ | ✅ |
| Guardian 集成 | ❌ | ✅ | ❌ |
| 多 Provider | ❌ | ❌ | ✅ |
| 自訂 Tier 覆寫 | ❌ | ❌ | ✅ |
| Metrics 收集 | 基礎 | 完整 | 完整 |
| 動態調整 | ❌ | ✅ | ✅ |
| 複雜度 | 簡單 | 中等 | 高 |

## 故障排查

### 錯誤: "ANTHROPIC_API_KEY environment variable not set"

**解決方案**:
```bash
export ANTHROPIC_API_KEY="your-api-key"
```

### 錯誤: "failed to create Thinking provider"

**原因**: API Key 無效或網路問題

**解決方案**:
1. 檢查 API Key 是否正確
2. 檢查網路連接
3. 檢查 Provider 服務狀態

### 錯誤: "context deadline exceeded"

**原因**: 請求超時

**解決方案**:
在 `TrinityClientConfig` 中增加超時時間：
```go
clientConfig := trinity.TrinityClientConfig{
    DefaultTimeout: 120 * time.Second, // 增加到 120 秒
}
```

## 學習路徑

1. **新手**: 從 `basic-usage.go` 開始
2. **進階**: 學習 `full-integration.go` 的 Guardian 集成
3. **專家**: 探索 `custom-config.go` 的高級配置

## 擴展範例

您可以基於這些範例創建自己的用法：

### 範例：添加自訂 Agent

```go
// 在 basic-usage.go 基礎上添加
customMessages := []client.Message{
    {Role: "user", Content: "Your custom prompt"},
}

resp, err := router.Route(ctx, "CustomAgent", customMessages)
// CustomAgent 會自動使用 Reactive tier（默認）
```

### 範例：使用 Thinking Tag 提取

```go
// 在 full-integration.go 基礎上添加
resp, err := llmClient.SendMessage(ctx, "JudgeAgent", messages)

if llmClient.HasThinking(resp) {
    thinkingChain, cleanContent := llmClient.ExtractThinking(resp)
    fmt.Println("思考過程:", thinkingChain)
    fmt.Println("最終答案:", cleanContent)
}
```

## 性能基準

在 Apple M1 Mac 上的典型性能（使用 Anthropic）：

| Tier | 平均延遲 | 成本/請求 |
|------|----------|-----------|
| Thinking (Opus) | 8-12 秒 | $0.02-0.05 |
| Reactive (Sonnet) | 3-5 秒 | $0.002-0.005 |
| Rapid (Haiku) | 1-2 秒 | $0.0003-0.001 |

## 常見問題

### Q: 可以使用免費的 API 嗎？

A: Anthropic、OpenAI、OpenRouter 都沒有免費方案，但提供試用額度。建議從 Anthropic 開始，獲得 $5 免費額度。

### Q: 範例會產生多少成本？

A: 每個範例約 $0.10-0.50，取決於請求次數和使用的模型。

### Q: 如何限制成本？

A: 使用 `custom-config.go` 中的低成本配置，或減少請求次數。

## 參考資料

- [Trinity 用戶指南](../../docs/trinity-guardian-guide.md)
- [Trinity API 參考](../../docs/trinity-api-reference.md)
- [Trinity 遷移指南](../../docs/trinity-migration-guide.md)
- [配置範例](../../docs/trinity-guardian-config-examples/)

## 貢獻

歡迎提交新的範例！請確保：
- ✅ 代碼可編譯運行
- ✅ 包含清晰的註釋
- ✅ 更新本 README
- ✅ 添加必要的錯誤處理

---

**最後更新**: 2025-01-15
**維護者**: Nightmare Assault 開發團隊
