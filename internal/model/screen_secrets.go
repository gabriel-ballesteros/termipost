package model

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gbrlballesteros/termipost/internal/ui"
)

// secretsScreen edits the single global secrets store. Values are masked by
// default; a reveal toggle shows them. Secrets persist to a gitignored file.
type secretsScreen struct {
	app      *App
	keys     []string
	cursor   int
	revealed bool
}

func newSecretsScreen(app *App) *secretsScreen {
	s := &secretsScreen{app: app}
	s.reloadKeys()
	return s
}

func (s *secretsScreen) reloadKeys() {
	s.keys = s.keys[:0]
	for k := range s.app.secrets {
		s.keys = append(s.keys, k)
	}
	sort.Strings(s.keys)
	if s.cursor >= len(s.keys) && s.cursor > 0 {
		s.cursor = len(s.keys) - 1
	}
}

func (s *secretsScreen) Init(*Model) tea.Cmd { return nil }

func (s *secretsScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch km.String() {
	case "esc":
		m.pop()
	case "v":
		s.revealed = !s.revealed
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(s.keys)-1 {
			s.cursor++
		}
	case "a":
		return m.push(newPromptScreen("Add secret (format: name: value):", "", func(m *Model, v string) tea.Cmd {
			if kv, ok := parseKV(v); ok {
				s.app.secrets[kv.Key] = kv.Value
				s.save(m)
				s.reloadKeys()
			} else {
				m.setError("Use the format name: value")
			}
			return nil
		}))
	case "e", "enter":
		if s.cursor < len(s.keys) {
			name := s.keys[s.cursor]
			return m.push(newPromptScreen("Edit secret "+name+":", s.app.secrets[name], func(m *Model, v string) tea.Cmd {
				s.app.secrets[name] = strings.TrimSpace(v)
				s.save(m)
				return nil
			}))
		}
	case "d":
		if s.cursor < len(s.keys) {
			delete(s.app.secrets, s.keys[s.cursor])
			s.save(m)
			s.reloadKeys()
		}
	}
	return nil
}

func (s *secretsScreen) save(m *Model) {
	if err := s.app.saveSecrets(); err != nil {
		m.setError("Save failed: " + err.Error())
	}
}

func (s *secretsScreen) View(m *Model) string {
	hint := ui.Subtle.Render("(values masked — press v to reveal)")
	if s.revealed {
		hint = ui.Warn.Render("(values revealed — press v to hide)")
	}
	header := ui.Label.Render("Global secrets ") + hint + "\n\n"
	if len(s.keys) == 0 {
		return header + ui.Subtle.Render("No secrets. Press ") + ui.Value.Render("a") + ui.Subtle.Render(" to add one.")
	}
	var b strings.Builder
	for i, k := range s.keys {
		val := "••••••"
		if s.revealed {
			val = s.app.secrets[k]
		}
		line := k + ": " + val
		if i == s.cursor {
			b.WriteString(ui.Selected.Render(" "+line+" ") + "\n")
		} else {
			b.WriteString("  " + ui.Value.Render(line) + "\n")
		}
	}
	return header + b.String()
}

func (s *secretsScreen) Title() string { return "Secrets" }

func (s *secretsScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Add, keys.Edit, keys.Delete, keys.Reveal, keys.Back}
}
