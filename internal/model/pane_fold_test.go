package model

import (
	"strings"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

// TestEditorBodyFoldingIsReadOnly verifies that collapsing a request body section
// shows the fold gutter, never mutates the stored body, and never marks the
// request dirty, and that folding is inert while the body is being edited.
func TestEditorBodyFoldingIsReadOnly(t *testing.T) {
	body := `{"a":1,"b":{"c":2}}`
	col := domain.Collection{ID: "c-1", Name: "Demo", Requests: []domain.Request{
		{ID: "r-1", Name: "Get", Method: domain.GET, URL: "https://x", Body: body},
	}}
	m := seededModel(t, col)
	w := ws(m)
	w.focus = paneEditor
	ed := w.editor
	ed.load("c-1", "r-1")
	ed.tab = etBody

	// Render once to seed the fold view from the body.
	if out := ed.View(m, 80, 24, true); !strings.Contains(out, "-") {
		t.Fatalf("expanded body should render a '-' gutter:\n%s", out)
	}
	if !ed.bodyFold.foldable {
		t.Fatal("multi-line JSON body should be foldable")
	}

	// Space collapses the section under the cursor (the root object).
	ed.key(m, keyMsg(" "))
	if !ed.bodyFold.collapsed[0] {
		t.Fatal("space should collapse the section at the cursor")
	}
	out := ed.View(m, 80, 24, true)
	if !strings.Contains(out, "+") || !strings.Contains(out, "…") {
		t.Errorf("collapsed body should show '+' and ellipsis:\n%s", out)
	}

	// Folding is view-only: stored body unchanged and request not dirty.
	if got := ed.storedReq().Body; got != body {
		t.Errorf("folding changed stored body: %q", got)
	}
	if ed.dirty() {
		t.Error("folding must not mark the request as dirty")
	}

	// While editing, the fold toggle is inert and space reaches the textarea.
	ed.editFld = true
	ed.body.Focus()
	before := ed.bodyFold.collapsed[0]
	ed.key(m, keyMsg(" "))
	if ed.bodyFold.collapsed[0] != before {
		t.Error("fold state must not change while editing the body")
	}
}

// TestResponseBodyFoldingIsReadOnly verifies the response pane folds a JSON body
// without altering the captured bytes.
func TestResponseBodyFoldingIsReadOnly(t *testing.T) {
	p := newResponsePane(NewApp(nil, &store.Data{}))
	raw := []byte(`{"a":1,"b":{"c":2}}`)
	p.setResponse(&domain.Request{Method: domain.GET}, &httpclient.Response{
		StatusCode: 200, Status: "200 OK", Body: raw,
	})
	p.tab = rtBody
	if !p.fold.foldable {
		t.Fatal("JSON response body should be foldable")
	}
	p.fold.toggle() // collapse root
	if !p.fold.collapsed[0] {
		t.Fatal("toggle should collapse the root")
	}
	if string(p.body) != string(raw) {
		t.Error("folding must not alter the captured response body")
	}
}
