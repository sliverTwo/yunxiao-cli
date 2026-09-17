package profile

import (
	"reflect"
	"testing"
)

func testDefaultsProfile() *Profile {
	return &Profile{
		Name: "test",
		WorkitemDefaults: map[string]WorkitemTypeDefaults{
			"req-type": {
				Name:     "产品类需求",
				Category: "Req",
				Fields: map[string]WorkitemDefaultField{
					"priority": {
						Value:     "prio-med",
						Display:   "中",
						FieldName: "优先级",
					},
					"6553533e855d4958ae9f6242b0": {
						Value:     "65cc38447d3e6485f086914e",
						FieldName: "测试负责人",
					},
					"e4ab47e4c00220b846f6b9f8d4": {
						Value:     "65cc3844ff6f3c5a80eb4310",
						FieldName: "验收负责人",
					},
					"trackers": {
						Value:     []any{"uid-tracker"},
						FieldName: "抄送",
					},
					"participants": {
						Value:     []any{"uid-part"},
						FieldName: "参与者",
					},
					"assignedTo": {
						Value: "uid-default-assignee",
					},
					"sprint": {
						Value: "sprint-default",
					},
					"emptyStr": {
						Value: "",
					},
					"emptyArr": {
						Value: []any{},
					},
				},
			},
			"bug-type": {
				Name: "缺陷",
				Fields: map[string]WorkitemDefaultField{
					"priority": {
						Value: "prio-high",
					},
					"seriousLevel": {
						Value: "sev-normal",
					},
					"trackers": {
						Value: []any{"uid-cc"},
					},
				},
			},
		},
	}
}

func TestApplyWorkitemDefaults_MergeCustomAndTopLevel(t *testing.T) {
	pf := testDefaultsProfile()
	body := map[string]any{
		"spaceId":        "s1",
		"workitemTypeId": "req-type",
		"subject":        "t",
		"assignedTo":     "uid-user", // already set — must keep
	}
	ApplyWorkitemDefaults(body, "req-type", pf)

	cf, ok := body["customFieldValues"].(map[string]any)
	if !ok {
		t.Fatalf("customFieldValues missing: %+v", body)
	}
	if cf["priority"] != "prio-med" {
		t.Fatalf("priority: %v", cf["priority"])
	}
	if cf["6553533e855d4958ae9f6242b0"] != "65cc38447d3e6485f086914e" {
		t.Fatalf("测试负责人: %v", cf["6553533e855d4958ae9f6242b0"])
	}
	if cf["e4ab47e4c00220b846f6b9f8d4"] != "65cc3844ff6f3c5a80eb4310" {
		t.Fatalf("验收负责人: %v", cf["e4ab47e4c00220b846f6b9f8d4"])
	}
	if _, ok := cf["emptyStr"]; ok {
		t.Fatal("empty string default should be skipped")
	}
	if _, ok := cf["emptyArr"]; ok {
		t.Fatal("empty array default should be skipped")
	}
	if _, ok := cf["trackers"]; ok {
		t.Fatal("trackers must be top-level, not custom")
	}

	trackers, ok := body["trackers"].([]any)
	if !ok || len(trackers) != 1 || trackers[0] != "uid-tracker" {
		t.Fatalf("trackers top-level: %v", body["trackers"])
	}
	parts, ok := body["participants"].([]any)
	if !ok || len(parts) != 1 || parts[0] != "uid-part" {
		t.Fatalf("participants: %v", body["participants"])
	}
	if body["assignedTo"] != "uid-user" {
		t.Fatalf("assignedTo must not override: %v", body["assignedTo"])
	}
	if body["sprint"] != "sprint-default" {
		t.Fatalf("sprint default when empty: %v", body["sprint"])
	}
}

func TestApplyWorkitemDefaults_NoOverride(t *testing.T) {
	pf := testDefaultsProfile()
	body := map[string]any{
		"assignedTo": "",
		"trackers":   []any{"explicit-tracker"},
		"customFieldValues": map[string]any{
			"priority":                  "caller-prio",
			"6553533e855d4958ae9f6242b0": "caller-tester",
		},
	}
	ApplyWorkitemDefaults(body, "req-type", pf)

	if body["assignedTo"] != "uid-default-assignee" {
		t.Fatalf("empty assignedTo should get default: %v", body["assignedTo"])
	}
	trackers := body["trackers"].([]any)
	if len(trackers) != 1 || trackers[0] != "explicit-tracker" {
		t.Fatalf("trackers must keep caller array: %v", trackers)
	}
	cf := body["customFieldValues"].(map[string]any)
	if cf["priority"] != "caller-prio" {
		t.Fatalf("priority override: %v", cf["priority"])
	}
	if cf["6553533e855d4958ae9f6242b0"] != "caller-tester" {
		t.Fatalf("测试负责人 override: %v", cf["6553533e855d4958ae9f6242b0"])
	}
	// 验收负责人 not set by caller — should fill
	if cf["e4ab47e4c00220b846f6b9f8d4"] != "65cc3844ff6f3c5a80eb4310" {
		t.Fatalf("验收负责人 missing: %v", cf["e4ab47e4c00220b846f6b9f8d4"])
	}
}

func TestApplyWorkitemDefaults_BugTrackersAndNoOverridePriority(t *testing.T) {
	pf := testDefaultsProfile()
	body := map[string]any{
		"assignedTo": "u1",
		"sprint":     "s1",
		"customFieldValues": map[string]any{
			"priority":     "from-build",
			"seriousLevel": "from-build",
		},
	}
	ApplyWorkitemDefaults(body, "bug-type", pf)

	cf := body["customFieldValues"].(map[string]any)
	if cf["priority"] != "from-build" || cf["seriousLevel"] != "from-build" {
		t.Fatalf("must not override BuildCreateBugArgs cf: %+v", cf)
	}
	trackers, ok := body["trackers"].([]any)
	if !ok || !reflect.DeepEqual(trackers, []any{"uid-cc"}) {
		t.Fatalf("trackers from defaults: %v", body["trackers"])
	}
}

func TestApplyWorkitemDefaults_NilSafe(t *testing.T) {
	ApplyWorkitemDefaults(nil, "x", testDefaultsProfile())
	body := map[string]any{"a": 1}
	ApplyWorkitemDefaults(body, "x", nil)
	ApplyWorkitemDefaults(body, "", testDefaultsProfile())
	ApplyWorkitemDefaults(body, "unknown", testDefaultsProfile())
	if len(body) != 1 {
		t.Fatalf("nil-safe should no-op: %+v", body)
	}
}
