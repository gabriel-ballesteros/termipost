package model

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// simpleItem is a list row carrying a stable id plus display text.
type simpleItem struct {
	id    string
	title string
	desc  string
}

func (i simpleItem) Title() string       { return i.title }
func (i simpleItem) Description() string { return i.desc }
func (i simpleItem) FilterValue() string { return i.title }

// newList builds a compact bubbles/list with our chrome stripped (we render the
// title and help bar ourselves) and filtering disabled so single-key actions
// like n/e/d are not captured as text.
func newList(items []list.Item, withDesc bool) list.Model {
	d := list.NewDefaultDelegate()
	d.ShowDescription = withDesc
	d.SetSpacing(0)
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(ui.ColorAccent).BorderForeground(ui.ColorAccent)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(ui.ColorAccent).BorderForeground(ui.ColorAccent)

	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(true)
	return l
}

// selectedID returns the id of the highlighted item, or "" if none.
func selectedID(l list.Model) string {
	if it, ok := l.SelectedItem().(simpleItem); ok {
		return it.id
	}
	return ""
}
