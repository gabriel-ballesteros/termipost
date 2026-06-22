package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/vars"
)

func TestSendResolvesAndCalls(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotHeader, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("q")
		gotHeader = r.Header.Get("X-Token")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	req := domain.Request{
		Method:      domain.POST,
		URL:         srv.URL + "/users",
		Headers:     []domain.KV{{Key: "X-Token", Value: "{{tok}}"}},
		QueryParams: []domain.KV{{Key: "q", Value: "hi"}},
		Body:        `{"name":"a"}`,
	}
	r := vars.New(map[string]string{"tok": "secret123"}, nil)

	resp, unresolved, err := Send(context.Background(), req, r)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(unresolved) != 0 {
		t.Errorf("unresolved = %v, want none", unresolved)
	}
	if resp.StatusCode != 201 {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}
	if string(resp.Body) != `{"ok":true}` {
		t.Errorf("Body = %q", resp.Body)
	}
	if resp.Elapsed <= 0 {
		t.Errorf("Elapsed = %v, want > 0", resp.Elapsed)
	}
	if gotMethod != "POST" || gotPath != "/users" || gotQuery != "hi" {
		t.Errorf("server saw method=%q path=%q query=%q", gotMethod, gotPath, gotQuery)
	}
	if gotHeader != "secret123" {
		t.Errorf("header X-Token = %q, want resolved value", gotHeader)
	}
	if gotBody != `{"name":"a"}` {
		t.Errorf("body = %q", gotBody)
	}
}

func TestSendReportsUnresolved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	req := domain.Request{
		Method:  domain.GET,
		URL:     srv.URL,
		Headers: []domain.KV{{Key: "X-A", Value: "{{missing}}"}},
		Body:    "{{missing}} {{alsoMissing}}",
	}
	_, unresolved, err := Send(context.Background(), req, vars.New(nil, nil))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	// dedupe: "missing" appears twice but should be reported once.
	want := map[string]bool{"missing": true, "alsoMissing": true}
	if len(unresolved) != len(want) {
		t.Fatalf("unresolved = %v, want keys %v", unresolved, want)
	}
	for _, u := range unresolved {
		if !want[u] {
			t.Errorf("unexpected unresolved name %q", u)
		}
	}
}

func TestSendDefaultsToGET(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
	}))
	defer srv.Close()

	req := domain.Request{URL: srv.URL} // no method
	if _, _, err := Send(context.Background(), req, vars.New(nil, nil)); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotMethod != "GET" {
		t.Errorf("default method = %q, want GET", gotMethod)
	}
}

func TestSendTransportError(t *testing.T) {
	req := domain.Request{Method: domain.GET, URL: "http://127.0.0.1:0"}
	if _, _, err := Send(context.Background(), req, vars.New(nil, nil)); err == nil {
		t.Fatal("expected transport error, got nil")
	}
}

func TestSendBadURL(t *testing.T) {
	req := domain.Request{Method: domain.GET, URL: "://bad url"}
	if _, _, err := Send(context.Background(), req, vars.New(nil, nil)); err == nil {
		t.Fatal("expected parse error for malformed URL, got nil")
	}
}

func TestSendCancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	req := domain.Request{Method: domain.GET, URL: srv.URL}
	if _, _, err := Send(ctx, req, vars.New(nil, nil)); err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}
