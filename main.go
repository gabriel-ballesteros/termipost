// Command termipost is a terminal UI for building, sending, and validating HTTP
// requests — a Postman-like client for the terminal.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gbrlballesteros/termipost/internal/model"
	"github.com/gbrlballesteros/termipost/internal/store"
)

func main() {
	dir, err := store.DefaultDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "termipost: cannot resolve data directory:", err)
		os.Exit(1)
	}

	s := store.New(dir)
	if err := s.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "termipost: cannot initialize data directory:", err)
		os.Exit(1)
	}

	data, err := s.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "termipost: cannot load data:", err)
		os.Exit(1)
	}

	app := model.NewApp(s, data)
	m := model.New(app, data.LoadErrors)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "termipost:", err)
		os.Exit(1)
	}
}
