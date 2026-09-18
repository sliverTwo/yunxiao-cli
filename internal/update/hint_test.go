package update

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCacheFresh(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	interval := 24 * time.Hour

	if CacheFresh(CheckCache{}, now, interval) {
		t.Fatal("zero cache should not be fresh")
	}
	if CacheFresh(CheckCache{LastCheckUnix: now.Add(-25 * time.Hour).Unix()}, now, interval) {
		t.Fatal("25h old should not be fresh")
	}
	if !CacheFresh(CheckCache{LastCheckUnix: now.Add(-23 * time.Hour).Unix()}, now, interval) {
		t.Fatal("23h old should be fresh")
	}
	if !CacheFresh(CheckCache{LastCheckUnix: now.Add(time.Hour).Unix()}, now, interval) {
		t.Fatal("future stamp should be treated as fresh")
	}
	if CacheFresh(CheckCache{LastCheckUnix: now.Unix()}, now, 0) {
		t.Fatal("non-positive interval is never fresh")
	}
}

func TestShouldSkipUpdateHint(t *testing.T) {
	cases := []struct {
		cmd, format string
		disabled    bool
		skip        bool
	}{
		{"whoami", "pretty", false, false},
		{"whoami", "json", false, true},
		{"whoami", "", false, true},
		{"whoami", "pretty", true, true},
		{"update", "pretty", false, true},
		{"self-update", "pretty", false, true},
		{"completion", "pretty", false, true},
		{"UPDATE", "pretty", false, true},
		{"doctor", "Pretty", false, false},
	}
	for _, c := range cases {
		if got := ShouldSkipUpdateHint(c.cmd, c.format, c.disabled); got != c.skip {
			t.Fatalf("ShouldSkipUpdateHint(%q,%q,%v)=%v want %v", c.cmd, c.format, c.disabled, got, c.skip)
		}
	}
}

func TestShouldHint(t *testing.T) {
	if !ShouldHint("0.16.0", "v0.16.1") {
		t.Fatal("expected hint")
	}
	if ShouldHint("0.16.1", "0.16.1") {
		t.Fatal("same version")
	}
	if ShouldHint("0.17.0", "0.16.1") {
		t.Fatal("newer current")
	}
	if ShouldHint("", "0.16.1") || ShouldHint("0.16.0", "") {
		t.Fatal("empty should not hint")
	}
}

func TestFormatHintMessage(t *testing.T) {
	s := FormatHintMessage("v0.16.0", "v0.16.1")
	want := "发现新版本 yunxiao：0.16.0 → 0.16.1。运行：yunxiao update"
	if s != want {
		t.Fatalf("got %q want %q", s, want)
	}
}

func TestCheckCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := CheckCachePath(dir)
	want := CheckCache{LastCheckUnix: 123, LatestTag: "v0.16.1"}
	if err := SaveCheckCache(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCheckCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %#v want %#v", got, want)
	}
	missing, err := LoadCheckCache(filepath.Join(dir, "nope.json"))
	if err != nil || missing != (CheckCache{}) {
		t.Fatalf("missing: %#v %v", missing, err)
	}
}

func TestMaybePrintUpdateHintUsesCacheWithoutNetwork(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	if err := SaveCheckCache(CheckCachePath(dir), CheckCache{
		LastCheckUnix: now.Add(-time.Hour).Unix(),
		LatestTag:     "v0.16.1",
	}); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	fetches := 0
	MaybePrintUpdateHint(context.Background(), HintConfig{
		CurrentVersion: "0.16.0",
		CommandName:    "whoami",
		Format:         "pretty",
		ConfigDir:      dir,
		Stderr:         &buf,
		Now:            now,
		Disabled:       false,
		FetchLatestTag: func(ctx context.Context) (string, error) {
			fetches++
			return "v9.9.9", nil
		},
	})
	if fetches != 0 {
		t.Fatalf("fresh cache must not fetch, got %d", fetches)
	}
	if !strings.Contains(buf.String(), "发现新版本 yunxiao：0.16.0 → 0.16.1。运行：yunxiao update") {
		t.Fatalf("hint: %q", buf.String())
	}
}

func TestMaybePrintUpdateHintFetchesWhenStale(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	if err := SaveCheckCache(CheckCachePath(dir), CheckCache{
		LastCheckUnix: now.Add(-25 * time.Hour).Unix(),
		LatestTag:     "v0.16.0",
	}); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	MaybePrintUpdateHint(context.Background(), HintConfig{
		CurrentVersion: "0.16.0",
		CommandName:    "doctor",
		Format:         "pretty",
		ConfigDir:      dir,
		Stderr:         &buf,
		Now:            now,
		FetchLatestTag: func(ctx context.Context) (string, error) {
			return "v0.16.2", nil
		},
	})
	if !strings.Contains(buf.String(), "发现新版本 yunxiao：0.16.0 → 0.16.2。运行：yunxiao update") {
		t.Fatalf("hint: %q", buf.String())
	}
	got, err := LoadCheckCache(CheckCachePath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if got.LatestTag != "v0.16.2" || got.LastCheckUnix != now.Unix() {
		t.Fatalf("cache not updated: %#v", got)
	}
}

func TestMaybePrintUpdateHintSilentOnFetchError(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	var buf bytes.Buffer
	MaybePrintUpdateHint(context.Background(), HintConfig{
		CurrentVersion: "0.16.0",
		CommandName:    "doctor",
		Format:         "pretty",
		ConfigDir:      dir,
		Stderr:         &buf,
		Now:            now,
		FetchLatestTag: func(ctx context.Context) (string, error) {
			return "", os.ErrDeadlineExceeded
		},
	})
	if buf.Len() != 0 {
		t.Fatalf("expected silent, got %q", buf.String())
	}
	got, err := LoadCheckCache(CheckCachePath(dir))
	if err != nil {
		t.Fatal(err)
	}
	// Rate-limit stamp even on failure.
	if got.LastCheckUnix != now.Unix() {
		t.Fatalf("expected last_check stamp, got %#v", got)
	}
}

func TestMaybePrintUpdateHintSkipsJSONAndUpdateCmd(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1_700_000_000, 0)
	_ = SaveCheckCache(CheckCachePath(dir), CheckCache{
		LastCheckUnix: now.Unix(),
		LatestTag:     "v9.0.0",
	})
	var buf bytes.Buffer
	cfg := HintConfig{
		CurrentVersion: "0.1.0",
		CommandName:    "whoami",
		Format:         "json",
		ConfigDir:      dir,
		Stderr:         &buf,
		Now:            now,
	}
	MaybePrintUpdateHint(context.Background(), cfg)
	if buf.Len() != 0 {
		t.Fatalf("json should skip: %q", buf.String())
	}
	cfg.Format = "pretty"
	cfg.CommandName = "update"
	MaybePrintUpdateHint(context.Background(), cfg)
	if buf.Len() != 0 {
		t.Fatalf("update should skip: %q", buf.String())
	}
}

func TestUpdateCheckDisabledEnv(t *testing.T) {
	t.Setenv(EnvUpdateCheck, "off")
	if !UpdateCheckDisabled() {
		t.Fatal("expected disabled")
	}
	t.Setenv(EnvUpdateCheck, "")
	if UpdateCheckDisabled() {
		t.Fatal("expected enabled")
	}
}
