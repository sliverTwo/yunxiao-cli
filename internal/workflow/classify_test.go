package workflow

import "testing"

func TestClassifyTransitionError(t *testing.T) {
	cases := []struct {
		msg  string
		want Outcome
	}{
		{"当前状态:暂不修复不能流转到目标状态:处理中", OutcomeDenied},
		{"不支持修改 status", OutcomeDenied},
		{"cannot transition to target", OutcomeDenied},
		{"字段【计划完成时间】不能为空", OutcomeNeedsFields},
		{"workitem does not contains field : [\"abc\"]", OutcomeNeedsFields},
		{"未启用此字段", OutcomeNeedsFields},
		{"missing required field developer", OutcomeNeedsFields},
		{"something weird exploded", OutcomeOther},
		{"", OutcomeOther},
	}
	for _, tc := range cases {
		if got := ClassifyTransitionError(tc.msg); got != tc.want {
			t.Fatalf("Classify(%q)=%v want %v", tc.msg, got, tc.want)
		}
	}
}

func TestExtractAPIErrorBody(t *testing.T) {
	msg := `yunxiao API PUT https://x -> HTTP 400: {"errorCode":"InvaildData.Failed","errorMessage":"当前状态:A不能流转到目标状态:B","errorMsg":"当前状态:A不能流转到目标状态:B"}`
	got := ExtractAPIErrorBody(msg)
	if got != "当前状态:A不能流转到目标状态:B" {
		t.Fatalf("got %q", got)
	}
}
