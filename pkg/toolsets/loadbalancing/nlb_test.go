package loadbalancing

import (
	"context"
	"testing"
)

// --- NLB List/Get validation tests ---

func TestListNetworkLoadBalancers_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{})
	_, err := listNetworkLoadBalancers(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestListNetworkLoadBalancers_EmptyDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "",
	})
	_, err := listNetworkLoadBalancers(context.Background(), params)
	if err == nil {
		t.Error("Expected error for empty datacenter_id")
	}
}

func TestGetNetworkLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id": "some-nlb-id",
	})
	_, err := getNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestGetNetworkLoadBalancer_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
	})
	_, err := getNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

func TestGetNetworkLoadBalancer_NonExistentIDs(t *testing.T) {
	skipIfNoCredentials(t)
	params := newTestParams(t, map[string]interface{}{
		"datacenter_id": "00000000-0000-0000-0000-000000000000",
		"nlb_id":        "00000000-0000-0000-0000-000000000000",
	})
	_, err := getNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for non-existent NLB")
	}
}

// --- NLB Create validation tests ---

func TestCreateNetworkLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":         "test-nlb",
		"listener_lan": float64(1),
		"target_lan":   float64(2),
	})
	_, err := createNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestCreateNetworkLoadBalancer_MissingName(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"listener_lan":  float64(1),
		"target_lan":    float64(2),
	})
	_, err := createNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestCreateNetworkLoadBalancer_MissingListenerLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-nlb",
		"target_lan":    float64(2),
	})
	_, err := createNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing listener_lan")
	}
}

func TestCreateNetworkLoadBalancer_MissingTargetLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-nlb",
		"listener_lan":  float64(1),
	})
	_, err := createNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing target_lan")
	}
}

func TestCreateNetworkLoadBalancer_InvalidListenerLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-nlb",
		"listener_lan":  float64(0),
		"target_lan":    float64(2),
	})
	_, err := createNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_lan < 1")
	}
}

func TestCreateNetworkLoadBalancer_InvalidTargetLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "test-nlb",
		"listener_lan":  float64(1),
		"target_lan":    float64(-1),
	})
	_, err := createNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target_lan < 1")
	}
}

// --- NLB Update validation tests ---

func TestUpdateNetworkLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id": "some-nlb-id",
		"name":   "updated-name",
	})
	_, err := updateNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestUpdateNetworkLoadBalancer_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"name":          "updated-name",
	})
	_, err := updateNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

func TestUpdateNetworkLoadBalancer_NoFieldsProvided(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
	})
	_, err := updateNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error when no update fields are provided")
	}
}

func TestUpdateNetworkLoadBalancer_InvalidListenerLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"listener_lan":  float64(0),
	})
	_, err := updateNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_lan < 1")
	}
}

func TestUpdateNetworkLoadBalancer_InvalidTargetLan(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"target_lan":    float64(-1),
	})
	_, err := updateNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target_lan < 1")
	}
}

// --- NLB Delete validation tests ---

func TestDeleteNetworkLoadBalancer_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id": "some-nlb-id",
	})
	_, err := deleteNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestDeleteNetworkLoadBalancer_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
	})
	_, err := deleteNetworkLoadBalancer(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

// --- NLB Forwarding Rules List/Get validation tests ---

func TestListNlbForwardingRules_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id": "some-nlb-id",
	})
	_, err := listNlbForwardingRules(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestListNlbForwardingRules_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
	})
	_, err := listNlbForwardingRules(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

func TestGetNlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id":  "some-nlb-id",
		"rule_id": "some-rule-id",
	})
	_, err := getNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestGetNlbForwardingRule_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"rule_id":       "some-rule-id",
	})
	_, err := getNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

func TestGetNlbForwardingRule_MissingRuleID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
	})
	_, err := getNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing rule_id")
	}
}

// --- NLB Forwarding Rule Create validation tests ---

func TestCreateNlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestCreateNlbForwardingRule_MissingName(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestCreateNlbForwardingRule_InvalidAlgorithm(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "INVALID_ALGORITHM",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid algorithm")
	}
}

func TestCreateNlbForwardingRule_InvalidProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "INVALID",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}
}

func TestCreateNlbForwardingRule_InvalidListenerPort_Zero(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(0),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_port = 0")
	}
}

func TestCreateNlbForwardingRule_InvalidListenerPort_TooHigh(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(70000),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_port > 65535")
	}
}

func TestCreateNlbForwardingRule_InvalidTargetIP(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "not-an-ip", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid target IP")
	}
}

func TestCreateNlbForwardingRule_InvalidTargetPort(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(70000), "weight": float64(1)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid target port")
	}
}

func TestCreateNlbForwardingRule_InvalidTargetWeight(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(8080), "weight": float64(300)},
		},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for weight > 256")
	}
}

func TestCreateNlbForwardingRule_EmptyTargets(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "test-rule",
		"algorithm":     "ROUND_ROBIN",
		"protocol":      "TCP",
		"listener_ip":   "10.0.0.1",
		"listener_port": float64(80),
		"targets":       []interface{}{},
	})
	_, err := createNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for empty targets array")
	}
}

// --- NLB Forwarding Rule Update validation tests ---

func TestUpdateNlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id":  "some-nlb-id",
		"rule_id": "some-rule-id",
		"name":    "updated-rule",
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestUpdateNlbForwardingRule_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"rule_id":       "some-rule-id",
		"name":          "updated-rule",
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

func TestUpdateNlbForwardingRule_MissingRuleID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"name":          "updated-rule",
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing rule_id")
	}
}

func TestUpdateNlbForwardingRule_NoFieldsProvided(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"rule_id":       "some-rule-id",
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error when no update fields are provided")
	}
}

func TestUpdateNlbForwardingRule_InvalidAlgorithm(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"rule_id":       "some-rule-id",
		"algorithm":     "INVALID_ALGORITHM",
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid algorithm")
	}
}

func TestUpdateNlbForwardingRule_InvalidProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"rule_id":       "some-rule-id",
		"protocol":      "INVALID",
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}
}

func TestUpdateNlbForwardingRule_InvalidListenerPort(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
		"rule_id":       "some-rule-id",
		"listener_port": float64(70000),
	})
	_, err := updateNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for listener_port > 65535")
	}
}

// --- NLB Forwarding Rule Delete validation tests ---

func TestDeleteNlbForwardingRule_MissingDatacenterID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"nlb_id":  "some-nlb-id",
		"rule_id": "some-rule-id",
	})
	_, err := deleteNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}
}

func TestDeleteNlbForwardingRule_MissingNlbID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"rule_id":       "some-rule-id",
	})
	_, err := deleteNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing nlb_id")
	}
}

func TestDeleteNlbForwardingRule_MissingRuleID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"datacenter_id": "some-dc-id",
		"nlb_id":        "some-nlb-id",
	})
	_, err := deleteNlbForwardingRule(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing rule_id")
	}
}
