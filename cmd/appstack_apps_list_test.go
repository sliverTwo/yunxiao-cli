package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/client"
)

func TestAppstackAppsListQueryIncludesPaginationKeyset(t *testing.T) {
	q := appstackAppsListQuery(20, "", "", "", "")
	if q["pagination"] != "keyset" {
		t.Fatalf("pagination=%q want keyset", q["pagination"])
	}
	if q["orderBy"] != "id" {
		t.Fatalf("orderBy=%q want id", q["orderBy"])
	}
	if q["perPage"] != "20" {
		t.Fatalf("perPage=%q", q["perPage"])
	}
	q2 := appstackAppsListQuery(0, "tok", "name", "desc", "a,b")
	if q2["pagination"] != "keyset" || q2["nextToken"] != "tok" || q2["orderBy"] != "name" || q2["sort"] != "desc" || q2["tags"] != "a,b" {
		t.Fatalf("%v", q2)
	}
	if _, ok := q2["perPage"]; ok {
		t.Fatalf("perPage should be omitted when 0: %v", q2)
	}
}

func TestAppstackAppsListQuerySentOverHTTP(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_ = json.NewEncoder(w).Encode([]any{})
	}))
	t.Cleanup(srv.Close)

	c := &client.Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	q := appstackAppsListQuery(20, "", "id", "asc", "")
	var out any
	if err := c.Get(context.Background(), "/oapi/v1/appstack/organizations/org1/apps:search", q, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "pagination=keyset") {
		t.Fatalf("missing pagination=keyset in %q", gotURL)
	}
	if !strings.Contains(gotURL, "orderBy=id") {
		t.Fatalf("missing orderBy in %q", gotURL)
	}
	if !strings.Contains(gotURL, "perPage=20") {
		t.Fatalf("missing perPage in %q", gotURL)
	}
}
