// Package cost calculates estimated monthly Hetzner Cloud costs.
// Estimates are based on published API pricing and are not invoices.
package cost

import (
	"fmt"
	"sort"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

// CostItem is a single billable line item.
type CostItem struct {
	ResourceType string
	ResourceName string
	MonthlyCost  float64
	Currency     string
	Category     string
}

// Summary is an aggregated cost estimate.
type Summary struct {
	Items    []CostItem
	Total    float64
	Currency string
}

// Categories used for grouping on the cost screen.
const (
	CategoryServers = "SERVERS"
	CategoryStorage = "STORAGE"
	CategoryNetwork = "NETWORK"
	CategoryOther   = "OTHER"
)

// Calculate builds an estimated monthly cost summary from infrastructure state.
func Calculate(infra model.Infrastructure) Summary {
	items := make([]CostItem, 0)

	currency := infra.Pricing.Currency
	if currency == "" {
		currency = "EUR"
	}

	for _, s := range infra.Servers {
		cur := s.Currency
		if cur == "" {
			cur = currency
		}
		items = append(items, CostItem{
			ResourceType: "server",
			ResourceName: s.Name,
			MonthlyCost:  s.MonthlyPrice,
			Currency:     cur,
			Category:     CategoryServers,
		})
	}

	for _, v := range infra.Volumes {
		cur := v.Currency
		if cur == "" {
			cur = currency
		}
		items = append(items, CostItem{
			ResourceType: "volume",
			ResourceName: v.Name,
			MonthlyCost:  v.MonthlyPrice,
			Currency:     cur,
			Category:     CategoryStorage,
		})
	}

	for _, snap := range infra.Snapshots {
		cur := snap.Currency
		if cur == "" {
			cur = currency
		}
		name := snap.Name
		if name == "" {
			name = snap.Description
		}
		if name == "" {
			name = fmt.Sprintf("snapshot-%d", snap.ID)
		}
		items = append(items, CostItem{
			ResourceType: "snapshot",
			ResourceName: name,
			MonthlyCost:  snap.MonthlyPrice,
			Currency:     cur,
			Category:     CategoryStorage,
		})
	}

	for _, ip := range infra.FloatingIPs {
		cur := ip.Currency
		if cur == "" {
			cur = currency
		}
		name := ip.Name
		if name == "" {
			name = ip.IP
		}
		items = append(items, CostItem{
			ResourceType: "floating_ip",
			ResourceName: name,
			MonthlyCost:  ip.MonthlyPrice,
			Currency:     cur,
			Category:     CategoryNetwork,
		})
	}

	total := 0.0
	for _, item := range items {
		total += item.MonthlyCost
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Category != items[j].Category {
			return categoryOrder(items[i].Category) < categoryOrder(items[j].Category)
		}
		if items[i].MonthlyCost != items[j].MonthlyCost {
			return items[i].MonthlyCost > items[j].MonthlyCost
		}
		return items[i].ResourceName < items[j].ResourceName
	})

	return Summary{
		Items:    items,
		Total:    total,
		Currency: currency,
	}
}

// AggregateByCategory sums costs per category for chart-style views.
func AggregateByCategory(summary Summary) []CostItem {
	totals := map[string]float64{}
	order := []string{CategoryServers, CategoryStorage, CategoryNetwork, CategoryOther}
	for _, item := range summary.Items {
		cat := item.Category
		if cat == "" {
			cat = CategoryOther
		}
		totals[cat] += item.MonthlyCost
	}
	out := make([]CostItem, 0, len(order))
	for _, cat := range order {
		if amount, ok := totals[cat]; ok {
			out = append(out, CostItem{
				ResourceType: "category",
				ResourceName: cat,
				MonthlyCost:  amount,
				Currency:     summary.Currency,
				Category:     cat,
			})
		}
	}
	return out
}

// FormatMoney formats an amount with a euro-style prefix for EUR, otherwise ISO code.
func FormatMoney(amount float64, currency string) string {
	if currency == "" || currency == "EUR" {
		return fmt.Sprintf("€%.2f", amount)
	}
	return fmt.Sprintf("%.2f %s", amount, currency)
}

func categoryOrder(cat string) int {
	switch cat {
	case CategoryServers:
		return 0
	case CategoryStorage:
		return 1
	case CategoryNetwork:
		return 2
	default:
		return 3
	}
}
