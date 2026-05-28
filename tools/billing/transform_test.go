package billing

import (
	"testing"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
)

func strp(s string) *string { return &s }
func f32p(f float32) *float32 { return &f }

// utilMeter is a test helper that builds a UtilizationMeter with the common fields populated.
func utilMeter(meterID, desc, meterType, region, resourceID, name string, qty float32, unit string) sdk.UtilizationMeter {
	rt := sdk.ResourceType(meterType)
	m := sdk.UtilizationMeter{
		MeterId:   strp(meterID),
		MeterDesc: strp(desc),
		Type:      &rt,
		Region:    strp(region),
		Name:      strp(name),
		Quantity: &sdk.UtilizationMeterQuantity{
			Quantity: f32p(qty),
			Unit:     strp(unit),
		},
	}
	if resourceID != "" {
		m.ResourceId = *sdk.NewNullableString(strp(resourceID))
	}
	return m
}

func utilResp(dcs []sdk.UtilizationDataCenter) sdk.UtilizationGet200Response {
	return sdk.UtilizationGet200Response{
		StartDate:   strp("2026-05-01"),
		EndDate:     strp("2026-05-28"),
		Metadata:    &sdk.Metadata{ContractId: strp("031909628")},
		Datacenters: dcs,
	}
}

func TestCompactUtilization_empty(t *testing.T) {
	out := CompactUtilizationGet(sdk.UtilizationGet200Response{}, CompactOptions{})
	if len(out.Datacenters) != 0 {
		t.Errorf("want 0 datacenters, got %d", len(out.Datacenters))
	}
	if out.MeterDefinitions != nil {
		t.Errorf("want nil meter_definitions on empty input, got %v", out.MeterDefinitions)
	}
}

func TestCompactUtilization_topLevelFields(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id:     strp("dc1"),
			Meters: []sdk.UtilizationMeter{utilMeter("LOGS1000", "Log Storage", "DB", "de/fra", "res-1", "", 1.5, "1G*30Days")},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{})
	if out.StartDate != "2026-05-01" || out.EndDate != "2026-05-28" {
		t.Errorf("dates lost: %+v", out)
	}
	if out.ContractID != "031909628" {
		t.Errorf("contract_id = %q, want %q", out.ContractID, "031909628")
	}
}

func TestCompactUtilization_meterDefinitionsHoisted(t *testing.T) {
	// Same meter_id appears twice — definition should be hoisted once.
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("DBMP1000", "MongoDB Playground", "DBAAS", "es/vit", "r1", "x", 1, "1hour"),
				utilMeter("DBMP1000", "MongoDB Playground", "DBAAS", "es/vit", "r2", "y", 2, "1hour"),
				utilMeter("DNSP1000", "DNS Primary", "DNS", "de/fra", "r3", "z", 3, "30Days"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{})
	if got := len(out.MeterDefinitions); got != 2 {
		t.Errorf("want 2 unique meter definitions, got %d (%v)", got, out.MeterDefinitions)
	}
	if out.MeterDefinitions["DBMP1000"] != "MongoDB Playground" {
		t.Errorf("DBMP1000 desc lost: %q", out.MeterDefinitions["DBMP1000"])
	}
	// Per-meter rows should NOT carry meterDesc anymore.
	for _, m := range out.Datacenters[0].Meters {
		_ = m // CompactUtilMeter has no MeterDesc field by design — compile-time guarantee.
	}
}

func TestCompactUtilization_dropsZeroByDefault(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("M1", "d", "DB", "r", "res1", "", 0, "u"),
				utilMeter("M2", "d", "DB", "r", "res2", "", 5, "u"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{})
	if len(out.Datacenters) != 1 {
		t.Fatalf("want 1 dc, got %d", len(out.Datacenters))
	}
	if len(out.Datacenters[0].Meters) != 1 {
		t.Errorf("want zero-quantity dropped, got %d meters", len(out.Datacenters[0].Meters))
	}
	if out.Datacenters[0].Meters[0].MeterID != "M2" {
		t.Errorf("kept the wrong meter: %v", out.Datacenters[0].Meters[0])
	}
}

func TestCompactUtilization_includeZeroKeepsAll(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("M1", "d", "DB", "r", "res1", "", 0, "u"),
				utilMeter("M2", "d", "DB", "r", "res2", "", 5, "u"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{IncludeZero: true})
	if len(out.Datacenters[0].Meters) != 2 {
		t.Errorf("include_zero=true should keep all, got %d", len(out.Datacenters[0].Meters))
	}
}

func TestCompactUtilization_emptyDatacentersDropped(t *testing.T) {
	// DC1 only has a zero-quantity meter — after filter it's empty and should drop.
	raw := utilResp([]sdk.UtilizationDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UtilizationMeter{utilMeter("M1", "d", "DB", "r", "res1", "", 0, "u")}},
		{Id: strp("dc2"), Meters: []sdk.UtilizationMeter{utilMeter("M2", "d", "DB", "r", "res2", "", 7, "u")}},
	})
	out := CompactUtilizationGet(raw, CompactOptions{})
	if len(out.Datacenters) != 1 {
		t.Fatalf("want 1 dc after empty drop, got %d", len(out.Datacenters))
	}
	if out.Datacenters[0].ID != "dc2" {
		t.Errorf("wrong dc kept: %q", out.Datacenters[0].ID)
	}
}

func TestCompactUtilization_datacenterFilter(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UtilizationMeter{utilMeter("M1", "d", "DB", "r", "res", "", 1, "u")}},
		{Id: strp("dc2"), Meters: []sdk.UtilizationMeter{utilMeter("M2", "d", "DB", "r", "res", "", 1, "u")}},
	})
	want := "dc2"
	out := CompactUtilizationGet(raw, CompactOptions{DatacenterID: &want})
	if len(out.Datacenters) != 1 || out.Datacenters[0].ID != "dc2" {
		t.Errorf("datacenter filter failed: %+v", out.Datacenters)
	}
}

func TestCompactUtilization_meterTypeFilter(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("M1", "d", "DBAAS", "r", "res1", "", 1, "u"),
				utilMeter("M2", "d", "DNS", "r", "res2", "", 1, "u"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{MeterTypes: []string{"DNS"}})
	if len(out.Datacenters[0].Meters) != 1 || out.Datacenters[0].Meters[0].MeterID != "M2" {
		t.Errorf("meter_types filter failed: %+v", out.Datacenters[0].Meters)
	}
}

func TestCompactUtilization_regionFilter(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("M1", "d", "DB", "de/fra", "res1", "", 1, "u"),
				utilMeter("M2", "d", "DB", "es/vit", "res2", "", 1, "u"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{Regions: []string{"es/vit"}})
	if len(out.Datacenters[0].Meters) != 1 || out.Datacenters[0].Meters[0].Region != "es/vit" {
		t.Errorf("region filter failed: %+v", out.Datacenters[0].Meters)
	}
}

func TestCompactUtilization_groupByMeter(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("M1", "d", "DB", "r", "res1", "n1", 2, "u"),
				utilMeter("M1", "d", "DB", "r", "res2", "n2", 3, "u"),
				utilMeter("M2", "d", "DB", "r", "res3", "n3", 7, "u"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{GroupBy: "meter"})
	if len(out.Datacenters[0].Meters) != 2 {
		t.Fatalf("want 2 groups after group_by=meter, got %d (%+v)", len(out.Datacenters[0].Meters), out.Datacenters[0].Meters)
	}
	// M1 should sum to 5; M2 stays 7.
	sums := map[string]float64{}
	for _, m := range out.Datacenters[0].Meters {
		sums[m.MeterID] = m.Quantity
		if m.ResourceID != "" || m.Name != "" {
			t.Errorf("group_by=meter should drop resource_id/name, got %+v", m)
		}
	}
	if sums["M1"] != 5 || sums["M2"] != 7 {
		t.Errorf("group_by=meter sums wrong: %v", sums)
	}
}

func TestCompactUtilization_groupByDatacenter(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UtilizationMeter{
				utilMeter("M1", "d", "DBAAS", "r", "res1", "n", 2, "u"),
				utilMeter("M2", "d", "DBAAS", "r", "res2", "n", 3, "u"),
				utilMeter("M3", "d", "DNS", "r", "res3", "n", 4, "u"),
			},
		},
	})
	out := CompactUtilizationGet(raw, CompactOptions{GroupBy: "datacenter"})
	if len(out.Datacenters[0].Meters) != 2 {
		t.Fatalf("want 2 type groups, got %d", len(out.Datacenters[0].Meters))
	}
	sums := map[string]float64{}
	for _, m := range out.Datacenters[0].Meters {
		sums[m.Type] = m.Quantity
	}
	if sums["DBAAS"] != 5 || sums["DNS"] != 4 {
		t.Errorf("group_by=datacenter sums wrong: %v", sums)
	}
}

func TestCompactUtilization_serverIdOmittedWhenNull(t *testing.T) {
	raw := utilResp([]sdk.UtilizationDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UtilizationMeter{utilMeter("M1", "d", "DB", "r", "res", "", 1, "u")}},
	})
	out := CompactUtilizationGet(raw, CompactOptions{})
	if out.Datacenters[0].Meters[0].ServerID != "" {
		t.Errorf("server_id should be empty when null, got %q", out.Datacenters[0].Meters[0].ServerID)
	}
}

// --- Usage tests ---

func usageMeter(meterID, desc, qty, unit string) sdk.UsageMeter {
	return sdk.UsageMeter{
		MeterId:   strp(meterID),
		MeterDesc: strp(desc),
		Quantity: &sdk.UsageMeterQuantity{
			Quantity: strp(qty),
			Unit:     strp(unit),
		},
	}
}

func usageResp(dcs []sdk.UsageDataCenter) sdk.UsageGet200Response {
	return sdk.UsageGet200Response{
		StartDate:   strp("2026-05-01"),
		EndDate:     strp("2026-05-28"),
		Metadata:    &sdk.Metadata{ContractId: strp("031909628")},
		Datacenters: dcs,
	}
}

func TestCompactUsage_dropsZeroByDefault(t *testing.T) {
	raw := usageResp([]sdk.UsageDataCenter{
		{
			Id: strp("dc1"),
			Meters: []sdk.UsageMeter{
				usageMeter("M1", "d", "0", "u"),
				usageMeter("M2", "d", "12.5", "u"),
			},
		},
	})
	out := CompactUsageGet(raw, CompactOptions{})
	if len(out.Datacenters[0].Meters) != 1 || out.Datacenters[0].Meters[0].MeterID != "M2" {
		t.Errorf("usage zero filter failed: %+v", out.Datacenters[0].Meters)
	}
}

func TestCompactUsage_includeZeroKeepsAll(t *testing.T) {
	raw := usageResp([]sdk.UsageDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UsageMeter{usageMeter("M1", "d", "0", "u"), usageMeter("M2", "d", "1", "u")}},
	})
	out := CompactUsageGet(raw, CompactOptions{IncludeZero: true})
	if len(out.Datacenters[0].Meters) != 2 {
		t.Errorf("include_zero=true should keep all, got %d", len(out.Datacenters[0].Meters))
	}
}

func TestCompactUsage_meterDefinitionsHoisted(t *testing.T) {
	raw := usageResp([]sdk.UsageDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UsageMeter{usageMeter("M1", "MongoDB", "1", "u"), usageMeter("M1", "MongoDB", "2", "u")}},
	})
	out := CompactUsageGet(raw, CompactOptions{})
	if len(out.MeterDefinitions) != 1 || out.MeterDefinitions["M1"] != "MongoDB" {
		t.Errorf("definitions hoist failed: %v", out.MeterDefinitions)
	}
}

func TestCompactUsage_datacenterFilter(t *testing.T) {
	raw := usageResp([]sdk.UsageDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UsageMeter{usageMeter("M1", "d", "1", "u")}},
		{Id: strp("dc2"), Meters: []sdk.UsageMeter{usageMeter("M2", "d", "1", "u")}},
	})
	want := "dc1"
	out := CompactUsageGet(raw, CompactOptions{DatacenterID: &want})
	if len(out.Datacenters) != 1 || out.Datacenters[0].ID != "dc1" {
		t.Errorf("usage datacenter filter failed: %+v", out.Datacenters)
	}
}

func TestCompactUsage_locationPreserved(t *testing.T) {
	raw := usageResp([]sdk.UsageDataCenter{
		{Id: strp("dc1"), Location: strp("de/fra"), Meters: []sdk.UsageMeter{usageMeter("M1", "d", "1", "u")}},
	})
	out := CompactUsageGet(raw, CompactOptions{})
	if out.Datacenters[0].Location != "de/fra" {
		t.Errorf("location lost: %q", out.Datacenters[0].Location)
	}
}

func TestCompactUsage_nonNumericQuantityKept(t *testing.T) {
	// Non-numeric quantity should not be filtered as zero — surface it for inspection.
	raw := usageResp([]sdk.UsageDataCenter{
		{Id: strp("dc1"), Meters: []sdk.UsageMeter{usageMeter("M1", "d", "N/A", "u")}},
	})
	out := CompactUsageGet(raw, CompactOptions{})
	if len(out.Datacenters[0].Meters) != 1 || out.Datacenters[0].Meters[0].Quantity != "N/A" {
		t.Errorf("non-numeric quantity should be kept, got %+v", out.Datacenters[0].Meters)
	}
}
