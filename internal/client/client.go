package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/version"
)

const (
	// maxAttempts is 1 initial try + up to 3 retries for idempotent GETs.
	maxAttempts = 4
	retryBase   = 200 * time.Millisecond
	// maxRetryAfter caps server Retry-After so a bad 86400s header cannot stall the CLI.
	maxRetryAfter = 30 * time.Second
)

// sleepWithContext waits d or until ctx is done. Tests may override to skip delays.
var sleepWithContext = func(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// SetRetrySleepForTest overrides retry backoff sleep. Tests must restore via the returned func.
func SetRetrySleepForTest(fn func(ctx context.Context, d time.Duration) error) (restore func()) {
	prev := sleepWithContext
	sleepWithContext = fn
	return func() { sleepWithContext = prev }
}

type Client struct {
	HTTP      *http.Client
	BaseURL   string
	Token     string
	Edition   string
	OrgID     string
	UserAgent string
}

func New(r config.Resolved) (*Client, error) {
	if r.AccessToken == "" {
		return nil, fmt.Errorf("missing access token: set %s, profile access_token, or run `yunxiao auth login`", config.EnvAccessToken)
	}
	return &Client{
		HTTP:      &http.Client{Timeout: 60 * time.Second},
		BaseURL:   strings.TrimRight(r.APIBaseURL, "/"),
		Token:     r.AccessToken,
		Edition:   r.Edition,
		OrgID:     r.OrganizationID,
		UserAgent: "yunxiao-cli/" + version.Version,
	}, nil
}

func (c *Client) IsRegion() bool {
	return strings.EqualFold(c.Edition, "region")
}

// ResolveOrgID returns configured org or fetches lastOrganization from /platform/user.
func (c *Client) ResolveOrgID(ctx context.Context) (string, error) {
	if c.IsRegion() {
		if c.OrgID != "" {
			return c.OrgID, nil
		}
		return "default", nil
	}
	if c.OrgID != "" && c.OrgID != "default" {
		return c.OrgID, nil
	}
	var user map[string]any
	if err := c.Get(ctx, "/oapi/v1/platform/user", nil, &user); err != nil {
		return "", err
	}
	if v, ok := user["lastOrganization"].(string); ok && v != "" {
		c.OrgID = v
		return v, nil
	}
	return "", fmt.Errorf("organizationId required for central edition; set YUNXIAO_ORGANIZATION_ID or config organization_id")
}

func (c *Client) domainPath(ctx context.Context, domain, suffix string) (string, error) {
	if !strings.HasPrefix(suffix, "/") {
		suffix = "/" + suffix
	}
	if c.IsRegion() {
		return "/oapi/v1/" + domain + suffix, nil
	}
	org, err := c.ResolveOrgID(ctx)
	if err != nil {
		return "", err
	}
	return "/oapi/v1/" + domain + "/organizations/" + org + suffix, nil
}

func (c *Client) PlatformPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "platform", suffix)
}
func (c *Client) ProjexPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "projex", suffix)
}
func (c *Client) CodeupPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "codeup", suffix)
}
func (c *Client) FlowPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "flow", suffix)
}
func (c *Client) PackagesPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "packages", suffix)
}
func (c *Client) AppstackPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "appstack", suffix)
}
func (c *Client) TesthubPath(ctx context.Context, suffix string) (string, error) {
	return c.domainPath(ctx, "testhub", suffix)
}

type RequestPreview struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body,omitempty"`
}

func (c *Client) Preview(method, path string, query map[string]string, body any) RequestPreview {
	return RequestPreview{
		Method: method,
		URL:    c.buildURL(path, query),
		Headers: map[string]string{
			"Accept":          "application/json",
			"Content-Type":    "application/json",
			"x-yunxiao-token": "(redacted)",
			"User-Agent":      c.UserAgent,
		},
		Body: body,
	}
}

func (c *Client) buildURL(path string, query map[string]string) string {
	u := c.BaseURL + path
	if len(query) == 0 {
		return u
	}
	q := url.Values{}
	for k, v := range query {
		if v != "" {
			q.Set(k, v)
		}
	}
	if enc := q.Encode(); enc != "" {
		u += "?" + enc
	}
	return u
}

func (c *Client) Get(ctx context.Context, path string, query map[string]string, out any) error {
	_, err := c.Do(ctx, "GET", path, query, nil, out)
	return err
}

func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	_, err := c.Do(ctx, "POST", path, nil, body, out)
	return err
}

func (c *Client) PostWithQuery(ctx context.Context, path string, query map[string]string, body any, out any) error {
	_, err := c.Do(ctx, "POST", path, query, body, out)
	return err
}

func (c *Client) Put(ctx context.Context, path string, body any, out any) error {
	_, err := c.Do(ctx, "PUT", path, nil, body, out)
	return err
}

func (c *Client) PutWithQuery(ctx context.Context, path string, query map[string]string, body any, out any) error {
	_, err := c.Do(ctx, "PUT", path, query, body, out)
	return err
}

func (c *Client) Delete(ctx context.Context, path string, query map[string]string, out any) error {
	_, err := c.Do(ctx, "DELETE", path, query, nil, out)
	return err
}

// DeleteJSON sends DELETE with a JSON body (needed by some Yunxiao endpoints).
func (c *Client) DeleteJSON(ctx context.Context, path string, body any, out any) error {
	_, err := c.Do(ctx, "DELETE", path, nil, body, out)
	return err
}

// PostMultipart sends multipart/form-data (e.g. workitem attachment upload).
// fileField is the form field name for the file (usually "file").
// Not retried (non-idempotent).
func (c *Client) PostMultipart(ctx context.Context, path string, query map[string]string, fileField, filename string, data []byte, fields map[string]string, out any) error {
	u := c.buildURL(path, query)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}
	part, err := w.CreateFormFile(fileField, filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", u, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("x-yunxiao-token", c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			Status: resp.StatusCode,
			Body:   RedactSecrets(truncate(string(raw), 2000), c.Token),
			URL:    RedactSecrets(u, c.Token),
			Method: "POST",
		}
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		if s, ok := out.(*string); ok {
			*s = string(raw)
			return nil
		}
		return fmt.Errorf("decode response: %w; body=%s", err, truncate(string(raw), 500))
	}
	return nil
}

// Do executes the request and returns response headers (for pagination meta).
// Prefer Do over DoRaw when only headers are needed; use DoRaw when status code matters.
func (c *Client) Do(ctx context.Context, method, path string, query map[string]string, body any, out any) (http.Header, error) {
	hdr, _, err := c.DoRaw(ctx, method, path, query, body, out)
	return hdr, err
}

// Pagination holds common Yunxiao x-* response headers.
type Pagination struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	NextPage   int `json:"next_page,omitempty"`
	PrevPage   int `json:"prev_page,omitempty"`
}

func PaginationFromHeader(h http.Header) *Pagination {
	if h == nil {
		return nil
	}
	getInt := func(k string) int {
		v := h.Get(k)
		if v == "" {
			return 0
		}
		n, _ := strconv.Atoi(v)
		return n
	}
	p := &Pagination{
		Page:       getInt("x-page"),
		PerPage:    getInt("x-per-page"),
		Total:      getInt("x-total"),
		TotalPages: getInt("x-total-pages"),
		NextPage:   getInt("x-next-page"),
		PrevPage:   getInt("x-prev-page"),
	}
	if p.Page == 0 && p.PerPage == 0 && p.Total == 0 {
		return nil
	}
	return p
}

// MetaWithPagination merges Yunxiao list pagination headers into envelope meta.
// Agents get top-level has_more / total / page (when present) so truncation is visible;
// the nested "pagination" object is retained for full header detail.
// Callers that need every page may use ListAll (wired behind --all on some list cmds).
//
// has_more is true when x-next-page > 0, OR when total/page/per_page imply more
// rows remain (page*per_page < total). Relying only on x-next-page can false-negative
// when the API omits that header but still returns total.
func MetaWithPagination(base map[string]any, h http.Header) map[string]any {
	if base == nil {
		base = map[string]any{}
	}
	if p := PaginationFromHeader(h); p != nil {
		base["pagination"] = p
		base["has_more"] = inferHasMore(p)
		if p.Total > 0 {
			base["total"] = p.Total
		}
		if p.Page > 0 {
			base["page"] = p.Page
		}
	}
	return base
}

func inferHasMore(p *Pagination) bool {
	if p == nil {
		return false
	}
	if p.NextPage > 0 {
		return true
	}
	if p.Total > 0 && p.Page > 0 && p.PerPage > 0 {
		return p.Page*p.PerPage < p.Total
	}
	if p.Total > 0 && p.TotalPages > 0 && p.Page > 0 {
		return p.Page < p.TotalPages
	}
	return false
}

func idempotentMethod(method string) bool {
	m := strings.ToUpper(method)
	return m == http.MethodGet || m == http.MethodHead
}

func retryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

func parseRetryAfter(h http.Header) time.Duration {
	if h == nil {
		return 0
	}
	ra := h.Get("Retry-After")
	if ra == "" {
		return 0
	}
	if secs, err := strconv.Atoi(ra); err == nil {
		if secs < 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(ra); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

func backoffDuration(attempt int, respHeader http.Header) time.Duration {
	if d := parseRetryAfter(respHeader); d > 0 {
		if d > maxRetryAfter {
			return maxRetryAfter
		}
		return d
	}
	// attempt is 0-based index of the failed try; delay before next try.
	return retryBase << attempt
}

func (c *Client) DoRaw(ctx context.Context, method, path string, query map[string]string, body any, out any) (http.Header, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	u := c.buildURL(path, query)

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyBytes = b
	}

	attempts := 1
	if idempotentMethod(method) {
		attempts = maxAttempts
	}

	var lastHeader http.Header
	var lastStatus int
	var lastErr error

	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return lastHeader, lastStatus, err
		}

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("x-yunxiao-token", c.Token)

		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			lastHeader, lastStatus = nil, 0
			if attempt+1 >= attempts || !idempotentMethod(method) {
				return nil, 0, lastErr
			}
			if err := sleepWithContext(ctx, backoffDuration(attempt, nil)); err != nil {
				return nil, 0, err
			}
			continue
		}

		raw, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastHeader, lastStatus, lastErr = resp.Header, resp.StatusCode, readErr
			if attempt+1 >= attempts || !idempotentMethod(method) {
				return lastHeader, lastStatus, lastErr
			}
			if err := sleepWithContext(ctx, backoffDuration(attempt, resp.Header)); err != nil {
				return lastHeader, lastStatus, err
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			ae := &APIError{
				Status: resp.StatusCode,
				Body:   RedactSecrets(truncate(string(raw), 2000), c.Token),
				URL:    RedactSecrets(u, c.Token),
				Method: method,
			}
			lastHeader, lastStatus, lastErr = resp.Header, resp.StatusCode, ae
			if idempotentMethod(method) && retryableStatus(resp.StatusCode) && attempt+1 < attempts {
				if err := sleepWithContext(ctx, backoffDuration(attempt, resp.Header)); err != nil {
					return lastHeader, lastStatus, err
				}
				continue
			}
			return lastHeader, lastStatus, lastErr
		}

		if out == nil || len(raw) == 0 {
			return resp.Header, resp.StatusCode, nil
		}
		if err := json.Unmarshal(raw, out); err != nil {
			if s, ok := out.(*string); ok {
				*s = string(raw)
				return resp.Header, resp.StatusCode, nil
			}
			return resp.Header, resp.StatusCode, fmt.Errorf("decode response: %w; body=%s", err, truncate(string(raw), 500))
		}
		return resp.Header, resp.StatusCode, nil
	}

	return lastHeader, lastStatus, lastErr
}

type APIError struct {
	Status int
	Body   string
	URL    string
	Method string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("yunxiao API %s %s -> HTTP %d: %s", e.Method, e.URL, e.Status, e.Body)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// RedactSecrets removes known secrets from strings (token, common query keys).
func RedactSecrets(s, token string) string {
	out := s
	if token != "" && len(token) >= 4 {
		out = strings.ReplaceAll(out, token, "(redacted)")
	}
	// belt-and-suspenders for accidental embedding
	out = strings.ReplaceAll(out, "x-yunxiao-token=", "x-yunxiao-token=(redacted)")
	return out
}

// EncodeRepoID encodes organizationId/repo-name style IDs.
func EncodeRepoID(id string) string {
	if strings.Contains(id, "/") && !strings.Contains(id, "%2F") && !strings.Contains(id, "%2f") {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			return parts[0] + "%2F" + strings.ReplaceAll(url.QueryEscape(parts[1]), "+", "%20")
		}
		return url.PathEscape(id)
	}
	return id
}

// EncodeFilePath URL-encodes a repository file path for path segments.
func EncodeFilePath(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return p
	}
	return url.PathEscape(p)
}

// PageQuery builds page/perPage query map.
func PageQuery(page, perPage int) map[string]string {
	q := map[string]string{}
	if page > 0 {
		q["page"] = strconv.Itoa(page)
	}
	if perPage > 0 {
		q["perPage"] = strconv.Itoa(perPage)
	}
	return q
}
