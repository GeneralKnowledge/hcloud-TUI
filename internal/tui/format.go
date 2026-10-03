package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/cost"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

func statusGlyph(status model.ServerStatus) string {
	switch status {
	case model.ServerStatusRunning:
		return "●"
	case model.ServerStatusOff:
		return "○"
	case model.ServerStatusError:
		return "✕"
	default:
		return "?"
	}
}

func statusLabel(s model.Server) string {
	glyph := statusGlyph(s.Status)
	label := strings.ToUpper(string(s.Status))
	if s.Status == model.ServerStatusOther && s.StatusRaw != "" {
		label = strings.ToUpper(s.StatusRaw)
	}
	styled := glyph + " " + label
	switch s.Status {
	case model.ServerStatusRunning:
		return styleOk.Render(styled)
	case model.ServerStatusError:
		return styleError.Render(styled)
	case model.ServerStatusOff:
		return styleMuted.Render(styled)
	default:
		return styleWarn.Render(styled)
	}
}

func trunc(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return trunc(s, width)
	}
	return s + strings.Repeat(" ", width-n)
}

func padLeft(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return trunc(s, width)
	}
	return strings.Repeat(" ", width-n) + s
}

func progressBar(percent float64, width int) string {
	if width < 4 {
		width = 4
	}
	if percent < 0 {
		return styleDim.Render(strings.Repeat("░", width)) + "  n/a"
	}
	if percent > 100 {
		percent = 100
	}
	filled := int((percent / 100) * float64(width))
	if filled > width {
		filled = width
	}
	bar := styleAccent.Render(strings.Repeat("█", filled)) +
		styleDim.Render(strings.Repeat("░", width-filled))
	return fmt.Sprintf("%s  %3.0f%%", bar, percent)
}

func money(amount float64, currency string) string {
	return cost.FormatMoney(amount, currency)
}

func cpuText(percent float64) string {
	if percent < 0 {
		return "n/a CPU"
	}
	return fmt.Sprintf("%.0f%% CPU", percent)
}

func helpBar(width int, parts ...string) string {
	text := strings.Join(parts, "  ")
	return styleFooter.Render(trunc(text, max(0, width-4)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
