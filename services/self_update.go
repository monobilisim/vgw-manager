package services

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	updateOwner       = "monobilisim"
	updateRepo        = "vgw-manager"
	updateProjectName = "vgw-manager"
	checksumsAsset    = "checksums.txt"

	maxArchiveSize = 200 << 20
	maxBinarySize  = 200 << 20
)

var updateClient = &http.Client{Timeout: 5 * time.Minute}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseTarget struct {
	tag          string
	assetName    string
	assetURL     string
	checksumsURL string
}

// SelfUpdate downloads the latest release, verifies it against the release
// checksums.txt, and replaces the current executable.
func SelfUpdate(currentVersion string) (string, bool, error) {
	if runtime.GOOS == "windows" {
		return "", false, errors.New("self-update is not supported on windows")
	}

	target, err := fetchLatestRelease(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", false, err
	}

	latestVersion := strings.TrimPrefix(target.tag, "v")
	if currentVersion != "" && currentVersion != "dev" && strings.TrimPrefix(currentVersion, "v") == latestVersion {
		return target.tag, false, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Same directory as the executable so the final rename never crosses filesystems.
	tempDir, err := os.MkdirTemp(filepath.Dir(exePath), ".vgw-manager-update-*")
	if err != nil {
		return "", false, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	checksums, err := fetch(target.checksumsURL, 1<<20)
	if err != nil {
		return "", false, fmt.Errorf("failed to download %s: %w", checksumsAsset, err)
	}
	wantSum, err := checksumFor(checksums, target.assetName)
	if err != nil {
		return "", false, err
	}

	archivePath := filepath.Join(tempDir, "release.tar.gz")
	gotSum, err := downloadFile(archivePath, target.assetURL)
	if err != nil {
		return "", false, err
	}
	if gotSum != wantSum {
		return "", false, fmt.Errorf("checksum mismatch for %s: got %s, want %s", target.assetName, gotSum, wantSum)
	}

	newBinary := filepath.Join(tempDir, updateProjectName)
	if err := extractBinary(archivePath, updateProjectName, newBinary); err != nil {
		return "", false, err
	}
	if err := os.Chmod(newBinary, 0o755); err != nil {
		return "", false, fmt.Errorf("failed to chmod new binary: %w", err)
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

	return target.tag, true, nil
}

func fetchLatestRelease(goos, goarch string) (releaseTarget, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", updateOwner, updateRepo)
	body, err := fetch(url, 1<<20)
	if err != nil {
		return releaseTarget{}, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	var info releaseInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return releaseTarget{}, fmt.Errorf("failed to decode release JSON: %w", err)
	}

	if info.TagName == "" {
		return releaseTarget{}, errors.New("release tag not found")
	}

	target := releaseTarget{
		tag:       info.TagName,
		assetName: fmt.Sprintf("%s_%s_%s_%s.tar.gz", updateProjectName, strings.TrimPrefix(info.TagName, "v"), goos, goarch),
	}
	for _, asset := range info.Assets {
		switch asset.Name {
		case target.assetName:
			target.assetURL = asset.BrowserDownloadURL
		case checksumsAsset:
			target.checksumsURL = asset.BrowserDownloadURL
		}
	}

	if target.assetURL == "" {
		return releaseTarget{}, fmt.Errorf("no release asset for %s/%s", goos, goarch)
	}
	if target.checksumsURL == "" {
		return releaseTarget{}, fmt.Errorf("release %s has no %s; refusing unverified update", info.TagName, checksumsAsset)
	}
	return target, nil
}

func get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", updateProjectName)
	resp, err := updateClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func fetch(url string, limit int64) ([]byte, error) {
	resp, err := get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response larger than %d bytes", limit)
	}
	return body, nil
}

// checksumFor finds name in goreleaser's "<sha256>  <file>" checksums list.
func checksumFor(checksums []byte, name string) (string, error) {
	for line := range strings.Lines(string(checksums)) {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			sum := strings.ToLower(fields[0])
			if len(sum) != sha256.Size*2 {
				return "", fmt.Errorf("malformed checksum for %s", name)
			}
			return sum, nil
		}
	}
	return "", fmt.Errorf("%s not listed in %s", name, checksumsAsset)
}

// downloadFile saves url to destination and returns its hex SHA-256.
func downloadFile(destination, url string) (string, error) {
	resp, err := get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("failed to create download file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(resp.Body, maxArchiveSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to write download file: %w", err)
	}
	if n > maxArchiveSize {
		return "", fmt.Errorf("release archive larger than %d bytes", maxArchiveSize)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
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
		if header.Size > maxBinarySize {
			return fmt.Errorf("binary %s in archive is larger than %d bytes", binaryName, maxBinarySize)
		}

		out, err := os.Create(destination)
		if err != nil {
			return fmt.Errorf("failed to create binary file: %w", err)
		}
		if _, err := io.Copy(out, io.LimitReader(tarReader, maxBinarySize)); err != nil {
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
