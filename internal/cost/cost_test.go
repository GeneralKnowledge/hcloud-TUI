package cost_test

import (
	"math"
	"testing"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/cost"
	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

func TestCalculate_Total(t *testing.T) {
	infra := model.Infrastructure{
		Servers: []model.Server{
			{Name: "A", MonthlyPrice: 5, Currency: "EUR"},
			{Name: "B", MonthlyPrice: 10, Currency: "EUR"},
		},
		Volumes: []model.Volume{
			{Name: "A", MonthlyPrice: 2, Currency: "EUR"},
		},
		FloatingIPs: []model.FloatingIP{
			{Name: "ip", MonthlyPrice: 4, Currency: "EUR"},
		},
		Pricing: model.Pricing{Currency: "EUR"},
	}

	summary := cost.Calculate(infra)
	if math.Abs(summary.Total-21) > 0.001 {
		t.Fatalf("Total = %.2f, want 21.00", summary.Total)
	}
	if summary.Currency != "EUR" {
		t.Fatalf("Currency = %q, want EUR", summary.Currency)
	}
	if len(summary.Items) != 4 {
		t.Fatalf("len(Items) = %d, want 4", len(summary.Items))
	}
}

func TestCalculate_Empty(t *testing.T) {
	summary := cost.Calculate(model.Infrastructure{})
	if summary.Total != 0 {
		t.Fatalf("Total = %.2f, want 0", summary.Total)
	}
}

func TestFormatMoney(t *testing.T) {
	if got := cost.FormatMoney(31.49, "EUR"); got != "€31.49" {
		t.Fatalf("FormatMoney EUR = %q, want €31.49", got)
	}
	if got := cost.FormatMoney(10, "USD"); got != "10.00 USD" {
		t.Fatalf("FormatMoney USD = %q, want 10.00 USD", got)
	}
}

func TestAggregateByCategory(t *testing.T) {
	summary := cost.Calculate(model.Infrastructure{
		Servers:     []model.Server{{Name: "s", MonthlyPrice: 10}},
		Volumes:     []model.Volume{{Name: "v", MonthlyPrice: 3}},
		FloatingIPs: []model.FloatingIP{{Name: "ip", MonthlyPrice: 4}},
		Pricing:     model.Pricing{Currency: "EUR"},
	})
	cats := cost.AggregateByCategory(summary)
	if len(cats) != 3 {
		t.Fatalf("len(cats) = %d, want 3", len(cats))
	}
}
