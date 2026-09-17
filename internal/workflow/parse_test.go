package workflow

import "testing"

func TestParseWorkflowResponse(t *testing.T) {
	raw := map[string]any{
		"id":              "wf1",
		"name":            "缺陷的默认工作流",
		"defaultStatusId": "28",
		"statuses": []any{
			map[string]any{"id": "28", "name": "待确认", "displayName": "待确认", "nameEn": "New"},
			map[string]any{"id": "100010", "name": "处理中", "displayName": "处理中"},
		},
	}
	id, name, def, st, err := ParseWorkflowResponse(raw)
	if err != nil || id != "wf1" || name == "" || def != "28" || len(st) != 2 {
		t.Fatalf("%v %s %s %s %v", err, id, name, def, st)
	}
}
