package networking

import (
	"context"
	"os"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
)

func skipIfNoCredentials(t *testing.T) {
	if os.Getenv("IONOS_USERNAME") == "" && os.Getenv("IONOS_TOKEN") == "" {
		t.Skip("Skipping E2E test: IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN not set")
	}
}

func newTestClient(t *testing.T) *ionos.Client {
	client, err := ionos.NewClient()
	if err != nil {
		t.Fatalf("Failed to create IONOS client: %v", err)
	}
	return client
}

func TestCreateNatGatewayValidation(t *testing.T) {
	skipIfNoCredentials(t)
	client := newTestClient(t)
	ctx := context.Background()

	// Test invalid public_ip type
	params := api.ToolHandlerParams{
		Context: ctx,
		Client:  client.Compute,
		Arguments: map[string]interface{}{
			"datacenter_id": "test-dc",
			"name":          "test-nat",
			"public_ips":    []interface{}{123}, // not a string
		},
	}
	_, err := createNatGateway(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid public_ip type")
	}

	// Test invalid lan type
	params = api.ToolHandlerParams{
		Context: ctx,
		Client:  client.Compute,
		Arguments: map[string]interface{}{
			"datacenter_id": "test-dc",
			"name":          "test-nat",
			"public_ips":    []interface{}{"1.2.3.4"},
			"lans":          []interface{}{"not-an-object"},
		},
	}
	_, err = createNatGateway(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid lan type")
	}
}

func TestCreateNatGatewayRuleValidation(t *testing.T) {
	skipIfNoCredentials(t)
	client := newTestClient(t)
	ctx := context.Background()

	// Test invalid protocol
	params := api.ToolHandlerParams{
		Context: ctx,
		Client:  client.Compute,
		Arguments: map[string]interface{}{
			"datacenter_id":  "test-dc",
			"nat_gateway_id": "test-nat",
			"name":           "test-rule",
			"source_subnet":  "10.0.0.0/24",
			"public_ip":      "1.2.3.4",
			"protocol":       "INVALID_PROTOCOL",
		},
	}
	_, err := createNatGatewayRule(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}

	// Test invalid rule type
	params = api.ToolHandlerParams{
		Context: ctx,
		Client:  client.Compute,
		Arguments: map[string]interface{}{
			"datacenter_id":  "test-dc",
			"nat_gateway_id": "test-nat",
			"name":           "test-rule",
			"source_subnet":  "10.0.0.0/24",
			"public_ip":      "1.2.3.4",
			"type":           "INVALID_TYPE",
		},
	}
	_, err = createNatGatewayRule(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid rule type")
	}

	// Test invalid port range
	params = api.ToolHandlerParams{
		Context: ctx,
		Client:  client.Compute,
		Arguments: map[string]interface{}{
			"datacenter_id":           "test-dc",
			"nat_gateway_id":          "test-nat",
			"name":                    "test-rule",
			"source_subnet":           "10.0.0.0/24",
			"public_ip":               "1.2.3.4",
			"target_port_range_start": float64(443),
			"target_port_range_end":   float64(80),
		},
	}
	_, err = createNatGatewayRule(ctx, params)
	if err == nil {
		t.Error("Expected error for port_range_start > port_range_end")
	}
}
