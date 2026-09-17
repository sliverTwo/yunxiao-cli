package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaskToken(t *testing.T) {
	if MaskToken("") != "" {
		t.Fatal("empty")
	}
	if MaskToken("abcd") != "****" {
		t.Fatal("short")
	}
	m := MaskToken("abcdefghijklmnop")
	if m[:4] != "abcd" || m[len(m)-4:] != "mnop" {
		t.Fatalf("got %s", m)
	}
}

func TestResolvePrefersEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(EnvAccessToken, "env-token-value-xx")
	t.Setenv(EnvAPIBaseURL, "https://example.test")
	t.Setenv(EnvOrganizationID, "org-env")
	t.Setenv(EnvEdition, "central")

	cfgDir := filepath.Join(dir, "yunxiao")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"access_token":"cfg-token","api_base_url":"https://cfg.test"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "env" || r.AccessToken != "env-token-value-xx" {
		t.Fatalf("%+v", r)
	}
	if r.APIBaseURL != "https://example.test" {
		t.Fatalf("base=%s", r.APIBaseURL)
	}
	if r.OrganizationID != "org-env" {
		t.Fatalf("org=%s", r.OrganizationID)
	}
}

func TestSaveAndLoadFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(EnvAccessToken, "")
	os.Unsetenv(EnvAccessToken)
	p, err := SaveFile(File{AccessToken: "tokentoken", OrganizationID: "o1", Edition: "central"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
	r, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "config" || r.AccessToken != "tokentoken" {
		t.Fatalf("%+v", r)
	}
}

func TestResolveTokenPrecedenceEnvBeatsProfileBeatsConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	os.Unsetenv(EnvAccessToken)

	cfgDir := filepath.Join(dir, "yunxiao")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"access_token":"cfg-token-xxxx"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// config only
	r, err := ResolveWithProfileToken("")
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "config" || r.AccessToken != "cfg-token-xxxx" {
		t.Fatalf("config-only: %+v", r)
	}

	// profile beats config
	r, err = ResolveWithProfileToken("profile-token-yy")
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "profile" || r.AccessToken != "profile-token-yy" {
		t.Fatalf("profile: %+v", r)
	}

	// env beats profile
	t.Setenv(EnvAccessToken, "env-token-zzzz")
	r, err = ResolveWithProfileToken("profile-token-yy")
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "env" || r.AccessToken != "env-token-zzzz" {
		t.Fatalf("env: %+v", r)
	}

	// empty profile token falls through to config when no env
	t.Setenv(EnvAccessToken, "")
	os.Unsetenv(EnvAccessToken)
	r, err = ResolveWithProfileToken("   ")
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "config" || r.AccessToken != "cfg-token-xxxx" {
		t.Fatalf("whitespace profile: %+v", r)
	}
}

func TestResolveWithProfileTokenNone(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	os.Unsetenv(EnvAccessToken)
	r, err := ResolveWithProfileToken("")
	if err != nil {
		t.Fatal(err)
	}
	if r.TokenSource != "none" || r.AccessToken != "" {
		t.Fatalf("%+v", r)
	}
}
