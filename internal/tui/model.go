package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hcloud-tui/hcloud-tui/internal/api"
	"github.com/hcloud-tui/hcloud-tui/internal/cost"
	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

// Model is the root Bubble Tea model for hcloud-tui.
type Model struct {
	cloud api.Cloud

	width  int
	height int

	screen     screen
	prevScreen screen

	inventory model.Inventory
	loaded    bool
	loading   bool
	errMsg    string
	statusMsg string
	spinner   spinner.Model

	focus          overviewFocus
	serverCursor   int
	resourceCursor int
	wasteCursor    int
	costCursor     int

	selectedServerID int64

	confirm confirmAction
	confirmTargetID int64
	confirmLabel    string
}

// New creates the initial TUI model and kicks off the first inventory fetch.
func New(cloud api.Cloud) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return Model{
		cloud:   cloud,
		screen:  screenOverview,
		loading: true,
		focus:   focusServers,
		spinner: s,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchInventoryCmd(m.cloud))
}

func (m Model) selectedServer() (model.Server, bool) {
	for _, s := range m.inventory.Servers {
		if s.ID == m.selectedServerID {
			return s, true
		}
	}
	if m.serverCursor >= 0 && m.serverCursor < len(m.inventory.Servers) {
		return m.inventory.Servers[m.serverCursor], true
	}
	return model.Server{}, false
}

func (m Model) costBreakdown() cost.Breakdown {
	return cost.Calculate(m.inventory)
}

func (m Model) wasteItems() []cost.WasteItem {
	return cost.FindWaste(m.inventory)
}

func (m Model) clampCursors() Model {
	if m.serverCursor < 0 {
		m.serverCursor = 0
	}
	if n := len(m.inventory.Servers); n > 0 && m.serverCursor >= n {
		m.serverCursor = n - 1
	}
	if m.resourceCursor < 0 {
		m.resourceCursor = 0
	}
	if m.resourceCursor > 2 {
		m.resourceCursor = 2
	}
	waste := m.wasteItems()
	if m.wasteCursor < 0 {
		m.wasteCursor = 0
	}
	if len(waste) > 0 && m.wasteCursor >= len(waste) {
		m.wasteCursor = len(waste) - 1
	}
	return m
}

func (m Model) setStatus(msg string) Model {
	m.statusMsg = msg
	return m
}

func (m Model) clearTransient() Model {
	m.errMsg = ""
	m.statusMsg = ""
	return m
}

func ago(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return formatUpdated(t)
}
