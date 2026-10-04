package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hcloud-tui/hcloud-tui/internal/cost"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case inventoryMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.inventory = msg.inv
		m.loaded = true
		m.errMsg = ""
		m.statusMsg = ""
		m = m.clampCursors()
		return m, nil

	case actionDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.statusMsg = msg.message
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, fetchInventoryCmd(m.cloud))

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.loading {
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.screen == screenConfirm {
		return m.handleConfirmKey(key)
	}

	// Global help / quit when not confirming
	switch key {
	case "q":
		if m.screen == screenOverview {
			return m, tea.Quit
		}
		m.screen = screenOverview
		return m, nil
	case "?":
		if m.screen != screenHelp {
			m.prevScreen = m.screen
			m.screen = screenHelp
		}
		return m, nil
	case "esc":
		return m.handleEsc()
	case "R", "r":
		// On server detail, R means reboot (with confirmation). Elsewhere it refreshes.
		if m.screen == screenServerDetail {
			return m.handleServerDetailKey(key)
		}
		if m.screen != screenHelp {
			m.loading = true
			m.statusMsg = "Fetching Hetzner Cloud data..."
			m.errMsg = ""
			return m, tea.Batch(m.spinner.Tick, fetchInventoryCmd(m.cloud))
		}
	}

	switch m.screen {
	case screenOverview:
		return m.handleOverviewKey(key)
	case screenServerDetail:
		return m.handleServerDetailKey(key)
	case screenCosts:
		return m.handleSimpleBackKey(key)
	case screenWaste:
		return m.handleWasteKey(key)
	case screenVolumes, screenFloatingIPs, screenSnapshots:
		return m.handleSimpleBackKey(key)
	case screenHelp:
		if key == "?" || key == "esc" || key == "enter" {
			m.screen = m.prevScreen
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleEsc() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenOverview:
		return m, tea.Quit
	case screenHelp:
		m.screen = m.prevScreen
	default:
		m.screen = screenOverview
	}
	return m, nil
}

func (m Model) handleOverviewKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.focus == focusServers {
			if m.serverCursor > 0 {
				m.serverCursor--
			}
		} else if m.resourceCursor > 0 {
			m.resourceCursor--
		} else {
			m.focus = focusServers
		}
	case "down", "j":
		if m.focus == focusServers {
			if m.serverCursor+1 < len(m.inventory.Servers) {
				m.serverCursor++
			} else {
				m.focus = focusResources
				m.resourceCursor = 0
			}
		} else if m.resourceCursor < 2 {
			m.resourceCursor++
		}
	case "tab":
		if m.focus == focusServers {
			m.focus = focusResources
		} else {
			m.focus = focusServers
		}
	case "enter":
		if m.focus == focusServers && len(m.inventory.Servers) > 0 {
			m.selectedServerID = m.inventory.Servers[m.serverCursor].ID
			m.screen = screenServerDetail
		} else if m.focus == focusResources {
			switch m.resourceCursor {
			case 0:
				m.screen = screenVolumes
			case 1:
				m.screen = screenFloatingIPs
			case 2:
				m.screen = screenSnapshots
			}
		}
	case "C", "c":
		m.screen = screenCosts
	case "U", "u":
		m.screen = screenWaste
		m.wasteCursor = 0
	}
	return m, nil
}

func (m Model) handleServerDetailKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "r", "R":
		srv, ok := m.selectedServer()
		if !ok {
			return m, nil
		}
		m.prevScreen = m.screen
		m.screen = screenConfirm
		m.confirm = confirmReboot
		m.confirmTargetID = srv.ID
		m.confirmLabel = fmt.Sprintf("Reboot server %q?\n\nThis sends a soft reboot via the Hetzner Cloud API.", srv.Name)
	case "s", "S":
		srv, ok := m.selectedServer()
		if !ok {
			return m, nil
		}
		m.prevScreen = m.screen
		m.screen = screenConfirm
		m.confirm = confirmSnapshot
		m.confirmTargetID = srv.ID
		m.confirmLabel = fmt.Sprintf("Create snapshot of %q?\n\nSnapshots incur storage costs on Hetzner Cloud.", srv.Name)
	}
	return m, nil
}

func (m Model) handleWasteKey(key string) (tea.Model, tea.Cmd) {
	items := m.wasteItems()
	switch key {
	case "up", "k":
		if m.wasteCursor > 0 {
			m.wasteCursor--
		}
	case "down", "j":
		if m.wasteCursor+1 < len(items) {
			m.wasteCursor++
		}
	case "enter":
		if len(items) == 0 {
			return m, nil
		}
		item := items[m.wasteCursor]
		switch item.Kind {
		case cost.WasteUnattachedVolume:
			m.screen = screenVolumes
		case cost.WasteUnassignedFloatingIP:
			m.screen = screenFloatingIPs
		}
	}
	return m, nil
}

func (m Model) handleSimpleBackKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		// no-op
	}
	return m, nil
}

func (m Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y", "enter":
		action := m.confirm
		id := m.confirmTargetID
		m.screen = m.prevScreen
		m.confirm = confirmNone
		m.loading = true
		m.statusMsg = "Requesting action..."
		switch action {
		case confirmReboot:
			return m, tea.Batch(m.spinner.Tick, rebootCmd(m.cloud, id))
		case confirmSnapshot:
			name := ""
			if srv, ok := m.selectedServer(); ok {
				name = srv.Name
			}
			return m, tea.Batch(m.spinner.Tick, snapshotCmd(m.cloud, id, name))
		}
	case "n", "N", "esc":
		m.screen = m.prevScreen
		m.confirm = confirmNone
		m.statusMsg = "Cancelled."
	}
	return m, nil
}
