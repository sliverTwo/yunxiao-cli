package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

// exitPanic is thrown by processExit override so cobra Run can unwind without
// killing the test process.
type exitPanic struct{ code int }

// TestCobraExecuteHighRiskExit10 drives a real cobra command path
// (workitem delete → runMutating HighRiskWrite) without --yes and asserts
// confirmation_required + exit 10. Auth/org come from env so the gate trips
// before any network call (mustClient + ProjexPath succeed locally).
func TestCobraExecuteHighRiskExit10(t *testing.T) {
	t.Setenv(config.EnvAccessToken, "test-token-for-gate-e2e-not-real")
	t.Setenv(config.EnvOrganizationID, "org-gate-e2e-test")
	t.Setenv(config.EnvEdition, "central")
	// Avoid profile side effects
	t.Setenv("YUNXIAO_PROFILE", "")

	var stderr bytes.Buffer
	prevErr := output.Stderr
	prevOut := output.Stdout
	prevJQ := output.JQ
	prevFmt := output.Format
	output.Stderr = &stderr
	output.Stdout = &bytes.Buffer{}
	output.JQ = ""
	output.Format = "json"
	t.Cleanup(func() {
		output.Stderr = prevErr
		output.Stdout = prevOut
		output.JQ = prevJQ
		output.Format = prevFmt
	})

	prevYes := globalYes
	prevDry := globalDryRun
	prevProfile := globalProfile
	prevOrg := globalOrg
	globalYes = false
	globalDryRun = false
	globalProfile = ""
	globalOrg = ""
	t.Cleanup(func() {
		globalYes = prevYes
		globalDryRun = prevDry
		globalProfile = prevProfile
		globalOrg = prevOrg
	})

	prevExit := processExit
	var gotCode int
	processExit = func(code int) {
		gotCode = code
		panic(exitPanic{code: code})
	}
	t.Cleanup(func() { processExit = prevExit })

	rootCmd.SetArgs([]string{"workitem", "delete", "--id", "wi-gate-e2e"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	var execErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(exitPanic); !ok {
					panic(r)
				}
			}
		}()
		execErr = rootCmd.Execute()
	}()

	if gotCode != risk.ExitConfirmationRequired {
		t.Fatalf("exit code=%d want %d; execErr=%v stderr=%s", gotCode, risk.ExitConfirmationRequired, execErr, stderr.String())
	}

	raw := stderr.Bytes()
	if len(raw) == 0 {
		t.Fatal("expected confirmation JSON on stderr")
	}
	var env output.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("stderr not JSON: %v / %s", err, raw)
	}
	if env.OK || env.Error == nil || env.Error.Subtype != "confirmation_required" {
		t.Fatalf("%+v", env)
	}
	if env.Error.Risk != string(risk.HighRiskWrite) {
		t.Fatalf("risk=%q", env.Error.Risk)
	}
	if !strings.Contains(env.Error.Action, "workitem delete") && env.Error.Action != "workitem delete" {
		// Action should be the runMutating action string
		if env.Error.Action == "" {
			t.Fatalf("empty action: %+v", env.Error)
		}
	}
	if env.Error.Action != "workitem delete" {
		t.Fatalf("action=%q", env.Error.Action)
	}
}
