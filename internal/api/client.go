// Package api provides a thin abstraction over the Hetzner Cloud API.
package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/model"
)

// Client is the subset of Hetzner Cloud operations used by hcloud-tui.
type Client interface {
	FetchInfrastructure(ctx context.Context) (model.Infrastructure, error)
	RebootServer(ctx context.Context, serverID int64) error
	CreateSnapshot(ctx context.Context, serverID int64, description string) error
	DeleteVolume(ctx context.Context, volumeID int64) error
	DeleteFloatingIP(ctx context.Context, floatingIPID int64) error
}

// HCloudClient implements Client using hcloud-go.
type HCloudClient struct {
	client *hcloud.Client
}

// New creates a Hetzner Cloud API client.
func New(token string) *HCloudClient {
	return &HCloudClient{
		client: hcloud.NewClient(
			hcloud.WithToken(token),
			hcloud.WithApplication("hcloud-tui", "0.1.0"),
		),
	}
}

// FetchInfrastructure loads servers, volumes, floating IPs, snapshots and pricing.
func (c *HCloudClient) FetchInfrastructure(ctx context.Context) (model.Infrastructure, error) {
	pricing, _, err := c.client.Pricing.Get(ctx)
	if err != nil {
		return model.Infrastructure{}, fmt.Errorf("retrieve pricing: %w", err)
	}
	parsedPricing := parsePricing(pricing)

	servers, err := c.client.Server.All(ctx)
	if err != nil {
		return model.Infrastructure{}, fmt.Errorf("retrieve servers: %w", err)
	}

	volumes, err := c.client.Volume.All(ctx)
	if err != nil {
		return model.Infrastructure{}, fmt.Errorf("retrieve volumes: %w", err)
	}

	floatingIPs, err := c.client.FloatingIP.All(ctx)
	if err != nil {
		return model.Infrastructure{}, fmt.Errorf("retrieve floating IPs: %w", err)
	}

	images, err := c.client.Image.AllWithOpts(ctx, hcloud.ImageListOpts{
		Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot},
	})
	if err != nil {
		return model.Infrastructure{}, fmt.Errorf("retrieve snapshots: %w", err)
	}

	serverByID := map[int64]*hcloud.Server{}
	for _, s := range servers {
		serverByID[s.ID] = s
	}

	outServers := make([]model.Server, 0, len(servers))
	for _, s := range servers {
		ms := mapServer(s, parsedPricing)
		if s.Status == hcloud.ServerStatusRunning {
			if cpu, err := c.averageCPU(ctx, s); err == nil {
				ms.CPUPercent = cpu
			}
		}
		outServers = append(outServers, ms)
	}

	outVolumes := make([]model.Volume, 0, len(volumes))
	for _, v := range volumes {
		outVolumes = append(outVolumes, mapVolume(v, serverByID, parsedPricing))
	}

	outIPs := make([]model.FloatingIP, 0, len(floatingIPs))
	for _, ip := range floatingIPs {
		outIPs = append(outIPs, mapFloatingIP(ip, serverByID, parsedPricing))
	}

	outSnaps := make([]model.Snapshot, 0, len(images))
	for _, img := range images {
		outSnaps = append(outSnaps, mapSnapshot(img, parsedPricing))
	}

	return model.Infrastructure{
		Servers:     outServers,
		Volumes:     outVolumes,
		FloatingIPs: outIPs,
		Snapshots:   outSnaps,
		Pricing:     parsedPricing,
		FetchedAt:   time.Now(),
	}, nil
}

// RebootServer soft-reboots a server.
func (c *HCloudClient) RebootServer(ctx context.Context, serverID int64) error {
	server, _, err := c.client.Server.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("lookup server: %w", err)
	}
	if server == nil {
		return fmt.Errorf("server %d not found", serverID)
	}
	_, _, err = c.client.Server.Reboot(ctx, server)
	if err != nil {
		return fmt.Errorf("reboot server: %w", err)
	}
	return nil
}

// DeleteVolume deletes a volume by ID.
func (c *HCloudClient) DeleteVolume(ctx context.Context, volumeID int64) error {
	volume, _, err := c.client.Volume.GetByID(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("lookup volume: %w", err)
	}
	if volume == nil {
		return fmt.Errorf("volume %d not found", volumeID)
	}
	if volume.Server != nil {
		return fmt.Errorf("volume is still attached to a server; detach it first")
	}
	if volume.Protection.Delete {
		return fmt.Errorf("volume has delete protection enabled")
	}
	_, err = c.client.Volume.Delete(ctx, volume)
	if err != nil {
		return fmt.Errorf("delete volume: %w", err)
	}
	return nil
}

// DeleteFloatingIP deletes a floating IP by ID.
func (c *HCloudClient) DeleteFloatingIP(ctx context.Context, floatingIPID int64) error {
	ip, _, err := c.client.FloatingIP.GetByID(ctx, floatingIPID)
	if err != nil {
		return fmt.Errorf("lookup floating IP: %w", err)
	}
	if ip == nil {
		return fmt.Errorf("floating IP %d not found", floatingIPID)
	}
	if ip.Server != nil {
		return fmt.Errorf("floating IP is still assigned to a server; unassign it first")
	}
	if ip.Protection.Delete {
		return fmt.Errorf("floating IP has delete protection enabled")
	}
	_, err = c.client.FloatingIP.Delete(ctx, ip)
	if err != nil {
		return fmt.Errorf("delete floating IP: %w", err)
	}
	return nil
}

// CreateSnapshot creates a snapshot image of a server.
func (c *HCloudClient) CreateSnapshot(ctx context.Context, serverID int64, description string) error {
	server, _, err := c.client.Server.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("lookup server: %w", err)
	}
	if server == nil {
		return fmt.Errorf("server %d not found", serverID)
	}
	if description == "" {
		description = fmt.Sprintf("hcloud-tui snapshot %s", time.Now().UTC().Format(time.RFC3339))
	}
	_, _, err = c.client.Server.CreateImage(ctx, server, &hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.Ptr(description),
	})
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}
	return nil
}

func (c *HCloudClient) averageCPU(ctx context.Context, server *hcloud.Server) (float64, error) {
	end := time.Now()
	start := end.Add(-15 * time.Minute)
	metrics, _, err := c.client.Server.GetMetrics(ctx, server, hcloud.ServerGetMetricsOpts{
		Types: []hcloud.ServerMetricType{hcloud.ServerMetricCPU},
		Start: start,
		End:   end,
	})
	if err != nil {
		return -1, err
	}
	series, ok := metrics.TimeSeries["cpu"]
	if !ok || len(series) == 0 {
		return -1, fmt.Errorf("no cpu metrics")
	}
	sum := 0.0
	count := 0
	for _, v := range series {
		f, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			continue
		}
		sum += f
		count++
	}
	if count == 0 {
		return -1, fmt.Errorf("no parseable cpu metrics")
	}
	return sum / float64(count), nil
}

func parsePricing(p hcloud.Pricing) model.Pricing {
	out := model.Pricing{
		Currency:          p.Currency,
		ServerMonthly:     map[string]map[string]float64{},
		FloatingIPMonthly: map[string]map[string]float64{},
		PrimaryIPMonthly:  map[string]map[string]float64{},
	}
	out.VolumePerGBMonthly = parsePrice(p.Volume.PerGBMonthly.Net)
	out.ImagePerGBMonthly = parsePrice(p.Image.PerGBMonth.Net)

	for _, st := range p.ServerTypes {
		if st.ServerType == nil {
			continue
		}
		locMap := map[string]float64{}
		for _, lp := range st.Pricings {
			locName := ""
			if lp.Location != nil {
				locName = lp.Location.Name
			}
			locMap[locName] = parsePrice(lp.Monthly.Net)
		}
		out.ServerMonthly[st.ServerType.Name] = locMap
	}

	for _, ft := range p.FloatingIPs {
		locMap := map[string]float64{}
		for _, lp := range ft.Pricings {
			locName := ""
			if lp.Location != nil {
				locName = lp.Location.Name
			}
			locMap[locName] = parsePrice(lp.Monthly.Net)
		}
		out.FloatingIPMonthly[string(ft.Type)] = locMap
	}

	for _, pt := range p.PrimaryIPs {
		locMap := map[string]float64{}
		for _, lp := range pt.Pricings {
			locMap[lp.Location] = parsePrice(lp.Monthly.Net)
		}
		out.PrimaryIPMonthly[pt.Type] = locMap
	}

	return out
}

func mapServer(s *hcloud.Server, pricing model.Pricing) model.Server {
	ms := model.Server{
		ID:         s.ID,
		Name:       s.Name,
		Status:     mapStatus(s.Status),
		StatusRaw:  string(s.Status),
		CPUPercent: -1,
		Currency:   pricing.Currency,
	}
	if s.ServerType != nil {
		ms.ServerType = s.ServerType.Name
		ms.Cores = s.ServerType.Cores
		ms.MemoryGB = float64(s.ServerType.Memory)
		ms.DiskGB = s.ServerType.Disk
	}
	if s.Location != nil {
		ms.Location = s.Location.City
		if ms.Location == "" {
			ms.Location = s.Location.Name
		}
		ms.Datacenter = s.Location.Name
	}
	if !s.PublicNet.IPv4.IsUnspecified() {
		ms.IPv4 = s.PublicNet.IPv4.IP.String()
	}
	if s.PublicNet.IPv6.IP != nil {
		ms.IPv6 = s.PublicNet.IPv6.IP.String()
	}
	ms.MonthlyPrice = lookupServerPrice(pricing, ms.ServerType, ms.Datacenter)
	return ms
}

func mapVolume(v *hcloud.Volume, servers map[int64]*hcloud.Server, pricing model.Pricing) model.Volume {
	mv := model.Volume{
		ID:               v.ID,
		Name:             v.Name,
		SizeGB:           v.Size,
		Attached:         v.Server != nil,
		MonthlyPrice:     float64(v.Size) * pricing.VolumePerGBMonthly,
		Currency:         pricing.Currency,
		DeleteProtection: v.Protection.Delete,
	}
	if v.Location != nil {
		mv.Location = v.Location.Name
	}
	if v.Server != nil {
		mv.ServerID = v.Server.ID
		if full, ok := servers[v.Server.ID]; ok {
			mv.ServerName = full.Name
		} else if v.Server.Name != "" {
			mv.ServerName = v.Server.Name
		} else {
			mv.ServerName = fmt.Sprintf("#%d", v.Server.ID)
		}
	}
	return mv
}

func mapFloatingIP(ip *hcloud.FloatingIP, servers map[int64]*hcloud.Server, pricing model.Pricing) model.FloatingIP {
	mf := model.FloatingIP{
		ID:               ip.ID,
		Name:             ip.Name,
		Type:             string(ip.Type),
		Assigned:         ip.Server != nil,
		Currency:         pricing.Currency,
		DeleteProtection: ip.Protection.Delete,
	}
	if ip.IP != nil {
		mf.IP = ip.IP.String()
	}
	loc := ""
	if ip.HomeLocation != nil {
		mf.Location = ip.HomeLocation.Name
		loc = ip.HomeLocation.Name
	}
	mf.MonthlyPrice = lookupFloatingIPPrice(pricing, string(ip.Type), loc)
	if ip.Server != nil {
		mf.ServerID = ip.Server.ID
		if full, ok := servers[ip.Server.ID]; ok {
			mf.ServerName = full.Name
		} else if ip.Server.Name != "" {
			mf.ServerName = ip.Server.Name
		} else {
			mf.ServerName = fmt.Sprintf("#%d", ip.Server.ID)
		}
	}
	return mf
}

func mapSnapshot(img *hcloud.Image, pricing model.Pricing) model.Snapshot {
	name := img.Name
	if name == "" {
		name = img.Description
	}
	ms := model.Snapshot{
		ID:           img.ID,
		Name:         name,
		Description:  img.Description,
		Created:      img.Created,
		SizeGB:       float64(img.ImageSize),
		MonthlyPrice: float64(img.ImageSize) * pricing.ImagePerGBMonthly,
		Currency:     pricing.Currency,
	}
	if img.CreatedFrom != nil {
		ms.ServerName = img.CreatedFrom.Name
		if ms.ServerName == "" {
			ms.ServerName = fmt.Sprintf("#%d", img.CreatedFrom.ID)
		}
	}
	return ms
}

func mapStatus(status hcloud.ServerStatus) model.ServerStatus {
	switch status {
	case hcloud.ServerStatusRunning:
		return model.ServerStatusRunning
	case hcloud.ServerStatusOff:
		return model.ServerStatusOff
	case hcloud.ServerStatusUnknown:
		return model.ServerStatusUnknown
	case hcloud.ServerStatusInitializing, hcloud.ServerStatusStarting,
		hcloud.ServerStatusStopping, hcloud.ServerStatusMigrating,
		hcloud.ServerStatusRebuilding, hcloud.ServerStatusDeleting:
		return model.ServerStatusOther
	default:
		if strings.Contains(string(status), "error") {
			return model.ServerStatusError
		}
		return model.ServerStatusUnknown
	}
}

func lookupServerPrice(pricing model.Pricing, serverType, location string) float64 {
	byLoc, ok := pricing.ServerMonthly[serverType]
	if !ok {
		return 0
	}
	if price, ok := byLoc[location]; ok {
		return price
	}
	for _, price := range byLoc {
		return price
	}
	return 0
}

func lookupFloatingIPPrice(pricing model.Pricing, ipType, location string) float64 {
	byLoc, ok := pricing.FloatingIPMonthly[ipType]
	if !ok {
		// Fall back to any type pricing.
		for _, m := range pricing.FloatingIPMonthly {
			byLoc = m
			break
		}
	}
	if byLoc == nil {
		return 0
	}
	if price, ok := byLoc[location]; ok {
		return price
	}
	for _, price := range byLoc {
		return price
	}
	return 0
}

func parsePrice(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// FormatAPIError produces a user-facing API error message without stack traces.
func FormatAPIError(action string, err error) string {
	if err == nil {
		return ""
	}
	if hcloud.IsError(err, hcloud.ErrorCodeUnauthorized) {
		return fmt.Sprintf("Unable to %s.\nHetzner API returned:\n401 Unauthorized\nCheck that HCLOUD_TOKEN is valid.", action)
	}
	var apiErr hcloud.Error
	if errors.As(err, &apiErr) {
		return fmt.Sprintf("Unable to %s.\nHetzner API returned:\n%s (%s)", action, apiErr.Message, apiErr.Code)
	}
	return fmt.Sprintf("Unable to %s.\n%s", action, err.Error())
}
