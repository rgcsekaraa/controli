package controli

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
)

const DefaultGitHubRepo = "rgcsekaraa/controli"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func UpdateFromLatestRelease(repo string, output io.Writer) error {
	if strings.TrimSpace(repo) == "" {
		repo = DefaultGitHubRepo
	}
	release, err := latestRelease(repo)
	if err != nil {
		return err
	}
	assetName, err := releaseAssetName()
	if err != nil {
		return err
	}
	downloadURL := ""
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("release %s does not contain asset %s", release.TagName, assetName)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return err
	}
	data, err := downloadFile(downloadURL)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		target := strings.TrimSuffix(executable, ".exe") + ".new.exe"
		if err := os.WriteFile(target, data, 0o755); err != nil {
			return err
		}
		fmt.Fprintf(output, "downloaded %s to %s\n", release.TagName, target)
		fmt.Fprintln(output, "replace the existing controli.exe with this file after closing running Controli windows")
		return nil
	}
	temp := executable + ".tmp"
	if err := os.WriteFile(temp, data, 0o755); err != nil {
		return err
	}
	if err := os.Rename(temp, executable); err != nil {
		_ = os.Remove(temp)
		return err
	}
	fmt.Fprintf(output, "updated controli to %s\n", release.TagName)
	return nil
}

func latestRelease(repo string) (githubRelease, error) {
	var release githubRelease
	client := http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return release, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "controli-updater")
	resp, err := client.Do(req)
	if err != nil {
		return release, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return release, fmt.Errorf("GitHub returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return release, err
	}
	if release.TagName == "" {
		return release, fmt.Errorf("latest release response did not include a tag")
	}
	return release, nil
}

func releaseAssetName() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
			return "", fmt.Errorf("unsupported macOS architecture: %s", runtime.GOARCH)
		}
		return "controli-darwin-" + runtime.GOARCH, nil
	case "linux":
		switch runtime.GOARCH {
		case "386", "amd64", "arm64", "ppc64le", "riscv64", "s390x":
			return "controli-linux-" + runtime.GOARCH, nil
		case "arm":
			return "controli-linux-armv7", nil
		default:
			return "", fmt.Errorf("unsupported Linux architecture: %s", runtime.GOARCH)
		}
	case "windows":
		switch runtime.GOARCH {
		case "386", "amd64", "arm", "arm64":
			return "controli-windows-" + runtime.GOARCH + ".exe", nil
		default:
			return "", fmt.Errorf("unsupported Windows architecture: %s", runtime.GOARCH)
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func downloadFile(url string) ([]byte, error) {
	client := http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "controli-updater")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("download returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}
