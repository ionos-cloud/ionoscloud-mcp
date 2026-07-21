//go:build e2e_live

// Package live runs the MCP server against the REAL IONOS API. It is gated
// behind the `e2e_live` build tag AND the presence of IONOS_TOKEN, so it never
// runs in the default suite. It is read-only and discovery-based: each chain
// lists a resource and, only if the account has one, drills into it — so the
// suite stays green on an empty or freshly-reset account. It asserts response
// shape (parses, not an error), never specific values.
//
// Run with: make test-live   (requires IONOS_TOKEN; object storage chains also
// require IONOS_S3_ACCESS_KEY / IONOS_S3_SECRET_KEY).
package live

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	activitylogSDK "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/activitylog"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	k8s "github.com/ionos-cloud/ionoscloud-mcp/tools/k8s"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
)

func requireToken(t *testing.T) {
	t.Helper()
	if os.Getenv("IONOS_TOKEN") == "" {
		t.Skip("IONOS_TOKEN not set; skipping live API suite")
	}
}

func hasS3Creds() bool {
	return os.Getenv("IONOS_S3_ACCESS_KEY") != "" && os.Getenv("IONOS_S3_SECRET_KEY") != ""
}

// session builds the full MCP server backed by real IONOS SDK clients and
// returns an in-memory client session.
func session(t *testing.T) *mcp.ClientSession {
	t.Helper()

	cfg := shared.NewConfigurationFromEnv()

	server := mcp.NewServer(&mcp.Implementation{Name: "ionos-cloud-mcp", Version: "live-test"}, nil)
	activitylog.RegisterAll(server, activitylogSDK.NewAPIClient(cfg))
	compute.RegisterAll(server, computeSDK.NewAPIClient(cfg), tools.Scope{}, tools.NewConfirmationStore())
	dns.RegisterAll(server, dnsSDK.NewAPIClient(cfg))
	billing.RegisterAll(server, billSDK.NewAPIClient(cfg), "")
	cert.RegisterAll(server, certSDK.NewAPIClient(cfg))
	k8s.RegisterAll(server, computeSDK.NewAPIClient(cfg))
	objectstorage.RegisterAll(server, objstSDK.NewAPIClient(cfg), objmgmtSDK.NewAPIClient(cfg), cfg)

	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "live-client", Version: "1.0.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

// callOK invokes a tool, fails on a tool error, and returns the parsed JSON
// (object or array) from the text result. A nil return means the body was not
// JSON (e.g. a zone file) — callers that need structure should not rely on it.
func callOK(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: protocol error: %v", name, err)
	}
	var text string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			text += tc.Text
		}
	}
	if res.IsError {
		t.Fatalf("%s returned error: %s", name, text)
	}
	var parsed any
	if json.Unmarshal([]byte(text), &parsed) != nil {
		return nil
	}
	return parsed
}

// firstID extracts items[0].id from a parsed list response, or "" if the list
// is empty / not shaped that way. Handles the common {"items":[{"id":...}]} form.
func firstID(parsed any) string {
	m, ok := parsed.(map[string]any)
	if !ok {
		return ""
	}
	items, ok := m["items"].([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		return ""
	}
	if id, ok := first["id"].(string); ok {
		return id
	}
	return ""
}

func TestLiveCompute(t *testing.T) {
	requireToken(t)
	cs := session(t)

	// Always-answerable, account-agnostic endpoints.
	callOK(t, cs, "get_contract", map[string]any{})
	callOK(t, cs, "list_locations", map[string]any{})
	callOK(t, cs, "list_images", map[string]any{})
	callOK(t, cs, "list_templates", map[string]any{})
	callOK(t, cs, "list_snapshots", map[string]any{})
	callOK(t, cs, "list_ip_blocks", map[string]any{})

	dcs := callOK(t, cs, "list_datacenters", map[string]any{})
	dc := firstID(dcs)
	if dc == "" {
		t.Skip("no datacenters on account; skipping datacenter-scoped chains")
	}
	callOK(t, cs, "get_datacenter", map[string]any{"datacenter_id": dc})
	callOK(t, cs, "list_lans", map[string]any{"datacenter_id": dc})
	callOK(t, cs, "list_volumes", map[string]any{"datacenter_id": dc})

	servers := callOK(t, cs, "list_servers", map[string]any{"datacenter_id": dc})
	if srv := firstID(servers); srv != "" {
		callOK(t, cs, "get_server", map[string]any{"datacenter_id": dc, "server_id": srv})
		callOK(t, cs, "list_server_volumes", map[string]any{"datacenter_id": dc, "server_id": srv})
		callOK(t, cs, "list_nics", map[string]any{"datacenter_id": dc, "server_id": srv})
	}
}

func TestLiveDNS(t *testing.T) {
	requireToken(t)
	cs := session(t)

	callOK(t, cs, "get_dns_quota", map[string]any{})
	callOK(t, cs, "list_dns_secondary_zones", map[string]any{})
	callOK(t, cs, "list_dns_reverse_records", map[string]any{})

	zones := callOK(t, cs, "list_dns_zones", map[string]any{})
	zone := firstID(zones)
	if zone == "" {
		t.Skip("no DNS zones on account")
	}
	callOK(t, cs, "get_dns_zone", map[string]any{"zone_id": zone})
	callOK(t, cs, "list_dns_zone_records", map[string]any{"zone_id": zone})
	callOK(t, cs, "list_dns_zone_dnssec_keys", map[string]any{"zone_id": zone})
}

func TestLiveK8s(t *testing.T) {
	requireToken(t)
	cs := session(t)

	callOK(t, cs, "list_k8s_versions", map[string]any{})
	callOK(t, cs, "get_k8s_default_version", map[string]any{})

	clusters := callOK(t, cs, "list_k8s_clusters", map[string]any{})
	cluster := firstID(clusters)
	if cluster == "" {
		t.Skip("no k8s clusters on account")
	}
	callOK(t, cs, "get_k8s_cluster", map[string]any{"k8s_cluster_id": cluster})
	nps := callOK(t, cs, "list_k8s_nodepools", map[string]any{"k8s_cluster_id": cluster})
	if np := firstID(nps); np != "" {
		callOK(t, cs, "get_k8s_nodepool", map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np})
		callOK(t, cs, "list_k8s_nodepool_nodes", map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np})
	}
}

func TestLiveBilling(t *testing.T) {
	requireToken(t)
	cs := session(t)

	profile := callOK(t, cs, "get_billing_profile", map[string]any{})
	contract := contractFromProfile(profile)
	if contract == 0 {
		t.Skip("could not resolve contract number from billing profile")
	}
	callOK(t, cs, "list_billing_invoices", map[string]any{"contract": contract})
	callOK(t, cs, "list_billing_products", map[string]any{"contract": contract, "filter": "RAM"})
	period := time.Now().UTC().Format("2006-01")
	callOK(t, cs, "list_billing_utilization_by_period", map[string]any{"contract": contract, "period": period})
}

// contractFromProfile extracts a contract number from get_billing_profile, which
// returns {"companies":[{"contractId":"<number-as-string>", ...}]}. The billing
// tools take an int32 contract, so the string id is parsed. Returns 0 when the
// profile has no company (→ caller skips).
func contractFromProfile(parsed any) int32 {
	m, ok := parsed.(map[string]any)
	if !ok {
		return 0
	}
	companies, ok := m["companies"].([]any)
	if !ok || len(companies) == 0 {
		return 0
	}
	first, ok := companies[0].(map[string]any)
	if !ok {
		return 0
	}
	cid, ok := first["contractId"].(string)
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(cid))
	if err != nil {
		return 0
	}
	return int32(n)
}

func TestLiveCert(t *testing.T) {
	requireToken(t)
	cs := session(t)

	callOK(t, cs, "list_cert_certificates", map[string]any{})
	callOK(t, cs, "list_cert_providers", map[string]any{})
	callOK(t, cs, "list_cert_auto_certificates", map[string]any{})
}

func TestLiveActivityLog(t *testing.T) {
	requireToken(t)
	cs := session(t)

	contracts := callOK(t, cs, "list_activitylog_contracts", map[string]any{})
	contract := firstContractNumber(contracts)
	if contract == 0 {
		t.Skip("no accessible activity-log contracts")
	}
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().UTC().Format("2006-01-02")
	callOK(t, cs, "list_activitylog_events", map[string]any{
		"contract":   contract,
		"date_start": yesterday,
		"date_end":   today,
		"limit":      int32(5),
	})
}

// firstContractNumber extracts a contract number from list_activitylog_contracts,
// which returns a JSON array of {"id":<number>, "type":..., "href":...}
// (activitylog ReferenceById). Returns 0 when no contract is accessible.
func firstContractNumber(parsed any) int32 {
	arr, ok := parsed.([]any)
	if !ok || len(arr) == 0 {
		return 0
	}
	m, ok := arr[0].(map[string]any)
	if !ok {
		return 0
	}
	if n, ok := m["id"].(float64); ok {
		return int32(n)
	}
	return 0
}

func TestLiveObjectStorage(t *testing.T) {
	requireToken(t)
	if !hasS3Creds() {
		t.Skip("IONOS_S3_ACCESS_KEY/SECRET not set; skipping object storage live suite")
	}
	cs := session(t)

	callOK(t, cs, "list_object_storage_regions", map[string]any{})
	callOK(t, cs, "list_object_storage_access_keys", map[string]any{})

	buckets := callOK(t, cs, "list_object_storage_buckets", map[string]any{})
	bucket := firstBucketName(buckets)
	if bucket == "" {
		t.Skip("no object storage buckets on account")
	}
	// Exercises the regional clientCache against real regional endpoints.
	callOK(t, cs, "get_object_storage_bucket_location", map[string]any{"bucket": bucket})
	callOK(t, cs, "head_object_storage_bucket", map[string]any{"bucket": bucket})
}

func firstBucketName(parsed any) string {
	m, ok := parsed.(map[string]any)
	if !ok {
		return ""
	}
	buckets, ok := m["Buckets"].([]any)
	if !ok {
		return ""
	}
	for _, b := range buckets {
		if bm, ok := b.(map[string]any); ok {
			if name, ok := bm["Name"].(string); ok && name != "" {
				return name
			}
		}
	}
	return ""
}
