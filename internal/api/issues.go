package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type IssueListParams struct {
	Limit      int
	Offset     int
	Sort       string
	Query      string
	Include    []string
	AssigneeID int
	DueDate    string
	StatusID   int
	PriorityID int
	Subject    string
	TaskTypeID int
	ProjectID  int
}

func (c *Client) ListIssues(ctx context.Context, params IssueListParams) (IssueListResponse, error) {
	query := url.Values{}
	hasFilter := false
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	if strings.TrimSpace(params.Sort) != "" {
		query.Set("sort", params.Sort)
	}
	if strings.TrimSpace(params.Query) != "" {
		query.Set("set_filter", "1")
		hasFilter = true
		query.Set("easy_query_q", params.Query)
	}
	if params.AssigneeID > 0 {
		query.Set("assigned_to_id", strconv.Itoa(params.AssigneeID))
		hasFilter = true
	}
	if strings.TrimSpace(params.DueDate) != "" {
		query.Set("due_date", params.DueDate)
		hasFilter = true
	}
	if params.StatusID > 0 {
		query.Set("status_id", strconv.Itoa(params.StatusID))
		hasFilter = true
	}
	if params.PriorityID > 0 {
		query.Set("priority_id", strconv.Itoa(params.PriorityID))
		hasFilter = true
	}
	if strings.TrimSpace(params.Subject) != "" {
		query.Set("subject", params.Subject)
		hasFilter = true
	}
	if params.TaskTypeID > 0 {
		query.Set("tracker_id", strconv.Itoa(params.TaskTypeID))
		hasFilter = true
	}
	if params.ProjectID > 0 {
		query.Set("project_id", strconv.Itoa(params.ProjectID))
		hasFilter = true
	}
	if hasFilter {
		query.Set("set_filter", "1")
	}
	if len(params.Include) > 0 {
		query.Set("include", strings.Join(params.Include, ","))
	}

	var resp IssueListResponse
	if err := c.doJSON(ctx, "GET", "/issues.json", query, nil, &resp); err != nil {
		return IssueListResponse{}, err
	}
	return resp, nil
}

func (c *Client) CreateIssue(ctx context.Context, input IssueInput) (IssueResponse, error) {
	var resp IssueResponse
	request := IssueRequest{Issue: input}
	if err := c.doJSON(ctx, "POST", "/issues.json", nil, request, &resp); err != nil {
		return IssueResponse{}, err
	}
	return resp, nil
}

func (c *Client) UpdateIssue(ctx context.Context, id int, input IssueInput) (IssueResponse, error) {
	if id == 0 {
		return IssueResponse{}, fmt.Errorf("missing issue id")
	}
	path := fmt.Sprintf("/issues/%d.json", id)
	var resp IssueResponse
	request := IssueRequest{Issue: input}
	if err := c.doJSON(ctx, "PUT", path, nil, request, &resp); err != nil {
		return IssueResponse{}, err
	}
	return resp, nil
}
