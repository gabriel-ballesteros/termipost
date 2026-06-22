package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/vars"
)

func TestEvaluateAssertionsAggregate(t *testing.T) {
	r := resp() // status 200, json body
	out := EvaluateAssertions([]domain.Assertion{
		{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"},
		{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "500"},
	}, r)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if !out[0].Passed || out[1].Passed {
		t.Fatalf("results = %+v, want [pass fail]", out)
	}
}

func TestEvaluateInvalidExpected(t *testing.T) {
	r := resp()
	if evaluate(domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "abc"}, r).Passed {
		t.Error("non-numeric status expected should fail")
	}
	if evaluate(domain.Assertion{Kind: domain.AssertLatency, Op: domain.OpMaxMS, Expected: "fast"}, r).Passed {
		t.Error("non-numeric latency expected should fail")
	}
	if evaluate(domain.Assertion{Kind: "bogus"}, r).Passed {
		t.Error("unknown assertion kind should fail")
	}
}

func TestMatchStringOps(t *testing.T) {
	// empty op behaves like equals
	if ok, _ := matchString("", "a", "a", "l"); !ok {
		t.Error("empty op should behave like equals")
	}
	// invalid regex fails cleanly
	if ok, detail := matchString(domain.OpRegex, "[", "x", "l"); ok || detail == "" {
		t.Errorf("invalid regex should fail with detail, got ok=%v detail=%q", ok, detail)
	}
	// unknown operator
	if ok, _ := matchString("weird", "a", "a", "l"); ok {
		t.Error("unknown operator should fail")
	}
}

func TestJSONPathErrors(t *testing.T) {
	cases := []struct {
		body []byte
		path string
	}{
		{[]byte("not json"), "a"},      // unmarshal fails
		{[]byte(`{"a":1}`), "a.b"},     // descend into non-container
		{[]byte(`[1,2]`), "5"},         // index out of range
		{[]byte(`[1,2]`), "x"},         // non-numeric index
		{[]byte(`{"a":1}`), "missing"}, // missing key
	}
	for _, c := range cases {
		if _, found := jsonPath(c.body, c.path); found {
			t.Errorf("jsonPath(%s, %q) found, want not found", c.body, c.path)
		}
	}
}

func TestJSONPathStringifyTypes(t *testing.T) {
	body := []byte(`{"s":"hi","n":3.5,"b":true,"nul":null,"obj":{"k":1},"arr":[1,2]}`)
	cases := map[string]string{
		"s":   "hi",
		"n":   "3.5",
		"b":   "true",
		"nul": "null",
		"obj": `{"k":1}`,
		"arr": "[1,2]",
	}
	for path, want := range cases {
		got, found := jsonPath(body, path)
		if !found || got != want {
			t.Errorf("jsonPath %q = %q (found=%v), want %q", path, got, found, want)
		}
	}
	// Leading/empty segments are skipped.
	if got, _ := jsonPath(body, ".s"); got != "hi" {
		t.Errorf("leading-dot path = %q, want hi", got)
	}
}

func TestRunRequestSkippedAndError(t *testing.T) {
	// No assertions -> skipped, no network.
	res := RunRequest(context.Background(), domain.Request{ID: "r", Name: "n"}, vars.New(nil, nil))
	if res.Status != domain.RunSkipped {
		t.Fatalf("status = %q, want skipped", res.Status)
	}

	// Transport failure -> error.
	bad := domain.Request{ID: "r", Name: "bad", Method: domain.GET, URL: "http://127.0.0.1:0",
		Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}}
	res = RunRequest(context.Background(), bad, vars.New(nil, nil))
	if res.Status != domain.RunError || res.Err == "" {
		t.Fatalf("expected error status with message, got %+v", res)
	}
}

func TestRunRequestPassedAndFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	mk := func(expected string) domain.RunResult {
		req := domain.Request{ID: "r", Name: "n", Method: domain.GET, URL: srv.URL,
			Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: expected}}}
		return RunRequest(context.Background(), req, vars.New(nil, nil))
	}
	if mk("200").Status != domain.RunPassed {
		t.Error("matching status should pass")
	}
	if mk("404").Status != domain.RunFailed {
		t.Error("mismatched status should fail")
	}
}
