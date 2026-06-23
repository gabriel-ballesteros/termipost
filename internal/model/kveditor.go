package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// kvEditor is an inline editor for an ordered list of key/value pairs (request
// headers or query params), embedded in the request editor pane. Rows are added,
// edited, and removed in place — no screen transition. The cursor can rest on
// any row or on a trailing "add" affordance.
type kvEditor struct {
	pairs  []domain.KV
	cursor int // 0..len(pairs); len(pairs) is the "+ add" row

	editing bool
	isNew   bool
	field   int // 0 = key, 1 = value
	keyIn   textinput.Model
	valIn   textinput.Model
}

func newKVEditor() *kvEditor {
	e := &kvEditor{}
	e.keyIn = textinput.New()
	e.keyIn.Prompt = ""
	e.valIn = textinput.New()
	e.valIn.Prompt = ""
	return e
}

// reset clears any in-progress edit and parks the cursor at the top.
func (e *kvEditor) reset() {
	e.editing, e.isNew, e.field, e.cursor = false, false, 0, 0
	e.keyIn.Blur()
	e.valIn.Blur()
}

func (e *kvEditor) addRow() int { return len(e.pairs) } // index of the add affordance

func (e *kvEditor) update(m *Model, msg tea.KeyMsg) tea.Cmd {
	if e.editing {
		return e.updateEditing(m, msg)
	}
	switch msg.String() {
	case "up", "k":
		if e.cursor > 0 {
			e.cursor--
		}
	case "down", "j":
		if e.cursor < e.addRow() {
			e.cursor++
		}
	case "tab":
		if e.cursor < e.addRow() {
			e.cursor++
		} else {
			e.cursor = 0
		}
	case "enter", "i", "e":
		if e.cursor == e.addRow() {
			e.startNew()
		} else {
			e.startEdit()
		}
		return textinput.Blink
	case "a":
		e.startNew()
		return textinput.Blink
	case "d":
		if e.cursor < len(e.pairs) {
			e.pairs = append(e.pairs[:e.cursor], e.pairs[e.cursor+1:]...)
			if e.cursor > 0 && e.cursor >= len(e.pairs) {
				e.cursor--
			}
		}
	}
	return nil
}

func (e *kvEditor) startNew() {
	e.editing, e.isNew, e.field = true, true, 0
	e.keyIn.SetValue("")
	e.valIn.SetValue("")
	e.focusField()
}

func (e *kvEditor) startEdit() {
	cur := e.pairs[e.cursor]
	e.editing, e.isNew, e.field = true, false, 0
	e.keyIn.SetValue(cur.Key)
	e.valIn.SetValue(cur.Value)
	e.keyIn.CursorEnd()
	e.valIn.CursorEnd()
	e.focusField()
}

func (e *kvEditor) focusField() {
	if e.field == 0 {
		e.keyIn.Focus()
		e.valIn.Blur()
	} else {
		e.valIn.Focus()
		e.keyIn.Blur()
	}
}

func (e *kvEditor) updateEditing(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		e.cancel()
		return nil
	case "tab", "shift+tab":
		e.field = 1 - e.field
		e.focusField()
		return nil
	case "enter":
		key := strings.TrimSpace(e.keyIn.Value())
		if key == "" {
			m.setError("Key cannot be empty")
			return nil
		}
		kv := domain.KV{Key: key, Value: e.valIn.Value()}
		if e.isNew {
			e.pairs = append(e.pairs, kv)
			e.cursor = len(e.pairs) - 1
		} else {
			e.pairs[e.cursor] = kv
		}
		e.cancel()
		return nil
	}
	var cmd tea.Cmd
	if e.field == 0 {
		e.keyIn, cmd = e.keyIn.Update(msg)
	} else {
		e.valIn, cmd = e.valIn.Update(msg)
	}
	return cmd
}

func (e *kvEditor) cancel() {
	e.editing, e.isNew = false, false
	e.keyIn.Blur()
	e.valIn.Blur()
}

func (e *kvEditor) view(w, h int, focused bool) string {
	var b strings.Builder
	if len(e.pairs) == 0 && !e.editing {
		b.WriteString(ui.Subtle.Render("No entries. Press ") + ui.Value.Render("a") + ui.Subtle.Render(" to add one.\n"))
	}
	for i, kv := range e.pairs {
		if e.editing && !e.isNew && i == e.cursor {
			b.WriteString(e.editRow())
			continue
		}
		line := fmt.Sprintf("%s: %s", kv.Key, kv.Value)
		if i == e.cursor && focused && !e.editing {
			b.WriteString(ui.Selected.Render(" "+line+" ") + "\n")
		} else {
			b.WriteString("  " + ui.Value.Render(line) + "\n")
		}
	}
	// Trailing add affordance / new-row editor.
	if e.editing && e.isNew {
		b.WriteString(e.editRow())
	} else {
		add := "+ add"
		if e.cursor == e.addRow() && focused {
			b.WriteString(ui.Selected.Render(" "+add+" ") + "\n")
		} else {
			b.WriteString("  " + ui.Subtle.Render(add) + "\n")
		}
	}
	return b.String()
}

func (e *kvEditor) editRow() string {
	k := e.keyIn.View()
	v := e.valIn.View()
	return ui.FieldFocused.Render("▸ ") + k + ui.Subtle.Render(" : ") + v + "\n"
}
