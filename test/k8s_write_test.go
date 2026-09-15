package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Write-tool tests for Managed Kubernetes. Both update endpoints are PUT rather
// than PATCH, so the carry-forward assertions here are load-bearing: they are what
// stops an unrelated field from being cleared by the request that changes another.

const (
	k8sPath      = "/cloudapi/v6/k8s"
	k8sClusterID = "k8s-1"
	k8sPoolID    = "np-1"
	k8sNodeID    = "node-1"
)

func k8sClusterPath() string { return k8sPath + "/" + k8sClusterID }
func k8sPoolsPath() string   { return k8sClusterPath() + "/nodepools" }
func k8sPoolPath() string    { return k8sPoolsPath() + "/" + k8sPoolID }
func k8sNodePath() string    { return k8sPoolPath() + "/nodes/" + k8sNodeID }
func k8sNodeReplace() string { return k8sNodePath() + "/replace" }

// clusterFixture is a cluster with every carry-forward-relevant field populated.
const clusterFixture = `{
  "id": "k8s-1",
  "metadata": {"state": "ACTIVE"},
  "properties": {
    "name": "prod",
    "k8sVersion": "1.31.2",
    "location": "de/fra",
    "public": true,
    "maintenanceWindow": {"dayOfTheWeek": "Sunday", "time": "02:00:00"},
    "apiSubnetAllowList": ["203.0.113.0/24"],
    "s3Buckets": [{"name": "audit-logs"}]
  }
}`

// poolFixture is a VCPU node pool with LANs, labels, annotations and taints, so a
// dropped carry-forward shows up as a missing key. autoScaling is {0,0} on purpose:
// the API rejects that shape on write, so a bad carry-forward fails here, not live.
const poolFixture = `{
  "id": "np-1",
  "metadata": {"state": "ACTIVE"},
  "properties": {
    "name": "workers",
    "datacenterId": "dc-1",
    "nodeCount": 4,
    "coresCount": 4,
    "ramSize": 4096,
    "availabilityZone": "AUTO",
    "storageType": "SSD",
    "storageSize": 100,
    "serverType": "VCPU",
    "k8sVersion": "1.31.2",
    "autoScaling": {"minNodeCount": 0, "maxNodeCount": 0},
    "maintenanceWindow": {"dayOfTheWeek": "Saturday", "time": "03:00:00"},
    "lans": [{"id": 3, "dhcp": true}],
    "labels": {"tier": "app"},
    "annotations": {"team": "sdk"},
    "taints": [{"key": "dedicated", "value": "gpu", "effect": "NoSchedule"}]
  }
}`

const nodeFixture = `{
  "id": "node-1",
  "metadata": {"state": "READY"},
  "properties": {"name": "worker-1", "k8sVersion": "1.31.2", "privateIP": "10.0.0.5", "publicIP": "198.51.100.7"}
}`

// putBody returns the sole PUT request's decoded properties object.
func putBody(t *testing.T, h *testSetup) map[string]any {
	t.Helper()
	var put recordedRequest
	found := false
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPut {
			put, found = r, true
		}
	}
	if !found {
		t.Fatalf("no PUT was issued; requests: %+v", h.log.allRequests())
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(put.Body), &body); err != nil {
		t.Fatalf("PUT body is not JSON (%v): %s", err, put.Body)
	}
	props, ok := body["properties"].(map[string]any)
	if !ok {
		t.Fatalf("PUT body has no properties object: %s", put.Body)
	}
	return props
}

// assertKeys asserts the exact key set of a request body's properties object.
func assertKeys(t *testing.T, what string, body map[string]any, want ...string) {
	t.Helper()
	wanted := map[string]bool{}
	for _, k := range want {
		wanted[k] = true
	}
	for k := range body {
		if !wanted[k] {
			t.Errorf("%s body contains unexpected field %q; body: %v", what, k, body)
		}
	}
	for k := range wanted {
		if _, ok := body[k]; !ok {
			t.Errorf("%s body is missing field %q; body: %v", what, k, body)
		}
	}
}

// ---------- cluster ----------

func TestCreateK8sClusterTwoPhase(t *testing.T) {
	h := destructiveSetup(t)

	args := map[string]any{
		"name":                  "prod",
		"k8s_version":           "1.31.2",
		"location":              "de/fra",
		"api_subnet_allow_list": []any{"203.0.113.0/24"},
		"maintenance_window":    map[string]any{"day_of_the_week": "sunday", "time": "02:00:00"},
	}
	h.resp.serve(k8sPath, `{"id":"k8s-new"}`)
	preview, res := previewThenExecute(t, h, "create_k8s_cluster", args)

	for _, want := range []string{"prod", "de/fra", "1.31.2", "203.0.113.0/24", "Sunday 02:00:00"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != k8sPath {
		t.Errorf("POST path = %s, want %s", req.Path, k8sPath)
	}
	for _, want := range []string{`"name":"prod"`, `"k8sVersion":"1.31.2"`, `"location":"de/fra"`, `"dayOfTheWeek":"Sunday"`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s: %s", want, req.Body)
		}
	}
}

// TestCreateK8sClusterWarnsOnOpenApiServer covers the one preview note that is a
// security statement rather than a convenience: with no allow list the Kubernetes
// API server accepts connections from anywhere.
func TestCreateK8sClusterWarnsOnOpenApiServer(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_k8s_cluster", map[string]any{"name": "prod"})
	if res.IsError {
		t.Fatalf("preview must not be an error: %s", resultText(res))
	}
	if !strings.Contains(resultText(res), "any source address") {
		t.Errorf("preview must warn that the API server is unrestricted:\n%s", resultText(res))
	}
}

func TestCreateK8sClusterValidation(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		wantText string
	}{
		{
			name:     "private without location",
			args:     map[string]any{"name": "prod", "public": false},
			wantText: "location is required for a private cluster",
		},
		{
			name:     "private without nat gateway ip",
			args:     map[string]any{"name": "prod", "public": false, "location": "de/fra"},
			wantText: "nat_gateway_ip is required for a private cluster",
		},
		{
			name:     "bad nat gateway ip",
			args:     map[string]any{"name": "prod", "public": false, "location": "de/fra", "nat_gateway_ip": "not-an-ip"},
			wantText: "is not an IP address",
		},
		{
			name:     "bad allow list entry",
			args:     map[string]any{"name": "prod", "api_subnet_allow_list": []any{"nonsense"}},
			wantText: "neither an IP address nor a CIDR",
		},
		{
			name:     "bad maintenance day",
			args:     map[string]any{"name": "prod", "maintenance_window": map[string]any{"day_of_the_week": "Funday", "time": "02:00:00"}},
			wantText: "is not a weekday",
		},
		{
			name:     "bad maintenance time",
			args:     map[string]any{"name": "prod", "maintenance_window": map[string]any{"day_of_the_week": "Monday", "time": "half past two"}},
			wantText: "is not a time of day",
		},
		{
			name:     "empty name",
			args:     map[string]any{"name": "   "},
			wantText: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_k8s_cluster", tt.args)
			if !res.IsError {
				t.Fatalf("expected an error, got: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantText) {
				t.Errorf("error missing %q: %s", tt.wantText, resultText(res))
			}
			assertNoMutation(t, h, "rejected create_k8s_cluster")
		})
	}
}

// TestUpdateK8sClusterCarriesFieldsForward is the regression guard for the PUT
// semantics: renaming a cluster must not drop its version, maintenance window,
// audit-log bucket or — most importantly — its API server allow list.
func TestUpdateK8sClusterCarriesFieldsForward(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sClusterPath(), clusterFixture)

	res := callTool(t, h, "update_k8s_cluster", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"name":           "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	body := putBody(t, h)
	assertKeys(t, "update_k8s_cluster PUT", body, "name", "k8sVersion", "maintenanceWindow", "apiSubnetAllowList", "s3Buckets")

	if body["name"] != "renamed" {
		t.Errorf("name = %v, want renamed", body["name"])
	}
	if body["k8sVersion"] != "1.31.2" {
		t.Errorf("k8sVersion = %v, want the carried-forward 1.31.2", body["k8sVersion"])
	}
	allow, _ := body["apiSubnetAllowList"].([]any)
	if len(allow) != 1 || allow[0] != "203.0.113.0/24" {
		t.Errorf("apiSubnetAllowList = %v, want the carried-forward [203.0.113.0/24]; losing it exposes the API server", body["apiSubnetAllowList"])
	}
}

// TestUpdateK8sClusterOmitsImmutableFields checks that a PUT never carries the
// fields the API treats as immutable, which it rejects rather than ignores.
func TestUpdateK8sClusterOmitsImmutableFields(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sClusterPath(), clusterFixture)

	res := callTool(t, h, "update_k8s_cluster", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"k8s_version":    "1.32.0",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	body := putBody(t, h)
	for _, immutable := range []string{"location", "natGatewayIp", "nodeSubnet", "public"} {
		if _, ok := body[immutable]; ok {
			t.Errorf("PUT body carries immutable field %q, which the API rejects: %v", immutable, body)
		}
	}
	if body["k8sVersion"] != "1.32.0" {
		t.Errorf("k8sVersion = %v, want 1.32.0", body["k8sVersion"])
	}
}

// TestUpdateK8sClusterEmptyListsClear covers the deliberate opposite of
// carry-forward: an explicitly empty array means "remove these".
func TestUpdateK8sClusterEmptyListsClear(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sClusterPath(), clusterFixture)

	res := callTool(t, h, "update_k8s_cluster", map[string]any{
		"k8s_cluster_id":        k8sClusterID,
		"api_subnet_allow_list": []any{},
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	body := putBody(t, h)
	allow, ok := body["apiSubnetAllowList"].([]any)
	if !ok || len(allow) != 0 {
		t.Errorf("apiSubnetAllowList = %v, want an empty array so the restriction is removed", body["apiSubnetAllowList"])
	}
}

func TestUpdateK8sClusterRejectsEmptyRequest(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_k8s_cluster", map[string]any{"k8s_cluster_id": k8sClusterID})
	if !res.IsError {
		t.Fatal("an update with no fields must be an error")
	}
	assertNoMutation(t, h, "empty update_k8s_cluster")
}

func TestDeleteK8sClusterTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	// Depth 2 populates the node pool collection, which is the blast radius.
	h.resp.serve(k8sClusterPath(), `{
	  "id": "k8s-1",
	  "metadata": {"state": "ACTIVE"},
	  "properties": {"name": "prod", "k8sVersion": "1.31.2", "location": "de/fra"},
	  "entities": {"nodepools": {"items": [
	    {"id": "np-1", "properties": {"name": "workers", "nodeCount": 3}},
	    {"id": "np-2", "properties": {"name": "gpu", "nodeCount": 2}}
	  ]}}
	}`)

	preview, res := previewThenExecute(t, h, "delete_k8s_cluster", map[string]any{"k8s_cluster_id": k8sClusterID})

	for _, want := range []string{"IRREVERSIBLE", "2 node pools", "5 worker nodes"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != k8sClusterPath() {
		t.Errorf("DELETE path = %s, want %s", req.Path, k8sClusterPath())
	}
}

func TestDeleteK8sClusterNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(k8sClusterPath(), http.StatusNotFound, `{"messages":[{"message":"not found"}]}`)

	res := callTool(t, h, "delete_k8s_cluster", map[string]any{"k8s_cluster_id": k8sClusterID})
	if !res.IsError {
		t.Fatal("deleting a missing cluster must be an error")
	}
	if !strings.Contains(resultText(res), "nothing to delete") {
		t.Errorf("error should say there is nothing to delete: %s", resultText(res))
	}
	assertNoMutation(t, h, "delete_k8s_cluster preview for a missing cluster")
}

// ---------- node pool ----------

func TestCreateK8sNodepoolTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolsPath(), `{"id":"np-new"}`)

	args := map[string]any{
		"k8s_cluster_id":    k8sClusterID,
		"name":              "workers",
		"datacenter_id":     "dc-1",
		"node_count":        2,
		"cores_count":       4,
		"ram_size":          4096,
		"availability_zone": "auto",
		"storage_type":      "ssd",
		"storage_size":      100,
		"server_type":       "vcpu",
		"labels":            map[string]any{"tier": "app"},
		"lans":              []any{map[string]any{"id": 3, "dhcp": true}},
	}
	preview, res := previewThenExecute(t, h, "create_k8s_nodepool", args)

	for _, want := range []string{"workers", "dc-1", "4096 MB", "100 GB SSD", "VCPU", "tier=app", "immutable"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != k8sPoolsPath() {
		t.Errorf("POST path = %s, want %s", req.Path, k8sPoolsPath())
	}
	// The enums are case-sensitive on the wire; the handler normalises them.
	for _, want := range []string{`"storageType":"SSD"`, `"availabilityZone":"AUTO"`, `"serverType":"VCPU"`, `"nodeCount":2`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s: %s", want, req.Body)
		}
	}
}

// manyTaints builds n distinct taints, to push past the API's per-pool limit.
func manyTaints(n int) []any {
	out := make([]any, 0, n)
	for i := range n {
		out = append(out, map[string]any{"key": fmt.Sprintf("k%d", i), "effect": "NoSchedule"})
	}
	return out
}

func TestCreateK8sNodepoolValidation(t *testing.T) {
	base := func(over map[string]any) map[string]any {
		args := map[string]any{
			"k8s_cluster_id":    k8sClusterID,
			"name":              "workers",
			"datacenter_id":     "dc-1",
			"node_count":        2,
			"cores_count":       4,
			"ram_size":          4096,
			"availability_zone": "AUTO",
			"storage_type":      "SSD",
			"storage_size":      100,
		}
		for k, v := range over {
			args[k] = v
		}
		return args
	}

	tests := []struct {
		name     string
		args     map[string]any
		wantText string
	}{
		{"ram not a multiple of 1024", base(map[string]any{"ram_size": 3000}), "multiple of 1024"},
		{"ram below the minimum", base(map[string]any{"ram_size": 1024}), "at least 2048"},
		{"zero cores", base(map[string]any{"cores_count": 0}), "cores_count must be at least 1"},
		{"bad storage type", base(map[string]any{"storage_type": "nvme"}), "use HDD or SSD"},
		{"bad availability zone", base(map[string]any{"availability_zone": "ZONE_9"}), "use AUTO, ZONE_1 or ZONE_2"},
		{"bad server type", base(map[string]any{"server_type": "shared"}), "use DedicatedCore"},
		{"zero node count", base(map[string]any{"node_count": 0}), "node_count must be at least 1"},
		{
			"autoscaling bounds inverted",
			base(map[string]any{"auto_scaling": map[string]any{"min_node_count": 5, "max_node_count": 2}}),
			">= min_node_count",
		},
		{
			"node count outside the autoscaling bounds",
			base(map[string]any{"node_count": 9, "auto_scaling": map[string]any{"min_node_count": 1, "max_node_count": 3}}),
			"must fall between the autoscaler bounds",
		},
		{
			"too few public ips for the node count",
			base(map[string]any{"public_ips": []any{"198.51.100.1", "198.51.100.2"}}),
			"needs at least 3",
		},
		{
			"too few public ips for the autoscaling maximum",
			base(map[string]any{
				"auto_scaling": map[string]any{"min_node_count": 1, "max_node_count": 5},
				"public_ips":   []any{"198.51.100.1", "198.51.100.2", "198.51.100.3"},
			}),
			"auto_scaling.max_node_count",
		},
		{"empty cpu_family", base(map[string]any{"cpu_family": "  "}), "cpu_family must not be empty"},
		{"bad lan id", base(map[string]any{"lans": []any{map[string]any{"id": 0}}}), "lans[0].id must be"},
		{
			"bad lan route network",
			base(map[string]any{"lans": []any{map[string]any{"id": 3, "routes": []any{map[string]any{"network": "10.0.0.0", "gateway_ip": "10.0.0.1"}}}}}),
			"is not a CIDR",
		},
		{
			"bad lan route gateway",
			base(map[string]any{"lans": []any{map[string]any{"id": 3, "routes": []any{map[string]any{"network": "10.0.0.0/24", "gateway_ip": "nope"}}}}}),
			"is not an IP address",
		},
		{"bad taint effect", base(map[string]any{"taints": []any{map[string]any{"key": "k", "effect": "Evict"}}}), "use NoSchedule, NoExecute or PreferNoSchedule"},
		{"empty taint key", base(map[string]any{"taints": []any{map[string]any{"key": " ", "effect": "NoSchedule"}}}), "taints[0].key is required"},
		{
			"bad taint key",
			base(map[string]any{"taints": []any{map[string]any{"key": "-nope-", "effect": "NoSchedule"}}}),
			"is not a Kubernetes label key",
		},
		{
			"bad taint value",
			base(map[string]any{"taints": []any{map[string]any{"key": "k", "value": "a b", "effect": "NoSchedule"}}}),
			"is not a Kubernetes label value",
		},
		{"more taints than the API accepts", base(map[string]any{"taints": manyTaints(51)}), "at most 50 per node pool"},
		{
			"taint key prefix with an empty label",
			base(map[string]any{"taints": []any{map[string]any{"key": "a..b/k", "effect": "NoSchedule"}}}),
			"is not a DNS subdomain",
		},
		{
			"taint key prefix with a label starting on a dash",
			base(map[string]any{"taints": []any{map[string]any{"key": "a.-b/k", "effect": "NoSchedule"}}}),
			"is not a DNS subdomain",
		},
		{
			"taint key prefix with an over-long label",
			base(map[string]any{"taints": []any{map[string]any{"key": strings.Repeat("a", 64) + ".com/k", "effect": "NoSchedule"}}}),
			"is not a DNS subdomain",
		},
		{
			// Both route fields are optional in the spec, so an entry with neither is
			// the only shape worth rejecting outright.
			"empty lan route",
			base(map[string]any{"lans": []any{map[string]any{"id": 3, "routes": []any{map[string]any{}}}}}),
			"is empty; give network, gateway_ip, or both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_k8s_nodepool", tt.args)
			if !res.IsError {
				t.Fatalf("expected an error, got: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantText) {
				t.Errorf("error missing %q: %s", tt.wantText, resultText(res))
			}
			assertNoMutation(t, h, "rejected create_k8s_nodepool")
		})
	}
}

// TestUpdateK8sNodepoolCarriesNodeCountForward guards NodeCount: a non-pointer field
// the SDK always serializes, so an update that skips it sends nodeCount=0 and drains
// every worker out of the pool.
func TestUpdateK8sNodepoolCarriesNodeCountForward(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), poolFixture)

	// A benign change to one field must not disturb anything else; this used to be a
	// rename, before the API rejected the pool name as immutable.
	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id":     k8sClusterID,
		"nodepool_id":        k8sPoolID,
		"maintenance_window": map[string]any{"day_of_the_week": "Sunday", "time": "04:00:00"},
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	body := putBody(t, h)
	if got := body["nodeCount"]; got != float64(4) {
		t.Fatalf("nodeCount = %v, want the carried-forward 4; sending 0 would drain the pool", got)
	}
	// No name: the API rejects it as immutable. No autoScaling: the pool has none, and
	// the {0,0} the API reports for that is rejected on a write.
	assertKeys(t, "update_k8s_nodepool PUT", body,
		"nodeCount", "serverType", "k8sVersion", "maintenanceWindow", "lans", "labels", "annotations", "taints")

	// A VCPU pool must not be silently converted by the constructor's
	// serverType=DedicatedCore default.
	if body["serverType"] != "VCPU" {
		t.Errorf("serverType = %v, want the carried-forward VCPU", body["serverType"])
	}
	// taints were not supplied, so the pool's own must survive the replacing PUT.
	taints, _ := body["taints"].([]any)
	if len(taints) != 1 {
		t.Errorf("taints = %v, want the carried-forward taint; losing it lets pods schedule onto reserved nodes", body["taints"])
	}
	lans, _ := body["lans"].([]any)
	if len(lans) != 1 {
		t.Errorf("lans = %v, want the carried-forward LAN", body["lans"])
	}
}

// TestUpdateK8sNodepoolOmitsImmutableHardware checks the PUT never carries the
// per-node hardware, which the API refuses to change.
func TestUpdateK8sNodepoolOmitsImmutableHardware(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), poolFixture)

	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_count":     6,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	body := putBody(t, h)
	for _, immutable := range []string{"datacenterId", "coresCount", "ramSize", "storageType", "storageSize", "availabilityZone", "cpuFamily"} {
		if _, ok := body[immutable]; ok {
			t.Errorf("PUT body carries immutable field %q, which the API rejects: %v", immutable, body)
		}
	}
	if body["nodeCount"] != float64(6) {
		t.Errorf("nodeCount = %v, want 6", body["nodeCount"])
	}
}

// TestUpdateK8sNodepoolValidatesAgainstCarriedForwardState checks that a supplied
// node count is validated against autoscaling bounds read from the existing pool,
// not only against bounds supplied in the same call.
func TestUpdateK8sNodepoolValidatesAgainstCarriedForwardState(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), `{
	  "id": "np-1",
	  "properties": {
	    "name": "workers", "datacenterId": "dc-1", "nodeCount": 2,
	    "coresCount": 4, "ramSize": 4096, "availabilityZone": "AUTO",
	    "storageType": "SSD", "storageSize": 100,
	    "autoScaling": {"minNodeCount": 1, "maxNodeCount": 3}
	  }
	}`)

	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_count":     10,
	})
	if !res.IsError {
		t.Fatal("a node count outside the pool's existing autoscaler bounds must be an error")
	}
	if !strings.Contains(resultText(res), "must fall between the autoscaler bounds") {
		t.Errorf("unexpected error text: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPut {
			t.Fatal("a rejected update must not issue a PUT")
		}
	}
}

// TestDeleteK8sNodePreviewFailsOnUnreadablePool pins fail-closed behaviour: without the
// pool we cannot check the autoscaler minimum, so no token may be minted.
func TestDeleteK8sNodePreviewFailsOnUnreadablePool(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sNodePath(), nodeFixture)
	h.resp.serveStatus(k8sPoolPath(), http.StatusInternalServerError, `{"messages":[{"message":"boom"}]}`)

	res := callTool(t, h, "delete_k8s_node", map[string]any{
		"k8s_cluster_id": k8sClusterID, "nodepool_id": k8sPoolID, "node_id": k8sNodeID,
	})
	if !res.IsError {
		t.Fatalf("an unreadable pool must fail the preview, got: %s", resultText(res))
	}
	if strings.Contains(resultText(res), "confirmation_token") {
		t.Error("no token may be minted when the pool cannot be read")
	}
}

// TestUpdateK8sNodepoolRejectsAutoScalingDisable pins a capability that doesn't ship:
// turning an autoscaler off has no valid request body (the API rejects zero bounds
// and ignores an omitted field), so the input is rejected rather than silently no-op'd.
func TestUpdateK8sNodepoolRejectsAutoScalingDisable(t *testing.T) {
	h := destructiveSetup(t)
	// A pool with a live autoscaler, so there would be something to switch off.
	h.resp.serve(k8sPoolPath(), `{
	  "id": "np-1",
	  "properties": {
	    "name": "workers", "datacenterId": "dc-1", "nodeCount": 2,
	    "coresCount": 4, "ramSize": 4096, "availabilityZone": "AUTO",
	    "storageType": "SSD", "storageSize": 100,
	    "autoScaling": {"minNodeCount": 1, "maxNodeCount": 3}
	  }
	}`)

	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"auto_scaling":   map[string]any{"min_node_count": 0, "max_node_count": 0},
	})
	if !res.IsError {
		t.Fatal("zero autoscaler bounds must be rejected, not accepted as a silent no-op")
	}
	if !strings.Contains(resultText(res), "cannot be turned off here") {
		t.Errorf("error must explain why disabling is impossible: %s", resultText(res))
	}
	// Rejected before any request goes out, so nothing is touched.
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPut {
			t.Fatal("a rejected update must not issue a PUT")
		}
	}
}

// TestUpdateK8sNodepoolChangesAutoScalingBounds is the other half: changing the bounds
// of an existing autoscaler does work, and must keep working.
func TestUpdateK8sNodepoolChangesAutoScalingBounds(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), `{
	  "id": "np-1",
	  "properties": {
	    "name": "workers", "datacenterId": "dc-1", "nodeCount": 2,
	    "coresCount": 4, "ramSize": 4096, "availabilityZone": "AUTO",
	    "storageType": "SSD", "storageSize": 100,
	    "autoScaling": {"minNodeCount": 1, "maxNodeCount": 3}
	  }
	}`)

	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"auto_scaling":   map[string]any{"min_node_count": 2, "max_node_count": 4},
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	auto, ok := putBody(t, h)["autoScaling"].(map[string]any)
	if !ok {
		t.Fatalf("autoScaling missing from body: %v", putBody(t, h))
	}
	if auto["minNodeCount"] != float64(2) || auto["maxNodeCount"] != float64(4) {
		t.Errorf("autoScaling = %v, want min 2 max 4", auto)
	}
}

// TestUpdateK8sNodepoolDropsInactiveAutoScaling guards against resending a pool's
// {0,0} autoScaling (what the API returns when no autoscaler is active), which is
// rejected on write with "autoScaling.minNodeCount must be > 0".
func TestUpdateK8sNodepoolDropsInactiveAutoScaling(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), poolFixture) // autoScaling {0,0}

	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_count":     3,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	body := putBody(t, h)
	if _, ok := body["autoScaling"]; ok {
		t.Errorf("autoScaling = %v, want the key ABSENT; resending the {0,0} the API reports for a pool without an autoscaler is rejected on write", body["autoScaling"])
	}
	if body["nodeCount"] != float64(3) {
		t.Errorf("nodeCount = %v, want 3", body["nodeCount"])
	}
}

// TestUpdateK8sNodepoolNeverSendsName guards that the immutable pool name is neither
// a tool input nor part of the PUT body; unlike the cluster, name here is optional,
// so there is nothing to carry forward.
func TestUpdateK8sNodepoolNeverSendsName(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if tool.Name != "update_k8s_nodepool" {
			continue
		}
		raw, mErr := json.Marshal(tool.InputSchema)
		if mErr != nil {
			t.Fatalf("marshalling input schema: %v", mErr)
		}
		var schema struct {
			Properties map[string]any `json:"properties"`
		}
		if uErr := json.Unmarshal(raw, &schema); uErr != nil {
			t.Fatalf("decoding input schema: %v", uErr)
		}
		if _, ok := schema.Properties["name"]; ok {
			t.Error("update_k8s_nodepool exposes a name parameter, but the API rejects the node pool name as immutable")
		}
	}

	h.resp.serve(k8sPoolPath(), poolFixture)
	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_count":     3,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	if _, ok := putBody(t, h)["name"]; ok {
		t.Error("PUT body contains name; the API rejects it as immutable")
	}
}

func TestUpdateK8sNodepoolRejectsEmptyRequest(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
	})
	if !res.IsError {
		t.Fatal("an update with no fields must be an error")
	}
	assertNoMutation(t, h, "empty update_k8s_nodepool")
}

func TestDeleteK8sNodepoolTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), poolFixture)

	preview, res := previewThenExecute(t, h, "delete_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
	})

	for _, want := range []string{"IRREVERSIBLE", "workers", "4 worker nodes"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != k8sPoolPath() {
		t.Errorf("DELETE path = %s, want %s", req.Path, k8sPoolPath())
	}
}

// ---------- node ----------

func TestRecreateK8sNodeTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sNodePath(), nodeFixture)

	preview, res := previewThenExecute(t, h, "recreate_k8s_node", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_id":        k8sNodeID,
	})

	for _, want := range []string{"RECREATE", "worker-1", "READY", "10.0.0.5"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != k8sNodeReplace() {
		t.Errorf("POST path = %s, want %s", req.Path, k8sNodeReplace())
	}
}

func TestDeleteK8sNodeTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sNodePath(), nodeFixture)
	// The preview also reads the pool, to check the delete would not breach the
	// autoscaler minimum. Give it a pool with room.
	h.resp.serve(k8sPoolPath(), nodepoolWithAutoScaling(3, 1, 5))

	preview, res := previewThenExecute(t, h, "delete_k8s_node", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_id":        k8sNodeID,
	})

	// The preview must honestly warn the pool is left short and may STAY that way if
	// the autoscaler is active — the opposite of what the tool's name suggests.
	for _, want := range []string{
		"IRREVERSIBLE", "worker-1", "one node short", "recreate_k8s_node",
		"ACTIVE autoscaler", "may STAY at the reduced size", "update_k8s_nodepool",
	} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != k8sNodePath() {
		t.Errorf("DELETE path = %s, want %s", req.Path, k8sNodePath())
	}
}

// nodepoolWithAutoScaling is a pool fixture with explicit node count and bounds, for
// the delete-node pre-flight cases.
func nodepoolWithAutoScaling(nodeCount, min, max int) string {
	return fmt.Sprintf(`{
	  "id": "np-1",
	  "properties": {
	    "name": "workers", "datacenterId": "dc-1", "nodeCount": %d,
	    "coresCount": 4, "ramSize": 4096, "availabilityZone": "AUTO",
	    "storageType": "SSD", "storageSize": 100,
	    "autoScaling": {"minNodeCount": %d, "maxNodeCount": %d}
	  }
	}`, nodeCount, min, max)
}

// TestDeleteK8sNodeBlockedByAutoScalingMinimum guards a pre-flight check: the API
// refuses to drop a pool below its autoscaler minimum, but only after a confirmation
// token has been spent — so the check must happen in preview.
func TestDeleteK8sNodeBlockedByAutoScalingMinimum(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		wantText string
		wantOK   bool
	}{
		{
			name:     "at the minimum",
			fixture:  nodepoolWithAutoScaling(2, 2, 2),
			wantText: "below its autoscaler minimum of 2",
		},
		{
			name:    "room above the minimum",
			fixture: nodepoolWithAutoScaling(2, 1, 3),
			wantOK:  true,
		},
		{
			name:     "only node in the pool",
			fixture:  nodepoolWithAutoScaling(1, 1, 3),
			wantText: "refuses to delete a pool's last node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(k8sNodePath(), nodeFixture)
			h.resp.serve(k8sPoolPath(), tt.fixture)

			res := callTool(t, h, "delete_k8s_node", map[string]any{
				"k8s_cluster_id": k8sClusterID,
				"nodepool_id":    k8sPoolID,
				"node_id":        k8sNodeID,
			})
			if tt.wantOK {
				if res.IsError {
					t.Fatalf("preview must succeed when the pool has room: %s", resultText(res))
				}
				if !strings.Contains(resultText(res), "confirmation_token") {
					t.Errorf("expected a preview with a token: %s", resultText(res))
				}
				return
			}
			if !res.IsError {
				t.Fatalf("expected a rejection, got: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantText) {
				t.Errorf("error missing %q: %s", tt.wantText, resultText(res))
			}
			// Rejected in the preview phase, so no token is spent and nothing is touched.
			assertNoMutation(t, h, "blocked delete_k8s_node")
		})
	}
}

// TestRecreateK8sNodeIgnoresAutoScalingMinimum pins the asymmetry: recreate brings the
// replacement up before draining the old node, so the count never dips and the minimum
// never applies. It must NOT inherit delete_k8s_node's pre-flight check.
func TestRecreateK8sNodeIgnoresAutoScalingMinimum(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sNodePath(), nodeFixture)
	h.resp.serve(k8sPoolPath(), nodepoolWithAutoScaling(2, 2, 2)) // pinned: would block a delete

	res := callTool(t, h, "recreate_k8s_node", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
		"node_id":        k8sNodeID,
	})
	if res.IsError {
		t.Fatalf("recreate must not be blocked by the autoscaler minimum: %s", resultText(res))
	}
	if !strings.Contains(resultText(res), "confirmation_token") {
		t.Errorf("expected a preview with a token: %s", resultText(res))
	}
}

func TestK8sNodeToolsRejectMissingIDs(t *testing.T) {
	h := destructiveSetup(t)
	for _, tool := range []string{"recreate_k8s_node", "delete_k8s_node"} {
		t.Run(tool, func(t *testing.T) {
			res := callTool(t, h, tool, map[string]any{
				"k8s_cluster_id": k8sClusterID,
				"nodepool_id":    k8sPoolID,
				"node_id":        "  ",
			})
			if !res.IsError {
				t.Fatal("a blank node_id must be an error")
			}
			if !strings.Contains(resultText(res), "node_id is required") {
				t.Errorf("unexpected error text: %s", resultText(res))
			}
		})
	}
}

// ---------- scope gating ----------

// TestK8sWriteToolsAreScopeGated pins which Kubernetes tools each scope exposes.
// recreate_k8s_node is the interesting case: it's a POST, so a method-based gate
// would misplace it in write scope, but it destroys a node and belongs with deletes.
func TestK8sWriteToolsAreScopeGated(t *testing.T) {
	reads := []string{
		"list_k8s_clusters", "get_k8s_cluster", "list_k8s_nodepools", "get_k8s_nodepool",
		"list_k8s_nodepool_nodes", "get_k8s_node", "list_k8s_versions",
		"get_k8s_default_version",
	}
	writes := []string{"create_k8s_cluster", "update_k8s_cluster", "create_k8s_nodepool", "update_k8s_nodepool"}
	destructives := []string{"delete_k8s_cluster", "delete_k8s_nodepool", "delete_k8s_node", "recreate_k8s_node"}

	tests := []struct {
		name    string
		scope   tools.Scope
		present []string
		absent  []string
	}{
		{"read only", tools.Scope{}, reads, append(append([]string{}, writes...), destructives...)},
		{"write", tools.Scope{Write: true}, append(append([]string{}, reads...), writes...), destructives},
		{
			"destructive",
			tools.Scope{Write: true, Destructive: true},
			append(append(append([]string{}, reads...), writes...), destructives...),
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupWithScope(t, tt.scope)
			names := toolNames(t, context.Background(), h)
			for _, n := range tt.present {
				if !names[n] {
					t.Errorf("scope %s: %s should be registered", tt.scope, n)
				}
			}
			for _, n := range tt.absent {
				if names[n] {
					t.Errorf("scope %s: %s must NOT be registered", tt.scope, n)
				}
			}
		})
	}
}

// TestK8sWriteToolAnnotations pins the annotations a client uses to decide whether
// to prompt before a call.
func TestK8sWriteToolAnnotations(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	want := map[string]struct {
		destructive bool
		idempotent  bool
	}{
		"create_k8s_cluster":  {destructive: false, idempotent: false},
		"update_k8s_cluster":  {destructive: false, idempotent: true},
		"delete_k8s_cluster":  {destructive: true, idempotent: true},
		"create_k8s_nodepool": {destructive: false, idempotent: false},
		"update_k8s_nodepool": {destructive: false, idempotent: true},
		"delete_k8s_nodepool": {destructive: true, idempotent: true},
		"delete_k8s_node":     {destructive: true, idempotent: true},
		// A second recreate replaces the replacement, so it is not idempotent.
		"recreate_k8s_node": {destructive: true, idempotent: false},
	}

	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		w, ok := want[tool.Name]
		if !ok {
			continue
		}
		seen[tool.Name] = true
		a := tool.Annotations
		if a == nil {
			t.Errorf("%s has no annotations", tool.Name)
			continue
		}
		if a.ReadOnlyHint {
			t.Errorf("%s ReadOnlyHint = true, want false", tool.Name)
		}
		if a.DestructiveHint == nil || *a.DestructiveHint != w.destructive {
			t.Errorf("%s DestructiveHint = %v, want %v", tool.Name, a.DestructiveHint, w.destructive)
		}
		if a.IdempotentHint != w.idempotent {
			t.Errorf("%s IdempotentHint = %v, want %v", tool.Name, a.IdempotentHint, w.idempotent)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("write tool %s was not registered at destructive scope", name)
		}
	}
}

// TestCreateK8sNodepoolSendsTaints covers the taint input the API exposed alongside
// SDK compute v2.0.7; before that the field was x-internal and no tool accepted it.
func TestCreateK8sNodepoolSendsTaints(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolsPath(), `{"id":"np-new"}`)

	preview, res := previewThenExecute(t, h, "create_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID, "name": "workers", "datacenter_id": "dc-1",
		"node_count": 2, "cores_count": 4, "ram_size": 4096,
		"availability_zone": "AUTO", "storage_type": "SSD", "storage_size": 100,
		"taints": []any{
			map[string]any{"key": "dedicated", "value": "gpu", "effect": "noschedule"},
			map[string]any{"key": "example.com/drain", "effect": "NoExecute"},
		},
	})
	// kubectl notation in the preview, so what is authorized reads like what is applied.
	for _, want := range []string{"dedicated=gpu:NoSchedule", "example.com/drain:NoExecute"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing taint %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	body := singleRequest(t, h, http.MethodPost).Body
	// The effect enum is case-sensitive on the wire; the handler normalises it.
	for _, want := range []string{`"key":"dedicated"`, `"value":"gpu"`, `"effect":"NoSchedule"`, `"effect":"NoExecute"`} {
		if !strings.Contains(body, want) {
			t.Errorf("POST body missing %s: %s", want, body)
		}
	}
	// A key-only taint must not carry an empty value, which would read back differently.
	if strings.Contains(body, `"value":""`) {
		t.Errorf("POST body sends an empty taint value: %s", body)
	}
}

// TestCreateK8sNodepoolTokenIsBoundToTheTaints guards the confirmation target: taints
// decide what the pool will accept, so a token previewed with one set must not create a
// pool with another. Only a test that swaps them between the phases catches a narrow target.
func TestCreateK8sNodepoolTokenIsBoundToTheTaints(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolsPath(), `{"id":"np-new"}`)

	args := func(taints []any) map[string]any {
		return map[string]any{
			"k8s_cluster_id": k8sClusterID, "name": "workers", "datacenter_id": "dc-1",
			"node_count": 2, "cores_count": 4, "ram_size": 4096,
			"availability_zone": "AUTO", "storage_type": "SSD", "storage_size": 100,
			"taints": taints,
		}
	}
	benign := []any{map[string]any{"key": "dedicated", "value": "gpu", "effect": "PreferNoSchedule"}}

	res := callTool(t, h, "create_k8s_nodepool", args(benign))
	if res.IsError {
		t.Fatalf("preview failed: %s", resultText(res))
	}
	token := extractToken(t, resultText(res))
	h.log.clear()

	// Same pool, a harsher taint than the one previewed.
	swapped := args([]any{map[string]any{"key": "dedicated", "value": "gpu", "effect": "NoExecute"}})
	swapped["confirmation_token"] = token
	if res = callTool(t, h, "create_k8s_nodepool", swapped); !res.IsError {
		t.Error("a token previewed with PreferNoSchedule must not create a NoExecute pool")
	}
	assertNoMutation(t, h, "create_k8s_nodepool with swapped taints")

	// Reordering the same taints is not a change, so the token must still work.
	reordered := args([]any{
		map[string]any{"key": "spot", "effect": "NoSchedule"},
		map[string]any{"key": "dedicated", "value": "gpu", "effect": "PreferNoSchedule"},
	})
	if res = callTool(t, h, "create_k8s_nodepool", reordered); res.IsError {
		t.Fatalf("preview failed: %s", resultText(res))
	}
	token = extractToken(t, resultText(res))
	h.log.clear()
	reordered = args([]any{
		map[string]any{"key": "dedicated", "value": "gpu", "effect": "PreferNoSchedule"},
		map[string]any{"key": "spot", "effect": "NoSchedule"},
	})
	reordered["confirmation_token"] = token
	if res = callTool(t, h, "create_k8s_nodepool", reordered); res.IsError {
		t.Errorf("reordering the same taints must not invalidate the token: %s", resultText(res))
	}
}

// TestCreateK8sNodepoolTokenIsBoundToTheWholeConfiguration covers the rest of the
// payload: every field in the preview costs money, is immutable, or decides what the
// pool accepts, so none of them may change between the two phases.
func TestCreateK8sNodepoolTokenIsBoundToTheWholeConfiguration(t *testing.T) {
	base := map[string]any{
		"k8s_cluster_id": k8sClusterID, "name": "workers", "datacenter_id": "dc-1",
		"node_count": 2, "cores_count": 4, "ram_size": 4096,
		"availability_zone": "AUTO", "storage_type": "SSD", "storage_size": 100,
		"labels": map[string]any{"tier": "app"},
		"lans":   []any{map[string]any{"id": 3, "dhcp": true}},
	}
	args := func(over map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		for k, v := range over {
			out[k] = v
		}
		return out
	}

	for _, tt := range []struct {
		name string
		over map[string]any
	}{
		{"more nodes than previewed", map[string]any{"node_count": 20}},
		{"bigger cores", map[string]any{"cores_count": 32}},
		{"more RAM", map[string]any{"ram_size": 65536}},
		{"bigger storage", map[string]any{"storage_size": 2000}},
		{"a different availability zone", map[string]any{"availability_zone": "ZONE_2"}},
		{"a different storage type", map[string]any{"storage_type": "HDD"}},
		{"a different server type", map[string]any{"server_type": "VCPU"}},
		{"different labels", map[string]any{"labels": map[string]any{"tier": "db"}}},
		{"a rerouted LAN", map[string]any{"lans": []any{map[string]any{"id": 9, "dhcp": true}}}},
		{"public IPs that were not previewed", map[string]any{"public_ips": []any{"203.0.113.1", "203.0.113.2", "203.0.113.3"}}},
		{"an autoscaler that was not previewed", map[string]any{"auto_scaling": map[string]any{"min_node_count": 2, "max_node_count": 40}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(k8sPoolsPath(), `{"id":"np-new"}`)

			res := callTool(t, h, "create_k8s_nodepool", args(nil))
			if res.IsError {
				t.Fatalf("preview failed: %s", resultText(res))
			}
			token := extractToken(t, resultText(res))
			h.log.clear()

			swapped := args(tt.over)
			swapped["confirmation_token"] = token
			if res = callTool(t, h, "create_k8s_nodepool", swapped); !res.IsError {
				t.Errorf("a token previewed without this change must not create it: %s", resultText(res))
			}
			assertNoMutation(t, h, "create_k8s_nodepool with a swapped field")
		})
	}

	// The unchanged configuration must still go through, or the binding is useless.
	t.Run("the previewed configuration still executes", func(t *testing.T) {
		h := destructiveSetup(t)
		h.resp.serve(k8sPoolsPath(), `{"id":"np-new"}`)
		if _, res := previewThenExecute(t, h, "create_k8s_nodepool", args(nil)); res.IsError {
			t.Fatalf("an unchanged replay must succeed: %s", resultText(res))
		}
		singleRequest(t, h, http.MethodPost)
	})
}

// TestUpdateK8sNodepoolReplacesTaints pins the replace-or-carry-forward contract the
// other list-valued fields follow: supplied replaces, empty clears, omitted keeps.
func TestUpdateK8sNodepoolReplacesTaints(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input any
		want  int
	}{
		{"supplied taints replace the current ones", []any{
			map[string]any{"key": "batch", "effect": "PreferNoSchedule"},
			map[string]any{"key": "spot", "effect": "NoExecute"},
		}, 2},
		{"an empty array removes them all", []any{}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(k8sPoolPath(), poolFixture)

			res := callTool(t, h, "update_k8s_nodepool", map[string]any{
				"k8s_cluster_id": k8sClusterID, "nodepool_id": k8sPoolID, "taints": tt.input,
			})
			if res.IsError {
				t.Fatalf("update failed: %s", resultText(res))
			}
			body := putBody(t, h)
			taints, ok := body["taints"].([]any)
			if !ok {
				t.Fatalf("PUT body has no taints array; an omitted field would keep the fixture's taint: %v", body)
			}
			if len(taints) != tt.want {
				t.Fatalf("taints = %v, want %d entries", taints, tt.want)
			}
			// nodeCount must still be carried forward alongside the taint change.
			if body["nodeCount"] != float64(4) {
				t.Errorf("nodeCount = %v, want the carried-forward 4", body["nodeCount"])
			}
		})
	}
}

// TestK8sWriteToolsDeclareAsyncBehaviour guards the async contract: every mutating
// Kubernetes endpoint answers 202, so each write tool must say so and name what to
// poll. Cluster and node pool state live on the resource; node state is its own enum.
func TestK8sWriteToolsDeclareAsyncBehaviour(t *testing.T) {
	h := destructiveSetup(t)
	wantPoller := map[string]string{
		"create_k8s_cluster":  "get_k8s_cluster",
		"update_k8s_cluster":  "get_k8s_cluster",
		"delete_k8s_cluster":  "get_k8s_cluster",
		"create_k8s_nodepool": "get_k8s_nodepool",
		"update_k8s_nodepool": "get_k8s_nodepool",
		"delete_k8s_nodepool": "get_k8s_nodepool",
		"recreate_k8s_node":   "list_k8s_nodepool_nodes",
		"delete_k8s_node":     "list_k8s_nodepool_nodes",
	}

	seen := map[string]bool{}
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		poller, ok := wantPoller[tool.Name]
		if !ok {
			continue
		}
		seen[tool.Name] = true
		if !strings.Contains(tool.Description, "Asynchronous (202)") {
			t.Errorf("%s does not declare it is asynchronous", tool.Name)
		}
		if !strings.Contains(tool.Description, poller) {
			t.Errorf("%s does not name %s as the way to follow progress", tool.Name, poller)
		}
	}
	for name := range wantPoller {
		if !seen[name] {
			t.Errorf("write tool %s was not registered at destructive scope", name)
		}
	}
}

// TestRecreateK8sNodeDeclaresExtraNodeCost pins the cost disclosure the spec calls
// out: "The node pool has an additional billable 'active' node during this process."
func TestRecreateK8sNodeDeclaresExtraNodeCost(t *testing.T) {
	h := destructiveSetup(t)
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if tool.Name != "recreate_k8s_node" {
			continue
		}
		if !strings.Contains(tool.Description, "billable") {
			t.Errorf("recreate_k8s_node must disclose the extra billable node: %s", tool.Description)
		}
		return
	}
	t.Fatal("recreate_k8s_node was not registered")
}

// TestCreateK8sNodepoolAcceptsPartialLanRoute covers the relaxed route validation:
// the spec marks both route fields optional, so one alone must be accepted.
func TestCreateK8sNodepoolAcceptsPartialLanRoute(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID, "name": "workers", "datacenter_id": "dc-1",
		"node_count": 2, "cores_count": 4, "ram_size": 4096,
		"availability_zone": "AUTO", "storage_type": "SSD", "storage_size": 100,
		"lans": []any{map[string]any{"id": 3, "routes": []any{map[string]any{"network": "10.0.0.0/24"}}}},
	})
	if res.IsError {
		t.Fatalf("a route with only a network must be accepted: %s", resultText(res))
	}
	if !strings.Contains(resultText(res), "confirmation_token") {
		t.Errorf("expected a preview with a token, got: %s", resultText(res))
	}
}

// TestK8sTokensAreBoundToTheirTarget checks a token minted for one resource cannot
// be spent on another.
func TestK8sTokensAreBoundToTheirTarget(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(k8sPoolPath(), poolFixture)

	res := callTool(t, h, "delete_k8s_nodepool", map[string]any{
		"k8s_cluster_id": k8sClusterID,
		"nodepool_id":    k8sPoolID,
	})
	if res.IsError {
		t.Fatalf("preview failed: %s", resultText(res))
	}
	token := extractToken(t, resultText(res))
	h.log.clear()

	// Same token, different node pool.
	res = callTool(t, h, "delete_k8s_nodepool", map[string]any{
		"k8s_cluster_id":     k8sClusterID,
		"nodepool_id":        "np-other",
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token minted for np-1 must not delete np-other")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatalf("a mismatched token must not issue a DELETE: %+v", r)
		}
	}

	// And it must not be replayable on its own target either, once spent.
	res = callTool(t, h, "delete_k8s_nodepool", map[string]any{
		"k8s_cluster_id":     k8sClusterID,
		"nodepool_id":        k8sPoolID,
		"confirmation_token": token,
	})
	if res.IsError {
		t.Fatalf("the token should still be valid for its own target: %s", resultText(res))
	}
	h.log.clear()
	res = callTool(t, h, "delete_k8s_nodepool", map[string]any{
		"k8s_cluster_id":     k8sClusterID,
		"nodepool_id":        k8sPoolID,
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a consumed token must not be reusable")
	}
}
