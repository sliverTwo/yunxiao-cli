package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/output"
	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

func TestRunMutatingHighRiskWithoutYes(t *testing.T) {
	called := false
	err := runMutating("workitem delete", risk.HighRiskWrite, false, false, map[string]any{"x": 1}, func() error {
		called = true
		return nil
	})
	if called {
		t.Fatal("execFn must not run without --yes")
	}
	g, ok := err.(risk.GateResult)
	if !ok {
		t.Fatalf("want GateResult, got %T %v", err, err)
	}
	if g.Risk != risk.HighRiskWrite {
		t.Fatalf("risk=%s", g.Risk)
	}

	// Envelope shape matches handleErr confirmation path (exit 10).
	var stderr bytes.Buffer
	prev := output.Stderr
	output.Stderr = &stderr
	t.Cleanup(func() { output.Stderr = prev })
	failErr := output.Fail(output.ErrorBody{
		Type:    "confirmation",
		Subtype: "confirmation_required",
		Message: g.Error(),
		Hint:    g.Hint,
		Risk:    string(g.Risk),
		Action:  g.Action,
	}, risk.ExitConfirmationRequired)
	ee, ok := failErr.(output.ExitError)
	if !ok || ee.Code != risk.ExitConfirmationRequired {
		t.Fatalf("exit envelope: %v", failErr)
	}
	var env output.Envelope
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Error == nil || env.Error.Subtype != "confirmation_required" {
		t.Fatalf("%+v", env)
	}
}

func TestRunMutatingDryRunRedactsToken(t *testing.T) {
	tok := "real-secret-token-xyz"
	c := &client.Client{BaseURL: "https://example.test", Token: tok, UserAgent: "t"}
	prevReq := c.Preview("DELETE", "/oapi/v1/projex/organizations/o/workitems/1", nil, nil)

	var stdout bytes.Buffer
	prevOut := output.Stdout
	output.Stdout = &stdout
	prevJQ := output.JQ
	prevFmt := output.Format
	output.JQ = ""
	output.Format = "json"
	t.Cleanup(func() {
		output.Stdout = prevOut
		output.JQ = prevJQ
		output.Format = prevFmt
	})

	err := runMutating("workitem delete", risk.HighRiskWrite, true, false, prevReq, func() error {
		t.Fatal("must not execute on dry-run")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if strings.Contains(out, tok) {
		t.Fatalf("token leaked in dry-run output: %s", out)
	}
	if !strings.Contains(out, "(redacted)") {
		t.Fatalf("expected redacted marker: %s", out)
	}
	var env output.Envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK || !env.DryRun || env.Risk != string(risk.HighRiskWrite) {
		t.Fatalf("%+v", env)
	}
}

// TestRunMutatingWriteWithoutYes locks the contract: runMutating only gates
// HighRiskWrite. Plain Write proceeds without --yes (multi-step Write commands
// that need confirmation call risk.CheckConfirmed directly instead).
func TestRunMutatingWriteWithoutYes(t *testing.T) {
	called := false
	err := runMutating("workitem create", risk.Write, false, false, map[string]any{"x": 1}, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Write via runMutating must not require --yes: %v", err)
	}
	if !called {
		t.Fatal("execFn must run for Write without --yes")
	}
}

// TestCheckConfirmedWriteRequiresYes documents that multi-step Write shortcuts
// (+bug-transition / +transition / +explore-workflow / +bug-create) use
// CheckConfirmed(Write) and therefore DO require --yes; missing --yes yields
// GateResult and the same confirmation_required / exit-10 envelope as handleErr.
func TestCheckConfirmedWriteRequiresYes(t *testing.T) {
	err := risk.CheckConfirmed("workitem +bug-transition", risk.Write, false)
	if err == nil {
		t.Fatal("Write CheckConfirmed without --yes must gate")
	}
	g, ok := err.(risk.GateResult)
	if !ok {
		t.Fatalf("want GateResult, got %T %v", err, err)
	}
	if g.Risk != risk.Write {
		t.Fatalf("risk=%s want write", g.Risk)
	}
	if g.Action != "workitem +bug-transition" {
		t.Fatalf("action=%s", g.Action)
	}

	var stderr bytes.Buffer
	prev := output.Stderr
	output.Stderr = &stderr
	t.Cleanup(func() { output.Stderr = prev })
	failErr := output.Fail(output.ErrorBody{
		Type:    "confirmation",
		Subtype: "confirmation_required",
		Message: g.Error(),
		Hint:    g.Hint,
		Risk:    string(g.Risk),
		Action:  g.Action,
	}, risk.ExitConfirmationRequired)
	ee, ok := failErr.(output.ExitError)
	if !ok || ee.Code != risk.ExitConfirmationRequired {
		t.Fatalf("exit envelope: %v", failErr)
	}
	var env output.Envelope
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Error == nil || env.Error.Subtype != "confirmation_required" {
		t.Fatalf("%+v", env)
	}
	if env.Error.Risk != string(risk.Write) {
		t.Fatalf("envelope risk=%q", env.Error.Risk)
	}
}

func TestCheckConfirmedWriteWithYes(t *testing.T) {
	if err := risk.CheckConfirmed("workitem +transition", risk.Write, true); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
