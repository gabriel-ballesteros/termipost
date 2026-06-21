package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// runResultsScreen shows the outcome of running a single request or a whole
// collection, scrollable in a viewport.
type runResultsScreen struct {
	title   string
	content string
	vp      viewport.Model
	ready   bool
}

func newRunResultsScreen(collectionName string, res domain.CollectionRunResult) *runResultsScreen {
	var b strings.Builder
	summary := fmt.Sprintf("%s  %s  %s",
		ui.Good.Render(fmt.Sprintf("%d passed", res.Passed)),
		ui.Bad.Render(fmt.Sprintf("%d failed", res.Failed)),
		ui.Subtle.Render(fmt.Sprintf("%d skipped", res.Skipped)),
	)
	b.WriteString(ui.Label.Render("Collection: ") + ui.Value.Render(collectionName) + "\n")
	b.WriteString(summary + "\n\n")
	for _, r := range res.Results {
		b.WriteString(renderRunResult(r))
		b.WriteString("\n")
	}
	return &runResultsScreen{title: "Run results", content: b.String()}
}

func newSingleRunResultsScreen(res domain.RunResult) *runResultsScreen {
	return &runResultsScreen{title: "Test result", content: renderRunResult(res)}
}

func renderRunResult(r domain.RunResult) string {
	var b strings.Builder
	var badge string
	switch r.Status {
	case domain.RunPassed:
		badge = ui.Good.Render("PASS")
	case domain.RunFailed:
		badge = ui.Bad.Render("FAIL")
	case domain.RunSkipped:
		badge = ui.Subtle.Render("SKIP")
	default:
		badge = ui.Bad.Render("ERROR")
	}
	b.WriteString(fmt.Sprintf("%s  %s", badge, ui.Value.Render(r.RequestName)))
	if r.Status != domain.RunSkipped && r.Status != domain.RunError {
		b.WriteString(ui.Subtle.Render(fmt.Sprintf("  [%d, %dms]", r.StatusCode, r.Elapsed.Milliseconds())))
	}
	b.WriteString("\n")
	if r.Status == domain.RunSkipped {
		b.WriteString("   " + ui.Subtle.Render("no assertions") + "\n")
	}
	if r.Err != "" {
		b.WriteString("   " + ui.Bad.Render("error: "+r.Err) + "\n")
	}
	for _, a := range r.Assertions {
		mark := ui.Good.Render("✓")
		if !a.Passed {
			mark = ui.Bad.Render("✗")
		}
		b.WriteString("   " + mark + " " + ui.Subtle.Render(a.Detail) + "\n")
	}
	return b.String()
}

func (s *runResultsScreen) Init(*Model) tea.Cmd { return nil }

func (s *runResultsScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.ensureVP(msg.Width, m.bodyHeight())
		return nil
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.pop()
			return nil
		}
	}
	if !s.ready {
		return nil
	}
	var cmd tea.Cmd
	s.vp, cmd = s.vp.Update(msg)
	return cmd
}

func (s *runResultsScreen) ensureVP(w, h int) {
	if !s.ready {
		s.vp = viewport.New(w, h)
		s.ready = true
	} else {
		s.vp.Width, s.vp.Height = w, h
	}
	s.vp.SetContent(s.content)
}

func (s *runResultsScreen) View(m *Model) string {
	if !s.ready {
		s.ensureVP(m.width, m.bodyHeight())
	}
	return s.vp.View()
}

func (s *runResultsScreen) Title() string { return s.title }

func (s *runResultsScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Back}
}
