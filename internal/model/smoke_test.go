package model

import (
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

// keyMsg builds a KeyMsg for a key name or a run of runes.
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func send(t *testing.T, m *Model, msg tea.Msg) {
	t.Helper()
	if _, cmd := m.Update(msg); cmd != nil {
		// Drain a single round of commands so async-free flows settle.
		if out := cmd(); out != nil {
			m.Update(out)
		}
	}
}

func newTestModel(t *testing.T) *Model {
	t.Helper()
	s := store.New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	data, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	root := New(NewApp(s, data), "dev", data.LoadErrors)
	send(t, root, tea.WindowSizeMsg{Width: 100, Height: 30})
	return root
}

// TestKeyboardWorkflow exercises the core flow with the keyboard only: create a
// collection, add a request, edit its URL, attach an assertion.
func TestKeyboardWorkflow(t *testing.T) {
	m := newTestModel(t)

	// Create a collection named "Demo" (tree focused at launch).
	send(t, m, keyMsg("N"))
	send(t, m, keyMsg("Demo"))
	send(t, m, keyMsg("enter"))
	if len(m.app.collections) != 1 || m.app.collections[0].Name != "Demo" {
		t.Fatalf("collection not created: %+v", m.app.collections)
	}

	// Add a request; creating it loads the editor pane.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("Get users"))
	send(t, m, keyMsg("enter"))
	col := m.app.collections[0]
	if len(col.Requests) != 1 || col.Requests[0].Name != "Get users" {
		t.Fatalf("request not created: %+v", col.Requests)
	}

	// Edit the URL in the editor pane and save.
	send(t, m, keyMsg("u"))     // jump to URL field
	send(t, m, keyMsg("enter")) // edit
	send(t, m, keyMsg("https://example.com/users"))
	send(t, m, keyMsg("enter"))  // commit
	send(t, m, keyMsg("ctrl+s")) // save
	if got := m.app.collections[0].Requests[0].URL; got != "https://example.com/users" {
		t.Fatalf("URL not saved, got %q", got)
	}

	// Add an assertion (defaults to status 200) via the assertions overlay.
	send(t, m, keyMsg("a"))      // open assertions
	send(t, m, keyMsg("a"))      // add -> assertion edit form
	send(t, m, keyMsg("ctrl+s")) // save assertion
	if n := len(m.app.collections[0].Requests[0].Assertions); n != 1 {
		t.Fatalf("expected 1 assertion, got %d", n)
	}
}

// TestResizeKeepsHelpBar verifies the layout adapts to resizes (including the
// single-pane fallback) and the action/help bar stays visible.
func TestResizeKeepsHelpBar(t *testing.T) {
	m := newTestModel(t)
	for _, sz := range []tea.WindowSizeMsg{{Width: 120, Height: 40}, {Width: 60, Height: 12}} {
		send(t, m, sz)
		view := m.View()
		if !strings.Contains(view, "quit") {
			t.Fatalf("help bar missing at %dx%d:\n%s", sz.Width, sz.Height, view)
		}
		if lines := strings.Count(view, "\n") + 1; lines > sz.Height {
			t.Fatalf("view height %d exceeds terminal height %d", lines, sz.Height)
		}
	}
}

// TestSecretsMasking verifies secrets are masked in the view until revealed.
func TestSecretsMasking(t *testing.T) {
	m := newTestModel(t)

	send(t, m, keyMsg("E")) // environments
	send(t, m, keyMsg("s")) // secrets editor
	send(t, m, keyMsg("a")) // add secret
	send(t, m, keyMsg("token: s3cr3t"))
	send(t, m, keyMsg("enter"))

	if m.app.secrets["token"] != "s3cr3t" {
		t.Fatalf("secret not stored: %+v", m.app.secrets)
	}

	view := m.View()
	if strings.Contains(view, "s3cr3t") {
		t.Fatalf("secret leaked into masked view:\n%s", view)
	}
	if !strings.Contains(view, "••••••") {
		t.Fatalf("expected mask in view:\n%s", view)
	}

	send(t, m, keyMsg("v"))
	if revealed := m.View(); !strings.Contains(revealed, "s3cr3t") {
		t.Fatalf("secret not revealed after toggle:\n%s", revealed)
	}
}

// TestResponseCopyBody verifies the response pane stores the raw body, exposes a
// copy binding, and handles the copy key.
func TestResponseCopyBody(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://example.com"},
	}}
	m := seededModel(t, col)
	w := ws(m)
	resp := &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{}, Body: []byte(`{"x":1}`)}
	w.Update(m, sendResultMsg{resp: resp})

	if string(w.response.body) != `{"x":1}` {
		t.Fatalf("raw body not stored, got %q", w.response.body)
	}
	hasCopy := false
	for _, b := range w.response.HelpBindings() {
		if b.Help().Key == "c" {
			hasCopy = true
		}
	}
	if !hasCopy {
		t.Fatal("copy binding not advertised in help")
	}

	send(t, m, keyMsg("c")) // response pane is focused after a send result
	if m.status == "" {
		t.Fatal("copy key produced no status feedback")
	}
}

// TestTreeShowsItemsOnOpen verifies the tree pane renders its request rows at
// startup, not just a header.
func TestTreeShowsItemsOnOpen(t *testing.T) {
	col := domain.Collection{ID: "c-1", Name: "Demo", Requests: []domain.Request{
		{ID: "r-1", Name: "List users", Method: domain.GET, URL: "https://example.com/users"},
		{ID: "r-2", Name: "Create user", Method: domain.POST, URL: "https://example.com/users"},
	}}
	m := seededModel(t, col)
	view := m.View()
	for _, name := range []string{"List users", "Create user"} {
		if !strings.Contains(view, name) {
			t.Fatalf("tree missing %q:\n%s", name, view)
		}
	}
}

// editorModel seeds one collection+request with the editor pane focused.
func editorModel(t *testing.T) *Model {
	col := domain.Collection{ID: "c-1", Name: "Demo", Requests: []domain.Request{
		{ID: "r-1", Name: "Get", Method: domain.GET, URL: "https://example.com"},
	}}
	m := seededModel(t, col)
	ws(m).focus = paneEditor
	return m
}

// TestEditorFieldShortcuts verifies first-letter shortcuts switch tabs/fields,
// and that they are inert while editing a text field.
func TestEditorFieldShortcuts(t *testing.T) {
	m := editorModel(t)
	ed := ws(m).editor

	send(t, m, keyMsg("h")) // Headers tab
	if ed.tab != etHeaders {
		t.Fatalf("`h` did not switch to Headers tab, got %d", ed.tab)
	}
	send(t, m, keyMsg("p")) // Query tab
	if ed.tab != etQuery {
		t.Fatalf("`p` did not switch to Query tab, got %d", ed.tab)
	}
	send(t, m, keyMsg("b")) // Body tab
	if ed.tab != etBody {
		t.Fatalf("`b` did not switch to Body tab, got %d", ed.tab)
	}

	// n -> Request tab, Name field, enter edit mode.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("enter"))
	if !ed.editFld || ed.tab != etRequest || ed.fieldIx != efName {
		t.Fatalf("`n`+enter did not edit Name (editing=%v tab=%d field=%d)", ed.editFld, ed.tab, ed.fieldIx)
	}
	// While editing, a shortcut letter types into the field, not switch tabs.
	send(t, m, keyMsg("p"))
	if ed.tab != etRequest {
		t.Fatal("shortcut fired while editing a text field")
	}
	if got := ed.name.Value(); got[len(got)-1] != 'p' {
		t.Fatalf("expected typed 'p' appended to name, got %q", got)
	}
}

// TestEditorMethodCyclesWithArrows verifies the method cycles with arrow keys.
func TestEditorMethodCyclesWithArrows(t *testing.T) {
	m := editorModel(t)
	send(t, m, keyMsg("m")) // focus Method
	ed := ws(m).editor
	if ed.fieldIx != efMethod {
		t.Fatalf("`m` did not focus Method, field=%d", ed.fieldIx)
	}
	before := ed.methodIdx
	send(t, m, tea.KeyMsg{Type: tea.KeyRight})
	if ed.methodIdx == before {
		t.Fatal("right arrow did not cycle the method")
	}
}

// TestEditorOpensAssertions verifies `a` opens the assertions overlay.
func TestEditorOpensAssertions(t *testing.T) {
	m := editorModel(t)
	send(t, m, keyMsg("a"))
	if _, ok := m.top().(*assertionsScreen); !ok {
		t.Fatalf("`a` did not open assertions, got %T", m.top())
	}
}

// TestAssertionEditorSkipsHiddenTarget guards the bug where ↓ from Operator
// landed on the hidden Target field.
func TestAssertionEditorSkipsHiddenTarget(t *testing.T) {
	a := domain.Assertion{Kind: domain.AssertLatency, Op: domain.OpMaxMS, Expected: "100"}
	s := newAssertionEditScreen(a, nil)
	s.focus = aOp
	s.Update(&Model{}, tea.KeyMsg{Type: tea.KeyDown})
	if s.focus != aExpected {
		t.Fatalf("expected focus on aExpected (max-ms) after one ↓, got %d", s.focus)
	}
}

// TestTitleShowsVersion verifies the version (and dev fallback) in the title.
func TestTitleShowsVersion(t *testing.T) {
	mk := func(version string) string {
		s := store.New(t.TempDir())
		if err := s.Init(); err != nil {
			t.Fatalf("init: %v", err)
		}
		data, _ := s.Load()
		m := New(NewApp(s, data), version, nil)
		send(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
		return m.View()
	}
	if v := mk("1.2.3"); !strings.Contains(v, "termipost v1.2.3") {
		t.Fatalf("expected 'termipost v1.2.3':\n%s", v)
	}
	if v := mk("dev"); !strings.Contains(v, "termipost (dev)") {
		t.Fatalf("expected 'termipost (dev)':\n%s", v)
	}
	if v := mk("v2.0.0"); !strings.Contains(v, "termipost v2.0.0") || strings.Contains(v, "vv2.0.0") {
		t.Fatalf("expected 'termipost v2.0.0' without double v:\n%s", v)
	}
}

// TestBreadcrumbReflectsStack verifies the breadcrumb tracks the workspace and
// open overlays, and omits transient prompts.
func TestBreadcrumbReflectsStack(t *testing.T) {
	m := editorModel(t)
	if v := m.View(); !strings.Contains(v, "Workspace") {
		t.Fatalf("breadcrumb missing Workspace:\n%s", v)
	}

	send(t, m, keyMsg("E")) // open environments (a real crumb)
	v := m.View()
	if !strings.Contains(v, "Environments") {
		t.Fatalf("expected Environments crumb:\n%s", v)
	}
	send(t, m, keyMsg("n")) // env new -> prompt overlay
	if v := m.View(); strings.Contains(v, "Input") {
		t.Fatalf("prompt overlay should be omitted from breadcrumb:\n%s", v)
	}
}
