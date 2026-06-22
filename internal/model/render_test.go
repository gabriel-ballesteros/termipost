package model

import (
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
)

func respFixture() *httpclient.Response {
	return &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{}, Body: []byte("{}")}
}

// renderTop sizes the model and returns the rendered view of the top screen,
// exercising its View/Title/HelpBindings/Crumb path.
func renderTop(t *testing.T, m *Model) string {
	t.Helper()
	send(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	return m.View()
}

func TestRenderAssertionsScreen(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://x",
			Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}},
	}}
	m := seededModel(t, col)
	openRequests(t, m)
	send(t, m, keyMsg("enter")) // editor
	send(t, m, keyMsg("a"))     // assertions screen
	v := renderTop(t, m)
	if !strings.Contains(v, "Assertions") || !strings.Contains(v, "status code == 200") {
		t.Fatalf("assertions screen view missing content:\n%s", v)
	}

	// Empty-state branch.
	m2 := seededModel(t, domain.Collection{ID: "c2", Name: "B", Requests: []domain.Request{{ID: "r9", Name: "n", Method: domain.GET}}})
	openRequests(t, m2)
	send(t, m2, keyMsg("enter"))
	send(t, m2, keyMsg("a"))
	if v := renderTop(t, m2); !strings.Contains(v, "No assertions") {
		t.Fatalf("expected empty assertions hint:\n%s", v)
	}
}

func TestRenderAssertionEditScreen(t *testing.T) {
	// Header kind shows a Target row; latency shows "Max ms".
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

	// Empty-state branch.
	m2 := seededModel(t)
	m2.push(newKVEditorScreen("Params", nil, nil))
	if v := renderTop(t, m2); !strings.Contains(v, "No entries") {
		t.Fatalf("expected empty kv hint:\n%s", v)
	}
}

func TestRenderResponseScreen(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://example.com"},
	}}
	m := seededModel(t, col)
	openRequests(t, m)
	send(t, m, keyMsg("enter")) // editor
	ed := m.top().(*requestEditScreen)
	ed.Update(m, sendResultMsg{resp: respFixture()})
	if v := renderTop(t, m); !strings.Contains(v, "Response") {
		t.Fatalf("response screen view missing title:\n%s", v)
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
	if m.Init() != nil {
		// collection list Init returns nil; just ensure it does not panic.
		t.Log("Init returned a cmd")
	}
	if m.depth() != 1 {
		t.Fatalf("depth = %d, want 1", m.depth())
	}
	openRequests(t, m)
	if m.depth() != 2 {
		t.Fatalf("depth after open = %d, want 2", m.depth())
	}
}
