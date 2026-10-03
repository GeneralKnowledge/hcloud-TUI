// Package tui implements the Bubble Tea terminal UI for hcloud-tui.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/api"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/cost"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/waste"
)

// App is the root Bubble Tea model.
type App struct {
	client api.Client

	width  int
	height int

	screen     screen
	prevScreen screen

	infra   model.Infrastructure
	summary cost.Summary
	waste   waste.Report

	loading bool
	errMsg  string
	status  string

	cursor int

	selectedServerID int64

	confirm confirmKind
	confirmID int64
	confirmLabel string

	spinner spinner.Model
	ready   bool
}

// New creates the TUI application model.
func New(client api.Client) App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styleAccent
	return App{
		client:  client,
		loading: true,
		spinner: sp,
		screen:  screenOverview,
	}
}

// Init starts the initial data fetch.
func (a App) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, a.fetchCmd())
}

func (a App) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		infra, err := a.client.FetchInfrastructure(ctx)
		return fetchMsg{infra: infra, err: err}
	}
}

func (a App) actionCmd(fn func(context.Context) error, okMsg string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		err := fn(ctx)
		return actionDoneMsg{err: err, message: okMsg}
	}
}

// Update handles Bubble Tea messages.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		return a, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case fetchMsg:
		a.loading = false
		if msg.err != nil {
			a.errMsg = api.FormatAPIError("retrieve Hetzner Cloud data", msg.err)
			return a, nil
		}
		a.errMsg = ""
		a.infra = msg.infra
		a.summary = cost.Calculate(msg.infra)
		a.waste = waste.Detect(msg.infra)
		a.status = "Updated " + msg.infra.FetchedAt.Local().Format("15:04:05")
		a.clampCursor()
		return a, nil

	case actionDoneMsg:
		a.loading = false
		if msg.err != nil {
			a.errMsg = api.FormatAPIError("complete action", msg.err)
			return a, nil
		}
		a.status = msg.message
		a.loading = true
		return a, a.fetchCmd()

	case tea.KeyMsg:
		return a.handleKey(msg)
	}
	return a, nil
}

func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.screen == screenConfirm {
		return a.handleConfirmKey(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		if a.screen == screenOverview || a.screen == screenHelp {
			return a, tea.Quit
		}
		a.screen = screenOverview
		a.errMsg = ""
		return a, nil

	case "esc":
		if a.screen == screenOverview {
			return a, tea.Quit
		}
		a.screen = screenOverview
		a.errMsg = ""
		a.clampCursor()
		return a, nil

	case "?":
		if a.screen != screenHelp {
			a.prevScreen = a.screen
			a.screen = screenHelp
		}
		return a, nil

	case "r", "R":
		if a.screen == screenServerDetail {
			return a.openConfirm(confirmReboot)
		}
		a.loading = true
		a.errMsg = ""
		a.status = "Fetching Hetzner Cloud data..."
		return a, tea.Batch(a.spinner.Tick, a.fetchCmd())

	case "c", "C":
		a.screen = screenCosts
		a.cursor = 0
		return a, nil

	case "u", "U":
		a.screen = screenWaste
		a.cursor = 0
		a.clampCursor()
		return a, nil

	case "up", "k":
		if a.cursor > 0 {
			a.cursor--
		}
		return a, nil

	case "down", "j":
		maxIdx := a.maxCursor()
		if a.cursor < maxIdx {
			a.cursor++
		}
		return a, nil

	case "enter":
		return a.handleEnter()

	case "s", "S":
		if a.screen == screenServerDetail {
			return a.openConfirm(confirmSnapshot)
		}

	case "d", "D":
		if a.screen == screenWaste {
			return a.openWasteDelete()
		}
	}

	return a, nil
}

func (a App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N", "q", "ctrl+c":
		a.screen = a.prevScreen
		return a, nil
	case "y", "Y", "enter":
		return a.runConfirm()
	}
	return a, nil
}

func (a App) openConfirm(kind confirmKind) (tea.Model, tea.Cmd) {
	srv := a.selectedServer()
	if srv == nil {
		return a, nil
	}
	a.prevScreen = a.screen
	a.screen = screenConfirm
	a.confirm = kind
	a.confirmID = srv.ID
	switch kind {
	case confirmReboot:
		a.confirmLabel = fmt.Sprintf("Reboot server %q?", srv.Name)
	case confirmSnapshot:
		a.confirmLabel = fmt.Sprintf("Create a snapshot of %q?", srv.Name)
	}
	return a, nil
}

func (a App) openWasteDelete() (tea.Model, tea.Cmd) {
	if a.cursor < 0 || a.cursor >= len(a.waste.Items) {
		return a, nil
	}
	item := a.waste.Items[a.cursor]
	a.prevScreen = a.screen
	a.screen = screenConfirm
	a.confirmID = item.ID
	switch item.Kind {
	case waste.KindVolume:
		a.confirm = confirmDeleteVolume
		a.confirmLabel = fmt.Sprintf("Delete unattached volume %q?\nThis cannot be undone.", item.ResourceName)
	case waste.KindFloatingIP:
		a.confirm = confirmDeleteFloatingIP
		a.confirmLabel = fmt.Sprintf("Delete unassigned floating IP?\n%s\nThis cannot be undone.", item.Detail)
	}
	return a, nil
}

func (a App) runConfirm() (tea.Model, tea.Cmd) {
	kind := a.confirm
	id := a.confirmID
	a.screen = a.prevScreen
	a.loading = true
	a.status = "Working..."

	switch kind {
	case confirmReboot:
		return a, a.actionCmd(func(ctx context.Context) error {
			return a.client.RebootServer(ctx, id)
		}, "Reboot requested")
	case confirmSnapshot:
		return a, a.actionCmd(func(ctx context.Context) error {
			return a.client.CreateSnapshot(ctx, id, "")
		}, "Snapshot creation requested")
	case confirmDeleteVolume:
		return a, a.actionCmd(func(ctx context.Context) error {
			return a.client.DeleteVolume(ctx, id)
		}, "Volume deleted")
	case confirmDeleteFloatingIP:
		return a, a.actionCmd(func(ctx context.Context) error {
			return a.client.DeleteFloatingIP(ctx, id)
		}, "Floating IP deleted")
	}
	return a, nil
}

func (a App) handleEnter() (tea.Model, tea.Cmd) {
	switch a.screen {
	case screenOverview:
		nServers := len(a.infra.Servers)
		switch {
		case nServers == 0 && a.cursor == 0:
			a.screen = screenVolumes
		case a.cursor < nServers:
			a.selectedServerID = a.infra.Servers[a.cursor].ID
			a.screen = screenServerDetail
		case a.cursor == nServers:
			a.screen = screenVolumes
		case a.cursor == nServers+1:
			a.screen = screenFloatingIPs
		case a.cursor == nServers+2:
			a.screen = screenSnapshots
		}
		a.cursor = 0
	case screenWaste:
		// Stay on waste; details are inline.
	case screenHelp:
		a.screen = a.prevScreen
	}
	return a, nil
}

func (a App) selectedServer() *model.Server {
	for i := range a.infra.Servers {
		if a.infra.Servers[i].ID == a.selectedServerID {
			return &a.infra.Servers[i]
		}
	}
	return nil
}

func (a App) maxCursor() int {
	switch a.screen {
	case screenOverview:
		n := len(a.infra.Servers) + 2 // volumes, floating IPs, snapshots
		if len(a.infra.Servers) == 0 {
			return 2
		}
		return n
	case screenWaste:
		return max(0, len(a.waste.Items)-1)
	case screenVolumes:
		return max(0, len(a.infra.Volumes)-1)
	case screenFloatingIPs:
		return max(0, len(a.infra.FloatingIPs)-1)
	case screenSnapshots:
		return max(0, len(a.infra.Snapshots)-1)
	case screenCosts:
		return max(0, len(a.summary.Items)-1)
	default:
		return 0
	}
}

func (a *App) clampCursor() {
	maxIdx := a.maxCursor()
	if a.cursor > maxIdx {
		a.cursor = maxIdx
	}
	if a.cursor < 0 {
		a.cursor = 0
	}
}

// View renders the current screen.
func (a App) View() string {
	if !a.ready {
		return "Starting hcloud-tui..."
	}

	var body string
	switch a.screen {
	case screenOverview:
		body = a.viewOverview()
	case screenServerDetail:
		body = a.viewServerDetail()
	case screenCosts:
		body = a.viewCosts()
	case screenWaste:
		body = a.viewWaste()
	case screenVolumes:
		body = a.viewVolumes()
	case screenFloatingIPs:
		body = a.viewFloatingIPs()
	case screenSnapshots:
		body = a.viewSnapshots()
	case screenHelp:
		body = a.viewHelp()
	case screenConfirm:
		body = a.viewConfirm()
	default:
		body = a.viewOverview()
	}

	contentWidth := max(20, a.width-4)
	framed := styleBorder.Width(contentWidth).Render(body)

	extra := ""
	if a.loading {
		extra += "\n" + a.spinner.View() + " " + styleMuted.Render("Fetching Hetzner Cloud data...")
	}
	if a.errMsg != "" {
		extra += "\n" + styleErrorBox.Width(contentWidth).Render(a.errMsg)
	}

	return lipgloss.JoinVertical(lipgloss.Left, framed, extra)
}

func (a App) header(title, right string) string {
	w := max(20, a.width-6)
	left := styleTitle.Render(title)
	if right == "" {
		return left
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return trunc(left+" "+right, w)
	}
	return left + strings.Repeat(" ", gap) + styleMuted.Render(right)
}

func (a App) footer(parts ...string) string {
	return helpBar(a.width, parts...)
}
