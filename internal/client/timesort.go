package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// listTimeKeys are tried in order when ranking list items for "what's recent".
// Prefer update/modified fields over create fields when both exist.
var listTimeKeys = []string{
	"updatedAt", "gmtModified", "modifiedTime", "gmtUpdate", "updateTime", "modifiedAt",
	"updated_at", "modified_at",
	"createdAt", "gmtCreate", "createTime", "creationDate", "created_at",
	"committedDate", "authoredDate", "committed_date", "authored_date",
	"gmtStart", "date", "timestamp", "time",
}

// listWrapperKeys are common Yunxiao/OpenAPI list container fields.
var listWrapperKeys = []string{
	"items", "list", "data", "comments", "commentList", "activities",
	"runs", "pipelineRuns", "changeRequests", "effortRecords", "estimatedEfforts",
	"executions", "logs", "result", "records", "pipelines", "members", "repositories",
}

// SortListByTime sorts list payloads newest-first (descending) or oldest-first.
// Accepts a bare []any or common wrapper maps; unknown shapes are returned unchanged.
// Items without a recognizable time field keep relative order at the "unknown" end
// (after timed items when descending, before them when ascending).
func SortListByTime(data any, descending bool) any {
	switch v := data.(type) {
	case []any:
		sortSliceByTime(v, descending)
		return v
	case map[string]any:
		for _, key := range listWrapperKeys {
			inner, ok := v[key]
			if !ok {
				continue
			}
			switch iv := inner.(type) {
			case []any:
				sortSliceByTime(iv, descending)
				return v
			case map[string]any:
				v[key] = SortListByTime(iv, descending)
				return v
			}
		}
		return v
	default:
		return data
	}
}

func sortSliceByTime(items []any, descending bool) {
	if len(items) < 2 {
		return
	}
	type ranked struct {
		idx int
		ms  int64
		ok  bool
	}
	ranks := make([]ranked, len(items))
	for i, it := range items {
		ms, ok := itemTimeMillis(it)
		ranks[i] = ranked{idx: i, ms: ms, ok: ok}
	}
	sort.SliceStable(ranks, func(i, j int) bool {
		a, b := ranks[i], ranks[j]
		if a.ok != b.ok {
			if descending {
				return a.ok && !b.ok
			}
			return !a.ok && b.ok
		}
		if !a.ok {
			return false
		}
		if descending {
			return a.ms > b.ms
		}
		return a.ms < b.ms
	})
	out := make([]any, len(items))
	for i, r := range ranks {
		out[i] = items[r.idx]
	}
	copy(items, out)
}

func itemTimeMillis(item any) (int64, bool) {
	m, ok := item.(map[string]any)
	if !ok || m == nil {
		return 0, false
	}
	for _, k := range listTimeKeys {
		if v, exists := m[k]; exists {
			if ms, ok := toMillis(v); ok {
				return ms, true
			}
		}
	}
	return 0, false
}

func toMillis(v any) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return normalizeEpoch(int64(t)), true
	case float32:
		return normalizeEpoch(int64(t)), true
	case int:
		return normalizeEpoch(int64(t)), true
	case int64:
		return normalizeEpoch(t), true
	case int32:
		return normalizeEpoch(int64(t)), true
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			f, err2 := t.Float64()
			if err2 != nil {
				return 0, false
			}
			return normalizeEpoch(int64(f)), true
		}
		return normalizeEpoch(i), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" || s == "<nil>" {
			return 0, false
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return normalizeEpoch(n), true
		}
		if n, err := strconv.ParseFloat(s, 64); err == nil {
			return normalizeEpoch(int64(n)), true
		}
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if tm, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return tm.UnixMilli(), true
			}
		}
		return 0, false
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "" || s == "<nil>" {
			return 0, false
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return normalizeEpoch(n), true
		}
		return 0, false
	}
}

// normalizeEpoch treats values < 1e12 as seconds, otherwise milliseconds.
func normalizeEpoch(n int64) int64 {
	if n <= 0 {
		return n
	}
	if n < 1_000_000_000_000 {
		return n * 1000
	}
	return n
}

// SortDescending reports whether a --sort flag value means newest-first.
// Empty / unknown values default to descending.
func SortDescending(sortFlag string) bool {
	switch strings.ToLower(strings.TrimSpace(sortFlag)) {
	case "asc", "ascending", "oldest", "oldest-first":
		return false
	default:
		return true
	}
}
