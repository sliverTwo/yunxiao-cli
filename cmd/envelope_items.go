package cmd

import "github.com/yunxiao-cli/yunxiao/internal/client"

// wrapDataAsItems builds the stable list shape:
//
//	{ "items": [...], "pagination": {...} }
//
// for opt-in --as-items / --envelope items. Pagination is taken from meta
// (nested "pagination" when present, else top-level page/perPage/total/…).
func wrapDataAsItems(out any, meta map[string]any) any {
	items := client.ExtractListItems(out)
	if items == nil {
		if s, ok := out.([]any); ok {
			items = s
		} else {
			items = []any{}
		}
	}
	data := map[string]any{"items": items}
	if meta == nil {
		return data
	}
	if p, ok := meta["pagination"]; ok && p != nil {
		data["pagination"] = p
		return data
	}
	pag := map[string]any{}
	for _, k := range []string{"page", "perPage", "total", "totalPages", "has_more"} {
		if v, ok := meta[k]; ok {
			pag[k] = v
		}
	}
	if len(pag) > 0 {
		data["pagination"] = pag
	}
	return data
}
