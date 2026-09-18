package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoMaps4xxToAPIError(t *testing.T) {
	tok := "secret-token-abc123"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != tok {
			t.Errorf("token header=%q", r.Header.Get("x-yunxiao-token"))
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errorCode":"Unauthorized","errorMsg":"bad token ` + tok + `"}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: tok, UserAgent: "t"}
	err := c.Get(context.Background(), "/oapi/v1/platform/user", nil, nil)
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("want APIError, got %T %v", err, err)
	}
	if ae.Status != 401 {
		t.Fatalf("status=%d", ae.Status)
	}
	if strings.Contains(ae.Body, tok) || strings.Contains(ae.Error(), tok) {
		t.Fatalf("token leaked: %s", ae.Error())
	}
	if !strings.Contains(ae.Body, "(redacted)") {
		t.Fatalf("expected redacted body: %s", ae.Body)
	}
}

func TestDoMaps5xxToAPIError(t *testing.T) {
	prev := sleepWithContext
	sleepWithContext = func(ctx context.Context, d time.Duration) error { return nil }
	t.Cleanup(func() { sleepWithContext = prev })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"errorMsg":"upstream"}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out map[string]any
	err := c.Get(context.Background(), "/x", nil, &out)
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 502 {
		t.Fatalf("%T %v", err, err)
	}
}

func TestGetRetries503ThenSucceeds(t *testing.T) {
	prev := sleepWithContext
	sleepWithContext = func(ctx context.Context, d time.Duration) error { return nil }
	t.Cleanup(func() { sleepWithContext = prev })

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"errorMsg":"unavailable"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "n": n})
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out map[string]any
	if err := c.Get(context.Background(), "/retry", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out["ok"] != true {
		t.Fatalf("%v", out)
	}
	if hits.Load() < 2 {
		t.Fatalf("hits=%d want >=2", hits.Load())
	}
}

func TestGet400NotRetried(t *testing.T) {
	prev := sleepWithContext
	sleepWithContext = func(ctx context.Context, d time.Duration) error { return nil }
	t.Cleanup(func() { sleepWithContext = prev })

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorMsg":"bad"}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	err := c.Get(context.Background(), "/bad", nil, nil)
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 400 {
		t.Fatalf("%T %v", err, err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits=%d want 1 (no retry on 400)", hits.Load())
	}
}

func TestGetRespectsRetryAfter(t *testing.T) {
	var slept time.Duration
	prev := sleepWithContext
	sleepWithContext = func(ctx context.Context, d time.Duration) error {
		slept = d
		return nil
	}
	t.Cleanup(func() { sleepWithContext = prev })

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"errorMsg":"slow down"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out map[string]any
	if err := c.Get(context.Background(), "/ra", nil, &out); err != nil {
		t.Fatal(err)
	}
	if slept != 2*time.Second {
		t.Fatalf("slept=%v want 2s (Retry-After)", slept)
	}
}

func TestGetCapsRetryAfterAt30s(t *testing.T) {
	var slept time.Duration
	prev := sleepWithContext
	sleepWithContext = func(ctx context.Context, d time.Duration) error {
		slept = d
		return nil
	}
	t.Cleanup(func() { sleepWithContext = prev })

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "86400")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"errorMsg":"slow down"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out map[string]any
	if err := c.Get(context.Background(), "/ra-cap", nil, &out); err != nil {
		t.Fatal(err)
	}
	if slept != 30*time.Second {
		t.Fatalf("slept=%v want 30s (Retry-After capped)", slept)
	}
}

func TestPostNotRetriedOn503(t *testing.T) {
	prev := sleepWithContext
	sleepWithContext = func(ctx context.Context, d time.Duration) error { return nil }
	t.Cleanup(func() { sleepWithContext = prev })

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"errorMsg":"unavailable"}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	err := c.Post(context.Background(), "/x", map[string]any{"a": 1}, nil)
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 503 {
		t.Fatalf("%T %v", err, err)
	}
	if hits.Load() != 1 {
		t.Fatalf("POST must not retry: hits=%d", hits.Load())
	}
}

func TestDoSuccessDecodesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "u1"})
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out map[string]any
	if err := c.Get(context.Background(), "/ok", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out["id"] != "u1" {
		t.Fatalf("%v", out)
	}
}

func TestPaginationFromHeaderNextPage(t *testing.T) {
	h := http.Header{}
	h.Set("x-page", "1")
	h.Set("x-per-page", "10")
	h.Set("x-total", "25")
	h.Set("x-next-page", "2")
	h.Set("x-prev-page", "0")
	p := PaginationFromHeader(h)
	if p == nil || p.NextPage != 2 || p.Total != 25 || p.Page != 1 {
		t.Fatalf("%+v", p)
	}
}

func TestPreviewNeverExposesRawToken(t *testing.T) {
	tok := "super-secret-pat-value"
	c := &Client{BaseURL: "https://openapi-rdc.aliyuncs.com", Token: tok, UserAgent: "t"}
	prev := c.Preview("POST", "/oapi/v1/x", map[string]string{"q": "1"}, map[string]any{"a": 1})
	raw, _ := json.Marshal(prev)
	if strings.Contains(string(raw), tok) {
		t.Fatalf("token in preview json: %s", raw)
	}
	if prev.Headers["x-yunxiao-token"] != "(redacted)" {
		t.Fatal(prev.Headers)
	}
}

func TestPostMultipartMaps4xxAndRedactsToken(t *testing.T) {
	tok := "multipart-secret-token-xyz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("content-type=%q", ct)
		}
		if r.Header.Get("x-yunxiao-token") != tok {
			t.Errorf("token header=%q", r.Header.Get("x-yunxiao-token"))
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorMsg":"upload failed token=` + tok + `"}`))
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: tok, UserAgent: "t"}
	err := c.PostMultipart(context.Background(), "/upload", nil, "file", "a.txt", []byte("hi"), map[string]string{"k": "v"}, nil)
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("want APIError, got %T %v", err, err)
	}
	if ae.Status != 400 {
		t.Fatalf("status=%d", ae.Status)
	}
	if strings.Contains(ae.Body, tok) || strings.Contains(ae.Error(), tok) {
		t.Fatalf("token leaked: %s", ae.Error())
	}
	if !strings.Contains(ae.Body, "(redacted)") {
		t.Fatalf("expected redacted body: %s", ae.Body)
	}
}

func TestPostMultipartSuccessDecodesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "att1"})
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out map[string]any
	if err := c.PostMultipart(context.Background(), "/ok", nil, "file", "a.txt", []byte("data"), nil, &out); err != nil {
		t.Fatal(err)
	}
	if out["id"] != "att1" {
		t.Fatalf("%v", out)
	}
}

func TestMetaWithPaginationHasMoreAndTotal(t *testing.T) {
	next := http.Header{}
	next.Set("x-page", "1")
	next.Set("x-per-page", "10")
	next.Set("x-total", "25")
	next.Set("x-next-page", "2")
	meta := MetaWithPagination(map[string]any{"risk": "read"}, next)
	if meta["has_more"] != true {
		t.Fatalf("has_more want true, got %v", meta["has_more"])
	}
	if meta["total"] != 25 {
		t.Fatalf("total=%v", meta["total"])
	}
	if meta["page"] != 1 {
		t.Fatalf("page=%v", meta["page"])
	}
	if meta["pagination"] == nil {
		t.Fatal("nested pagination missing")
	}

	last := http.Header{}
	last.Set("x-page", "3")
	last.Set("x-per-page", "10")
	last.Set("x-total", "25")
	last.Set("x-total-pages", "3")
	// no x-next-page → last page
	meta2 := MetaWithPagination(map[string]any{"risk": "read"}, last)
	if meta2["has_more"] != false {
		t.Fatalf("has_more want false on last page, got %v", meta2["has_more"])
	}
	if meta2["total"] != 25 || meta2["page"] != 3 {
		t.Fatalf("%v", meta2)
	}

	empty := MetaWithPagination(map[string]any{"risk": "read"}, http.Header{})
	if _, ok := empty["has_more"]; ok {
		t.Fatalf("no pagination headers should not set has_more: %v", empty)
	}

	// False-negative lock: API omits x-next-page but total implies more pages.
	noNext := http.Header{}
	noNext.Set("x-page", "1")
	noNext.Set("x-per-page", "10")
	noNext.Set("x-total", "25")
	// deliberately no x-next-page
	meta3 := MetaWithPagination(map[string]any{"risk": "read"}, noNext)
	if meta3["has_more"] != true {
		t.Fatalf("has_more want true via total/page/per_page (no x-next-page), got %v", meta3["has_more"])
	}

	// Exact last page via page*per_page == total, no next
	exact := http.Header{}
	exact.Set("x-page", "2")
	exact.Set("x-per-page", "10")
	exact.Set("x-total", "20")
	meta4 := MetaWithPagination(map[string]any{"risk": "read"}, exact)
	if meta4["has_more"] != false {
		t.Fatalf("has_more want false when page*per_page >= total, got %v", meta4["has_more"])
	}
}

func TestInferHasMoreViaTotalPages(t *testing.T) {
	h := http.Header{}
	h.Set("x-page", "1")
	h.Set("x-total", "100")
	h.Set("x-total-pages", "5")
	// no per_page, no next-page — infer from total_pages
	meta := MetaWithPagination(nil, h)
	if meta["has_more"] != true {
		t.Fatalf("want true via total_pages, got %v", meta)
	}
}

func TestDoReturnsHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-page", "2")
		w.Header().Set("x-per-page", "10")
		w.Header().Set("x-total", "25")
		w.Header().Set("x-next-page", "3")
		_ = json.NewEncoder(w).Encode([]any{map[string]any{"id": 1}})
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var out any
	hdr, err := c.Do(context.Background(), "GET", "/list", nil, nil, &out)
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Get("x-total") != "25" {
		t.Fatalf("hdr=%v", hdr)
	}
	meta := MetaWithPagination(map[string]any{"risk": "read"}, hdr)
	if meta["has_more"] != true || meta["total"] != 25 || meta["page"] != 2 {
		t.Fatalf("meta=%v", meta)
	}
}

func TestRawPaginationHeadersAndTopLevelPerPageTotalPages(t *testing.T) {
	h := http.Header{}
	h.Set("x-page", "1")
	h.Set("x-per-page", "200")
	h.Set("x-total", "450")
	h.Set("x-total-pages", "3")
	h.Set("x-next-page", "2")
	meta := MetaWithPagination(map[string]any{"risk": "read"}, h)
	if meta["perPage"] != 200 || meta["totalPages"] != 3 || meta["total"] != 450 {
		t.Fatalf("top-level=%v", meta)
	}
	raw, ok := meta["pagination_headers"].(map[string]string)
	if !ok || raw["x-total"] != "450" || raw["x-next-page"] != "2" {
		t.Fatalf("pagination_headers=%v", meta["pagination_headers"])
	}
}

func TestApplyFullPageHasMoreHeuristic(t *testing.T) {
	meta := map[string]any{"risk": "read"}
	ApplyFullPageHasMoreHeuristic(meta, 200, 200)
	if meta["has_more"] != true {
		t.Fatalf("%v", meta)
	}
	if meta["has_more_reason"] == nil {
		t.Fatal("missing reason")
	}
	// does not override existing has_more
	meta2 := map[string]any{"has_more": false}
	ApplyFullPageHasMoreHeuristic(meta2, 200, 200)
	if meta2["has_more"] != false {
		t.Fatalf("should not override: %v", meta2)
	}
}

func TestInferHasMoreIgnoresSpuriousNextPage(t *testing.T) {
	// Live Yunxiao: x-total=0 still sends x-next-page=2.
	empty := http.Header{}
	empty.Set("x-page", "1")
	empty.Set("x-per-page", "50")
	empty.Set("x-total", "0")
	empty.Set("x-total-pages", "0")
	empty.Set("x-next-page", "2")
	meta := MetaWithPagination(nil, empty)
	if meta["has_more"] != false {
		t.Fatalf("empty total should not has_more: %v", meta)
	}

	// Last page still advertises x-next-page.
	last := http.Header{}
	last.Set("x-page", "2")
	last.Set("x-per-page", "50")
	last.Set("x-total", "83")
	last.Set("x-total-pages", "2")
	last.Set("x-next-page", "3")
	meta2 := MetaWithPagination(nil, last)
	if meta2["has_more"] != false {
		t.Fatalf("page==totalPages should not has_more: %v", meta2)
	}
}
