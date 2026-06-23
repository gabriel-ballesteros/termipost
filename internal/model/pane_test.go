package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

// ctrlw sends the vim window chord ctrl+w followed by a direction.
func ctrlw(t *testing.T, m *Model, dir string) {
	t.Helper()
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlW})
	send(t, m, keyMsg(dir))
}

func twoReqModel(t *testing.T) *Model {
	col := domain.Collection{ID: "c1", Name: "API", Requests: []domain.Request{
		{ID: "r1", Name: "First", Method: domain.GET, URL: "https://a"},
		{ID: "r2", Name: "Second", Method: domain.POST, URL: "https://b"},
	}}
	return seededModel(t, col)
}

// ---- 6.2 focus transitions ----

func TestFocusWindowChord(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.focus = paneTree

	ctrlw(t, m, "l")
	if w.focus != paneEditor {
		t.Fatalf("ctrl+w l from tree should focus editor, got %d", w.focus)
	}
	ctrlw(t, m, "j")
	if w.focus != paneResponse {
		t.Fatalf("ctrl+w j from editor should focus response, got %d", w.focus)
	}
	ctrlw(t, m, "k")
	if w.focus != paneEditor {
		t.Fatalf("ctrl+w k from response should focus editor, got %d", w.focus)
	}
	ctrlw(t, m, "h")
	if w.focus != paneTree {
		t.Fatalf("ctrl+w h from editor should focus tree, got %d", w.focus)
	}
	// No-op at an edge: tree has no pane to its left.
	ctrlw(t, m, "h")
	if w.focus != paneTree {
		t.Fatalf("ctrl+w h at the left edge should be a no-op, got %d", w.focus)
	}
}

func TestKeysOnlyHitFocusedPane(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.focus = paneTree
	w.tree.cursor = 0

	// j moves the tree cursor (focused), and must not scroll the response pane.
	send(t, m, keyMsg("j"))
	if w.tree.cursor != 1 {
		t.Fatalf("j should move the tree cursor when tree focused, got %d", w.tree.cursor)
	}
}

// ---- 6.4 in-pane keys do not change focus ----

func TestInPaneKeysDoNotMoveFocus(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.focus = paneEditor
	for _, k := range []string{"tab", "j", "k", "[", "]"} {
		send(t, m, keyMsg(k))
		if w.focus != paneEditor {
			t.Fatalf("key %q moved focus away from the editor (got %d)", k, w.focus)
		}
	}
}

// ---- 6.3 single-pane fallback ----

func TestSinglePaneFallback(t *testing.T) {
	m := twoReqModel(t)
	send(t, m, tea.WindowSizeMsg{Width: 50, Height: 10})
	w := ws(m)
	if w.multiPane(50, m.bodyHeight()) {
		t.Fatal("50x10 should not be multi-pane")
	}
	w.focus = paneTree
	v := m.View()
	if !strings.Contains(v, "Collections") {
		t.Fatalf("fallback should show the focused (tree) pane:\n%s", v)
	}
	if strings.Contains(v, "Response") {
		t.Fatalf("fallback should hide the unfocused response pane:\n%s", v)
	}
	// The window chord switches which single pane is shown.
	ctrlw(t, m, "l")
	if v := m.View(); !strings.Contains(v, "Request") {
		t.Fatalf("ctrl+w l should switch the shown pane to the editor:\n%s", v)
	}
}

// ---- 6.5 unsaved-edit guard on request switch ----

func TestUnsavedGuardOnSwitchConfirm(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.editor.name.SetValue("dirty edit") // make the editor dirty without saving
	w.focus = paneTree
	w.tree.cursor = 2 // the second request row

	send(t, m, keyMsg("enter")) // attempt to switch
	if _, ok := m.top().(*confirmScreen); !ok {
		t.Fatalf("switching away from a dirty editor should confirm, got %T", m.top())
	}
	send(t, m, keyMsg("y")) // confirm discard
	if w.editor.reqID != "r2" {
		t.Fatalf("after confirming, editor should load r2, got %q", w.editor.reqID)
	}
}

func TestUnsavedGuardOnSwitchCancel(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.editor.name.SetValue("dirty edit")
	w.focus = paneTree
	w.tree.cursor = 2

	send(t, m, keyMsg("enter"))
	send(t, m, keyMsg("n")) // cancel
	if w.editor.reqID != "r1" {
		t.Fatalf("cancelling should keep r1 loaded, got %q", w.editor.reqID)
	}
}

func TestCleanSwitchNoPrompt(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.focus = paneTree
	w.tree.cursor = 2
	send(t, m, keyMsg("enter"))
	if _, ok := m.top().(*confirmScreen); ok {
		t.Fatal("a clean editor should switch without a prompt")
	}
	if w.editor.reqID != "r2" {
		t.Fatalf("clean switch should load r2, got %q", w.editor.reqID)
	}
}

// ---- 6.6 inline KV editing persists ----

func TestInlineKVAddPersists(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.focus = paneEditor
	send(t, m, keyMsg("h")) // Headers tab
	send(t, m, keyMsg("a")) // add row -> inline edit
	send(t, m, keyMsg("X-Test"))
	send(t, m, keyMsg("tab")) // move to value field
	send(t, m, keyMsg("yes"))
	send(t, m, keyMsg("enter"))  // commit row
	send(t, m, keyMsg("ctrl+s")) // persist

	hs := m.app.findCollection("c1").Requests[0].Headers
	if len(hs) != 1 || hs[0].Key != "X-Test" || hs[0].Value != "yes" {
		t.Fatalf("inline header not persisted: %+v", hs)
	}
}

// ---- 6.8 quit guard ----

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestQuitGuard(t *testing.T) {
	// ctrl+c always quits.
	m := twoReqModel(t)
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); !isQuit(cmd) {
		t.Fatal("ctrl+c should always quit")
	}

	// Clean editor: q quits.
	m = twoReqModel(t)
	if _, cmd := m.Update(keyMsg("q")); !isQuit(cmd) {
		t.Fatal("q on a clean editor should quit")
	}

	// Dirty editor: q prompts instead of quitting.
	m = twoReqModel(t)
	ws(m).editor.name.SetValue("dirty")
	if _, cmd := m.Update(keyMsg("q")); isQuit(cmd) {
		t.Fatal("q on a dirty editor should not quit immediately")
	}
	if _, ok := m.top().(*confirmScreen); !ok {
		t.Fatalf("q on a dirty editor should push a confirm, got %T", m.top())
	}
}

// ---- 6.9 tree collapse ----

func TestTreeCollapseToggle(t *testing.T) {
	m := twoReqModel(t)
	w := ws(m)
	w.focus = paneTree
	w.tree.cursor = 0 // the collection row

	// Expanded by default: 1 collection + 2 requests = 3 rows.
	if got := len(w.tree.rows); got != 3 {
		t.Fatalf("expected 3 rows expanded, got %d", got)
	}
	send(t, m, keyMsg("enter")) // collapse the collection
	if got := len(w.tree.rows); got != 1 {
		t.Fatalf("collapsing should hide requests, got %d rows", got)
	}
	send(t, m, keyMsg("enter")) // expand again
	if got := len(w.tree.rows); got != 3 {
		t.Fatalf("expanding should show requests again, got %d rows", got)
	}

	// Enter on a request row loads it instead of toggling.
	w.tree.cursor = 2
	send(t, m, keyMsg("enter"))
	if w.editor.reqID != "r2" {
		t.Fatalf("enter on a request row should load it, got %q", w.editor.reqID)
	}
}
