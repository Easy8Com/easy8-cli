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
		if query.Get("assigned_to_id") != "42" {
			t.Fatalf("assigned_to_id = %s", query.Get("assigned_to_id"))
		}
		if query.Get("due_date") != "2024-01-10" {
			t.Fatalf("due_date = %s", query.Get("due_date"))
		}
		if query.Get("status_id") != "2" {
			t.Fatalf("status_id = %s", query.Get("status_id"))
		}
		if query.Get("priority_id") != "3" {
			t.Fatalf("priority_id = %s", query.Get("priority_id"))
		}
		if query.Get("subject") != "Fix" {
			t.Fatalf("subject = %s", query.Get("subject"))
		}
		if query.Get("tracker_id") != "7" {
			t.Fatalf("tracker_id = %s", query.Get("tracker_id"))
		}
		if query.Get("project_id") != "11" {
			t.Fatalf("project_id = %s", query.Get("project_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"issues\":[{\"id\":1,\"subject\":\"Test\"}],\"total_count\":1,\"offset\":0,\"limit\":25}"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	params := IssueListParams{
		Limit:      10,
		Offset:     5,
		Sort:       "priority:desc",
		Query:      "onboarding",
		Include:    []string{"attachments", "relations"},
		AssigneeID: 42,
		DueDate:    "2024-01-10",
		StatusID:   2,
		PriorityID: 3,
		Subject:    "Fix",
		TaskTypeID: 7,
		ProjectID:  11,
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

func TestLookupEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/trackers.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"trackers\":[{\"id\":1,\"name\":\"Task\"}]}"))
		case "/issue_statuses.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"issue_statuses\":[{\"id\":2,\"name\":\"New\"}]}"))
		case "/enumerations/issue_priorities.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"issue_priorities\":[{\"id\":3,\"name\":\"High\"}]}"))
		case "/users.json":
			q := r.URL.Query()
			if q.Get("offset") == "0" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{\"users\":[{\"id\":10,\"login\":\"alice\",\"firstname\":\"Alice\",\"lastname\":\"Doe\"}],\"total_count\":2,\"offset\":0,\"limit\":1}"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"users\":[{\"id\":11,\"login\":\"bob\",\"firstname\":\"Bob\",\"lastname\":\"Smith\"}],\"total_count\":2,\"offset\":1,\"limit\":1}"))
		case "/projects.json":
			q := r.URL.Query()
			if q.Get("offset") == "0" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{\"projects\":[{\"id\":20,\"name\":\"Alpha\"}],\"total_count\":2,\"offset\":0,\"limit\":1}"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"projects\":[{\"id\":21,\"name\":\"Beta\"}],\"total_count\":2,\"offset\":1,\"limit\":1}"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	trackers, err := client.ListTrackers(context.Background())
	if err != nil || len(trackers) != 1 || trackers[0].ID != 1 {
		t.Fatalf("trackers: %v %v", trackers, err)
	}
	statuses, err := client.ListIssueStatuses(context.Background())
	if err != nil || len(statuses) != 1 || statuses[0].ID != 2 {
		t.Fatalf("statuses: %v %v", statuses, err)
	}
	priorities, err := client.ListIssuePriorities(context.Background())
	if err != nil || len(priorities) != 1 || priorities[0].ID != 3 {
		t.Fatalf("priorities: %v %v", priorities, err)
	}
	users, err := client.ListUsers(context.Background())
	if err != nil || len(users) != 2 {
		t.Fatalf("users: %v %v", users, err)
	}
	projects, err := client.ListProjects(context.Background())
	if err != nil || len(projects) != 2 {
		t.Fatalf("projects: %v %v", projects, err)
	}
}
