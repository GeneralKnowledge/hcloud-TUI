package cost

import (
	"testing"

	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

func TestFindWasteReportsOnlyUnattached(t *testing.T) {
	inv := model.Inventory{
		Volumes: []model.Volume{
			{ID: 1, Name: "attached", ServerID: 10, ServerName: "web-01", SizeGB: 20, MonthlyCost: 0.88},
			{ID: 2, Name: "old-backup", ServerID: 0, SizeGB: 50, MonthlyCost: 2.40},
		},
		FloatingIPs: []model.FloatingIP{
			{ID: 3, Name: "prod", IP: "1.1.1.1", Type: "ipv4", ServerID: 10, ServerName: "web-01", MonthlyCost: 4.00},
			{ID: 4, Name: "spare", IP: "2.2.2.2", Type: "ipv4", ServerID: 0, MonthlyCost: 4.00},
		},
		Pricing: model.Pricing{Currency: "EUR"},
	}

	items := FindWaste(inv)
	if len(items) != 2 {
		t.Fatalf("waste items = %d, want 2", len(items))
	}

	var sawVolume, sawIP bool
	for _, item := range items {
		switch item.Kind {
		case WasteUnattachedVolume:
			sawVolume = true
			if item.ResourceName != "old-backup" {
				t.Fatalf("volume name = %q, want old-backup", item.ResourceName)
			}
		case WasteUnassignedFloatingIP:
			sawIP = true
			if item.ID != 4 {
				t.Fatalf("floating IP id = %d, want 4", item.ID)
			}
		default:
			t.Fatalf("unexpected kind %q", item.Kind)
		}
	}
	if !sawVolume || !sawIP {
		t.Fatalf("missing expected waste kinds: volume=%v ip=%v", sawVolume, sawIP)
	}

	if got := WasteTotal(items); got != 6.40 {
		t.Fatalf("WasteTotal = %.2f, want 6.40", got)
	}
}

func TestFindWasteIgnoresLowCPUServers(t *testing.T) {
	cpu := 1.0
	inv := model.Inventory{
		Servers: []model.Server{
			{Name: "prod", Status: model.ServerStatusRunning, CPUPercent: &cpu, MonthlyCost: 4.35},
		},
	}
	if items := FindWaste(inv); len(items) != 0 {
		t.Fatalf("expected no waste from low CPU, got %#v", items)
	}
}
