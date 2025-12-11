package update

import "time"

// ReleaseInfo 代表GitHub release的資訊
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	Assets      []Asset   `json:"assets"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
}

// Asset 代表release中的附件檔案
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// UpdateStatus 更新狀態
type UpdateStatus int

const (
	UpdateStatusUnknown UpdateStatus = iota
	UpdateStatusChecking
	UpdateStatusAvailable
	UpdateStatusUpToDate
	UpdateStatusDownloading
	UpdateStatusVerifying
	UpdateStatusInstalling
	UpdateStatusCompleted
	UpdateStatusFailed
)

func (s UpdateStatus) String() string {
	switch s {
	case UpdateStatusChecking:
		return "檢查中"
	case UpdateStatusAvailable:
		return "有新版本"
	case UpdateStatusUpToDate:
		return "已是最新版本"
	case UpdateStatusDownloading:
		return "下載中"
	case UpdateStatusVerifying:
		return "驗證中"
	case UpdateStatusInstalling:
		return "安裝中"
	case UpdateStatusCompleted:
		return "完成"
	case UpdateStatusFailed:
		return "失敗"
	default:
		return "未知"
	}
}

// UpdateConfig 更新配置
type UpdateConfig struct {
	Owner          string        // GitHub repo owner
	Repo           string        // GitHub repo name
	CurrentVersion string        // 當前版本
	CheckInterval  time.Duration // 檢查間隔
	AllowPrerelease bool         // 是否接受pre-release版本
	CacheDir       string        // 緩存目錄
}

// UpdateResult 更新結果
type UpdateResult struct {
	Status       UpdateStatus
	NewVersion   string
	CurrentVersion string
	ErrorMessage string
	DownloadURL  string
	Checksum     string
}
