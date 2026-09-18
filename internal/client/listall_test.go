package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func TestExtractListItems(t *testing.T) {
	if got := ExtractListItems([]any{"a", "b"}); len(got) != 2 {
		t.Fatalf("slice: %v", got)
	}
	wrapped := map[string]any{"data": []any{map[string]any{"id": "1"}}}
	if got := ExtractListItems(wrapped); len(got) != 1 {
		t.Fatalf("wrapped data: %v", got)
	}
	if ExtractListItems(map[string]any{"foo": "bar"}) != nil {
		t.Fatal("no list key should return nil")
	}
	if ExtractListItems(nil) != nil {
		t.Fatal("nil")
	}
}

func TestListAllFollowsPages(t *testing.T) {
	calls := 0
	fetch := func(ctx context.Context, query map[string]string) (any, http.Header, error) {
		calls++
		page := query["page"]
		h := http.Header{}
		h.Set("x-per-page", "2")
		h.Set("x-total", "5")
		switch page {
		case "1":
			h.Set("x-page", "1")
			h.Set("x-next-page", "2")
			return []any{"a", "b"}, h, nil
		case "2":
			h.Set("x-page", "2")
			h.Set("x-next-page", "3")
			return []any{"c", "d"}, h, nil
		case "3":
			h.Set("x-page", "3")
			// last page: no next
			return []any{"e"}, h, nil
		default:
			return nil, nil, fmt.Errorf("unexpected page %s", page)
		}
	}
	res, err := ListAll(context.Background(), 1, 2, 50, nil, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls=%d", calls)
	}
	if len(res.Items) != 5 {
		t.Fatalf("items=%v", res.Items)
	}
	if res.Truncated || res.Meta["has_more"] != false {
		t.Fatalf("meta=%v truncated=%v", res.Meta, res.Truncated)
	}
	if res.Meta["list_all"] != true || res.Pages != 3 {
		t.Fatalf("pages/meta=%d %v", res.Pages, res.Meta)
	}
	if res.Meta["total"] != 5 {
		t.Fatalf("total=%v", res.Meta["total"])
	}
}

func TestListAllRespectsMaxPages(t *testing.T) {
	fetch := func(ctx context.Context, query map[string]string) (any, http.Header, error) {
		h := http.Header{}
		h.Set("x-page", query["page"])
		h.Set("x-per-page", "1")
		h.Set("x-total", "100")
		h.Set("x-total-pages", "100")
		page, _ := strconv.Atoi(query["page"])
		h.Set("x-next-page", strconv.Itoa(page+1)) // sequential; totals keep has_more true
		return []any{query["page"]}, h, nil
	}
	res, err := ListAll(context.Background(), 1, 1, 3, map[string]string{"state": "opened"}, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Truncated || res.Meta["has_more"] != true || res.Meta["truncated"] != true {
		t.Fatalf("want truncated has_more: %v", res.Meta)
	}
	if len(res.Items) != 3 || res.Pages != 3 {
		t.Fatalf("items=%v pages=%d", res.Items, res.Pages)
	}
}

func TestListAllNoPaginationHeaders(t *testing.T) {
	fetch := func(ctx context.Context, query map[string]string) (any, http.Header, error) {
		return []any{"only"}, http.Header{}, nil
	}
	res, err := ListAll(context.Background(), 0, 0, 0, nil, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 || res.Pages != 1 || res.Meta["has_more"] != false {
		t.Fatalf("%+v meta=%v", res, res.Meta)
	}
}

func TestListAllNilFetch(t *testing.T) {
	_, err := ListAll(context.Background(), 1, 10, 5, nil, nil)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestListAllPropagatesFetchError(t *testing.T) {
	fetch := func(ctx context.Context, query map[string]string) (any, http.Header, error) {
		return nil, nil, fmt.Errorf("boom")
	}
	_, err := ListAll(context.Background(), 1, 10, 5, nil, fetch)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err=%v", err)
	}
}

func TestDedupItemsByID(t *testing.T) {
	items := []any{
		map[string]any{"id": "a", "n": 1},
		map[string]any{"id": "b", "n": 2},
		map[string]any{"id": "a", "n": 3},
		map[string]any{"n": 4}, // no id kept
		"plain",
	}
	out := DedupItemsByID(items)
	if len(out) != 4 {
		t.Fatalf("len=%d out=%v", len(out), out)
	}
	if out[0].(map[string]any)["n"] != 1 {
		t.Fatalf("first wins: %v", out[0])
	}
}

func TestListAllPagesPOSTStyle(t *testing.T) {
	calls := 0
	fetch := func(ctx context.Context, page, perPage int) (any, http.Header, error) {
		calls++
		h := http.Header{}
		h.Set("x-per-page", strconv.Itoa(perPage))
		h.Set("x-total", "3")
		h.Set("x-page", strconv.Itoa(page))
		if page == 1 {
			h.Set("x-next-page", "2")
			return []any{map[string]any{"id": "1"}, map[string]any{"id": "2"}}, h, nil
		}
		return []any{map[string]any{"id": "2"}, map[string]any{"id": "3"}}, h, nil // id 2 dup across pages
	}
	res, err := ListAllPages(context.Background(), 1, 2, 50, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(res.Items) != 4 {
		t.Fatalf("calls=%d items=%d", calls, len(res.Items))
	}
	dedup := DedupItemsByID(res.Items)
	if len(dedup) != 3 {
		t.Fatalf("dedup=%d", len(dedup))
	}
}

func TestListAllPagesStopsOnSpuriousNextWithEmptyPage(t *testing.T) {
	calls := 0
	fetch := func(ctx context.Context, page, perPage int) (any, http.Header, error) {
		calls++
		h := http.Header{}
		h.Set("x-page", strconv.Itoa(page))
		h.Set("x-per-page", strconv.Itoa(perPage))
		h.Set("x-total", "0")
		h.Set("x-total-pages", "0")
		h.Set("x-next-page", strconv.Itoa(page+1))
		return []any{}, h, nil
	}
	res, err := ListAllPages(context.Background(), 1, 5, 50, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || res.Pages != 1 || res.Meta["has_more"] != false {
		t.Fatalf("calls=%d pages=%d meta=%v", calls, res.Pages, res.Meta)
	}
}

func TestListAllPagesStopsWhenTotalReachedDespiteNext(t *testing.T) {
	calls := 0
	fetch := func(ctx context.Context, page, perPage int) (any, http.Header, error) {
		calls++
		h := http.Header{}
		h.Set("x-page", strconv.Itoa(page))
		h.Set("x-per-page", "50")
		h.Set("x-total", "83")
		h.Set("x-total-pages", "2")
		h.Set("x-next-page", strconv.Itoa(page+1))
		if page == 1 {
			items := make([]any, 50)
			for i := range items {
				items[i] = map[string]any{"id": strconv.Itoa(i)}
			}
			return items, h, nil
		}
		items := make([]any, 33)
		for i := range items {
			items[i] = map[string]any{"id": strconv.Itoa(50 + i)}
		}
		return items, h, nil
	}
	res, err := ListAllPages(context.Background(), 1, 50, 50, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(res.Items) != 83 || res.Meta["has_more"] != false || res.Truncated {
		t.Fatalf("calls=%d n=%d meta=%v trunc=%v", calls, len(res.Items), res.Meta, res.Truncated)
	}
}
