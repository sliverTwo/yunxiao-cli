package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewPKCE_S256(t *testing.T) {
	p, err := NewPKCE()
	if err != nil {
		t.Fatal(err)
	}
	if p.Method != "S256" {
		t.Fatalf("method=%s", p.Method)
	}
	if len(p.Verifier) < 43 {
		t.Fatalf("verifier too short: %d", len(p.Verifier))
	}
	sum := sha256.Sum256([]byte(p.Verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if p.Challenge != want {
		t.Fatalf("challenge mismatch")
	}
}

func TestValidateLoopbackRedirectURI(t *testing.T) {
	if err := ValidateLoopbackRedirectURI("http://127.0.0.1:9/callback"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLoopbackRedirectURI("http://localhost:9/callback"); err == nil {
		t.Fatal("expected reject localhost")
	}
	if err := ValidateLoopbackRedirectURI("http://0.0.0.0:9/callback"); err == nil {
		t.Fatal("expected reject 0.0.0.0")
	}
}

func TestDiscoverRegisterExchangeRefresh(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Metadata{
			Issuer:                "https://example.test",
			AuthorizationEndpoint: "https://account.example.test/authorize",
			TokenEndpoint:         "http://" + r.Host + "/v1/oauth2/token",
			RegistrationEndpoint:  "http://" + r.Host + "/v1/oauth2/register",
		})
	})
	mux.HandleFunc("/v1/oauth2/register", func(w http.ResponseWriter, r *http.Request) {
		var body RegisterRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.TokenEndpointAuthMethod != "none" {
			t.Errorf("auth_method=%s", body.TokenEndpointAuthMethod)
		}
		if len(body.RedirectURIs) != 1 || !strings.HasPrefix(body.RedirectURIs[0], "http://127.0.0.1:") {
			t.Errorf("redirect=%v", body.RedirectURIs)
		}
		_ = json.NewEncoder(w).Encode(RegisterResponse{ClientID: "cid-test"})
	})
	mux.HandleFunc("/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(b))
		switch v.Get("grant_type") {
		case "authorization_code":
			if v.Get("code_verifier") == "" || v.Get("client_id") == "" {
				http.Error(w, "bad", 400)
				return
			}
			_ = json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "oat-access", RefreshToken: "ort-refresh", TokenType: "Bearer", ExpiresIn: 3600,
			})
		case "refresh_token":
			if v.Get("refresh_token") != "ort-refresh" {
				http.Error(w, "bad refresh", 400)
				return
			}
			_ = json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "oat-new", RefreshToken: "ort-new", TokenType: "Bearer", ExpiresIn: 3600,
			})
		default:
			http.Error(w, "bad grant", 400)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx := context.Background()
	meta, err := Discover(ctx, srv.Client(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if meta.AuthorizationEndpoint == "" {
		t.Fatal("missing auth endpoint")
	}
	reg, err := Register(ctx, srv.Client(), meta.RegistrationEndpoint, "yunxiao-cli-test", "http://127.0.0.1:12345/callback")
	if err != nil {
		t.Fatal(err)
	}
	if reg.ClientID != "cid-test" {
		t.Fatalf("client_id=%s", reg.ClientID)
	}
	pkce, _ := NewPKCE()
	state, _ := RandomState()
	authURL, err := AuthURL(AuthURLParams{
		AuthorizationEndpoint: meta.AuthorizationEndpoint,
		ClientID:              reg.ClientID,
		RedirectURI:           "http://127.0.0.1:12345/callback",
		State:                 state,
		CodeChallenge:         pkce.Challenge,
		CodeChallengeMethod:   "S256",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(authURL, "code_challenge_method=S256") {
		t.Fatalf("url=%s", authURL)
	}
	tok, err := Exchange(ctx, srv.Client(), meta.TokenEndpoint, reg.ClientID, "code1", "http://127.0.0.1:12345/callback", pkce.Verifier)
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "oat-access" {
		t.Fatalf("%+v", tok)
	}
	ref, err := Refresh(ctx, srv.Client(), meta.TokenEndpoint, reg.ClientID, tok.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if ref.AccessToken != "oat-new" {
		t.Fatalf("%+v", ref)
	}
}

func TestProbeWhoami_PrefersYunxiaoToken(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "" {
			seen = append(seen, "x-yunxiao-token")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "u1", "name": "n", "lastOrganization": "o1"})
			return
		}
		if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			seen = append(seen, "bearer")
			http.Error(w, "nope", 401)
			return
		}
		http.Error(w, "no auth", 401)
	}))
	defer srv.Close()
	res, err := ProbeWhoami(context.Background(), srv.Client(), srv.URL, "oat-fake")
	if err != nil {
		t.Fatal(err)
	}
	if res.Strategy != HeaderYunxiaoToken || !res.OK {
		t.Fatalf("%+v", res)
	}
	if len(seen) != 1 || seen[0] != "x-yunxiao-token" {
		t.Fatalf("seen=%v", seen)
	}
}

func TestProbeWhoami_FallsBackToBearer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "" {
			http.Error(w, "nope", 401)
			return
		}
		if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "u2", "name": "n2"})
			return
		}
		http.Error(w, "no auth", 401)
	}))
	defer srv.Close()
	res, err := ProbeWhoami(context.Background(), srv.Client(), srv.URL, "oat-fake")
	if err != nil {
		t.Fatal(err)
	}
	if res.Strategy != HeaderBearer {
		t.Fatalf("%+v", res)
	}
}

func TestProbeWhoami_BothFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", 401)
	}))
	defer srv.Close()
	_, err := ProbeWhoami(context.Background(), srv.Client(), srv.URL, "oat-fake")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "STOP") {
		t.Fatalf("err=%v", err)
	}
}

func TestStartLoopbackListener_SingleShot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lr, err := StartLoopbackListener(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(lr.RedirectURI, "http://127.0.0.1:") {
		t.Fatalf("uri=%s", lr.RedirectURI)
	}
	if strings.Contains(lr.RedirectURI, "localhost") {
		t.Fatal("localhost forbidden")
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		resp, err := http.Get(lr.RedirectURI + "?code=abc&state=xyz")
		if err == nil {
			_, _ = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
		}
	}()
	res, err := WaitCallback(ctx, lr, "xyz")
	if err != nil {
		t.Fatal(err)
	}
	if res.Code != "abc" {
		t.Fatalf("%+v", res)
	}
}

func TestExpiresAt(t *testing.T) {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	got := ExpiresAt(3600, now)
	if !got.Equal(now.Add(time.Hour)) {
		t.Fatalf("%v", got)
	}
}
