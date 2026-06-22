package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

// openSecrets navigates Collections -> Environments -> Secrets.
func openSecrets(t *testing.T, m *Model) *secretsScreen {
	t.Helper()
	send(t, m, keyMsg("e"))
	send(t, m, keyMsg("s"))
	ss, ok := m.top().(*secretsScreen)
	if !ok {
		t.Fatalf("expected secrets screen, got %T", m.top())
	}
	return ss
}

func TestSecretAddEditDelete(t *testing.T) {
	m := seededModel(t)
	openSecrets(t, m)

	// Add via "name: value" prompt.
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("token: abc123"))
	send(t, m, keyMsg("enter"))
	if m.app.secrets["token"] != "abc123" {
		t.Fatalf("secret not added: %+v", m.app.secrets)
	}

	// Edit it.
	send(t, m, keyMsg("enter")) // edit prompt prefilled
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("xyz789"))
	send(t, m, keyMsg("enter"))
	if m.app.secrets["token"] != "xyz789" {
		t.Fatalf("secret not edited: %+v", m.app.secrets)
	}

	// Delete it.
	send(t, m, keyMsg("d"))
	if _, exists := m.app.secrets["token"]; exists {
		t.Fatalf("secret not deleted: %+v", m.app.secrets)
	}
}

func TestSecretBadFormatRejected(t *testing.T) {
	m := seededModel(t)
	openSecrets(t, m)
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("noseparator"))
	send(t, m, keyMsg("enter"))
	if !m.statusErr {
		t.Fatal("secret without ':' should error")
	}
	if len(m.app.secrets) != 0 {
		t.Fatalf("no secret should be stored: %+v", m.app.secrets)
	}
}

func TestSecretRevealToggle(t *testing.T) {
	m := seededModel(t)
	ss := openSecrets(t, m)
	m.app.secrets["k"] = "plainval"
	ss.reloadKeys()

	if v := m.View(); strings.Contains(v, "plainval") || !strings.Contains(v, "••••••") {
		t.Fatalf("masked view should hide value:\n%s", v)
	}
	send(t, m, keyMsg("v")) // reveal
	if v := m.View(); !strings.Contains(v, "plainval") {
		t.Fatalf("revealed view should show value:\n%s", v)
	}
}

func TestParseKV(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		k, v string
	}{
		{"a: b", true, "a", "b"},
		{"host:api.test", true, "host", "api.test"},
		{"  spaced  :  val  ", true, "spaced", "val"},
		{"url: http://x:8080/p", true, "url", "http://x:8080/p"}, // only first colon splits
		{"noseparator", false, "", ""},
		{": noKey", false, "", ""},
	}
	for _, c := range cases {
		kv, ok := parseKV(c.in)
		if ok != c.ok {
			t.Errorf("parseKV(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && (kv.Key != c.k || kv.Value != c.v) {
			t.Errorf("parseKV(%q) = %+v, want {%q %q}", c.in, kv, c.k, c.v)
		}
	}
}

// ---- KV editor screen (direct, also reached via the request editor) ----

func TestKVEditorAddEditDeleteAndCommit(t *testing.T) {
	m := seededModel(t)
	var committed []domain.KV
	scr := newKVEditorScreen("Headers", nil, func(m *Model, pairs []domain.KV) tea.Cmd {
		committed = pairs
		return nil
	})
	m.push(scr)
	send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	// Add two entries.
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("X-A: 1"))
	send(t, m, keyMsg("enter"))
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("X-B: 2"))
	send(t, m, keyMsg("enter"))
	if len(scr.pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %+v", scr.pairs)
	}

	// Bad format rejected.
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("nope"))
	send(t, m, keyMsg("enter"))
	if !m.statusErr || len(scr.pairs) != 2 {
		t.Fatalf("bad KV format should be rejected: pairs=%+v", scr.pairs)
	}

	// Delete the first (cursor moved to last after adds).
	send(t, m, keyMsg("up")) // cursor -> first
	send(t, m, keyMsg("d"))
	if len(scr.pairs) != 1 || scr.pairs[0].Key != "X-B" {
		t.Fatalf("delete wrong entry: %+v", scr.pairs)
	}

	// Esc commits the remaining pairs via onDone.
	send(t, m, keyMsg("esc"))
	if len(committed) != 1 || committed[0].Key != "X-B" {
		t.Fatalf("onDone got %+v, want [X-B]", committed)
	}
}

func TestKVEditorEditEntry(t *testing.T) {
	m := seededModel(t)
	scr := newKVEditorScreen("Params", []domain.KV{{Key: "q", Value: "old"}}, nil)
	m.push(scr)
	send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	send(t, m, keyMsg("enter")) // edit prompt prefilled "q: old"
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("q: new"))
	send(t, m, keyMsg("enter"))
	if scr.pairs[0].Value != "new" {
		t.Fatalf("entry not edited: %+v", scr.pairs)
	}
}
