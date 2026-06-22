package model

import (
	"strings"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

func TestDescribeAssertion(t *testing.T) {
	cases := []struct {
		a    domain.Assertion
		want string
	}{
		{domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}, "status code == 200"},
		{domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpNotEquals, Expected: "500"}, "status code != 500"},
		{domain.Assertion{Kind: domain.AssertHeader, Target: "Content-Type", Op: domain.OpContains, Expected: "json"}, `header "Content-Type" contains "json"`},
		{domain.Assertion{Kind: domain.AssertBody, Op: domain.OpJSONPath, Target: "data.id", Expected: "42"}, `body json "data.id" == "42"`},
		{domain.Assertion{Kind: domain.AssertBody, Op: domain.OpContains, Expected: "ok"}, `body contains "ok"`},
		{domain.Assertion{Kind: domain.AssertLatency, Op: domain.OpMaxMS, Expected: "100"}, "latency <= 100ms"},
		{domain.Assertion{Kind: "unknown"}, "unknown"},
	}
	for _, c := range cases {
		if got := describeAssertion(c.a); got != c.want {
			t.Errorf("describeAssertion(%+v) = %q, want %q", c.a, got, c.want)
		}
	}
}

func TestOpsFor(t *testing.T) {
	for _, k := range assertionKinds {
		ops := opsFor(k)
		if len(ops) == 0 {
			t.Errorf("opsFor(%q) returned no ops", k)
		}
	}
	if opsFor(domain.AssertLatency)[0] != domain.OpMaxMS {
		t.Error("latency op must be max_ms")
	}
	if opsFor("garbage")[0] != domain.OpEquals {
		t.Error("unknown kind should fall back to status_code ops (equals first)")
	}
}

func TestUsesTarget(t *testing.T) {
	yes := []domain.Assertion{
		{Kind: domain.AssertHeader},
		{Kind: domain.AssertBody, Op: domain.OpJSONPath},
	}
	no := []domain.Assertion{
		{Kind: domain.AssertStatusCode},
		{Kind: domain.AssertLatency},
		{Kind: domain.AssertBody, Op: domain.OpContains},
	}
	for _, a := range yes {
		if !usesTarget(a) {
			t.Errorf("usesTarget(%+v) = false, want true", a)
		}
	}
	for _, a := range no {
		if usesTarget(a) {
			t.Errorf("usesTarget(%+v) = true, want false", a)
		}
	}
}

func TestStepField(t *testing.T) {
	fields := []assertionField{aKind, aOp, aExpected}
	if got := stepField(fields, aKind, 1); got != aOp {
		t.Errorf("step +1 from aKind = %d, want aOp", got)
	}
	if got := stepField(fields, aKind, -1); got != aExpected {
		t.Errorf("step -1 from aKind should wrap to aExpected, got %d", got)
	}
	if got := stepField(fields, aExpected, 1); got != aKind {
		t.Errorf("step +1 from last should wrap to aKind, got %d", got)
	}
	// focus not in visible set -> first field
	if got := stepField(fields, aTarget, 1); got != aKind {
		t.Errorf("step from hidden field = %d, want aKind", got)
	}
}

func TestFieldVisible(t *testing.T) {
	fields := []assertionField{aKind, aOp, aExpected}
	if !fieldVisible(fields, aOp) {
		t.Error("aOp should be visible")
	}
	if fieldVisible(fields, aTarget) {
		t.Error("aTarget should not be visible")
	}
}

func TestWrap(t *testing.T) {
	if got := wrap(0, true, 3); got != 1 {
		t.Errorf("wrap forward = %d, want 1", got)
	}
	if got := wrap(2, true, 3); got != 0 {
		t.Errorf("wrap forward overflow = %d, want 0", got)
	}
	if got := wrap(0, false, 3); got != 2 {
		t.Errorf("wrap backward underflow = %d, want 2", got)
	}
}

func TestElideLeft(t *testing.T) {
	if got := elideLeft("hello", 0); got != "" {
		t.Errorf("width 0 = %q, want empty", got)
	}
	if got := elideLeft("hello", 10); got != "hello" {
		t.Errorf("fits = %q, want hello", got)
	}
	got := elideLeft("/very/long/path/here", 8)
	if !strings.HasPrefix(got, "…") {
		t.Errorf("elided = %q, want leading ellipsis", got)
	}
	if len([]rune(got)) > 8 {
		t.Errorf("elided %q exceeds width 8", got)
	}
}

func TestAppName(t *testing.T) {
	cases := map[string]string{
		"":       "termipost (dev)",
		"dev":    "termipost (dev)",
		"1.2.3":  "termipost v1.2.3",
		"v2.0.0": "termipost v2.0.0",
	}
	for ver, want := range cases {
		m := &Model{version: ver}
		if got := m.appName(); got != want {
			t.Errorf("appName(%q) = %q, want %q", ver, got, want)
		}
	}
}
