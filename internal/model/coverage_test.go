package model

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

// ---- kvEditor ----

func TestKVEditorAddEditDelete(t *testing.T) {
	e := newKVEditor()
	m := &Model{}

	// Add a row via 'a' then type key, tab to value, type, commit.
	e.update(m, keyMsg("a"))
	if !e.editing || !e.isNew {
		t.Fatal("'a' should start a new row edit")
	}
	e.update(m, keyMsg("X-One"))
	e.update(m, keyMsg("tab")) // -> value field
	if e.field != 1 {
		t.Fatalf("tab should move to the value field, got %d", e.field)
	}
	e.update(m, keyMsg("v1"))
	e.update(m, keyMsg("enter")) // commit
	if len(e.pairs) != 1 || e.pairs[0].Key != "X-One" || e.pairs[0].Value != "v1" {
		t.Fatalf("row not added: %+v", e.pairs)
	}

	// Edit the existing row: 'e' loads it, change value, commit.
	e.cursor = 0
	e.update(m, keyMsg("e"))
	if !e.editing || e.isNew {
		t.Fatal("'e' should edit the existing row")
	}
	if e.keyIn.Value() != "X-One" {
		t.Fatalf("edit should preload the key, got %q", e.keyIn.Value())
	}
	e.update(m, keyMsg("tab"))
	e.update(m, keyMsg("tab")) // back to key field
	e.update(m, keyMsg("esc")) // cancel keeps the row unchanged
	if len(e.pairs) != 1 {
		t.Fatalf("cancel should keep the row: %+v", e.pairs)
	}

	// Empty key is rejected on commit.
	e.update(m, keyMsg("a"))
	e.update(m, keyMsg("enter"))
	if !m.statusErr {
		t.Fatal("empty key should be rejected")
	}
	e.cancel()

	// Delete the row.
	e.cursor = 0
	e.update(m, keyMsg("d"))
	if len(e.pairs) != 0 {
		t.Fatalf("'d' should delete the row: %+v", e.pairs)
	}
}

func TestKVEditorNavAndView(t *testing.T) {
	e := newKVEditor()
	e.pairs = []domain.KV{{Key: "A", Value: "1"}, {Key: "B", Value: "2"}}
	m := &Model{}

	e.cursor = 0
	e.update(m, keyMsg("j")) // down
	if e.cursor != 1 {
		t.Fatalf("j should move down, got %d", e.cursor)
	}
	e.update(m, keyMsg("j")) // onto the add row
	if e.cursor != e.addRow() {
		t.Fatalf("cursor should reach the add row, got %d", e.cursor)
	}
	e.update(m, keyMsg("k"))
	if e.cursor != 1 {
		t.Fatalf("k should move up, got %d", e.cursor)
	}

	// View: rows + add affordance.
	v := e.view(40, 10, true)
	if !strings.Contains(v, "A: 1") || !strings.Contains(v, "+ add") {
		t.Fatalf("view missing rows or add affordance:\n%s", v)
	}
	// Editing view shows the inline inputs.
	e.startEdit()
	if ev := e.view(40, 10, true); !strings.Contains(ev, "▸") {
		t.Fatalf("editing view should show the inline edit row:\n%s", ev)
	}
	// Empty editor hint.
	empty := newKVEditor()
	if v := empty.view(40, 10, true); !strings.Contains(v, "No entries") {
		t.Fatalf("empty kv editor should hint:\n%s", v)
	}
}

// ---- response pane ----

func TestResponsePaneStatusBadgeAndContent(t *testing.T) {
	p := newResponsePane(NewApp(store.New(t.TempDir()), &store.Data{}))

	if !strings.Contains(p.statusBadge(), "no response") {
		t.Fatal("nil response should report no response")
	}
	if !strings.Contains(p.content(), "No response yet") {
		t.Fatal("empty content should hint to send")
	}

	cases := []struct {
		code   int
		status string
	}{{200, "200 OK"}, {301, "301 Moved"}, {404, "404 Not Found"}}
	for _, c := range cases {
		h := http.Header{}
		h.Set("Content-Type", "application/json")
		resp := &httpclient.Response{StatusCode: c.code, Status: c.status, Headers: h,
			Body: []byte(`{"a":1}`), Elapsed: 5 * time.Millisecond}
		p.setResponse(&domain.Request{Method: domain.GET}, resp)
		if !strings.Contains(p.statusBadge(), c.status) {
			t.Fatalf("badge missing status %q", c.status)
		}
		// Body tab content is the pretty body; Headers tab lists the header.
		p.tab = rtBody
		if !strings.Contains(p.content(), "a") {
			t.Fatal("body content missing")
		}
		p.tab = rtHeaders
		if !strings.Contains(p.content(), "Content-Type") {
			t.Fatal("headers content missing")
		}
	}
}

func TestResponsePaneUpdate(t *testing.T) {
	m := &Model{}
	p := newResponsePane(NewApp(store.New(t.TempDir()), &store.Data{}))
	resp := &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{}, Body: []byte("hi")}
	p.setResponse(&domain.Request{Method: domain.GET}, resp)
	p.View(nil, 40, 10, true) // initialise the viewport

	p.Update(m, keyMsg("]"), true)
	if p.tab != rtHeaders {
		t.Fatalf("] should switch to Headers, got %d", p.tab)
	}
	p.Update(m, keyMsg("["), true)
	if p.tab != rtBody {
		t.Fatalf("[ should switch back to Body, got %d", p.tab)
	}
	p.Update(m, keyMsg("c"), true) // copy
	if m.status == "" {
		t.Fatal("copy should set a status")
	}
	// Unfocused keys are ignored.
	before := p.tab
	p.Update(m, keyMsg("]"), false)
	if p.tab != before {
		t.Fatal("unfocused pane should ignore keys")
	}
}

// ---- editor pane: body tab + edit ----

func TestEditorBodyTabEditing(t *testing.T) {
	m := withRequest(t)
	w := ws(m)
	ed := w.editor

	send(t, m, keyMsg("b")) // Body tab
	if ed.tab != etBody {
		t.Fatalf("'b' should switch to Body, got %d", ed.tab)
	}
	// Non-editing view shows the preview box.
	if v := ed.View(nil, 60, 12, true); !strings.Contains(v, "Body") {
		t.Fatalf("body tab view missing label:\n%s", v)
	}
	send(t, m, keyMsg("enter")) // edit body
	if !ed.editFld {
		t.Fatal("enter on Body should enter edit mode")
	}
	send(t, m, keyMsg("hello"))
	send(t, m, keyMsg("esc")) // stop editing
	if !strings.Contains(ed.body.Value(), "hello") {
		t.Fatalf("body not updated: %q", ed.body.Value())
	}
	// Editing-mode view renders the textarea.
	send(t, m, keyMsg("enter"))
	if v := ed.View(nil, 60, 12, true); v == "" {
		t.Fatal("editing body view should render")
	}
}

func TestEditorHelpBindingsEditing(t *testing.T) {
	m := withRequest(t)
	ed := ws(m).editor
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("enter")) // editing name
	hb := ed.HelpBindings()
	if len(hb) != 1 || hb[0].Help().Key != "esc" {
		t.Fatalf("editing help should be just esc, got %+v", hb)
	}
}

// ---- tree pane: rename request, run-collection ----

func TestTreeRenameRequest(t *testing.T) {
	m := withRequest(t)
	w := ws(m)
	w.focus = paneTree
	w.tree.cursor = 1 // the request row
	send(t, m, keyMsg("r"))
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("Renamed"))
	send(t, m, keyMsg("enter"))
	if got := m.app.findCollection("c1").Requests[0].Name; got != "Renamed" {
		t.Fatalf("request not renamed, got %q", got)
	}
}

func TestTreeEmptyCursor(t *testing.T) {
	m := seededModel(t)
	w := ws(m)
	if _, ok := w.tree.cur(); ok {
		t.Fatal("empty tree should have no current row")
	}
	if w.tree.collectionForCursor() != "" {
		t.Fatal("empty tree should have no collection for cursor")
	}
}

// ---- workspace: send guards + focus edges ----

func TestWorkspaceSendNoRequest(t *testing.T) {
	m := seededModel(t) // no requests
	send(t, m, keyMsg("R"))
	if !m.statusErr {
		t.Fatal("sending with no request should error")
	}
}

func TestWorkspaceMoveFocusEdges(t *testing.T) {
	m := withRequest(t)
	w := ws(m)
	w.focus = paneTree
	w.moveFocus("k") // no upward neighbour from tree
	if w.focus != paneTree {
		t.Fatalf("invalid direction should be a no-op, got %d", w.focus)
	}
	w.focus = paneResponse
	w.moveFocus("l") // nothing to the right of response
	if w.focus != paneResponse {
		t.Fatalf("edge move should be a no-op, got %d", w.focus)
	}
}

// ---- model: footer rendering ----

func TestRenderHelpWrapsAndCaps(t *testing.T) {
	m := &Model{width: 30}
	// Empty bindings -> single (blank) line.
	if _, n := m.renderHelp(nil); n != 1 {
		t.Fatalf("empty help should be 1 line, got %d", n)
	}
	// Many bindings on a narrow width wrap onto multiple lines.
	var bs []key.Binding
	for _, kv := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		bs = append(bs, key.NewBinding(key.WithKeys(kv), key.WithHelp(kv, "do "+kv+" thing")))
	}
	_, n := m.renderHelp(bs)
	if n < 2 {
		t.Fatalf("narrow footer should wrap to multiple lines, got %d", n)
	}
	if n > maxHelpLines {
		t.Fatalf("footer should cap at %d lines, got %d", maxHelpLines, n)
	}
}

// ---- misc helpers / branches ----

func TestBodyPreview(t *testing.T) {
	if !strings.Contains(bodyPreview("  "), "empty") {
		t.Fatal("blank body should show the empty hint")
	}
	long := strings.Repeat("line\n", 12)
	out := bodyPreview(long)
	if !strings.Contains(out, "…") {
		t.Fatalf("a long body should be truncated with an ellipsis:\n%s", out)
	}
}

func TestEditorFieldTextURLEditing(t *testing.T) {
	m := withRequest(t)
	ed := ws(m).editor
	send(t, m, keyMsg("u"))     // focus URL
	send(t, m, keyMsg("enter")) // edit URL
	// While editing, the Request tab view renders the live URL input.
	if v := ed.View(nil, 60, 12, true); v == "" {
		t.Fatal("editing URL should render the field view")
	}
	if got := ed.fieldText(efURL, ""); got == "" {
		t.Fatal("fieldText should render the live URL input while editing")
	}
}

func TestResponsePaneScroll(t *testing.T) {
	m := &Model{}
	p := newResponsePane(NewApp(store.New(t.TempDir()), &store.Data{}))
	resp := &httpclient.Response{StatusCode: 200, Status: "200 OK", Headers: http.Header{},
		Body: []byte(strings.Repeat("x\n", 200))}
	p.setResponse(&domain.Request{Method: domain.GET}, resp)
	p.View(nil, 20, 5, true)
	// A movement key reaches the viewport without changing the tab.
	p.Update(m, tea.KeyMsg{Type: tea.KeyDown}, true)
	if p.tab != rtBody {
		t.Fatal("scrolling should not change the tab")
	}
}

func TestWorkspaceSendSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "ping", Method: domain.GET, URL: srv.URL},
	}}
	m := seededModel(t, col)
	w := ws(m)
	w.focus = paneEditor
	send(t, m, keyMsg("R")) // exercises workspace.send (sets sending, returns batch)
	if !w.sending {
		t.Fatal("R should start sending")
	}
	// send() does not unwrap tea.Batch, so run the network round explicitly.
	req := w.editor.currentReq()
	m.Update(sendCmd(m.app, req)())
	if w.response.resp == nil {
		t.Fatal("a successful send should populate the response pane")
	}
}
