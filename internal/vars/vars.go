// Package vars resolves {{name}} references in request fields against the active
// environment and the global secrets store, and masks secret values for display.
package vars

import (
	"regexp"
	"strings"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

// refRe matches {{ name }} references with optional surrounding whitespace.
var refRe = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.\-]+)\s*\}\}`)

// Resolver substitutes references using a single layer: active-environment
// variables first, then the global secrets store.
type Resolver struct {
	env     map[string]string
	secrets map[string]string
}

// New builds a Resolver from the active environment's variables (may be nil)
// and the global secrets store (may be nil).
func New(env map[string]string, secrets domain.Secrets) *Resolver {
	return &Resolver{env: env, secrets: secrets}
}

// Resolve substitutes every {{name}} reference in s. References that resolve
// from neither the environment nor secrets are left intact and their names are
// returned (de-duplicated) so the caller can warn the user.
func (r *Resolver) Resolve(s string) (out string, unresolved []string) {
	seen := map[string]bool{}
	out = refRe.ReplaceAllStringFunc(s, func(match string) string {
		name := strings.TrimSpace(refRe.FindStringSubmatch(match)[1])
		if v, ok := r.env[name]; ok {
			return v
		}
		if v, ok := r.secrets[name]; ok {
			return v
		}
		if !seen[name] {
			seen[name] = true
			unresolved = append(unresolved, name)
		}
		return match // leave unresolved reference untouched
	})
	return out, unresolved
}

// ResolveKVs resolves both key and value of each pair, accumulating unresolved
// names across all of them.
func (r *Resolver) ResolveKVs(kvs []domain.KV) (out []domain.KV, unresolved []string) {
	for _, kv := range kvs {
		k, u1 := r.Resolve(kv.Key)
		v, u2 := r.Resolve(kv.Value)
		out = append(out, domain.KV{Key: k, Value: v})
		unresolved = append(unresolved, u1...)
		unresolved = append(unresolved, u2...)
	}
	return out, unresolved
}

// Mask replaces every non-empty secret value occurring in s with a fixed-width
// mask, so secret values never leak into a rendered preview. The real value is
// still what gets sent over the wire (callers mask only for display).
func Mask(s string, secrets domain.Secrets) string {
	for _, v := range secrets {
		if v == "" {
			continue
		}
		s = strings.ReplaceAll(s, v, "••••••")
	}
	return s
}
