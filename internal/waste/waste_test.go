package waste_test

import (
	"testing"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/waste"
)

func TestDetect_UnattachedResources(t *testing.T) {
	infra := model.Infrastructure{
		Volumes: []model.Volume{
			{ID: 1, Name: "attached", SizeGB: 10, Attached: true, ServerName: "web-01", MonthlyPrice: 0.44},
			{ID: 2, Name: "old-backup", SizeGB: 50, Attached: false, MonthlyPrice: 2.40},
		},
		FloatingIPs: []model.FloatingIP{
			{ID: 10, Name: "prod", IP: "1.2.3.4", Type: "ipv4", Assigned: true, ServerName: "web-01", MonthlyPrice: 4.00},
			{ID: 11, Name: "spare", IP: "5.6.7.8", Type: "ipv4", Assigned: false, MonthlyPrice: 4.00},
		},
		Pricing: model.Pricing{Currency: "EUR"},
	}

	report := waste.Detect(infra)

	if len(report.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(report.Items))
	}

	var sawVolume, sawIP bool
	for _, item := range report.Items {
		switch item.Kind {
		case waste.KindVolume:
			sawVolume = true
			if item.ID != 2 {
				t.Fatalf("volume ID = %d, want 2", item.ID)
			}
		case waste.KindFloatingIP:
			sawIP = true
			if item.ID != 11 {
				t.Fatalf("floating IP ID = %d, want 11", item.ID)
			}
		}
	}
	if !sawVolume || !sawIP {
		t.Fatalf("expected both volume and floating IP findings, got %#v", report.Items)
	}
	if report.PotentialSaving < 6.39 || report.PotentialSaving > 6.41 {
		t.Fatalf("PotentialSaving = %.2f, want 6.40", report.PotentialSaving)
	}
}

func TestDetect_AttachedNotReported(t *testing.T) {
	infra := model.Infrastructure{
		Volumes: []model.Volume{
			{ID: 1, Name: "data", Attached: true, MonthlyPrice: 1},
		},
		FloatingIPs: []model.FloatingIP{
			{ID: 2, Name: "vip", Assigned: true, MonthlyPrice: 4},
		},
	}
	report := waste.Detect(infra)
	if len(report.Items) != 0 {
		t.Fatalf("len(Items) = %d, want 0", len(report.Items))
	}
}
