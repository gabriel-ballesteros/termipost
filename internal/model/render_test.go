package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

// renderTop sizes the model and returns the rendered view of the top screen,
// exercising its View/Title/HelpBindings/Crumb path.
func renderTop(t *testing.T, m *Model) string {
	t.Helper()
	send(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	return m.View()
}

func TestRenderWorkspacePanes(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://example.com"},
	}}
	m := seededModel(t, col)
	v := renderTop(t, m)
	for _, want := range []string{"Collections", "Request", "Response", "Send", "API", "Get"} {
		if !strings.Contains(v, want) {
			t.Fatalf("workspace view missing %q:\n%s", want, v)
		}
	}
}

func TestRenderAssertionsScreen(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://x",
			Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}},
	}}
	m := seededModel(t, col)
	ws(m).focus = paneEditor
	send(t, m, keyMsg("a")) // assertions screen
	v := renderTop(t, m)
	if !strings.Contains(v, "Assertions") || !strings.Contains(v, "status code == 200") {
		t.Fatalf("assertions screen view missing content:\n%s", v)
	}

	// Empty-state branch.
	m2 := seededModel(t, domain.Collection{ID: "c2", Name: "B", Requests: []domain.Request{{ID: "r9", Name: "n", Method: domain.GET}}})
	ws(m2).focus = paneEditor
	send(t, m2, keyMsg("a"))
	if v := renderTop(t, m2); !strings.Contains(v, "No assertions") {
		t.Fatalf("expected empty assertions hint:\n%s", v)
	}
}

func TestRenderAssertionEditScreen(t *testing.T) {
	cases := []struct {
		a    domain.Assertion
		want string
	}{
		{domain.Assertion{Kind: domain.AssertHeader, Op: domain.OpEquals, Target: "Content-Type", Expected: "json"}, "Target"},
		{domain.Assertion{Kind: domain.AssertBody, Op: domain.OpJSONPath, Target: "data.id", Expected: "1"}, "JSON path"},
		{domain.Assertion{Kind: domain.AssertLatency, Op: domain.OpMaxMS, Expected: "100"}, "Max ms"},
	}
	for _, c := range cases {
		m := seededModel(t)
		scr := newAssertionEditScreen(c.a, nil)
		m.push(scr)
		v := renderTop(t, m)
		if !strings.Contains(v, "Edit assertion") || !strings.Contains(v, c.want) {
			t.Fatalf("assertion editor view missing %q:\n%s", c.want, v)
		}
	}
}

func TestRenderKVEditor(t *testing.T) {
	m := seededModel(t)
	scr := newKVEditorScreen("Headers", []domain.KV{{Key: "X-A", Value: "1"}}, nil)
	m.push(scr)
	if v := renderTop(t, m); !strings.Contains(v, "Headers") || !strings.Contains(v, "X-A: 1") {
		t.Fatalf("kv editor view missing content:\n%s", v)
	}

	m2 := seededModel(t)
	m2.push(newKVEditorScreen("Params", nil, nil))
	if v := renderTop(t, m2); !strings.Contains(v, "No entries") {
		t.Fatalf("expected empty kv hint:\n%s", v)
	}
}

func TestRenderInlineKVEditor(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://x",
			Headers: []domain.KV{{Key: "Content-Type", Value: "application/json"}}},
	}}
	m := seededModel(t, col)
	ws(m).focus = paneEditor
	send(t, m, keyMsg("h")) // Headers tab
	v := renderTop(t, m)
	if !strings.Contains(v, "Content-Type") || !strings.Contains(v, "application/json") {
		t.Fatalf("inline headers not rendered in pane:\n%s", v)
	}
}

func TestRenderConfirmScreen(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	send(t, m, keyMsg("d")) // push confirm
	cs, ok := m.top().(*confirmScreen)
	if !ok {
		t.Fatalf("expected confirm screen, got %T", m.top())
	}
	v := renderTop(t, m)
	if !strings.Contains(v, "Delete collection") || !strings.Contains(v, "y = yes") {
		t.Fatalf("confirm view missing content:\n%s", v)
	}
	if cs.Title() != "Confirm" || cs.Crumb() != "" {
		t.Fatalf("confirm Title/Crumb wrong: %q / %q", cs.Title(), cs.Crumb())
	}
	if len(cs.HelpBindings()) == 0 {
		t.Fatal("confirm should advertise help bindings")
	}
}

func TestSimpleItemFilterValue(t *testing.T) {
	it := simpleItem{id: "i", title: "Hello", desc: "world"}
	if it.FilterValue() != "Hello" {
		t.Fatalf("FilterValue = %q, want title", it.FilterValue())
	}
}

func TestModelInitAndDepth(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	if m.depth() != 1 {
		t.Fatalf("depth = %d, want 1", m.depth())
	}
	send(t, m, keyMsg("E")) // open environments overlay
	if m.depth() != 2 {
		t.Fatalf("depth after open = %d, want 2", m.depth())
	}
}
