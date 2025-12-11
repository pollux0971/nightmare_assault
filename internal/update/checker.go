package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	defaultTimeout   = 10 * time.Second
	maxRetries       = 3
)

// Checker 版本檢查器
type Checker struct {
	config     UpdateConfig
	httpClient *http.Client
	lastCheck  time.Time
	cachedInfo *ReleaseInfo
}

// NewChecker 創建新的版本檢查器
func NewChecker(config UpdateConfig) *Checker {
	return &Checker{
		config: config,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// CheckForUpdates 檢查是否有新版本可用
func (c *Checker) CheckForUpdates() (*UpdateResult, error) {
	// 如果有緩存且未過期，使用緩存
	if c.cachedInfo != nil && time.Since(c.lastCheck) < c.config.CheckInterval {
		return c.compareVersions(c.cachedInfo), nil
	}

	// 獲取最新release資訊
	releaseInfo, err := c.fetchLatestRelease()
	if err != nil {
		return &UpdateResult{
			Status:       UpdateStatusFailed,
			ErrorMessage: fmt.Sprintf("獲取版本資訊失敗: %v", err),
		}, err
	}

	// 更新緩存
	c.cachedInfo = releaseInfo
	c.lastCheck = time.Now()

	return c.compareVersions(releaseInfo), nil
}

// fetchLatestRelease 從GitHub API獲取最新release
func (c *Checker) fetchLatestRelease() (*ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest",
		githubAPIBaseURL, c.config.Owner, c.config.Repo)

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("創建請求失敗: %w", err)
		}

		// 設置User-Agent避免被rate limit
		req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", c.config.Repo, c.config.CurrentVersion))
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1)) // 指數退避
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			// Rate limited - 等待後重試
			lastErr = fmt.Errorf("API rate limit exceeded")
			time.Sleep(time.Minute)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API返回錯誤: %d - %s", resp.StatusCode, string(body))
		}

		var releaseInfo ReleaseInfo
		if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
			return nil, fmt.Errorf("解析release資訊失敗: %w", err)
		}

		return &releaseInfo, nil
	}

	return nil, fmt.Errorf("重試%d次後仍失敗: %w", maxRetries, lastErr)
}

// compareVersions 比較版本並返回結果
func (c *Checker) compareVersions(releaseInfo *ReleaseInfo) *UpdateResult {
	result := &UpdateResult{
		CurrentVersion: c.config.CurrentVersion,
		NewVersion:     releaseInfo.Version,
	}

	// 跳過draft和prerelease（除非用戶選擇）
	if releaseInfo.Draft {
		result.Status = UpdateStatusUpToDate
		result.ErrorMessage = "最新版本是draft，跳過"
		return result
	}

	if releaseInfo.Prerelease && !c.config.AllowPrerelease {
		result.Status = UpdateStatusUpToDate
		result.ErrorMessage = "最新版本是pre-release，跳過（可在設置中啟用）"
		return result
	}

	// 語義化版本比對
	currentVer, err := semver.NewVersion(strings.TrimPrefix(c.config.CurrentVersion, "v"))
	if err != nil {
		result.Status = UpdateStatusFailed
		result.ErrorMessage = fmt.Sprintf("解析當前版本失敗: %v", err)
		return result
	}

	newVer, err := semver.NewVersion(strings.TrimPrefix(releaseInfo.Version, "v"))
	if err != nil {
		result.Status = UpdateStatusFailed
		result.ErrorMessage = fmt.Sprintf("解析新版本失敗: %v", err)
		return result
	}

	// 比較版本
	if newVer.GreaterThan(currentVer) {
		result.Status = UpdateStatusAvailable

		// 找到當前平台的下載URL
		downloadURL, checksum := c.findPlatformAsset(releaseInfo.Assets)
		result.DownloadURL = downloadURL
		result.Checksum = checksum
	} else {
		result.Status = UpdateStatusUpToDate
	}

	return result
}

// findPlatformAsset 找到當前平台的執行檔asset
func (c *Checker) findPlatformAsset(assets []Asset) (downloadURL, checksum string) {
	// 構建當前平台的檔案名稱模式
	platformPattern := fmt.Sprintf("nightmare-%s-%s", runtime.GOOS, runtime.GOARCH)

	// Windows需要.exe後綴
	if runtime.GOOS == "windows" {
		platformPattern += ".exe"
	}

	var checksumAsset *Asset

	for _, asset := range assets {
		// 找到checksum文件
		if strings.Contains(asset.Name, "checksums") || strings.Contains(asset.Name, "SHA256") {
			checksumAsset = &asset
		}

		// 找到當前平台的執行檔
		if strings.Contains(asset.Name, platformPattern) {
			downloadURL = asset.DownloadURL
		}
	}

	// 如果找到checksum文件，獲取對應的checksum
	if checksumAsset != nil && downloadURL != "" {
		// 這裡簡化處理，實際應該下載並解析checksums文件
		// 暫時返回checksums文件URL，由下載器處理
		checksum = checksumAsset.DownloadURL
	}

	return downloadURL, checksum
}
