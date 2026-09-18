package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// DefaultListAllMaxPages caps auto-follow pagination for ListAll.
const DefaultListAllMaxPages = 50

// ListPageFetcher fetches one page. query already includes page/perPage.
type ListPageFetcher func(ctx context.Context, query map[string]string) (body any, hdr http.Header, err error)

// ListBodyPageFetcher fetches one page by page/perPage numbers (POST body pagination).
type ListBodyPageFetcher func(ctx context.Context, page, perPage int) (body any, hdr http.Header, err error)

// ListAllResult holds concatenated list items and final pagination meta.
type ListAllResult struct {
	Items     []any
	Meta      map[string]any
	Pages     int
	Truncated bool // true if stopped at maxPages while has_more
}

// ExtractListItems pulls a []any from common Yunxiao list JSON shapes.
// Returns nil when no list slice is found (callers treat as empty page).
func ExtractListItems(body any) []any {
	switch v := body.(type) {
	case []any:
		return v
	case nil:
		return nil
	case map[string]any:
		for _, key := range []string{"items", "list", "data", "pipelines", "changeRequests", "members", "repositories"} {
			if inner, ok := v[key]; ok {
				if s := ExtractListItems(inner); s != nil {
					return s
				}
			}
		}
	}
	return nil
}

// DedupItemsByID drops duplicate map items that share a non-empty "id" string.
// Items without id are kept in order. Stable: first occurrence wins.
func DedupItemsByID(items []any) []any {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]any, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			out = append(out, it)
			continue
		}
		id, _ := m["id"].(string)
		if id == "" {
			out = append(out, it)
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, it)
	}
	return out
}

// ListAll loops pages using PaginationFromHeader / has_more until done or maxPages.
// startPage defaults to 1; perPage defaults to 20; maxPages defaults to DefaultListAllMaxPages.
// baseQuery is copied each page; page/perPage keys are overwritten.
// When pagination headers are absent, a single page is treated as complete
// (has_more is not set by MetaWithPagination — agents must not assume completeness
// from a plain first page without --all; ListAll marks list_all=true).
func ListAll(ctx context.Context, startPage, perPage, maxPages int, baseQuery map[string]string, fetch ListPageFetcher) (*ListAllResult, error) {
	if fetch == nil {
		return nil, fmt.Errorf("ListAll: nil fetch")
	}
	return ListAllPages(ctx, startPage, perPage, maxPages, func(ctx context.Context, page, pp int) (any, http.Header, error) {
		q := map[string]string{}
		for k, v := range baseQuery {
			if k == "page" || k == "perPage" {
				continue
			}
			q[k] = v
		}
		q["page"] = strconv.Itoa(page)
		q["perPage"] = strconv.Itoa(pp)
		return fetch(ctx, q)
	})
}

// ListAllPages is like ListAll but takes an explicit page/perPage fetcher (POST bodies).
func ListAllPages(ctx context.Context, startPage, perPage, maxPages int, fetch ListBodyPageFetcher) (*ListAllResult, error) {
	if fetch == nil {
		return nil, fmt.Errorf("ListAllPages: nil fetch")
	}
	if startPage <= 0 {
		startPage = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if maxPages <= 0 {
		maxPages = DefaultListAllMaxPages
	}

	var all []any
	page := startPage
	var lastMeta map[string]any

	for pages := 1; pages <= maxPages; pages++ {
		body, hdr, err := fetch(ctx, page, perPage)
		if err != nil {
			return nil, err
		}
		items := ExtractListItems(body)
		if len(items) > 0 {
			all = append(all, items...)
		}

		meta := MetaWithPagination(map[string]any{}, hdr)
		meta["list_all"] = true
		meta["pages_fetched"] = pages
		lastMeta = meta

		// Empty page → done (API may still emit a spurious x-next-page).
		if len(items) == 0 {
			meta["has_more"] = false
			return &ListAllResult{Items: all, Meta: meta, Pages: pages}, nil
		}
		// Collected at least x-total → done even if headers keep advertising next.
		if total, ok := meta["total"].(int); ok && total > 0 && len(all) >= total {
			meta["has_more"] = false
			return &ListAllResult{Items: all, Meta: meta, Pages: pages}, nil
		}

		hasMore, hasFlag := meta["has_more"].(bool)
		if !hasFlag {
			// No pagination headers → treat as single complete page under --all.
			meta["has_more"] = false
			return &ListAllResult{Items: all, Meta: meta, Pages: pages}, nil
		}
		if !hasMore {
			return &ListAllResult{Items: all, Meta: meta, Pages: pages}, nil
		}

		if pages == maxPages {
			meta["truncated"] = true
			return &ListAllResult{Items: all, Meta: meta, Pages: pages, Truncated: true}, nil
		}

		if p, ok := meta["pagination"].(*Pagination); ok && p != nil && p.NextPage > 0 {
			page = p.NextPage
		} else {
			page++
		}
	}

	// unreachable, but keep compiler happy
	return &ListAllResult{Items: all, Meta: lastMeta, Pages: maxPages, Truncated: true}, nil
}
