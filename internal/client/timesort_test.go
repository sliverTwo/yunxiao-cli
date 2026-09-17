package client

import (
	"testing"
)

func TestSortListByTimeDescBareSlice(t *testing.T) {
	in := []any{
		map[string]any{"id": "old", "gmtCreate": "2024-01-01T00:00:00Z"},
		map[string]any{"id": "new", "gmtCreate": "2025-06-01T00:00:00Z"},
		map[string]any{"id": "mid", "gmtCreate": "2024-06-01T00:00:00Z"},
	}
	out := SortListByTime(in, true).([]any)
	if out[0].(map[string]any)["id"] != "new" || out[2].(map[string]any)["id"] != "old" {
		t.Fatalf("got %#v", out)
	}
}

func TestSortListByTimeAsc(t *testing.T) {
	in := []any{
		map[string]any{"id": "b", "gmtCreate": float64(2000)},
		map[string]any{"id": "a", "gmtCreate": float64(1000)},
	}
	out := SortListByTime(in, false).([]any)
	if out[0].(map[string]any)["id"] != "a" {
		t.Fatalf("got %#v", out)
	}
}

func TestSortListByTimePrefersUpdatedOverCreated(t *testing.T) {
	in := []any{
		map[string]any{"id": "stale", "gmtCreate": "2025-01-01T00:00:00Z", "gmtModified": "2025-01-02T00:00:00Z"},
		map[string]any{"id": "fresh", "gmtCreate": "2024-01-01T00:00:00Z", "gmtModified": "2025-06-01T00:00:00Z"},
	}
	out := SortListByTime(in, true).([]any)
	if out[0].(map[string]any)["id"] != "fresh" {
		t.Fatalf("prefer modified: %#v", out)
	}
}

func TestSortListByTimeWrapperComments(t *testing.T) {
	in := map[string]any{
		"comments": []any{
			map[string]any{"id": "1", "createTime": float64(1_700_000_000_000)},
			map[string]any{"id": "2", "createTime": float64(1_800_000_000_000)},
		},
	}
	out := SortListByTime(in, true).(map[string]any)
	list := out["comments"].([]any)
	if list[0].(map[string]any)["id"] != "2" {
		t.Fatalf("%#v", list)
	}
}

func TestSortListByTimeEpochSeconds(t *testing.T) {
	in := []any{
		map[string]any{"id": "a", "createdAt": float64(1_700_000_000)},
		map[string]any{"id": "b", "createdAt": float64(1_800_000_000)},
	}
	out := SortListByTime(in, true).([]any)
	if out[0].(map[string]any)["id"] != "b" {
		t.Fatalf("%#v", out)
	}
}

func TestSortDescending(t *testing.T) {
	if !SortDescending("") || !SortDescending("desc") || !SortDescending("DESC") {
		t.Fatal("default/desc")
	}
	if SortDescending("asc") || SortDescending("oldest") {
		t.Fatal("asc")
	}
}

func TestSortListByTimeNoopScalar(t *testing.T) {
	if SortListByTime("x", true) != "x" {
		t.Fatal()
	}
}
