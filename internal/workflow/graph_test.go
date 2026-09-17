package workflow

import (
	"reflect"
	"testing"
)

func TestBFSPath(t *testing.T) {
	edges := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"d"},
	}
	path := BFSPath("a", "d", edges)
	if !reflect.DeepEqual(path, []string{"b", "c", "d"}) {
		t.Fatalf("%v", path)
	}
	if BFSPath("a", "a", edges) != nil {
		t.Fatal("same")
	}
	if BFSPath("d", "a", edges) != nil {
		t.Fatal("unreachable")
	}
}

func TestAddEdgeIdempotent(t *testing.T) {
	e := map[string][]string{}
	AddEdge(e, "1", "2")
	AddEdge(e, "1", "2")
	if len(e["1"]) != 1 {
		t.Fatal(e)
	}
}
