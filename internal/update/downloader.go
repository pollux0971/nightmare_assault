package update

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Downloader 檔案下載器
type Downloader struct {
	httpClient *http.Client
	cacheDir   string
}

// NewDownloader 創建新的下載器
func NewDownloader(cacheDir string) *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 0, // 下載大文件時不設置timeout
		},
		cacheDir: cacheDir,
	}
}

// DownloadProgress 下載進度回調
type DownloadProgress func(downloaded, total int64)

// Download 下載文件
func (d *Downloader) Download(url string, progress DownloadProgress) (string, error) {
	// 創建緩存目錄
	if err := os.MkdirAll(d.cacheDir, 0755); err != nil {
		return "", fmt.Errorf("創建緩存目錄失敗: %w", err)
	}

	// 提取文件名
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	destPath := filepath.Join(d.cacheDir, filename)

	// 檢查文件是否已存在
	if _, err := os.Stat(destPath); err == nil {
		// 文件已存在，檢查是否需要重新下載
		return destPath, nil
	}

	// 創建請求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("創建下載請求失敗: %w", err)
	}

	// 執行請求
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下載失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下載失敗: HTTP %d", resp.StatusCode)
	}

	// 創建臨時文件
	tmpFile, err := os.CreateTemp(d.cacheDir, "download-*.tmp")
	if err != nil {
		return "", fmt.Errorf("創建臨時文件失敗: %w", err)
	}
	defer tmpFile.Close()

	// 下載並寫入
	totalSize := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				os.Remove(tmpFile.Name())
				return "", fmt.Errorf("寫入文件失敗: %w", writeErr)
			}

			downloaded += int64(n)

			// 報告進度
			if progress != nil {
				progress(downloaded, totalSize)
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("讀取數據失敗: %w", err)
		}
	}

	// 關閉臨時文件
	tmpFile.Close()

	// 重命名為最終文件
	if err := os.Rename(tmpFile.Name(), destPath); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("移動文件失敗: %w", err)
	}

	return destPath, nil
}

// VerifyChecksum 驗證文件checksum
func (d *Downloader) VerifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打開文件失敗: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("計算checksum失敗: %w", err)
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))

	// 比對checksum（不區分大小寫）
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum不符: 預期=%s, 實際=%s", expectedChecksum, actualChecksum)
	}

	return nil
}

// DownloadWithRetry 帶重試的下載
func (d *Downloader) DownloadWithRetry(url string, maxRetries int, progress DownloadProgress) (string, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		filePath, err := d.Download(url, progress)
		if err == nil {
			return filePath, nil
		}

		lastErr = err

		// 如果是最後一次重試，不再等待
		if i < maxRetries-1 {
			// 清理可能的部分下載文件
			if filePath != "" {
				os.Remove(filePath)
			}
		}
	}

	return "", fmt.Errorf("重試%d次後仍失敗: %w", maxRetries, lastErr)
}

// FetchChecksums 下載並解析checksums文件
func (d *Downloader) FetchChecksums(checksumURL string) (map[string]string, error) {
	// 下載checksums文件
	filePath, err := d.Download(checksumURL, nil)
	if err != nil {
		return nil, fmt.Errorf("下載checksums文件失敗: %w", err)
	}

	// 讀取文件內容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("讀取checksums文件失敗: %w", err)
	}

	// 解析checksums (格式: <checksum>  <filename>)
	checksums := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[1]
			checksums[filename] = checksum
		}
	}

	return checksums, nil
}
