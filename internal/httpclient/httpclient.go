// Package httpclient sends a domain.Request over HTTP after resolving its
// {{variable}} references, and returns the response with timing.
package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gbrlballesteros/termipost/internal/domain"
	"github.com/gbrlballesteros/termipost/internal/vars"
)

// Response is the captured result of sending a request.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Elapsed    time.Duration
}

// Send resolves variables in req, performs the HTTP call using ctx for
// cancellation, and returns the response. Unresolved variable names are
// returned so the caller can warn the user. A transport-level failure returns a
// non-nil error.
func Send(ctx context.Context, req domain.Request, r *vars.Resolver) (*Response, []string, error) {
	var unresolved []string
	collect := func(s string, u []string) string { unresolved = append(unresolved, u...); return s }

	rawURL := collect(r.Resolve(req.URL))
	headers, hu := r.ResolveKVs(req.Headers)
	unresolved = append(unresolved, hu...)
	params, pu := r.ResolveKVs(req.QueryParams)
	unresolved = append(unresolved, pu...)
	body := collect(r.Resolve(req.Body))

	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, unresolved, err
	}
	if len(params) > 0 {
		q := u.Query()
		for _, p := range params {
			if p.Key != "" {
				q.Add(p.Key, p.Value)
			}
		}
		u.RawQuery = q.Encode()
	}

	method := string(req.Method)
	if method == "" {
		method = string(domain.GET)
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, unresolved, err
	}
	for _, h := range headers {
		if h.Key != "" {
			httpReq.Header.Add(h.Key, h.Value)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, unresolved, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	elapsed := time.Since(start)
	if err != nil {
		return nil, unresolved, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       data,
		Elapsed:    elapsed,
	}, dedupe(unresolved), nil
}

func dedupe(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
