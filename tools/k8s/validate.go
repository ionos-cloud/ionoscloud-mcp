package k8s

import (
	"fmt"
	"net"
	"strings"
	"time"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Input validation for the Kubernetes write tools. Each builder returns
// (value, message); a non-empty message means reject the call with that text.
// The enums are plain strings on the wire, so a bad value would otherwise surface as
// a whole-request rejection that never names the field.

var maintenanceDays = map[string]string{
	"monday": "Monday", "tuesday": "Tuesday", "wednesday": "Wednesday",
	"thursday": "Thursday", "friday": "Friday", "saturday": "Saturday", "sunday": "Sunday",
}

func buildMaintenanceWindow(in *tools.K8sMaintenanceWindowInput) (*ionos.KubernetesMaintenanceWindow, string) {
	if in == nil {
		return nil, ""
	}
	day, ok := maintenanceDays[strings.ToLower(strings.TrimSpace(in.DayOfTheWeek))]
	if !ok {
		return nil, fmt.Sprintf("maintenance_window.day_of_the_week %q is not a weekday", in.DayOfTheWeek)
	}
	// The API takes HH:mm:ss, HH:mm:ssZ or HH:mm:ss"Z"; only the clock part is checked.
	t := strings.TrimSpace(in.Time)
	clock := strings.TrimSuffix(strings.TrimSuffix(t, `"Z"`), "Z")
	if _, err := time.Parse("15:04:05", clock); err != nil {
		return nil, fmt.Sprintf("maintenance_window.time %q is not a time of day; use HH:mm:ss", in.Time)
	}
	return ionos.NewKubernetesMaintenanceWindow(day, t), ""
}

func maintenanceWindowText(w *ionos.KubernetesMaintenanceWindow) string {
	if w == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", w.GetDayOfTheWeek(), w.GetTime())
}

var serverTypes = map[string]ionos.KubernetesNodePoolServerType{
	"dedicatedcore":  ionos.DEDICATED_CORE,
	"dedicated_core": ionos.DEDICATED_CORE,
	"dedicated":      ionos.DEDICATED_CORE,
	"vcpu":           ionos.VCPU,
}

func normalizeServerType(v string) (ionos.KubernetesNodePoolServerType, string) {
	st, ok := serverTypes[strings.ToLower(strings.TrimSpace(v))]
	if !ok {
		return "", fmt.Sprintf("server_type %q is not valid; use DedicatedCore or VCPU", v)
	}
	return st, ""
}

var storageTypes = map[string]string{"hdd": "HDD", "ssd": "SSD"}

func normalizeStorageType(v string) (string, string) {
	st, ok := storageTypes[strings.ToLower(strings.TrimSpace(v))]
	if !ok {
		return "", fmt.Sprintf("storage_type %q is not valid; use HDD or SSD", v)
	}
	return st, ""
}

var availabilityZones = map[string]string{
	"auto": "AUTO", "zone_1": "ZONE_1", "zone1": "ZONE_1", "zone_2": "ZONE_2", "zone2": "ZONE_2",
}

func normalizeAvailabilityZone(v string) (string, string) {
	az, ok := availabilityZones[strings.ToLower(strings.TrimSpace(v))]
	if !ok {
		return "", fmt.Sprintf("availability_zone %q is not valid; use AUTO, ZONE_1 or ZONE_2", v)
	}
	return az, ""
}

// validateNodeHardware checks the immutable per-node sizing. RAM must be a multiple
// of 1024 MB and at least 2048, per the spec.
func validateNodeHardware(cores, ram, storage int32) string {
	switch {
	case cores < 1:
		return fmt.Sprintf("cores_count must be at least 1, got %d", cores)
	case ram < 2048:
		return fmt.Sprintf("ram_size must be at least 2048 MB, got %d", ram)
	case ram%1024 != 0:
		return fmt.Sprintf("ram_size must be a multiple of 1024 MB, got %d", ram)
	case storage < 1:
		return fmt.Sprintf("storage_size must be at least 1 GB, got %d", storage)
	}
	return ""
}

func validateCpuFamily(v *string) string {
	if v != nil && strings.TrimSpace(*v) == "" {
		return "cpu_family must not be empty; omit it to let IONOS choose, or set server_type instead"
	}
	return ""
}

func buildLans(in []tools.K8sNodePoolLanInput) ([]ionos.KubernetesNodePoolLan, string) {
	if in == nil {
		return nil, ""
	}
	out := make([]ionos.KubernetesNodePoolLan, 0, len(in))
	for i, l := range in {
		if l.ID < 1 {
			return nil, fmt.Sprintf("lans[%d].id must be an existing LAN ID (1 or greater), got %d", i, l.ID)
		}
		lan := ionos.NewKubernetesNodePoolLan(l.ID)
		if l.Dhcp != nil {
			lan.SetDhcp(*l.Dhcp)
		}
		if l.Routes != nil {
			routes, msg := buildLanRoutes(i, l.Routes)
			if msg != "" {
				return nil, msg
			}
			lan.SetRoutes(routes)
		}
		out = append(out, *lan)
	}
	return out, ""
}

// buildLanRoutes validates each route field only when supplied: both are optional in
// the spec, so only an entry with neither is rejected.
func buildLanRoutes(lanIdx int, in []tools.K8sNodePoolLanRouteInput) ([]ionos.KubernetesNodePoolLanRoutes, string) {
	out := make([]ionos.KubernetesNodePoolLanRoutes, 0, len(in))
	for j, r := range in {
		if r.Network == nil && r.GatewayIp == nil {
			return nil, fmt.Sprintf("lans[%d].routes[%d] is empty; give network, gateway_ip, or both", lanIdx, j)
		}
		route := ionos.NewKubernetesNodePoolLanRoutes()
		if r.Network != nil {
			network := strings.TrimSpace(*r.Network)
			if _, _, err := net.ParseCIDR(network); err != nil {
				return nil, fmt.Sprintf("lans[%d].routes[%d].network %q is not a CIDR", lanIdx, j, *r.Network)
			}
			route.SetNetwork(network)
		}
		if r.GatewayIp != nil {
			gateway := strings.TrimSpace(*r.GatewayIp)
			if net.ParseIP(gateway) == nil {
				return nil, fmt.Sprintf("lans[%d].routes[%d].gateway_ip %q is not an IP address", lanIdx, j, *r.GatewayIp)
			}
			route.SetGatewayIp(gateway)
		}
		out = append(out, *route)
	}
	return out, ""
}

func lansText(lans []ionos.KubernetesNodePoolLan) string {
	parts := make([]string, 0, len(lans))
	for _, l := range lans {
		desc := fmt.Sprintf("%d", l.GetId())
		if l.HasDhcp() {
			desc += fmt.Sprintf(" (dhcp %t)", l.GetDhcp())
		}
		if n := len(l.GetRoutes()); n > 0 {
			desc += fmt.Sprintf(" +%d routes", n)
		}
		parts = append(parts, desc)
	}
	return strings.Join(parts, ", ")
}

func validateIPs(field string, vals []string) string {
	for i, v := range vals {
		if net.ParseIP(strings.TrimSpace(v)) == nil {
			return fmt.Sprintf("%s[%d] %q is not an IP address", field, i, v)
		}
	}
	return ""
}

func validateIPsOrCIDRs(field string, vals []string) string {
	for i, v := range vals {
		e := strings.TrimSpace(v)
		if net.ParseIP(e) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(e); err != nil {
			return fmt.Sprintf("%s[%d] %q is neither an IP address nor a CIDR", field, i, v)
		}
	}
	return ""
}

// buildS3Buckets takes bucket names only. IONOS writes the audit logs itself, so no
// Object Storage credentials are involved; the bucket must already exist.
func buildS3Buckets(names []string) ([]ionos.S3Bucket, string) {
	if names == nil {
		return nil, ""
	}
	out := make([]ionos.S3Bucket, 0, len(names))
	for i, n := range names {
		name := strings.TrimSpace(n)
		if name == "" {
			return nil, fmt.Sprintf("s3_buckets[%d] is empty", i)
		}
		out = append(out, *ionos.NewS3Bucket(name))
	}
	return out, ""
}

func s3BucketNames(buckets []ionos.S3Bucket) string {
	names := make([]string, 0, len(buckets))
	for _, b := range buckets {
		names = append(names, b.GetName())
	}
	return strings.Join(names, ", ")
}

// Verified live: the API answers 422 for zero bounds and ignores an omitted
// autoScaling, so nothing removes an autoscaler. Rejected rather than silently no-op'd.
const disableAutoScalingUnsupported = "auto_scaling cannot be turned off here: the API rejects zero bounds and ignores an omitted autoScaling, so no request removes an autoscaler. " +
	"For fixed-size behaviour set both bounds to the same value, or recreate the node pool without auto_scaling."

func buildAutoScaling(in *tools.K8sAutoScalingInput) (*ionos.KubernetesAutoScaling, string) {
	if in == nil {
		return nil, ""
	}
	switch {
	case in.MinNodeCount == 0 && in.MaxNodeCount == 0:
		return nil, disableAutoScalingUnsupported
	case in.MinNodeCount < 1:
		return nil, fmt.Sprintf("auto_scaling.min_node_count must be at least 1, got %d. %s", in.MinNodeCount, disableAutoScalingUnsupported)
	case in.MaxNodeCount < in.MinNodeCount:
		return nil, fmt.Sprintf("auto_scaling.max_node_count (%d) must be >= min_node_count (%d)", in.MaxNodeCount, in.MinNodeCount)
	}
	return ionos.NewKubernetesAutoScaling(in.MinNodeCount, in.MaxNodeCount), ""
}

// autoScalingActive distinguishes a live autoscaler from the {0,0} the API returns for
// a pool without one — that shape must never be written back.
func autoScalingActive(a *ionos.KubernetesAutoScaling) bool {
	return a != nil && (a.GetMinNodeCount() > 0 || a.GetMaxNodeCount() > 0)
}

// validateNodeCountWithinAutoScaling follows the Terraform provider (nodeCount >=
// minNodeCount), not the spec's minNodeCount text, which contradicts its own
// maxNodeCount text. Do not flip it without testing against a real account.
func validateNodeCountWithinAutoScaling(a *ionos.KubernetesAutoScaling, nodeCount int32) string {
	if !autoScalingActive(a) {
		return ""
	}
	if nodeCount < a.GetMinNodeCount() || nodeCount > a.GetMaxNodeCount() {
		return fmt.Sprintf("node_count (%d) must fall between the autoscaler bounds %d and %d", nodeCount, a.GetMinNodeCount(), a.GetMaxNodeCount())
	}
	return ""
}

func autoScalingText(a *ionos.KubernetesAutoScaling) string {
	if !autoScalingActive(a) {
		return ""
	}
	return fmt.Sprintf("%d-%d nodes", a.GetMinNodeCount(), a.GetMaxNodeCount())
}

// validatePublicIps enforces the spec's spare-IP rule: one more than the most nodes
// the pool can run, since the extra covers a node being rebuilt.
func validatePublicIps(ips []string, nodeCount int32, auto *ionos.KubernetesAutoScaling) string {
	if len(ips) == 0 {
		return ""
	}
	if msg := validateIPs("public_ips", ips); msg != "" {
		return msg
	}
	maxNodes, source := nodeCount, "node_count"
	if autoScalingActive(auto) {
		maxNodes, source = auto.GetMaxNodeCount(), "auto_scaling.max_node_count"
	}
	if want := int(maxNodes) + 1; len(ips) < want {
		return fmt.Sprintf("public_ips has %d addresses but needs at least %d: one more than %s (%d)", len(ips), want, source, maxNodes)
	}
	return ""
}
