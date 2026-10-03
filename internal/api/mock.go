package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

// MockClient is an in-memory Client for tests and demos.
type MockClient struct {
	mu       sync.Mutex
	Infra    model.Infrastructure
	FetchErr error
	Reboots  []int64
	Snaps    []int64
}

// NewMock returns a MockClient populated with sample Hetzner-like data.
func NewMock() *MockClient {
	return &MockClient{
		Infra: model.Infrastructure{
			Servers: []model.Server{
				{
					ID: 1, Name: "web-01", Status: model.ServerStatusRunning, StatusRaw: "running",
					ServerType: "cx22", Location: "Falkenstein", Datacenter: "fsn1",
					IPv4: "203.0.113.10", IPv6: "2001:db8::10",
					Cores: 2, MemoryGB: 4, DiskGB: 40, MonthlyPrice: 4.35, Currency: "EUR",
					CPUPercent: 12,
				},
				{
					ID: 2, Name: "api-01", Status: model.ServerStatusRunning, StatusRaw: "running",
					ServerType: "cx22", Location: "Nuremberg", Datacenter: "nbg1",
					IPv4: "203.0.113.20", IPv6: "2001:db8::20",
					Cores: 2, MemoryGB: 4, DiskGB: 40, MonthlyPrice: 4.35, Currency: "EUR",
					CPUPercent: 8,
				},
				{
					ID: 3, Name: "minecraft", Status: model.ServerStatusRunning, StatusRaw: "running",
					ServerType: "cpx31", Location: "Helsinki", Datacenter: "hel1",
					IPv4: "203.0.113.30", IPv6: "2001:db8::30",
					Cores: 4, MemoryGB: 8, DiskGB: 160, MonthlyPrice: 15.59, Currency: "EUR",
					CPUPercent: 71,
				},
			},
			Volumes: []model.Volume{
				{ID: 11, Name: "data", SizeGB: 30, Location: "fsn1", Attached: true, ServerID: 1, ServerName: "web-01", MonthlyPrice: 1.32, Currency: "EUR"},
				{ID: 12, Name: "old-backup", SizeGB: 50, Location: "fsn1", Attached: false, MonthlyPrice: 2.20, Currency: "EUR"},
			},
			FloatingIPs: []model.FloatingIP{
				{ID: 21, Name: "vip", IP: "203.0.113.100", Type: "ipv4", Location: "fsn1", Assigned: true, ServerID: 1, ServerName: "web-01", MonthlyPrice: 3.57, Currency: "EUR"},
				{ID: 22, Name: "spare", IP: "203.0.113.101", Type: "ipv4", Location: "nbg1", Assigned: false, MonthlyPrice: 3.57, Currency: "EUR"},
			},
			Snapshots: []model.Snapshot{
				{ID: 31, Name: "web-01-backup", ServerName: "web-01", Created: time.Now().Add(-48 * time.Hour), SizeGB: 5, MonthlyPrice: 0.05, Currency: "EUR"},
			},
			Pricing: model.Pricing{
				Currency:           "EUR",
				VolumePerGBMonthly: 0.044,
				ImagePerGBMonthly:  0.01,
			},
			FetchedAt: time.Now(),
		},
	}
}

// FetchInfrastructure returns the mock infrastructure snapshot.
func (m *MockClient) FetchInfrastructure(ctx context.Context) (model.Infrastructure, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FetchErr != nil {
		return model.Infrastructure{}, m.FetchErr
	}
	select {
	case <-ctx.Done():
		return model.Infrastructure{}, ctx.Err()
	default:
	}
	infra := m.Infra
	infra.FetchedAt = time.Now()
	return infra, nil
}

// RebootServer records a reboot request.
func (m *MockClient) RebootServer(ctx context.Context, serverID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Reboots = append(m.Reboots, serverID)
	return nil
}

// CreateSnapshot records a snapshot request.
func (m *MockClient) CreateSnapshot(ctx context.Context, serverID int64, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Snaps = append(m.Snaps, serverID)
	_ = description
	return nil
}

// DeleteVolume removes an unattached volume from the mock inventory.
func (m *MockClient) DeleteVolume(ctx context.Context, volumeID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.Infra.Volumes[:0]
	found := false
	for _, v := range m.Infra.Volumes {
		if v.ID == volumeID {
			found = true
			if v.Attached {
				return fmt.Errorf("volume is still attached to a server; detach it first")
			}
			continue
		}
		kept = append(kept, v)
	}
	if !found {
		return fmt.Errorf("volume %d not found", volumeID)
	}
	m.Infra.Volumes = kept
	return nil
}

// DeleteFloatingIP removes an unassigned floating IP from the mock inventory.
func (m *MockClient) DeleteFloatingIP(ctx context.Context, floatingIPID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.Infra.FloatingIPs[:0]
	found := false
	for _, ip := range m.Infra.FloatingIPs {
		if ip.ID == floatingIPID {
			found = true
			if ip.Assigned {
				return fmt.Errorf("floating IP is still assigned to a server; unassign it first")
			}
			continue
		}
		kept = append(kept, ip)
	}
	if !found {
		return fmt.Errorf("floating IP %d not found", floatingIPID)
	}
	m.Infra.FloatingIPs = kept
	return nil
}

// SetFetchError configures the next FetchInfrastructure call to fail.
func (m *MockClient) SetFetchError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FetchErr = err
}

// Ensure MockClient satisfies Client.
var _ Client = (*MockClient)(nil)

// ErrDemo is returned when mock mode is intentionally broken for UI testing.
var ErrDemo = fmt.Errorf("demo error")
