// Package model defines domain types shared across the API, cost, and TUI layers.
package model

import "time"

// ServerStatus represents a Hetzner Cloud server status.
type ServerStatus string

const (
	ServerStatusRunning      ServerStatus = "running"
	ServerStatusOff          ServerStatus = "off"
	ServerStatusInitializing ServerStatus = "initializing"
	ServerStatusStarting     ServerStatus = "starting"
	ServerStatusStopping     ServerStatus = "stopping"
	ServerStatusMigrating    ServerStatus = "migrating"
	ServerStatusRebuilding   ServerStatus = "rebuilding"
	ServerStatusDeleting     ServerStatus = "deleting"
	ServerStatusUnknown      ServerStatus = "unknown"
)

// Server is a Hetzner Cloud server as displayed by hcloud-tui.
type Server struct {
	ID           int64
	Name         string
	Status       ServerStatus
	ServerType   string
	Location     string
	IPv4         string
	IPv6         string
	CPUPercent   *float64 // nil when metrics are unavailable
	MemoryGB     float64
	DiskGB       int
	MonthlyCost  float64
	Currency     string
	Created      time.Time
}

// Volume is a Hetzner Cloud volume.
type Volume struct {
	ID              int64
	Name            string
	SizeGB          int
	Location        string
	ServerID        int64  // 0 when unattached
	ServerName      string // empty when unattached
	MonthlyCost     float64
	Currency        string
	ProtectionDelete bool
}

// FloatingIP is a Hetzner Cloud floating IP.
type FloatingIP struct {
	ID               int64
	Name             string
	IP               string
	Type             string // ipv4 or ipv6
	Location         string
	ServerID         int64  // 0 when unassigned
	ServerName       string // empty when unassigned
	MonthlyCost      float64
	Currency         string
	ProtectionDelete bool
}

// Snapshot is a Hetzner Cloud snapshot image.
type Snapshot struct {
	ID          int64
	Name        string
	Description string
	ServerName  string
	SizeGB      float64
	Created     time.Time
	MonthlyCost float64
	Currency    string
}

// Pricing holds currency and per-resource unit prices from the Hetzner API.
type Pricing struct {
	Currency            string
	VolumePerGBMonthly  float64
	ImagePerGBMonthly   float64
	FloatingIPv4Monthly float64
	FloatingIPv6Monthly float64
	// ServerMonthly maps "serverType|location" -> monthly net price.
	ServerMonthly map[string]float64
}

// Inventory is a snapshot of billable Hetzner Cloud resources.
type Inventory struct {
	Servers     []Server
	Volumes     []Volume
	FloatingIPs []FloatingIP
	Snapshots   []Snapshot
	Pricing     Pricing
	FetchedAt   time.Time
}

// StatusSymbol returns a colour-independent status indicator.
func StatusSymbol(status ServerStatus) string {
	switch status {
	case ServerStatusRunning:
		return "●"
	case ServerStatusOff, ServerStatusStopping:
		return "○"
	case ServerStatusInitializing, ServerStatusStarting, ServerStatusMigrating, ServerStatusRebuilding, ServerStatusDeleting:
		return "…"
	default:
		return "?"
	}
}

// StatusLabel returns a human-readable status label.
func StatusLabel(status ServerStatus) string {
	switch status {
	case ServerStatusRunning:
		return "RUNNING"
	case ServerStatusOff:
		return "OFF"
	case ServerStatusInitializing:
		return "INITIALIZING"
	case ServerStatusStarting:
		return "STARTING"
	case ServerStatusStopping:
		return "STOPPING"
	case ServerStatusMigrating:
		return "MIGRATING"
	case ServerStatusRebuilding:
		return "REBUILDING"
	case ServerStatusDeleting:
		return "DELETING"
	default:
		return "UNKNOWN"
	}
}
