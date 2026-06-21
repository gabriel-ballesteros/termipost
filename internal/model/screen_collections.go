package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gbrlballesteros/termipost/internal/domain"
	"github.com/gbrlballesteros/termipost/internal/ui"
)

// collectionListScreen is the top-level screen: browse collections, manage them,
// and jump to the environments manager.
type collectionListScreen struct {
	app  *App
	list list.Model
}

var (
	keyEnvironments = key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "environments"))
	keyRunCol       = key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "run all"))
)

func newCollectionListScreen(app *App) *collectionListScreen {
	s := &collectionListScreen{app: app}
	s.list = newList(s.items(), true)
	return s
}

func (s *collectionListScreen) items() []list.Item {
	items := make([]list.Item, 0, len(s.app.collections))
	for _, c := range s.app.collections {
		items = append(items, simpleItem{
			id:    c.ID,
			title: c.Name,
			desc:  fmt.Sprintf("%d request(s)", len(c.Requests)),
		})
	}
	return items
}

func (s *collectionListScreen) refresh() { s.list.SetItems(s.items()) }

func (s *collectionListScreen) Init(*Model) tea.Cmd { return nil }

func (s *collectionListScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, m.bodyHeight())
		return nil

	case collRunMsg:
		return m.push(newRunResultsScreen(msg.collectionName, msg.result))

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return tea.Quit
		case "n":
			return m.push(newPromptScreen("New collection name:", "", s.createCollection))
		case "r":
			if id := selectedID(s.list); id != "" {
				cur := s.app.findCollection(id)
				return m.push(newPromptScreen("Rename collection:", cur.Name, func(m *Model, v string) tea.Cmd {
					return s.renameCollection(m, id, v)
				}))
			}
		case "d":
			if id := selectedID(s.list); id != "" {
				c := s.app.findCollection(id)
				return m.push(newConfirmScreen(
					fmt.Sprintf("Delete collection %q and its requests?", c.Name),
					func(m *Model) tea.Cmd { return s.deleteCollection(m, id) }))
			}
		case "R":
			if id := selectedID(s.list); id != "" {
				c := s.app.findCollection(id)
				m.setStatus("Running collection…")
				return runCollectionCmd(s.app, *c)
			}
		case "e":
			return m.push(newEnvListScreen(s.app))
		case "enter":
			if id := selectedID(s.list); id != "" {
				return m.push(newRequestListScreen(s.app, id))
			}
			return nil
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return cmd
}

func (s *collectionListScreen) createCollection(m *Model, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Collection name cannot be empty")
		return nil
	}
	if s.app.collectionNameTaken(name, "") {
		m.setError(fmt.Sprintf("A collection named %q already exists", name))
		return nil
	}
	c := domain.Collection{ID: domain.NewID(name), Name: name}
	if err := s.app.saveCollection(c); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.app.collections = append(s.app.collections, c)
	s.refresh()
	m.setStatus(fmt.Sprintf("Created collection %q", name))
	return nil
}

func (s *collectionListScreen) renameCollection(m *Model, id, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Collection name cannot be empty")
		return nil
	}
	if s.app.collectionNameTaken(name, id) {
		m.setError(fmt.Sprintf("A collection named %q already exists", name))
		return nil
	}
	c := s.app.findCollection(id)
	c.Name = name
	if err := s.app.saveCollection(*c); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.refresh()
	m.setStatus("Renamed collection")
	return nil
}

func (s *collectionListScreen) deleteCollection(m *Model, id string) tea.Cmd {
	if err := s.app.deleteCollection(id); err != nil {
		m.setError("Delete failed: " + err.Error())
		return nil
	}
	s.refresh()
	m.setStatus("Deleted collection")
	return nil
}

func (s *collectionListScreen) View(m *Model) string {
	if len(s.app.collections) == 0 {
		return "\n" + ui.Subtle.Render("No collections yet. Press ") +
			ui.Value.Render("n") + ui.Subtle.Render(" to create your first one.")
	}
	return s.list.View()
}

func (s *collectionListScreen) Title() string { return "Collections" }

func (s *collectionListScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.New, keys.Rename, keys.Delete, keyRunCol, keyEnvironments, keys.Quit}
}
