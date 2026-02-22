package api

import (
	"context"
	"net/url"
	"strconv"
)

func (c *Client) ListTrackers(ctx context.Context) ([]Tracker, error) {
	var resp TrackerListResponse
	if err := c.doJSON(ctx, "GET", "/trackers.json", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Trackers, nil
}

func (c *Client) ListIssueStatuses(ctx context.Context) ([]IssueStatus, error) {
	var resp IssueStatusListResponse
	if err := c.doJSON(ctx, "GET", "/issue_statuses.json", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.IssueStatuses, nil
}

func (c *Client) ListIssuePriorities(ctx context.Context) ([]IssuePriority, error) {
	var resp IssuePriorityListResponse
	if err := c.doJSON(ctx, "GET", "/enumerations/issue_priorities.json", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.IssuePriorities, nil
}

func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	return listUsersPaged(ctx, c)
}

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	return listProjectsPaged(ctx, c)
}

func listUsersPaged(ctx context.Context, c *Client) ([]User, error) {
	limit := 100
	offset := 0
	var all []User
	for {
		query := url.Values{}
		query.Set("limit", strconv.Itoa(limit))
		query.Set("offset", strconv.Itoa(offset))
		var resp UserListResponse
		if err := c.doJSON(ctx, "GET", "/users.json", query, nil, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Users...)
		offset += resp.Limit
		if offset >= resp.TotalCount || resp.Limit == 0 {
			break
		}
	}
	return all, nil
}

func listProjectsPaged(ctx context.Context, c *Client) ([]Project, error) {
	limit := 100
	offset := 0
	var all []Project
	for {
		query := url.Values{}
		query.Set("limit", strconv.Itoa(limit))
		query.Set("offset", strconv.Itoa(offset))
		var resp ProjectListResponse
		if err := c.doJSON(ctx, "GET", "/projects.json", query, nil, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Projects...)
		offset += resp.Limit
		if offset >= resp.TotalCount || resp.Limit == 0 {
			break
		}
	}
	return all, nil
}
