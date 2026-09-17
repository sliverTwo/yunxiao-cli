package cmd

import (
	"strings"
	"testing"
)

func TestAfterSortByTimeDefaultDesc(t *testing.T) {
	in := []any{
		map[string]any{"id": "old", "gmtCreate": "2020-01-01T00:00:00Z"},
		map[string]any{"id": "new", "gmtCreate": "2024-01-01T00:00:00Z"},
	}
	after, err := afterSortByTime("desc", nil)
	if err != nil {
		t.Fatal(err)
	}
	out, meta := after(in, map[string]any{"risk": "read"})
	list := out.([]any)
	if list[0].(map[string]any)["id"] != "new" {
		t.Fatalf("%#v", list)
	}
	if meta["sort"] != "desc" {
		t.Fatalf("meta=%v", meta)
	}
}

func TestAfterSortByTimeAsc(t *testing.T) {
	in := []any{
		map[string]any{"id": "new", "gmtCreate": "2024-01-01T00:00:00Z"},
		map[string]any{"id": "old", "gmtCreate": "2020-01-01T00:00:00Z"},
	}
	after, err := afterSortByTime("asc", nil)
	if err != nil {
		t.Fatal(err)
	}
	out, meta := after(in, nil)
	list := out.([]any)
	if list[0].(map[string]any)["id"] != "old" {
		t.Fatalf("%#v", list)
	}
	if meta["sort"] != "asc" {
		t.Fatalf("meta=%v", meta)
	}
}

func TestAfterSortByTimeInvalid(t *testing.T) {
	_, err := afterSortByTime("dessc", nil)
	if err == nil || !strings.Contains(err.Error(), "asc or desc") {
		t.Fatalf("want invalid --sort error, got %v", err)
	}
}

func TestAfterSortByCreateTimePrefersCreate(t *testing.T) {
	in := []any{
		map[string]any{"id": "older-edited", "gmtCreate": "2024-01-01T00:00:00Z", "gmtModified": "2025-12-01T00:00:00Z"},
		map[string]any{"id": "newer-posted", "gmtCreate": "2025-06-01T00:00:00Z", "gmtModified": "2025-06-01T00:00:00Z"},
	}
	after, err := afterSortByCreateTime("desc", nil)
	if err != nil {
		t.Fatal(err)
	}
	out, _ := after(in, nil)
	list := out.([]any)
	if list[0].(map[string]any)["id"] != "newer-posted" {
		t.Fatalf("create-first: %#v", list)
	}
}
