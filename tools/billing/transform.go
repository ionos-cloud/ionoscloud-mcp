package billing

import (
	"sort"
	"strconv"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
)

// CompactOptions controls how the raw SDK utilization/usage response is reduced.
type CompactOptions struct {
	IncludeZero  bool     // include meters with quantity == 0 (default false)
	GroupBy      string   // "" | "meter" | "datacenter" — utilization only
	DatacenterID *string  // scope to a single datacenter (utilization + list_billing_usage)
	MeterTypes   []string // filter by meter type (utilization only)
	Regions      []string // filter by region (utilization only)
}

// CompactUtilizationResponse is the compacted shape for utilization endpoints.
type CompactUtilizationResponse struct {
	StartDate        string             `json:"start_date,omitempty"`
	EndDate          string             `json:"end_date,omitempty"`
	ContractID       string             `json:"contract_id,omitempty"`
	MeterDefinitions map[string]string  `json:"meter_definitions,omitempty"`
	Datacenters      []CompactUtilDC    `json:"datacenters"`
}

type CompactUtilDC struct {
	ID     string             `json:"id,omitempty"`
	Name   string             `json:"name,omitempty"`
	Meters []CompactUtilMeter `json:"meters"`
}

type CompactUtilMeter struct {
	MeterID    string  `json:"meter_id,omitempty"`
	Type       string  `json:"type,omitempty"`
	Region     string  `json:"region,omitempty"`
	ResourceID string  `json:"resource_id,omitempty"`
	ServerID   string  `json:"server_id,omitempty"`
	Name       string  `json:"name,omitempty"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit,omitempty"`
}

// CompactUsageResponse is the compacted shape for usage endpoints.
// UsageMeter carries fewer fields than UtilizationMeter (no type/region/resource_id/server_id/name),
// so the shape is correspondingly leaner.
type CompactUsageResponse struct {
	StartDate        string              `json:"start_date,omitempty"`
	EndDate          string              `json:"end_date,omitempty"`
	ContractID       string              `json:"contract_id,omitempty"`
	MeterDefinitions map[string]string   `json:"meter_definitions,omitempty"`
	Datacenters      []CompactUsageDC    `json:"datacenters"`
}

type CompactUsageDC struct {
	ID       string              `json:"id,omitempty"`
	Name     string              `json:"name,omitempty"`
	Location string              `json:"location,omitempty"`
	Meters   []CompactUsageMeter `json:"meters"`
}

type CompactUsageMeter struct {
	MeterID  string `json:"meter_id,omitempty"`
	Quantity string `json:"quantity,omitempty"` // string per SDK — preserves precision
	Unit     string `json:"unit,omitempty"`
}

// CompactUtilizationGet projects a UtilizationGet200Response into the compact shape.
func CompactUtilizationGet(raw sdk.UtilizationGet200Response, opts CompactOptions) CompactUtilizationResponse {
	return compactUtilization(raw.StartDate, raw.EndDate, raw.Metadata, raw.Datacenters, opts)
}

// CompactUtilizationDaily projects a UtilizationDailyFindByDate200Response into the compact shape.
func CompactUtilizationDaily(raw sdk.UtilizationDailyFindByDate200Response, opts CompactOptions) CompactUtilizationResponse {
	return compactUtilization(raw.StartDate, raw.EndDate, raw.Metadata, raw.Datacenters, opts)
}

func compactUtilization(start, end *string, meta *sdk.Metadata, dcs []sdk.UtilizationDataCenter, opts CompactOptions) CompactUtilizationResponse {
	out := CompactUtilizationResponse{
		StartDate:   deref(start),
		EndDate:     deref(end),
		Datacenters: []CompactUtilDC{},
	}
	if meta != nil {
		out.ContractID = deref(meta.ContractId)
	}

	typeSet := stringSet(opts.MeterTypes)
	regionSet := stringSet(opts.Regions)
	defs := map[string]string{}

	for _, dc := range dcs {
		dcID := deref(dc.Id)
		if opts.DatacenterID != nil && *opts.DatacenterID != dcID {
			continue
		}

		compact := CompactUtilDC{
			ID:     dcID,
			Name:   deref(dc.Name),
			Meters: []CompactUtilMeter{},
		}
		for _, m := range dc.Meters {
			typeStr := ""
			if m.Type != nil {
				typeStr = string(*m.Type)
			}
			if len(typeSet) > 0 && !typeSet[typeStr] {
				continue
			}
			region := deref(m.Region)
			if len(regionSet) > 0 && !regionSet[region] {
				continue
			}

			qty, unit := utilQuantity(m.Quantity)
			if !opts.IncludeZero && qty == 0 {
				continue
			}

			if meterID := deref(m.MeterId); meterID != "" {
				if desc := deref(m.MeterDesc); desc != "" {
					if _, seen := defs[meterID]; !seen {
						defs[meterID] = desc
					}
				}
			}

			compact.Meters = append(compact.Meters, CompactUtilMeter{
				MeterID:    deref(m.MeterId),
				Type:       typeStr,
				Region:     region,
				ResourceID: nullableString(m.ResourceId),
				ServerID:   nullableString(m.ServerId),
				Name:       deref(m.Name),
				Quantity:   qty,
				Unit:       unit,
			})
		}

		if len(compact.Meters) == 0 {
			continue
		}
		out.Datacenters = append(out.Datacenters, compact)
	}

	if opts.GroupBy == "meter" || opts.GroupBy == "datacenter" {
		out.Datacenters = applyGroupBy(out.Datacenters, opts.GroupBy)
	}

	if len(defs) > 0 {
		out.MeterDefinitions = defs
	}
	return out
}

func utilQuantity(q *sdk.UtilizationMeterQuantity) (float64, string) {
	if q == nil {
		return 0, ""
	}
	var qty float64
	if q.Quantity != nil {
		qty = float64(*q.Quantity)
	}
	return qty, deref(q.Unit)
}

// applyGroupBy collapses meters within each datacenter.
// "meter": sum quantities per (meter_id, unit). Resource/server/name dropped.
// "datacenter": sum per (type, unit). All meter_ids fold together.
func applyGroupBy(dcs []CompactUtilDC, mode string) []CompactUtilDC {
	out := make([]CompactUtilDC, 0, len(dcs))
	for _, dc := range dcs {
		groups := map[string]*CompactUtilMeter{}
		var order []string
		for _, m := range dc.Meters {
			var key string
			if mode == "meter" {
				key = m.MeterID + "|" + m.Unit
			} else {
				key = m.Type + "|" + m.Unit
			}
			if existing, ok := groups[key]; ok {
				existing.Quantity += m.Quantity
				continue
			}
			rolled := CompactUtilMeter{
				Quantity: m.Quantity,
				Unit:     m.Unit,
			}
			if mode == "meter" {
				rolled.MeterID = m.MeterID
				rolled.Type = m.Type
			} else {
				rolled.Type = m.Type
			}
			groups[key] = &rolled
			order = append(order, key)
		}
		sort.Strings(order)
		rolled := make([]CompactUtilMeter, 0, len(order))
		for _, k := range order {
			rolled = append(rolled, *groups[k])
		}
		out = append(out, CompactUtilDC{
			ID:     dc.ID,
			Name:   dc.Name,
			Meters: rolled,
		})
	}
	return out
}

// CompactUsageGet projects a UsageGet200Response into the compact shape.
// UsageMeter has fewer fields than UtilizationMeter — only meter_id, quantity, unit, desc.
// Filters that don't apply (MeterTypes, Regions, GroupBy) are ignored.
func CompactUsageGet(raw sdk.UsageGet200Response, opts CompactOptions) CompactUsageResponse {
	out := CompactUsageResponse{
		StartDate:   deref(raw.StartDate),
		EndDate:     deref(raw.EndDate),
		Datacenters: []CompactUsageDC{},
	}
	if raw.Metadata != nil {
		out.ContractID = deref(raw.Metadata.ContractId)
	}

	defs := map[string]string{}
	for _, dc := range raw.Datacenters {
		dcID := deref(dc.Id)
		if opts.DatacenterID != nil && *opts.DatacenterID != dcID {
			continue
		}

		compact := CompactUsageDC{
			ID:       dcID,
			Name:     deref(dc.Name),
			Location: deref(dc.Location),
			Meters:   []CompactUsageMeter{},
		}
		for _, m := range dc.Meters {
			qtyStr, unit := usageQuantity(m.Quantity)
			if !opts.IncludeZero && isZeroString(qtyStr) {
				continue
			}

			if meterID := deref(m.MeterId); meterID != "" {
				if desc := deref(m.MeterDesc); desc != "" {
					if _, seen := defs[meterID]; !seen {
						defs[meterID] = desc
					}
				}
			}

			compact.Meters = append(compact.Meters, CompactUsageMeter{
				MeterID:  deref(m.MeterId),
				Quantity: qtyStr,
				Unit:     unit,
			})
		}

		if len(compact.Meters) == 0 {
			continue
		}
		out.Datacenters = append(out.Datacenters, compact)
	}

	if len(defs) > 0 {
		out.MeterDefinitions = defs
	}
	return out
}

func usageQuantity(q *sdk.UsageMeterQuantity) (string, string) {
	if q == nil {
		return "", ""
	}
	return deref(q.Quantity), deref(q.Unit)
}

// isZeroString reports whether s represents a numeric zero ("0", "0.0", "0.00", "", etc.).
// Non-numeric strings are treated as non-zero so they survive the filter.
func isZeroString(s string) bool {
	if s == "" {
		return true
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return f == 0
}

func stringSet(xs []string) map[string]bool {
	if len(xs) == 0 {
		return nil
	}
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nullableString returns the value of a NullableString or "" when nil/unset.
func nullableString(n sdk.NullableString) string {
	if n.Get() == nil {
		return ""
	}
	return *n.Get()
}
