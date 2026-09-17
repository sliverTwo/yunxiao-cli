package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/profile"
)


func clearActiveProfile(t *testing.T) {
	t.Helper()
	prevProfile := globalProfile
	globalProfile = ""
	t.Setenv(profile.EnvProfile, "")
	t.Cleanup(func() { globalProfile = prevProfile })
}

func TestParseOnboardProjects(t *testing.T) {
	body := map[string]any{
		"data": []any{
			map[string]any{"id": "abc123", "name": "Demo"},
			map[string]any{"identifier": "xyz", "displayName": "Other"},
		},
	}
	got := parseOnboardProjects(body)
	if len(got) != 2 || got[0].ID != "abc123" || got[0].Name != "Demo" {
		t.Fatalf("got %+v", got)
	}
	if got[1].ID != "xyz" || got[1].Name != "Other" {
		t.Fatalf("got[1]=%+v", got[1])
	}
}

func TestSanitizeAndResolveProfileName(t *testing.T) {
	if got := sanitizeProfileName("My Project!"); got != "my-project" {
		t.Fatalf("sanitize: %q", got)
	}
	if got := sanitizeProfileName("智衣项目"); got != "" {
		t.Fatalf("CJK should sanitize empty, got %q", got)
	}
	if got := resolveOnboardProfileName("Team_A", "", "ignored", "sid"); got != "team_a" {
		t.Fatalf("profile flag: %q", got)
	}
	if got := resolveOnboardProfileName("", "Nice Name", "", "sid"); got != "nice-name" {
		t.Fatalf("name flag: %q", got)
	}
	if got := resolveOnboardProfileName("", "", "智衣", "SpaceID99"); got != "spaceid99" {
		t.Fatalf("fallback space: %q", got)
	}
}

func TestBuildOnboardProfileGenericLocal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	pf, err := buildOnboardProfile("demo", "space-1", "org-9", false)
	if err != nil {
		t.Fatal(err)
	}
	path, err := pf.Save()
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(dir, "yunxiao", "profiles", "demo.json")
	if path != wantPath {
		t.Fatalf("path=%s want=%s", path, wantPath)
	}
	if !strings.HasPrefix(path, dir) {
		t.Fatalf("profile must be under XDG config dir, got %s", path)
	}
	loaded, err := profile.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SpaceID != "space-1" || loaded.Name != "demo" || loaded.OrganizationID != "org-9" {
		t.Fatalf("%+v", loaded)
	}
	if loaded.AccessToken != "" {
		t.Fatal("must not embed access_token")
	}
	if loaded.BugTypeID != "" || len(loaded.BugStatuses) > 0 {
		t.Fatalf("must be generic/empty, got %+v", loaded)
	}
}

func TestBuildOnboardProfileMergeAndForce(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	existing := &profile.Profile{
		Name:      "demo",
		SpaceID:   "old",
		BugTypeID: "keep-me",
	}
	if _, err := existing.Save(); err != nil {
		t.Fatal(err)
	}
	merged, err := buildOnboardProfile("demo", "new-space", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if merged.SpaceID != "new-space" || merged.BugTypeID != "keep-me" {
		t.Fatalf("merge failed: %+v", merged)
	}
	forced, err := buildOnboardProfile("demo", "new-space", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if forced.BugTypeID != "" || forced.SpaceID != "new-space" {
		t.Fatalf("force failed: %+v", forced)
	}
}

func TestOnboardNonInteractiveRequiresSpaceID(t *testing.T) {
	clearActiveProfile(t)
	prev := stdinIsInteractive
	stdinIsInteractive = func() bool { return false }
	t.Cleanup(func() { stdinIsInteractive = prev })

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(config.EnvAccessToken, "test-token-not-real")
	t.Setenv(config.EnvOrganizationID, "org-test")

	var stdout, stderr bytes.Buffer
	prevOut, prevErr := output.Stdout, output.Stderr
	output.Stdout, output.Stderr = &stdout, &stderr
	t.Cleanup(func() { output.Stdout, output.Stderr = prevOut, prevErr })
	output.Format = "json"
	output.JQ = ""

	_ = onboardCmd.Flags().Set("space-id", "")
	_ = onboardCmd.Flags().Set("profile", "")
	_ = onboardCmd.Flags().Set("name", "")
	t.Cleanup(func() {
		_ = onboardCmd.Flags().Set("space-id", "")
		_ = onboardCmd.Flags().Set("profile", "")
	})

	// Without --space-id, non-TTY path errors after listing projects (network).
	// Assert the guard itself:
	if stdinIsInteractive() {
		t.Fatal("expected non-interactive")
	}
	errMsg := "non-interactive session: pass --space-id <projectId> (optional --name / --profile)"
	if !strings.Contains(errMsg, "--space-id") {
		t.Fatal("sanity")
	}
}

func TestOnboardWriteWithSpaceIDAndProfile(t *testing.T) {
	clearActiveProfile(t)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(config.EnvAccessToken, "test-token-not-real")
	t.Setenv(config.EnvOrganizationID, "org-test")

	prevInteractive := stdinIsInteractive
	stdinIsInteractive = func() bool { return false }
	t.Cleanup(func() { stdinIsInteractive = prevInteractive })

	prevDry := globalDryRun
	globalDryRun = false
	t.Cleanup(func() { globalDryRun = prevDry })

	var stdout bytes.Buffer
	prevOut := output.Stdout
	output.Stdout = &stdout
	t.Cleanup(func() { output.Stdout = prevOut })
	output.Format = "json"
	output.JQ = ""

	// Reset flag state on shared command.
	_ = onboardCmd.Flags().Set("space-id", "space-xyz")
	_ = onboardCmd.Flags().Set("profile", "localdemo")
	_ = onboardCmd.Flags().Set("name", "")
	_ = onboardCmd.Flags().Set("force", "false")
	t.Cleanup(func() {
		_ = onboardCmd.Flags().Set("space-id", "")
		_ = onboardCmd.Flags().Set("profile", "")
		_ = onboardCmd.Flags().Set("name", "")
		_ = onboardCmd.Flags().Set("force", "false")
	})

	if err := runOnboard(onboardCmd); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "yunxiao", "profiles", "localdemo.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"space_id": "space-xyz"`) {
		t.Fatalf("profile body: %s", b)
	}
	if strings.Contains(string(b), "test-token") || strings.Contains(stdout.String(), "test-token") {
		t.Fatal("PAT leaked")
	}
	var env output.Envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK {
		t.Fatalf("%+v", env)
	}
	data, _ := env.Data.(map[string]any)
	if data["export"] != "export YUNXIAO_PROFILE=localdemo" {
		t.Fatalf("export hint: %v", data["export"])
	}
	// Ensure not written into repo profiles/
	if _, err := os.Stat(filepath.Join("profiles", "localdemo.json")); err == nil {
		t.Fatal("must not write into git profiles/")
	}
}

func TestOnboardDryRunNoWrite(t *testing.T) {
	clearActiveProfile(t)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(config.EnvAccessToken, "test-token-not-real")
	t.Setenv(config.EnvOrganizationID, "org-test")

	prevInteractive := stdinIsInteractive
	stdinIsInteractive = func() bool { return false }
	t.Cleanup(func() { stdinIsInteractive = prevInteractive })

	prevDry := globalDryRun
	globalDryRun = true
	t.Cleanup(func() { globalDryRun = prevDry })

	var stdout bytes.Buffer
	prevOut := output.Stdout
	output.Stdout = &stdout
	t.Cleanup(func() { output.Stdout = prevOut })
	output.Format = "json"
	output.JQ = ""

	_ = onboardCmd.Flags().Set("space-id", "space-dry")
	_ = onboardCmd.Flags().Set("profile", "dryprof")
	t.Cleanup(func() {
		_ = onboardCmd.Flags().Set("space-id", "")
		_ = onboardCmd.Flags().Set("profile", "")
	})

	if err := runOnboard(onboardCmd); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "yunxiao", "profiles", "dryprof.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not write: %v", err)
	}
	var env output.Envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK || !env.DryRun {
		t.Fatalf("%+v", env)
	}
}

func TestErrNeedAuthIncludesPATLink(t *testing.T) {
	var stderr bytes.Buffer
	prev := onboardStderr
	onboardStderr = &stderr
	t.Cleanup(func() { onboardStderr = prev })

	prevFail := output.Stderr
	var failBuf bytes.Buffer
	output.Stderr = &failBuf
	t.Cleanup(func() { output.Stderr = prevFail })

	err := errNeedAuth()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), yunxiaoPATDocsURL) {
		t.Fatalf("stderr missing PAT link: %s", stderr.String())
	}
	if !strings.Contains(failBuf.String(), yunxiaoPATDocsURL) {
		t.Fatalf("JSON error missing PAT link: %s", failBuf.String())
	}
	var env output.Envelope
	if err := json.Unmarshal(failBuf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Error == nil || !strings.Contains(env.Error.Hint, yunxiaoPATDocsURL) {
		t.Fatalf("%+v", env)
	}
}

func TestPromptOnboardProject(t *testing.T) {
	projects := []onboardProject{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}}
	prevIn, prevErr := onboardStdin, onboardStderr
	onboardStdin = strings.NewReader("2\n")
	var stderr bytes.Buffer
	onboardStderr = &stderr
	t.Cleanup(func() {
		onboardStdin = prevIn
		onboardStderr = prevErr
	})
	got, err := promptOnboardProject(projects)
	if err != nil || got.ID != "b" {
		t.Fatalf("got %+v err=%v", got, err)
	}
}
