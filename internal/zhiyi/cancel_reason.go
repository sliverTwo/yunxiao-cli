package zhiyi

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

const CancelReasonFieldName = "取消原因"

// Known cancel status identifiers (display name / English / common id).
var cancelStatusTokens = map[string]bool{
	"已取消":       true,
	"canceled":   true,
	"cancelled":  true,
	"141230":     true,
}

var (
	cancelReasonFieldCache sync.Map // typeID -> fieldID (empty string = miss)
	requiredFieldNameRe    = regexp.MustCompile(`【([^】]+)】`)
	requiredPrefixRe       = regexp.MustCompile(`([^【\]\s:：,，]{2,32})必填`)
)

// LooksLikeCancelStatus reports whether status id or display name looks like cancelled.
func LooksLikeCancelStatus(status string) bool {
	s := strings.TrimSpace(status)
	if s == "" {
		return false
	}
	if cancelStatusTokens[s] {
		return true
	}
	return cancelStatusTokens[strings.ToLower(s)]
}

// FindFieldIDByName searches a fields API payload for a field whose name matches
// (exact, trimmed). Returns field id or empty.
func FindFieldIDByName(fields any, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	list := flattenFields(fields)
	for _, f := range list {
		fname := strings.TrimSpace(fmt.Sprint(f["name"]))
		if fname == name {
			id := strings.TrimSpace(fmt.Sprint(f["id"]))
			if id != "" && id != "<nil>" {
				return id
			}
		}
	}
	return ""
}

func flattenFields(fields any) []map[string]any {
	switch v := fields.(type) {
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, it := range v {
			if m, ok := it.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return v
	case map[string]any:
		for _, key := range []string{"fields", "data", "items", "list"} {
			if inner, ok := v[key]; ok {
				return flattenFields(inner)
			}
		}
	}
	return nil
}

// CachedCancelReasonFieldID returns cached field id for typeID, or ok=false.
func CachedCancelReasonFieldID(typeID string) (fieldID string, ok bool) {
	typeID = strings.TrimSpace(typeID)
	if typeID == "" {
		return "", false
	}
	v, hit := cancelReasonFieldCache.Load(typeID)
	if !hit {
		return "", false
	}
	return v.(string), true
}

// StoreCancelReasonFieldID caches field id (may be empty for "not found").
func StoreCancelReasonFieldID(typeID, fieldID string) {
	typeID = strings.TrimSpace(typeID)
	if typeID == "" {
		return
	}
	cancelReasonFieldCache.Store(typeID, strings.TrimSpace(fieldID))
}

// ClearCancelReasonFieldCache is for tests.
func ClearCancelReasonFieldCache() {
	cancelReasonFieldCache = sync.Map{}
}

// MergeCancelReasonIntoBody sets body[fieldID]=text when both non-empty.
// Does not overwrite an existing non-empty value for that field id.
func MergeCancelReasonIntoBody(body map[string]any, fieldID, text string) bool {
	if body == nil {
		return false
	}
	fieldID = strings.TrimSpace(fieldID)
	text = strings.TrimSpace(text)
	if fieldID == "" || text == "" {
		return false
	}
	if existing, ok := body[fieldID]; ok && existing != nil {
		s := strings.TrimSpace(fmt.Sprint(existing))
		if s != "" && s != "<nil>" {
			return false
		}
	}
	body[fieldID] = text
	return true
}

// RequiredFieldHint builds a CLI hint when an API error mentions 必填.
// Prefer --cancel-reason when the required field looks like 取消原因.
func RequiredFieldHint(apiMessage string) string {
	msg := strings.TrimSpace(apiMessage)
	if msg == "" || !strings.Contains(msg, "必填") {
		return ""
	}
	fieldName := parseRequiredFieldName(msg)
	if fieldName == CancelReasonFieldName || strings.Contains(msg, CancelReasonFieldName) {
		return "status change requires 取消原因; pass --cancel-reason <text> or --custom-fields '{\"<fieldId>\":\"…\"}' (discover id via: workitem fields)"
	}
	if fieldName != "" {
		return fmt.Sprintf("required field %q; pass --cancel-reason if it is 取消原因, else --custom-fields '{\"<fieldId>\":\"…\"}' (workitem fields --space-id … --type-id …)", fieldName)
	}
	return "API reports a required field (必填); try --cancel-reason <text> for 取消原因, or --custom-fields with the field id from: workitem fields"
}

func parseRequiredFieldName(msg string) string {
	if m := requiredFieldNameRe.FindStringSubmatch(msg); len(m) > 1 {
		parts := strings.Split(m[1], ",")
		return strings.TrimSpace(parts[0])
	}
	if m := requiredPrefixRe.FindStringSubmatch(msg); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// CancelReasonSoftWarning returns a soft warning when targeting cancel without --cancel-reason
// while the type has a 取消原因 field.
func CancelReasonSoftWarning(status, cancelReasonFlag, fieldID string) string {
	if strings.TrimSpace(cancelReasonFlag) != "" {
		return ""
	}
	if !LooksLikeCancelStatus(status) {
		return ""
	}
	if strings.TrimSpace(fieldID) == "" {
		return ""
	}
	return fmt.Sprintf("status looks like cancelled (%s) and type has field 取消原因 (id %s); consider --cancel-reason to avoid 必填 errors", status, fieldID)
}
