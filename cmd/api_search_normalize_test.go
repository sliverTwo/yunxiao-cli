package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeWorkitemSearchAPIBodyAliases(t *testing.T) {
	body := map[string]any{
		"category":         "Req",
		"createdAfter":     "2026-09-01 00:00:00",
		"createdBefore":    "2026-09-07 23:59:59",
		"updatedAfter":     "2026-08-01 00:00:00",
		"finishTimeBefore": "2026-09-30 23:59:59",
		"spaceId":          "space-1",
	}
	aliases := peekWorkitemSearchDateAliases(body)
	if len(aliases) != 4 {
		t.Fatalf("aliases=%v", aliases)
	}
	ok, err := normalizeWorkitemSearchAPIBody(body)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected normalized")
	}
	for _, k := range []string{"createdAfter", "createdBefore", "updatedAfter", "finishTimeBefore"} {
		if _, still := body[k]; still {
			t.Fatalf("alias %s still present", k)
		}
	}
	condStr, _ := body["conditions"].(string)
	if condStr == "" {
		t.Fatal("conditions missing")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(condStr), &parsed); err != nil {
		t.Fatal(err)
	}
	group := parsed["conditionGroups"].([]any)[0].([]any)
	fields := map[string]bool{}
	for _, item := range group {
		m := item.(map[string]any)
		fields[m["fieldIdentifier"].(string)] = true
		if m["operator"] != "BETWEEN" || m["className"] != "dateTime" {
			t.Fatalf("shape=%v", m)
		}
	}
	for _, f := range []string{"gmtCreate", "gmtModified", "finishTime"} {
		if !fields[f] {
			t.Fatalf("missing field %s in %s", f, condStr)
		}
	}
	req := workitemSearchAPIRequestMeta(body, aliases)
	if req["conditions"] != condStr {
		t.Fatalf("request.conditions mismatch")
	}
	na, _ := req["normalized_aliases"].([]string)
	if len(na) != 4 {
		t.Fatalf("normalized_aliases=%v", na)
	}
}

func TestNormalizeWorkitemSearchAPIBodyMergesExistingConditions(t *testing.T) {
	existing := `{"conditionGroups":[[{"className":"status","fieldIdentifier":"status","format":"list","operator":"CONTAINS","value":["100005"]}]]}`
	body := map[string]any{
		"category":     "Bug",
		"conditions":   existing,
		"createdAfter": "2026-09-01 00:00:00",
	}
	ok, err := normalizeWorkitemSearchAPIBody(body)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body["conditions"].(string)), &parsed); err != nil {
		t.Fatal(err)
	}
	group := parsed["conditionGroups"].([]any)[0].([]any)
	if len(group) != 2 {
		t.Fatalf("group len=%d raw=%v", len(group), body["conditions"])
	}
	foundStatus, foundCreate := false, false
	for _, item := range group {
		m := item.(map[string]any)
		switch m["fieldIdentifier"] {
		case "status":
			foundStatus = true
		case "gmtCreate":
			foundCreate = true
		}
	}
	if !foundStatus || !foundCreate {
		t.Fatalf("status=%v create=%v", foundStatus, foundCreate)
	}
}

func TestNormalizeWorkitemSearchAPIBodyNoop(t *testing.T) {
	body := map[string]any{"category": "Req", "spaceId": "s"}
	ok, err := normalizeWorkitemSearchAPIBody(body)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if _, has := body["conditions"]; has {
		t.Fatal("should not invent conditions")
	}
}

func TestNormalizeWorkitemSearchAPIBodyRejectsEmptyString(t *testing.T) {
	for _, alias := range acceptedWorkitemSearchDateAliasKeys() {
		t.Run(alias, func(t *testing.T) {
			body := map[string]any{alias: "   "}
			_, err := normalizeWorkitemSearchAPIBody(body)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), alias) || !strings.Contains(err.Error(), "must not be empty") {
				t.Fatalf("err=%v", err)
			}
			if _, still := body[alias]; !still {
				t.Fatalf("alias %s should remain when normalization fails", alias)
			}
		})
	}
}

func TestNormalizeWorkitemSearchAPIBodyRejectsNonString(t *testing.T) {
	body := map[string]any{"createdAfter": 123}
	_, err := normalizeWorkitemSearchAPIBody(body)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "createdAfter") || !strings.Contains(err.Error(), "accepted top-level date aliases") {
		t.Fatalf("err=%v", err)
	}
}

func TestIsWorkitemSearchPath(t *testing.T) {
	if !isWorkitemSearchPath("/oapi/v1/projex/organizations/x/workitems:search") {
		t.Fatal()
	}
	if isWorkitemSearchPath("/oapi/v1/projex/organizations/x/projects:search") {
		t.Fatal()
	}
}
