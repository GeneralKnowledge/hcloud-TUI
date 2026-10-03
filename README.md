# hcloud-tui

**Hetzner Cloud TUI** — a lightweight terminal dashboard for monitoring and managing Hetzner Cloud infrastructure and estimating monthly costs.

Built for operators who live in a terminal (including over SSH): fast, keyboard-driven, and focused on Hetzner Cloud rather than generic multi-cloud management.

## Features

- Hetzner Cloud server monitoring
- Resource overview (volumes, floating IPs, snapshots)
- Estimated monthly costs from Hetzner list prices
- Cost breakdown by resource
- Unattached / unassigned resource detection (“potential waste”)
- Server details with safe actions (reboot, snapshot) and confirmation
- Keyboard-driven interface with Vim-style navigation
- SSH-friendly TUI
- Lightweight single binary

## Requirements

- Go 1.22 or later (to build from source)
- A Hetzner Cloud API token with read access (write access only if you use reboot/snapshot)

## Installation

### From source

```bash
go install github.com/hcloud-tui/hcloud-tui/cmd/hcloud-tui@latest
```

Or clone and build:

```bash
git clone https://github.com/hcloud-tui/hcloud-tui.git
cd hcloud-tui
go build -o hcloud-tui ./cmd/hcloud-tui
```

### Binary releases

Download a pre-built binary from the [GitHub Releases](https://github.com/hcloud-tui/hcloud-tui/releases) page (when available), make it executable, and place it on your `PATH`:

```bash
chmod +x hcloud-tui
sudo mv hcloud-tui /usr/local/bin/
```

## Configuration

Set your Hetzner Cloud API token (recommended):

```bash
export HCLOUD_TOKEN="..."
```

Or create `~/.config/hcloud-tui/config`:

```ini
token = "..."
```

Environment variables take precedence over the config file.

The token is only sent to the Hetzner Cloud API and is never displayed in full by the TUI.

## Usage

```bash
hcloud-tui
```

Offline demo with sample data (no API token required):

```bash
hcloud-tui --demo
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Open details |
| `Esc` | Back |
| `C` | Cost overview |
| `U` | Potential waste |
| `R` | Refresh (reboot on server detail) |
| `S` | Create snapshot (server detail) |
| `?` | Help |
| `Q` | Quit |

## Screenshots

_Screenshots coming soon._

Suggested captures for contributors:

1. Main infrastructure overview
2. Server detail view
3. Monthly cost breakdown
4. Potential waste view

## How costs work

hcloud-tui estimates monthly spend from the Hetzner Cloud pricing API and your current resources:

- Cloud servers (by type and location)
- Volumes (€/GB-month)
- Floating IPs
- Snapshots (€/GB-month)

These are **estimates**, not invoices. Taxes, traffic overages, discounts, and billing-period timing are not modelled.

## Potential waste

The waste view only flags resources that are clearly unattached or unassigned in the API:

- Volumes with no associated server
- Floating IPs with no assigned server

It does **not** treat low CPU utilisation as waste.

## Development

```bash
go test ./...
go vet ./...
go run ./cmd/hcloud-tui --demo
go build -o hcloud-tui ./cmd/hcloud-tui
```

Project layout:

```
cmd/hcloud-tui/     # main entrypoint
internal/api/      # Hetzner Cloud API client + mock
internal/config/   # token / config loading
internal/cost/     # cost calculation and waste detection
internal/model/    # shared domain types
internal/tui/      # Bubble Tea UI
```

## Contributing

Contributions are welcome.

1. Fork the repository
2. Create a feature branch
3. Add tests for non-UI logic where practical
4. Run `go test ./...` and `go vet ./...`
5. Open a pull request with a clear description

Please keep the scope focused on Hetzner Cloud monitoring and cost visibility. Avoid turning this into a generic cloud console.

## License

[MIT](LICENSE)
