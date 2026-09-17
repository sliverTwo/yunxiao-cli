package workflow

import "fmt"

// ParseWorkflowResponse extracts workflow metadata + statuses from GET .../workflows JSON.
func ParseWorkflowResponse(raw any) (workflowID, workflowName, defaultStatusID string, statuses []StatusInfo, err error) {
	m, ok := asMap(raw)
	if !ok {
		// Sometimes wrapped as list of workflows — take first.
		if arr, ok := raw.([]any); ok && len(arr) > 0 {
			m, ok = asMap(arr[0])
			if !ok {
				return "", "", "", nil, fmt.Errorf("unexpected workflow response type")
			}
		} else {
			return "", "", "", nil, fmt.Errorf("unexpected workflow response type %T", raw)
		}
	}
	workflowID = asString(m["id"])
	workflowName = asString(m["name"])
	defaultStatusID = asString(m["defaultStatusId"])
	if defaultStatusID == "" {
		defaultStatusID = asString(m["defaultStatusID"])
	}
	rawStatuses, _ := m["statuses"]
	switch list := rawStatuses.(type) {
	case []any:
		for _, it := range list {
			sm, ok := asMap(it)
			if !ok {
				continue
			}
			st := StatusInfo{
				ID:          asString(sm["id"]),
				Name:        asString(sm["name"]),
				DisplayName: asString(sm["displayName"]),
				NameEn:      asString(sm["nameEn"]),
			}
			if st.ID != "" {
				statuses = append(statuses, st)
			}
		}
	case []map[string]any:
		for _, sm := range list {
			st := StatusInfo{
				ID:          asString(sm["id"]),
				Name:        asString(sm["name"]),
				DisplayName: asString(sm["displayName"]),
				NameEn:      asString(sm["nameEn"]),
			}
			if st.ID != "" {
				statuses = append(statuses, st)
			}
		}
	}
	if len(statuses) == 0 {
		return workflowID, workflowName, defaultStatusID, statuses, fmt.Errorf("workflow has no statuses")
	}
	return workflowID, workflowName, defaultStatusID, statuses, nil
}

// StatusIDs returns ids from StatusInfo list.
func StatusIDs(statuses []StatusInfo) []string {
	out := make([]string, 0, len(statuses))
	for _, s := range statuses {
		if s.ID != "" {
			out = append(out, s.ID)
		}
	}
	return out
}

func asMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%.0f", t)
	case nil:
		return ""
	default:
		s := fmt.Sprint(t)
		if s == "<nil>" {
			return ""
		}
		return s
	}
}
