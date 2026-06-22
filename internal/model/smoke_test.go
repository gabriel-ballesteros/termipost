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

// key builds a KeyMsg for a key name or a run of runes.
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
	root := New(NewApp(s, data), data.LoadErrors)
	send(t, root, tea.WindowSizeMsg{Width: 80, Height: 24})
	return root
}

// TestKeyboardWorkflow exercises the core flow with the keyboard only:
// create a collection, add a request, edit it, attach an assertion.
func TestKeyboardWorkflow(t *testing.T) {
	m := newTestModel(t)

	// Create a collection named "Demo".
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("Demo"))
	send(t, m, keyMsg("enter"))
	if len(m.app.collections) != 1 || m.app.collections[0].Name != "Demo" {
		t.Fatalf("collection not created: %+v", m.app.collections)
	}

	// Open it and create a request.
	send(t, m, keyMsg("enter")) // open collection -> request list
	send(t, m, keyMsg("n"))     // new request prompt
	send(t, m, keyMsg("Get users"))
	send(t, m, keyMsg("enter")) // creates request, opens editor
	col := m.app.collections[0]
	if len(col.Requests) != 1 || col.Requests[0].Name != "Get users" {
		t.Fatalf("request not created: %+v", col.Requests)
	}

	// In the editor, navigate to URL, edit it.
	send(t, m, keyMsg("tab")) // Name -> Method
	send(t, m, keyMsg("tab")) // Method -> URL
	send(t, m, keyMsg("enter"))
	send(t, m, keyMsg("https://example.com/users"))
	send(t, m, keyMsg("enter"))  // commit URL
	send(t, m, keyMsg("ctrl+s")) // explicit save
	if got := m.app.collections[0].Requests[0].URL; got != "https://example.com/users" {
		t.Fatalf("URL not saved, got %q", got)
	}

	// Add an assertion (defaults to status 200) via the assertions screen.
	send(t, m, keyMsg("a"))      // open assertions
	send(t, m, keyMsg("a"))      // add -> assertion edit form (default status 200)
	send(t, m, keyMsg("ctrl+s")) // save assertion
	if n := len(m.app.collections[0].Requests[0].Assertions); n != 1 {
		t.Fatalf("expected 1 assertion, got %d", n)
	}
}

// TestResizeKeepsHelpBar verifies the layout adapts to terminal resizes and the
// action/help bar stays visible.
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

	send(t, m, keyMsg("e")) // environments
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

	// Reveal and confirm the value appears.
	send(t, m, keyMsg("v"))
	if revealed := m.View(); !strings.Contains(revealed, "s3cr3t") {
		t.Fatalf("secret not revealed after toggle:\n%s", revealed)
	}
}

// TestResponseCopyBody verifies the response screen stores the raw body, exposes
// a copy binding, and handles the copy key (success or a clean error status).
func TestResponseCopyBody(t *testing.T) {
	s := store.New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	data, _ := s.Load()
	app := NewApp(s, data)

	resp := &httpclient.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    http.Header{},
		Body:       []byte(`{"x":1}`),
	}
	scr := newResponseScreen(app, domain.Request{Method: domain.GET, URL: "https://example.com"}, resp)

	if string(scr.body) != `{"x":1}` {
		t.Fatalf("raw body not stored, got %q", scr.body)
	}
	hasCopy := false
	for _, b := range scr.HelpBindings() {
		if b.Help().Key == "c" {
			hasCopy = true
		}
	}
	if !hasCopy {
		t.Fatal("copy binding not advertised in help")
	}

	m := New(app, nil)
	m.push(scr)
	send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	send(t, m, keyMsg("c"))
	if m.status == "" {
		t.Fatal("copy key produced no status feedback")
	}
}

// TestRequestListShowsItemsOnOpen guards the bug where a screen opened after
// startup was never sized, so its list rendered only a pagination footer. After
// opening a collection, the request rows must be visible.
func TestRequestListShowsItemsOnOpen(t *testing.T) {
	s := store.New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	col := domain.Collection{ID: "c-1", Name: "Demo", Requests: []domain.Request{
		{ID: "r-1", Name: "List users", Method: domain.GET, URL: "https://example.com/users"},
		{ID: "r-2", Name: "Create user", Method: domain.POST, URL: "https://example.com/users"},
	}}
	if err := s.SaveCollection(col); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, _ := s.Load()
	m := New(NewApp(s, data), nil)

	// Size the root, then open the collection (push the request-list screen).
	send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	send(t, m, keyMsg("enter"))

	view := m.View()
	for _, name := range []string{"List users", "Create user"} {
		if !strings.Contains(view, name) {
			t.Fatalf("request list missing %q (only pagination?):\n%s", name, view)
		}
	}
}
