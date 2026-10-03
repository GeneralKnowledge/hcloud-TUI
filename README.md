# hcloud-tui

**Hetzner Cloud TUI** — a lightweight terminal dashboard for monitoring and managing Hetzner Cloud infrastructure and estimating monthly costs.

Built for people who live in the terminal: fast, keyboard-driven, SSH-friendly, and focused entirely on Hetzner Cloud.

## Features

- Hetzner Cloud server monitoring
- Resource overview (volumes, floating IPs, snapshots)
- Estimated monthly costs from Hetzner list prices
- Cost breakdown by resource
- Unattached / unassigned resource detection
- Server details with reboot and snapshot actions
- Keyboard-driven Bubble Tea interface
- SSH-friendly layout (works in limited-colour terminals)
- Lightweight single binary

## Requirements

- Go 1.25 or newer (to build from source)
- A Hetzner Cloud API token with read access (write access needed for reboot / snapshot / delete)

## Installation

### From source

```bash
go install github.com/GeneralKnowledge/hcloud-TUI/cmd/hcloud-tui@latest
```

Or clone and build:

```bash
git clone https://github.com/GeneralKnowledge/hcloud-TUI.git
cd hcloud-TUI
go build -o hcloud-tui ./cmd/hcloud-tui
```

### Binary releases

Pre-built binaries for Linux (and later macOS / Windows) will be published on the [GitHub Releases](https://github.com/GeneralKnowledge/hcloud-TUI/releases) page.

```bash
# Example once a release exists:
curl -sL https://github.com/GeneralKnowledge/hcloud-TUI/releases/latest/download/hcloud-tui_linux_amd64.tar.gz \
  | tar -xz
sudo mv hcloud-tui /usr/local/bin/
```

## Configuration

Set your Hetzner Cloud API token:

```bash
export HCLOUD_TOKEN="..."
```

Alternatively, create `~/.config/hcloud-tui/config.yaml`:

```yaml
token: "..."
```

`HCLOUD_TOKEN` always takes precedence over the config file.

The token is only sent to the Hetzner Cloud API. It is never printed in full by the UI.

## Usage

```bash
hcloud-tui
```

Demo mode (no token required — uses sample data):

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
| `R` | Refresh (reboot on server detail) |
| `C` | Cost overview |
| `U` | Potential waste (unattached resources) |
| `S` | Create snapshot (server detail) |
| `D` | Delete selected unattached resource (with confirmation) |
| `?` | Help |
| `Q` | Quit |

## Screenshots

_Screenshots coming soon._

```
┌─ Hetzner Cloud ─────────────────────────────────────────────────────┐
│ hcloud-tui                                        Updated 10:42:18  │
├─────────────────────────────────────────────────────────────────────┤
│ SERVERS                                                             │
│ ● web-01          cx22    Falkenstein      12% CPU      €4.35/mo   │
│ ● api-01          cx22    Nuremberg         8% CPU      €4.35/mo   │
│ RESOURCES                                                           │
│ Volumes             2       80 GB                         €3.52/mo  │
│ Floating IPs        1                                      €3.57/mo │
│ Estimated monthly cost                             €…               │
│ ↑↓ Navigate  Enter Details  C Costs  U Waste  R Refresh  ? Help    │
└─────────────────────────────────────────────────────────────────────┘
```

## Cost estimates

Costs shown by hcloud-tui are **estimates** derived from the Hetzner Cloud pricing API (list prices). They are not invoices and may differ from your final bill (VAT, traffic overages, discounts, primary IP billing nuances, etc.).

## Development

```bash
# Run against your account
export HCLOUD_TOKEN="..."
go run ./cmd/hcloud-tui

# Or run the demo dataset
go run ./cmd/hcloud-tui --demo

# Tests (no Hetzner account required)
go test ./...

# Static checks
go vet ./...

# Build a standalone binary
go build -o hcloud-tui ./cmd/hcloud-tui
```

### Project layout

```
cmd/hcloud-tui/     # main entrypoint
internal/api/       # Hetzner Cloud API client + mock
internal/config/    # token / config loading
internal/cost/      # estimated monthly cost calculations
internal/model/     # domain types
internal/tui/       # Bubble Tea UI
internal/waste/     # unattached resource detection
```

## Contributing

Contributions are welcome.

1. Fork the repository
2. Create a feature branch
3. Add tests for non-UI logic where practical
4. Ensure `go test ./...` and `go vet ./...` pass
5. Open a pull request with a clear description

Please keep the scope focused on Hetzner Cloud monitoring and cost visibility. Multi-cloud support, web UIs, and large framework layers are out of scope.

## License

[MIT](LICENSE)
