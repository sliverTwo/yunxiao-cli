package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

func TestRefreshAfterTransitionOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method=%s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"identifier": "wi-1", "status": "processing"})
	}))
	defer srv.Close()

	c := &client.Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var warn bytes.Buffer
	prev := refreshWarnOut
	refreshWarnOut = &warn
	t.Cleanup(func() { refreshWarnOut = prev })

	refreshed, ok := refreshAfterTransition(context.Background(), c, "/workitems/wi-1", "wi-1")
	if !ok {
		t.Fatal("refresh_ok want true")
	}
	if refreshed == nil || refreshed["identifier"] != "wi-1" {
		t.Fatalf("%v", refreshed)
	}
	if warn.Len() != 0 {
		t.Fatalf("unexpected warning: %s", warn.String())
	}
}

func TestRefreshAfterTransitionFailWarning(t *testing.T) {
	restore := client.SetRetrySleepForTest(func(ctx context.Context, d time.Duration) error { return nil })
	t.Cleanup(restore)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errorMsg":"boom"}`))
	}))
	defer srv.Close()

	c := &client.Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok", UserAgent: "t"}
	var warn bytes.Buffer
	prev := refreshWarnOut
	refreshWarnOut = &warn
	t.Cleanup(func() { refreshWarnOut = prev })

	refreshed, ok := refreshAfterTransition(context.Background(), c, "/workitems/ZYPT-1", "ZYPT-1")
	if ok {
		t.Fatal("refresh_ok want false")
	}
	// success path still uses refresh_ok=false; refreshed may be nil
	_ = refreshed
	msg := warn.String()
	if !strings.Contains(msg, "warning: transition succeeded but refresh failed") {
		t.Fatalf("warning missing: %q", msg)
	}
	if !strings.Contains(msg, "ZYPT-1") {
		t.Fatalf("id missing in warning: %q", msg)
	}

	// Envelope contract: Success with refresh_ok false still ok=true
	var stdout bytes.Buffer
	prevOut := output.Stdout
	prevJQ := output.JQ
	prevFmt := output.Format
	output.Stdout = &stdout
	output.JQ = ""
	output.Format = "json"
	t.Cleanup(func() {
		output.Stdout = prevOut
		output.JQ = prevJQ
		output.Format = prevFmt
	})
	if err := output.Success(map[string]any{"refresh_ok": false, "work_item": "ZYPT-1"}, map[string]any{"risk": "write"}); err != nil {
		t.Fatal(err)
	}
	var env output.Envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK {
		t.Fatalf("success path must stay ok=true: %+v", env)
	}
	data, _ := env.Data.(map[string]any)
	if data["refresh_ok"] != false {
		t.Fatalf("refresh_ok=%v", data["refresh_ok"])
	}
}

// compile-time sanity: fake getter error path without HTTP
type fakeGetter struct{ err error }

func (f fakeGetter) Get(ctx context.Context, path string, query map[string]string, out any) error {
	if f.err != nil {
		return f.err
	}
	if m, ok := out.(*map[string]any); ok {
		*m = map[string]any{"ok": true}
	}
	return nil
}

func TestRefreshAfterTransitionFakeGetter(t *testing.T) {
	var warn bytes.Buffer
	prev := refreshWarnOut
	refreshWarnOut = &warn
	t.Cleanup(func() { refreshWarnOut = prev })

	_, ok := refreshAfterTransition(context.Background(), fakeGetter{err: fmt.Errorf("net down")}, "/p", "id-9")
	if ok {
		t.Fatal("want false")
	}
	if !strings.Contains(warn.String(), "warning: transition succeeded but refresh failed for id-9") {
		t.Fatalf("%q", warn.String())
	}

	ref, ok := refreshAfterTransition(context.Background(), fakeGetter{}, "/p", "id-9")
	if !ok || ref["ok"] != true {
		t.Fatalf("%v %v", ok, ref)
	}
}
