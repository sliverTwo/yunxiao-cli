package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MCP-like top-level date aliases accepted on raw api POST …/workitems:search bodies.
// They are normalized into official conditions (BETWEEN / dateTime) on the named fields.
var workitemSearchDateAliases = []struct {
	Key   string // top-level JSON key
	Field string // gmtCreate | gmtModified | finishTime
	Side  string // after | before
}{
	{"createdAfter", "gmtCreate", "after"},
	{"createdBefore", "gmtCreate", "before"},
	{"updatedAfter", "gmtModified", "after"},
	{"updatedBefore", "gmtModified", "before"},
	{"finishTimeAfter", "finishTime", "after"},
	{"finishTimeBefore", "finishTime", "before"},
}

func acceptedWorkitemSearchDateAliasKeys() []string {
	keys := make([]string, 0, len(workitemSearchDateAliases))
	for _, a := range workitemSearchDateAliases {
		keys = append(keys, a.Key)
	}
	return keys
}

func isWorkitemSearchPath(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "workitems:search") || strings.HasSuffix(p, "/workitems:search")
}

// anyToTrimmedString coerces JSON scalar values used as date bounds into strings.
func anyToTrimmedString(v any) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return strings.TrimSpace(t), nil
	case json.Number:
		return strings.TrimSpace(t.String()), nil
	case float64:
		// JSON numbers are uncommon for datetimes; reject to avoid silent bad filters.
		return "", fmt.Errorf("date alias value must be a string (got number)")
	case bool:
		return "", fmt.Errorf("date alias value must be a string (got bool)")
	default:
		return "", fmt.Errorf("date alias value must be a string (got %T)", v)
	}
}

type dateRangeSides struct {
	After  string
	Before string
}

// normalizeWorkitemSearchAPIBody rewrites known MCP-like top-level date aliases
// (createdAfter/Before, updatedAfter/Before, finishTimeAfter/Before) into official
// conditions JSON. Returns whether any alias was normalized.
//
// Unknown top-level keys are left untouched. Known aliases are never silently dropped:
// they are always folded into conditions (merged with any existing conditionGroups).
func normalizeWorkitemSearchAPIBody(body map[string]any) (normalized bool, err error) {
	if body == nil {
		return false, nil
	}
	ranges := map[string]*dateRangeSides{} // field -> sides
	var seen []string
	for _, a := range workitemSearchDateAliases {
		raw, ok := body[a.Key]
		if !ok {
			continue
		}
		s, err := anyToTrimmedString(raw)
		if err != nil {
			return false, fmt.Errorf("%s: %w; accepted top-level date aliases: %s",
				a.Key, err, strings.Join(acceptedWorkitemSearchDateAliasKeys(), ", "))
		}
		delete(body, a.Key)
		seen = append(seen, a.Key)
		r := ranges[a.Field]
		if r == nil {
			r = &dateRangeSides{}
			ranges[a.Field] = r
		}
		if a.Side == "after" {
			r.After = s
		} else {
			r.Before = s
		}
	}
	if len(seen) == 0 {
		return false, nil
	}

	var filters []any
	// Stable field order matching typed search.
	for _, field := range []string{"gmtCreate", "gmtModified", "finishTime"} {
		r := ranges[field]
		if r == nil {
			continue
		}
		filters = appendDateRangeFilter(filters, field, r.After, r.Before)
	}
	if err := mergeConditionFilters(body, filters); err != nil {
		return false, err
	}
	return true, nil
}

// mergeConditionFilters appends filters into body["conditions"] as a JSON string
// with conditionGroups. Existing conditions (string or object) are preserved and
// extended: new filters are appended to the first group when present, else a new
// group is added.
func mergeConditionFilters(body map[string]any, filters []any) error {
	if len(filters) == 0 {
		return nil
	}
	conds, err := parseConditionsValue(body["conditions"])
	if err != nil {
		return err
	}
	if conds == nil {
		conds = map[string]any{}
	}
	groups, _ := conds["conditionGroups"].([]any)
	if len(groups) == 0 {
		conds["conditionGroups"] = []any{filters}
	} else {
		first, ok := groups[0].([]any)
		if !ok {
			// Sometimes unmarshaled as []interface{} of maps only — wrap as new group.
			groups = append(groups, filters)
			conds["conditionGroups"] = groups
		} else {
			first = append(first, filters...)
			groups[0] = first
			conds["conditionGroups"] = groups
		}
	}
	raw, err := json.Marshal(conds)
	if err != nil {
		return fmt.Errorf("marshal conditions: %w", err)
	}
	body["conditions"] = string(raw)
	return nil
}

func parseConditionsValue(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, nil
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return nil, fmt.Errorf("invalid conditions JSON string: %w", err)
		}
		return m, nil
	case map[string]any:
		// Copy so we do not mutate caller's nested map unexpectedly beyond merge.
		out := map[string]any{}
		for k, val := range t {
			out[k] = val
		}
		return out, nil
	default:
		return nil, fmt.Errorf("conditions must be a JSON string or object (got %T)", v)
	}
}

// workitemSearchAPIRequestMeta builds meta.request for raw api :search calls,
// including the actual conditions string after normalization.
func workitemSearchAPIRequestMeta(body map[string]any, normalizedAliases []string) map[string]any {
	req := map[string]any{
		"body_keys": mapKeys(body),
	}
	if cond, ok := body["conditions"]; ok {
		req["conditions"] = cond
	}
	if len(normalizedAliases) > 0 {
		req["normalized_aliases"] = normalizedAliases
	}
	return req
}

// extractNormalizedAliasNames returns which known aliases were present before normalize.
// Call before normalizeWorkitemSearchAPIBody if you need the list; otherwise use
// peekWorkitemSearchDateAliases.
func peekWorkitemSearchDateAliases(body map[string]any) []string {
	if body == nil {
		return nil
	}
	var out []string
	for _, a := range workitemSearchDateAliases {
		if _, ok := body[a.Key]; ok {
			out = append(out, a.Key)
		}
	}
	return out
}
