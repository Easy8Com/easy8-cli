package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListIssuesBuildsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("limit") != "10" {
			t.Fatalf("limit = %s", query.Get("limit"))
		}
		if query.Get("offset") != "5" {
			t.Fatalf("offset = %s", query.Get("offset"))
		}
		if query.Get("sort") != "priority:desc" {
			t.Fatalf("sort = %s", query.Get("sort"))
		}
		if query.Get("set_filter") != "1" {
			t.Fatalf("set_filter = %s", query.Get("set_filter"))
		}
		if query.Get("easy_query_q") != "onboarding" {
			t.Fatalf("easy_query_q = %s", query.Get("easy_query_q"))
		}
		if query.Get("include") != "attachments,relations" {
			t.Fatalf("include = %s", query.Get("include"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issues\":[{\"id\":1,\"subject\":\"Test\"}],\"total_count\":1,\"offset\":0,\"limit\":25}"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	params := IssueListParams{
		Limit:   10,
		Offset:  5,
		Sort:    "priority:desc",
		Query:   "onboarding",
		Include: []string{"attachments", "relations"},
	}
	resp, err := client.ListIssues(context.Background(), params)
	if err != nil {
		t.Fatalf("ListIssues error: %v", err)
	}
	if len(resp.Issues) != 1 || resp.Issues[0].ID != 1 {
		t.Fatalf("unexpected issues: %+v", resp.Issues)
	}
}

func TestCreateIssueSendsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		var request IssueRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if request.Issue.Subject == nil || *request.Issue.Subject != "New" {
			t.Fatalf("subject = %v", request.Issue.Subject)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issue\":{\"id\":10,\"subject\":\"New\"}}"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	subject := "New"
	projectID := 1
	input := IssueInput{
		Subject:   &subject,
		ProjectID: &projectID,
	}
	resp, err := client.CreateIssue(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateIssue error: %v", err)
	}
	if resp.Issue.ID != 10 {
		t.Fatalf("issue id = %d", resp.Issue.ID)
	}
}

func TestUpdateIssueMissingID(t *testing.T) {
	client := &Client{BaseURL: "https://example.com", APIKey: "key", HTTP: http.DefaultClient}
	_, err := client.UpdateIssue(context.Background(), 0, IssueInput{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSearchBuildsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("q") != "hello" {
			t.Fatalf("q = %s", query.Get("q"))
		}
		if query.Get("open_issues") != "1" {
			t.Fatalf("open_issues = %s", query.Get("open_issues"))
		}
		if query.Get("scope") != "7" {
			t.Fatalf("scope = %s", query.Get("scope"))
		}
		if query.Get("issues") != "1" {
			t.Fatalf("issues = %s", query.Get("issues"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"results\":[{\"id\":5,\"type\":\"issue\",\"title\":\"Hello\",\"url\":\"https://example.com\",\"description\":\"\",\"datetime\":\"2024-01-01\"}],\"total_count\":1,\"offset\":0,\"limit\":25}"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	resp, err := client.Search(context.Background(), SearchParams{Query: "hello", OpenIssues: true, Scope: 7, IssuesOnly: true})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].ID != 5 {
		t.Fatalf("unexpected results: %+v", resp.Results)
	}
}

func TestAPIErrorIncludesBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	_, err := client.ListIssues(context.Background(), IssueListParams{})
	if err == nil {
		t.Fatalf("expected error")
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError")
	}
	if !strings.Contains(err.Error(), "api error 400") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingAPIKey(t *testing.T) {
	client := &Client{BaseURL: "https://example.com", APIKey: "", HTTP: http.DefaultClient}
	_, err := client.ListIssues(context.Background(), IssueListParams{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
