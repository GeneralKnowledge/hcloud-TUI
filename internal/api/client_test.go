package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/api"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/cost"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/waste"
)

func TestMockFetchInfrastructure(t *testing.T) {
	client := api.NewMock()
	infra, err := client.FetchInfrastructure(context.Background())
	if err != nil {
		t.Fatalf("FetchInfrastructure() error = %v", err)
	}
	if len(infra.Servers) == 0 {
		t.Fatal("expected sample servers")
	}

	summary := cost.Calculate(infra)
	if summary.Total <= 0 {
		t.Fatalf("expected positive total, got %.2f", summary.Total)
	}

	report := waste.Detect(infra)
	if len(report.Items) == 0 {
		t.Fatal("expected unattached resources in mock data")
	}
}

func TestMockDeleteUnattachedOnly(t *testing.T) {
	client := api.NewMock()
	if err := client.DeleteVolume(context.Background(), 11); err == nil {
		t.Fatal("expected error deleting attached volume")
	}
	if err := client.DeleteVolume(context.Background(), 12); err != nil {
		t.Fatalf("DeleteVolume unattached: %v", err)
	}
	if err := client.DeleteFloatingIP(context.Background(), 21); err == nil {
		t.Fatal("expected error deleting assigned floating IP")
	}
	if err := client.DeleteFloatingIP(context.Background(), 22); err != nil {
		t.Fatalf("DeleteFloatingIP unassigned: %v", err)
	}
}

func TestFormatAPIError_Unauthorized(t *testing.T) {
	err := hcloud.Error{Code: hcloud.ErrorCodeUnauthorized, Message: "unauthorized"}
	msg := api.FormatAPIError("retrieve servers", err)
	if msg == "" {
		t.Fatal("empty message")
	}
	if !contains(msg, "401 Unauthorized") {
		t.Fatalf("message missing 401 text: %s", msg)
	}
	if !contains(msg, "HCLOUD_TOKEN") {
		t.Fatalf("message missing token hint: %s", msg)
	}
}

func TestFormatAPIError_Generic(t *testing.T) {
	msg := api.FormatAPIError("retrieve volumes", errors.New("connection reset"))
	if !contains(msg, "Unable to retrieve volumes") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
