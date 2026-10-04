package cost

import (
	"math"
	"testing"

	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

func TestCalculateTotal(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.Server{
			{Name: "A", MonthlyCost: 5},
			{Name: "B", MonthlyCost: 10},
		},
		Volumes: []model.Volume{
			{Name: "A", MonthlyCost: 2},
		},
		FloatingIPs: []model.FloatingIP{
			{Name: "ip", IP: "1.2.3.4", MonthlyCost: 4},
		},
		Pricing: model.Pricing{Currency: "EUR"},
	}

	b := Calculate(inv)
	if math.Abs(b.Total-21) > 0.001 {
		t.Fatalf("total = %.2f, want 21.00", b.Total)
	}
	if b.Currency != "EUR" {
		t.Fatalf("currency = %q, want EUR", b.Currency)
	}
	if len(b.Items) != 4 {
		t.Fatalf("items = %d, want 4", len(b.Items))
	}
}

func TestCalculateOrdersByCostWithinCategory(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.Server{
			{Name: "cheap", MonthlyCost: 4.35},
			{Name: "expensive", MonthlyCost: 15.59},
		},
		Pricing: model.Pricing{Currency: "EUR"},
	}
	b := Calculate(inv)
	if b.Items[0].ResourceName != "expensive" {
		t.Fatalf("first item = %q, want expensive", b.Items[0].ResourceName)
	}
}

func TestFormatMoney(t *testing.T) {
	if got := FormatMoney(31.49, "EUR"); got != "€31.49" {
		t.Fatalf("FormatMoney = %q, want €31.49", got)
	}
}
