package model

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/syntax"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// renderHeaders renders HTTP response headers sorted by name.
func renderHeaders(h http.Header) string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(ui.Label.Render(k+": ") + ui.Value.Render(strings.Join(h[k], ", ")) + "\n")
	}
	return b.String()
}

// prettyBody pretty-prints the body when it is JSON, otherwise returns it as-is.
func prettyBody(resp *httpclient.Response) string {
	body := resp.Body
	ct := resp.Headers.Get("Content-Type")
	if strings.Contains(ct, "json") || looksLikeJSON(body) {
		var out bytes.Buffer
		if err := json.Indent(&out, body, "", "  "); err == nil {
			return syntax.HighlightJSON(out.String())
		}
	}
	return string(body)
}

func looksLikeJSON(b []byte) bool {
	t := bytes.TrimSpace(b)
	return len(t) > 0 && (t[0] == '{' || t[0] == '[')
}

// bodyPreview shows the first lines of a request body for the non-editing view.
func bodyPreview(body string) string {
	if strings.TrimSpace(body) == "" {
		return ui.Subtle.Render("(empty — press enter to edit)")
	}
	shown := body
	if looksLikeJSON([]byte(body)) {
		shown = syntax.HighlightJSON(body)
	}
	lines := strings.Split(shown, "\n")
	if len(lines) > 6 {
		lines = append(lines[:6], "…")
	}
	return strings.Join(lines, "\n")
}

// parseKV parses a "key: value" string into a domain.KV.
func parseKV(s string) (domain.KV, bool) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return domain.KV{}, false
	}
	key := strings.TrimSpace(s[:idx])
	val := strings.TrimSpace(s[idx+1:])
	if key == "" {
		return domain.KV{}, false
	}
	return domain.KV{Key: key, Value: val}, true
}
