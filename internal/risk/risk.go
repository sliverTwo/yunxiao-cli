package risk

import "fmt"

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
