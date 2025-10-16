package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jarv/newsgoat/internal/logging"
	"github.com/jarv/newsgoat/internal/version"
)

const (
	githubAPIURL = "https://api.github.com/repos/jarv/newsgoat/releases/latest"
	downloadURL  = "https://github.com/jarv/newsgoat/releases/download/%s/%s"
	timeout      = 10 * time.Second
)

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	DownloadURL    string
	ReleaseURL     string
}

// CheckForUpdate queries GitHub API for the latest release
func CheckForUpdate() (*UpdateInfo, error) {
	currentVersion := version.GetVersion()

	// Skip update check for dev builds
	if currentVersion == "dev" {
		logging.Debug("Skipping update check for dev build")
		return nil, nil
	}

	logging.Debug("Checking for updates", "api_url", githubAPIURL)

	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", version.GetUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logging.Debug("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	latestVersion := release.TagName

	logging.Debug("Version comparison", "current", currentVersion, "latest", latestVersion)

	// If versions are the same, no update needed
	if currentVersion == latestVersion {
		logging.Debug("Already on latest version")
		return nil, nil
	}

	// Find the appropriate binary for this platform
	binaryName := getBinaryName()
	logging.Debug("Looking for binary", "name", binaryName, "platform", runtime.GOOS, "arch", runtime.GOARCH)

	var downloadURL string

	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no binary found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	logging.Info("Update available", "version", latestVersion, "download_url", downloadURL)

	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		DownloadURL:    downloadURL,
		ReleaseURL:     release.HTMLURL,
	}, nil
}

// DownloadAndInstall downloads the latest version and replaces the current binary
func DownloadAndInstall(downloadURL string) error {
	logging.Info("Starting update installation", "download_url", downloadURL)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	logging.Debug("Current executable path", "path", execPath)

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	logging.Debug("Resolved executable path", "path", execPath)

	// Download the new binary
	logging.Info("Downloading update", "url", downloadURL)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logging.Debug("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	logging.Debug("Download response received", "status", resp.StatusCode, "content_length", resp.ContentLength)

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "newsgoat-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if removeErr := os.Remove(tmpPath); removeErr != nil {
			logging.Debug("Failed to remove temp file", "path", tmpPath, "error", removeErr)
		}
	}()

	logging.Debug("Created temporary file", "path", tmpPath)

	// Write downloaded content to temp file
	bytesWritten, err := io.Copy(tmpFile, resp.Body)
	if closeErr := tmpFile.Close(); closeErr != nil {
		return fmt.Errorf("failed to close temp file: %w", closeErr)
	}
	if err != nil {
		return fmt.Errorf("failed to write update: %w", err)
	}

	logging.Debug("Downloaded binary written to temp file", "bytes", bytesWritten, "path", tmpPath)

	// Make it executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	logging.Debug("Set executable permissions on temp file")

	// Backup current binary
	backupPath := execPath + ".backup"
	logging.Debug("Backing up current binary", "from", execPath, "to", backupPath)

	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to current location
	logging.Info("Installing new binary", "from", tmpPath, "to", execPath)

	if err := os.Rename(tmpPath, execPath); err != nil {
		// Restore backup if move fails
		logging.Error("Failed to install update, restoring backup", "error", err)
		if restoreErr := os.Rename(backupPath, execPath); restoreErr != nil {
			logging.Error("Failed to restore backup", "error", restoreErr)
		}
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Remove backup on success
	logging.Debug("Removing backup file", "path", backupPath)
	if removeErr := os.Remove(backupPath); removeErr != nil {
		logging.Debug("Failed to remove backup file", "path", backupPath, "error", removeErr)
	}

	logging.Info("Update installed successfully", "path", execPath)

	return nil
}

// getBinaryName returns the expected binary name for the current platform
func getBinaryName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Normalize architecture names to match release naming
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
	}

	arch, ok := archMap[goarch]
	if !ok {
		arch = goarch
	}

	return fmt.Sprintf("newsgoat-%s-%s", goos, arch)
}

// IsNewerVersion compares two version strings (format: v1.2.3)
func IsNewerVersion(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Simple string comparison (works for semantic versions)
	return latest > current
}
