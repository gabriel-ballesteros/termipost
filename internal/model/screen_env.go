package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// envListScreen manages environments: create/rename/delete, set active, edit
// variables, and open the global secrets editor.
type envListScreen struct {
	app  *App
	list list.Model
}

var (
	keySetActive = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "set active"))
	keySecrets   = key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "secrets"))
	keyVars      = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit vars"))
)

func newEnvListScreen(app *App) *envListScreen {
	s := &envListScreen{app: app}
	s.list = newList(s.items(), true)
	return s
}

func (s *envListScreen) items() []list.Item {
	items := make([]list.Item, 0, len(s.app.environments))
	for _, e := range s.app.environments {
		title := e.Name
		if e.ID == s.app.cfg.ActiveEnvironmentID {
			title += "  " + ui.Good.Render("(active)")
		}
		items = append(items, simpleItem{id: e.ID, title: title, desc: fmt.Sprintf("%d variable(s)", len(e.Vars))})
	}
	return items
}

func (s *envListScreen) refresh() { s.list.SetItems(s.items()) }

func (s *envListScreen) findEnv(id string) *domain.Environment {
	for i := range s.app.environments {
		if s.app.environments[i].ID == id {
			return &s.app.environments[i]
		}
	}
	return nil
}

func (s *envListScreen) Init(*Model) tea.Cmd { return nil }

func (s *envListScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, m.bodyHeight())
		return nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.pop()
			return nil
		case "s":
			return m.push(newSecretsScreen(s.app))
		case "n":
			return m.push(newPromptScreen("New environment name:", "", s.createEnv))
		case "r":
			if id := selectedID(s.list); id != "" {
				cur := s.findEnv(id)
				return m.push(newPromptScreen("Rename environment:", cur.Name, func(m *Model, v string) tea.Cmd {
					return s.renameEnv(m, id, v)
				}))
			}
		case "d":
			if id := selectedID(s.list); id != "" {
				e := s.findEnv(id)
				return m.push(newConfirmScreen(fmt.Sprintf("Delete environment %q?", e.Name),
					func(m *Model) tea.Cmd { return s.deleteEnv(m, id) }))
			}
		case "a":
			if id := selectedID(s.list); id != "" {
				s.app.cfg.ActiveEnvironmentID = id
				if err := s.app.saveConfig(); err != nil {
					m.setError("Save failed: " + err.Error())
					return nil
				}
				s.refresh()
				m.setStatus("Active environment set")
			}
			return nil
		case "enter":
			if id := selectedID(s.list); id != "" {
				e := s.findEnv(id)
				return m.push(newKVEditorScreen("Variables: "+e.Name, mapToKVs(e.Vars), func(m *Model, p []domain.KV) tea.Cmd {
					return s.saveVars(m, id, p)
				}))
			}
			return nil
		}
	}
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return cmd
}

func (s *envListScreen) createEnv(m *Model, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Environment name cannot be empty")
		return nil
	}
	if s.app.envNameTaken(name, "") {
		m.setError(fmt.Sprintf("An environment named %q already exists", name))
		return nil
	}
	e := domain.Environment{ID: domain.NewID(name), Name: name, Vars: map[string]string{}}
	if err := s.app.saveEnvironment(e); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.app.environments = append(s.app.environments, e)
	s.refresh()
	m.setStatus(fmt.Sprintf("Created environment %q", name))
	return nil
}

func (s *envListScreen) renameEnv(m *Model, id, name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		m.setError("Environment name cannot be empty")
		return nil
	}
	if s.app.envNameTaken(name, id) {
		m.setError(fmt.Sprintf("An environment named %q already exists", name))
		return nil
	}
	e := s.findEnv(id)
	e.Name = name
	if err := s.app.saveEnvironment(*e); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.refresh()
	m.setStatus("Renamed environment")
	return nil
}

func (s *envListScreen) deleteEnv(m *Model, id string) tea.Cmd {
	if err := s.app.deleteEnvironment(id); err != nil {
		m.setError("Delete failed: " + err.Error())
		return nil
	}
	s.refresh()
	m.setStatus("Deleted environment")
	return nil
}

func (s *envListScreen) saveVars(m *Model, id string, pairs []domain.KV) tea.Cmd {
	e := s.findEnv(id)
	e.Vars = kvsToMap(pairs)
	if err := s.app.saveEnvironment(*e); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	s.refresh()
	return nil
}

func (s *envListScreen) View(m *Model) string {
	if len(s.app.environments) == 0 {
		return "\n" + ui.Subtle.Render("No environments yet. Press ") + ui.Value.Render("n") +
			ui.Subtle.Render(" to create one, or ") + ui.Value.Render("s") + ui.Subtle.Render(" to edit secrets.")
	}
	return s.list.View()
}

func (s *envListScreen) Title() string { return "Environments" }

func (s *envListScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keyVars, keys.New, keys.Rename, keys.Delete, keySetActive, keySecrets, keys.Back}
}

func mapToKVs(m map[string]string) []domain.KV {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]domain.KV, 0, len(keys))
	for _, k := range keys {
		out = append(out, domain.KV{Key: k, Value: m[k]})
	}
	return out
}

func kvsToMap(pairs []domain.KV) map[string]string {
	out := map[string]string{}
	for _, kv := range pairs {
		if kv.Key != "" {
			out[kv.Key] = kv.Value
		}
	}
	return out
}
