// Command hcloud-tui is a lightweight terminal dashboard for monitoring
// Hetzner Cloud infrastructure and estimating monthly costs.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/api"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/config"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help", "help":
			printUsage()
			return nil
		case "-v", "--version", "version":
			fmt.Println("hcloud-tui 0.1.0")
			return nil
		case "--demo":
			return startUI(api.NewMock())
		}
	}

	cfg, err := config.Load("")
	if err != nil {
		if err == config.ErrMissingToken {
			fmt.Fprintln(os.Stderr, config.MissingTokenMessage())
			os.Exit(1)
		}
		return err
	}

	client := api.New(cfg.Token)
	return startUI(client)
}

func startUI(client api.Client) error {
	app := tui.New(client)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func printUsage() {
	fmt.Print(`hcloud-tui — Hetzner Cloud TUI

A lightweight terminal dashboard for monitoring and managing
Hetzner Cloud infrastructure and estimating monthly costs.

Usage:
  hcloud-tui
  hcloud-tui --demo

Authentication:
  export HCLOUD_TOKEN="..."

Options:
  --demo       Run with sample data (no API token required)
  -h, --help   Show this help
  -v, --version
`)
}
