package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RegisterRequest is the DCR body for a public (PKCE) client.
type RegisterRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

// RegisterResponse is the DCR success payload.
type RegisterResponse struct {
	ClientID                string   `json:"client_id"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
}

// Register performs anonymous Dynamic Client Registration.
// redirectURI must use host 127.0.0.1 (never the hostname "localhost").
func Register(ctx context.Context, httpClient *http.Client, registrationEndpoint, clientName, redirectURI string) (*RegisterResponse, error) {
	if err := ValidateLoopbackRedirectURI(redirectURI); err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	body := RegisterRequest{
		ClientName:              clientName,
		RedirectURIs:            []string{redirectURI},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registrationEndpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oauth register: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oauth register: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 400))
	}
	var out RegisterResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("oauth register: decode: %w", err)
	}
	if strings.TrimSpace(out.ClientID) == "" {
		return nil, fmt.Errorf("oauth register: empty client_id")
	}
	return &out, nil
}

// ValidateLoopbackRedirectURI requires http://127.0.0.1:<port>/... — rejects "localhost".
func ValidateLoopbackRedirectURI(uri string) error {
	u := strings.TrimSpace(uri)
	if !strings.HasPrefix(u, "http://127.0.0.1:") && !strings.HasPrefix(u, "http://127.0.0.1/") {
		return fmt.Errorf("oauth: redirect_uri must use host 127.0.0.1 (got %q); do not use localhost", uri)
	}
	if strings.Contains(strings.ToLower(u), "localhost") {
		return fmt.Errorf("oauth: redirect_uri must not contain hostname localhost")
	}
	return nil
}
