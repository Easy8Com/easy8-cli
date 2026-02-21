package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"easy8-cli/internal/api"
)

func TestRunNoArgs(t *testing.T) {
	setTestHome(t)

	code := Run([]string{})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	setTestHome(t)

	code := Run([]string{"nope"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
}

func TestIssueCreateMissingSubject(t *testing.T) {
	setTestHome(t)

	code := Run([]string{"issue", "create", "--project-id", "1", "--tracker-id", "1", "--status-id", "1", "--priority-id", "1", "--author-id", "1", "--assigned-to-id", "2"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
}

func TestIssueSearchMissingQuery(t *testing.T) {
	setTestHome(t)

	code := Run([]string{"issue", "search"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
}

func TestIssueListTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "list"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Subject") || !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueListJSONOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "list", "--json"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	var resp api.IssueListResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if len(resp.Issues) != 1 || resp.Issues[0].ID != 101 {
		t.Fatalf("unexpected issues: %+v", resp.Issues)
	}
}

func TestIssueCreateJSONOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	args := []string{"issue", "create", "--subject", "New task", "--project-id", "1", "--tracker-id", "1", "--status-id", "1", "--priority-id", "1", "--author-id", "1", "--assigned-to-id", "2", "--json"}
	stdout, stderr, code := captureRun(t, args)
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	var resp api.IssueResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp.Issue.ID != 202 {
		t.Fatalf("unexpected issue id: %d", resp.Issue.ID)
	}
}

func TestIssueUpdateTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "update", "--id", "101", "--status-id", "2"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueSearchTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "search", "--q", "onboarding"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "issue") || !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueListAPIError(t *testing.T) {
	server := newErrorServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "list"})
	if code != 1 {
		t.Fatalf("code = %d stdout=%s", code, stdout)
	}
	if !strings.Contains(stderr, "api error 500") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func setTestHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func setTestEnv(t *testing.T, baseURL string) {
	t.Helper()
	setTestHome(t)
	t.Setenv("EASY8_BASE_URL", baseURL)
	t.Setenv("EASY8_API_KEY", "test-key")
}

func captureRun(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	code := Run(args)

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if _, err := stdout.ReadFrom(stdoutReader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrReader); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	_ = stdoutReader.Close()
	_ = stderrReader.Close()

	return stdout.String(), stderr.String(), code
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	handler := http.NewServeMux()
	handler.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"issues\":[{\"id\":101,\"subject\":\"Fix onboarding\",\"status\":{\"id\":1,\"name\":\"New\"},\"assigned_to\":{\"id\":2,\"name\":\"Alice\"},\"updated_on\":\"2024-01-01\"}],\"total_count\":1,\"offset\":0,\"limit\":25}"))
			return
		}
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"issue\":{\"id\":202,\"subject\":\"New task\"}}"))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	handler.HandleFunc("/issues/101.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issue\":{\"id\":101,\"subject\":\"Fix onboarding\",\"status\":{\"id\":2,\"name\":\"In Progress\"}}}"))
	})

	handler.HandleFunc("/search.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"results\":[{\"id\":101,\"type\":\"issue\",\"title\":\"Fix onboarding\",\"url\":\"https://example.com/issues/101\",\"description\":\"\",\"datetime\":\"2024-01-01\"}],\"total_count\":1,\"offset\":0,\"limit\":25}"))
	})

	return httptest.NewServer(handler)
}

func newErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	handler := http.NewServeMux()
	handler.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})
	return httptest.NewServer(handler)
}
