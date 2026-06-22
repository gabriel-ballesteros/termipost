package domain

import (
	"regexp"
	"strings"
	"testing"
)

func TestValidMethod(t *testing.T) {
	for _, m := range Methods {
		if !ValidMethod(m) {
			t.Errorf("ValidMethod(%q) = false, want true", m)
		}
	}
	for _, m := range []Method{"", "get", "FETCH", "TRACE"} {
		if ValidMethod(m) {
			t.Errorf("ValidMethod(%q) = true, want false", m)
		}
	}
}

func TestNewID(t *testing.T) {
	idRe := regexp.MustCompile(`^[a-z0-9-]+-[0-9a-f]{4}$`)
	cases := map[string]string{
		"Get Users":      "get-users-",
		"  Spaced  Out ": "spaced-out-",
		"MixedCASE":      "mixedcase-",
		"a/b\\c":         "a-b-c-",
		"!!!":            "item-", // no alphanumerics -> "item"
		"":               "item-",
	}
	for name, prefix := range cases {
		id := NewID(name)
		if !strings.HasPrefix(id, prefix) {
			t.Errorf("NewID(%q) = %q, want prefix %q", name, id, prefix)
		}
		if !idRe.MatchString(id) {
			t.Errorf("NewID(%q) = %q, does not match id pattern", name, id)
		}
	}
}

func TestNewIDUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := NewID("dup")
		if seen[id] {
			t.Fatalf("duplicate id generated: %q", id)
		}
		seen[id] = true
	}
}
