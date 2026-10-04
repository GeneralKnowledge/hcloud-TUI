// Package api provides a thin abstraction over the Hetzner Cloud API.
package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/hcloud-tui/hcloud-tui/internal/model"
)

// Cloud is the subset of Hetzner Cloud operations used by hcloud-tui.
type Cloud interface {
	FetchInventory(ctx context.Context) (model.Inventory, error)
	RebootServer(ctx context.Context, serverID int64) error
	CreateSnapshot(ctx context.Context, serverID int64, description string) error
}

// Client wraps hcloud.Client.
type Client struct {
	hc *hcloud.Client
}

// New creates a Hetzner Cloud API client. The token is only sent to the Hetzner API.
func New(token string) *Client {
	return &Client{
		hc: hcloud.NewClient(
			hcloud.WithToken(token),
			hcloud.WithApplication("hcloud-tui", "0.1.0"),
		),
	}
}

// FetchInventory loads servers, volumes, floating IPs, snapshots, and pricing.
func (c *Client) FetchInventory(ctx context.Context) (model.Inventory, error) {
	pricing, _, err := c.hc.Pricing.Get(ctx)
	if err != nil {
		return model.Inventory{}, formatAPIError("Unable to retrieve pricing", err)
	}
	modelPricing := convertPricing(pricing)

	servers, err := c.hc.Server.All(ctx)
	if err != nil {
		return model.Inventory{}, formatAPIError("Unable to retrieve servers", err)
	}

	volumes, err := c.hc.Volume.All(ctx)
	if err != nil {
		return model.Inventory{}, formatAPIError("Unable to retrieve volumes", err)
	}

	floatingIPs, err := c.hc.FloatingIP.All(ctx)
	if err != nil {
		return model.Inventory{}, formatAPIError("Unable to retrieve floating IPs", err)
	}

	images, err := c.hc.Image.AllWithOpts(ctx, hcloud.ImageListOpts{
		Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot},
	})
	if err != nil {
		return model.Inventory{}, formatAPIError("Unable to retrieve snapshots", err)
	}

	serverByID := map[int64]*hcloud.Server{}
	for _, s := range servers {
		serverByID[s.ID] = s
	}

	out := model.Inventory{
		Servers:     make([]model.Server, 0, len(servers)),
		Volumes:     make([]model.Volume, 0, len(volumes)),
		FloatingIPs: make([]model.FloatingIP, 0, len(floatingIPs)),
		Snapshots:   make([]model.Snapshot, 0, len(images)),
		Pricing:     modelPricing,
		FetchedAt:   time.Now(),
	}

	for _, s := range servers {
		ms := convertServer(s, modelPricing)
		if cpu, ok := c.fetchCPUPercent(ctx, s); ok {
			ms.CPUPercent = &cpu
		}
		out.Servers = append(out.Servers, ms)
	}

	for _, v := range volumes {
		out.Volumes = append(out.Volumes, convertVolume(v, modelPricing, serverByID))
	}
	for _, ip := range floatingIPs {
		out.FloatingIPs = append(out.FloatingIPs, convertFloatingIP(ip, modelPricing, serverByID))
	}
	for _, img := range images {
		out.Snapshots = append(out.Snapshots, convertSnapshot(img, modelPricing))
	}

	return out, nil
}

// RebootServer soft-reboots a server.
func (c *Client) RebootServer(ctx context.Context, serverID int64) error {
	server, _, err := c.hc.Server.GetByID(ctx, serverID)
	if err != nil {
		return formatAPIError("Unable to reboot server", err)
	}
	if server == nil {
		return fmt.Errorf("server %d not found", serverID)
	}
	_, _, err = c.hc.Server.Reboot(ctx, server)
	if err != nil {
		return formatAPIError("Unable to reboot server", err)
	}
	return nil
}

// CreateSnapshot creates a snapshot image from a server.
func (c *Client) CreateSnapshot(ctx context.Context, serverID int64, description string) error {
	server, _, err := c.hc.Server.GetByID(ctx, serverID)
	if err != nil {
		return formatAPIError("Unable to create snapshot", err)
	}
	if server == nil {
		return fmt.Errorf("server %d not found", serverID)
	}
	opts := hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.Ptr(description),
	}
	_, _, err = c.hc.Server.CreateImage(ctx, server, &opts)
	if err != nil {
		return formatAPIError("Unable to create snapshot", err)
	}
	return nil
}

func (c *Client) fetchCPUPercent(ctx context.Context, server *hcloud.Server) (float64, bool) {
	if server.Status != hcloud.ServerStatusRunning {
		return 0, false
	}
	end := time.Now()
	start := end.Add(-15 * time.Minute)
	metrics, _, err := c.hc.Server.GetMetrics(ctx, server, hcloud.ServerGetMetricsOpts{
		Types: []hcloud.ServerMetricType{hcloud.ServerMetricCPU},
		Start: start,
		End:   end,
		Step:  60,
	})
	if err != nil || metrics == nil {
		return 0, false
	}
	series, ok := metrics.TimeSeries["cpu"]
	if !ok || len(series) == 0 {
		// Some API responses use a namespaced key.
		for key, values := range metrics.TimeSeries {
			if strings.Contains(strings.ToLower(key), "cpu") && len(values) > 0 {
				series = values
				ok = true
				break
			}
		}
	}
	if !ok || len(series) == 0 {
		return 0, false
	}
	last := series[len(series)-1]
	v, err := strconv.ParseFloat(last.Value, 64)
	if err != nil {
		return 0, false
	}
	// Hetzner CPU metrics are typically 0–1; normalise to percent.
	if v <= 1.5 {
		v *= 100
	}
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return v, true
}

func convertPricing(p hcloud.Pricing) model.Pricing {
	currency := p.Image.PerGBMonth.Currency
	if currency == "" {
		currency = p.Volume.PerGBMonthly.Currency
	}
	if currency == "" {
		currency = "EUR"
	}
	out := model.Pricing{
		Currency:      currency,
		ServerMonthly: map[string]float64{},
	}
	out.VolumePerGBMonthly = parsePrice(p.Volume.PerGBMonthly.Net)
	out.ImagePerGBMonthly = parsePrice(p.Image.PerGBMonth.Net)

	for _, fp := range p.FloatingIPs {
		for _, loc := range fp.Pricings {
			monthly := parsePrice(loc.Monthly.Net)
			switch fp.Type {
			case hcloud.FloatingIPTypeIPv4:
				if out.FloatingIPv4Monthly == 0 || monthly > 0 {
					out.FloatingIPv4Monthly = monthly
				}
			case hcloud.FloatingIPTypeIPv6:
				if out.FloatingIPv6Monthly == 0 || monthly > 0 {
					out.FloatingIPv6Monthly = monthly
				}
			}
		}
	}
	// Fallback for older pricing payloads.
	if out.FloatingIPv4Monthly == 0 {
		out.FloatingIPv4Monthly = parsePrice(p.FloatingIP.Monthly.Net)
	}

	for _, st := range p.ServerTypes {
		if st.ServerType == nil {
			continue
		}
		for _, locPrice := range st.Pricings {
			locName := ""
			if locPrice.Location != nil {
				locName = locPrice.Location.Name
			}
			key := serverPriceKey(st.ServerType.Name, locName)
			out.ServerMonthly[key] = parsePrice(locPrice.Monthly.Net)
			// Also store type-only fallback.
			if _, exists := out.ServerMonthly[st.ServerType.Name]; !exists {
				out.ServerMonthly[st.ServerType.Name] = parsePrice(locPrice.Monthly.Net)
			}
		}
	}
	return out
}

func convertServer(s *hcloud.Server, pricing model.Pricing) model.Server {
	out := model.Server{
		ID:       s.ID,
		Name:     s.Name,
		Status:   model.ServerStatus(s.Status),
		Created:  s.Created,
		Currency: pricing.Currency,
	}
	if s.ServerType != nil {
		out.ServerType = s.ServerType.Name
		out.MemoryGB = float64(s.ServerType.Memory)
		out.DiskGB = s.ServerType.Disk
	}
	loc := serverLocation(s)
	if loc != nil {
		out.Location = locationDisplay(loc)
	}
	if !s.PublicNet.IPv4.IsUnspecified() {
		out.IPv4 = s.PublicNet.IPv4.IP.String()
	}
	if !s.PublicNet.IPv6.IsUnspecified() {
		out.IPv6 = s.PublicNet.IPv6.IP.String()
	}
	out.MonthlyCost = lookupServerPrice(pricing, out.ServerType, locationName(loc))
	return out
}

func serverLocation(s *hcloud.Server) *hcloud.Location {
	if s == nil || s.Datacenter == nil {
		return nil
	}
	return s.Datacenter.Location
}

func convertVolume(v *hcloud.Volume, pricing model.Pricing, servers map[int64]*hcloud.Server) model.Volume {
	out := model.Volume{
		ID:               v.ID,
		Name:             v.Name,
		SizeGB:           v.Size,
		Currency:         pricing.Currency,
		ProtectionDelete: v.Protection.Delete,
		MonthlyCost:      float64(v.Size) * pricing.VolumePerGBMonthly,
	}
	if v.Location != nil {
		out.Location = locationDisplay(v.Location)
	}
	if v.Server != nil {
		out.ServerID = v.Server.ID
		if s, ok := servers[v.Server.ID]; ok {
			out.ServerName = s.Name
		}
	}
	return out
}

func convertFloatingIP(ip *hcloud.FloatingIP, pricing model.Pricing, servers map[int64]*hcloud.Server) model.FloatingIP {
	out := model.FloatingIP{
		ID:               ip.ID,
		Name:             ip.Name,
		Type:             string(ip.Type),
		Currency:         pricing.Currency,
		ProtectionDelete: ip.Protection.Delete,
	}
	if ip.IP != nil {
		out.IP = ip.IP.String()
	}
	if ip.HomeLocation != nil {
		out.Location = locationDisplay(ip.HomeLocation)
	}
	switch ip.Type {
	case hcloud.FloatingIPTypeIPv6:
		out.MonthlyCost = pricing.FloatingIPv6Monthly
	default:
		out.MonthlyCost = pricing.FloatingIPv4Monthly
	}
	if ip.Server != nil {
		out.ServerID = ip.Server.ID
		if s, ok := servers[ip.Server.ID]; ok {
			out.ServerName = s.Name
		}
	}
	return out
}

func convertSnapshot(img *hcloud.Image, pricing model.Pricing) model.Snapshot {
	out := model.Snapshot{
		ID:          img.ID,
		Name:        img.Name,
		Description: img.Description,
		SizeGB:      float64(img.ImageSize),
		Created:     img.Created,
		Currency:    pricing.Currency,
		MonthlyCost: float64(img.ImageSize) * pricing.ImagePerGBMonthly,
	}
	if img.CreatedFrom != nil {
		out.ServerName = img.CreatedFrom.Name
	} else if img.BoundTo != nil {
		out.ServerName = img.BoundTo.Name
	}
	if out.Name == "" {
		out.Name = img.Description
	}
	return out
}

func lookupServerPrice(pricing model.Pricing, serverType, location string) float64 {
	if pricing.ServerMonthly == nil {
		return 0
	}
	if location != "" {
		if p, ok := pricing.ServerMonthly[serverPriceKey(serverType, location)]; ok {
			return p
		}
	}
	if p, ok := pricing.ServerMonthly[serverType]; ok {
		return p
	}
	return 0
}

func serverPriceKey(serverType, location string) string {
	return serverType + "|" + location
}

func locationName(loc *hcloud.Location) string {
	if loc == nil {
		return ""
	}
	return loc.Name
}

func locationDisplay(loc *hcloud.Location) string {
	if loc == nil {
		return ""
	}
	if loc.City != "" {
		return loc.City
	}
	return loc.Name
}

func parsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func formatAPIError(prefix string, err error) error {
	if err == nil {
		return nil
	}
	var apiErr hcloud.Error
	if hcloud.IsError(err, hcloud.ErrorCodeUnauthorized) || asHCloudError(err, &apiErr) && apiErr.Code == hcloud.ErrorCodeUnauthorized {
		return fmt.Errorf("%s.\nHetzner API returned:\n401 Unauthorized\nCheck that HCLOUD_TOKEN is valid", prefix)
	}
	if asHCloudError(err, &apiErr) {
		return fmt.Errorf("%s.\nHetzner API returned:\n%s (%s)", prefix, apiErr.Message, apiErr.Code)
	}
	return fmt.Errorf("%s.\n%s", prefix, err.Error())
}

func asHCloudError(err error, target *hcloud.Error) bool {
	e, ok := err.(hcloud.Error)
	if !ok {
		return false
	}
	*target = e
	return true
}
