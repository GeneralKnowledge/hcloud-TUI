package cost

import (
	"fmt"

	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

// WasteKind identifies a confidently detectable unattached / unassigned resource.
type WasteKind string

const (
	WasteUnattachedVolume     WasteKind = "unattached_volume"
	WasteUnassignedFloatingIP WasteKind = "unassigned_floating_ip"
)

// WasteItem is a resource that appears unused based on API attachment state only.
// Low CPU utilisation is never treated as waste.
type WasteItem struct {
	Kind         WasteKind
	ResourceType string
	ResourceName string
	Detail       string
	MonthlyCost  float64
	Currency     string
	ID           int64
}

// FindWaste returns unattached volumes and unassigned floating IPs.
func FindWaste(inv model.Inventory) []WasteItem {
	currency := inv.Pricing.Currency
	if currency == "" {
		currency = "EUR"
	}

	items := make([]WasteItem, 0)

	for _, v := range inv.Volumes {
		if v.ServerID != 0 {
			continue
		}
		items = append(items, WasteItem{
			Kind:         WasteUnattachedVolume,
			ResourceType: "Volume",
			ResourceName: v.Name,
			Detail:       formatVolumeDetail(v),
			MonthlyCost:  v.MonthlyCost,
			Currency:     currencyOr(v.Currency, currency),
			ID:           v.ID,
		})
	}

	for _, ip := range inv.FloatingIPs {
		if ip.ServerID != 0 {
			continue
		}
		name := ip.Name
		if name == "" {
			name = ip.IP
		}
		items = append(items, WasteItem{
			Kind:         WasteUnassignedFloatingIP,
			ResourceType: "Floating IP",
			ResourceName: name,
			Detail:       formatFloatingIPDetail(ip),
			MonthlyCost:  ip.MonthlyCost,
			Currency:     currencyOr(ip.Currency, currency),
			ID:           ip.ID,
		})
	}

	return items
}

// WasteTotal returns the sum of potential monthly savings from waste items.
func WasteTotal(items []WasteItem) float64 {
	var total float64
	for _, item := range items {
		total += item.MonthlyCost
	}
	return total
}

func formatVolumeDetail(v model.Volume) string {
	if v.SizeGB > 0 {
		return fmt.Sprintf("%q · unattached · %d GB", v.Name, v.SizeGB)
	}
	return fmt.Sprintf("%q · unattached", v.Name)
}

func formatFloatingIPDetail(ip model.FloatingIP) string {
	typ := ip.Type
	if typ == "" {
		typ = "IP"
	}
	return fmt.Sprintf("unassigned %s address %s", typ, ip.IP)
}
