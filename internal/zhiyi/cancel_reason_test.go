package zhiyi

import "testing"

func TestFindFieldIDByName(t *testing.T) {
	fields := []any{
		map[string]any{"id": "80", "name": "计划完成时间"},
		map[string]any{"id": "879b8243adcfb0eeceb1e925a8", "name": "取消原因"},
	}
	id := FindFieldIDByName(fields, "取消原因")
	if id != "879b8243adcfb0eeceb1e925a8" {
		t.Fatalf("%q", id)
	}
	if FindFieldIDByName(fields, "不存在") != "" {
		t.Fatal()
	}
	wrapped := map[string]any{"data": fields}
	if FindFieldIDByName(wrapped, "取消原因") != "879b8243adcfb0eeceb1e925a8" {
		t.Fatal("wrapped")
	}
}

func TestMergeCancelReasonIntoBody(t *testing.T) {
	body := map[string]any{"status": "141230"}
	if !MergeCancelReasonIntoBody(body, "fid", "不再需要") {
		t.Fatal("merge")
	}
	if body["fid"] != "不再需要" {
		t.Fatalf("%v", body)
	}
	// do not overwrite
	if MergeCancelReasonIntoBody(body, "fid", "other") {
		t.Fatal("should not overwrite")
	}
	if body["fid"] != "不再需要" {
		t.Fatal()
	}
}

func TestRequiredFieldHint(t *testing.T) {
	h := RequiredFieldHint("取消原因必填")
	if h == "" || !containsAll(h, "--cancel-reason") {
		t.Fatalf("%q", h)
	}
	h2 := RequiredFieldHint(`{"errorMsg":"【验收说明】必填"}`)
	if h2 == "" || !containsAll(h2, "验收说明") {
		t.Fatalf("%q", h2)
	}
	if RequiredFieldHint("ok") != "" {
		t.Fatal()
	}
}

func TestLooksLikeCancelStatus(t *testing.T) {
	for _, s := range []string{"已取消", "Canceled", "cancelled", "141230"} {
		if !LooksLikeCancelStatus(s) {
			t.Fatalf("%s", s)
		}
	}
	if LooksLikeCancelStatus("处理中") {
		t.Fatal()
	}
}

func TestCancelReasonSoftWarning(t *testing.T) {
	w := CancelReasonSoftWarning("已取消", "", "fid-1")
	if w == "" || !containsAll(w, "--cancel-reason") {
		t.Fatalf("%q", w)
	}
	if CancelReasonSoftWarning("已取消", "reason", "fid-1") != "" {
		t.Fatal("flag set")
	}
	if CancelReasonSoftWarning("处理中", "", "fid-1") != "" {
		t.Fatal("not cancel")
	}
}

func TestCancelReasonFieldCache(t *testing.T) {
	ClearCancelReasonFieldCache()
	if _, ok := CachedCancelReasonFieldID("t1"); ok {
		t.Fatal()
	}
	StoreCancelReasonFieldID("t1", "abc")
	id, ok := CachedCancelReasonFieldID("t1")
	if !ok || id != "abc" {
		t.Fatalf("%v %v", id, ok)
	}
	ClearCancelReasonFieldCache()
}

func containsAll(s string, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
