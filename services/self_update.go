package services

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	updateOwner       = "monobilisim"
	updateRepo        = "vgw-manager"
	updateProjectName = "vgw-manager"
)

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// SelfUpdate downloads the latest release and replaces the current executable.
func SelfUpdate(currentVersion string) (string, bool, error) {
	if runtime.GOOS == "windows" {
		return "", false, errors.New("self-update is not supported on windows")
	}

	latestTag, assetURL, err := fetchLatestRelease(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", false, err
	}

	latestVersion := strings.TrimPrefix(latestTag, "v")
	if currentVersion != "" && currentVersion != "dev" && strings.TrimPrefix(currentVersion, "v") == latestVersion {
		return latestTag, false, nil
	}

	tempDir, err := os.MkdirTemp("", "vgw-manager-update-*")
	if err != nil {
		return "", false, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "release.tar.gz")
	if err := downloadFile(archivePath, assetURL); err != nil {
		return "", false, err
	}

	newBinary := filepath.Join(tempDir, updateProjectName)
	if err := extractBinary(archivePath, updateProjectName, newBinary); err != nil {
		return "", false, err
	}
	if err := os.Chmod(newBinary, 0o755); err != nil {
		return "", false, fmt.Errorf("failed to chmod new binary: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	backupPath := exePath + ".bak"
	_ = os.Remove(backupPath)
	if err := os.Rename(exePath, backupPath); err != nil {
		return "", false, fmt.Errorf("failed to replace current binary: %w", err)
	}

	if err := os.Rename(newBinary, exePath); err != nil {
		_ = os.Rename(backupPath, exePath)
		return "", false, fmt.Errorf("failed to install new binary: %w", err)
	}
	_ = os.Remove(backupPath)

	return latestTag, true, nil
}

func fetchLatestRelease(goos, goarch string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", updateOwner, updateRepo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", updateProjectName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", "", fmt.Errorf("failed to fetch latest release: %s", strings.TrimSpace(string(body)))
	}

	var info releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", "", fmt.Errorf("failed to decode release JSON: %w", err)
	}

	if info.TagName == "" {
		return "", "", errors.New("release tag not found")
	}

	assetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", updateProjectName, strings.TrimPrefix(info.TagName, "v"), goos, goarch)
	for _, asset := range info.Assets {
		if asset.Name == assetName {
			return info.TagName, asset.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("no release asset for %s/%s", goos, goarch)
}

func downloadFile(destination, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("failed to download asset: %s", strings.TrimSpace(string(body)))
	}

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create download file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write download file: %w", err)
	}

	return nil
}

func extractBinary(archivePath, binaryName, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if filepath.Base(header.Name) != binaryName {
			continue
		}

		out, err := os.Create(destination)
		if err != nil {
			return fmt.Errorf("failed to create binary file: %w", err)
		}
		if _, err := io.Copy(out, tarReader); err != nil {
			out.Close()
			return fmt.Errorf("failed to extract binary: %w", err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("failed to close binary file: %w", err)
		}
		return nil
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}
