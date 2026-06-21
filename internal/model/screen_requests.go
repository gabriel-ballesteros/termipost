package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// requestListScreen lists the requests inside one collection.
type requestListScreen struct {
	app          *App
	collectionID string
	list         list.Model
}

func newRequestListScreen(app *App, collectionID string) *requestListScreen {
	s := &requestListScreen{app: app, collectionID: collectionID}
	s.list = newList(s.items(), true)
	return s
}

func (s *requestListScreen) collection() *domain.Collection {
	return s.app.findCollection(s.collectionID)
}

func (s *requestListScreen) items() []list.Item {
	c := s.collection()
	if c == nil {
		return nil
	}
	items := make([]list.Item, 0, len(c.Requests))
	for _, r := range c.Requests {
		desc := r.URL
		if n := len(r.Assertions); n > 0 {
			desc += fmt.Sprintf("   (%d assertion(s))", n)
		}
		items = append(items, simpleItem{id: r.ID, title: fmt.Sprintf("%-6s %s", r.Method, r.Name), desc: desc})
	}
	return items
}

func (s *requestListScreen) refresh() { s.list.SetItems(s.items()) }

func (s *requestListScreen) Init(*Model) tea.Cmd { return nil }

func (s *requestListScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, m.bodyHeight())
		return nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.pop()
			return nil
		case "n":
			return m.push(newPromptScreen("New request name:", "", s.createRequest))
		case "d":
			if id := selectedID(s.list); id != "" {
				r := s.findRequest(id)
				return m.push(newConfirmScreen(
					fmt.Sprintf("Delete request %q?", r.Name),
					func(m *Model) tea.Cmd { return s.deleteRequest(m, id) }))
			}
		case "enter", "e":
			if id := selectedID(s.list); id != "" {
				return m.push(newRequestEditScreen(s.app, s.collectionID, id))
			}
			return nil
		}
	}
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return cmd
}

func (s *requestListScreen) findRequest(id string) *domain.Request {
	c := s.collection()
	for i := range c.Requests {
		if c.Requests[i].ID == id {
			return &c.Requests[i]
		}
	}
	return nil
}

func (s *requestListScreen) createRequest(m *Model, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Request name cannot be empty")
		return nil
	}
	c := s.collection()
	r := domain.Request{ID: domain.NewID(name), Name: name, Method: domain.GET}
	c.Requests = append(c.Requests, r)
	if err := s.app.saveCollection(*c); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.refresh()
	m.setStatus(fmt.Sprintf("Created request %q", name))
	return m.push(newRequestEditScreen(s.app, s.collectionID, r.ID))
}

func (s *requestListScreen) deleteRequest(m *Model, id string) tea.Cmd {
	c := s.collection()
	c.Requests = removeByID(c.Requests, id, func(r domain.Request) string { return r.ID })
	if err := s.app.saveCollection(*c); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.refresh()
	m.setStatus("Deleted request")
	return nil
}

func (s *requestListScreen) View(m *Model) string {
	c := s.collection()
	if c == nil {
		return ui.Bad.Render("collection not found")
	}
	header := ui.Label.Render("Collection: ") + ui.Value.Render(c.Name) + "\n\n"
	if len(c.Requests) == 0 {
		return header + ui.Subtle.Render("No requests yet. Press ") +
			ui.Value.Render("n") + ui.Subtle.Render(" to add one.")
	}
	return header + s.list.View()
}

func (s *requestListScreen) Title() string { return "Requests" }

func (s *requestListScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.New, keys.Delete, keys.Back}
}
