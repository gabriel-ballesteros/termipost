package model

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// promptScreen is a reusable single-line text prompt pushed onto the stack. On
// Enter it pops itself and invokes onSubmit; on Esc it pops without calling it.
// Because it is a text-entry screen, global single-key shortcuts are inert here
// (it consumes all key input except Enter/Esc), satisfying edit-vs-nav mode.
type promptScreen struct {
	prompt   string
	input    textinput.Model
	onSubmit func(m *Model, value string) tea.Cmd
}

func newPromptScreen(prompt, initial string, onSubmit func(m *Model, value string) tea.Cmd) *promptScreen {
	ti := textinput.New()
	ti.SetValue(initial)
	ti.CursorEnd()
	ti.Focus()
	ti.Prompt = "> "
	return &promptScreen{prompt: prompt, input: ti, onSubmit: onSubmit}
}

func (s *promptScreen) Init(*Model) tea.Cmd { return textinput.Blink }

func (s *promptScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "enter":
			value := s.input.Value()
			fn := s.onSubmit
			m.pop()
			if fn != nil {
				return fn(m, value)
			}
			return nil
		case "esc":
			m.pop()
			return nil
		}
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

func (s *promptScreen) View(m *Model) string {
	return "\n" + ui.Label.Render(s.prompt) + "\n\n" + s.input.View()
}

func (s *promptScreen) Title() string { return "Input" }

// Crumb omits the prompt overlay from the breadcrumb trail.
func (s *promptScreen) Crumb() string { return "" }

func (s *promptScreen) HelpBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

// confirmScreen asks a yes/no question. y confirms (pops + onConfirm), n/esc
// cancels.
type confirmScreen struct {
	question  string
	onConfirm func(m *Model) tea.Cmd
}

func newConfirmScreen(question string, onConfirm func(m *Model) tea.Cmd) *confirmScreen {
	return &confirmScreen{question: question, onConfirm: onConfirm}
}

func (s *confirmScreen) Init(*Model) tea.Cmd { return nil }

func (s *confirmScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "y", "Y", "enter":
			fn := s.onConfirm
			m.pop()
			if fn != nil {
				return fn(m)
			}
			return nil
		case "n", "N", "esc":
			m.pop()
			return nil
		}
	}
	return nil
}

func (s *confirmScreen) View(m *Model) string {
	return "\n" + ui.Warn.Render(s.question) + "\n\n" + ui.Subtle.Render("y = yes   n = no")
}

func (s *confirmScreen) Title() string { return "Confirm" }

// Crumb omits the confirmation overlay from the breadcrumb trail.
func (s *confirmScreen) Crumb() string { return "" }

func (s *confirmScreen) HelpBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
	}
}
