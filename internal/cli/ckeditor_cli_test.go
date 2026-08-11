package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"easy8-cli/internal/api"
)

func TestIssueUpdateFormatsCKEditorFields(t *testing.T) {
	var seen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/issues/101.json" && r.Method == http.MethodPut:
			seen = true
			var request api.IssueRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if request.Issue.Description == nil || !strings.Contains(*request.Issue.Description, "<h2>Souhrn</h2>") {
				t.Fatalf("description = %v", request.Issue.Description)
			}
			if request.Issue.Notes == nil || !strings.Contains(*request.Issue.Notes, "<li>Pridane testy</li>") {
				t.Fatalf("notes = %v", request.Issue.Notes)
			}
			if request.Issue.AutomationSource == nil || *request.Issue.AutomationSource != api.AutomationSourceEasy8CLI {
				t.Fatalf("automation_source = %v", request.Issue.AutomationSource)
			}
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/issues/101.json" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issue":{"id":101,"subject":"Fix onboarding","status":{"id":1,"name":"New"},"assigned_to":{"id":2,"name":"Alice"},"updated_on":"2024-01-15"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{
		"issue", "update", "101",
		"--description", "## Souhrn\n\nText",
		"--notes", "- Pridane testy\n- Overeno lokalne",
	})
	if code != 0 {
		t.Fatalf("code = %d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !seen {
		t.Fatal("expected update request")
	}
}

func TestIssueCreateFormatsCKEditorDescription(t *testing.T) {
	var seen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		seen = true
		var request api.IssueRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if request.Issue.Description == nil || !strings.Contains(*request.Issue.Description, "<strong>dulezite</strong>") {
			t.Fatalf("description = %v", request.Issue.Description)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issue":{"id":202,"subject":"New task"}}`))
	}))
	defer server.Close()
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{
		"issue", "create",
		"--subject", "New task",
		"--project-id", "1",
		"--tracker-id", "1",
		"--status-id", "1",
		"--priority-id", "1",
		"--author-id", "1",
		"--assigned-to-id", "2",
		"--description", "Tohle je **dulezite**.",
	})
	if code != 0 {
		t.Fatalf("code = %d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !seen {
		t.Fatal("expected create request")
	}
}

func TestPBIUpdateFormatsCKEditorDescription(t *testing.T) {
	var seen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/easy_product_backlog_items/42.json" || r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		seen = true
		var request api.PBIRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if request.PBI.Description == nil || !strings.Contains(*request.PBI.Description, "<ol>") || !strings.Contains(*request.PBI.Description, "<li>Prvni</li>") {
			t.Fatalf("description = %v", request.PBI.Description)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	setTestEnv(t, server.URL)

	stdout, stderr, code := captureRun(t, []string{
		"pbi", "update", "42",
		"--description", "1. Prvni\n2. Druhy",
	})
	if code != 0 {
		t.Fatalf("code = %d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !seen {
		t.Fatal("expected update request")
	}
}
