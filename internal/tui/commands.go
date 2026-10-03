package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hcloud-tui/hcloud-tui/internal/api"
)

func fetchInventoryCmd(cloud api.Cloud) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		inv, err := cloud.FetchInventory(ctx)
		return inventoryMsg{inv: inv, err: err}
	}
}

func rebootCmd(cloud api.Cloud, serverID int64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := cloud.RebootServer(ctx, serverID)
		if err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{message: "Reboot requested."}
	}
}

func snapshotCmd(cloud api.Cloud, serverID int64, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		desc := fmt.Sprintf("hcloud-tui-%s-%s", name, time.Now().Format("20060102-150405"))
		err := cloud.CreateSnapshot(ctx, serverID, desc)
		if err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{message: "Snapshot requested: " + desc}
	}
}
