package workflow

import "sort"

// BFSPath returns the list of status ids to visit (excluding start, including target)
// along a shortest path in edges. Nil if unreachable.
func BFSPath(start, target string, edges map[string][]string) []string {
	if start == "" || target == "" {
		return nil
	}
	if start == target {
		return nil
	}
	type node struct {
		id   string
		path []string
	}
	queue := []node{{id: start, path: nil}}
	seen := map[string]bool{start: true}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, next := range edges[n.id] {
			if seen[next] {
				continue
			}
			path := append(append([]string{}, n.path...), next)
			if next == target {
				return path
			}
			seen[next] = true
			queue = append(queue, node{id: next, path: path})
		}
	}
	return nil
}

// AddEdge appends to under from if not already present.
func AddEdge(edges map[string][]string, from, to string) {
	if from == "" || to == "" || from == to {
		return
	}
	for _, x := range edges[from] {
		if x == to {
			return
		}
	}
	edges[from] = append(edges[from], to)
}

// SortedCopy returns a deep copy of edges with sorted adjacency lists and keys order-stable for JSON.
func SortedCopy(edges map[string][]string) map[string][]string {
	out := map[string][]string{}
	keys := make([]string, 0, len(edges))
	for k := range edges {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tos := append([]string{}, edges[k]...)
		sort.Strings(tos)
		out[k] = tos
	}
	return out
}

// CountEdges returns total directed edges.
func CountEdges(edges map[string][]string) int {
	n := 0
	for _, tos := range edges {
		n += len(tos)
	}
	return n
}
