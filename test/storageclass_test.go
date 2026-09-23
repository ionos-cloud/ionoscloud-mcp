package test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// The jsonschema tags in tools/inputs.go are literals and cannot reference the
// shared lists, so this is what keeps the two from drifting apart.
func TestStorageTypeDescriptionsListEveryValue(t *testing.T) {
	h := destructiveSetup(t)
	want := map[string][]string{
		"create_volume":       tools.VolumeStorageTypes,
		"create_server":       tools.VolumeStorageTypes,
		"create_k8s_nodepool": tools.K8sStorageTypes,
	}

	seen := map[string]bool{}
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		values, ok := want[tool.Name]
		if !ok {
			continue
		}
		seen[tool.Name] = true
		raw, mErr := json.Marshal(tool.InputSchema)
		if mErr != nil {
			t.Fatalf("marshalling %s input schema: %v", tool.Name, mErr)
		}
		for _, v := range values {
			if !strings.Contains(string(raw), v) {
				t.Errorf("%s input schema never mentions storage type %q", tool.Name, v)
			}
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("%s was not registered, so its storage types went unchecked", name)
		}
	}
}

// BALANCED is a volume-only class; the node pool tool must say so rather than
// leaving a model to infer it from an absence.
func TestK8sNodepoolSchemaSaysBalancedIsNotOffered(t *testing.T) {
	h := destructiveSetup(t)
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if tool.Name != "create_k8s_nodepool" {
			continue
		}
		raw, mErr := json.Marshal(tool.InputSchema)
		if mErr != nil {
			t.Fatalf("marshalling input schema: %v", mErr)
		}
		if !strings.Contains(string(raw), tools.StorageBalanced) {
			t.Error("create_k8s_nodepool never mentions BALANCED, so nothing tells a model it is volume-only")
		}
	}
}
