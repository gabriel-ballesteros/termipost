// Package domain defines the core data types for termipost: collections,
// requests, assertions, environments, and secrets, plus the result types
// produced when requests are run as tests.
package domain

import "time"

// Method is an HTTP method.
type Method string

// Supported HTTP methods.
const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	PATCH   Method = "PATCH"
	DELETE  Method = "DELETE"
	HEAD    Method = "HEAD"
	OPTIONS Method = "OPTIONS"
)

// Methods lists every selectable HTTP method, in display order.
var Methods = []Method{GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS}

// ValidMethod reports whether m is one of the supported methods.
func ValidMethod(m Method) bool {
	for _, x := range Methods {
		if x == m {
			return true
		}
	}
	return false
}

// KV is an ordered key/value pair used for headers and query parameters.
type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AssertionKind identifies what part of a response an assertion checks.
type AssertionKind string

// Assertion kinds.
const (
	AssertStatusCode AssertionKind = "status_code"
	AssertHeader     AssertionKind = "header"
	AssertBody       AssertionKind = "body"
	AssertLatency    AssertionKind = "latency"
)

// MatchOp identifies how an assertion compares expected vs actual values.
type MatchOp string

// Match operators.
const (
	OpEquals   MatchOp = "equals"
	OpContains MatchOp = "contains"
	OpRegex    MatchOp = "regex"
	OpJSONPath MatchOp = "json_path" // body: dotted JSON path equals Expected
	OpMaxMS    MatchOp = "max_ms"    // latency: elapsed must be <= Expected ms
)

// Assertion is a single expectation evaluated against a response.
//
// Target meaning by Kind:
//   - status_code: ignored
//   - header:      the header name to inspect
//   - body:        for OpJSONPath, the dotted path (e.g. "data.id"); otherwise ignored
//   - latency:     ignored
type Assertion struct {
	Kind     AssertionKind `json:"kind"`
	Target   string        `json:"target,omitempty"`
	Op       MatchOp       `json:"op"`
	Expected string        `json:"expected"`
}

// Request is an HTTP request and any assertions attached to it. A request that
// carries assertions is treated as a test; there is no separate test entity.
type Request struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Method      Method      `json:"method"`
	URL         string      `json:"url"`
	Headers     []KV        `json:"headers,omitempty"`
	QueryParams []KV        `json:"queryParams,omitempty"`
	Body        string      `json:"body,omitempty"`
	Assertions  []Assertion `json:"assertions,omitempty"`
}

// Collection groups related requests.
type Collection struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Requests []Request `json:"requests"`
}

// Environment is a named set of variables. One environment is active at a time.
type Environment struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Vars map[string]string `json:"vars"`
}

// Secrets is the single global secret store (gitignored on disk).
type Secrets map[string]string

// Config holds application settings persisted in config.json.
type Config struct {
	ActiveEnvironmentID string `json:"activeEnvironmentId"`
}

// RunStatus is the outcome of running a request as a test.
type RunStatus string

// Run statuses.
const (
	RunPassed  RunStatus = "passed"
	RunFailed  RunStatus = "failed"
	RunSkipped RunStatus = "skipped"
	RunError   RunStatus = "error"
)

// AssertionResult is the outcome of evaluating one assertion.
type AssertionResult struct {
	Assertion Assertion
	Passed    bool
	Detail    string // human-readable expected vs actual
}

// RunResult is the outcome of running a single request's assertions.
type RunResult struct {
	RequestID   string
	RequestName string
	Status      RunStatus
	StatusCode  int
	Elapsed     time.Duration
	Assertions  []AssertionResult
	Err         string
}

// CollectionRunResult aggregates the results of running a whole collection.
type CollectionRunResult struct {
	Results []RunResult
	Passed  int
	Failed  int
	Skipped int
}
