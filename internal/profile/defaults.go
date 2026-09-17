package profile

import (
	"strings"
)

// topLevelCreateKeys are system Role / workitem props applied on the create body root
// (not under customFieldValues). Matches OpenAPI create workitem shape.
var topLevelCreateKeys = map[string]bool{
	"trackers":     true,
	"participants": true,
	"verifier":     true,
	"assignedTo":   true,
	"sprint":       true,
}

// ApplyWorkitemDefaults merges pf.WorkitemDefaults[typeID].Fields into body.
//
// Top-level system keys (trackers, participants, verifier, assignedTo, sprint)
// are written on body when missing/empty. Everything else merges into
// body["customFieldValues"] as fieldId→value (create API shape, same as +bug-create).
//
// Never overrides keys already set by the caller (non-empty string, or any
// non-nil already-present value in body or customFieldValues). Skips empty defaults.
// No-op when pf is nil, typeID empty, or no defaults entry.
func ApplyWorkitemDefaults(body map[string]any, typeID string, pf *Profile) {
	if body == nil || pf == nil {
		return
	}
	typeID = strings.TrimSpace(typeID)
	if typeID == "" || pf.WorkitemDefaults == nil {
		return
	}
	defs, ok := pf.WorkitemDefaults[typeID]
	if !ok || len(defs.Fields) == 0 {
		return
	}

	var cf map[string]any
	if existing, ok := body["customFieldValues"]; ok && existing != nil {
		switch m := existing.(type) {
		case map[string]any:
			cf = m
		default:
			// Unexpected type — start fresh map but keep reference via body later.
			cf = map[string]any{}
		}
	}

	for key, field := range defs.Fields {
		key = strings.TrimSpace(key)
		if key == "" || isEmptyDefault(field.Value) {
			continue
		}
		if topLevelCreateKeys[key] {
			if isAlreadySet(body[key]) {
				continue
			}
			body[key] = field.Value
			continue
		}
		if cf == nil {
			cf = map[string]any{}
		}
		if isAlreadySet(cf[key]) {
			continue
		}
		cf[key] = field.Value
	}

	if cf != nil && len(cf) > 0 {
		body["customFieldValues"] = cf
	}
}

func isEmptyDefault(v any) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case []any:
		return len(t) == 0
	case []string:
		return len(t) == 0
	default:
		return false
	}
}

// isAlreadySet reports whether a caller-provided value should be preserved.
// Non-empty strings and any non-nil non-string value count as set.
func isAlreadySet(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) != ""
	default:
		return true
	}
}
