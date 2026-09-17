package workflow

import (
	"fmt"
	"testing"
	"time"
)

func TestExploreTransitionsOneWayPartial(t *testing.T) {
	// One-way chain: after leaving a we may not rediscover all a→* edges.
	allowed := map[string]map[string]bool{
		"a": {"b": true, "c": true},
		"b": {"c": true},
		"c": {},
	}
	current := "a"
	res, err := ExploreTransitions(ProbeOptions{
		StatusIDs:     []string{"a", "b", "c"},
		StartStatus:   "a",
		DefaultStatus: "a",
		Sleep:         func(d time.Duration) {},
		Get:           func() (string, error) { return current, nil },
		Put: func(to string) error {
			from := current
			if allowed[from][to] {
				current = to
				return nil
			}
			return fmt.Errorf("当前状态:%s不能流转到目标状态:%s", from, to)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasEdge(res.Edges, "a", "b") && !hasEdge(res.Edges, "a", "c") {
		t.Fatalf("expected at least one outbound from a: %v", res.Edges)
	}
	if !hasEdge(res.Edges, "b", "c") && CountEdges(res.Edges) < 1 {
		t.Fatalf("%v", res.Edges)
	}
}

func TestExploreTransitionsWithReturnEdges(t *testing.T) {
	// Bidirectional enough to finish all pairs from a.
	allowed := map[string]map[string]bool{
		"a": {"b": true, "c": true},
		"b": {"a": true, "c": true},
		"c": {"a": true},
	}
	current := "a"
	res, err := ExploreTransitions(ProbeOptions{
		StatusIDs:     []string{"a", "b", "c"},
		StartStatus:   "a",
		DefaultStatus: "a",
		Sleep:         func(d time.Duration) {},
		Get:           func() (string, error) { return current, nil },
		Put: func(to string) error {
			from := current
			if allowed[from][to] {
				current = to
				return nil
			}
			return fmt.Errorf("当前状态:%s不能流转到目标状态:%s", from, to)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range [][2]string{{"a", "b"}, {"a", "c"}, {"b", "c"}, {"b", "a"}, {"c", "a"}} {
		if !hasEdge(res.Edges, e[0], e[1]) {
			t.Fatalf("missing %s→%s in %v", e[0], e[1], res.Edges)
		}
	}
}

func TestExploreNeedsFieldsCountsAsEdge(t *testing.T) {
	current := "a"
	res, err := ExploreTransitions(ProbeOptions{
		StatusIDs:   []string{"a", "b"},
		StartStatus: "a",
		Sleep:       func(d time.Duration) {},
		Get:         func() (string, error) { return current, nil },
		Put: func(to string) error {
			if current == "a" && to == "b" {
				return fmt.Errorf("字段【计划完成时间】不能为空")
			}
			return fmt.Errorf("当前状态不能流转到目标状态")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasEdge(res.Edges, "a", "b") {
		t.Fatalf("expected needs_fields edge, got %v hints=%v", res.Edges, res.RequiredHints)
	}
	if res.RequiredHints["b"] == "" && res.RequiredHints["a→b"] == "" {
		t.Fatalf("hints=%v", res.RequiredHints)
	}
}

func TestExploreTransitionsResetUnlocksSource(t *testing.T) {
	allowed := map[string]map[string]bool{
		"a": {"b": true, "c": true},
		"b": {"c": true},
		"c": {},
	}
	current := "a"
	resets := 0
	res, err := ExploreTransitions(ProbeOptions{
		StatusIDs:     []string{"a", "b", "c"},
		StartStatus:   "a",
		DefaultStatus: "a",
		Sleep:         func(d time.Duration) {},
		Get:           func() (string, error) { return current, nil },
		Put: func(to string) error {
			from := current
			if allowed[from][to] {
				current = to
				return nil
			}
			return fmt.Errorf("当前状态:%s不能流转到目标状态:%s", from, to)
		},
		Reset: func() (string, error) {
			resets++
			current = "a"
			return "a", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resets == 0 {
		t.Fatal("expected reset")
	}
	if !hasEdge(res.Edges, "a", "b") || !hasEdge(res.Edges, "a", "c") || !hasEdge(res.Edges, "b", "c") {
		t.Fatalf("edges=%v resets=%d", res.Edges, resets)
	}
}
