// Package model implements the Bubble Tea TUI for termipost: a root model that
// owns shared state and a navigation stack of screens, each of which renders a
// view and a contextual action/help bar.
package model

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// Screen is a single view in the navigation stack. Update may mutate the root
// Model (to push/pop screens, set status, or change app data) and returns a
// command. Screens own their own key handling.
type Screen interface {
	Init(m *Model) tea.Cmd
	Update(m *Model, msg tea.Msg) tea.Cmd
	View(m *Model) string
	Title() string
	HelpBindings() []key.Binding
}

// Model is the root Bubble Tea model.
type Model struct {
	app    *App
	stack  []Screen
	width  int
	height int

	status    string
	statusErr bool

	help help.Model

	version string
}

// New builds the root model with the collection list as the initial screen.
func New(app *App, version string, loadErrs []error) *Model {
	m := &Model{app: app, version: version, help: help.New()}
	m.push(newCollectionListScreen(app))
	if len(loadErrs) > 0 {
		var names []string
		for _, e := range loadErrs {
			names = append(names, e.Error())
		}
		m.setError("Some data files could not be loaded: " + strings.Join(names, "; "))
	}
	return m
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	if s := m.top(); s != nil {
		return s.Init(m)
	}
	return nil
}

// --- navigation stack ---

func (m *Model) top() Screen {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

func (m *Model) push(s Screen) tea.Cmd {
	m.stack = append(m.stack, s)
	m.clearStatus()
	cmd := s.Init(m)
	// Bubble Tea only emits WindowSizeMsg at startup and on resize, so a screen
	// opened later never learns the terminal size on its own. Hand it the
	// current size now (once known) so list/viewport screens render their
	// content immediately instead of just a pagination footer.
	if m.width > 0 {
		sizeCmd := s.Update(m, tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmd = tea.Batch(cmd, sizeCmd)
	}
	return cmd
}

// pop removes the top screen. It is a no-op when only one screen remains.
func (m *Model) pop() {
	if len(m.stack) > 1 {
		m.stack = m.stack[:len(m.stack)-1]
		m.clearStatus()
	}
}

func (m *Model) depth() int { return len(m.stack) }

func (m *Model) setStatus(s string) { m.status, m.statusErr = s, false }
func (m *Model) setError(s string)  { m.status, m.statusErr = s, true }
func (m *Model) clearStatus()       { m.status, m.statusErr = "", false }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
		// fall through to let the active screen react to resize too
	case tea.KeyMsg:
		// Ctrl+C always quits, regardless of screen or edit mode.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if s := m.top(); s != nil {
		cmd := s.Update(m, msg)
		return m, cmd
	}
	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	s := m.top()
	if s == nil || m.width == 0 {
		return "loading…"
	}

	env := ui.Subtle.Render("   env: ") + ui.Value.Render("none")
	if e := m.app.ActiveEnv(); e != nil {
		env = ui.Subtle.Render("   env: ") + ui.Value.Render(e.Name)
	}
	name := ui.Title.Render(m.appName())
	// Breadcrumb fills the space between the name and the env indicator, eliding
	// from the front when the terminal is too narrow.
	used := lipgloss.Width(name) + lipgloss.Width(env) + 3
	crumbs := ui.Subtle.Render("  " + elideLeft(m.breadcrumb(), max(m.width-used, 0)))
	title := name + crumbs + env

	help := ui.HelpBar.Render(m.help.ShortHelpView(s.HelpBindings()))

	status := ""
	if m.status != "" {
		if m.statusErr {
			status = ui.Error.Render(m.status)
		} else {
			status = ui.Subtle.Render(m.status)
		}
	}

	// Reserve lines for title, blank, status, help.
	chrome := 4
	bodyHeight := m.height - chrome
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	body := s.View(m)
	body = lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(body)

	return strings.Join([]string{title, body, status, help}, "\n")
}

// bodyHeight returns the height available to a screen's body content.
func (m *Model) bodyHeight() int {
	h := m.height - 4
	if h < 1 {
		return 1
	}
	return h
}

// crumbSep separates breadcrumb segments.
const crumbSep = " › "

// crumber lets a screen provide a short breadcrumb label, or "" to be omitted
// from the trail (used by transient overlays like prompts and confirmations).
type crumber interface{ Crumb() string }

// appName renders "termipost vX.Y.Z", or "termipost (dev)" for local builds.
func (m *Model) appName() string {
	if m.version == "" || m.version == "dev" {
		return "termipost (dev)"
	}
	return "termipost v" + strings.TrimPrefix(m.version, "v")
}

// breadcrumb joins the open screens' labels, skipping omitted ones.
func (m *Model) breadcrumb() string {
	var parts []string
	for _, s := range m.stack {
		label := s.Title()
		if c, ok := s.(crumber); ok {
			label = c.Crumb()
		}
		if label != "" {
			parts = append(parts, label)
		}
	}
	return strings.Join(parts, crumbSep)
}

// elideLeft trims s from the left to fit width display columns, prefixing an
// ellipsis when it had to cut. width <= 0 yields an empty string.
func elideLeft(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	// Drop leading runes until the remainder (plus "…") fits.
	for i := 0; i < len(r); i++ {
		cand := "…" + string(r[i:])
		if lipgloss.Width(cand) <= width {
			return cand
		}
	}
	return "…"
}
