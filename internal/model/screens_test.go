package model

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

// seededModel builds a root Model whose store already holds cols.
func seededModel(t *testing.T, cols ...domain.Collection) *Model {
	t.Helper()
	s := store.New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, c := range cols {
		if err := s.SaveCollection(c); err != nil {
			t.Fatalf("seed save: %v", err)
		}
	}
	data, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	m := New(NewApp(s, data), "dev", nil)
	send(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	return m
}

// ws returns the workspace screen at the bottom of the stack.
func ws(m *Model) *workspaceScreen { return m.stack[0].(*workspaceScreen) }

// errString is a minimal error for injecting send failures.
type errString string

func (e errString) Error() string { return string(e) }

// ---- tree pane: collection CRUD ----

func TestCreateCollectionValidation(t *testing.T) {
	m := seededModel(t)

	// Empty name is rejected.
	send(t, m, keyMsg("N"))
	send(t, m, keyMsg("enter"))
	if !m.statusErr {
		t.Fatal("empty collection name should set an error")
	}
	if len(m.app.collections) != 0 {
		t.Fatalf("no collection should be created, got %+v", m.app.collections)
	}

	// Valid name creates one.
	send(t, m, keyMsg("N"))
	send(t, m, keyMsg("API"))
	send(t, m, keyMsg("enter"))
	if len(m.app.collections) != 1 || m.app.collections[0].Name != "API" {
		t.Fatalf("collection not created: %+v", m.app.collections)
	}

	// Duplicate name is rejected.
	send(t, m, keyMsg("N"))
	send(t, m, keyMsg("API"))
	send(t, m, keyMsg("enter"))
	if !m.statusErr {
		t.Fatal("duplicate collection name should set an error")
	}
	if len(m.app.collections) != 1 {
		t.Fatalf("duplicate must not be added: %+v", m.app.collections)
	}
}

func TestRenameCollection(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "Old"})
	send(t, m, keyMsg("r"))
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("New"))
	send(t, m, keyMsg("enter"))
	if got := m.app.findCollection("c1").Name; got != "New" {
		t.Fatalf("rename failed, name = %q", got)
	}
}

func TestDeleteCollectionConfirmed(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "Doomed"})
	send(t, m, keyMsg("d"))
	if _, ok := m.top().(*confirmScreen); !ok {
		t.Fatalf("d should push a confirm screen, got %T", m.top())
	}
	send(t, m, keyMsg("y"))
	if len(m.app.collections) != 0 {
		t.Fatalf("collection not deleted: %+v", m.app.collections)
	}
}

func TestDeleteCollectionCancelled(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "Keep"})
	send(t, m, keyMsg("d"))
	send(t, m, keyMsg("n")) // decline
	if len(m.app.collections) != 1 {
		t.Fatalf("collection should survive a declined delete: %+v", m.app.collections)
	}
}

func TestRunCollectionShowsResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	col := domain.Collection{ID: "c1", Name: "Suite", Requests: []domain.Request{
		{ID: "r1", Name: "ping", Method: domain.GET, URL: srv.URL,
			Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}},
	}}
	m := seededModel(t, col)
	// Cursor starts on the collection row; x runs the whole collection.
	send(t, m, keyMsg("x"))
	if _, ok := m.top().(*runResultsScreen); !ok {
		t.Fatalf("running a collection should push run results, got %T", m.top())
	}
	if !strings.Contains(m.View(), "1 passed") {
		t.Fatalf("results should report 1 passed:\n%s", m.View())
	}
}

// ---- tree pane: request CRUD ----

// withRequest seeds one collection+request and focuses the editor on it.
func withRequest(t *testing.T) *Model {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://example.com"},
	}}
	m := seededModel(t, col)
	// Constructor loads the first request; make sure the editor is focused.
	ws(m).focus = paneEditor
	return m
}

func TestCreateRequestLoadsEditor(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	// Tree focused, cursor on the collection.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("List"))
	send(t, m, keyMsg("enter"))

	col := m.app.findCollection("c1")
	if len(col.Requests) != 1 || col.Requests[0].Name != "List" {
		t.Fatalf("request not created: %+v", col.Requests)
	}
	if col.Requests[0].Method != domain.GET {
		t.Fatalf("new request should default to GET, got %s", col.Requests[0].Method)
	}
	w := ws(m)
	if !w.editor.loaded || w.editor.reqID != col.Requests[0].ID {
		t.Fatalf("new request should be loaded in the editor, got %q", w.editor.reqID)
	}
	if w.focus != paneEditor {
		t.Fatalf("focus should move to the editor, got %d", w.focus)
	}
}

func TestCreateRequestEmptyRejected(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("enter")) // empty name
	if !m.statusErr {
		t.Fatal("empty request name should error")
	}
	if len(m.app.findCollection("c1").Requests) != 0 {
		t.Fatal("no request should be created")
	}
}

func TestDeleteRequestConfirmed(t *testing.T) {
	m := withRequest(t)
	// Move focus to the tree and select the request row (collection is row 0).
	ws(m).focus = paneTree
	ws(m).tree.cursor = 1
	send(t, m, keyMsg("d"))
	send(t, m, keyMsg("y"))
	if n := len(m.app.findCollection("c1").Requests); n != 0 {
		t.Fatalf("request not deleted, %d remain", n)
	}
}

// ---- editor pane ----

func TestEditorPersistsURL(t *testing.T) {
	m := withRequest(t)
	send(t, m, keyMsg("u")) // jump to URL field
	send(t, m, keyMsg("enter"))
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("https://api.test/v1"))
	send(t, m, keyMsg("enter"))  // commit
	send(t, m, keyMsg("ctrl+s")) // persist
	if got := m.app.findCollection("c1").Requests[0].URL; got != "https://api.test/v1" {
		t.Fatalf("URL not persisted, got %q", got)
	}
	if m.status != "Saved" {
		t.Fatalf("expected Saved status, got %q", m.status)
	}
}

func TestEditorMethodCycle(t *testing.T) {
	m := withRequest(t)
	ed := ws(m).editor
	send(t, m, keyMsg("m"))     // focus Method
	send(t, m, keyMsg("enter")) // Enter cycles method
	if domain.Methods[ed.methodIdx] == domain.GET {
		t.Fatalf("Enter on Method should advance from GET, methodIdx=%d", ed.methodIdx)
	}
}

func TestEditorTestWithoutAssertionsErrors(t *testing.T) {
	m := withRequest(t)
	send(t, m, keyMsg("T")) // test, but request has no assertions
	if !m.statusErr {
		t.Fatal("testing a request with no assertions should error")
	}
}

func TestSendResultPopulatesResponsePane(t *testing.T) {
	m := withRequest(t)
	resp := &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{}, Body: []byte("{}")}
	w := ws(m)
	w.Update(m, sendResultMsg{resp: resp})
	if w.response.resp != resp {
		t.Fatal("send result should populate the response pane")
	}
	if w.focus != paneResponse {
		t.Fatalf("focus should move to the response pane, got %d", w.focus)
	}
}

func TestSendErrorSetsError(t *testing.T) {
	m := withRequest(t)
	ws(m).Update(m, sendResultMsg{err: errString("boom")})
	if !m.statusErr || !strings.Contains(m.status, "boom") {
		t.Fatalf("send error not surfaced: status=%q err=%v", m.status, m.statusErr)
	}
}

func TestRunResultPushesResults(t *testing.T) {
	m := withRequest(t)
	ws(m).Update(m, reqRunMsg{result: domain.RunResult{RequestName: "Get", Status: domain.RunPassed}})
	if _, ok := m.top().(*runResultsScreen); !ok {
		t.Fatalf("a run result should push the results screen, got %T", m.top())
	}
}

// ---- assertions (overlay launched from the editor) ----

func openAssertions(t *testing.T) *Model {
	m := withRequest(t)
	send(t, m, keyMsg("a")) // editor -> assertions screen
	if _, ok := m.top().(*assertionsScreen); !ok {
		t.Fatalf("expected assertions screen, got %T", m.top())
	}
	return m
}

func TestAssertionAddAndDelete(t *testing.T) {
	m := openAssertions(t)

	send(t, m, keyMsg("a"))      // add default status-200 assertion
	send(t, m, keyMsg("ctrl+s")) // save it
	if n := len(m.app.findCollection("c1").Requests[0].Assertions); n != 1 {
		t.Fatalf("expected 1 assertion after add, got %d", n)
	}

	send(t, m, keyMsg("d"))
	if n := len(m.app.findCollection("c1").Requests[0].Assertions); n != 0 {
		t.Fatalf("expected 0 assertions after delete, got %d", n)
	}
}

func TestAssertionEditExisting(t *testing.T) {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://example.com",
			Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}}},
	}}
	m := seededModel(t, col)
	ws(m).focus = paneEditor
	send(t, m, keyMsg("a"))     // assertions
	send(t, m, keyMsg("enter")) // edit assertion 0
	ed, ok := m.top().(*assertionEditScreen)
	if !ok {
		t.Fatalf("expected assertion editor, got %T", m.top())
	}
	ed.focus = aExpected
	send(t, m, keyMsg("enter")) // begin editing expected
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("404"))
	send(t, m, keyMsg("esc"))    // stop editing
	send(t, m, keyMsg("ctrl+s")) // save
	if got := m.app.findCollection("c1").Requests[0].Assertions[0].Expected; got != "404" {
		t.Fatalf("assertion not updated, Expected = %q", got)
	}
}

func TestAssertionEmptyExpectedRejected(t *testing.T) {
	m := openAssertions(t)
	send(t, m, keyMsg("a")) // add
	ed := m.top().(*assertionEditScreen)
	ed.expected.SetValue("") // clear default 200
	send(t, m, keyMsg("ctrl+s"))
	if !m.statusErr {
		t.Fatal("empty Expected should be rejected")
	}
	if _, ok := m.top().(*assertionEditScreen); !ok {
		t.Fatalf("should remain on assertion editor, got %T", m.top())
	}
}

func TestAssertionEditorKindCycleResetsOp(t *testing.T) {
	s := newAssertionEditScreen(domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}, nil)
	s.focus = aKind
	s.cycle(true) // status_code -> header
	if s.a.Kind != domain.AssertHeader {
		t.Fatalf("kind = %q, want header", s.a.Kind)
	}
	if s.a.Op != opsFor(domain.AssertHeader)[0] {
		t.Fatalf("op not reset to first valid op for new kind: %q", s.a.Op)
	}
}

// ---- response rendering helpers ----

func TestPrettyBodyIndentsJSON(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	resp := &httpclient.Response{Headers: h, Body: []byte(`{"error":"nope"}`)}
	out := prettyBody(resp)
	for _, want := range []string{"error", "nope", "\n  "} {
		if !strings.Contains(out, want) {
			t.Errorf("pretty body missing %q:\n%s", want, out)
		}
	}
}

func TestLooksLikeJSON(t *testing.T) {
	if !looksLikeJSON([]byte(`  {"a":1}`)) || !looksLikeJSON([]byte(`[1,2]`)) {
		t.Error("object/array should look like JSON")
	}
	if looksLikeJSON([]byte("plain text")) || looksLikeJSON([]byte("")) {
		t.Error("non-JSON should not look like JSON")
	}
}

func TestPrettyBodyNonJSON(t *testing.T) {
	resp := &httpclient.Response{Headers: http.Header{}, Body: []byte("just text")}
	if got := prettyBody(resp); got != "just text" {
		t.Fatalf("non-JSON body changed: %q", got)
	}
}

// ---- body prettify + live validation (payload-formatting) ----

// editorOf seeds a model with one request and returns its loaded editor pane.
func editorOf(t *testing.T, body string) (*Model, *editorPane) {
	t.Helper()
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "First", Method: domain.POST, URL: "https://a", Body: body},
	}}
	m := seededModel(t, col)
	w := ws(m)
	w.selectRequest(m, "c1", "r1")
	return m, w.editor
}

func TestPrettifyBodyFormatsValidJSON(t *testing.T) {
	m, ed := editorOf(t, `{"a":1,"b":[2,3]}`)
	ed.tab = etBody
	ed.prettifyBody(m)
	got := ed.body.Value()
	if !strings.Contains(got, "\n  \"a\": 1") {
		t.Fatalf("body not prettified:\n%s", got)
	}
	if !ed.dirty() {
		t.Error("prettify changed the body but request is not marked dirty")
	}
}

func TestPrettifyBodyInvalidLeavesUnchanged(t *testing.T) {
	const bad = `{"a": }`
	m, ed := editorOf(t, bad)
	ed.tab = etBody
	ed.prettifyBody(m)
	if ed.body.Value() != bad {
		t.Errorf("body mutated on invalid JSON: %q", ed.body.Value())
	}
	if !m.statusErr || !strings.Contains(m.status, "Invalid JSON") {
		t.Errorf("expected an Invalid JSON error, got status=%q err=%v", m.status, m.statusErr)
	}
}

func TestBodyValidityIndicator(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string // substring; "" means no indicator
	}{
		{"valid", `{"a":1}`, "valid JSON"},
		{"invalid", `{"a":`, "✗"},
		{"non-json", "plain text", ""},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ed := editorOf(t, tc.body)
			got := ed.bodyValidity()
			if tc.want == "" {
				if got != "" {
					t.Errorf("bodyValidity = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("bodyValidity = %q, want substring %q", got, tc.want)
			}
		})
	}
}

// ---- run results screen rendering ----

func TestRenderRunResultStatuses(t *testing.T) {
	cases := []struct {
		r    domain.RunResult
		want string
	}{
		{domain.RunResult{RequestName: "ok", Status: domain.RunPassed, StatusCode: 200}, "PASS"},
		{domain.RunResult{RequestName: "bad", Status: domain.RunFailed, StatusCode: 500}, "FAIL"},
		{domain.RunResult{RequestName: "none", Status: domain.RunSkipped}, "SKIP"},
		{domain.RunResult{RequestName: "boom", Status: domain.RunError, Err: "dial fail"}, "ERROR"},
	}
	for _, c := range cases {
		out := renderRunResult(c.r)
		if !strings.Contains(out, c.want) || !strings.Contains(out, c.r.RequestName) {
			t.Errorf("renderRunResult(%s) missing %q or name:\n%s", c.r.Status, c.want, out)
		}
	}
	if !strings.Contains(renderRunResult(cases[2].r), "no assertions") {
		t.Error("skipped result should say 'no assertions'")
	}
	if !strings.Contains(renderRunResult(cases[3].r), "dial fail") {
		t.Error("error result should include the error message")
	}
}

func TestNewRunResultsScreenSummary(t *testing.T) {
	res := domain.CollectionRunResult{
		Passed: 2, Failed: 1, Skipped: 3,
		Results: []domain.RunResult{
			{RequestName: "a", Status: domain.RunPassed, Assertions: []domain.AssertionResult{{Passed: true, Detail: "ok"}}},
		},
	}
	scr := newRunResultsScreen("Suite", res)
	for _, want := range []string{"Suite", "2 passed", "1 failed", "3 skipped", "a"} {
		if !strings.Contains(scr.content, want) {
			t.Errorf("results summary missing %q:\n%s", want, scr.content)
		}
	}
	if scr.Title() != "Run results" {
		t.Errorf("title = %q", scr.Title())
	}
}

func TestNewSingleRunResultsScreen(t *testing.T) {
	scr := newSingleRunResultsScreen(domain.RunResult{RequestName: "solo", Status: domain.RunPassed})
	if scr.Title() != "Test result" {
		t.Errorf("title = %q, want Test result", scr.Title())
	}
	if !strings.Contains(scr.content, "solo") {
		t.Errorf("single result missing request name:\n%s", scr.content)
	}
}

func TestRunResultsBackPops(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	m.push(newSingleRunResultsScreen(domain.RunResult{RequestName: "x", Status: domain.RunPassed}))
	send(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	send(t, m, keyMsg("esc"))
	if _, ok := m.top().(*runResultsScreen); ok {
		t.Fatal("esc should pop the run results screen")
	}
}
