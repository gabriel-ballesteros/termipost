package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// assertionsScreen lists and manages the assertions attached to a request. It
// edits the request pointer in place and persists via its collection.
type assertionsScreen struct {
	app          *App
	collectionID string
	req          *domain.Request
	cursor       int
}

func newAssertionsScreen(app *App, collectionID string, req *domain.Request) *assertionsScreen {
	return &assertionsScreen{app: app, collectionID: collectionID, req: req}
}

func (s *assertionsScreen) persist(m *Model) {
	c := s.app.findCollection(s.collectionID)
	for i := range c.Requests {
		if c.Requests[i].ID == s.req.ID {
			c.Requests[i] = *s.req
			if err := s.app.saveCollection(*c); err != nil {
				m.setError("Save failed: " + err.Error())
			}
			return
		}
	}
}

func (s *assertionsScreen) Init(*Model) tea.Cmd { return nil }

func (s *assertionsScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch km.String() {
	case "esc":
		s.persist(m)
		m.pop()
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(s.req.Assertions)-1 {
			s.cursor++
		}
	case "a":
		return m.push(newAssertionEditScreen(domain.Assertion{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}, func(m *Model, a domain.Assertion) tea.Cmd {
			s.req.Assertions = append(s.req.Assertions, a)
			s.persist(m)
			return nil
		}))
	case "e", "enter":
		if s.cursor < len(s.req.Assertions) {
			idx := s.cursor
			return m.push(newAssertionEditScreen(s.req.Assertions[idx], func(m *Model, a domain.Assertion) tea.Cmd {
				s.req.Assertions[idx] = a
				s.persist(m)
				return nil
			}))
		}
	case "d":
		if s.cursor < len(s.req.Assertions) {
			s.req.Assertions = append(s.req.Assertions[:s.cursor], s.req.Assertions[s.cursor+1:]...)
			if s.cursor > 0 && s.cursor >= len(s.req.Assertions) {
				s.cursor--
			}
			s.persist(m)
		}
	}
	return nil
}

func (s *assertionsScreen) View(m *Model) string {
	header := ui.Label.Render("Request: ") + ui.Value.Render(s.req.Name) + "\n\n"
	if len(s.req.Assertions) == 0 {
		return header + ui.Subtle.Render("No assertions. Press ") + ui.Value.Render("a") + ui.Subtle.Render(" to add one.")
	}
	var b strings.Builder
	for i, a := range s.req.Assertions {
		line := describeAssertion(a)
		if i == s.cursor {
			b.WriteString(ui.Selected.Render(" "+line+" ") + "\n")
		} else {
			b.WriteString("  " + ui.Value.Render(line) + "\n")
		}
	}
	return header + b.String()
}

func describeAssertion(a domain.Assertion) string {
	switch a.Kind {
	case domain.AssertStatusCode:
		return "status code == " + a.Expected
	case domain.AssertHeader:
		return fmt.Sprintf("header %q %s %q", a.Target, a.Op, a.Expected)
	case domain.AssertBody:
		if a.Op == domain.OpJSONPath {
			return fmt.Sprintf("body json %q == %q", a.Target, a.Expected)
		}
		return fmt.Sprintf("body %s %q", a.Op, a.Expected)
	case domain.AssertLatency:
		return "latency <= " + a.Expected + "ms"
	}
	return string(a.Kind)
}

func (s *assertionsScreen) Title() string { return "Assertions" }

func (s *assertionsScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Add, keys.Edit, keys.Delete, keys.Back}
}

// ---- assertion edit form ----

type assertionField int

const (
	aKind assertionField = iota
	aOp
	aTarget
	aExpected
	aFieldCount
)

// assertionEditScreen is a small form to build/edit one assertion.
type assertionEditScreen struct {
	a        domain.Assertion
	focus    assertionField
	editing  bool
	target   textinput.Model
	expected textinput.Model
	onDone   func(m *Model, a domain.Assertion) tea.Cmd
}

var assertionKinds = []domain.AssertionKind{
	domain.AssertStatusCode, domain.AssertHeader, domain.AssertBody, domain.AssertLatency,
}

func opsFor(k domain.AssertionKind) []domain.MatchOp {
	switch k {
	case domain.AssertHeader:
		return []domain.MatchOp{domain.OpEquals, domain.OpContains, domain.OpRegex}
	case domain.AssertBody:
		return []domain.MatchOp{domain.OpContains, domain.OpEquals, domain.OpJSONPath}
	case domain.AssertLatency:
		return []domain.MatchOp{domain.OpMaxMS}
	default: // status_code
		return []domain.MatchOp{domain.OpEquals}
	}
}

// usesTarget reports whether the assertion needs a Target field.
func usesTarget(a domain.Assertion) bool {
	return a.Kind == domain.AssertHeader || (a.Kind == domain.AssertBody && a.Op == domain.OpJSONPath)
}

func newAssertionEditScreen(a domain.Assertion, onDone func(m *Model, a domain.Assertion) tea.Cmd) *assertionEditScreen {
	s := &assertionEditScreen{a: a, onDone: onDone}
	s.target = textinput.New()
	s.target.SetValue(a.Target)
	s.expected = textinput.New()
	s.expected.SetValue(a.Expected)
	return s
}

func (s *assertionEditScreen) Init(*Model) tea.Cmd { return nil }

func (s *assertionEditScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	if s.editing {
		if km.String() == "esc" || km.String() == "enter" {
			s.editing = false
			s.target.Blur()
			s.expected.Blur()
			return nil
		}
		var cmd tea.Cmd
		switch s.focus {
		case aTarget:
			s.target, cmd = s.target.Update(msg)
		case aExpected:
			s.expected, cmd = s.expected.Update(msg)
		}
		return cmd
	}

	switch km.String() {
	case "esc":
		m.pop()
	case "tab", "down", "j":
		s.focus = (s.focus + 1) % aFieldCount
	case "shift+tab", "up", "k":
		s.focus = (s.focus + aFieldCount - 1) % aFieldCount
	case "left", "h", "right", "l":
		s.cycle(km.String() == "right" || km.String() == "l")
	case "enter", "i":
		switch s.focus {
		case aKind, aOp:
			s.cycle(true)
		case aTarget:
			s.editing = true
			s.target.Focus()
		case aExpected:
			s.editing = true
			s.expected.Focus()
		}
	case "ctrl+s", "S":
		return s.save(m)
	}
	return nil
}

// cycle advances (or reverses) the selector for the focused Kind/Op field.
func (s *assertionEditScreen) cycle(forward bool) {
	switch s.focus {
	case aKind:
		idx := 0
		for i, k := range assertionKinds {
			if k == s.a.Kind {
				idx = i
			}
		}
		idx = wrap(idx, forward, len(assertionKinds))
		s.a.Kind = assertionKinds[idx]
		// reset op to the first valid op for the new kind
		s.a.Op = opsFor(s.a.Kind)[0]
	case aOp:
		ops := opsFor(s.a.Kind)
		idx := 0
		for i, o := range ops {
			if o == s.a.Op {
				idx = i
			}
		}
		idx = wrap(idx, forward, len(ops))
		s.a.Op = ops[idx]
	}
}

func wrap(idx int, forward bool, n int) int {
	if forward {
		return (idx + 1) % n
	}
	return (idx + n - 1) % n
}

func (s *assertionEditScreen) save(m *Model) tea.Cmd {
	s.a.Target = strings.TrimSpace(s.target.Value())
	s.a.Expected = strings.TrimSpace(s.expected.Value())
	if !usesTarget(s.a) {
		s.a.Target = ""
	}
	if s.a.Expected == "" {
		m.setError("Expected value is required")
		return nil
	}
	if usesTarget(s.a) && s.a.Target == "" {
		m.setError("Target (header name / json path) is required")
		return nil
	}
	fn := s.onDone
	a := s.a
	m.pop()
	if fn != nil {
		return fn(m, a)
	}
	return nil
}

func (s *assertionEditScreen) View(m *Model) string {
	var b strings.Builder
	b.WriteString(s.row("Kind", aKind, string(s.a.Kind)))
	b.WriteString(s.row("Operator", aOp, string(s.a.Op)))
	if usesTarget(s.a) {
		label := "Target"
		if s.a.Kind == domain.AssertBody {
			label = "JSON path"
		}
		b.WriteString(s.row(label, aTarget, s.fieldText(aTarget, s.a.Target)))
	}
	expLabel := "Expected"
	if s.a.Kind == domain.AssertLatency {
		expLabel = "Max ms"
	}
	b.WriteString(s.row(expLabel, aExpected, s.fieldText(aExpected, s.expected.Value())))
	b.WriteString("\n" + ui.Subtle.Render("←/→ change selector · enter edit field · ctrl+s save"))
	return b.String()
}

func (s *assertionEditScreen) row(label string, f assertionField, value string) string {
	lbl := fmt.Sprintf("%-11s", label+":")
	if f == s.focus {
		return ui.FieldFocused.Render("▸ "+lbl) + " " + value + "\n"
	}
	return ui.Label.Render("  "+lbl) + " " + ui.Value.Render(value) + "\n"
}

func (s *assertionEditScreen) fieldText(f assertionField, fallback string) string {
	if s.editing && s.focus == f {
		switch f {
		case aTarget:
			return s.target.View()
		case aExpected:
			return s.expected.View()
		}
	}
	if fallback == "" {
		return ui.Subtle.Render("(empty)")
	}
	return fallback
}

func (s *assertionEditScreen) Title() string { return "Edit assertion" }

func (s *assertionEditScreen) HelpBindings() []key.Binding {
	if s.editing {
		return []key.Binding{key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "stop editing"))}
	}
	return []key.Binding{keys.Up, keys.Down, keys.Left, keys.Right, keys.Enter,
		key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")), keys.Back}
}
