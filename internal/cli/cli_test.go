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

func TestVersionCommand(t *testing.T) {
	setTestHome(t)

	stdout, _, code := captureRun(t, []string{"version"})
	if code != 0 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stdout, "dev") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	setTestHome(t)

	code := Run([]string{"nope"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
}

func TestIssueCreateDoneRatioInvalid(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	args := []string{"issue", "create", "--subject", "Test", "--project-id", "1", "--tracker-id", "1", "--status-id", "1", "--priority-id", "1", "--author-id", "1", "--assigned-to-id", "2", "--done-ratio", "150"}
	_, stderr, code := captureRun(t, args)
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "done-ratio must be between 0 and 100") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestIssueUpdateDoneRatioInvalid(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	_, stderr, code := captureRun(t, []string{"issue", "update", "--id", "101", "--done-ratio", "-5"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "done-ratio must be between 0 and 100") {
		t.Fatalf("unexpected stderr: %s", stderr)
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

	_, stderr, code := captureRun(t, []string{"issue", "search"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "at least one filter is required") {
		t.Fatalf("unexpected stderr: %s", stderr)
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

func TestIssueUpdateWithNotesRefetch(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "update", "--id", "101", "--notes", "a test note"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	// After PUT (empty body), the CLI re-fetches via GET and displays the issue.
	if !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("expected re-fetched issue in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "101") {
		t.Fatalf("expected issue ID in stdout: %s", stdout)
	}
}

func TestIssueShowTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "show", "--id", "101"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("expected subject in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "Alice") {
		t.Fatalf("expected assignee in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "30%") {
		t.Fatalf("expected done ratio in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "Onboarding flow needs fixing") {
		t.Fatalf("expected description in stdout: %s", stdout)
	}
}

func TestIssueShowJSONOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "show", "--id", "101", "--json"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	var resp api.IssueResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp.Issue.ID != 101 {
		t.Fatalf("issue id = %d", resp.Issue.ID)
	}
	if resp.Issue.Description != "Onboarding flow needs fixing" {
		t.Fatalf("description = %q", resp.Issue.Description)
	}
}

func TestIssueShowMissingID(t *testing.T) {
	setTestHome(t)

	_, stderr, code := captureRun(t, []string{"issue", "show"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "--id is required") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestIssueShowWithJournalsAndAttachments(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "show", "--id", "101", "--include", "journals,attachments"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	// Basic fields still present
	if !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("expected subject in stdout: %s", stdout)
	}
	// Journals section
	if !strings.Contains(stdout, "Journals:") {
		t.Fatalf("expected Journals section in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "Bob") {
		t.Fatalf("expected journal author Bob in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "Started work on this") {
		t.Fatalf("expected journal notes in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "status_id") {
		t.Fatalf("expected journal detail name in stdout: %s", stdout)
	}
	// Attachments section
	if !strings.Contains(stdout, "Attachments:") {
		t.Fatalf("expected Attachments section in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "design.pdf") {
		t.Fatalf("expected attachment filename in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "200.0 KB") {
		t.Fatalf("expected attachment filesize in stdout: %s", stdout)
	}
}

func TestIssueShowWithJournalsJSON(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "show", "--id", "101", "--include", "journals,attachments", "--json"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	var resp api.IssueResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if len(resp.Issue.Journals) != 2 {
		t.Fatalf("journals count = %d", len(resp.Issue.Journals))
	}
	if resp.Issue.Journals[0].Notes != "Started work on this" {
		t.Fatalf("journal notes = %q", resp.Issue.Journals[0].Notes)
	}
	if len(resp.Issue.Journals[1].Details) != 1 {
		t.Fatalf("journal details count = %d", len(resp.Issue.Journals[1].Details))
	}
	if len(resp.Issue.Attachments) != 1 {
		t.Fatalf("attachments count = %d", len(resp.Issue.Attachments))
	}
	if resp.Issue.Attachments[0].Filename != "design.pdf" {
		t.Fatalf("attachment filename = %q", resp.Issue.Attachments[0].Filename)
	}
}

func TestIssueSearchTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"issue", "search", "--q", "onboarding"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Fix onboarding") {
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

func TestIssueSearchNameFilters(t *testing.T) {
	server := newLookupServer(t)
	setTestEnv(t, server.URL)

	args := []string{
		"issue", "search", "--q", "petr",
		"--assignee", "Alice Doe",
		"--status", "New",
		"--priority", "High",
		"--task-type", "Task",
		"--project", "Project A",
	}
	stdout, stderr, code := captureRun(t, args)
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Fix onboarding") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueSearchNameConflict(t *testing.T) {
	server := newLookupServer(t)
	setTestEnv(t, server.URL)

	args := []string{"issue", "search", "--q", "petr", "--status-id", "1", "--status", "New"}
	_, stderr, code := captureRun(t, args)
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "status-id does not match status name") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

// --- PBI tests ---

func TestPBIListTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"pbi", "list", "--limit", "5"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "PBI Alpha") {
		t.Fatalf("expected name in stdout: %s", stdout)
	}
	if !strings.Contains(stdout, "to_do") {
		t.Fatalf("expected status in stdout: %s", stdout)
	}
}

func TestPBIListJSONOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"pbi", "list", "--json"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	var resp api.PBIListResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if len(resp.PBIs) != 1 || resp.PBIs[0].ID != 1 {
		t.Fatalf("unexpected: %+v", resp)
	}
}

func TestPBIShowTableOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"pbi", "show", "--id", "1"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "PBI Alpha") {
		t.Fatalf("expected name: %s", stdout)
	}
	if !strings.Contains(stdout, "Design") {
		t.Fatalf("expected board: %s", stdout)
	}
	if !strings.Contains(stdout, "Full description") {
		t.Fatalf("expected description: %s", stdout)
	}
	// Issue details should include status, assignee, done ratio (fetched via batch)
	if !strings.Contains(stdout, "Task 100") {
		t.Fatalf("expected issue subject: %s", stdout)
	}
	if !strings.Contains(stdout, "New") {
		t.Fatalf("expected issue status: %s", stdout)
	}
	if !strings.Contains(stdout, "Alice") {
		t.Fatalf("expected issue assignee: %s", stdout)
	}
	if !strings.Contains(stdout, "50%") {
		t.Fatalf("expected issue done ratio: %s", stdout)
	}
	if !strings.Contains(stdout, "Review") {
		t.Fatalf("expected sticky note: %s", stdout)
	}
}

func TestPBIShowJSONOutput(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"pbi", "show", "--id", "1", "--json"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	var resp api.PBIResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp.PBI.ID != 1 {
		t.Fatalf("id = %d", resp.PBI.ID)
	}
}

func TestPBIShowMissingID(t *testing.T) {
	setTestHome(t)

	_, stderr, code := captureRun(t, []string{"pbi", "show"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "--id is required") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestPBIUpdateSuccess(t *testing.T) {
	server := newTestServer(t)
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{"pbi", "update", "--id", "1", "--status", "done"})
	if code != 0 {
		t.Fatalf("code = %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "PBI #1 updated") {
		t.Fatalf("expected update message: %s", stdout)
	}
}

func TestPBIUpdateMissingID(t *testing.T) {
	setTestHome(t)

	_, stderr, code := captureRun(t, []string{"pbi", "update", "--status", "done"})
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr, "--id is required") {
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
			// If issue_id filter is set, return matching issues with full details
			if issueID := r.URL.Query().Get("issue_id"); issueID == "100" {
				_, _ = w.Write([]byte(`{"issues":[{"id":100,"subject":"Task 100","status":{"id":1,"name":"New"},"assigned_to":{"id":2,"name":"Alice"},"done_ratio":50,"updated_on":"2024-01-10"}],"total_count":1,"offset":0,"limit":25}`))
				return
			}
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
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			inc := r.URL.Query().Get("include")
			if strings.Contains(inc, "journals") || strings.Contains(inc, "attachments") {
				_, _ = w.Write([]byte(`{"issue":{"id":101,"subject":"Fix onboarding","description":"Onboarding flow needs fixing","status":{"id":1,"name":"New"},"priority":{"id":4,"name":"Normal"},"project":{"id":5,"name":"Alpha"},"tracker":{"id":7,"name":"Task"},"author":{"id":1,"name":"Bob"},"assigned_to":{"id":2,"name":"Alice"},"done_ratio":30,"start_date":"2024-01-01","due_date":"2024-02-01","created_on":"2024-01-01","updated_on":"2024-01-15","journals":[{"id":500,"user":{"id":1,"name":"Bob"},"notes":"Started work on this","created_on":"2024-01-02","private_notes":false,"details":[]},{"id":501,"user":{"id":2,"name":"Alice"},"notes":"","created_on":"2024-01-03","private_notes":false,"details":[{"property":"attr","name":"status_id","old_value":"1","new_value":"2"}]}],"attachments":[{"id":300,"filename":"design.pdf","filesize":204800,"content_type":"application/pdf","description":"Design doc","version":1,"content_url":"https://example.com/att/300","author":{"id":1,"name":"Bob"},"created_on":"2024-01-01"}]}}`))
			} else {
				_, _ = w.Write([]byte(`{"issue":{"id":101,"subject":"Fix onboarding","description":"Onboarding flow needs fixing","status":{"id":1,"name":"New"},"priority":{"id":4,"name":"Normal"},"project":{"id":5,"name":"Alpha"},"tracker":{"id":7,"name":"Task"},"author":{"id":1,"name":"Bob"},"assigned_to":{"id":2,"name":"Alice"},"done_ratio":30,"start_date":"2024-01-01","due_date":"2024-02-01","created_on":"2024-01-01","updated_on":"2024-01-15"}}`))
			}
			return
		}
		if r.Method == http.MethodPut {
			// Real Redmine returns 200 with empty body on successful PUT.
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	handler.HandleFunc("/easy_product_backlog_items.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"easy_product_backlog_items":[{"id":1,"name":"PBI Alpha","status":"to_do","estimate":"3","author":{"id":1,"name":"Alice"},"updated_at":"2024-01-01"}],"total_count":1,"offset":0,"limit":25}`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	handler.HandleFunc("/easy_product_backlog_items/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"easy_product_backlog_item":{"id":1,"name":"PBI Alpha","description":"Full description","estimate":"3","estimate_float":3.0,"status":"to_do","author":{"id":1,"name":"Alice"},"easy_product_backlog_board":{"id":17,"name":"Design"},"issues":[{"id":100,"subject":"Task 100"}],"easy_sticky_notes":[{"id":10,"name":"Review","status":"done"}],"created_at":"2024-01-01","updated_at":"2024-01-15"}}`))
			return
		}
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return httptest.NewServer(handler)
}

func newLookupServer(t *testing.T) *httptest.Server {
	t.Helper()

	handler := http.NewServeMux()
	handler.HandleFunc("/users.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"users\":[{\"id\":11,\"login\":\"alice\",\"firstname\":\"Alice\",\"lastname\":\"Doe\"}],\"total_count\":1,\"offset\":0,\"limit\":100}"))
	})
	handler.HandleFunc("/issue_statuses.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issue_statuses\":[{\"id\":2,\"name\":\"New\"}]}"))
	})
	handler.HandleFunc("/enumerations/issue_priorities.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issue_priorities\":[{\"id\":3,\"name\":\"High\"}]}"))
	})
	handler.HandleFunc("/trackers.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"trackers\":[{\"id\":4,\"name\":\"Task\"}]}"))
	})
	handler.HandleFunc("/projects.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"projects\":[{\"id\":5,\"name\":\"Project A\"}],\"total_count\":1,\"offset\":0,\"limit\":100}"))
	})
	handler.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("assigned_to_id") != "11" {
			t.Fatalf("assigned_to_id = %s", query.Get("assigned_to_id"))
		}
		if query.Get("status_id") != "2" {
			t.Fatalf("status_id = %s", query.Get("status_id"))
		}
		if query.Get("priority_id") != "3" {
			t.Fatalf("priority_id = %s", query.Get("priority_id"))
		}
		if query.Get("tracker_id") != "4" {
			t.Fatalf("tracker_id = %s", query.Get("tracker_id"))
		}
		if query.Get("project_id") != "5" {
			t.Fatalf("project_id = %s", query.Get("project_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issues\":[{\"id\":101,\"subject\":\"Fix onboarding\",\"status\":{\"id\":1,\"name\":\"New\"},\"assigned_to\":{\"id\":2,\"name\":\"Alice\"},\"updated_on\":\"2024-01-01\"}],\"total_count\":1,\"offset\":0,\"limit\":25}"))
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
