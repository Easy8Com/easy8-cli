package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestUpdateCommandAlreadyCurrent(t *testing.T) {
	setTestHome(t)
	resetUpdateTestDeps(t)
	Version = "0.1.2"

	server := newUpdateTestServer(t, "v0.1.2", "unused", nil, "")
	defer server.Close()
	updateReleaseURL = server.URL + "/latest"

	stdout, stderr, code := captureRun(t, []string{"update", "--quiet"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}

	var result updateResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if result.Updated {
		t.Fatalf("expected no update: %+v", result)
	}
	if result.CurrentVersion != "0.1.2" || result.LatestVersion != "0.1.2" {
		t.Fatalf("unexpected versions: %+v", result)
	}
}

func TestUpdateCommandDownloadsAndReplacesExecutable(t *testing.T) {
	setTestHome(t)
	resetUpdateTestDeps(t)
	Version = "0.1.2"

	assetName, err := updateAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	binary := []byte("updated binary")
	server := newUpdateTestServer(t, "v0.1.3", assetName, binary, "")
	defer server.Close()
	updateReleaseURL = server.URL + "/latest"

	executablePath := filepath.Join(t.TempDir(), "easy8")
	if runtime.GOOS == "windows" {
		executablePath += ".exe"
	}
	if err := os.WriteFile(executablePath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}
	updateExecutable = func() (string, error) { return executablePath, nil }

	stdout, stderr, code := captureRun(t, []string{"update", "--quiet"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	updatedBinary, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read executable: %v", err)
	}
	if string(updatedBinary) != string(binary) {
		t.Fatalf("executable was not replaced: %q", string(updatedBinary))
	}

	var result updateResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if !result.Updated {
		t.Fatalf("expected update: %+v", result)
	}
	if result.Asset != assetName || result.Path != executablePath {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestUpdateCommandChecksumMismatch(t *testing.T) {
	setTestHome(t)
	resetUpdateTestDeps(t)
	Version = "0.1.2"

	assetName, err := updateAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported test platform: %v", err)
	}
	server := newUpdateTestServer(t, "v0.1.3", assetName, []byte("updated binary"), strings.Repeat("0", 64))
	defer server.Close()
	updateReleaseURL = server.URL + "/latest"

	executablePath := filepath.Join(t.TempDir(), "easy8")
	if runtime.GOOS == "windows" {
		executablePath += ".exe"
	}
	if err := os.WriteFile(executablePath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}
	updateExecutable = func() (string, error) { return executablePath, nil }

	_, stderr, code := captureRun(t, []string{"update"})
	if code != 1 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got stderr=%s", stderr)
	}
	currentBinary, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read executable: %v", err)
	}
	if string(currentBinary) != "old binary" {
		t.Fatalf("executable changed despite checksum failure: %q", string(currentBinary))
	}
}

func TestUpdateCommandUnsupportedPlatform(t *testing.T) {
	setTestHome(t)
	resetUpdateTestDeps(t)
	Version = "0.1.2"
	updateRuntimeGOOS = func() string { return "plan9" }
	updateRuntimeGOARCH = func() string { return "amd64" }

	server := newUpdateTestServer(t, "v0.1.3", "unused", nil, "")
	defer server.Close()
	updateReleaseURL = server.URL + "/latest"

	_, stderr, code := captureRun(t, []string{"update"})
	if code != 1 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "unsupported platform: plan9/amd64") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestUpdateCommandInCatalog(t *testing.T) {
	setTestHome(t)

	stdout, stderr, code := captureRun(t, []string{"commands", "--quiet"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}

	var catalog []commandInfo
	if err := json.Unmarshal([]byte(stdout), &catalog); err != nil {
		t.Fatalf("json error: %v", err)
	}
	for _, command := range catalog {
		if command.Name == "easy8 update" {
			return
		}
	}
	t.Fatalf("expected easy8 update in catalog: %+v", catalog)
}

func TestStartupAutoUpdateRunsOncePerDayAndStaysSilent(t *testing.T) {
	setTestHome(t)
	resetUpdateTestDeps(t)
	t.Setenv("EASY8_BASE_URL", "")
	t.Setenv("EASY8_API_KEY", "")
	t.Setenv("EASY8_AUTOUPDATE", "true")

	statePath := filepath.Join(t.TempDir(), "update-state.yaml")
	autoUpdateStateFilePath = func() (string, error) { return statePath, nil }
	autoUpdateNow = func() time.Time { return time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC) }
	runs := 0
	autoUpdateRunner = func(ctx context.Context) (updateResult, error) {
		runs++
		return updateResult{}, fmt.Errorf("network unavailable")
	}

	stdout, stderr, code := captureRun(t, []string{"auth", "status", "--quiet"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected silent autoupdate, got stderr=%s", stderr)
	}
	if !strings.Contains(stdout, `"authenticated": false`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	if runs != 1 {
		t.Fatalf("runs = %d", runs)
	}

	_, stderr, code = captureRun(t, []string{"auth", "status", "--quiet"})
	if code != 0 {
		t.Fatalf("second code = %d stderr=%s", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected silent second run, got stderr=%s", stderr)
	}
	if runs != 1 {
		t.Fatalf("expected no second autoupdate run, got %d", runs)
	}
}

func TestStartupAutoUpdateCommandFilter(t *testing.T) {
	tests := map[string]bool{
		"issue":    true,
		"pbi":      true,
		"auth":     true,
		"skill":    true,
		"setup":    false,
		"update":   false,
		"commands": false,
		"version":  false,
		"help":     false,
		"--help":   false,
	}
	for command, want := range tests {
		if got := shouldRunStartupAutoUpdate(command); got != want {
			t.Fatalf("shouldRunStartupAutoUpdate(%q) = %t, want %t", command, got, want)
		}
	}
}

func resetUpdateTestDeps(t *testing.T) {
	t.Helper()
	oldVersion := Version
	oldReleaseURL := updateReleaseURL
	oldExecutable := updateExecutable
	oldGOOS := updateRuntimeGOOS
	oldGOARCH := updateRuntimeGOARCH
	oldAutoUpdateRunner := autoUpdateRunner
	oldAutoUpdateNow := autoUpdateNow
	oldAutoUpdateStateFilePath := autoUpdateStateFilePath
	t.Cleanup(func() {
		Version = oldVersion
		updateReleaseURL = oldReleaseURL
		updateExecutable = oldExecutable
		updateRuntimeGOOS = oldGOOS
		updateRuntimeGOARCH = oldGOARCH
		autoUpdateRunner = oldAutoUpdateRunner
		autoUpdateNow = oldAutoUpdateNow
		autoUpdateStateFilePath = oldAutoUpdateStateFilePath
	})
}

func newUpdateTestServer(t *testing.T, tagName, assetName string, binary []byte, checksumOverride string) *httptest.Server {
	t.Helper()

	checksumBytes := sha256.Sum256(binary)
	checksum := fmt.Sprintf("%x", checksumBytes)
	if checksumOverride != "" {
		checksum = checksumOverride
	}

	mux := http.NewServeMux()
	var server *httptest.Server
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		release := githubRelease{
			TagName: tagName,
			Assets: []githubAsset{
				{Name: assetName, BrowserDownloadURL: server.URL + "/download/" + assetName},
				{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums.txt"},
			},
		}
		_ = json.NewEncoder(w).Encode(release)
	})
	mux.HandleFunc("/download/"+assetName, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(binary)
	})
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", checksum, assetName)
	})

	server = httptest.NewServer(mux)
	return server
}
