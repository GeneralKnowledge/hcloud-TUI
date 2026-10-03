// Package waste detects conservatively identifiable unattached Hetzner resources.
package waste

import (
	"fmt"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

// Kind identifies the type of potentially unused resource.
type Kind string

const (
	KindVolume     Kind = "volume"
	KindFloatingIP Kind = "floating_ip"
)

// Item is a resource that appears unattached / not assigned based on API state.
type Item struct {
	Kind         Kind
	ID           int64
	Title        string
	Detail       string
	MonthlyCost  float64
	Currency     string
	ResourceName string
}

// Report lists conservatively detected unattached resources.
type Report struct {
	Items            []Item
	PotentialSaving  float64
	Currency         string
}

// Detect finds unattached volumes and unassigned floating IPs.
// It does not consider low CPU utilisation as waste.
func Detect(infra model.Infrastructure) Report {
	currency := infra.Pricing.Currency
	if currency == "" {
		currency = "EUR"
	}

	items := make([]Item, 0)

	for _, v := range infra.Volumes {
		if v.Attached {
			continue
		}
		cur := v.Currency
		if cur == "" {
			cur = currency
		}
		items = append(items, Item{
			Kind:         KindVolume,
			ID:           v.ID,
			Title:        "Volume",
			Detail:       fmt.Sprintf("%q · %d GB · no associated server", v.Name, v.SizeGB),
			MonthlyCost:  v.MonthlyPrice,
			Currency:     cur,
			ResourceName: v.Name,
		})
	}

	for _, ip := range infra.FloatingIPs {
		if ip.Assigned {
			continue
		}
		cur := ip.Currency
		if cur == "" {
			cur = currency
		}
		label := ip.IP
		if ip.Type != "" {
			label = fmt.Sprintf("%s %s", ip.Type, ip.IP)
		}
		items = append(items, Item{
			Kind:         KindFloatingIP,
			ID:           ip.ID,
			Title:        "Floating IP",
			Detail:       fmt.Sprintf("%s · not assigned to a server", label),
			MonthlyCost:  ip.MonthlyPrice,
			Currency:     cur,
			ResourceName: ip.Name,
		})
	}

	saving := 0.0
	for _, item := range items {
		saving += item.MonthlyCost
	}

	return Report{
		Items:           items,
		PotentialSaving: saving,
		Currency:        currency,
	}
}
