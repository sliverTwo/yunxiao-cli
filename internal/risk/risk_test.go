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

func TestIsReadOnlyHTTP(t *testing.T) {
	cases := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/oapi/v1/platform/user", true},
		{"HEAD", "/anything", true},
		{"POST", "/oapi/v1/projex/organizations/o/workitems:search", true},
		{"post", "/oapi/v1/projex/organizations/o/projects:search", true},
		{"POST", "/oapi/v1/projex/organizations/o/workitems:search?x=1", true},
		{"POST", "/oapi/v1/projex/organizations/o/workitems", false},
		{"POST", "/oapi/v1/projex/organizations/o/workitems/abc", false},
		{"PUT", "/oapi/v1/projex/organizations/o/workitems:search", false},
		{"DELETE", "/oapi/v1/x", false},
	}
	for _, tc := range cases {
		got := IsReadOnlyHTTP(tc.method, tc.path)
		if got != tc.want {
			t.Fatalf("%s %s: got %v want %v", tc.method, tc.path, got, tc.want)
		}
	}
}
