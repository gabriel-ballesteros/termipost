package model

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

func cmdApp(t *testing.T) *App {
	t.Helper()
	return NewApp(store.New(t.TempDir()), &store.Data{})
}

func TestSendCmd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	msg := sendCmd(cmdApp(t), domain.Request{Method: domain.GET, URL: srv.URL})()
	res, ok := msg.(sendResultMsg)
	if !ok {
		t.Fatalf("sendCmd returned %T, want sendResultMsg", msg)
	}
	if res.err != nil {
		t.Fatalf("unexpected err: %v", res.err)
	}
	if res.resp.StatusCode != 204 {
		t.Fatalf("status = %d, want 204", res.resp.StatusCode)
	}
}

func TestSendCmdError(t *testing.T) {
	msg := sendCmd(cmdApp(t), domain.Request{Method: domain.GET, URL: "http://127.0.0.1:0"})()
	res := msg.(sendResultMsg)
	if res.err == nil {
		t.Fatal("expected transport error from sendCmd")
	}
}

func TestRunRequestCmd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	req := domain.Request{
		Method: domain.GET, URL: srv.URL,
		Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}},
	}
	msg := runRequestCmd(cmdApp(t), req)()
	res, ok := msg.(reqRunMsg)
	if !ok {
		t.Fatalf("runRequestCmd returned %T, want reqRunMsg", msg)
	}
	if res.result.Status != domain.RunPassed {
		t.Fatalf("status = %q, want passed", res.result.Status)
	}
}

func TestRunCollectionCmd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	col := domain.Collection{Name: "Suite", Requests: []domain.Request{
		{ID: "r1", Name: "ok", Method: domain.GET, URL: srv.URL,
			Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}},
		{ID: "r2", Name: "skip", Method: domain.GET, URL: srv.URL}, // no assertions -> skipped
	}}
	msg := runCollectionCmd(cmdApp(t), col)()
	res, ok := msg.(collRunMsg)
	if !ok {
		t.Fatalf("runCollectionCmd returned %T, want collRunMsg", msg)
	}
	if res.collectionName != "Suite" {
		t.Fatalf("collectionName = %q", res.collectionName)
	}
	if res.result.Passed != 1 || res.result.Skipped != 1 {
		t.Fatalf("result = %+v, want 1 passed / 1 skipped", res.result)
	}
}
