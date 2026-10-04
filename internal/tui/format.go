package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/hcloud-tui/hcloud-tui/internal/cost"
	"github.com/hcloud-tui/hcloud-tui/internal/model"
	"github.com/hcloud-tui/hcloud-tui/internal/tui/styles"
)

func formatMoney(amount float64, currency string) string {
	return cost.FormatMoney(amount, currency)
}

func formatCPU(cpu *float64) string {
	if cpu == nil {
		return "  n/a CPU"
	}
	return fmt.Sprintf("%3.0f%% CPU", *cpu)
}

func statusLine(status model.ServerStatus) string {
	sym := model.StatusSymbol(status)
	label := model.StatusLabel(status)
	text := sym + " " + label
	switch status {
	case model.ServerStatusRunning:
		return styles.StatusRunning.Render(text)
	case model.ServerStatusOff, model.ServerStatusStopping:
		return styles.StatusOff.Render(text)
	case model.ServerStatusUnknown:
		return styles.StatusWarn.Render(text)
	default:
		return styles.StatusWarn.Render(text)
	}
}

func progressBar(percent float64, width int) string {
	if width < 4 {
		width = 4
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int((percent / 100) * float64(width))
	if percent > 0 && filled == 0 {
		filled = 1
	}
	empty := width - filled
	return styles.BarFill.Render(strings.Repeat("█", filled)) +
		styles.BarEmpty.Render(strings.Repeat("░", empty))
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-len(r))
}

func padLeft(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return strings.Repeat(" ", width-len(r)) + s
}

func formatUpdated(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "Updated " + t.Format("15:04:05")
}

func frame(title, body, footer string, width, height int) string {
	if width < 40 {
		width = 40
	}
	innerWidth := width - 4
	header := styles.Title.Render(truncate(title, innerWidth))
	content := header + "\n" + body
	if footer != "" {
		content += "\n" + styles.Footer.Render(footer)
	}
	boxed := styles.Border.Width(width - 2).Render(content)
	if height > 0 {
		lines := strings.Split(boxed, "\n")
		if len(lines) > height {
			boxed = strings.Join(lines[:height], "\n")
		}
	}
	return boxed
}

func helpFooter(parts ...string) string {
	return strings.Join(parts, "  ")
}

func renderKeyHelp(key, desc string) string {
	return styles.HelpKey.Render(padRight(key, 12)) + styles.HelpDesc.Render(desc)
}

func maxBarWidth(totalWidth int) int {
	w := totalWidth / 3
	if w < 8 {
		w = 8
	}
	if w > 24 {
		w = 24
	}
	return w
}

func styleSelected(selected bool, line string) string {
	if selected {
		return styles.Selected.Render(line)
	}
	return styles.Normal.Render(line)
}

func sectionHeader(title string) string {
	return styles.Section.Render(strings.ToUpper(title))
}

// Ensure lipgloss is referenced when styles package is used alone in tests.
var _ = lipgloss.NewStyle
