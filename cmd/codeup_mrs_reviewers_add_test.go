package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
)

func resetMrsReviewersAddFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"repo", "local-id", "reviewer"} {
		f := codeupMrsReviewersAddCmd.Flags().Lookup(name)
		if f == nil {
			continue
		}
		_ = codeupMrsReviewersAddCmd.Flags().Set(name, f.DefValue)
		f.Changed = false
	}
}

// TestMrsReviewersAddDryRunBody wires --reviewer CSV into OpenAPI person/REVIEWER
// body userIds via dry-run request preview (no network).
func TestMrsReviewersAddDryRunBody(t *testing.T) {
	t.Setenv(config.EnvAccessToken, "test-token-mrs-reviewers-add-not-real")
	t.Setenv(config.EnvOrganizationID, "org-mrs-reviewers-add-test")
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

	resetMrsReviewersAddFlags(t)
	rootCmd.SetArgs([]string{
		"codeup", "mrs", "reviewers", "add",
		"--repo", "4951320",
		"--local-id", "42",
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
	raw, err := json.Marshal(env.Request)
	if err != nil {
		t.Fatalf("request type %T: %v", env.Request, env.Request)
	}
	var req map[string]any
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("request re-decode: %v", err)
	}
	urlStr, _ := req["url"].(string)
	if !strings.Contains(urlStr, "/changeRequests/42/person/REVIEWER") {
		t.Fatalf("url should contain person/REVIEWER: %q", urlStr)
	}
	method, _ := req["method"].(string)
	if method != "POST" {
		t.Fatalf("method=%q", method)
	}
	body, ok := req["body"].(map[string]any)
	if !ok {
		t.Fatalf("body missing: %#v", req)
	}
	ids, ok := body["userIds"].([]any)
	if !ok {
		t.Fatalf("userIds missing/wrong type: %#v", body["userIds"])
	}
	if len(ids) != 2 || ids[0] != "uid-a" || ids[1] != "uid-b" {
		t.Fatalf("userIds=%#v", ids)
	}
	if _, hasWrong := body["reviewerUserIds"]; hasWrong {
		t.Fatalf("must not send create-only reviewerUserIds on person API: %#v", body)
	}
}

func TestMrsReviewersAddHelpMentionsPersonAPI(t *testing.T) {
	var buf bytes.Buffer
	codeupMrsReviewersAddCmd.SetOut(&buf)
	codeupMrsReviewersAddCmd.SetErr(&buf)
	codeupMrsReviewersAddCmd.Help()
	out := buf.String()
	if !strings.Contains(out, "person/REVIEWER") {
		t.Fatalf("help should mention person/REVIEWER:\n%s", out)
	}
	if !strings.Contains(out, "userIds") {
		t.Fatalf("help should mention userIds:\n%s", out)
	}
}
