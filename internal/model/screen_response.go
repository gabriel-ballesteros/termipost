package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
	"github.com/gabriel-ballesteros/termipost/internal/vars"
)

// responseScreen renders an HTTP response: a masked resolved-request line,
// status, timing, headers, and a scrollable, pretty-printed body. The raw body
// can be copied to the clipboard.
type responseScreen struct {
	content string
	body    []byte
	vp      viewport.Model
	ready   bool
}

func newResponseScreen(app *App, req domain.Request, resp *httpclient.Response) *responseScreen {
	var b strings.Builder

	// Resolved request line, with secret-sourced values masked so a token never
	// leaks into the preview.
	resolvedURL, _ := app.Resolver().Resolve(req.URL)
	masked := vars.Mask(resolvedURL, app.secrets)
	b.WriteString(ui.Label.Render("Request:  ") + ui.Value.Render(string(req.Method)+" "+masked) + "\n")

	statusStyle := ui.Good
	if resp.StatusCode >= 400 {
		statusStyle = ui.Bad
	} else if resp.StatusCode >= 300 {
		statusStyle = ui.Warn
	}
	b.WriteString(ui.Label.Render("Status:   ") + statusStyle.Render(resp.Status) + "\n")
	b.WriteString(ui.Label.Render("Time:     ") + ui.Value.Render(fmt.Sprintf("%dms", resp.Elapsed.Milliseconds())) + "\n")
	b.WriteString(ui.Label.Render("Size:     ") + ui.Value.Render(fmt.Sprintf("%d bytes", len(resp.Body))) + "\n\n")

	b.WriteString(ui.Subtle.Render("── Headers ──") + "\n")
	b.WriteString(renderHeaders(resp.Headers))
	b.WriteString("\n" + ui.Subtle.Render("── Body ──") + "\n")
	b.WriteString(prettyBody(resp))

	return &responseScreen{content: b.String(), body: resp.Body}
}

func renderHeaders(h http.Header) string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(ui.Label.Render(k+": ") + ui.Value.Render(strings.Join(h[k], ", ")) + "\n")
	}
	return b.String()
}

// prettyBody pretty-prints the body when it is JSON, otherwise returns it as-is.
func prettyBody(resp *httpclient.Response) string {
	body := resp.Body
	ct := resp.Headers.Get("Content-Type")
	if strings.Contains(ct, "json") || looksLikeJSON(body) {
		var out bytes.Buffer
		if err := json.Indent(&out, body, "", "  "); err == nil {
			return out.String()
		}
	}
	return string(body)
}

func looksLikeJSON(b []byte) bool {
	t := bytes.TrimSpace(b)
	return len(t) > 0 && (t[0] == '{' || t[0] == '[')
}

func (s *responseScreen) Init(*Model) tea.Cmd { return nil }

func (s *responseScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.ensureVP(msg.Width, m.bodyHeight())
		return nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.pop()
			return nil
		case "c", "y":
			// Ctrl+C is reserved for quit, so copy uses c / y (vim-style yank).
			if err := clipboard.WriteAll(string(s.body)); err != nil {
				m.setError("Copy failed: " + err.Error())
			} else {
				m.setStatus(fmt.Sprintf("Copied %d bytes to clipboard", len(s.body)))
			}
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

func (s *responseScreen) ensureVP(w, h int) {
	if !s.ready {
		s.vp = viewport.New(w, h)
		s.ready = true
	} else {
		s.vp.Width, s.vp.Height = w, h
	}
	s.vp.SetContent(s.content)
}

func (s *responseScreen) View(m *Model) string {
	if !s.ready {
		s.ensureVP(m.width, m.bodyHeight())
	}
	return s.vp.View()
}

func (s *responseScreen) Title() string { return "Response" }

func (s *responseScreen) HelpBindings() []key.Binding {
	return []key.Binding{keys.Up, keys.Down, keys.Copy, keys.Back}
}
