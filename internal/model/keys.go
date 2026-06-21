package model

import "github.com/charmbracelet/bubbles/key"

// keys holds the key bindings shared across screens. Each screen advertises the
// subset relevant to it via HelpBindings so the action bar always matches the
// real behavior.
var keys = struct {
	Up, Down, Left, Right key.Binding
	Enter, Back, Quit     key.Binding
	New, Edit, Delete     key.Binding
	Rename, Run, Test     key.Binding
	Send, Copy            key.Binding
	Save, Add, Toggle     key.Binding
	Reveal, SetActive     key.Binding
	Tab, ShiftTab         key.Binding
}{
	Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Left:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
	Right: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Back:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),

	New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Rename: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
	Run:    key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "run")),
	Test:   key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "test")),
	Send:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "send")),
	Copy:   key.NewBinding(key.WithKeys("c", "y"), key.WithHelp("c", "copy body")),

	Save:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Add:       key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Toggle:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	Reveal:    key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "reveal/hide")),
	SetActive: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "set active")),
	Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
	ShiftTab:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
}
