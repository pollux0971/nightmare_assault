package update

import (
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	config := UpdateConfig{
		Owner:          "test-owner",
		Repo:           "test-repo",
		CurrentVersion: "v1.0.0",
		CheckInterval:  24 * time.Hour,
	}

	checker := NewChecker(config)

	if checker == nil {
		t.Fatal("Expected checker to be created")
	}

	if checker.config.Owner != "test-owner" {
		t.Errorf("Expected owner to be 'test-owner', got '%s'", checker.config.Owner)
	}
}

func TestUpdateStatusString(t *testing.T) {
	tests := []struct {
		status   UpdateStatus
		expected string
	}{
		{UpdateStatusChecking, "檢查中"},
		{UpdateStatusAvailable, "有新版本"},
		{UpdateStatusUpToDate, "已是最新版本"},
		{UpdateStatusDownloading, "下載中"},
		{UpdateStatusFailed, "失敗"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("UpdateStatus.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	checker := NewChecker(UpdateConfig{
		CurrentVersion: "v1.0.0",
	})

	tests := []struct {
		name           string
		releaseVersion string
		prerelease     bool
		draft          bool
		expectedStatus UpdateStatus
	}{
		{
			name:           "newer version available",
			releaseVersion: "v1.1.0",
			expectedStatus: UpdateStatusAvailable,
		},
		{
			name:           "same version",
			releaseVersion: "v1.0.0",
			expectedStatus: UpdateStatusUpToDate,
		},
		{
			name:           "older version",
			releaseVersion: "v0.9.0",
			expectedStatus: UpdateStatusUpToDate,
		},
		{
			name:           "draft version skipped",
			releaseVersion: "v1.2.0",
			draft:          true,
			expectedStatus: UpdateStatusUpToDate,
		},
		{
			name:           "prerelease skipped",
			releaseVersion: "v1.2.0-beta.1",
			prerelease:     true,
			expectedStatus: UpdateStatusUpToDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseInfo := &ReleaseInfo{
				Version:    tt.releaseVersion,
				Prerelease: tt.prerelease,
				Draft:      tt.draft,
			}

			result := checker.compareVersions(releaseInfo)

			if result.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, result.Status)
			}
		})
	}
}

func TestFindPlatformAsset(t *testing.T) {
	checker := NewChecker(UpdateConfig{})

	assets := []Asset{
		{Name: "nightmare-linux-amd64", DownloadURL: "http://example.com/linux"},
		{Name: "nightmare-darwin-amd64", DownloadURL: "http://example.com/darwin"},
		{Name: "nightmare-windows-amd64.exe", DownloadURL: "http://example.com/windows"},
		{Name: "checksums.txt", DownloadURL: "http://example.com/checksums"},
	}

	downloadURL, checksumURL := checker.findPlatformAsset(assets)

	if downloadURL == "" {
		t.Error("Expected to find platform asset")
	}

	if checksumURL == "" {
		t.Error("Expected to find checksum file")
	}
}
