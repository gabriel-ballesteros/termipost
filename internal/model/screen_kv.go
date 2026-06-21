package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gbrlballesteros/termipost/internal/domain"
	"github.com/gbrlballesteros/termipost/internal/ui"
)

// kvEditorScreen edits an ordered list of key/value pairs (headers or query
// params). Entries are added/edited via a "key: value" prompt. On Esc it calls
// onDone with the edited pairs and pops.
type kvEditorScreen struct {
	title  string
	pairs  []domain.KV
	cursor int
	onDone func(m *Model, pairs []domain.KV) tea.Cmd
}

func newKVEditorScreen(title string, pairs []domain.KV, onDone func(m *Model, pairs []domain.KV) tea.Cmd) *kvEditorScreen {
	cp := append([]domain.KV(nil), pairs...)
	return &kvEditorScreen{title: title, pairs: cp, onDone: onDone}
}

func (s *kvEditorScreen) Init(*Model) tea.Cmd { return nil }

func (s *kvEditorScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch km.String() {
	case "esc":
		fn := s.onDone
		pairs := s.pairs
		m.pop()
		if fn != nil {
			return fn(m, pairs)
		}
		return nil
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(s.pairs)-1 {
			s.cursor++
		}
	case "a":
		return m.push(newPromptScreen("Add (format: key: value):", "", func(m *Model, v string) tea.Cmd {
			if kv, ok := parseKV(v); ok {
				s.pairs = append(s.pairs, kv)
				s.cursor = len(s.pairs) - 1
			} else {
				m.setError("Use the format key: value")
			}
			return nil
		}))
	case "e", "enter":
		if s.cursor < len(s.pairs) {
			cur := s.pairs[s.cursor]
			return m.push(newPromptScreen("Edit (format: key: value):", cur.Key+": "+cur.Value, func(m *Model, v string) tea.Cmd {
				if kv, ok := parseKV(v); ok {
					s.pairs[s.cursor] = kv
				} else {
					m.setError("Use the format key: value")
				}
				return nil
			}))
		}
	case "d":
		if s.cursor < len(s.pairs) {
			s.pairs = append(s.pairs[:s.cursor], s.pairs[s.cursor+1:]...)
			if s.cursor > 0 && s.cursor >= len(s.pairs) {
				s.cursor--
			}
		}
	}
	return nil
}

func parseKV(s string) (domain.KV, bool) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return domain.KV{}, false
	}
	key := strings.TrimSpace(s[:idx])
	val := strings.TrimSpace(s[idx+1:])
	if key == "" {
		return domain.KV{}, false
	}
	return domain.KV{Key: key, Value: val}, true
}

func (s *kvEditorScreen) View(m *Model) string {
	if len(s.pairs) == 0 {
		return "\n" + ui.Subtle.Render("No entries. Press ") + ui.Value.Render("a") + ui.Subtle.Render(" to add one.")
	}
	var b strings.Builder
	for i, kv := range s.pairs {
		line := fmt.Sprintf("%s: %s", kv.Key, kv.Value)
		if i == s.cursor {
			b.WriteString(ui.Selected.Render(" "+line+" ") + "\n")
		} else {
			b.WriteString("  " + ui.Value.Render(line) + "\n")
		}
	}
	return "\n" + b.String()
}

func (s *kvEditorScreen) Title() string { return s.title }

func (s *kvEditorScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Add, keys.Edit, keys.Delete, keys.Back}
}
