package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// treeRow is one visible line in the tree: either a collection header or a
// request nested under an expanded collection.
type treeRow struct {
	isCollection bool
	collectionID string
	requestID    string
	label        string
}

// treePane is the collapsible two-level collection/request browser. Each
// collection expands or collapses independently; requests under a collapsed
// collection are hidden and skipped by cursor movement.
type treePane struct {
	app      *App
	expanded map[string]bool // collection id -> expanded
	rows     []treeRow
	cursor   int

	// activate loads a request into the rest of the workspace.
	activate func(m *Model, collID, reqID string) tea.Cmd
}

func newTreePane(app *App) *treePane {
	t := &treePane{app: app, expanded: map[string]bool{}}
	// Expand every collection by default so requests are visible on launch.
	for _, c := range app.collections {
		t.expanded[c.ID] = true
	}
	t.rebuild()
	return t
}

// rebuild recomputes the visible rows from the collections and expand state.
func (t *treePane) rebuild() {
	t.rows = t.rows[:0]
	for _, c := range t.app.collections {
		glyph := "▸"
		if t.expanded[c.ID] {
			glyph = "▾"
		}
		t.rows = append(t.rows, treeRow{
			isCollection: true,
			collectionID: c.ID,
			label:        fmt.Sprintf("%s %s  (%d)", glyph, c.Name, len(c.Requests)),
		})
		if t.expanded[c.ID] {
			for _, r := range c.Requests {
				t.rows = append(t.rows, treeRow{
					collectionID: c.ID,
					requestID:    r.ID,
					label:        "    " + fmtRequestRow(r),
				})
			}
		}
	}
	if t.cursor >= len(t.rows) {
		t.cursor = len(t.rows) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
}

// firstRequest returns the collection/request id of the first request, or empty.
func (t *treePane) firstRequest() (string, string) {
	for _, c := range t.app.collections {
		if len(c.Requests) > 0 {
			return c.ID, c.Requests[0].ID
		}
	}
	return "", ""
}

func (t *treePane) cur() (treeRow, bool) {
	if t.cursor < 0 || t.cursor >= len(t.rows) {
		return treeRow{}, false
	}
	return t.rows[t.cursor], true
}

func (t *treePane) collectionForCursor() string {
	if r, ok := t.cur(); ok {
		return r.collectionID
	}
	return ""
}

func (t *treePane) editing() bool { return false }

func (t *treePane) Update(m *Model, msg tea.Msg, focused bool) tea.Cmd {
	if !focused {
		return nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch km.String() {
	case "up", "k":
		if t.cursor > 0 {
			t.cursor--
		}
	case "down", "j":
		if t.cursor < len(t.rows)-1 {
			t.cursor++
		}
	case "enter", " ":
		row, ok := t.cur()
		if !ok {
			return nil
		}
		if row.isCollection {
			t.expanded[row.collectionID] = !t.expanded[row.collectionID]
			t.rebuild()
			return nil
		}
		if t.activate != nil {
			return t.activate(m, row.collectionID, row.requestID)
		}
	case "N":
		return m.push(newPromptScreen("New collection name:", "", t.createCollection))
	case "n":
		cid := t.collectionForCursor()
		if cid == "" {
			m.setError("Select a collection first")
			return nil
		}
		return m.push(newPromptScreen("New request name:", "", func(m *Model, v string) tea.Cmd {
			return t.createRequest(m, cid, v)
		}))
	case "r":
		return t.rename(m)
	case "d":
		return t.del(m)
	case "x":
		cid := t.collectionForCursor()
		if cid == "" {
			return nil
		}
		c := t.app.findCollection(cid)
		if c == nil {
			return nil
		}
		m.setStatus("Running collection…")
		return runCollectionCmd(t.app, *c)
	}
	return nil
}

func (t *treePane) createCollection(m *Model, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Collection name cannot be empty")
		return nil
	}
	if t.app.collectionNameTaken(name, "") {
		m.setError(fmt.Sprintf("A collection named %q already exists", name))
		return nil
	}
	c := domain.Collection{ID: domain.NewID(name), Name: name}
	if err := t.app.saveCollection(c); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	t.app.collections = append(t.app.collections, c)
	t.expanded[c.ID] = true
	t.rebuild()
	m.setStatus(fmt.Sprintf("Created collection %q", name))
	return nil
}

func (t *treePane) createRequest(m *Model, collID, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Request name cannot be empty")
		return nil
	}
	c := t.app.findCollection(collID)
	if c == nil {
		return nil
	}
	r := domain.Request{ID: domain.NewID(name), Name: name, Method: domain.GET}
	c.Requests = append(c.Requests, r)
	if err := t.app.saveCollection(*c); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	t.expanded[collID] = true
	t.rebuild()
	m.setStatus(fmt.Sprintf("Created request %q", name))
	if t.activate != nil {
		return t.activate(m, collID, r.ID)
	}
	return nil
}

func (t *treePane) rename(m *Model) tea.Cmd {
	row, ok := t.cur()
	if !ok {
		return nil
	}
	if row.isCollection {
		c := t.app.findCollection(row.collectionID)
		return m.push(newPromptScreen("Rename collection:", c.Name, func(m *Model, v string) tea.Cmd {
			v = strings.TrimSpace(v)
			if v == "" {
				m.setError("Name cannot be empty")
				return nil
			}
			if t.app.collectionNameTaken(v, c.ID) {
				m.setError(fmt.Sprintf("A collection named %q already exists", v))
				return nil
			}
			c.Name = v
			if err := t.app.saveCollection(*c); err != nil {
				m.setError("Save failed: " + err.Error())
				return nil
			}
			t.rebuild()
			m.setStatus("Renamed collection")
			return nil
		}))
	}
	c := t.app.findCollection(row.collectionID)
	var cur *domain.Request
	for i := range c.Requests {
		if c.Requests[i].ID == row.requestID {
			cur = &c.Requests[i]
		}
	}
	if cur == nil {
		return nil
	}
	return m.push(newPromptScreen("Rename request:", cur.Name, func(m *Model, v string) tea.Cmd {
		v = strings.TrimSpace(v)
		if v == "" {
			m.setError("Name cannot be empty")
			return nil
		}
		cur.Name = v
		if err := t.app.saveCollection(*c); err != nil {
			m.setError("Save failed: " + err.Error())
			return nil
		}
		t.rebuild()
		m.setStatus("Renamed request")
		return nil
	}))
}

func (t *treePane) del(m *Model) tea.Cmd {
	row, ok := t.cur()
	if !ok {
		return nil
	}
	if row.isCollection {
		c := t.app.findCollection(row.collectionID)
		return m.push(newConfirmScreen(fmt.Sprintf("Delete collection %q and its requests?", c.Name), func(m *Model) tea.Cmd {
			if err := t.app.deleteCollection(row.collectionID); err != nil {
				m.setError("Delete failed: " + err.Error())
				return nil
			}
			t.rebuild()
			m.setStatus("Deleted collection")
			return nil
		}))
	}
	c := t.app.findCollection(row.collectionID)
	var name string
	for _, r := range c.Requests {
		if r.ID == row.requestID {
			name = r.Name
		}
	}
	return m.push(newConfirmScreen(fmt.Sprintf("Delete request %q?", name), func(m *Model) tea.Cmd {
		c.Requests = removeByID(c.Requests, row.requestID, func(r domain.Request) string { return r.ID })
		if err := t.app.saveCollection(*c); err != nil {
			m.setError("Delete failed: " + err.Error())
			return nil
		}
		t.rebuild()
		m.setStatus("Deleted request")
		return nil
	}))
}

func (t *treePane) View(m *Model, w, h int, focused bool) string {
	t.rebuild()
	if len(t.rows) == 0 {
		return ui.Subtle.Render("No collections. Press ") + ui.Value.Render("N") + ui.Subtle.Render(" to create one.")
	}
	var b strings.Builder
	// Window the rows around the cursor so a long tree stays scrolled into view.
	start := 0
	if h > 0 && t.cursor >= h {
		start = t.cursor - h + 1
	}
	end := len(t.rows)
	if h > 0 && start+h < end {
		end = start + h
	}
	for i := start; i < end; i++ {
		row := t.rows[i]
		line := row.label
		if i == t.cursor {
			if focused {
				line = ui.Selected.Render(" " + strings.TrimLeft(row.label, " ") + " ")
				if !row.isCollection {
					line = "    " + ui.Selected.Render(" "+strings.TrimSpace(row.label)+" ")
				}
			} else {
				line = ui.FieldFocused.Render(row.label)
			}
		} else if row.isCollection {
			line = ui.Value.Render(row.label)
		} else {
			line = ui.Subtle.Render(row.label)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (t *treePane) Title() string { return "Collections" }

var keyRunCol = key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "run collection"))

func (t *treePane) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.New, keys.Rename, keys.Delete, keyRunCol}
}
