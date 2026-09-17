package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAPIBaseURL = "https://openapi-rdc.aliyuncs.com"
	EnvAccessToken    = "YUNXIAO_ACCESS_TOKEN"
	EnvAPIBaseURL     = "YUNXIAO_API_BASE_URL"
	EnvOrganizationID = "YUNXIAO_ORGANIZATION_ID"
	EnvEdition        = "YUNXIAO_EDITION" // central | region
)

type File struct {
	AccessToken    string `json:"access_token,omitempty"`
	APIBaseURL     string `json:"api_base_url,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	Edition        string `json:"edition,omitempty"`
}

type Resolved struct {
	AccessToken    string
	APIBaseURL     string
	OrganizationID string
	Edition        string
	ConfigPath     string
	// TokenSource is env | profile | config | none
	TokenSource string
}

func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "yunxiao"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "yunxiao"), nil
}

func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.json"), nil
}

func LoadFile() (File, string, error) {
	p, err := Path()
	if err != nil {
		return File{}, "", err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, p, nil
		}
		return File{}, p, err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return File{}, p, err
	}
	return f, p, nil
}

func SaveFile(f File) (string, error) {
	p, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	b = append(b, '\n')
	if err := os.WriteFile(p, b, 0o600); err != nil {
		return "", err
	}
	return p, nil
}

// Resolve loads config and resolves settings without a profile token.
// Token precedence for Resolve / ResolveWithProfileToken:
//  1. YUNXIAO_ACCESS_TOKEN env (highest)
//  2. active profile's access_token (when passed via ResolveWithProfileToken)
//  3. ~/.config/yunxiao/config.json access_token
//  4. none
func Resolve() (Resolved, error) {
	return ResolveWithProfileToken("")
}

// ResolveWithProfileToken is like Resolve but considers an optional profile PAT.
// Pass the active profile's access_token (or empty). Does not import profile
// (avoids config↔profile cycle). Callers load the profile and pass the token.
// TokenSource reports env | profile | config | none.
func ResolveWithProfileToken(profileToken string) (Resolved, error) {
	f, p, err := LoadFile()
	if err != nil {
		return Resolved{}, err
	}
	r := Resolved{
		ConfigPath:     p,
		APIBaseURL:     DefaultAPIBaseURL,
		OrganizationID: f.OrganizationID,
		Edition:        f.Edition,
	}
	if f.APIBaseURL != "" {
		r.APIBaseURL = strings.TrimRight(f.APIBaseURL, "/")
	}
	if v := strings.TrimSpace(os.Getenv(EnvAPIBaseURL)); v != "" {
		r.APIBaseURL = strings.TrimRight(v, "/")
	}
	if v := strings.TrimSpace(os.Getenv(EnvOrganizationID)); v != "" {
		r.OrganizationID = v
	}
	if v := strings.TrimSpace(os.Getenv(EnvEdition)); v != "" {
		r.Edition = v
	}
	if t := strings.TrimSpace(os.Getenv(EnvAccessToken)); t != "" {
		r.AccessToken = t
		r.TokenSource = "env"
	} else if t := strings.TrimSpace(profileToken); t != "" {
		r.AccessToken = t
		r.TokenSource = "profile"
	} else if f.AccessToken != "" {
		r.AccessToken = f.AccessToken
		r.TokenSource = "config"
	} else {
		r.TokenSource = "none"
	}
	if r.Edition == "" {
		if strings.Contains(r.APIBaseURL, "openapi-rdc.aliyuncs.com") {
			r.Edition = "central"
		} else {
			r.Edition = "region"
		}
	}
	return r, nil
}

func MaskToken(t string) string {
	if t == "" {
		return ""
	}
	if len(t) <= 8 {
		return "****"
	}
	return t[:4] + "****" + t[len(t)-4:]
}
