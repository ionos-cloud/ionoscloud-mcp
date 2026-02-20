package loadbalancing

import (
	"context"
	"testing"
)

// --- Target Group List/Get validation tests ---

func TestGetTargetGroup_MissingTargetGroupID(t *testing.T) {
	params := validationParams(map[string]interface{}{})
	_, err := getTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing target_group_id")
	}
}

func TestGetTargetGroup_EmptyTargetGroupID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"target_group_id": "",
	})
	_, err := getTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for empty target_group_id")
	}
}

func TestGetTargetGroup_NonExistentID(t *testing.T) {
	skipIfNoCredentials(t)
	params := newTestParams(t, map[string]interface{}{
		"target_group_id": "00000000-0000-0000-0000-000000000000",
	})
	_, err := getTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for non-existent target group")
	}
}

// --- Target Group Create validation tests ---

func TestCreateTargetGroup_MissingName(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"algorithm": "ROUND_ROBIN",
		"protocol":  "HTTP",
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestCreateTargetGroup_MissingAlgorithm(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":     "test-tg",
		"protocol": "HTTP",
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing algorithm")
	}
}

func TestCreateTargetGroup_MissingProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing protocol")
	}
}

func TestCreateTargetGroup_InvalidAlgorithm(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "INVALID_ALGORITHM",
		"protocol":  "HTTP",
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid algorithm")
	}
}

func TestCreateTargetGroup_InvalidProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
		"protocol":  "INVALID",
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}
}

func TestCreateTargetGroup_InvalidTargetIP(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
		"protocol":  "HTTP",
		"targets": []interface{}{
			map[string]interface{}{"ip": "not-an-ip", "port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid target IP")
	}
}

func TestCreateTargetGroup_InvalidTargetPort(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
		"protocol":  "HTTP",
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.1", "port": float64(70000), "weight": float64(1)},
		},
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target port > 65535")
	}
}

func TestCreateTargetGroup_InvalidTargetWeight(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
		"protocol":  "HTTP",
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.1", "port": float64(8080), "weight": float64(300)},
		},
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for weight > 256")
	}
}

func TestCreateTargetGroup_TargetMissingIP(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
		"protocol":  "HTTP",
		"targets": []interface{}{
			map[string]interface{}{"port": float64(8080), "weight": float64(1)},
		},
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target missing IP")
	}
}

func TestCreateTargetGroup_TargetMissingPort(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name":      "test-tg",
		"algorithm": "ROUND_ROBIN",
		"protocol":  "HTTP",
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.1", "weight": float64(1)},
		},
	})
	_, err := createTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for target missing port")
	}
}

// --- Target Group Update validation tests ---

func TestUpdateTargetGroup_MissingTargetGroupID(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"name": "updated-tg",
	})
	_, err := updateTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing target_group_id")
	}
}

func TestUpdateTargetGroup_NoFieldsProvided(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"target_group_id": "some-tg-id",
	})
	_, err := updateTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error when no update fields are provided")
	}
}

func TestUpdateTargetGroup_InvalidAlgorithm(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"target_group_id": "some-tg-id",
		"algorithm":       "INVALID_ALGORITHM",
	})
	_, err := updateTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid algorithm")
	}
}

func TestUpdateTargetGroup_InvalidProtocol(t *testing.T) {
	params := validationParams(map[string]interface{}{
		"target_group_id": "some-tg-id",
		"protocol":        "INVALID",
	})
	_, err := updateTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}
}

// --- Target Group Delete validation tests ---

func TestDeleteTargetGroup_MissingTargetGroupID(t *testing.T) {
	params := validationParams(map[string]interface{}{})
	_, err := deleteTargetGroup(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing target_group_id")
	}
}
