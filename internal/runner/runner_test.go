package runner

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gbrlballesteros/termipost/internal/domain"
	"github.com/gbrlballesteros/termipost/internal/httpclient"
	"github.com/gbrlballesteros/termipost/internal/vars"
)

func resp() *httpclient.Response {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	return &httpclient.Response{
		StatusCode: 200,
		Headers:    h,
		Body:       []byte(`{"data":{"id":42,"name":"ok"},"items":["a","b"]}`),
		Elapsed:    50 * time.Millisecond,
	}
}

func TestEvaluateAssertions(t *testing.T) {
	cases := []struct {
		name string
		a    domain.Assertion
		want bool
	}{
		{"status pass", domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}, true},
		{"status fail", domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "404"}, false},
		{"header equals", domain.Assertion{Kind: domain.AssertHeader, Target: "Content-Type", Op: domain.OpEquals, Expected: "application/json"}, true},
		{"header contains", domain.Assertion{Kind: domain.AssertHeader, Target: "Content-Type", Op: domain.OpContains, Expected: "json"}, true},
		{"header regex", domain.Assertion{Kind: domain.AssertHeader, Target: "Content-Type", Op: domain.OpRegex, Expected: `application/.*`}, true},
		{"body contains", domain.Assertion{Kind: domain.AssertBody, Op: domain.OpContains, Expected: "ok"}, true},
		{"body jsonpath obj", domain.Assertion{Kind: domain.AssertBody, Op: domain.OpJSONPath, Target: "data.id", Expected: "42"}, true},
		{"body jsonpath arr", domain.Assertion{Kind: domain.AssertBody, Op: domain.OpJSONPath, Target: "items.1", Expected: "b"}, true},
		{"body jsonpath missing", domain.Assertion{Kind: domain.AssertBody, Op: domain.OpJSONPath, Target: "data.nope", Expected: "x"}, false},
		{"latency pass", domain.Assertion{Kind: domain.AssertLatency, Op: domain.OpMaxMS, Expected: "100"}, true},
		{"latency fail", domain.Assertion{Kind: domain.AssertLatency, Op: domain.OpMaxMS, Expected: "10"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := evaluate(c.a, resp())
			if res.Passed != c.want {
				t.Fatalf("got passed=%v want %v (detail: %s)", res.Passed, c.want, res.Detail)
			}
		})
	}
}

func TestRunCollectionSummary(t *testing.T) {
	col := domain.Collection{
		ID: "c", Name: "Col",
		Requests: []domain.Request{
			{ID: "noassert", Name: "no assertions"}, // skipped
			{ID: "bad", Name: "bad url", Method: domain.GET, URL: "http://127.0.0.1:0",
				Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}}, // error -> failed
		},
	}
	// The no-assertion request never hits the network, and the bad-url request
	// fails fast at the transport layer.
	agg := RunCollection(context.Background(), col, vars.New(nil, nil))
	if agg.Skipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", agg.Skipped)
	}
	if agg.Failed != 1 {
		t.Fatalf("expected 1 failed, got %d", agg.Failed)
	}
	if agg.Passed != 0 {
		t.Fatalf("expected 0 passed, got %d", agg.Passed)
	}
}
