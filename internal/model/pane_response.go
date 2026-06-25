package model

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

type respTab int

const (
	rtBody respTab = iota
	rtHeaders
	respTabCount
)

var respTabNames = []string{"Body", "Headers"}

// responsePane shows the selected request's most recent response, with Body and
// Headers tabs and a scrollable, pretty-printed body.
type responsePane struct {
	app  *App
	tab  respTab
	vp   viewport.Model
	rdy  bool
	req  *domain.Request
	resp *httpclient.Response
	body []byte
	fold foldView
}

func newResponsePane(app *App) *responsePane { return &responsePane{app: app} }

// setResponse points the pane at a request and its response (resp may be nil).
func (p *responsePane) setResponse(req *domain.Request, resp *httpclient.Response) {
	p.req, p.resp = req, resp
	if resp != nil {
		p.body = resp.Body
	} else {
		p.body = nil
	}
	p.fold.setJSON(string(p.body))
}

func (p *responsePane) editing() bool { return false }

func (p *responsePane) Update(m *Model, msg tea.Msg, focused bool) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok || !focused {
		return nil
	}
	switch km.String() {
	case "]":
		p.tab = (p.tab + 1) % respTabCount
		return nil
	case "[":
		p.tab = (p.tab + respTabCount - 1) % respTabCount
		return nil
	case "c", "y":
		if p.resp == nil {
			return nil
		}
		if err := clipboard.WriteAll(string(p.body)); err != nil {
			m.setError("Copy failed: " + err.Error())
		} else {
			m.setStatus(fmt.Sprintf("Copied %d bytes to clipboard", len(p.body)))
		}
		return nil
	}
	// On a foldable JSON body, arrows drive the fold cursor and space toggles the
	// section; PageUp/Down and mouse still scroll the viewport (handled below).
	if p.tab == rtBody && p.fold.foldable {
		switch km.String() {
		case "up", "k":
			p.fold.moveUp()
			return nil
		case "down", "j":
			p.fold.moveDown()
			return nil
		case " ":
			p.fold.toggle()
			return nil
		}
	}
	if p.rdy {
		var cmd tea.Cmd
		p.vp, cmd = p.vp.Update(msg)
		return cmd
	}
	return nil
}

// statusBadge renders the response status with a severity colour.
func (p *responsePane) statusBadge() string {
	if p.resp == nil {
		return ui.Subtle.Render("no response")
	}
	style := ui.Good
	if p.resp.StatusCode >= 400 {
		style = ui.Bad
	} else if p.resp.StatusCode >= 300 {
		style = ui.Warn
	}
	return style.Render(p.resp.Status) +
		ui.Subtle.Render(fmt.Sprintf("  %dms  %dB", p.resp.Elapsed.Milliseconds(), len(p.resp.Body)))
}

func (p *responsePane) content(focused bool) string {
	if p.resp == nil {
		return ui.Subtle.Render("No response yet — press ") + ui.Value.Render("R") + ui.Subtle.Render(" to send the request.")
	}
	switch p.tab {
	case rtHeaders:
		return renderHeaders(p.resp.Headers)
	default:
		if p.fold.foldable {
			rows, _ := p.fold.renderLines(focused)
			return strings.Join(rows, "\n")
		}
		return prettyBody(p.resp)
	}
}

func (p *responsePane) View(m *Model, w, h int, focused bool) string {
	tabs := renderTabs(respTabNames, int(p.tab), focused)
	badge := p.statusBadge()
	header := tabs + "   " + badge

	vpH := h - 1
	if vpH < 1 {
		vpH = 1
	}
	if !p.rdy {
		p.vp = viewport.New(w, vpH)
		p.rdy = true
	} else {
		p.vp.Width, p.vp.Height = w, vpH
	}
	// Soft-wrap content to the current pane width so long body/header lines stay
	// visible (no horizontal scroll). Re-wrapping each render reflows on resize.
	wrapW := w
	if wrapW < 1 {
		wrapW = 1
	}
	p.vp.SetContent(lipgloss.NewStyle().Width(wrapW).Render(p.content(focused)))
	// Keep the fold cursor inside the viewport window, accounting for soft-wrap.
	if p.tab == rtBody && p.fold.foldable {
		top, bottom := p.fold.rowRange(wrapW)
		if top < p.vp.YOffset {
			p.vp.SetYOffset(top)
		} else if bottom >= p.vp.YOffset+p.vp.Height {
			p.vp.SetYOffset(bottom - p.vp.Height + 1)
		}
	}
	return header + "\n" + p.vp.View()
}

func (p *responsePane) Title() string { return "Response" }

func (p *responsePane) HelpBindings() []key.Binding {
	b := []key.Binding{keys.Up, keys.Down, keys.TabPrev, keys.TabNext, keys.Copy}
	if p.tab == rtBody && p.fold.foldable {
		b = append(b, keys.Fold)
	}
	return b
}

// renderTabs draws a tab strip with the active index highlighted.
func renderTabs(names []string, active int, focused bool) string {
	var parts []string
	for i, n := range names {
		if i == active {
			parts = append(parts, ui.TabActive.Render(n))
		} else {
			parts = append(parts, ui.TabInactive.Render(n))
		}
	}
	return strings.Join(parts, "")
}
