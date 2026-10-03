// Package styles defines lipgloss styles for hcloud-tui.
package styles

import "github.com/charmbracelet/lipgloss"

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))

	Subtitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	Section = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		MarginTop(1).
		MarginBottom(0)

	Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	StatusRunning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("10"))

	StatusOff = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	StatusWarn = lipgloss.NewStyle().
		Foreground(lipgloss.Color("11"))

	StatusError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9"))

	Cost = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	HelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9"))

	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	Footer = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		MarginTop(1)

	BarFill = lipgloss.NewStyle().
		Foreground(lipgloss.Color("12"))

	BarEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)
