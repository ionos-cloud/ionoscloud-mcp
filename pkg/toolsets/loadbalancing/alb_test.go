package loadbalancing

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

func newTestParams(t *testing.T, args map[string]interface{}) api.ToolHandlerParams {
	client := newTestClient(t)
	return api.ToolHandlerParams{
		Context:   context.Background(),
		Client:    client.Compute,
		DNSClient: client.DNS,
		Arguments: args,
	}
}

// validationParams creates params for validation-only tests that don't need real credentials.
func validationParams(args map[string]interface{}) api.ToolHandlerParams {
	return api.ToolHandlerParams{
		Context:   context.Background(),
		Arguments: args,
	}
}

// --- ALB List/Get validation tests ---

func TestListApplicationLoadBalancers_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{})
	_, err := listApplicationLoadBalancers(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestListApplicationLoadBalancers_EmptyDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "",
	})
	_, err := listApplicationLoadBalancers(context.Background(), params)
	if err == nil {
		t.Error("Expected error for empty datacenter_id")
	}
}

func TestGetApplicationLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id": "some-alb-id",
	})
	_, err := getApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestGetApplicationLoadBalancer_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
	})
	_, err := getApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestGetApplicationLoadBalancer_NonExistentIDs(t *testing.T) {
	skipIfNoCredentials(t)
	params := newTestParams(t, map[string]interface{}{
		"datacenter_id": "00000000-0000-0000-0000-000000000000",
		"alb_id":        "00000000-0000-0000-0000-000000000000",
	})
	_, err := getApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for non-existent ALB")
	}
}

// --- ALB Create validation tests ---

func TestCreateApplicationLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":         "test-alb",
		"listener_lan": float64(1),
		"target_lan":   float64(2),
	})
	_, err := createApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestCreateApplicationLoadBalancer_MissingName(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"listener_lan":  float64(1),
		"target_lan":    float64(2),
	})
	_, err := createApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestCreateApplicationLoadBalancer_MissingListenerLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-alb",
		"target_lan":    float64(2),
	})
	_, err := createApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing listener_lan")
	}
}

func TestCreateApplicationLoadBalancer_MissingTargetLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-alb",
		"listener_lan":  float64(1),
	})
	_, err := createApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing target_lan")
	}
}

func TestCreateApplicationLoadBalancer_InvalidListenerLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-alb",
		"listener_lan":  float64(0),
		"target_lan":    float64(2),
	})
	_, err := createApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_lan < 1")
	}
}

func TestCreateApplicationLoadBalancer_InvalidTargetLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-alb",
		"listener_lan":  float64(1),
		"target_lan":    float64(-1),
	})
	_, err := createApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target_lan < 1")
	}
}

// --- ALB Update validation tests ---

func TestUpdateApplicationLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id": "some-alb-id",
		"name":   "updated-name",
	})
	_, err := updateApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestUpdateApplicationLoadBalancer_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "updated-name",
	})
	_, err := updateApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestUpdateApplicationLoadBalancer_NoFieldsProvided(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
	})
	_, err := updateApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error when no update fields are provided")
	}
}

func TestUpdateApplicationLoadBalancer_InvalidListenerLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"listener_lan":  float64(0),
	})
	_, err := updateApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_lan < 1")
	}
}

func TestUpdateApplicationLoadBalancer_InvalidTargetLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"target_lan":    float64(-1),
	})
	_, err := updateApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target_lan < 1")
	}
}

// --- ALB Delete validation tests ---

func TestDeleteApplicationLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id": "some-alb-id",
	})
	_, err := deleteApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestDeleteApplicationLoadBalancer_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
	})
	_, err := deleteApplicationLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

// --- ALB Forwarding Rules List/Get validation tests ---

func TestListAlbForwardingRules_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id": "some-alb-id",
	})
	_, err := listAlbForwardingRules(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestListAlbForwardingRules_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
	})
	_, err := listAlbForwardingRules(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestGetAlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id":  "some-alb-id",
		"rule_id": "some-rule-id",
	})
	_, err := getAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestGetAlbForwardingRule_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"rule_id":       "some-rule-id",
	})
	_, err := getAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestGetAlbForwardingRule_MissingRuleID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
	})
	_, err := getAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing rule_id")
	}
}

func TestGetAlbForwardingRule_NonExistentIDs(t *testing.T) {
	skipIfNoCredentials(t)
	params := newTestParams(t, map[string]interface{}{
		"datacenter_id": "00000000-0000-0000-0000-000000000000",
		"alb_id":        "00000000-0000-0000-0000-000000000000",
		"rule_id":       "00000000-0000-0000-0000-000000000000",
	})
	_, err := getAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for non-existent forwarding rule")
	}
}

// --- ALB Forwarding Rule Create validation tests ---

func TestCreateAlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"protocol":      "HTTP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestCreateAlbForwardingRule_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-rule",
		"protocol":      "HTTP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestCreateAlbForwardingRule_MissingName(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"protocol":      "HTTP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestCreateAlbForwardingRule_MissingProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing protocol")
	}
}

func TestCreateAlbForwardingRule_InvalidProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid protocol (only HTTP is valid)")
	}
}

func TestCreateAlbForwardingRule_MissingListenerIP(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"protocol":      "HTTP",
		"listener_port": float64(80),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing listener_ip")
	}
}

func TestCreateAlbForwardingRule_MissingListenerPort(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"protocol":      "HTTP",
		"listener_ip":   "10.0.0.1",
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing listener_port")
	}
}

func TestCreateAlbForwardingRule_InvalidListenerPort_Zero(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"protocol":      "HTTP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(0),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_port = 0")
	}
}

func TestCreateAlbForwardingRule_InvalidListenerPort_TooHigh(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "test-rule",
		"protocol":      "HTTP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(70000),
	})
	_, err := createAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_port > 65535")
	}
}

// --- ALB Forwarding Rule Update validation tests ---

func TestUpdateAlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id":  "some-alb-id",
		"rule_id": "some-rule-id",
		"name":    "updated-rule",
	})
	_, err := updateAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestUpdateAlbForwardingRule_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"rule_id":       "some-rule-id",
		"name":          "updated-rule",
	})
	_, err := updateAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestUpdateAlbForwardingRule_MissingRuleID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"name":          "updated-rule",
	})
	_, err := updateAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing rule_id")
	}
}

func TestUpdateAlbForwardingRule_NoFieldsProvided(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"rule_id":       "some-rule-id",
	})
	_, err := updateAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error when no update fields are provided")
	}
}

func TestUpdateAlbForwardingRule_InvalidProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"rule_id":       "some-rule-id",
		"protocol":      "INVALID",
	})
	_, err := updateAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}
}

func TestUpdateAlbForwardingRule_InvalidListenerPort(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
		"rule_id":       "some-rule-id",
		"listener_port": float64(70000),
	})
	_, err := updateAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_port > 65535")
	}
}

// --- ALB Forwarding Rule Delete validation tests ---

func TestDeleteAlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"alb_id":  "some-alb-id",
		"rule_id": "some-rule-id",
	})
	_, err := deleteAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestDeleteAlbForwardingRule_MissingAlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"rule_id":       "some-rule-id",
	})
	_, err := deleteAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing alb_id")
	}
}

func TestDeleteAlbForwardingRule_MissingRuleID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"alb_id":        "some-alb-id",
	})
	_, err := deleteAlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing rule_id")
	}
}

// --- Toolset registration tests ---

func TestToolsetRegistration(t *testing.T) {
	ts := &Toolset{}

	if ts.GetName() != "loadbalancing" {
		t.Errorf("Expected toolset name 'loadbalancing', got '%s'", ts.GetName())
	}

	if ts.GetDescription() == "" {
		t.Error("Expected non-empty toolset description")
	}

	tools := ts.GetTools()
	if len(tools) == 0 {
		t.Fatal("Expected at least one tool in loadbalancing toolset")
	}

	for _, tool := range tools {
		if tool.Tool.Name == "" {
			t.Error("Found tool with empty name")
		}
		if tool.Tool.Description == "" {
			t.Errorf("Tool %s has empty description", tool.Tool.Name)
		}
		if tool.Tool.InputSchema == nil {
			t.Errorf("Tool %s has nil InputSchema", tool.Tool.Name)
		}
		if tool.Handler == nil {
			t.Errorf("Tool %s has nil handler", tool.Tool.Name)
		}
		if tool.Tool.Annotations == nil {
			t.Errorf("Tool %s has nil annotations", tool.Tool.Name)
		}
	}
}

func TestToolsetToolNames(t *testing.T) {
	ts := &Toolset{}
	tools := ts.GetTools()

	expectedTools := map[string]bool{
		// ALB
		"list_application_load_balancers":  false,
		"get_application_load_balancer":    false,
		"create_application_load_balancer": false,
		"update_application_load_balancer": false,
		"delete_application_load_balancer": false,
		// ALB forwarding rules
		"list_alb_forwarding_rules":  false,
		"get_alb_forwarding_rule":    false,
		"create_alb_forwarding_rule": false,
		"update_alb_forwarding_rule": false,
		"delete_alb_forwarding_rule": false,
		// NLB
		"list_network_load_balancers":  false,
		"get_network_load_balancer":    false,
		"create_network_load_balancer": false,
		"update_network_load_balancer": false,
		"delete_network_load_balancer": false,
		// NLB forwarding rules
		"list_nlb_forwarding_rules":  false,
		"get_nlb_forwarding_rule":    false,
		"create_nlb_forwarding_rule": false,
		"update_nlb_forwarding_rule": false,
		"delete_nlb_forwarding_rule": false,
		// Target groups
		"list_target_groups":  false,
		"get_target_group":    false,
		"create_target_group": false,
		"update_target_group": false,
		"delete_target_group": false,
	}

	for _, tool := range tools {
		if _, ok := expectedTools[tool.Tool.Name]; ok {
			expectedTools[tool.Tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool %s not found in toolset", name)
		}
	}
}

func TestToolsetToolCount(t *testing.T) {
	ts := &Toolset{}
	tools := ts.GetTools()
	// 10 ALB + 10 NLB + 5 Target Groups = 25 tools
	if len(tools) < 25 {
		t.Errorf("Expected at least 25 tools, got %d", len(tools))
	}
}

func TestToolAnnotations(t *testing.T) {
	ts := &Toolset{}
	tools := ts.GetTools()

	readOnlyTools := map[string]bool{
		"list_application_load_balancers": true,
		"get_application_load_balancer":   true,
		"list_alb_forwarding_rules":       true,
		"get_alb_forwarding_rule":         true,
		"list_network_load_balancers":     true,
		"get_network_load_balancer":       true,
		"list_nlb_forwarding_rules":       true,
		"get_nlb_forwarding_rule":         true,
		"list_target_groups":              true,
		"get_target_group":                true,
	}

	destructiveTools := map[string]bool{
		"delete_application_load_balancer": true,
		"delete_alb_forwarding_rule":       true,
		"delete_network_load_balancer":     true,
		"delete_nlb_forwarding_rule":       true,
		"delete_target_group":              true,
	}

	for _, tool := range tools {
		if readOnlyTools[tool.Tool.Name] {
			if tool.Tool.Annotations == nil || tool.Tool.Annotations.ReadOnlyHint == nil || !*tool.Tool.Annotations.ReadOnlyHint {
				t.Errorf("Tool %s should be marked as read-only", tool.Tool.Name)
			}
		}
		if destructiveTools[tool.Tool.Name] {
			if tool.Tool.Annotations == nil || tool.Tool.Annotations.DestructiveHint == nil || !*tool.Tool.Annotations.DestructiveHint {
				t.Errorf("Tool %s should be marked as destructive", tool.Tool.Name)
			}
		}
	}
}
