package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

// Mock is an in-memory Cloud implementation for tests and demos.
type Mock struct {
	mu        sync.Mutex
	Inventory model.Inventory
	FailFetch error
	Reboots   []int64
	Snapshots []string
}

// FetchInventory returns the configured inventory or FailFetch.
func (m *Mock) FetchInventory(ctx context.Context) (model.Inventory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FailFetch != nil {
		return model.Inventory{}, m.FailFetch
	}
	inv := m.Inventory
	inv.FetchedAt = time.Now()
	return inv, nil
}

// RebootServer records a reboot request.
func (m *Mock) RebootServer(ctx context.Context, serverID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Reboots = append(m.Reboots, serverID)
	return nil
}

// CreateSnapshot records a snapshot request.
func (m *Mock) CreateSnapshot(ctx context.Context, serverID int64, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Snapshots = append(m.Snapshots, fmt.Sprintf("%d:%s", serverID, description))
	return nil
}

// SampleInventory returns a realistic demo inventory for offline use.
func SampleInventory() model.Inventory {
	cpuWeb := 12.0
	cpuAPI := 8.0
	cpuMC := 71.0
	return model.Inventory{
		Servers: []model.Server{
			{
				ID: 1, Name: "web-01", Status: model.ServerStatusRunning,
				ServerType: "CX22", Location: "Falkenstein",
				IPv4: "203.0.113.10", IPv6: "2001:db8::10",
				CPUPercent: &cpuWeb, MemoryGB: 4, DiskGB: 40,
				MonthlyCost: 4.35, Currency: "EUR",
			},
			{
				ID: 2, Name: "api-01", Status: model.ServerStatusRunning,
				ServerType: "CX22", Location: "Nuremberg",
				IPv4: "203.0.113.20", IPv6: "2001:db8::20",
				CPUPercent: &cpuAPI, MemoryGB: 4, DiskGB: 40,
				MonthlyCost: 4.35, Currency: "EUR",
			},
			{
				ID: 3, Name: "minecraft", Status: model.ServerStatusRunning,
				ServerType: "CPX31", Location: "Helsinki",
				IPv4: "203.0.113.30", IPv6: "2001:db8::30",
				CPUPercent: &cpuMC, MemoryGB: 8, DiskGB: 160,
				MonthlyCost: 15.59, Currency: "EUR",
			},
		},
		Volumes: []model.Volume{
			{ID: 10, Name: "data", SizeGB: 50, Location: "Falkenstein", ServerID: 1, ServerName: "web-01", MonthlyCost: 2.20, Currency: "EUR"},
			{ID: 11, Name: "old-backup", SizeGB: 30, Location: "Nuremberg", ServerID: 0, MonthlyCost: 1.00, Currency: "EUR"},
		},
		FloatingIPs: []model.FloatingIP{
			{ID: 20, Name: "lb-ip", IP: "203.0.113.100", Type: "ipv4", Location: "Falkenstein", ServerID: 1, ServerName: "web-01", MonthlyCost: 4.00, Currency: "EUR"},
			{ID: 21, Name: "spare-ip", IP: "203.0.113.101", Type: "ipv4", Location: "Nuremberg", ServerID: 0, MonthlyCost: 4.00, Currency: "EUR"},
		},
		Snapshots: []model.Snapshot{
			{ID: 30, Name: "minecraft-backup", Description: "minecraft-backup", ServerName: "minecraft", SizeGB: 12.5, Created: time.Now().Add(-48 * time.Hour), MonthlyCost: 0.14, Currency: "EUR"},
		},
		Pricing: model.Pricing{
			Currency:            "EUR",
			VolumePerGBMonthly:  0.044,
			ImagePerGBMonthly:   0.011,
			FloatingIPv4Monthly: 4.00,
			FloatingIPv6Monthly: 4.00,
			ServerMonthly: map[string]float64{
				"CX22":  4.35,
				"CPX31": 15.59,
			},
		},
		FetchedAt: time.Now(),
	}
}
