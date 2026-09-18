package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

func resetMrsCreateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"repo", "source", "target", "title", "description", "reviewer", "source-project-id", "target-project-id", "create-from"} {
		f := codeupMrsCreateCmd.Flags().Lookup(name)
		if f == nil {
			continue
		}
		def := f.DefValue
		if name == "create-from" && def == "" {
			def = "WEB"
		}
		_ = codeupMrsCreateCmd.Flags().Set(name, def)
		f.Changed = false
	}
}

// TestMrsCreateReviewerDryRunBody wires --reviewer CSV into OpenAPI reviewerUserIds
// via dry-run request preview (no network).
func TestMrsCreateReviewerDryRunBody(t *testing.T) {
	t.Setenv(config.EnvAccessToken, "test-token-mrs-create-reviewer-not-real")
	t.Setenv(config.EnvOrganizationID, "org-mrs-create-reviewer-test")
	t.Setenv(config.EnvEdition, "central")
	t.Setenv("YUNXIAO_PROFILE", "")

	var stdout bytes.Buffer
	prevOut := output.Stdout
	prevErr := output.Stderr
	prevJQ := output.JQ
	prevFmt := output.Format
	output.Stdout = &stdout
	output.Stderr = &bytes.Buffer{}
	output.JQ = ""
	output.Format = "json"
	t.Cleanup(func() {
		output.Stdout = prevOut
		output.Stderr = prevErr
		output.JQ = prevJQ
		output.Format = prevFmt
	})

	prevYes := globalYes
	prevDry := globalDryRun
	prevProfile := globalProfile
	prevOrg := globalOrg
	globalYes = false
	globalDryRun = true
	globalProfile = ""
	globalOrg = ""
	t.Cleanup(func() {
		globalYes = prevYes
		globalDryRun = prevDry
		globalProfile = prevProfile
		globalOrg = prevOrg
	})

	resetMrsCreateFlags(t)
	rootCmd.SetArgs([]string{
		"codeup", "mrs", "create",
		"--repo", "4951320",
		"--source", "feat/x",
		"--target", "master",
		"--title", "feat: x",
		"--reviewer", "uid-a, uid-b",
		"--dry-run",
	})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v\nstdout=%s", err, stdout.String())
	}

	var env output.Envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout JSON: %v / %s", err, stdout.Bytes())
	}
	if !env.OK || !env.DryRun {
		t.Fatalf("envelope: %+v", env)
	}
	req, ok := env.Request.(map[string]any)
	if !ok {
		// Request may decode as map via json.Unmarshal into any
		raw, err := json.Marshal(env.Request)
		if err != nil {
			t.Fatalf("request type %T: %v", env.Request, env.Request)
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Fatalf("request re-decode: %v", err)
		}
	}
	body, ok := req["body"].(map[string]any)
	if !ok {
		t.Fatalf("body missing: %#v", req)
	}
	ids, ok := body["reviewerUserIds"].([]any)
	if !ok {
		t.Fatalf("reviewerUserIds missing/wrong type: %#v", body["reviewerUserIds"])
	}
	if len(ids) != 2 || ids[0] != "uid-a" || ids[1] != "uid-b" {
		t.Fatalf("reviewerUserIds=%#v", ids)
	}
	if _, hasWrong := body["reviewerIds"]; hasWrong {
		t.Fatalf("must not send legacy reviewerIds: %#v", body)
	}
}

func TestMrsCreateOmitsReviewerWhenEmpty(t *testing.T) {
	t.Setenv(config.EnvAccessToken, "test-token-mrs-create-reviewer-not-real")
	t.Setenv(config.EnvOrganizationID, "org-mrs-create-reviewer-test")
	t.Setenv(config.EnvEdition, "central")
	t.Setenv("YUNXIAO_PROFILE", "")

	var stdout bytes.Buffer
	prevOut := output.Stdout
	prevErr := output.Stderr
	prevJQ := output.JQ
	prevFmt := output.Format
	output.Stdout = &stdout
	output.Stderr = &bytes.Buffer{}
	output.JQ = ""
	output.Format = "json"
	t.Cleanup(func() {
		output.Stdout = prevOut
		output.Stderr = prevErr
		output.JQ = prevJQ
		output.Format = prevFmt
	})

	prevYes := globalYes
	prevDry := globalDryRun
	prevProfile := globalProfile
	prevOrg := globalOrg
	globalYes = false
	globalDryRun = true
	globalProfile = ""
	globalOrg = ""
	t.Cleanup(func() {
		globalYes = prevYes
		globalDryRun = prevDry
		globalProfile = prevProfile
		globalOrg = prevOrg
	})

	resetMrsCreateFlags(t)
	rootCmd.SetArgs([]string{
		"codeup", "mrs", "create",
		"--repo", "4951320",
		"--source", "feat/x",
		"--target", "master",
		"--title", "feat: x",
		"--dry-run",
	})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v\nstdout=%s", err, stdout.String())
	}

	var env output.Envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout JSON: %v / %s", err, stdout.Bytes())
	}
	raw, _ := json.Marshal(env.Request)
	var req map[string]any
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	body, _ := req["body"].(map[string]any)
	if _, has := body["reviewerUserIds"]; has {
		t.Fatalf("empty --reviewer must omit reviewerUserIds: %#v", body)
	}
}
