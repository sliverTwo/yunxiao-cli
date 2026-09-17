package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// CallbackResult is the single-shot authorization callback payload.
type CallbackResult struct {
	Code  string
	State string
	Error string
	Desc  string
}

// ListenResult holds the listener address and a channel for the one callback.
type ListenResult struct {
	Listener   net.Listener
	RedirectURI string // http://127.0.0.1:<port>/callback
	Port       int
	ResultCh   <-chan CallbackResult
	Server     *http.Server
}

// StartLoopbackListener binds 127.0.0.1 only (never 0.0.0.0 / ::) and serves a
// single /callback; after one valid or error response the server shuts down.
func StartLoopbackListener(ctx context.Context) (*ListenResult, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("oauth callback listen: %w", err)
	}
	tcp, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		_ = ln.Close()
		return nil, fmt.Errorf("oauth callback: unexpected listener addr type")
	}
	port := tcp.Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	ch := make(chan CallbackResult, 1)
	mux := http.NewServeMux()
	var srv *http.Server
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Reject non-loopback remote (belt-and-suspenders; we only bind 127.0.0.1).
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		q := r.URL.Query()
		res := CallbackResult{
			Code:  q.Get("code"),
			State: q.Get("state"),
			Error: q.Get("error"),
			Desc:  q.Get("error_description"),
		}
		select {
		case ch <- res:
		default:
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if res.Error != "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`<!doctype html><html><body><h1>Authorization failed</h1><p>You can close this window and return to the CLI.</p></body></html>`))
		} else {
			_, _ = w.Write([]byte(`<!doctype html><html><body><h1>Authorization complete</h1><p>You can close this window and return to the CLI.</p></body></html>`))
		}
		// Single-shot: close listener after this response.
		go func() {
			time.Sleep(100 * time.Millisecond)
			if srv != nil {
				_ = srv.Shutdown(context.Background())
			}
		}()
	})
	srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		_ = srv.Serve(ln)
		close(ch)
	}()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	return &ListenResult{
		Listener:    ln,
		RedirectURI: redirectURI,
		Port:        port,
		ResultCh:    ch,
		Server:      srv,
	}, nil
}

// WaitCallback waits for the single callback or ctx/timeout.
func WaitCallback(ctx context.Context, lr *ListenResult, expectState string) (CallbackResult, error) {
	select {
	case <-ctx.Done():
		if lr.Server != nil {
			_ = lr.Server.Shutdown(context.Background())
		}
		return CallbackResult{}, fmt.Errorf("oauth callback: %w", ctx.Err())
	case res, ok := <-lr.ResultCh:
		if !ok {
			return CallbackResult{}, fmt.Errorf("oauth callback: listener closed without result")
		}
		if res.Error != "" {
			return res, fmt.Errorf("oauth authorization error: %s (%s)", res.Error, res.Desc)
		}
		if res.Code == "" {
			return res, fmt.Errorf("oauth callback: missing code")
		}
		if expectState != "" && res.State != expectState {
			return res, fmt.Errorf("oauth callback: state mismatch")
		}
		return res, nil
	}
}

// IsLoopbackAddr reports whether host is a loopback IP literal (not "localhost").
func IsLoopbackAddr(host string) bool {
	h := strings.TrimSpace(host)
	if h == "localhost" {
		return false
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}
