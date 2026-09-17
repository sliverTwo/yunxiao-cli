package zhiyi

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
)

func TestMergeRelationEnrichmentOK(t *testing.T) {
	rec := map[string]any{
		"id":           "rel1",
		"relationType": "ASSOCIATED",
		"resourceId":   "abc",
		"resourceType": "Bug",
	}
	item := map[string]any{
		"id":           "abc",
		"serialNumber": "ZYPT-5424",
		"subject":      "linked bug",
		"categoryId":   "Bug",
		"spaceId":      "space-1",
	}
	out := MergeRelationEnrichment(rec, item, "space-1", nil)
	if out["serial_number"] != "ZYPT-5424" {
		t.Fatalf("serial=%v", out["serial_number"])
	}
	if out["subject"] != "linked bug" {
		t.Fatalf("subject=%v", out["subject"])
	}
	if out["category"] != "Bug" {
		t.Fatalf("category=%v", out["category"])
	}
	u, _ := out["url"].(string)
	if u == "" || out["relationType"] != "ASSOCIATED" {
		t.Fatalf("url=%q keep=%v", u, out["relationType"])
	}
	// original map not required to stay untouched, but resourceId preserved
	if out["resourceId"] != "abc" {
		t.Fatal(out)
	}
}

func TestMergeRelationEnrichmentError(t *testing.T) {
	rec := map[string]any{"resourceId": "x"}
	out := MergeRelationEnrichment(rec, nil, "", errors.New("boom"))
	if out["resolve_error"] != "boom" || out["resourceId"] != "x" {
		t.Fatalf("%v", out)
	}
}

func TestEnrichRelationRecordsConcurrency(t *testing.T) {
	list := []any{
		map[string]any{"resourceId": "a", "resourceType": "Req"},
		map[string]any{"resourceId": "b", "resourceType": "Bug"},
		map[string]any{"resourceId": "fail"},
		map[string]any{"noId": true},
	}
	var calls int32
	fetch := func(id string) (map[string]any, error) {
		atomic.AddInt32(&calls, 1)
		if id == "fail" {
			return nil, fmt.Errorf("not found")
		}
		return map[string]any{
			"id":           id,
			"serialNumber": "ZYPT-" + id,
			"subject":      "sub-" + id,
			"categoryId":   "Req",
			"spaceId":      "sid",
		}, nil
	}
	out := EnrichRelationRecords(list, fetch, "sid", 4).([]any)
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("calls=%d", calls)
	}
	a := out[0].(map[string]any)
	if a["serial_number"] != "ZYPT-a" || a["subject"] != "sub-a" {
		t.Fatalf("%v", a)
	}
	fail := out[2].(map[string]any)
	if fail["resolve_error"] == nil {
		t.Fatalf("expected resolve_error: %v", fail)
	}
	noop := out[3].(map[string]any)
	if _, ok := noop["serial_number"]; ok {
		t.Fatal("should not enrich without resourceId")
	}
}

func TestExtractRelationRecordsWrapped(t *testing.T) {
	data := map[string]any{
		"data": []any{map[string]any{"resourceId": "1"}},
	}
	recs := ExtractRelationRecords(data)
	if len(recs) != 1 || RelationResourceID(recs[0]) != "1" {
		t.Fatalf("%v", recs)
	}
}
