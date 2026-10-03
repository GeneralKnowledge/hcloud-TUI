// Command hcloud-tui is a lightweight terminal dashboard for Hetzner Cloud.
package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hcloud-tui/hcloud-tui/internal/api"
	"github.com/hcloud-tui/hcloud-tui/internal/config"
	"github.com/hcloud-tui/hcloud-tui/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	demo := false
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help":
			fmt.Print(usage())
			return nil
		case "--demo":
			demo = true
		case "-v", "--version", "version":
			fmt.Println("hcloud-tui 0.1.0")
			return nil
		default:
			return fmt.Errorf("unknown argument: %s\n\n%s", arg, usage())
		}
	}

	var cloud api.Cloud
	if demo || os.Getenv("HCLOUD_TUI_DEMO") == "1" {
		cloud = &api.Mock{Inventory: api.SampleInventory()}
	} else {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrTokenNotFound) {
				return errors.New(config.TokenMissingMessage())
			}
			return err
		}
		// Never print the token.
		_ = cfg.TokenSource
		cloud = api.New(cfg.Token)
	}

	m := tui.New(cloud)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func usage() string {
	return `hcloud-tui — Hetzner Cloud TUI

A lightweight terminal dashboard for monitoring and managing
Hetzner Cloud infrastructure and estimating monthly costs.

Usage:
  hcloud-tui
  hcloud-tui --demo

Authentication:
  export HCLOUD_TOKEN="..."

  Or place the token in ~/.config/hcloud-tui/config:
    token = "..."

Environment variables take precedence over the config file.

Keys:
  ↑/k ↓/j   Navigate
  Enter     Open details
  C         Cost overview
  U         Potential waste
  R         Refresh
  ?         Help
  Q         Quit
`
}
