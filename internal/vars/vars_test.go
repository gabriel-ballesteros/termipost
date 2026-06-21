package vars

import (
	"reflect"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

func TestResolvePrecedenceEnvBeforeSecrets(t *testing.T) {
	r := New(map[string]string{"key": "env"}, domain.Secrets{"key": "secret", "tok": "s3cr3t"})

	out, unresolved := r.Resolve("{{key}} and {{tok}}")
	if out != "env and s3cr3t" {
		t.Fatalf("expected env to win, got %q", out)
	}
	if len(unresolved) != 0 {
		t.Fatalf("unexpected unresolved: %v", unresolved)
	}
}

func TestResolveUnresolvedLeftIntact(t *testing.T) {
	r := New(map[string]string{}, nil)
	out, unresolved := r.Resolve("a {{missing}} b {{missing}}")
	if out != "a {{missing}} b {{missing}}" {
		t.Fatalf("unresolved ref should be left intact, got %q", out)
	}
	if !reflect.DeepEqual(unresolved, []string{"missing"}) {
		t.Fatalf("expected [missing] once, got %v", unresolved)
	}
}

func TestResolveWhitespaceInRef(t *testing.T) {
	r := New(map[string]string{"base": "https://api"}, nil)
	out, _ := r.Resolve("{{ base }}/v1")
	if out != "https://api/v1" {
		t.Fatalf("whitespace ref not resolved: %q", out)
	}
}

func TestMaskSecrets(t *testing.T) {
	sec := domain.Secrets{"tok": "s3cr3t", "empty": ""}
	got := Mask("Authorization: Bearer s3cr3t", sec)
	if got != "Authorization: Bearer ••••••" {
		t.Fatalf("secret not masked: %q", got)
	}
	// Empty secret values must not mask everything.
	if Mask("hello", sec) != "hello" {
		t.Fatal("empty secret value should be ignored when masking")
	}
}
