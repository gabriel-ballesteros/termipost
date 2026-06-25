package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/syntax"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

type editorTab int

const (
	etRequest editorTab = iota
	etHeaders
	etQuery
	etBody
	editorTabCount
)

var editorTabNames = []string{"Request", "Headers", "Query", "Body"}

// requestFields on the Request tab.
const (
	efName = iota
	efMethod
	efURL
	efFieldCount
)

// editorPane edits the selected request. Fields are grouped into tabs (Request,
// Headers, Query, Body); Headers and Query use an inline key/value editor. The
// pane tracks a dirty flag against the persisted request. Assertions are edited
// through a pushed overlay that saves immediately, so they are read from the
// stored request and never counted as unsaved edits.
type editorPane struct {
	app          *App
	collectionID string
	reqID        string
	loaded       bool

	persisted domain.Request // last saved snapshot, for dirty detection

	name      textinput.Model
	url       textinput.Model
	body      textarea.Model
	methodIdx int
	headers   *kvEditor
	query     *kvEditor

	tab     editorTab
	fieldIx int  // focused field within the Request tab
	editFld bool // editing name/url/body
}

func newEditorPane(app *App) *editorPane {
	p := &editorPane{app: app}
	p.name = textinput.New()
	p.url = textinput.New()
	p.body = textarea.New()
	p.body.ShowLineNumbers = false
	p.headers = newKVEditor()
	p.query = newKVEditor()
	return p
}

// load swaps the editor to a different request, resetting all working state.
func (p *editorPane) load(collID, reqID string) {
	p.collectionID, p.reqID = collID, reqID
	r := p.storedReq()
	p.persisted = cloneRequest(r)
	p.name.SetValue(r.Name)
	p.url.SetValue(r.URL)
	p.body.SetValue(r.Body)
	p.methodIdx = 0
	for i, mth := range domain.Methods {
		if mth == r.Method {
			p.methodIdx = i
		}
	}
	p.headers.pairs = append([]domain.KV(nil), r.Headers...)
	p.query.pairs = append([]domain.KV(nil), r.QueryParams...)
	p.headers.reset()
	p.query.reset()
	p.tab, p.fieldIx, p.editFld = etRequest, efName, false
	p.loaded = true
}

// storedReq returns the request as currently persisted on disk.
func (p *editorPane) storedReq() domain.Request {
	c := p.app.findCollection(p.collectionID)
	if c != nil {
		for _, r := range c.Requests {
			if r.ID == p.reqID {
				return r
			}
		}
	}
	return domain.Request{}
}

// currentReq builds the working request from the live widgets, taking assertions
// from the stored copy (they are saved through their own overlay).
func (p *editorPane) currentReq() domain.Request {
	return domain.Request{
		ID:          p.reqID,
		Name:        strings.TrimSpace(p.name.Value()),
		Method:      domain.Methods[p.methodIdx],
		URL:         strings.TrimSpace(p.url.Value()),
		Headers:     p.headers.pairs,
		QueryParams: p.query.pairs,
		Body:        p.body.Value(),
		Assertions:  p.storedReq().Assertions,
	}
}

func cloneRequest(r domain.Request) domain.Request {
	r.Headers = append([]domain.KV(nil), r.Headers...)
	r.QueryParams = append([]domain.KV(nil), r.QueryParams...)
	r.Assertions = append([]domain.Assertion(nil), r.Assertions...)
	return r
}

// dirty reports whether the working request differs from the persisted one,
// ignoring assertions (which auto-save through their overlay).
func (p *editorPane) dirty() bool {
	if !p.loaded {
		return false
	}
	a, b := cloneRequest(p.currentReq()), cloneRequest(p.persisted)
	a.Assertions, b.Assertions = nil, nil
	return !reflect.DeepEqual(a, b)
}

// persist writes the working request back into its collection on disk.
func (p *editorPane) persist(m *Model) error {
	if !p.loaded {
		return nil
	}
	req := p.currentReq()
	c := p.app.findCollection(p.collectionID)
	if c == nil {
		return nil
	}
	for i := range c.Requests {
		if c.Requests[i].ID == req.ID {
			c.Requests[i] = req
			if err := p.app.saveCollection(*c); err != nil {
				return err
			}
			p.persisted = cloneRequest(req)
			return nil
		}
	}
	return nil
}

// editing satisfies the pane interface: true while any text/KV input is active.
func (p *editorPane) editing() bool {
	return p.editFld || p.headers.editing || p.query.editing
}

func (p *editorPane) Update(m *Model, msg tea.Msg, focused bool) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return p.forward(msg)
	}
	if !focused {
		return nil
	}
	return p.key(m, km)
}

// forward sends non-key messages to whichever text widget is active so it can
// animate (cursor blink) or process paste etc.
func (p *editorPane) forward(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case p.editFld && p.tab == etBody:
		p.body, cmd = p.body.Update(msg)
	case p.editFld && p.fieldIx == efName:
		p.name, cmd = p.name.Update(msg)
	case p.editFld && p.fieldIx == efURL:
		p.url, cmd = p.url.Update(msg)
	}
	return cmd
}

func (p *editorPane) key(m *Model, msg tea.KeyMsg) tea.Cmd {
	if !p.loaded {
		return nil
	}
	// Prettify works on the Body tab in both navigation and edit mode; it uses a
	// chord (ctrl+f) so it never collides with text input.
	if p.tab == etBody && key.Matches(msg, keys.Prettify) {
		return p.prettifyBody(m)
	}
	// Inline KV editors own input while active.
	if p.tab == etHeaders && p.headers.editing {
		return p.headers.update(m, msg)
	}
	if p.tab == etQuery && p.query.editing {
		return p.query.update(m, msg)
	}
	// Text-field edit mode (name/url/body).
	if p.editFld {
		return p.updateEditing(msg)
	}

	switch msg.String() {
	case "]":
		p.tab = (p.tab + 1) % editorTabCount
		return nil
	case "[":
		p.tab = (p.tab + editorTabCount - 1) % editorTabCount
		return nil
	case "a":
		// On the KV tabs 'a' adds a row; elsewhere it opens the assertions overlay.
		if p.tab == etHeaders {
			return p.headers.update(m, msg)
		}
		if p.tab == etQuery {
			return p.query.update(m, msg)
		}
		return p.openAssertions(m)
	// field jump shortcuts
	case "n":
		p.tab, p.fieldIx = etRequest, efName
		return nil
	case "m":
		p.tab, p.fieldIx = etRequest, efMethod
		return nil
	case "u":
		p.tab, p.fieldIx = etRequest, efURL
		return nil
	case "h":
		p.tab = etHeaders
		return nil
	case "p":
		p.tab = etQuery
		return nil
	case "b":
		p.tab = etBody
		return nil
	}

	switch p.tab {
	case etRequest:
		return p.updateRequestTab(msg)
	case etHeaders:
		return p.headers.update(m, msg)
	case etQuery:
		return p.query.update(m, msg)
	case etBody:
		if msg.String() == "enter" || msg.String() == "i" {
			p.editFld = true
			p.body.Focus()
			return textarea.Blink
		}
	}
	return nil
}

// openAssertions persists the working request first (so the stored copy carries
// the latest fields), then opens the assertions overlay on the stored request.
func (p *editorPane) openAssertions(m *Model) tea.Cmd {
	if err := p.persist(m); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	c := p.app.findCollection(p.collectionID)
	if c == nil {
		return nil
	}
	for i := range c.Requests {
		if c.Requests[i].ID == p.reqID {
			return m.push(newAssertionsScreen(p.app, p.collectionID, &c.Requests[i]))
		}
	}
	return nil
}

func (p *editorPane) updateRequestTab(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab", "down", "j":
		p.fieldIx = (p.fieldIx + 1) % efFieldCount
	case "shift+tab", "up", "k":
		p.fieldIx = (p.fieldIx + efFieldCount - 1) % efFieldCount
	case "left":
		if p.fieldIx == efMethod {
			p.methodIdx = (p.methodIdx + len(domain.Methods) - 1) % len(domain.Methods)
		}
	case "right":
		if p.fieldIx == efMethod {
			p.methodIdx = (p.methodIdx + 1) % len(domain.Methods)
		}
	case "enter", "i":
		switch p.fieldIx {
		case efName:
			p.editFld = true
			p.name.Focus()
			return textinput.Blink
		case efURL:
			p.editFld = true
			p.url.Focus()
			return textinput.Blink
		case efMethod:
			p.methodIdx = (p.methodIdx + 1) % len(domain.Methods)
		}
	}
	return nil
}

func (p *editorPane) updateEditing(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		p.editFld = false
		p.name.Blur()
		p.url.Blur()
		p.body.Blur()
		return nil
	}
	// Enter commits single-line fields; the body keeps Enter for newlines.
	if msg.String() == "enter" && p.tab == etRequest && (p.fieldIx == efName || p.fieldIx == efURL) {
		p.editFld = false
		p.name.Blur()
		p.url.Blur()
		return nil
	}
	var cmd tea.Cmd
	switch {
	case p.tab == etBody:
		p.body, cmd = p.body.Update(msg)
	case p.fieldIx == efName:
		p.name, cmd = p.name.Update(msg)
	case p.fieldIx == efURL:
		p.url, cmd = p.url.Update(msg)
	}
	return cmd
}

func (p *editorPane) View(m *Model, w, h int, focused bool) string {
	if !p.loaded {
		return ui.Subtle.Render("Select a request in the tree.")
	}
	tabs := renderTabs(editorTabNames, int(p.tab), focused)
	dirt := ""
	if p.dirty() {
		dirt = ui.Warn.Render("  ●")
	}
	var body string
	switch p.tab {
	case etRequest:
		body = p.viewRequestTab()
	case etHeaders:
		body = p.headers.view(w, h-2, focused)
	case etQuery:
		body = p.query.view(w, h-2, focused)
	case etBody:
		body = p.viewBodyTab(w, h-2)
	}
	return tabs + dirt + "\n" + body
}

func (p *editorPane) viewRequestTab() string {
	var b strings.Builder
	b.WriteString(p.fieldRow("Name", efName, p.fieldText(efName, p.name.Value())))
	b.WriteString(p.fieldRow("Method", efMethod, fmt.Sprintf("< %s >", domain.Methods[p.methodIdx])))
	b.WriteString(p.fieldRow("URL", efURL, p.fieldText(efURL, p.url.Value())))
	b.WriteString("\n" + ui.Subtle.Render(fmt.Sprintf("Assertions: %d  (press a)", len(p.storedReq().Assertions))))
	return b.String()
}

func (p *editorPane) fieldRow(label string, f int, value string) string {
	lbl := fmt.Sprintf("%-8s", label+":")
	if f == p.fieldIx && p.tab == etRequest {
		return ui.FieldFocused.Render("▸ "+lbl) + " " + value + "\n"
	}
	return ui.Label.Render("  "+lbl) + " " + ui.Value.Render(value) + "\n"
}

func (p *editorPane) fieldText(f int, fallback string) string {
	if p.editFld && p.tab == etRequest && p.fieldIx == f {
		switch f {
		case efName:
			return p.name.View()
		case efURL:
			return p.url.View()
		}
	}
	if fallback == "" {
		return ui.Subtle.Render("(empty)")
	}
	return fallback
}

func (p *editorPane) viewBodyTab(w, h int) string {
	if p.editFld && p.tab == etBody {
		p.body.SetWidth(min(w, 80))
		p.body.SetHeight(max(h-2, 3))
		return p.body.View() + "\n" + p.bodyValidity()
	}
	return ui.Box.Render(bodyPreview(p.body.Value()))
}

// prettifyBody formats and validates the JSON body in place. On a parse error the
// body is left untouched and the error is surfaced; otherwise the formatted body
// replaces the current one (marking the request dirty if it changed).
func (p *editorPane) prettifyBody(m *Model) tea.Cmd {
	cur := p.body.Value()
	out, err := syntax.Prettify(cur)
	if err != nil {
		m.setError("Invalid JSON: " + err.Error())
		return nil
	}
	if out != cur {
		p.body.SetValue(out)
		m.setStatus("Body prettified")
	} else {
		m.setStatus("Body already formatted")
	}
	return nil
}

// bodyValidity returns a live valid/invalid indicator for the body while editing.
// It is gated on JSON-looking content so plain-text or form bodies show nothing.
func (p *editorPane) bodyValidity() string {
	v := p.body.Value()
	if !looksLikeJSON([]byte(v)) {
		return ""
	}
	if ok, err := syntax.ValidateJSON(v); ok {
		return ui.Good.Render("✓ valid JSON")
	} else {
		return ui.Bad.Render("✗ " + err.Error())
	}
}

func (p *editorPane) Title() string { return "Request" }

func (p *editorPane) HelpBindings() []key.Binding {
	if p.editing() {
		b := []key.Binding{key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "stop editing"))}
		if p.tab == etBody {
			b = append(b, keys.Prettify)
		}
		return b
	}
	b := []key.Binding{keys.Tab, keys.Enter, keys.TabPrev, keys.TabNext,
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "assertions"))}
	if p.tab == etBody {
		b = append(b, keys.Prettify)
	}
	return b
}
