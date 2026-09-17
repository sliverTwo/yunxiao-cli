package client

import (
	"context"
	"fmt"
	"net/http"
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
		h.Set("x-next-page", "999") // always more
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
