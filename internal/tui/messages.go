package tui

import (
	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

type screen int

const (
	screenOverview screen = iota
	screenServerDetail
	screenCosts
	screenWaste
	screenVolumes
	screenFloatingIPs
	screenSnapshots
	screenHelp
	screenConfirm
)

type overviewFocus int

const (
	focusServers overviewFocus = iota
	focusResources
)

type resourceKind int

const (
	resourceVolumes resourceKind = iota
	resourceFloatingIPs
	resourceSnapshots
)

type inventoryMsg struct {
	inv model.Inventory
	err error
}

type actionDoneMsg struct {
	message string
	err     error
}

type confirmAction int

const (
	confirmNone confirmAction = iota
	confirmReboot
	confirmSnapshot
)
