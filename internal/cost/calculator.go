// Package cost calculates estimated monthly Hetzner Cloud costs and detects unattached resources.
package cost

import (
	"fmt"
	"sort"

	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

// CostItem is a single estimated monthly cost line.
type CostItem struct {
	ResourceType string
	ResourceName string
	MonthlyCost  float64
	Currency     string
	Category     string // servers, storage, network, images
}

// Breakdown is a categorised monthly cost estimate.
type Breakdown struct {
	Items      []CostItem
	Total      float64
	Currency   string
	Categories []CategoryTotal
}

// CategoryTotal aggregates costs for a display category.
type CategoryTotal struct {
	Name  string
	Total float64
}

// Calculate builds a cost breakdown from inventory.
// Prices are estimates based on Hetzner list prices (net), not invoices.
func Calculate(inv model.Inventory) Breakdown {
	currency := inv.Pricing.Currency
	if currency == "" {
		currency = "EUR"
	}

	items := make([]CostItem, 0, len(inv.Servers)+len(inv.Volumes)+len(inv.FloatingIPs)+len(inv.Snapshots))

	for _, s := range inv.Servers {
		items = append(items, CostItem{
			ResourceType: "server",
			ResourceName: s.Name,
			MonthlyCost:  s.MonthlyCost,
			Currency:     currencyOr(s.Currency, currency),
			Category:     "servers",
		})
	}
	for _, v := range inv.Volumes {
		items = append(items, CostItem{
			ResourceType: "volume",
			ResourceName: v.Name,
			MonthlyCost:  v.MonthlyCost,
			Currency:     currencyOr(v.Currency, currency),
			Category:     "storage",
		})
	}
	for _, ip := range inv.FloatingIPs {
		name := ip.Name
		if name == "" {
			name = ip.IP
		}
		items = append(items, CostItem{
			ResourceType: "floating_ip",
			ResourceName: name,
			MonthlyCost:  ip.MonthlyCost,
			Currency:     currencyOr(ip.Currency, currency),
			Category:     "network",
		})
	}
	for _, snap := range inv.Snapshots {
		name := snap.Name
		if name == "" {
			name = snap.Description
		}
		items = append(items, CostItem{
			ResourceType: "snapshot",
			ResourceName: name,
			MonthlyCost:  snap.MonthlyCost,
			Currency:     currencyOr(snap.Currency, currency),
			Category:     "images",
		})
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

	var total float64
	catMap := map[string]float64{}
	for _, item := range items {
		total += item.MonthlyCost
		catMap[item.Category] += item.MonthlyCost
	}

	cats := make([]CategoryTotal, 0, len(catMap))
	for _, name := range []string{"servers", "storage", "network", "images"} {
		if t, ok := catMap[name]; ok {
			cats = append(cats, CategoryTotal{Name: name, Total: t})
		}
	}

	return Breakdown{
		Items:      items,
		Total:      total,
		Currency:   currency,
		Categories: cats,
	}
}

// FormatEUR formats a euro amount for display.
func FormatEUR(amount float64) string {
	return fmt.Sprintf("€%.2f", amount)
}

// FormatMoney formats an amount with a currency symbol when known.
func FormatMoney(amount float64, currency string) string {
	switch currency {
	case "", "EUR", "€":
		return FormatEUR(amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

func currencyOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func categoryOrder(name string) int {
	switch name {
	case "servers":
		return 0
	case "storage":
		return 1
	case "network":
		return 2
	case "images":
		return 3
	default:
		return 9
	}
}
