package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListPBIsBuildsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("set_filter") != "1" {
			t.Fatalf("set_filter = %s", q.Get("set_filter"))
		}
		if q.Get("limit") != "5" {
			t.Fatalf("limit = %s", q.Get("limit"))
		}
		if q.Get("offset") != "10" {
			t.Fatalf("offset = %s", q.Get("offset"))
		}
		if q.Get("sort") != "updated_at:desc" {
			t.Fatalf("sort = %s", q.Get("sort"))
		}
		if q.Get("easy_query_q") != "design" {
			t.Fatalf("easy_query_q = %s", q.Get("easy_query_q"))
		}
		if q.Get("status") != "to_do" {
			t.Fatalf("status = %s", q.Get("status"))
		}
		if q.Get("author_id") != "51" {
			t.Fatalf("author_id = %s", q.Get("author_id"))
		}
		if q.Get("easy_product_backlog_board_id") != "17" {
			t.Fatalf("board_id = %s", q.Get("easy_product_backlog_board_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"easy_product_backlog_items":[{"id":1,"name":"PBI 1","status":"to_do"}],"total_count":1,"offset":0,"limit":5}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	resp, err := client.ListPBIs(context.Background(), PBIListParams{
		Limit:    5,
		Offset:   10,
		Sort:     "updated_at:desc",
		Query:    "design",
		Status:   "to_do",
		AuthorID: 51,
		BoardID:  17,
	})
	if err != nil {
		t.Fatalf("ListPBIs error: %v", err)
	}
	if len(resp.PBIs) != 1 || resp.PBIs[0].ID != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetPBI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/easy_product_backlog_items/42.json" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"easy_product_backlog_item":{"id":42,"name":"Test PBI","description":"desc","estimate":"3","estimate_float":3.0,"status":"to_do","author":{"id":1,"name":"Alice"},"easy_product_backlog_board":{"id":17,"name":"Board"},"issues":[{"id":100,"subject":"Task"}],"easy_sticky_notes":[{"id":1,"name":"Note","status":"done"}],"created_at":"2024-01-01","updated_at":"2024-01-15"}}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	resp, err := client.GetPBI(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetPBI error: %v", err)
	}
	if resp.PBI.ID != 42 {
		t.Fatalf("id = %d", resp.PBI.ID)
	}
	if resp.PBI.Name != "Test PBI" {
		t.Fatalf("name = %q", resp.PBI.Name)
	}
	if resp.PBI.EstimateFloat != 3.0 {
		t.Fatalf("estimate_float = %f", resp.PBI.EstimateFloat)
	}
	if resp.PBI.Board == nil || resp.PBI.Board.ID != 17 {
		t.Fatalf("board = %+v", resp.PBI.Board)
	}
	if len(resp.PBI.Issues) != 1 || resp.PBI.Issues[0].ID != 100 {
		t.Fatalf("issues = %+v", resp.PBI.Issues)
	}
	if len(resp.PBI.StickyNotes) != 1 || resp.PBI.StickyNotes[0].Name != "Note" {
		t.Fatalf("sticky_notes = %+v", resp.PBI.StickyNotes)
	}
}

func TestGetPBIMissingID(t *testing.T) {
	client := &Client{BaseURL: "https://example.com", APIKey: "key", HTTP: http.DefaultClient}
	_, err := client.GetPBI(context.Background(), 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUpdatePBI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/easy_product_backlog_items/42.json" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"name":"Updated"`) {
			t.Fatalf("body = %s", string(body))
		}
		if !strings.Contains(string(body), `"status":"done"`) {
			t.Fatalf("body missing status: %s", string(body))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTP: server.Client()}
	name := "Updated"
	status := "done"
	err := client.UpdatePBI(context.Background(), 42, PBIInput{Name: &name, Status: &status})
	if err != nil {
		t.Fatalf("UpdatePBI error: %v", err)
	}
}

func TestUpdatePBIMissingID(t *testing.T) {
	client := &Client{BaseURL: "https://example.com", APIKey: "key", HTTP: http.DefaultClient}
	err := client.UpdatePBI(context.Background(), 0, PBIInput{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
