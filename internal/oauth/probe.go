package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AuthHeaderStrategy is how the access token is sent to OpenAPI.
type AuthHeaderStrategy string

const (
	HeaderYunxiaoToken AuthHeaderStrategy = "x-yunxiao-token"
	HeaderBearer       AuthHeaderStrategy = "authorization-bearer"
)

// DefaultOpenAPIHeader is the preferred header after a successful live probe.
// Updated when ProbeWhoami confirms a working strategy (O1 hard gate).
// Prefer x-yunxiao-token first per peer review; fall back to Bearer.
var DefaultOpenAPIHeader = HeaderYunxiaoToken

// ProbeResult documents which header worked against /oapi/v1/platform/user.
type ProbeResult struct {
	Strategy   AuthHeaderStrategy `json:"strategy"`
	OK         bool               `json:"ok"`
	Status     int                `json:"status,omitempty"`
	UserID     string             `json:"user_id,omitempty"`
	UserName   string             `json:"user_name,omitempty"`
	LastOrg    string             `json:"last_organization,omitempty"`
	Error      string             `json:"error,omitempty"`
	TriedOrder []string           `json:"tried_order"`
}

// ProbeWhoami tries x-yunxiao-token first, then Authorization: Bearer.
// Returns the first successful strategy, or an error if both fail (O1 hard gate).
func ProbeWhoami(ctx context.Context, httpClient *http.Client, apiBase, accessToken string) (*ProbeResult, error) {
	base := strings.TrimRight(strings.TrimSpace(apiBase), "/")
	if base == "" || accessToken == "" {
		return nil, fmt.Errorf("oauth probe: api_base and access_token required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	url := base + "/oapi/v1/platform/user"
	order := []AuthHeaderStrategy{HeaderYunxiaoToken, HeaderBearer}
	out := &ProbeResult{TriedOrder: []string{string(HeaderYunxiaoToken), string(HeaderBearer)}}
	var lastErr error
	for _, strat := range order {
		status, user, err := probeOnce(ctx, httpClient, url, accessToken, strat)
		if err == nil && status >= 200 && status < 300 {
			out.OK = true
			out.Strategy = strat
			out.Status = status
			if id, ok := user["id"].(string); ok {
				out.UserID = id
			}
			if n, ok := user["name"].(string); ok {
				out.UserName = n
			}
			if o, ok := user["lastOrganization"].(string); ok {
				out.LastOrg = o
			}
			DefaultOpenAPIHeader = strat
			return out, nil
		}
		lastErr = err
		if err == nil {
			lastErr = fmt.Errorf("HTTP %d", status)
		}
		out.Status = status
		out.Error = lastErr.Error()
	}
	return out, fmt.Errorf("oauth probe: both x-yunxiao-token and Authorization Bearer failed for OpenAPI whoami (last: %v); STOP — do not ship O2 CLI login", lastErr)
}

func probeOnce(ctx context.Context, httpClient *http.Client, url, token string, strat AuthHeaderStrategy) (int, map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	switch strat {
	case HeaderYunxiaoToken:
		req.Header.Set("x-yunxiao-token", token)
	case HeaderBearer:
		req.Header.Set("Authorization", "Bearer "+token)
	default:
		return 0, nil, fmt.Errorf("unknown strategy %s", strat)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var user map[string]any
	if err := json.Unmarshal(raw, &user); err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, user, nil
}
