package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/Easy8Com/easy8-cli/releases/latest"
	maxUpdateDownloadSize  = 200 << 20
	autoUpdateInterval     = 24 * time.Hour
	autoUpdateTimeout      = 10 * time.Second
)

var (
	updateReleaseURL        = githubLatestReleaseURL
	updateHTTPClient        = &http.Client{Timeout: 60 * time.Second}
	updateExecutable        = os.Executable
	updateRuntimeGOOS       = func() string { return runtime.GOOS }
	updateRuntimeGOARCH     = func() string { return runtime.GOARCH }
	autoUpdateRunner        = updateFromGitHub
	autoUpdateNow           = time.Now
	autoUpdateStateFilePath = defaultAutoUpdateStateFilePath
)

type updateResult struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Updated        bool   `json:"updated"`
	Asset          string `json:"asset,omitempty"`
	Path           string `json:"path,omitempty"`
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type autoUpdateState struct {
	LastCheck string `yaml:"last_check"`
}

func runUpdate(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	jsonOut := fs.Bool("json", false, "JSON envelope output")
	quietOut := fs.Bool("quiet", false, "Raw JSON data output")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := validateMachineFlags(*jsonOut, *quietOut); err != nil {
		return usageError(err)
	}
	if fs.NArg() > 0 {
		return usageError(fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " ")))
	}

	result, err := updateFromGitHub(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "update error:", err)
		return 1
	}

	if *quietOut {
		return outputJSON(result)
	}
	if *jsonOut {
		summary := "Already up to date"
		if result.Updated {
			summary = fmt.Sprintf("Updated to %s", result.LatestVersion)
		}
		return outputJSONEnvelope(result, summary, nil, nil)
	}

	if result.Updated {
		fmt.Fprintf(os.Stdout, "easy8 updated %s -> %s at %s\n", result.CurrentVersion, result.LatestVersion, result.Path)
		return 0
	}

	fmt.Fprintf(os.Stdout, "easy8 %s is already up to date.\n", result.CurrentVersion)
	return 0
}

func updateFromGitHub(ctx context.Context) (updateResult, error) {
	currentVersion := normalizeReleaseVersion(Version)
	result := updateResult{CurrentVersion: currentVersion}

	release, err := fetchLatestRelease(ctx)
	if err != nil {
		return result, err
	}

	latestVersion := normalizeReleaseVersion(release.TagName)
	if latestVersion == "" {
		return result, fmt.Errorf("latest GitHub release is missing tag_name")
	}
	result.LatestVersion = latestVersion

	if currentVersion == latestVersion {
		return result, nil
	}

	assetName, err := updateAssetName(updateRuntimeGOOS(), updateRuntimeGOARCH())
	if err != nil {
		return result, err
	}
	result.Asset = assetName

	assetURL, err := findReleaseAsset(release, assetName)
	if err != nil {
		return result, err
	}
	checksumsURL, err := findReleaseAsset(release, "checksums.txt")
	if err != nil {
		return result, err
	}

	binary, err := downloadUpdateFile(ctx, assetURL)
	if err != nil {
		return result, err
	}
	checksums, err := downloadUpdateFile(ctx, checksumsURL)
	if err != nil {
		return result, err
	}
	if err := verifyUpdateChecksum(assetName, binary, string(checksums)); err != nil {
		return result, err
	}

	executablePath, err := updateExecutable()
	if err != nil {
		return result, err
	}
	if err := replaceExecutable(executablePath, binary); err != nil {
		return result, err
	}

	result.Updated = true
	result.Path = executablePath
	return result, nil
}

func shouldRunStartupAutoUpdate(command string) bool {
	switch command {
	case "issue", "pbi", "auth", "skill":
		return true
	default:
		return false
	}
}

func maybeRunAutoUpdate(enabled bool) {
	if !enabled {
		return
	}
	now := autoUpdateNow()
	due, err := isAutoUpdateDue(now)
	if err != nil || !due {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), autoUpdateTimeout)
	defer cancel()
	_, _ = autoUpdateRunner(ctx)
	_ = writeAutoUpdateState(now)
}

func isAutoUpdateDue(now time.Time) (bool, error) {
	path, err := autoUpdateStateFilePath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	var state autoUpdateState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return true, nil
	}
	lastCheck, err := time.Parse(time.RFC3339, strings.TrimSpace(state.LastCheck))
	if err != nil {
		return true, nil
	}
	return !now.Before(lastCheck.Add(autoUpdateInterval)), nil
}

func writeAutoUpdateState(now time.Time) error {
	path, err := autoUpdateStateFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(autoUpdateState{LastCheck: now.UTC().Format(time.RFC3339)})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func defaultAutoUpdateStateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "easy8", "update-state.yaml"), nil
}

func fetchLatestRelease(ctx context.Context) (githubRelease, error) {
	var release githubRelease
	body, err := downloadUpdateFile(ctx, updateReleaseURL)
	if err != nil {
		return release, err
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return release, err
	}
	return release, nil
}

func downloadUpdateFile(ctx context.Context, urlValue string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlValue, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json, application/octet-stream")
	req.Header.Set("User-Agent", "easy8-cli/"+Version)

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxUpdateDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxUpdateDownloadSize {
		return nil, fmt.Errorf("download exceeds %d bytes: %s", maxUpdateDownloadSize, urlValue)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("GET %s failed with HTTP %d: %s", urlValue, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func updateAssetName(goos, goarch string) (string, error) {
	switch goos {
	case "linux", "darwin", "windows":
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	switch goarch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	name := fmt.Sprintf("easy8-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name, nil
}

func findReleaseAsset(release githubRelease, name string) (string, error) {
	for _, asset := range release.Assets {
		if asset.Name == name && strings.TrimSpace(asset.BrowserDownloadURL) != "" {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("GitHub release %s does not include asset %s", release.TagName, name)
}

func verifyUpdateChecksum(assetName string, binary []byte, checksums string) error {
	expected := ""
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == assetName {
			expected = strings.ToLower(fields[0])
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum entry for %s not found", assetName)
	}

	actualBytes := sha256.Sum256(binary)
	actual := fmt.Sprintf("%x", actualBytes)
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", assetName, expected, actual)
	}
	return nil
}

func replaceExecutable(path string, binary []byte) error {
	path = filepath.Clean(path)
	mode := os.FileMode(0o755)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(binary); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return replaceExecutableWindows(path, tmpPath, &removeTemp)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	removeTemp = false
	return nil
}

func replaceExecutableWindows(path, tmpPath string, removeTemp *bool) error {
	backupPath := path + ".old"
	_ = os.Remove(backupPath)
	if err := os.Rename(path, backupPath); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Rename(backupPath, path)
		return fmt.Errorf("replace %s: %w", path, err)
	}
	*removeTemp = false
	_ = os.Remove(backupPath)
	return nil
}

func normalizeReleaseVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
