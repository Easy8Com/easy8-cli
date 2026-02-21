package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type IssueListParams struct {
	Limit   int
	Offset  int
	Sort    string
	Query   string
	Include []string
}

func (c *Client) ListIssues(ctx context.Context, params IssueListParams) (IssueListResponse, error) {
	query := url.Values{}
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
		query.Set("easy_query_q", params.Query)
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
