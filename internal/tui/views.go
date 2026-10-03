package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

func (a App) viewOverview() string {
	w := max(40, a.width-6)
	narrow := a.width < 80

	var b strings.Builder
	b.WriteString(a.header("hcloud-tui", a.status))
	b.WriteString("\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", max(10, w))))
	b.WriteString("\n\n")
	b.WriteString(styleSection.Render("SERVERS"))
	b.WriteString("\n\n")

	if len(a.infra.Servers) == 0 && !a.loading {
		b.WriteString(styleMuted.Render("  No servers found in this Hetzner Cloud project."))
		b.WriteString("\n")
	}

	for i, s := range a.infra.Servers {
		line := a.formatServerRow(s, w, narrow)
		if a.cursor == i {
			b.WriteString(styleSelected.Render("> " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleSection.Render("RESOURCES"))
	b.WriteString("\n\n")

	volCost := sumVolumeCost(a.infra.Volumes)
	ipCost := sumIPCost(a.infra.FloatingIPs)
	snapCost := sumSnapCost(a.infra.Snapshots)
	volSize := sumVolumeSize(a.infra.Volumes)

	resourceRows := []struct {
		label string
		value string
		cost  string
	}{
		{"Volumes", fmt.Sprintf("%d       %d GB", len(a.infra.Volumes), volSize), money(volCost, a.summary.Currency)},
		{"Floating IPs", fmt.Sprintf("%d", len(a.infra.FloatingIPs)), money(ipCost, a.summary.Currency)},
		{"Snapshots", fmt.Sprintf("%d", len(a.infra.Snapshots)), money(snapCost, a.summary.Currency)},
	}

	base := len(a.infra.Servers)
	for i, row := range resourceRows {
		idx := base + i
		left := padRight(row.label, 16) + "  " + padRight(row.value, 18)
		line := left + padLeft(row.cost+"/mo", max(8, w-lipgloss.Width(left)-2))
		if narrow {
			line = padRight(row.label, 14) + padLeft(row.cost+"/mo", 12)
		}
		if a.cursor == idx {
			b.WriteString(styleSelected.Render("> " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", max(10, w))))
	b.WriteString("\n")
	totalLabel := "Estimated monthly cost"
	totalValue := money(a.summary.Total, a.summary.Currency)
	gap := max(1, w-lipgloss.Width(totalLabel)-lipgloss.Width(totalValue))
	b.WriteString(styleAccent.Render(totalLabel + strings.Repeat(" ", gap) + totalValue))
	b.WriteString("\n")
	b.WriteString(styleDim.Render("Estimates based on Hetzner list prices — not an invoice."))
	b.WriteString("\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", max(10, w))))
	b.WriteString("\n")
	b.WriteString(a.footer("↑↓ Navigate", "Enter Details", "C Costs", "U Waste", "R Refresh", "? Help", "Q Quit"))
	return b.String()
}

func (a App) formatServerRow(s model.Server, w int, narrow bool) string {
	glyph := statusGlyph(s.Status)
	var styledGlyph string
	switch s.Status {
	case model.ServerStatusRunning:
		styledGlyph = styleOk.Render(glyph)
	case model.ServerStatusError:
		styledGlyph = styleError.Render(glyph)
	case model.ServerStatusOff:
		styledGlyph = styleMuted.Render(glyph)
	default:
		styledGlyph = styleWarn.Render(glyph)
	}

	name := padRight(s.Name, 14)
	stype := padRight(s.ServerType, 8)
	loc := padRight(s.Location, 14)
	cpu := padRight(cpuText(s.CPUPercent), 10)
	price := money(s.MonthlyPrice, s.Currency) + "/mo"

	if narrow || w < 70 {
		return fmt.Sprintf("%s %s  %s  %s", styledGlyph, trunc(s.Name, 16), s.ServerType, price)
	}
	if w < 100 {
		return fmt.Sprintf("%s %s %s %s %s", styledGlyph, name, stype, cpu, padLeft(price, 10))
	}
	return fmt.Sprintf("%s %s %s %s %s %s", styledGlyph, name, stype, loc, cpu, padLeft(price, 10))
}

func (a App) viewServerDetail() string {
	s := a.selectedServer()
	if s == nil {
		return "Server not found.\n\n" + a.footer("Esc Back")
	}
	w := max(40, a.width-6)
	var b strings.Builder
	b.WriteString(a.header(s.Name, ""))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Status", statusLabel(*s)))
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Type", s.ServerType))
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Location", s.Location))
	if s.IPv4 != "" {
		b.WriteString(fmt.Sprintf("%-16s %s\n", "IPv4", s.IPv4))
	}
	if s.IPv6 != "" && a.width >= 60 {
		b.WriteString(fmt.Sprintf("%-16s %s\n", "IPv6", s.IPv6))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%-16s %s\n", "CPU", progressBar(s.CPUPercent, min(20, max(8, w/4)))))
	memLabel := "n/a"
	if s.MemoryGB > 0 {
		memLabel = fmt.Sprintf("%.0f GB allocated", s.MemoryGB)
	}
	diskLabel := "n/a"
	if s.DiskGB > 0 {
		diskLabel = fmt.Sprintf("%d GB allocated", s.DiskGB)
	}
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Memory", styleMuted.Render(memLabel)))
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Disk", styleMuted.Render(diskLabel)))
	b.WriteString("\n")
	label := "Estimated monthly cost"
	val := money(s.MonthlyPrice, s.Currency)
	gap := max(1, w-lipgloss.Width(label)-lipgloss.Width(val))
	b.WriteString(styleAccent.Render(label + strings.Repeat(" ", gap) + val))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", max(10, w))))
	b.WriteString("\n")
	b.WriteString(a.footer("R Reboot", "S Snapshot", "Esc Back"))
	return b.String()
}

func (a App) viewCosts() string {
	w := max(40, a.width-6)
	var b strings.Builder
	b.WriteString(a.header("Monthly Cost", "estimates"))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("These figures are estimates from Hetzner list prices, not your invoice."))
	b.WriteString("\n\n")

	if len(a.summary.Items) == 0 {
		b.WriteString(styleMuted.Render("No billable resources found."))
		b.WriteString("\n")
	}

	maxCost := 0.0
	for _, item := range a.summary.Items {
		if item.MonthlyCost > maxCost {
			maxCost = item.MonthlyCost
		}
	}

	currentCat := ""
	for i, item := range a.summary.Items {
		if item.Category != currentCat {
			if currentCat != "" {
				b.WriteString("\n")
			}
			currentCat = item.Category
			b.WriteString(styleSection.Render(currentCat))
			b.WriteString("\n\n")
		}
		barWidth := min(24, max(6, w/3))
		bar := costBar(item.MonthlyCost, maxCost, barWidth)
		name := padRight(trunc(item.ResourceName, 20), 20)
		price := padLeft(money(item.MonthlyCost, item.Currency), 8)
		line := fmt.Sprintf("%s %s  %s", name, price, bar)
		if a.width < 60 {
			line = fmt.Sprintf("%s %s", trunc(item.ResourceName, 16), price)
		}
		if a.cursor == i {
			b.WriteString(styleSelected.Render("> " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleDim.Render(strings.Repeat("─", max(10, w))))
	b.WriteString("\n")
	total := "ESTIMATED TOTAL"
	val := money(a.summary.Total, a.summary.Currency)
	gap := max(1, w-lipgloss.Width(total)-lipgloss.Width(val))
	b.WriteString(styleAccent.Render(total + strings.Repeat(" ", gap) + val))
	b.WriteString("\n\n")
	b.WriteString(a.footer("Esc Back", "R Refresh", "Q Quit"))
	return b.String()
}

func (a App) viewWaste() string {
	w := max(40, a.width-6)
	var b strings.Builder
	b.WriteString(a.header("Potential Waste", "unattached resources"))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("Only resources with no associated server are listed."))
	b.WriteString("\n")
	b.WriteString(styleDim.Render("Low utilisation alone is never treated as waste."))
	b.WriteString("\n\n")

	if len(a.waste.Items) == 0 {
		b.WriteString(styleOk.Render("No unattached volumes or unassigned floating IPs found."))
		b.WriteString("\n\n")
		b.WriteString(a.footer("Esc Back"))
		return b.String()
	}

	for i, item := range a.waste.Items {
		title := styleWarn.Render("⚠ "+item.Title)
		detail := padRight(trunc(item.Detail, max(20, w-18)), max(20, w-18))
		price := money(item.MonthlyCost, item.Currency) + "/month"
		block := title + "\n  " + detail + "  " + price
		if a.width < 70 {
			block = title + "\n  " + trunc(item.Detail, w-4) + "\n  " + price
		}
		if a.cursor == i {
			b.WriteString(styleSelected.Render(block))
		} else {
			b.WriteString(block)
		}
		b.WriteString("\n\n")
	}

	label := "Potential monthly saving"
	val := money(a.waste.PotentialSaving, a.waste.Currency)
	gap := max(1, w-lipgloss.Width(label)-lipgloss.Width(val))
	b.WriteString(styleAccent.Render(label + strings.Repeat(" ", gap) + val))
	b.WriteString("\n\n")
	b.WriteString(a.footer("D Delete", "Esc Back", "R Refresh"))
	return b.String()
}

func (a App) viewVolumes() string {
	w := max(40, a.width-6)
	var b strings.Builder
	b.WriteString(a.header("Volumes", ""))
	b.WriteString("\n\n")
	if len(a.infra.Volumes) == 0 {
		b.WriteString(styleMuted.Render("No volumes found."))
		b.WriteString("\n\n")
		b.WriteString(a.footer("Esc Back"))
		return b.String()
	}
	for i, v := range a.infra.Volumes {
		attached := "unattached"
		if v.Attached {
			attached = "→ " + v.ServerName
		}
		line := fmt.Sprintf("%s  %d GB  %s  %s  %s/mo",
			padRight(trunc(v.Name, 16), 16),
			v.SizeGB,
			padRight(v.Location, 8),
			padRight(attached, 18),
			money(v.MonthlyPrice, v.Currency),
		)
		if a.width < 70 {
			line = fmt.Sprintf("%s  %dGB  %s  %s/mo", trunc(v.Name, 14), v.SizeGB, attached, money(v.MonthlyPrice, v.Currency))
		}
		if a.cursor == i {
			b.WriteString(styleSelected.Render("> " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	_ = w
	b.WriteString("\n")
	b.WriteString(a.footer("Esc Back", "R Refresh"))
	return b.String()
}

func (a App) viewFloatingIPs() string {
	var b strings.Builder
	b.WriteString(a.header("Floating IPs", ""))
	b.WriteString("\n\n")
	if len(a.infra.FloatingIPs) == 0 {
		b.WriteString(styleMuted.Render("No floating IPs found."))
		b.WriteString("\n\n")
		b.WriteString(a.footer("Esc Back"))
		return b.String()
	}
	for i, ip := range a.infra.FloatingIPs {
		assigned := "not assigned"
		if ip.Assigned {
			assigned = "→ " + ip.ServerName
		}
		name := ip.Name
		if name == "" {
			name = ip.IP
		}
		line := fmt.Sprintf("%s  %s  %s  %s  %s  %s/mo",
			padRight(trunc(name, 12), 12),
			padRight(ip.IP, 16),
			padRight(ip.Type, 6),
			padRight(ip.Location, 8),
			padRight(assigned, 16),
			money(ip.MonthlyPrice, ip.Currency),
		)
		if a.width < 80 {
			line = fmt.Sprintf("%s  %s  %s  %s/mo", trunc(ip.IP, 16), ip.Type, assigned, money(ip.MonthlyPrice, ip.Currency))
		}
		if a.cursor == i {
			b.WriteString(styleSelected.Render("> " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(a.footer("Esc Back", "R Refresh"))
	return b.String()
}

func (a App) viewSnapshots() string {
	var b strings.Builder
	b.WriteString(a.header("Snapshots", ""))
	b.WriteString("\n\n")
	if len(a.infra.Snapshots) == 0 {
		b.WriteString(styleMuted.Render("No snapshots found."))
		b.WriteString("\n\n")
		b.WriteString(a.footer("Esc Back"))
		return b.String()
	}
	for i, s := range a.infra.Snapshots {
		created := s.Created.Local().Format("2006-01-02 15:04")
		line := fmt.Sprintf("%s  %s  %s  %.1f GB  %s/mo",
			padRight(trunc(s.Name, 18), 18),
			padRight(trunc(s.ServerName, 12), 12),
			created,
			s.SizeGB,
			money(s.MonthlyPrice, s.Currency),
		)
		if a.width < 80 {
			line = fmt.Sprintf("%s  %s  %.1fGB", trunc(s.Name, 18), created, s.SizeGB)
		}
		if a.cursor == i {
			b.WriteString(styleSelected.Render("> " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(a.footer("Esc Back", "R Refresh"))
	return b.String()
}

func (a App) viewHelp() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")
	b.WriteString(styleSection.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(helpLine("↑ / k", "Up"))
	b.WriteString(helpLine("↓ / j", "Down"))
	b.WriteString(helpLine("Enter", "Select"))
	b.WriteString(helpLine("Esc", "Back"))
	b.WriteString("\n")
	b.WriteString(styleSection.Render("Actions"))
	b.WriteString("\n")
	b.WriteString(helpLine("R", "Refresh (Reboot on server detail)"))
	b.WriteString(helpLine("C", "Cost overview"))
	b.WriteString(helpLine("U", "Potential waste"))
	b.WriteString(helpLine("S", "Snapshot (server detail)"))
	b.WriteString(helpLine("D", "Delete (waste view, with confirm)"))
	b.WriteString("\n")
	b.WriteString(styleSection.Render("General"))
	b.WriteString("\n")
	b.WriteString(helpLine("?", "Help"))
	b.WriteString(helpLine("Q", "Quit"))
	b.WriteString("\n")
	b.WriteString(a.footer("Esc Back"))
	return b.String()
}

func (a App) viewConfirm() string {
	var b strings.Builder
	b.WriteString(styleWarn.Render("Confirm"))
	b.WriteString("\n\n")
	b.WriteString(a.confirmLabel)
	b.WriteString("\n\n")
	b.WriteString(a.footer("Y Confirm", "N Cancel", "Esc Cancel"))
	return b.String()
}

func helpLine(key, desc string) string {
	return "  " + styleHelpKey.Render(padRight(key, 12)) + " " + desc + "\n"
}

func costBar(value, maxValue float64, width int) string {
	if maxValue <= 0 || width <= 0 {
		return ""
	}
	filled := int((value / maxValue) * float64(width))
	if filled < 1 && value > 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	return styleAccent.Render(strings.Repeat("█", filled)) + styleDim.Render(strings.Repeat(" ", width-filled))
}

func sumVolumeCost(vols []model.Volume) float64 {
	t := 0.0
	for _, v := range vols {
		t += v.MonthlyPrice
	}
	return t
}

func sumIPCost(ips []model.FloatingIP) float64 {
	t := 0.0
	for _, ip := range ips {
		t += ip.MonthlyPrice
	}
	return t
}

func sumSnapCost(snaps []model.Snapshot) float64 {
	t := 0.0
	for _, s := range snaps {
		t += s.MonthlyPrice
	}
	return t
}

func sumVolumeSize(vols []model.Volume) int {
	t := 0
	for _, v := range vols {
		t += v.SizeGB
	}
	return t
}
