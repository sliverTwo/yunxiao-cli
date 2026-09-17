package risk

import "testing"

func TestCheckHighRiskRequiresYes(t *testing.T) {
	err := CheckHighRisk("codeup mrs create", false)
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	if !IsConfirmation(err) {
		t.Fatalf("expected GateResult, got %T", err)
	}
	g := err.(GateResult)
	if g.Risk != HighRiskWrite {
		t.Fatalf("risk=%s", g.Risk)
	}
	if g.Hint == "" {
		t.Fatal("hint empty")
	}
}

func TestCheckHighRiskWithYes(t *testing.T) {
	if err := CheckHighRisk("codeup mrs create", true); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestExitCodeConstant(t *testing.T) {
	if ExitConfirmationRequired != 10 {
		t.Fatalf("exit code must be 10, got %d", ExitConfirmationRequired)
	}
}


func TestCheckConfirmedWriteRequiresYes(t *testing.T) {
	err := CheckConfirmed("workitem +bug-transition", Write, false)
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	if !IsConfirmation(err) {
		t.Fatalf("expected GateResult, got %T", err)
	}
	g := err.(GateResult)
	if g.Risk != Write {
		t.Fatalf("risk=%s", g.Risk)
	}
}

func TestCheckConfirmedWriteWithYes(t *testing.T) {
	if err := CheckConfirmed("workitem +bug-transition", Write, true); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
