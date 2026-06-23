package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// pane is one region of the composite workspace. Unlike a Screen it never owns
// the navigation stack: it renders into an explicit width/height, reacts only to
// the messages the workspace routes to it, and reports whether it is mid text
// entry (so the workspace can suspend its global keys).
type pane interface {
	Update(m *Model, msg tea.Msg, focused bool) tea.Cmd
	View(m *Model, w, h int, focused bool) string
	HelpBindings() []key.Binding
	Title() string
	editing() bool
}

// paneID identifies a pane for focus movement.
type paneID int

const (
	paneTree paneID = iota
	paneEditor
	paneResponse
	paneCount
)

// Minimum sizes used to decide between the full multi-pane layout and the
// single-pane fallback (see design Decision 9).
const (
	minTreeW   = 24
	minRightW  = 40
	minPaneH   = 6
	topBarRows = 1
)

// workspaceScreen is the primary composite view: a top method+URL+Send bar, a
// collapsible tree pane on the left, and a request editor + response pane
// stacked on the right. Focus moves between panes with the vim window chord
// (ctrl+w then h/j/k/l); everything else stays in-pane.
type workspaceScreen struct {
	app *App

	tree     *treePane
	editor   *editorPane
	response *responsePane

	focus      paneID
	pendingWin bool // ctrl+w pressed, awaiting a direction

	spin    spinner.Model
	sending bool

	// lastResp keeps the most recent response per request id so switching the
	// selection in the tree restores what that request last returned.
	lastResp map[string]*httpclient.Response
}

func newWorkspaceScreen(app *App) *workspaceScreen {
	w := &workspaceScreen{app: app, lastResp: map[string]*httpclient.Response{}}
	w.spin = spinner.New()
	w.spin.Spinner = spinner.Dot

	w.tree = newTreePane(app)
	w.editor = newEditorPane(app)
	w.response = newResponsePane(app)
	w.tree.activate = w.selectRequest

	// Open the first request, if any, so the panes are populated on launch.
	if cid, rid := w.tree.firstRequest(); rid != "" {
		w.editor.load(cid, rid)
		w.refreshResponse()
	}
	return w
}

func (w *workspaceScreen) panes() []pane { return []pane{w.tree, w.editor, w.response} }

func (w *workspaceScreen) focused() pane { return w.panes()[w.focus] }

func (w *workspaceScreen) Init(*Model) tea.Cmd { return nil }

// selectRequest loads a request into the editor + response panes, guarding
// against discarding unsaved edits.
func (w *workspaceScreen) selectRequest(m *Model, collID, reqID string) tea.Cmd {
	if w.editor.reqID == reqID && w.editor.loaded {
		return nil
	}
	load := func(m *Model) tea.Cmd {
		w.editor.load(collID, reqID)
		w.refreshResponse()
		w.focus = paneEditor
		return nil
	}
	if w.editor.dirty() {
		return m.push(newConfirmScreen("Discard unsaved changes to this request?", load))
	}
	return load(m)
}

// refreshResponse points the response pane at the selected request's last
// response (or clears it if there is none).
func (w *workspaceScreen) refreshResponse() {
	req := w.editor.currentReq()
	w.response.setResponse(&req, w.lastResp[w.editor.reqID])
}

func (w *workspaceScreen) Update(m *Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if w.sending {
			var cmd tea.Cmd
			w.spin, cmd = w.spin.Update(msg)
			return cmd
		}
		return nil

	case sendResultMsg:
		w.sending = false
		if msg.err != nil {
			m.setError("Request failed: " + msg.err.Error())
			return nil
		}
		if len(msg.unresolved) > 0 {
			m.setStatus("Sent (unresolved vars: " + strings.Join(msg.unresolved, ", ") + ")")
		} else {
			m.setStatus("Response received")
		}
		w.lastResp[w.editor.reqID] = msg.resp
		w.refreshResponse()
		w.focus = paneResponse
		return nil

	case reqRunMsg:
		w.sending = false
		return m.push(newSingleRunResultsScreen(msg.result))

	case collRunMsg:
		return m.push(newRunResultsScreen(msg.collectionName, msg.result))

	case tea.KeyMsg:
		return w.handleKey(m, msg)
	}

	// Non-key, non-lifecycle messages are broadcast to every pane so background
	// widgets (viewport, inputs) stay live regardless of focus.
	var cmds []tea.Cmd
	for i, p := range w.panes() {
		cmds = append(cmds, p.Update(m, msg, paneID(i) == w.focus))
	}
	return tea.Batch(cmds...)
}

func (w *workspaceScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	// While a pane is mid text-entry, every key belongs to it.
	if w.focused().editing() {
		return w.focused().Update(m, msg, true)
	}

	// Window chord: ctrl+w then a direction.
	if w.pendingWin {
		w.pendingWin = false
		switch msg.String() {
		case "h", "j", "k", "l":
			w.moveFocus(msg.String())
			return nil
		}
		// Any other key falls through as a normal key this turn.
	}

	switch {
	case key.Matches(msg, keys.Window):
		w.pendingWin = true
		return nil
	case msg.String() == "q":
		return w.quit(m)
	case msg.String() == "R":
		return w.send(m, false)
	case msg.String() == "T":
		return w.send(m, true)
	case key.Matches(msg, keys.Save):
		if err := w.editor.persist(m); err != nil {
			m.setError("Save failed: " + err.Error())
		} else {
			m.setStatus("Saved")
			w.refreshResponse()
		}
		return nil
	case msg.String() == "E":
		return m.push(newEnvListScreen(w.app))
	}

	return w.focused().Update(m, msg, true)
}

// quit exits, guarding the soft-quit key against discarding unsaved edits. The
// hard-quit key (Ctrl+C) is handled in the root model and never reaches here.
func (w *workspaceScreen) quit(m *Model) tea.Cmd {
	if w.editor.dirty() {
		return m.push(newConfirmScreen("Discard unsaved changes and quit?", func(*Model) tea.Cmd {
			return tea.Quit
		}))
	}
	return tea.Quit
}

func (w *workspaceScreen) send(m *Model, test bool) tea.Cmd {
	if !w.editor.loaded {
		m.setError("No request selected")
		return nil
	}
	if err := w.editor.persist(m); err != nil {
		m.setError("Save failed: " + err.Error())
		return nil
	}
	req := w.editor.currentReq()
	if test {
		if len(req.Assertions) == 0 {
			m.setError("Add assertions first (press a) to test this request")
			return nil
		}
		w.sending = true
		m.setStatus("Testing…")
		return tea.Batch(w.spin.Tick, runRequestCmd(w.app, req))
	}
	w.sending = true
	m.setStatus("Running…")
	return tea.Batch(w.spin.Tick, sendCmd(w.app, req))
}

// neighbors maps the directional move available from each pane.
func (w *workspaceScreen) moveFocus(dir string) {
	type mv struct {
		from paneID
		dir  string
		to   paneID
	}
	moves := []mv{
		{paneTree, "l", paneEditor},
		{paneEditor, "h", paneTree},
		{paneEditor, "j", paneResponse},
		{paneResponse, "h", paneTree},
		{paneResponse, "k", paneEditor},
	}
	for _, mvv := range moves {
		if mvv.from == w.focus && mvv.dir == dir {
			w.focus = mvv.to
			return
		}
	}
}

// --- layout ---

// multiPane reports whether the terminal is large enough for the full grid.
func (w *workspaceScreen) multiPane(bodyW, bodyH int) bool {
	gridH := bodyH - topBarRows
	return bodyW >= minTreeW+minRightW+4 && gridH >= 2*minPaneH
}

func (w *workspaceScreen) View(m *Model) string {
	bodyW, bodyH := m.width, m.bodyHeight()
	topBar := w.topBar(bodyW)
	gridH := bodyH - topBarRows
	if gridH < 1 {
		gridH = 1
	}

	if !w.multiPane(bodyW, bodyH) {
		// Single-pane fallback: only the focused pane, full area.
		p := w.focused()
		box := w.renderPane(p, paneID(w.focus), bodyW, gridH)
		return topBar + "\n" + box
	}

	treeW := bodyW * 30 / 100
	if treeW < minTreeW {
		treeW = minTreeW
	}
	if treeW > 45 {
		treeW = 45
	}
	rightW := bodyW - treeW

	editorH := gridH * 45 / 100
	if editorH < minPaneH {
		editorH = minPaneH
	}
	responseH := gridH - editorH
	if responseH < minPaneH {
		responseH = minPaneH
		editorH = gridH - responseH
	}

	treeBox := w.renderPane(w.tree, paneTree, treeW, gridH)
	editorBox := w.renderPane(w.editor, paneEditor, rightW, editorH)
	responseBox := w.renderPane(w.response, paneResponse, rightW, responseH)

	right := lipgloss.JoinVertical(lipgloss.Left, editorBox, responseBox)
	grid := lipgloss.JoinHorizontal(lipgloss.Top, treeBox, right)
	return topBar + "\n" + grid
}

// renderPane wraps a pane's content in a bordered box of total size w x h,
// highlighting the border when focused. The pane renders into the inner area.
func (w *workspaceScreen) renderPane(p pane, id paneID, totalW, totalH int) string {
	innerW := totalW - 2
	innerH := totalH - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}
	focused := id == w.focus
	titleStyle := ui.PaneTitle
	border := ui.Pane
	if focused {
		titleStyle = ui.PaneTitleFocused
		border = ui.PaneFocused
	}

	title := titleStyle.Render(p.Title())
	content := p.View(nil, innerW, innerH-1, focused)
	inner := lipgloss.NewStyle().Width(innerW).Height(innerH).MaxHeight(innerH).
		Render(title + "\n" + content)
	return border.Width(innerW).Height(innerH).Render(inner)
}

// topBar renders the method chip, URL, and Send affordance.
func (w *workspaceScreen) topBar(width int) string {
	method := "GET"
	url := ui.Subtle.Render("(no request selected)")
	if w.editor.loaded {
		r := w.editor.currentReq()
		method = string(r.Method)
		if strings.TrimSpace(r.URL) != "" {
			url = ui.Value.Render(r.URL)
		} else {
			url = ui.Subtle.Render("(no url)")
		}
	}
	send := ui.SendButton.Render("Send")
	if w.sending {
		send = w.spin.View() + ui.Subtle.Render(" sending")
	}

	left := ui.Method.Render(method) + " " + url
	gap := width - lipgloss.Width(left) - lipgloss.Width(send)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + send
}

func (w *workspaceScreen) Title() string { return "Workspace" }

func (w *workspaceScreen) Crumb() string { return "Workspace" }

var keyEnvironments = key.NewBinding(key.WithKeys("E"), key.WithHelp("E", "environments"))

func (w *workspaceScreen) HelpBindings() []key.Binding {
	b := []key.Binding{keys.Window, keys.Quit, keys.Run, keys.Test, keys.Save, keyEnvironments}
	b = append(b, w.focused().HelpBindings()...)
	return b
}

// fmtRequestRow formats a request's method+name the way the tree shows it.
func fmtRequestRow(r domain.Request) string {
	return fmt.Sprintf("%-6s %s", r.Method, r.Name)
}
