// Command termipost is a terminal UI for building, sending, and validating HTTP
// requests — a Postman-like client for the terminal.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/model"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

// version is set at build time via -ldflags (e.g. by GoReleaser); it defaults to
// "dev" for local builds.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version", "version":
			fmt.Println("termipost", version)
			return
		}
	}

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
