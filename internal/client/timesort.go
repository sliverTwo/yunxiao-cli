package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TimeKeyPreference selects which timestamp family ranks list items.
type TimeKeyPreference int

const (
	// PreferUpdateTime ranks by update/modified fields first, then create
	// (activity, MR lists, pipeline runs, efforts — "what's recently touched").
	PreferUpdateTime TimeKeyPreference = iota
	// PreferCreateTime ranks by create fields first, then update
	// (comment lists — "newest comment" means most recently posted).
	PreferCreateTime
)

// listTimeKeysUpdateFirst: update/modified before create (default for activity-style lists).
var listTimeKeysUpdateFirst = []string{
	"updatedAt", "gmtModified", "modifiedTime", "gmtUpdate", "updateTime", "modifiedAt",
	"updated_at", "modified_at",
	"createdAt", "gmtCreate", "createTime", "creationDate", "created_at",
	"committedDate", "authoredDate", "committed_date", "authored_date",
	"gmtStart", "date", "timestamp", "time",
}

// listTimeKeysCreateFirst: create before update (comment lists).
var listTimeKeysCreateFirst = []string{
	"createdAt", "gmtCreate", "createTime", "creationDate", "created_at",
	"updatedAt", "gmtModified", "modifiedTime", "gmtUpdate", "updateTime", "modifiedAt",
	"updated_at", "modified_at",
	"committedDate", "authoredDate", "committed_date", "authored_date",
	"gmtStart", "date", "timestamp", "time",
}

// listWrapperKeys are common Yunxiao/OpenAPI list container fields.
var listWrapperKeys = []string{
	"items", "list", "data", "comments", "commentList", "activities",
	"runs", "pipelineRuns", "changeRequests", "effortRecords", "estimatedEfforts",
	"executions", "logs", "result", "records", "pipelines", "members", "repositories",
}

// SortListByTime sorts list payloads newest-first (descending) or oldest-first,
// preferring update/modified timestamps when both exist.
// Accepts a bare []any or common wrapper maps; unknown shapes are returned unchanged.
// Items without a recognizable time field keep relative order at the "unknown" end
// (after timed items when descending, before them when ascending).
func SortListByTime(data any, descending bool) any {
	return SortListByTimePref(data, descending, PreferUpdateTime)
}

// SortListByTimePref is like SortListByTime but selects create- vs update-first keys.
func SortListByTimePref(data any, descending bool, pref TimeKeyPreference) any {
	keys := listTimeKeysUpdateFirst
	if pref == PreferCreateTime {
		keys = listTimeKeysCreateFirst
	}
	switch v := data.(type) {
	case []any:
		sortSliceByTime(v, descending, keys)
		return v
	case map[string]any:
		for _, key := range listWrapperKeys {
			inner, ok := v[key]
			if !ok {
				continue
			}
			switch iv := inner.(type) {
			case []any:
				sortSliceByTime(iv, descending, keys)
				return v
			case map[string]any:
				v[key] = SortListByTimePref(iv, descending, pref)
				return v
			}
		}
		return v
	default:
		return data
	}
}

func sortSliceByTime(items []any, descending bool, keys []string) {
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
		ms, ok := itemTimeMillis(it, keys)
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

func itemTimeMillis(item any, keys []string) (int64, bool) {
	m, ok := item.(map[string]any)
	if !ok || m == nil {
		return 0, false
	}
	for _, k := range keys {
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

// ParseSortDescending reports whether a --sort flag value means newest-first.
// Empty defaults to descending. Allowed: asc|desc and documented aliases.
// Unknown values return an error (never silently treated as desc).
func ParseSortDescending(sortFlag string) (descending bool, err error) {
	switch strings.ToLower(strings.TrimSpace(sortFlag)) {
	case "", "desc", "descending", "newest", "newest-first":
		return true, nil
	case "asc", "ascending", "oldest", "oldest-first":
		return false, nil
	default:
		return false, fmt.Errorf("--sort must be asc or desc (got %q)", sortFlag)
	}
}
