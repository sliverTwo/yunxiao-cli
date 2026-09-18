package risk

import (
	"fmt"
	"strings"
)

// Level classifies command risk for agents and humans.
type Level string

const (
	Read          Level = "read"
	Write         Level = "write"
	HighRiskWrite Level = "high-risk-write"
)

const ExitConfirmationRequired = 10

// GateResult is returned when a write lacks --yes (high-risk, or multi-step write).
type GateResult struct {
	Action string
	Risk   Level
	Hint   string
}

func (g GateResult) Error() string {
	return fmt.Sprintf("%s requires confirmation (%s)", g.Action, g.Risk)
}

// CheckConfirmed returns a GateResult when yes is false.
func CheckConfirmed(action string, level Level, yes bool) error {
	if yes {
		return nil
	}
	return GateResult{
		Action: action,
		Risk:   level,
		Hint:   "add --yes to confirm",
	}
}

// CheckHighRisk returns a GateResult when yes is false for high-risk-write.
func CheckHighRisk(action string, yes bool) error {
	return CheckConfirmed(action, HighRiskWrite, yes)
}

// IsConfirmation reports whether err is a confirmation gate.
func IsConfirmation(err error) bool {
	_, ok := err.(GateResult)
	return ok
}

// IsReadOnlyHTTP reports whether method+path is a genuine read that must not
// require --yes (including read-only POSTs such as .../workitems:search).
// Unknown mutating methods stay gated by the caller.
func IsReadOnlyHTTP(method, path string) bool {
	m := strings.ToUpper(method)
	if m == "GET" || m == "HEAD" || m == "OPTIONS" {
		return true
	}
	if m != "POST" {
		return false
	}
	return isReadOnlyPostPath(path)
}

func isReadOnlyPostPath(path string) bool {
	// Typed search commands and schema registry mark these as Risk: read.
	// Match ":search" action suffix used across Projex/Platform/AppStack/Testhub.
	base := path
	if i := strings.IndexByte(base, '?'); i >= 0 {
		base = base[:i]
	}
	return strings.Contains(base, ":search")
}
