package cmd

import (
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/client"
)

func TestWrapDataAsItemsFromArray(t *testing.T) {
	out := []any{
		map[string]any{"id": "a"},
		map[string]any{"id": "b"},
	}
	meta := map[string]any{
		"page":    1,
		"perPage": 20,
		"total":   2,
		"has_more": false,
	}
	wrapped := wrapDataAsItems(out, meta).(map[string]any)
	items, ok := wrapped["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items=%v", wrapped["items"])
	}
	pag, ok := wrapped["pagination"].(map[string]any)
	if !ok || pag["total"] != 2 || pag["page"] != 1 {
		t.Fatalf("pagination=%v", wrapped["pagination"])
	}
}

func TestWrapDataAsItemsUsesNestedPagination(t *testing.T) {
	p := &client.Pagination{Page: 2, PerPage: 50, Total: 100}
	meta := map[string]any{"pagination": p, "page": 2}
	wrapped := wrapDataAsItems([]any{"x"}, meta).(map[string]any)
	if wrapped["pagination"] != p {
		t.Fatalf("want nested pagination pointer, got %v", wrapped["pagination"])
	}
	items := wrapped["items"].([]any)
	if len(items) != 1 || items[0] != "x" {
		t.Fatalf("items=%v", items)
	}
}

func TestWrapDataAsItemsEmpty(t *testing.T) {
	wrapped := wrapDataAsItems(nil, nil).(map[string]any)
	items := wrapped["items"].([]any)
	if len(items) != 0 {
		t.Fatalf("%v", items)
	}
	if _, ok := wrapped["pagination"]; ok {
		t.Fatal("unexpected pagination")
	}
}
