package tui

import (
	"fmt"
	"strings"

	"github.com/hcloud-tui/hcloud-tui/internal/cost"
	"github.com/hcloud-tui/hcloud-tui/internal/model"
	"github.com/hcloud-tui/hcloud-tui/internal/tui/styles"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading hcloud-tui..."
	}

	switch m.screen {
	case screenServerDetail:
		return m.viewServerDetail()
	case screenCosts:
		return m.viewCosts()
	case screenWaste:
		return m.viewWaste()
	case screenVolumes:
		return m.viewVolumes()
	case screenFloatingIPs:
		return m.viewFloatingIPs()
	case screenSnapshots:
		return m.viewSnapshots()
	case screenHelp:
		return m.viewHelp()
	case screenConfirm:
		return m.viewConfirm()
	default:
		return m.viewOverview()
	}
}

func (m Model) viewOverview() string {
	w := m.width
	narrow := w < 80
	var b strings.Builder

	titleLeft := styles.Title.Render("hcloud-tui")
	titleRight := styles.Dim.Render(ago(m.inventory.FetchedAt))
	b.WriteString(m.headerLine("Hetzner Cloud", titleLeft, titleRight))
	b.WriteString("\n\n")

	if m.loading && !m.loaded {
		b.WriteString(m.spinner.View() + " " + styles.Dim.Render("Fetching Hetzner Cloud data..."))
		b.WriteString("\n")
		return frame("Hetzner Cloud", b.String(), helpFooter("Q Quit", "? Help"), m.width, m.height)
	}

	if m.errMsg != "" && !m.loaded {
		b.WriteString(styles.Error.Render(m.errMsg))
		b.WriteString("\n")
		return frame("Hetzner Cloud", b.String(), helpFooter("R Refresh", "Q Quit", "? Help"), m.width, m.height)
	}

	b.WriteString(sectionHeader("Servers"))
	b.WriteString("\n")
	if len(m.inventory.Servers) == 0 {
		b.WriteString(styles.Dim.Render("  No servers in this project."))
		b.WriteString("\n")
	} else {
		for i, s := range m.inventory.Servers {
			selected := m.focus == focusServers && i == m.serverCursor
			line := m.formatServerRow(s, narrow)
			b.WriteString(styleSelected(selected, "  "+line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(sectionHeader("Resources"))
	b.WriteString("\n")

	volCost := sumVolumeCost(m.inventory.Volumes)
	ipCost := sumFloatingIPCost(m.inventory.FloatingIPs)
	snapCost := sumSnapshotCost(m.inventory.Snapshots)
	volSize := sumVolumeSize(m.inventory.Volumes)

	resources := []struct {
		label string
		value string
		cost  float64
	}{
		{"Volumes", fmt.Sprintf("%d       %d GB", len(m.inventory.Volumes), volSize), volCost},
		{"Floating IPs", fmt.Sprintf("%d", len(m.inventory.FloatingIPs)), ipCost},
		{"Snapshots", fmt.Sprintf("%d", len(m.inventory.Snapshots)), snapCost},
	}
	for i, r := range resources {
		selected := m.focus == focusResources && i == m.resourceCursor
		costStr := formatMoney(r.cost, m.inventory.Pricing.Currency)
		var line string
		if narrow {
			line = fmt.Sprintf("%-14s %-10s %10s", r.label, r.value, costStr)
		} else {
			line = fmt.Sprintf("%-16s %-20s %12s", r.label, r.value, costStr)
		}
		b.WriteString(styleSelected(selected, "  "+line))
		b.WriteString("\n")
	}

	total := cost.Calculate(m.inventory).Total
	b.WriteString("\n")
	b.WriteString(styles.Dim.Render(strings.Repeat("─", max(20, min(m.width-8, 60)))))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s %s",
		padRight("Estimated monthly cost", max(24, min(m.width-20, 40))),
		styles.Cost.Render(formatMoney(total, m.inventory.Pricing.Currency)),
	))
	b.WriteString("\n")
	b.WriteString(styles.Dim.Render("Estimates are based on Hetzner list prices, not invoices."))
	b.WriteString("\n")

	if m.loading {
		b.WriteString("\n")
		b.WriteString(m.spinner.View() + " " + styles.Dim.Render("Fetching Hetzner Cloud data..."))
		b.WriteString("\n")
	}
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.Error.Render(m.errMsg))
		b.WriteString("\n")
	}
	if m.statusMsg != "" && !m.loading {
		b.WriteString("\n")
		b.WriteString(styles.Dim.Render(m.statusMsg))
		b.WriteString("\n")
	}

	footer := helpFooter("↑↓ Navigate", "Enter Details", "C Costs", "U Unused", "R Refresh", "? Help", "Q Quit")
	return frame("Hetzner Cloud", b.String(), footer, m.width, m.height)
}

func (m Model) formatServerRow(s model.Server, narrow bool) string {
	sym := model.StatusSymbol(s.Status)
	name := truncate(s.Name, 14)
	stype := truncate(s.ServerType, 8)
	loc := truncate(s.Location, 12)
	cpu := formatCPU(s.CPUPercent)
	price := formatMoney(s.MonthlyCost, s.Currency)

	if narrow {
		return fmt.Sprintf("%s %-12s %-6s %8s", sym, name, stype, price)
	}
	return fmt.Sprintf("%s %-14s %-8s %-12s %-10s %10s", sym, name, stype, loc, cpu, price)
}

func (m Model) viewServerDetail() string {
	srv, ok := m.selectedServer()
	if !ok {
		return frame("Server", styles.Dim.Render("Server not found."), "Esc Back", m.width, m.height)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Status", statusLine(srv.Status)))
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Type", srv.ServerType))
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Location", srv.Location))
	if srv.IPv4 != "" {
		b.WriteString(fmt.Sprintf("%-16s %s\n", "IPv4", srv.IPv4))
	}
	if srv.IPv6 != "" && m.width >= 70 {
		b.WriteString(fmt.Sprintf("%-16s %s\n", "IPv6", srv.IPv6))
	}
	b.WriteString("\n")

	barW := maxBarWidth(m.width)
	if srv.CPUPercent != nil {
		b.WriteString(fmt.Sprintf("%-16s %s  %3.0f%%\n", "CPU", progressBar(*srv.CPUPercent, barW), *srv.CPUPercent))
	} else {
		b.WriteString(fmt.Sprintf("%-16s %s\n", "CPU", styles.Dim.Render("metrics unavailable")))
	}
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Memory", fmt.Sprintf("%.0f GB", srv.MemoryGB)))
	b.WriteString(fmt.Sprintf("%-16s %s\n", "Disk", fmt.Sprintf("%d GB", srv.DiskGB)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s %s\n",
		padRight("Estimated monthly cost", 36),
		styles.Cost.Render(formatMoney(srv.MonthlyCost, srv.Currency)),
	))

	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.Dim.Render(m.statusMsg))
		b.WriteString("\n")
	}
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.Error.Render(m.errMsg))
		b.WriteString("\n")
	}

	footer := helpFooter("R Reboot", "S Snapshot", "Esc Back")
	return frame(srv.Name, b.String(), footer, m.width, m.height)
}

func (m Model) viewCosts() string {
	bdown := m.costBreakdown()
	var b strings.Builder
	b.WriteString(styles.Dim.Render("Estimated monthly costs (not an invoice)"))
	b.WriteString("\n")

	writeCat := func(title, category string) {
		items := filterCategory(bdown.Items, category)
		if len(items) == 0 {
			return
		}
		b.WriteString("\n")
		b.WriteString(sectionHeader(title))
		b.WriteString("\n")
		maxCost := 0.0
		for _, it := range items {
			if it.MonthlyCost > maxCost {
				maxCost = it.MonthlyCost
			}
		}
		barW := maxBarWidth(m.width)
		for _, it := range items {
			name := truncate(it.ResourceName, 18)
			price := formatMoney(it.MonthlyCost, it.Currency)
			ratio := 0.0
			if maxCost > 0 {
				ratio = it.MonthlyCost / maxCost
			}
			filled := int(ratio * float64(barW))
			if it.MonthlyCost > 0 && filled == 0 {
				filled = 1
			}
			bar := styles.BarFill.Render(strings.Repeat("█", filled)) +
				styles.BarEmpty.Render(strings.Repeat("░", barW-filled))
			b.WriteString(fmt.Sprintf("  %-18s %10s   %s\n", name, price, bar))
		}
	}

	writeCat("Servers", "servers")
	writeCat("Storage", "storage")
	writeCat("Network", "network")
	writeCat("Images", "images")

	b.WriteString("\n")
	b.WriteString(styles.Dim.Render(strings.Repeat("─", max(20, min(m.width-8, 60)))))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%-22s %s\n", "ESTIMATED TOTAL", styles.Cost.Render(formatMoney(bdown.Total, bdown.Currency))))

	return frame("Monthly Cost", b.String(), helpFooter("Esc Back", "R Refresh", "? Help"), m.width, m.height)
}

func (m Model) viewWaste() string {
	items := m.wasteItems()
	var b strings.Builder
	b.WriteString(styles.Dim.Render("Unattached / unassigned resources only."))
	b.WriteString("\n")
	b.WriteString(styles.Dim.Render("Low CPU usage is never treated as waste."))
	b.WriteString("\n")

	if len(items) == 0 {
		b.WriteString("\n")
		b.WriteString(styles.Dim.Render("No unattached volumes or unassigned floating IPs found."))
		b.WriteString("\n")
	} else {
		for i, item := range items {
			selected := i == m.wasteCursor
			block := fmt.Sprintf("⚠ %s\n  %s%*s",
				item.ResourceType,
				item.Detail,
				max(1, min(m.width-8, 50)-len(item.Detail)),
				formatMoney(item.MonthlyCost, item.Currency)+"/month",
			)
			// Simpler two-line layout:
			line1 := fmt.Sprintf("⚠ %s", item.ResourceType)
			line2 := fmt.Sprintf("  %s", truncate(item.Detail, max(20, m.width-24)))
			line3 := fmt.Sprintf("  %s", formatMoney(item.MonthlyCost, item.Currency)+"/month")
			if selected {
				b.WriteString(styles.Selected.Render(line1))
				b.WriteString("\n")
				b.WriteString(styles.Selected.Render(line2))
				b.WriteString("\n")
				b.WriteString(styles.Selected.Render(line3))
			} else {
				b.WriteString(styles.StatusWarn.Render(line1))
				b.WriteString("\n")
				b.WriteString(styles.Normal.Render(line2))
				b.WriteString("\n")
				b.WriteString(styles.Dim.Render(line3))
			}
			b.WriteString("\n\n")
			_ = block
		}
		total := cost.WasteTotal(items)
		b.WriteString(fmt.Sprintf("%s %s\n",
			padRight("Potential monthly saving", 36),
			styles.Cost.Render(formatMoney(total, m.inventory.Pricing.Currency)),
		))
	}

	return frame("Potential Waste", b.String(), helpFooter("Enter Details", "Esc Back"), m.width, m.height)
}

func (m Model) viewVolumes() string {
	var b strings.Builder
	if len(m.inventory.Volumes) == 0 {
		b.WriteString(styles.Dim.Render("No volumes."))
	} else {
		for _, v := range m.inventory.Volumes {
			attached := "unattached"
			if v.ServerName != "" {
				attached = "attached to " + v.ServerName
			} else if v.ServerID != 0 {
				attached = fmt.Sprintf("attached to server %d", v.ServerID)
			}
			b.WriteString(fmt.Sprintf("%-16s %4d GB  %-12s  %-28s  %8s\n",
				truncate(v.Name, 16),
				v.SizeGB,
				truncate(v.Location, 12),
				truncate(attached, 28),
				formatMoney(v.MonthlyCost, v.Currency),
			))
		}
	}
	return frame("Volumes", b.String(), helpFooter("Esc Back"), m.width, m.height)
}

func (m Model) viewFloatingIPs() string {
	var b strings.Builder
	if len(m.inventory.FloatingIPs) == 0 {
		b.WriteString(styles.Dim.Render("No floating IPs."))
	} else {
		for _, ip := range m.inventory.FloatingIPs {
			assigned := "not assigned"
			if ip.ServerName != "" {
				assigned = "assigned to " + ip.ServerName
			} else if ip.ServerID != 0 {
				assigned = fmt.Sprintf("assigned to server %d", ip.ServerID)
			}
			name := ip.Name
			if name == "" {
				name = ip.IP
			}
			b.WriteString(fmt.Sprintf("%-16s %-16s %-6s %-12s  %-24s  %8s\n",
				truncate(name, 16),
				truncate(ip.IP, 16),
				ip.Type,
				truncate(ip.Location, 12),
				truncate(assigned, 24),
				formatMoney(ip.MonthlyCost, ip.Currency),
			))
		}
	}
	return frame("Floating IPs", b.String(), helpFooter("Esc Back"), m.width, m.height)
}

func (m Model) viewSnapshots() string {
	var b strings.Builder
	if len(m.inventory.Snapshots) == 0 {
		b.WriteString(styles.Dim.Render("No snapshots."))
	} else {
		for _, s := range m.inventory.Snapshots {
			name := s.Name
			if name == "" {
				name = s.Description
			}
			server := s.ServerName
			if server == "" {
				server = "-"
			}
			b.WriteString(fmt.Sprintf("%-20s %-14s %-12s %6.1f GB  %8s\n",
				truncate(name, 20),
				truncate(server, 14),
				s.Created.Format("2006-01-02"),
				s.SizeGB,
				formatMoney(s.MonthlyCost, s.Currency),
			))
		}
	}
	return frame("Snapshots", b.String(), helpFooter("Esc Back"), m.width, m.height)
}

func (m Model) viewHelp() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")
	b.WriteString(sectionHeader("Navigation"))
	b.WriteString("\n")
	b.WriteString(renderKeyHelp("↑ / k", "Up") + "\n")
	b.WriteString(renderKeyHelp("↓ / j", "Down") + "\n")
	b.WriteString(renderKeyHelp("Enter", "Select") + "\n")
	b.WriteString(renderKeyHelp("Esc", "Back") + "\n")
	b.WriteString("\n")
	b.WriteString(sectionHeader("Actions"))
	b.WriteString("\n")
	b.WriteString(renderKeyHelp("R", "Refresh (Reboot on server detail)") + "\n")
	b.WriteString(renderKeyHelp("C", "Cost overview") + "\n")
	b.WriteString(renderKeyHelp("U", "Potential waste") + "\n")
	b.WriteString(renderKeyHelp("S", "Snapshot (server detail)") + "\n")
	b.WriteString("\n")
	b.WriteString(sectionHeader("General"))
	b.WriteString("\n")
	b.WriteString(renderKeyHelp("?", "Help") + "\n")
	b.WriteString(renderKeyHelp("Q", "Quit") + "\n")
	return frame("Help", b.String(), helpFooter("Esc Back"), m.width, m.height)
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.confirmLabel)
	b.WriteString("\n\n")
	b.WriteString(styles.Dim.Render("Y Confirm    N Cancel"))
	b.WriteString("\n")
	return frame("Confirm", b.String(), "", m.width, m.height)
}

func (m Model) headerLine(borderTitle, left, right string) string {
	_ = borderTitle
	gap := m.width - 8 - lipglossWidth(left) - lipglossWidth(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func lipglossWidth(s string) int {
	// Approximate visible width by stripping ANSI; lipgloss.Width is ideal but
	// plain rune count is fine for our styled short strings.
	plain := stripANSI(s)
	return len([]rune(plain))
}

func stripANSI(s string) string {
	var b strings.Builder
	inESC := false
	for _, r := range s {
		if r == 0x1b {
			inESC = true
			continue
		}
		if inESC {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inESC = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func filterCategory(items []cost.CostItem, category string) []cost.CostItem {
	out := make([]cost.CostItem, 0)
	for _, it := range items {
		if it.Category == category {
			out = append(out, it)
		}
	}
	return out
}

func sumVolumeCost(vs []model.Volume) float64 {
	var t float64
	for _, v := range vs {
		t += v.MonthlyCost
	}
	return t
}

func sumFloatingIPCost(ips []model.FloatingIP) float64 {
	var t float64
	for _, ip := range ips {
		t += ip.MonthlyCost
	}
	return t
}

func sumSnapshotCost(ss []model.Snapshot) float64 {
	var t float64
	for _, s := range ss {
		t += s.MonthlyCost
	}
	return t
}

func sumVolumeSize(vs []model.Volume) int {
	var t int
	for _, v := range vs {
		t += v.SizeGB
	}
	return t
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
