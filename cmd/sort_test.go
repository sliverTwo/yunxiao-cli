package cmd

import (
	"testing"
)

func TestAfterSortByTimeDefaultDesc(t *testing.T) {
	in := []any{
		map[string]any{"id": "old", "gmtCreate": "2020-01-01T00:00:00Z"},
		map[string]any{"id": "new", "gmtCreate": "2024-01-01T00:00:00Z"},
	}
	out, meta := afterSortByTime("desc", nil)(in, map[string]any{"risk": "read"})
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
	out, meta := afterSortByTime("asc", nil)(in, nil)
	list := out.([]any)
	if list[0].(map[string]any)["id"] != "old" {
		t.Fatalf("%#v", list)
	}
	if meta["sort"] != "asc" {
		t.Fatalf("meta=%v", meta)
	}
}
