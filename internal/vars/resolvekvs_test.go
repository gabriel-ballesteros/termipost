package vars

import (
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

func TestResolveKVs(t *testing.T) {
	r := New(map[string]string{"h": "X-Token", "v": "abc"}, domain.Secrets{"sec": "shh"})
	kvs := []domain.KV{
		{Key: "{{h}}", Value: "{{v}}"},
		{Key: "Authorization", Value: "Bearer {{sec}}"},
		{Key: "{{missingKey}}", Value: "{{missingVal}}"},
	}
	out, unresolved := r.ResolveKVs(kvs)

	if out[0].Key != "X-Token" || out[0].Value != "abc" {
		t.Fatalf("pair 0 = %+v, want {X-Token abc}", out[0])
	}
	if out[1].Value != "Bearer shh" {
		t.Fatalf("pair 1 value = %q, want secret resolved", out[1].Value)
	}
	// Unresolved names accumulate across keys and values.
	want := map[string]bool{"missingKey": true, "missingVal": true}
	if len(unresolved) != len(want) {
		t.Fatalf("unresolved = %v, want %v", unresolved, want)
	}
	for _, u := range unresolved {
		if !want[u] {
			t.Errorf("unexpected unresolved %q", u)
		}
	}
}

func TestResolveKVsEmpty(t *testing.T) {
	out, unresolved := New(nil, nil).ResolveKVs(nil)
	if len(out) != 0 || len(unresolved) != 0 {
		t.Fatalf("empty input should yield empty output: out=%v unresolved=%v", out, unresolved)
	}
}
