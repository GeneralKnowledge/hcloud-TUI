package api

import (
	"context"
	"testing"
)

func TestMockFetchInventory(t *testing.T) {
	m := &Mock{Inventory: SampleInventory()}
	inv, err := m.FetchInventory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Servers) != 3 {
		t.Fatalf("servers = %d, want 3", len(inv.Servers))
	}
	if len(inv.Volumes) != 2 {
		t.Fatalf("volumes = %d, want 2", len(inv.Volumes))
	}
}

func TestMockActions(t *testing.T) {
	m := &Mock{Inventory: SampleInventory()}
	if err := m.RebootServer(context.Background(), 3); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateSnapshot(context.Background(), 3, "test"); err != nil {
		t.Fatal(err)
	}
	if len(m.Reboots) != 1 || m.Reboots[0] != 3 {
		t.Fatalf("reboots = %#v", m.Reboots)
	}
	if len(m.Snapshots) != 1 {
		t.Fatalf("snapshots = %#v", m.Snapshots)
	}
}
