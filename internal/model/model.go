// Package model defines domain types for Hetzner Cloud resources
// independent of the hcloud-go client types.
package model

import "time"

// ServerStatus is a simplified server status for display.
type ServerStatus string

const (
	ServerStatusRunning ServerStatus = "running"
	ServerStatusOff     ServerStatus = "off"
	ServerStatusError   ServerStatus = "error"
	ServerStatusUnknown ServerStatus = "unknown"
	ServerStatusOther   ServerStatus = "other"
)

// Server is a Hetzner Cloud server summary used by the TUI and cost layer.
type Server struct {
	ID           int64
	Name         string
	Status       ServerStatus
	StatusRaw    string
	ServerType   string
	Location     string
	Datacenter   string
	IPv4         string
	IPv6         string
	Cores        int
	MemoryGB     float64
	DiskGB       int
	MonthlyPrice float64
	Currency     string
	// CPUPercent is the recent average CPU utilisation when available.
	// Negative means unavailable.
	CPUPercent float64
}

// Volume is a Hetzner Cloud volume.
type Volume struct {
	ID               int64
	Name             string
	SizeGB           int
	Location         string
	ServerID         int64
	ServerName       string
	Attached         bool
	MonthlyPrice     float64
	Currency         string
	DeleteProtection bool
}

// FloatingIP is a Hetzner Cloud floating IP.
type FloatingIP struct {
	ID               int64
	Name             string
	IP               string
	Type             string
	Location         string
	ServerID         int64
	ServerName       string
	Assigned         bool
	MonthlyPrice     float64
	Currency         string
	DeleteProtection bool
}

// Snapshot is a Hetzner Cloud snapshot image.
type Snapshot struct {
	ID           int64
	Name         string
	Description  string
	ServerName   string
	Created      time.Time
	SizeGB       float64
	MonthlyPrice float64
	Currency     string
}

// Pricing holds parsed monthly prices from the Hetzner pricing API.
type Pricing struct {
	Currency            string
	ServerMonthly       map[string]map[string]float64 // serverType -> location -> monthly net
	VolumePerGBMonthly  float64
	FloatingIPMonthly   map[string]map[string]float64 // type -> location -> monthly net
	ImagePerGBMonthly   float64
	PrimaryIPMonthly    map[string]map[string]float64 // type -> location -> monthly net
}

// Infrastructure is a snapshot of account resources used by the UI and cost layer.
type Infrastructure struct {
	Servers     []Server
	Volumes     []Volume
	FloatingIPs []FloatingIP
	Snapshots   []Snapshot
	Pricing     Pricing
	FetchedAt   time.Time
}
