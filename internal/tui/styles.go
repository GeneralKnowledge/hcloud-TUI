package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorTitle   = lipgloss.Color("14")
	colorMuted   = lipgloss.Color("245")
	colorOk      = lipgloss.Color("10")
	colorWarn    = lipgloss.Color("11")
	colorError   = lipgloss.Color("9")
	colorAccent  = lipgloss.Color("12")
	colorBorder  = lipgloss.Color("240")
	colorSelect  = lipgloss.Color("15")
	colorDim     = lipgloss.Color("244")
	colorBarFill = lipgloss.Color("12")
	colorBarEmpty = lipgloss.Color("238")
)

var (
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(colorTitle)
	styleMuted = lipgloss.NewStyle().Foreground(colorMuted)
	styleDim   = lipgloss.NewStyle().Foreground(colorDim)
	styleOk    = lipgloss.NewStyle().Foreground(colorOk)
	styleWarn  = lipgloss.NewStyle().Foreground(colorWarn)
	styleError = lipgloss.NewStyle().Foreground(colorError)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleSection = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	styleSelected = lipgloss.NewStyle().Foreground(colorSelect).Bold(true)
	styleHelpKey = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)
	styleFooter = lipgloss.NewStyle().Foreground(colorMuted)
	styleErrorBox = lipgloss.NewStyle().
			Foreground(colorError).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorError).
			Padding(0, 1)
)
