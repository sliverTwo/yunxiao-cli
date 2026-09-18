package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// CheckCacheFile is stored under the config dir (alongside profiles/).
	CheckCacheFile = "update_check.json"
	// DefaultCheckInterval is the minimum time between network release checks.
	DefaultCheckInterval = 24 * time.Hour
	// DefaultHintFetchTimeout bounds opportunistic Latest Release lookups.
	DefaultHintFetchTimeout = 2 * time.Second
)

// CheckCache persists the last opportunistic update check.
type CheckCache struct {
	LastCheckUnix int64  `json:"last_check_unix"`
	LatestTag     string `json:"latest_tag"`
}

// CheckCachePath returns <configDir>/update_check.json.
func CheckCachePath(configDir string) string {
	return filepath.Join(configDir, CheckCacheFile)
}

// LoadCheckCache reads the cache file. Missing file yields a zero cache and nil error.
func LoadCheckCache(path string) (CheckCache, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return CheckCache{}, nil
		}
		return CheckCache{}, err
	}
	var c CheckCache
	if err := json.Unmarshal(b, &c); err != nil {
		return CheckCache{}, err
	}
	return c, nil
}

// SaveCheckCache writes the cache file (0600), creating the parent dir if needed.
func SaveCheckCache(path string, c CheckCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o600)
}

// CacheFresh reports whether last_check_unix is within interval of now.
func CacheFresh(c CheckCache, now time.Time, interval time.Duration) bool {
	if c.LastCheckUnix <= 0 || interval <= 0 {
		return false
	}
	last := time.Unix(c.LastCheckUnix, 0)
	if last.After(now) {
		// Clock skew / future stamp: treat as fresh to avoid hammering the network.
		return true
	}
	return now.Sub(last) < interval
}

// ShouldSkipUpdateHint is true for update/self-update/completion, JSON stdout, or when disabled.
func ShouldSkipUpdateHint(commandName, format string, disabled bool) bool {
	if disabled {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(commandName)) {
	case "update", "self-update", "completion":
		return true
	}
	f := strings.ToLower(strings.TrimSpace(format))
	if f == "" || f == "json" {
		return true
	}
	return false
}

// ShouldHint reports whether latest is strictly newer than current (after normalize).
func ShouldHint(current, latest string) bool {
	current = strings.TrimSpace(current)
	latest = strings.TrimSpace(latest)
	if current == "" || latest == "" {
		return false
	}
	return NewerAvailable(current, latest)
}

// FormatHintMessage is the one-liner printed to stderr.
func FormatHintMessage(current, latest string) string {
	cur := NormalizeVersion(current)
	lat := NormalizeVersion(latest)
	return fmt.Sprintf("A newer yunxiao is available (%s → %s). Run: yunxiao update", cur, lat)
}

// HintConfig drives MaybePrintUpdateHint (tests inject clock/IO/fetch).
type HintConfig struct {
	CurrentVersion string
	CommandName    string
	Format         string
	ConfigDir      string
	Stderr         io.Writer
	Now            time.Time
	Interval       time.Duration
	FetchTimeout   time.Duration
	Client         HTTPDoer
	Repo           string
	Disabled       bool
	// FetchLatestTag optional override for tests (no network).
	FetchLatestTag func(ctx context.Context) (string, error)
}

// MaybePrintUpdateHint optionally checks GitHub Latest Release (≤1×/24h) and prints a stderr hint.
// Never returns an error; failures are silent. Does not download.
func MaybePrintUpdateHint(ctx context.Context, cfg HintConfig) {
	defer func() {
		// Never let a panic from hinting fail the user command.
		_ = recover()
	}()

	if ShouldSkipUpdateHint(cfg.CommandName, cfg.Format, cfg.Disabled) {
		return
	}
	if strings.TrimSpace(cfg.ConfigDir) == "" || strings.TrimSpace(cfg.CurrentVersion) == "" {
		return
	}
	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}
	now := cfg.Now
	if now.IsZero() {
		now = time.Now()
	}
	interval := cfg.Interval
	if interval <= 0 {
		interval = DefaultCheckInterval
	}
	timeout := cfg.FetchTimeout
	if timeout <= 0 {
		timeout = DefaultHintFetchTimeout
	}

	path := CheckCachePath(cfg.ConfigDir)
	cache, err := LoadCheckCache(path)
	if err != nil {
		cache = CheckCache{}
	}

	latest := strings.TrimSpace(cache.LatestTag)
	if !CacheFresh(cache, now, interval) {
		tag, fetchErr := fetchLatestTagForHint(ctx, cfg, timeout)
		// Always stamp last_check on attempt so failures still rate-limit.
		cache.LastCheckUnix = now.Unix()
		if fetchErr == nil && strings.TrimSpace(tag) != "" {
			cache.LatestTag = strings.TrimSpace(tag)
			latest = cache.LatestTag
		}
		_ = SaveCheckCache(path, cache)
	}

	if !ShouldHint(cfg.CurrentVersion, latest) {
		return
	}
	_, _ = fmt.Fprintln(cfg.Stderr, FormatHintMessage(cfg.CurrentVersion, latest))
}

func fetchLatestTagForHint(ctx context.Context, cfg HintConfig, timeout time.Duration) (string, error) {
	if cfg.FetchLatestTag != nil {
		cctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return cfg.FetchLatestTag(cctx)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	repo := cfg.Repo
	if repo == "" {
		repo = GithubRepo()
	}
	rel, err := FetchLatestRelease(cctx, client, repo)
	if err != nil {
		return "", err
	}
	return rel.TagName, nil
}
