// Package runner evaluates a request's assertions against a response and runs
// requests (and whole collections) as tests.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/vars"
)

// RunRequest sends req and evaluates its assertions, returning a RunResult.
// A request with no assertions yields a Skipped result.
func RunRequest(ctx context.Context, req domain.Request, r *vars.Resolver) domain.RunResult {
	res := domain.RunResult{RequestID: req.ID, RequestName: req.Name}
	if len(req.Assertions) == 0 {
		res.Status = domain.RunSkipped
		return res
	}

	resp, _, err := httpclient.Send(ctx, req, r)
	if err != nil {
		res.Status = domain.RunError
		res.Err = err.Error()
		return res
	}
	res.StatusCode = resp.StatusCode
	res.Elapsed = resp.Elapsed
	res.Assertions = EvaluateAssertions(req.Assertions, resp)

	res.Status = domain.RunPassed
	for _, a := range res.Assertions {
		if !a.Passed {
			res.Status = domain.RunFailed
			break
		}
	}
	return res
}

// RunCollection runs every request in col that has assertions, skipping the
// rest, and aggregates the outcome.
func RunCollection(ctx context.Context, col domain.Collection, r *vars.Resolver) domain.CollectionRunResult {
	var agg domain.CollectionRunResult
	for _, req := range col.Requests {
		rr := RunRequest(ctx, req, r)
		agg.Results = append(agg.Results, rr)
		switch rr.Status {
		case domain.RunPassed:
			agg.Passed++
		case domain.RunSkipped:
			agg.Skipped++
		default: // failed or error
			agg.Failed++
		}
	}
	return agg
}

// EvaluateAssertions checks every assertion against resp.
func EvaluateAssertions(assertions []domain.Assertion, resp *httpclient.Response) []domain.AssertionResult {
	out := make([]domain.AssertionResult, 0, len(assertions))
	for _, a := range assertions {
		out = append(out, evaluate(a, resp))
	}
	return out
}

func evaluate(a domain.Assertion, resp *httpclient.Response) domain.AssertionResult {
	switch a.Kind {
	case domain.AssertStatusCode:
		want, err := strconv.Atoi(strings.TrimSpace(a.Expected))
		if err != nil {
			return fail(a, fmt.Sprintf("invalid expected status %q", a.Expected))
		}
		if a.Op == domain.OpNotEquals {
			return result(a, resp.StatusCode != want, fmt.Sprintf("status: expected != %d, got %d", want, resp.StatusCode))
		}
		return result(a, resp.StatusCode == want, fmt.Sprintf("status: expected %d, got %d", want, resp.StatusCode))

	case domain.AssertHeader:
		got := resp.Headers.Get(a.Target)
		ok, detail := matchString(a.Op, a.Expected, got, a.Target)
		return result(a, ok, detail)

	case domain.AssertBody:
		if a.Op == domain.OpJSONPath {
			got, found := jsonPath(resp.Body, a.Target)
			if !found {
				return fail(a, fmt.Sprintf("body: path %q not found", a.Target))
			}
			return result(a, got == a.Expected, fmt.Sprintf("body[%s]: expected %q, got %q", a.Target, a.Expected, got))
		}
		ok, detail := matchString(a.Op, a.Expected, string(resp.Body), "body")
		return result(a, ok, detail)

	case domain.AssertLatency:
		max, err := strconv.ParseInt(strings.TrimSpace(a.Expected), 10, 64)
		if err != nil {
			return fail(a, fmt.Sprintf("invalid expected latency %q", a.Expected))
		}
		got := resp.Elapsed.Milliseconds()
		return result(a, got <= max, fmt.Sprintf("latency: expected <= %dms, got %dms", max, got))

	default:
		return fail(a, fmt.Sprintf("unknown assertion kind %q", a.Kind))
	}
}

// matchString applies an op (equals/contains/regex) against got, returning a
// pass flag and a human-readable detail.
func matchString(op domain.MatchOp, expected, got, label string) (bool, string) {
	switch op {
	case domain.OpEquals, "":
		return got == expected, fmt.Sprintf("%s: expected %q, got %q", label, expected, got)
	case domain.OpContains:
		return strings.Contains(got, expected), fmt.Sprintf("%s: expected to contain %q, got %q", label, expected, got)
	case domain.OpRegex:
		re, err := regexp.Compile(expected)
		if err != nil {
			return false, fmt.Sprintf("%s: invalid regex %q", label, expected)
		}
		return re.MatchString(got), fmt.Sprintf("%s: expected to match /%s/, got %q", label, expected, got)
	default:
		return false, fmt.Sprintf("%s: unknown operator %q", label, op)
	}
}

// jsonPath looks up a dotted path (e.g. "data.items.0.id") in JSON body and
// returns the value rendered as a string.
func jsonPath(body []byte, path string) (string, bool) {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return "", false
	}
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			continue
		}
		switch cur := v.(type) {
		case map[string]any:
			next, ok := cur[seg]
			if !ok {
				return "", false
			}
			v = next
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(cur) {
				return "", false
			}
			v = cur[idx]
		default:
			return "", false
		}
	}
	return stringify(v), true
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func result(a domain.Assertion, passed bool, detail string) domain.AssertionResult {
	return domain.AssertionResult{Assertion: a, Passed: passed, Detail: detail}
}

func fail(a domain.Assertion, detail string) domain.AssertionResult {
	return domain.AssertionResult{Assertion: a, Passed: false, Detail: detail}
}
