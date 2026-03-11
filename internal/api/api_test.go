package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"easy8-cli/internal/config"
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

func TestListIssuesByIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("issue_id") != "100|200|300" {
			t.Fatalf("issue_id = %s", q.Get("issue_id"))
		}
		if q.Get("set_filter") != "1" {
			t.Fatalf("set_filter = %s", q.Get("set_filter"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"id":100,"subject":"A","status":{"id":1,"name":"New"}},{"id":200,"subject":"B","status":{"id":2,"name":"Done"}},{"id":300,"subject":"C","status":{"id":1,"name":"New"}}],"total_count":3,"offset":0,"limit":25}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	resp, err := client.ListIssues(context.Background(), IssueListParams{
		IssueIDs: []int{100, 200, 300},
		Limit:    3,
	})
	if err != nil {
		t.Fatalf("ListIssues error: %v", err)
	}
	if len(resp.Issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(resp.Issues))
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

func TestGetIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/issues/42.json" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("include") != "journals,attachments" {
			t.Fatalf("include = %s", r.URL.Query().Get("include"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issue":{"id":42,"subject":"Test issue","description":"desc","status":{"id":1,"name":"New"},"project":{"id":5,"name":"Alpha"},"tracker":{"id":7,"name":"Task"},"priority":{"id":4,"name":"Normal"},"author":{"id":51,"name":"Petr"},"assigned_to":{"id":51,"name":"Petr"},"done_ratio":50,"start_date":"2024-01-01","due_date":"2024-02-01","created_on":"2024-01-01","updated_on":"2024-01-15","journals":[{"id":100,"user":{"id":51,"name":"Petr"},"notes":"fixed the bug","created_on":"2024-01-10","private_notes":false,"details":[{"property":"attr","name":"status_id","old_value":"1","new_value":"2"}]}],"attachments":[{"id":200,"filename":"screenshot.png","filesize":51200,"content_type":"image/png","description":"UI screenshot","version":1,"content_url":"https://example.com/att/200","author":{"id":51,"name":"Petr"},"created_on":"2024-01-05"}]}}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	resp, err := client.GetIssue(context.Background(), 42, []string{"journals", "attachments"})
	if err != nil {
		t.Fatalf("GetIssue error: %v", err)
	}
	if resp.Issue.ID != 42 {
		t.Fatalf("issue id = %d", resp.Issue.ID)
	}
	if resp.Issue.Subject != "Test issue" {
		t.Fatalf("subject = %q", resp.Issue.Subject)
	}
	if resp.Issue.DoneRatio != 50 {
		t.Fatalf("done_ratio = %d", resp.Issue.DoneRatio)
	}
	if len(resp.Issue.Journals) != 1 {
		t.Fatalf("journals count = %d", len(resp.Issue.Journals))
	}
	j := resp.Issue.Journals[0]
	if j.ID != 100 {
		t.Fatalf("journal id = %d", j.ID)
	}
	if j.Notes != "fixed the bug" {
		t.Fatalf("journal notes = %q", j.Notes)
	}
	if len(j.Details) != 1 || j.Details[0].Name != "status_id" {
		t.Fatalf("journal details = %+v", j.Details)
	}
	if j.Details[0].OldValue != "1" || j.Details[0].NewValue != "2" {
		t.Fatalf("journal detail values = %q -> %q", j.Details[0].OldValue, j.Details[0].NewValue)
	}
	if len(resp.Issue.Attachments) != 1 {
		t.Fatalf("attachments count = %d", len(resp.Issue.Attachments))
	}
	a := resp.Issue.Attachments[0]
	if a.ID != 200 {
		t.Fatalf("attachment id = %d", a.ID)
	}
	if a.Filename != "screenshot.png" {
		t.Fatalf("attachment filename = %q", a.Filename)
	}
	if a.Filesize != 51200 {
		t.Fatalf("attachment filesize = %d", a.Filesize)
	}
}

func TestGetIssueNoInclude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include") != "" {
			t.Fatalf("unexpected include param: %s", r.URL.Query().Get("include"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issue":{"id":10,"subject":"Simple"}}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	resp, err := client.GetIssue(context.Background(), 10, nil)
	if err != nil {
		t.Fatalf("GetIssue error: %v", err)
	}
	if resp.Issue.ID != 10 {
		t.Fatalf("issue id = %d", resp.Issue.ID)
	}
}

func TestGetIssueMissingID(t *testing.T) {
	client := &Client{BaseURL: "https://example.com", APIKey: "key", HTTP: http.DefaultClient}
	_, err := client.GetIssue(context.Background(), 0, nil)
	if err == nil {
		t.Fatalf("expected error")
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

func TestUpdateIssueEmptyResponse(t *testing.T) {
	var receivedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		var request IssueRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if request.Issue.Notes == nil || *request.Issue.Notes != "test note" {
			t.Fatalf("notes = %v", request.Issue.Notes)
		}
		// Redmine returns 200 with empty body on successful PUT.
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	notes := "test note"
	resp, err := client.UpdateIssue(context.Background(), 42, IssueInput{Notes: &notes})
	if err != nil {
		t.Fatalf("UpdateIssue error: %v", err)
	}
	if receivedMethod != http.MethodPut {
		t.Fatalf("server received %s, want PUT", receivedMethod)
	}
	// Empty response means zero-value issue.
	if resp.Issue.ID != 0 {
		t.Fatalf("expected zero ID from empty response, got %d", resp.Issue.ID)
	}
}

func TestRedirectPreservesMethod(t *testing.T) {
	var finalMethod string
	// Target server that records the received method.
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	// Redirect server that sends 302 (which would change PUT→GET by default).
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.Path, http.StatusFound)
	}))
	defer redirect.Close()

	client := NewClient(config.Config{BaseURL: redirect.URL, APIKey: "key"})
	notes := "test"
	_, err := client.UpdateIssue(context.Background(), 1, IssueInput{Notes: &notes})
	// The redirect changes PUT→GET, so our CheckRedirect should refuse it.
	// doJSON sees the 302 status and returns APIError.
	if err == nil {
		t.Fatal("expected error for method-changing redirect")
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusFound {
		t.Fatalf("status = %d, want %d", apiErr.StatusCode, http.StatusFound)
	}
	// Target should NOT have been hit, or if it was, it must be a GET (which we prevented).
	if finalMethod == http.MethodPut {
		t.Fatal("redirect should not have forwarded the PUT")
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
			if q.Get("set_filter") != "1" {
				t.Fatalf("users: set_filter = %s", q.Get("set_filter"))
			}
			if q.Get("offset") == "0" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{\"users\":[{\"id\":10,\"login\":\"alice\",\"firstname\":\"Alice\",\"lastname\":\"Doe\"}],\"total_count\":2,\"offset\":0,\"limit\":1}"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"users\":[{\"id\":11,\"login\":\"bob\",\"firstname\":\"Bob\",\"lastname\":\"Smith\"}],\"total_count\":2,\"offset\":1,\"limit\":1}"))
		case "/projects.json":
			q := r.URL.Query()
			if q.Get("set_filter") != "1" {
				t.Fatalf("projects: set_filter = %s", q.Get("set_filter"))
			}
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
