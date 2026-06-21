package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

type reqField int

const (
	fName reqField = iota
	fMethod
	fURL
	fHeaders
	fParams
	fBody
	fieldCount
)

// requestEditScreen edits a request and can send it or run it as a test.
//
// It has two input modes (satisfying edit-vs-navigation): navigation mode, where
// single keys are actions (tab to move, s to send, R to run, a for assertions);
// and edit mode, entered with Enter on a text field, where keystrokes go to the
// focused input and Esc returns to navigation.
type requestEditScreen struct {
	app          *App
	collectionID string
	req          domain.Request

	focus     reqField
	editing   bool
	methodIdx int

	name textinput.Model
	url  textinput.Model
	body textarea.Model

	sending bool
	spin    spinner.Model
}

func newRequestEditScreen(app *App, collectionID, requestID string) *requestEditScreen {
	s := &requestEditScreen{app: app, collectionID: collectionID}
	c := app.findCollection(collectionID)
	for _, r := range c.Requests {
		if r.ID == requestID {
			s.req = r
			break
		}
	}

	s.name = textinput.New()
	s.name.SetValue(s.req.Name)
	s.url = textinput.New()
	s.url.SetValue(s.req.URL)
	s.url.Width = 60
	s.body = textarea.New()
	s.body.SetValue(s.req.Body)
	s.body.ShowLineNumbers = false

	for i, mth := range domain.Methods {
		if mth == s.req.Method {
			s.methodIdx = i
		}
	}
	s.spin = spinner.New()
	s.spin.Spinner = spinner.Dot
	return s
}

func (s *requestEditScreen) Init(*Model) tea.Cmd { return nil }

func (s *requestEditScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.body.SetWidth(min(msg.Width-2, 80))
		s.body.SetHeight(max(m.bodyHeight()-9, 3))
		return nil

	case spinner.TickMsg:
		if s.sending {
			var cmd tea.Cmd
			s.spin, cmd = s.spin.Update(msg)
			return cmd
		}
		return nil

	case sendResultMsg:
		s.sending = false
		if msg.err != nil {
			m.setError("Request failed: " + msg.err.Error())
			return nil
		}
		if len(msg.unresolved) > 0 {
			m.setStatus("Sent (unresolved vars: " + strings.Join(msg.unresolved, ", ") + ")")
		} else {
			m.setStatus("Response received")
		}
		return m.push(newResponseScreen(s.app, s.req, msg.resp))

	case reqRunMsg:
		s.sending = false
		return m.push(newSingleRunResultsScreen(msg.result))

	case tea.KeyMsg:
		if s.editing {
			return s.updateEditing(m, msg)
		}
		return s.updateNav(m, msg)
	}
	return nil
}

func (s *requestEditScreen) updateNav(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.pop()
	case "tab", "down", "j":
		s.focus = (s.focus + 1) % fieldCount
	case "shift+tab", "up", "k":
		s.focus = (s.focus + fieldCount - 1) % fieldCount
	case "left", "h":
		if s.focus == fMethod {
			s.methodIdx = (s.methodIdx + len(domain.Methods) - 1) % len(domain.Methods)
		}
	case "right", "l":
		if s.focus == fMethod {
			s.methodIdx = (s.methodIdx + 1) % len(domain.Methods)
		}
	case "enter", "i":
		switch s.focus {
		case fName:
			s.editing = true
			s.name.Focus()
		case fURL:
			s.editing = true
			s.url.Focus()
		case fBody:
			s.editing = true
			s.body.Focus()
		case fMethod:
			s.methodIdx = (s.methodIdx + 1) % len(domain.Methods)
		case fHeaders:
			return m.push(newKVEditorScreen("Headers", s.req.Headers, func(m *Model, p []domain.KV) tea.Cmd {
				s.req.Headers = p
				return nil
			}))
		case fParams:
			return m.push(newKVEditorScreen("Query params", s.req.QueryParams, func(m *Model, p []domain.KV) tea.Cmd {
				s.req.QueryParams = p
				return nil
			}))
		}
	case "ctrl+s":
		if err := s.persist(m); err != nil {
			m.setError("Save failed: " + err.Error())
		} else {
			m.setStatus("Saved")
		}
	case "R":
		// Run: send the request and show the response, ignoring assertions.
		if err := s.persist(m); err != nil {
			m.setError("Save failed: " + err.Error())
			return nil
		}
		s.sending = true
		m.setStatus("Running…")
		return tea.Batch(s.spin.Tick, sendCmd(s.app, s.req))
	case "T":
		// Test: send the request and evaluate its assertions.
		if err := s.persist(m); err != nil {
			m.setError("Save failed: " + err.Error())
			return nil
		}
		if len(s.req.Assertions) == 0 {
			m.setError("Add assertions first (press a) to test this request")
			return nil
		}
		s.sending = true
		m.setStatus("Testing…")
		return tea.Batch(s.spin.Tick, runRequestCmd(s.app, s.req))
	case "a":
		return m.push(newAssertionsScreen(s.app, s.collectionID, &s.req))
	}
	return nil
}

func (s *requestEditScreen) updateEditing(m *Model, msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		s.editing = false
		s.name.Blur()
		s.url.Blur()
		s.body.Blur()
		return nil
	}
	// Enter commits single-line fields; the body keeps Enter for newlines.
	if msg.String() == "enter" && (s.focus == fName || s.focus == fURL) {
		s.editing = false
		s.name.Blur()
		s.url.Blur()
		return nil
	}
	var cmd tea.Cmd
	switch s.focus {
	case fName:
		s.name, cmd = s.name.Update(msg)
	case fURL:
		s.url, cmd = s.url.Update(msg)
	case fBody:
		s.body, cmd = s.body.Update(msg)
	}
	return cmd
}

// syncFromInputs copies the live input widgets back into the working request.
func (s *requestEditScreen) syncFromInputs() {
	s.req.Name = strings.TrimSpace(s.name.Value())
	s.req.URL = strings.TrimSpace(s.url.Value())
	s.req.Method = domain.Methods[s.methodIdx]
	s.req.Body = s.body.Value()
}

// persist writes the working request back into its collection on disk.
func (s *requestEditScreen) persist(m *Model) error {
	s.syncFromInputs()
	c := s.app.findCollection(s.collectionID)
	for i := range c.Requests {
		if c.Requests[i].ID == s.req.ID {
			c.Requests[i] = s.req
			return s.app.saveCollection(*c)
		}
	}
	return nil
}

func (s *requestEditScreen) View(m *Model) string {
	s.syncFromInputs()
	var b strings.Builder

	b.WriteString(s.fieldRow("Name", fName, s.fieldText(fName, s.req.Name)))
	b.WriteString(s.fieldRow("Method", fMethod, fmt.Sprintf("< %s >", s.req.Method)))
	b.WriteString(s.fieldRow("URL", fURL, s.fieldText(fURL, s.req.URL)))
	b.WriteString(s.fieldRow("Headers", fHeaders, fmt.Sprintf("%d entry(ies)  (enter to edit)", len(s.req.Headers))))
	b.WriteString(s.fieldRow("Params", fParams, fmt.Sprintf("%d entry(ies)  (enter to edit)", len(s.req.QueryParams))))
	b.WriteString(s.fieldRow("Assertions", -1, fmt.Sprintf("%d  (press a to edit)", len(s.req.Assertions))))

	bodyLabel := ui.Label.Render("Body:")
	if s.focus == fBody {
		bodyLabel = ui.FieldFocused.Render("Body:")
	}
	b.WriteString("\n" + bodyLabel + "\n")
	if s.editing && s.focus == fBody {
		b.WriteString(s.body.View())
	} else {
		b.WriteString(ui.Box.Render(bodyPreview(s.req.Body)))
	}

	if s.sending {
		b.WriteString("\n\n" + s.spin.View() + ui.Subtle.Render(" working…"))
	}
	return b.String()
}

func (s *requestEditScreen) fieldRow(label string, f reqField, value string) string {
	lbl := fmt.Sprintf("%-11s", label+":")
	if f == s.focus {
		return ui.FieldFocused.Render("▸ "+lbl) + " " + value + "\n"
	}
	return ui.Label.Render("  "+lbl) + " " + ui.Value.Render(value) + "\n"
}

// fieldText shows the live input when its field is being edited.
func (s *requestEditScreen) fieldText(f reqField, fallback string) string {
	if s.editing && s.focus == f {
		switch f {
		case fName:
			return s.name.View()
		case fURL:
			return s.url.View()
		}
	}
	if fallback == "" {
		return ui.Subtle.Render("(empty)")
	}
	return fallback
}

func bodyPreview(body string) string {
	if strings.TrimSpace(body) == "" {
		return ui.Subtle.Render("(empty — press enter to edit)")
	}
	lines := strings.Split(body, "\n")
	if len(lines) > 6 {
		lines = append(lines[:6], "…")
	}
	return strings.Join(lines, "\n")
}

func (s *requestEditScreen) Title() string {
	if s.editing {
		return "Edit request (editing field — esc to stop)"
	}
	return "Edit request"
}

func (s *requestEditScreen) HelpBindings() []key.Binding {
	if s.editing {
		return []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "stop editing")),
		}
	}
	return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.Run, keys.Test,
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "assertions")), keys.Save, keys.Back}
}
