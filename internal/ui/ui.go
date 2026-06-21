// Package ui holds shared Lip Gloss styles and small rendering helpers used
// across termipost screens.
package ui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	ColorAccent = lipgloss.Color("63")  // violet
	ColorMuted  = lipgloss.Color("242") // grey
	ColorGood   = lipgloss.Color("78")  // green
	ColorBad    = lipgloss.Color("203") // red
	ColorWarn   = lipgloss.Color("214") // orange
	ColorFg     = lipgloss.Color("252")
)

// Shared styles.
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(ColorAccent).
		Padding(0, 1)

	Subtle = lipgloss.NewStyle().Foreground(ColorMuted)

	Label = lipgloss.NewStyle().Foreground(ColorMuted)

	Value = lipgloss.NewStyle().Foreground(ColorFg)

	Selected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(ColorAccent).
			Bold(true)

	FieldFocused = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)

	Good  = lipgloss.NewStyle().Foreground(ColorGood).Bold(true)
	Bad   = lipgloss.NewStyle().Foreground(ColorBad).Bold(true)
	Warn  = lipgloss.NewStyle().Foreground(ColorWarn)
	Error = lipgloss.NewStyle().Foreground(ColorBad).Bold(true)

	HelpBar = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Padding(0, 1)

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)
)
