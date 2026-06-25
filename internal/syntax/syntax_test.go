package syntax

import (
	"strings"
	"testing"
)

func TestValidateJSONValid(t *testing.T) {
	valid := []string{
		`{}`,
		`[]`,
		`{"a":1}`,
		`{"a": 1, "b": [2, 3, {"c": true}]}`,
		`"just a string"`,
		`123`,
		`-0`,
		`1e10`,
		`1.5E-3`,
		`1.5e+3`,
		`[true, false, null]`,
		`"escapes \" \\ \/ \b \f \n \r \t"`,
		`"unicode é ꯍ"`,
		`{"nested":{"deep":{"deeper":[1,[2,[3,[4]]]]}}}`,
		"  {\n  \"x\": 1\n}  ",
		`{"empty":""}`,
		`0.5`,
	}
	for _, s := range valid {
		if ok, err := ValidateJSON(s); !ok {
			t.Errorf("ValidateJSON(%q) = false (false positive), err=%v", s, err)
		}
	}
}

func TestValidateJSONInvalid(t *testing.T) {
	invalid := []string{
		``,
		`{`,
		`{"a":}`,
		`{"a":1,}`,
		`[1,2,]`,
		`{a:1}`,
		`{"a" 1}`,
		`01`,
		`1.`,
		`1e`,
		`-`,
		`tru`,
		`"unterminated`,
		`"bad \x escape"`,
		`"bad \u00gz"`,
		`{} extra`,
		`{"a":1} {"b":2}`,
		`nul`,
	}
	for _, s := range invalid {
		if ok, _ := ValidateJSON(s); ok {
			t.Errorf("ValidateJSON(%q) = true, want invalid", s)
		}
	}
}

func TestValidateJSONErrorLocation(t *testing.T) {
	// Error should point at the offending position on line 2.
	ok, err := ValidateJSON("{\n  \"a\": }\n}")
	if ok || err == nil {
		t.Fatal("expected invalid with error")
	}
	if err.Line != 2 {
		t.Errorf("error line = %d, want 2 (%s)", err.Line, err.Error())
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("error message missing location: %s", err.Error())
	}
}

func TestHighlightJSONRoundTripsValid(t *testing.T) {
	// Highlight styles auto-degrade to plain text in non-terminal test output, so
	// the rendered string must reconstruct the original content exactly.
	src := `{"a": 1, "b": [true, null, "x"]}`
	out := HighlightJSON(src)
	if out != src {
		t.Errorf("HighlightJSON changed content:\n got %q\nwant %q", out, src)
	}
}

func TestHighlightJSONBailsOnMalformed(t *testing.T) {
	src := `not json at all {`
	if out := HighlightJSON(src); out != src {
		t.Errorf("HighlightJSON(malformed) = %q, want passthrough %q", out, src)
	}
}

func TestPrettifyValid(t *testing.T) {
	out, err := Prettify(`{"a":1,"b":[2,3]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"\"a\": 1", "\n  ", "[\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("prettified output missing %q:\n%s", want, out)
		}
	}
}

func TestPrettifyIdempotent(t *testing.T) {
	once, err := Prettify(`{"a":1,"b":2}`)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := Prettify(once)
	if err != nil {
		t.Fatal(err)
	}
	if once != twice {
		t.Errorf("prettify not idempotent:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
}

func TestPrettifyEmpty(t *testing.T) {
	for _, s := range []string{"", "   ", "\n\t "} {
		out, err := Prettify(s)
		if err != nil {
			t.Errorf("Prettify(%q) error = %v, want nil", s, err)
		}
		if out != s {
			t.Errorf("Prettify(%q) = %q, want unchanged", s, out)
		}
	}
}

func TestPrettifyInvalidUnchanged(t *testing.T) {
	in := `{"a": }`
	out, err := Prettify(in)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if out != in {
		t.Errorf("body mutated on invalid input: %q", out)
	}
	if err.Line == 0 || err.Col == 0 {
		t.Errorf("error missing line/column: %+v", err)
	}
}
