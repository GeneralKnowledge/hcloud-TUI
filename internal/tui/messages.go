package tui

import (
	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
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

type fetchMsg struct {
	infra model.Infrastructure
	err   error
}

type actionDoneMsg struct {
	err     error
	message string
}

type confirmKind int

const (
	confirmReboot confirmKind = iota
	confirmSnapshot
	confirmDeleteVolume
	confirmDeleteFloatingIP
)
