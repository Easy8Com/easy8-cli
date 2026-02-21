package api

import (
	"context"
	"net/url"
	"strconv"
)

type SearchParams struct {
	Query      string
	OpenIssues bool
	Scope      int
	IssuesOnly bool
}

func (c *Client) Search(ctx context.Context, params SearchParams) (SearchResponse, error) {
	query := url.Values{}
	query.Set("q", params.Query)
	if params.OpenIssues {
		query.Set("open_issues", "1")
	}
	if params.Scope > 0 {
		query.Set("scope", strconv.Itoa(params.Scope))
	}
	if params.IssuesOnly {
		query.Set("issues", "1")
	}

	var resp SearchResponse
	if err := c.doJSON(ctx, "GET", "/search.json", query, nil, &resp); err != nil {
		return SearchResponse{}, err
	}
	return resp, nil
}
