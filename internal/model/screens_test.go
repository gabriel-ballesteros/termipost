package model

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

// ---- collections screen ----

func TestCreateCollectionValidation(t *testing.T) {
	m := seededModel(t)

	// Empty name is rejected.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("enter"))
	if !m.statusErr {
		t.Fatal("empty collection name should set an error")
	}
	if len(m.app.collections) != 0 {
		t.Fatalf("no collection should be created, got %+v", m.app.collections)
	}

	// Valid name creates one.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("API"))
	send(t, m, keyMsg("enter"))
	if len(m.app.collections) != 1 || m.app.collections[0].Name != "API" {
		t.Fatalf("collection not created: %+v", m.app.collections)
	}

	// Duplicate name is rejected.
	send(t, m, keyMsg("n"))
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
	// Clear the prefilled value then type a new name.
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
	send(t, m, keyMsg("R")) // run all; send() drains the network round
	if _, ok := m.top().(*runResultsScreen); !ok {
		t.Fatalf("running a collection should push run results, got %T", m.top())
	}
	if !strings.Contains(m.View(), "1 passed") {
		t.Fatalf("results should report 1 passed:\n%s", m.View())
	}
}

// ---- requests screen ----

func openRequests(t *testing.T, m *Model) *requestListScreen {
	t.Helper()
	send(t, m, keyMsg("enter")) // open the selected collection
	rs, ok := m.top().(*requestListScreen)
	if !ok {
		t.Fatalf("enter should open the request list, got %T", m.top())
	}
	return rs
}

func TestCreateRequestOpensEditor(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	openRequests(t, m)

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
	if _, ok := m.top().(*requestEditScreen); !ok {
		t.Fatalf("creating a request should open its editor, got %T", m.top())
	}
}

func TestCreateRequestEmptyRejected(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	openRequests(t, m)
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
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "one", Method: domain.GET},
	}}
	m := seededModel(t, col)
	openRequests(t, m)
	send(t, m, keyMsg("d"))
	send(t, m, keyMsg("y"))
	if n := len(m.app.findCollection("c1").Requests); n != 0 {
		t.Fatalf("request not deleted, %d remain", n)
	}
}

func TestRequestsBackPopsToCollections(t *testing.T) {
	m := seededModel(t, domain.Collection{ID: "c1", Name: "API"})
	openRequests(t, m)
	send(t, m, keyMsg("esc"))
	if _, ok := m.top().(*collectionListScreen); !ok {
		t.Fatalf("esc from requests should return to collections, got %T", m.top())
	}
}

// ---- request editor ----

func openEditor(t *testing.T) *Model {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "Get", Method: domain.GET, URL: "https://example.com"},
	}}
	m := seededModel(t, col)
	openRequests(t, m)
	send(t, m, keyMsg("enter")) // open editor
	if _, ok := m.top().(*requestEditScreen); !ok {
		t.Fatalf("expected editor, got %T", m.top())
	}
	return m
}

func TestEditorPersistsURL(t *testing.T) {
	m := openEditor(t)
	send(t, m, keyMsg("u")) // focus + edit URL
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
	m := openEditor(t)
	ed := m.top().(*requestEditScreen)
	send(t, m, keyMsg("m"))     // focus Method
	send(t, m, keyMsg("enter")) // Enter cycles method (advances methodIdx)
	if domain.Methods[ed.methodIdx] == domain.GET {
		t.Fatalf("Enter on Method should advance from GET, methodIdx=%d", ed.methodIdx)
	}
}

func TestEditorTestWithoutAssertionsErrors(t *testing.T) {
	m := openEditor(t)
	send(t, m, keyMsg("T")) // test, but request has no assertions
	if !m.statusErr {
		t.Fatal("testing a request with no assertions should error")
	}
}

func TestEditorSendResultPushesResponse(t *testing.T) {
	m := openEditor(t)
	ed := m.top().(*requestEditScreen)
	resp := &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{}, Body: []byte("{}")}
	ed.Update(m, sendResultMsg{resp: resp})
	// push() clears the transient status, so we assert the navigation, not status.
	if _, ok := m.top().(*responseScreen); !ok {
		t.Fatalf("a successful send result should push the response screen, got %T", m.top())
	}
}

func TestEditorSendErrorSetsError(t *testing.T) {
	m := openEditor(t)
	ed := m.top().(*requestEditScreen)
	ed.Update(m, sendResultMsg{err: errString("boom")})
	if !m.statusErr || !strings.Contains(m.status, "boom") {
		t.Fatalf("send error not surfaced: status=%q err=%v", m.status, m.statusErr)
	}
}

func TestEditorRunResultPushesResults(t *testing.T) {
	m := openEditor(t)
	ed := m.top().(*requestEditScreen)
	ed.Update(m, reqRunMsg{result: domain.RunResult{RequestName: "Get", Status: domain.RunPassed}})
	if _, ok := m.top().(*runResultsScreen); !ok {
		t.Fatalf("a run result should push the results screen, got %T", m.top())
	}
}

// errString is a minimal error for injecting send failures.
type errString string

func (e errString) Error() string { return string(e) }

// ---- assertions screen ----

func openAssertions(t *testing.T) *Model {
	m := openEditor(t)
	send(t, m, keyMsg("a")) // editor -> assertions screen
	if _, ok := m.top().(*assertionsScreen); !ok {
		t.Fatalf("expected assertions screen, got %T", m.top())
	}
	return m
}

func TestAssertionAddAndDelete(t *testing.T) {
	m := openAssertions(t)

	// Add default status-200 assertion.
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("ctrl+s"))
	if n := len(m.app.findCollection("c1").Requests[0].Assertions); n != 1 {
		t.Fatalf("expected 1 assertion after add, got %d", n)
	}

	// Delete it.
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
	openRequests(t, m)
	send(t, m, keyMsg("enter")) // editor
	send(t, m, keyMsg("a"))     // assertions
	send(t, m, keyMsg("enter")) // edit assertion 0
	ed, ok := m.top().(*assertionEditScreen)
	if !ok {
		t.Fatalf("expected assertion editor, got %T", m.top())
	}
	// Change Expected to 404.
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
	// Still on the editor, no assertion saved.
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

// ---- response screen rendering ----

func TestResponseScreenRendersStatusAndBody(t *testing.T) {
	app := NewApp(store.New(t.TempDir()), &store.Data{})
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	resp := &httpclient.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Headers:    h,
		Body:       []byte(`{"error":"nope"}`),
		Elapsed:    12 * time.Millisecond,
	}
	scr := newResponseScreen(app, domain.Request{Method: domain.GET, URL: "https://example.com"}, resp)
	for _, want := range []string{"404 Not Found", "12ms", "Content-Type", "error", "nope"} {
		if !strings.Contains(scr.content, want) {
			t.Errorf("response content missing %q:\n%s", want, scr.content)
		}
	}
	// JSON body must be pretty-printed (indented).
	if !strings.Contains(scr.content, "\n  ") {
		t.Errorf("JSON body not indented:\n%s", scr.content)
	}
}

func TestResponseScreenMasksSecrets(t *testing.T) {
	app := NewApp(store.New(t.TempDir()), &store.Data{})
	app.secrets = domain.Secrets{"tok": "s3cr3t"}
	app.environments = []domain.Environment{{ID: "e1", Name: "e", Vars: map[string]string{"base": "https://h"}}}
	app.cfg.ActiveEnvironmentID = "e1"
	resp := &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{}, Body: []byte("ok")}
	scr := newResponseScreen(app, domain.Request{Method: domain.GET, URL: "{{base}}/{{tok}}"}, resp)
	if strings.Contains(scr.content, "s3cr3t") {
		t.Fatalf("secret leaked into response request line:\n%s", scr.content)
	}
	if !strings.Contains(scr.content, "••••••") {
		t.Fatalf("expected masked secret in request line:\n%s", scr.content)
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
	// Skipped explains itself; error shows the message.
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
	send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	send(t, m, keyMsg("esc"))
	if _, ok := m.top().(*runResultsScreen); ok {
		t.Fatal("esc should pop the run results screen")
	}
}
