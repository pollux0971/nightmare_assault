# API 配置設定指南

本專案支援兩種 API 配置方式：**環境變數**（推薦）和**配置文件**。

## 🔐 方法一：使用環境變數（推薦）

環境變數是最安全的方式，不會將 API key 存入文件。

### 步驟：

1. **複製範例文件**
   ```bash
   cp .env.example .env
   ```

2. **編輯 .env 文件**
   ```bash
   nano .env  # 或使用你喜歡的編輯器
   ```

3. **填入你的 API 配置**
   ```bash
   # 選擇 Provider（openrouter、anthropic、openai、google、cohere）
   NIGHTMARE_API_PROVIDER=openrouter

   # 選擇模型
   NIGHTMARE_API_MODEL=anthropic/claude-3.5-sonnet

   # 設定最大 token 數
   NIGHTMARE_API_MAX_TOKENS=100000

   # 填入對應的 API Key
   OPENROUTER_API_KEY=sk-or-v1-你的KEY
   ```

4. **載入環境變數並執行**
   ```bash
   # Linux/macOS
   export $(cat .env | xargs)
   ./nightmare-assault

   # 或直接在執行時載入
   source .env && ./nightmare-assault
   ```

## 📄 方法二：使用配置文件

配置文件會存在你的家目錄，可以透過遊戲介面設定。

### 自動配置：
遊戲首次啟動時會引導你設定 API。

### 手動配置：

1. **複製範例文件到家目錄**
   ```bash
   mkdir -p ~/.nightmare
   cp config.example.json ~/.nightmare/config.json
   ```

2. **編輯配置文件**
   ```bash
   nano ~/.nightmare/config.json
   ```

3. **填入你的 API 配置**
   ```json
   {
     "api": {
       "provider": {
         "provider_id": "openrouter",
         "model": "anthropic/claude-3.5-sonnet",
         "max_tokens": 100000
       },
       "api_keys": {
         "openrouter": "sk-or-v1-你的KEY",
         "anthropic": "sk-ant-你的KEY",
         "openai": "sk-你的KEY"
       }
     }
   }
   ```

## 🔑 支援的 Provider

| Provider | 環境變數 | Provider ID | 說明 |
|----------|---------|------------|------|
| **OpenRouter** | `OPENROUTER_API_KEY` | `openrouter` | 推薦！支援多種模型 |
| **Anthropic** | `ANTHROPIC_API_KEY` | `anthropic` | Claude 官方 API |
| **OpenAI** | `OPENAI_API_KEY` | `openai` | GPT 系列 |
| **Google** | `GOOGLE_API_KEY` | `google` | Gemini 系列 |
| **Cohere** | `COHERE_API_KEY` | `cohere` | Cohere 模型 |

## 🎯 推薦設定

### OpenRouter（最推薦）
```bash
NIGHTMARE_API_PROVIDER=openrouter
NIGHTMARE_API_MODEL=anthropic/claude-3.5-sonnet
NIGHTMARE_API_MAX_TOKENS=100000
OPENROUTER_API_KEY=sk-or-v1-你的KEY
```

**優點：**
- 單一 API key 可使用多種模型
- 價格透明，按使用付費
- 支援 Claude、GPT、Gemini 等多種模型

### Anthropic（直接使用 Claude）
```bash
NIGHTMARE_API_PROVIDER=anthropic
NIGHTMARE_API_MODEL=claude-3-5-sonnet-20241022
NIGHTMARE_API_MAX_TOKENS=100000
ANTHROPIC_API_KEY=sk-ant-你的KEY
```

## 🔒 安全注意事項

1. **絕對不要提交 API key 到 Git**
   - `.env` 和 `~/.nightmare/config.json` 已在 `.gitignore` 中
   - 只提交 `.env.example` 和 `config.example.json`

2. **優先順序**
   - 環境變數 > 配置文件
   - 如果兩者都設定，環境變數會覆蓋配置文件

3. **撤銷洩露的 Key**
   - 如果不小心洩露了 API key，立即到對應平台撤銷

## 🧪 測試配置

啟動遊戲後，在 API 設定畫面可以測試連線：

```bash
./nightmare-assault
# 進入 Settings > API Setup > Test Connection
```

## 💻 程式碼中使用

```go
import "github.com/nightmare-assault/nightmare-assault/internal/config"

// 載入配置（自動從環境變數和配置文件載入）
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// 獲取 API Key（優先使用環境變數）
apiKey := cfg.GetAPIKey("openrouter")

// 設定 API Key 並存檔
cfg.SetAPIKey("openrouter", "sk-or-v1-...")
```

## ❓ 常見問題

### Q: 環境變數沒有生效？
A: 確保在執行程式前已載入環境變數：
```bash
source .env
./nightmare-assault
```

### Q: 如何切換不同的 Provider？
A: 修改環境變數或配置文件中的 `provider_id` 和對應的 API key。

### Q: 配置文件在哪裡？
A:
- Linux/macOS: `~/.nightmare/config.json`
- Windows: `%USERPROFILE%\.nightmare\config.json`

### Q: 可以同時使用多個 Provider 嗎？
A: 可以！填入多個 API key，然後在遊戲中切換 provider。

## 📚 更多資訊

- [OpenRouter 註冊](https://openrouter.ai/)
- [Anthropic API 文檔](https://docs.anthropic.com/)
- [專案 README](README.md)
