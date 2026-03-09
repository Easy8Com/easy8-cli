package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type PBI struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description,omitempty"`
	Estimate      string          `json:"estimate,omitempty"`
	EstimateFloat float64         `json:"estimate_float,omitempty"`
	Status        string          `json:"status,omitempty"`
	Color         string          `json:"color,omitempty"`
	Icon          string          `json:"icon,omitempty"`
	Author        *NamedRef       `json:"author,omitempty"`
	Board         *NamedRef       `json:"easy_product_backlog_board,omitempty"`
	Swimlane      *NamedRef       `json:"easy_swimlane,omitempty"`
	Issues        []PBIIssue      `json:"issues,omitempty"`
	StickyNotes   []PBIStickyNote `json:"easy_sticky_notes,omitempty"`
	CreatedAt     string          `json:"created_at,omitempty"`
	UpdatedAt     string          `json:"updated_at,omitempty"`
}

type PBIIssue struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
}

type PBIStickyNote struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PBIInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Estimate    *string `json:"estimate,omitempty"`
}

type PBIRequest struct {
	PBI PBIInput `json:"easy_product_backlog_item"`
}

type PBIResponse struct {
	PBI PBI `json:"easy_product_backlog_item"`
}

type PBIListResponse struct {
	PBIs       []PBI `json:"easy_product_backlog_items"`
	TotalCount int   `json:"total_count"`
	Offset     int   `json:"offset"`
	Limit      int   `json:"limit"`
}

type PBIListParams struct {
	Limit    int
	Offset   int
	Sort     string
	Query    string
	Status   string
	AuthorID int
	BoardID  int
}

func (c *Client) ListPBIs(ctx context.Context, params PBIListParams) (PBIListResponse, error) {
	query := url.Values{}
	query.Set("set_filter", "1")
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
		query.Set("easy_query_q", params.Query)
	}
	if strings.TrimSpace(params.Status) != "" {
		query.Set("status", params.Status)
	}
	if params.AuthorID > 0 {
		query.Set("author_id", strconv.Itoa(params.AuthorID))
	}
	if params.BoardID > 0 {
		query.Set("easy_product_backlog_board_id", strconv.Itoa(params.BoardID))
	}

	var resp PBIListResponse
	if err := c.doJSON(ctx, "GET", "/easy_product_backlog_items.json", query, nil, &resp); err != nil {
		return PBIListResponse{}, err
	}
	return resp, nil
}

func (c *Client) GetPBI(ctx context.Context, id int) (PBIResponse, error) {
	if id == 0 {
		return PBIResponse{}, fmt.Errorf("missing PBI id")
	}
	path := fmt.Sprintf("/easy_product_backlog_items/%d.json", id)
	var resp PBIResponse
	if err := c.doJSON(ctx, "GET", path, nil, nil, &resp); err != nil {
		return PBIResponse{}, err
	}
	return resp, nil
}

func (c *Client) UpdatePBI(ctx context.Context, id int, input PBIInput) error {
	if id == 0 {
		return fmt.Errorf("missing PBI id")
	}
	path := fmt.Sprintf("/easy_product_backlog_items/%d.json", id)
	request := PBIRequest{PBI: input}
	if err := c.doJSON(ctx, "PUT", path, nil, request, nil); err != nil {
		return err
	}
	return nil
}
