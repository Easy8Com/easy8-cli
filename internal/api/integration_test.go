//go:build integration

package api

import (
	"context"
	"easy8-cli/internal/config"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

// integrationClient returns a configured Client for integration tests.
// It skips the test if EASY8_BASE_URL or EASY8_API_KEY are not set.
// Usage: source ./setup_env.sh && go test -tags integration -v -timeout 600s ./internal/api/
func integrationClient(t *testing.T) *Client {
	t.Helper()
	baseURL := os.Getenv("EASY8_BASE_URL")
	apiKey := os.Getenv("EASY8_API_KEY")
	if baseURL == "" || apiKey == "" {
		t.Skip("EASY8_BASE_URL and EASY8_API_KEY must be set for integration tests")
	}
	cfg := config.Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
	client := NewClient(cfg)
	client.HTTP = &http.Client{Timeout: 120 * time.Second}
	return client
}

func TestIntegrationListIssues(t *testing.T) {
	client := integrationClient(t)

	resp, err := client.ListIssues(context.Background(), IssueListParams{Limit: 3})
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	t.Logf("ListIssues: %d issues returned (total_count=%d)", len(resp.Issues), resp.TotalCount)
	for _, issue := range resp.Issues {
		t.Logf("  #%d %s", issue.ID, issue.Subject)
	}
}

func TestIntegrationLookups(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	trackers, err := client.ListTrackers(ctx)
	if err != nil {
		t.Fatalf("ListTrackers: %v", err)
	}
	if len(trackers) == 0 {
		t.Fatal("ListTrackers: 0 items")
	}
	t.Logf("Trackers: %d items (first: id=%d name=%q)", len(trackers), trackers[0].ID, trackers[0].Name)

	statuses, err := client.ListIssueStatuses(ctx)
	if err != nil {
		t.Fatalf("ListIssueStatuses: %v", err)
	}
	if len(statuses) == 0 {
		t.Fatal("ListIssueStatuses: 0 items")
	}
	t.Logf("Statuses: %d items (first: id=%d name=%q)", len(statuses), statuses[0].ID, statuses[0].Name)

	priorities, err := client.ListIssuePriorities(ctx)
	if err != nil {
		t.Fatalf("ListIssuePriorities: %v", err)
	}
	if len(priorities) == 0 {
		t.Fatal("ListIssuePriorities: 0 items")
	}
	t.Logf("Priorities: %d items (first: id=%d name=%q)", len(priorities), priorities[0].ID, priorities[0].Name)

	users, err := client.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) == 0 {
		t.Fatal("ListUsers: 0 items")
	}
	t.Logf("Users: %d items (first: id=%d login=%q)", len(users), users[0].ID, users[0].Login)

	projects, err := client.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) == 0 {
		t.Fatal("ListProjects: 0 items")
	}
	t.Logf("Projects: %d items (first: id=%d name=%q)", len(projects), projects[0].ID, projects[0].Name)
}

func TestIntegrationCreateUpdateIssue(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	// Read required IDs from env -- these must be valid for the target server.
	projectID := envInt(t, "EASY8_DEFAULT_PROJECT_ID")
	trackerID := envInt(t, "EASY8_DEFAULT_TRACKER_ID")
	statusID := envInt(t, "EASY8_DEFAULT_STATUS_ID")
	priorityID := envInt(t, "EASY8_DEFAULT_PRIORITY_ID")
	authorID := envInt(t, "EASY8_DEFAULT_AUTHOR_ID")
	assigneeID := envInt(t, "EASY8_DEFAULT_ASSIGNED_TO_ID")

	// Create
	subject := "[integration-test] Smoke test issue"
	description := "Created by easy8-cli integration test. Safe to delete."
	input := IssueInput{
		Subject:      &subject,
		ProjectID:    &projectID,
		TrackerID:    &trackerID,
		StatusID:     &statusID,
		PriorityID:   &priorityID,
		AuthorID:     &authorID,
		AssignedToID: &assigneeID,
		Description:  &description,
	}

	createResp, err := client.CreateIssue(ctx, input)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	issueID := createResp.Issue.ID
	if issueID == 0 {
		t.Fatal("CreateIssue returned issue ID 0")
	}
	t.Logf("Created issue #%d", issueID)

	attachmentFile, err := os.CreateTemp(t.TempDir(), "easy8-attachment-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer attachmentFile.Close()
	attachmentContent := "Attached by easy8-cli integration test."
	if _, err := attachmentFile.WriteString(attachmentContent); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if _, err := attachmentFile.Seek(0, 0); err != nil {
		t.Fatalf("Seek: %v", err)
	}
	attachmentInfo, err := attachmentFile.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	attachmentDescription := "Attached by integration test."
	uploadResp, err := client.UploadAttachment(ctx, attachmentInfo.Name(), attachmentFile)
	if err != nil {
		t.Fatalf("UploadAttachment: %v", err)
	}

	// Update
	updatedSubject := "[integration-test] Updated smoke test"
	notes := "Updated by integration test."
	updateInput := IssueInput{
		Subject: &updatedSubject,
		Notes:   &notes,
		Uploads: []IssueUpload{{
			Token:       uploadResp.Upload.Token,
			Filename:    attachmentInfo.Name(),
			Description: attachmentDescription,
		}},
	}

	_, err = client.UpdateIssue(ctx, issueID, updateInput)
	if err != nil {
		t.Fatalf("UpdateIssue #%d: %v", issueID, err)
	}

	// Verify the update actually took effect by re-fetching.
	getResp, err := client.GetIssue(ctx, issueID, []string{"attachments"})
	if err != nil {
		t.Fatalf("GetIssue #%d after update: %v", issueID, err)
	}
	if getResp.Issue.Subject != updatedSubject {
		t.Fatalf("subject after update = %q, want %q", getResp.Issue.Subject, updatedSubject)
	}
	matchedAttachment := false
	for _, attachment := range getResp.Issue.Attachments {
		if attachment.Filename != attachmentInfo.Name() {
			continue
		}
		matchedAttachment = true
		if attachment.Description != attachmentDescription {
			t.Fatalf("attachment description = %q, want %q", attachment.Description, attachmentDescription)
		}
		break
	}
	if !matchedAttachment {
		t.Fatalf("attachment %q not found in issue attachments", attachmentInfo.Name())
	}
	t.Logf("Verified update for issue #%d: %s", getResp.Issue.ID, getResp.Issue.Subject)
}

// envInt reads a required integer env var; skips the test if missing or invalid.
func envInt(t *testing.T, key string) int {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skipf("%s must be set for this test", key)
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		t.Skipf("%s is not a valid integer: %q", key, val)
	}
	return n
}

func TestIntegrationSearchIssues(t *testing.T) {
	client := integrationClient(t)

	resp, err := client.ListIssues(context.Background(), IssueListParams{
		Limit: 3,
		Query: "test",
	})
	if err != nil {
		t.Fatalf("ListIssues (search): %v", err)
	}
	t.Logf("Search 'test': %d issues returned (total_count=%d)", len(resp.Issues), resp.TotalCount)
	for _, issue := range resp.Issues {
		t.Logf("  #%d %s", issue.ID, issue.Subject)
	}
}

func TestIntegrationShowIssue(t *testing.T) {
	client := integrationClient(t)

	// First list to get a valid issue ID
	listResp, err := client.ListIssues(context.Background(), IssueListParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(listResp.Issues) == 0 {
		t.Skip("no issues found to show")
	}
	issueID := listResp.Issues[0].ID

	// Get the issue detail
	resp, err := client.GetIssue(context.Background(), issueID, nil)
	if err != nil {
		t.Fatalf("GetIssue #%d: %v", issueID, err)
	}
	if resp.Issue.ID != issueID {
		t.Fatalf("expected issue %d, got %d", issueID, resp.Issue.ID)
	}
	t.Logf("Issue #%d: %s (status=%s, assignee=%s)",
		resp.Issue.ID, resp.Issue.Subject,
		nameOrEmpty(resp.Issue.Status), nameOrEmpty(resp.Issue.AssignedTo))
}

// --- PBI integration tests ---

func TestIntegrationListPBIs(t *testing.T) {
	client := integrationClient(t)

	resp, err := client.ListPBIs(context.Background(), PBIListParams{Limit: 3})
	if err != nil {
		t.Fatalf("ListPBIs: %v", err)
	}
	t.Logf("ListPBIs: %d items returned (total_count=%d)", len(resp.PBIs), resp.TotalCount)
	for _, pbi := range resp.PBIs {
		t.Logf("  #%d %s [%s]", pbi.ID, pbi.Name, pbi.Status)
	}
}

func TestIntegrationShowPBI(t *testing.T) {
	client := integrationClient(t)

	listResp, err := client.ListPBIs(context.Background(), PBIListParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListPBIs: %v", err)
	}
	if len(listResp.PBIs) == 0 {
		t.Skip("no PBIs found")
	}
	pbiID := listResp.PBIs[0].ID

	resp, err := client.GetPBI(context.Background(), pbiID)
	if err != nil {
		t.Fatalf("GetPBI #%d: %v", pbiID, err)
	}
	if resp.PBI.ID != pbiID {
		t.Fatalf("expected PBI %d, got %d", pbiID, resp.PBI.ID)
	}
	t.Logf("PBI #%d: %s (status=%s, estimate=%s, board=%s)",
		resp.PBI.ID, resp.PBI.Name, resp.PBI.Status, resp.PBI.Estimate, nameOrEmpty(resp.PBI.Board))
}

func TestIntegrationUpdatePBI(t *testing.T) {
	client := integrationClient(t)

	// Find a PBI to update (no-op: set name to same value)
	listResp, err := client.ListPBIs(context.Background(), PBIListParams{Limit: 1, Status: "to_do"})
	if err != nil {
		t.Fatalf("ListPBIs: %v", err)
	}
	if len(listResp.PBIs) == 0 {
		t.Skip("no to_do PBIs found")
	}
	pbi := listResp.PBIs[0]
	name := pbi.Name

	err = client.UpdatePBI(context.Background(), pbi.ID, PBIInput{Name: &name})
	if err != nil {
		t.Fatalf("UpdatePBI #%d: %v", pbi.ID, err)
	}
	t.Logf("Updated PBI #%d (no-op rename)", pbi.ID)
}

func nameOrEmpty(ref *NamedRef) string {
	if ref == nil {
		return ""
	}
	return ref.Name
}
