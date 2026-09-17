package zhiyi

import (
	"fmt"
	"strings"
	"sync"
)

// RelationFetcher loads a work item by id for relation enrichment (read-only).
type RelationFetcher func(resourceID string) (map[string]any, error)

// RelationResourceID extracts resourceId from a relation record.
func RelationResourceID(rec map[string]any) string {
	if rec == nil {
		return ""
	}
	for _, key := range []string{"resourceId", "resource_id", "workitemId", "workItemId"} {
		if v, ok := rec[key]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

// Subject extracts work item subject/title when present.
func Subject(item map[string]any) string {
	if item == nil {
		return ""
	}
	nested, _ := item["data"].(map[string]any)
	for _, m := range []map[string]any{item, nested} {
		if m == nil {
			continue
		}
		for _, key := range []string{"subject", "title", "name"} {
			if v, ok := m[key]; ok && v != nil {
				s := strings.TrimSpace(fmt.Sprint(v))
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

// CategoryID extracts categoryId (Req/Bug/Task/…).
func CategoryID(item map[string]any) string {
	if item == nil {
		return ""
	}
	nested, _ := item["data"].(map[string]any)
	for _, m := range []map[string]any{item, nested} {
		if m == nil {
			continue
		}
		if v, ok := m["categoryId"]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

// MergeRelationEnrichment attaches serial_number/subject/category/url onto a
// relation record, or resolve_error when fetch failed. Original keys are kept.
func MergeRelationEnrichment(rec map[string]any, item map[string]any, profileSpace string, resolveErr error) map[string]any {
	if rec == nil {
		rec = map[string]any{}
	}
	out := cloneStringMap(rec)
	if resolveErr != nil {
		out["resolve_error"] = resolveErr.Error()
		return out
	}
	if item == nil {
		out["resolve_error"] = "empty workitem"
		return out
	}
	if sn := SerialNumber(item); sn != "" {
		out["serial_number"] = sn
	}
	if subj := Subject(item); subj != "" {
		out["subject"] = subj
	}
	cat := CategoryID(item)
	if cat == "" {
		if rt, ok := out["resourceType"]; ok && rt != nil {
			cat = strings.TrimSpace(fmt.Sprint(rt))
		}
	}
	if cat != "" {
		out["category"] = cat
		// Keep resourceType if already set; otherwise mirror category.
		if _, ok := out["resourceType"]; !ok {
			out["resourceType"] = cat
		}
	}
	sid := ResolveSpaceID(item, profileSpace, "")
	if u := WorkItemURL(item, sid); u != "" {
		out["url"] = u
	}
	return out
}

func cloneStringMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m)+8)
	for k, v := range m {
		out[k] = v
	}
	return out
}

// ExtractRelationRecords returns the mutable list of relation maps from a list
// payload (bare []any or common wrappers). Non-map entries are skipped.
func ExtractRelationRecords(data any) []map[string]any {
	switch v := data.(type) {
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, it := range v {
			if m, ok := it.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return v
	case map[string]any:
		for _, key := range []string{"items", "list", "data", "relationRecords", "records"} {
			if inner, ok := v[key]; ok {
				return ExtractRelationRecords(inner)
			}
		}
	}
	return nil
}

// EnrichRelationRecords resolves each relation's resourceId via fetch (bounded
// concurrency) and merges serial_number/subject/category/url. Best-effort:
// failures keep the original record plus resolve_error. Mutates list payloads
// in place when possible; returns the (possibly same) data root.
func EnrichRelationRecords(data any, fetch RelationFetcher, profileSpace string, concurrency int) any {
	if fetch == nil {
		return data
	}
	if concurrency <= 0 {
		concurrency = 6
	}
	if concurrency > 8 {
		concurrency = 8
	}

	switch root := data.(type) {
	case []any:
		enrichRelationSlice(root, fetch, profileSpace, concurrency)
		return root
	case map[string]any:
		for _, key := range []string{"items", "list", "data", "relationRecords", "records"} {
			if inner, ok := root[key]; ok {
				root[key] = EnrichRelationRecords(inner, fetch, profileSpace, concurrency)
			}
		}
		return root
	default:
		return data
	}
}

func enrichRelationSlice(list []any, fetch RelationFetcher, profileSpace string, concurrency int) {
	type job struct {
		idx int
		id  string
	}
	jobs := make([]job, 0, len(list))
	for i, it := range list {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		id := RelationResourceID(m)
		if id == "" {
			continue
		}
		jobs = append(jobs, job{idx: i, id: id})
	}
	if len(jobs) == 0 {
		return
	}

	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			item, err := fetch(j.id)
			mu.Lock()
			defer mu.Unlock()
			orig, _ := list[j.idx].(map[string]any)
			list[j.idx] = MergeRelationEnrichment(orig, item, profileSpace, err)
		}()
	}
	wg.Wait()
}
