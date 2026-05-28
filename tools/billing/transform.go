package billing

import (
	"math"
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
	TopN         *int32   // when set, return a flat top-N list across datacenters instead of nested data (utilization only)
}

// CompactUtilizationResponse is the compacted shape for utilization endpoints.
// When CompactOptions.TopN is set, Datacenters is omitted and TopMeters carries
// the N largest meters globally (already sorted descending by quantity).
type CompactUtilizationResponse struct {
	StartDate        string                  `json:"start_date,omitempty"`
	EndDate          string                  `json:"end_date,omitempty"`
	ContractID       string                  `json:"contract_id,omitempty"`
	MeterDefinitions map[string]string       `json:"meter_definitions,omitempty"`
	Datacenters      []CompactUtilDC         `json:"datacenters,omitempty"`
	TopMeters        []CompactUtilMeterWithDC `json:"top_meters,omitempty"`
}

// CompactUtilMeterWithDC is a flattened meter row carrying its datacenter context.
// Used only in TopN mode.
type CompactUtilMeterWithDC struct {
	DCID   string `json:"dc_id,omitempty"`
	DCName string `json:"dc_name,omitempty"`
	CompactUtilMeter
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
				Quantity:   roundQty(qty),
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

	if opts.TopN != nil && *opts.TopN > 0 {
		out.TopMeters = flattenTopN(out.Datacenters, int(*opts.TopN))
		out.Datacenters = nil
	}

	// Trim meter_definitions to keys actually present in the final output.
	// Without this, post-filter/group/top_n leaves ~all input meter_ids in the map,
	// inflating the response with descriptions for meters that aren't emitted.
	// group_by=datacenter drops MeterID (key is type+unit) → no surviving keys → empty map.
	if used := usedMeterIDs(out); len(used) > 0 {
		trimmed := make(map[string]string, len(used))
		for id := range used {
			if d, ok := defs[id]; ok {
				trimmed[id] = d
			}
		}
		if len(trimmed) > 0 {
			out.MeterDefinitions = trimmed
		}
	}
	return out
}

// usedMeterIDs returns the set of meter_ids actually present in the final emitted output.
func usedMeterIDs(out CompactUtilizationResponse) map[string]bool {
	used := map[string]bool{}
	for _, dc := range out.Datacenters {
		for _, m := range dc.Meters {
			if m.MeterID != "" {
				used[m.MeterID] = true
			}
		}
	}
	for _, m := range out.TopMeters {
		if m.MeterID != "" {
			used[m.MeterID] = true
		}
	}
	return used
}

// roundQty trims float precision noise from float32→float64 conversion.
// Six decimals preserves practical precision for billing quantities (CPU-hours, GB, etc.)
// while eliminating values like 14.933333396911621 → 14.933333.
func roundQty(f float64) float64 {
	return math.Round(f*1e6) / 1e6
}

// flattenTopN flattens nested datacenter meters into a single list, sorted
// descending by quantity, capped at n. Each row carries its source dc_id/dc_name.
func flattenTopN(dcs []CompactUtilDC, n int) []CompactUtilMeterWithDC {
	var flat []CompactUtilMeterWithDC
	for _, dc := range dcs {
		for _, m := range dc.Meters {
			flat = append(flat, CompactUtilMeterWithDC{
				DCID:             dc.ID,
				DCName:           dc.Name,
				CompactUtilMeter: m,
			})
		}
	}
	sort.SliceStable(flat, func(i, j int) bool {
		return flat[i].Quantity > flat[j].Quantity
	})
	if len(flat) > n {
		flat = flat[:n]
	}
	return flat
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
			g := *groups[k]
			g.Quantity = roundQty(g.Quantity)
			rolled = append(rolled, g)
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

	// Trim meter_definitions to keys actually present in the final output.
	used := map[string]bool{}
	for _, dc := range out.Datacenters {
		for _, m := range dc.Meters {
			if m.MeterID != "" {
				used[m.MeterID] = true
			}
		}
	}
	if len(used) > 0 {
		trimmed := make(map[string]string, len(used))
		for id := range used {
			if d, ok := defs[id]; ok {
				trimmed[id] = d
			}
		}
		if len(trimmed) > 0 {
			out.MeterDefinitions = trimmed
		}
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
