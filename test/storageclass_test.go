package test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// storageTypeProperties locates the one property on each tool that carries a
// storage type. Checking the whole schema would let an unrelated description
// satisfy the guard.
var storageTypeProperties = map[string][]string{
	"create_volume":       {"type"},
	"create_server":       {"boot_volume", "type"},
	"create_k8s_nodepool": {"storage_type"},
}

func propertyDescription(t *testing.T, schema any, path []string) string {
	t.Helper()
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshalling input schema: %v", err)
	}
	node := map[string]any{}
	if err := json.Unmarshal(raw, &node); err != nil {
		t.Fatalf("decoding input schema: %v", err)
	}
	for i, key := range path {
		props, ok := node["properties"].(map[string]any)
		if !ok {
			t.Fatalf("no properties object at %v", path[:i])
		}
		child, ok := props[key].(map[string]any)
		if !ok {
			t.Fatalf("no property %q at %v", key, path[:i])
		}
		node = child
	}
	desc, _ := node["description"].(string)
	return desc
}

func isWordByte(b byte) bool {
	return b == '_' || b == '-' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// namesValue reports whether desc lists value in its own right rather than only as
// the head of a longer entry, so "SSD Standard" alone does not satisfy "SSD".
// known disambiguates: it must hold every longer value that starts with one of them.
func namesValue(desc, value string, known []string) bool {
	for i := 0; i+len(value) <= len(desc); i++ {
		if desc[i:i+len(value)] != value {
			continue
		}
		if i > 0 && isWordByte(desc[i-1]) {
			continue
		}
		if end := i + len(value); end < len(desc) && isWordByte(desc[end]) {
			continue
		}
		swallowed := false
		for _, other := range known {
			if len(other) > len(value) && strings.HasPrefix(desc[i:], other) {
				swallowed = true
				break
			}
		}
		if !swallowed {
			return true
		}
	}
	return false
}

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
		desc := propertyDescription(t, tool.InputSchema, storageTypeProperties[tool.Name])
		if desc == "" {
			t.Errorf("%s: storage type property has no description", tool.Name)
			continue
		}
		for _, v := range values {
			// VolumeStorageTypes is the superset, so it disambiguates both lists.
			if !namesValue(desc, v, tools.VolumeStorageTypes) {
				t.Errorf("%s storage type description never names %q in its own right: %s", tool.Name, v, desc)
			}
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("%s was not registered, so its storage types went unchecked", name)
		}
	}
}

// Values the node pool API refuses must be named as refused, so a model does not
// have to infer it from an absence.
func TestK8sNodepoolSchemaNamesTheRefusedValues(t *testing.T) {
	h := destructiveSetup(t)
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if tool.Name != "create_k8s_nodepool" {
			continue
		}
		desc := propertyDescription(t, tool.InputSchema, storageTypeProperties[tool.Name])
		for _, v := range []string{tools.StorageBalanced, "SSD Premium"} {
			if !strings.Contains(desc, v) {
				t.Errorf("create_k8s_nodepool never mentions %q, so nothing tells a model it is refused here", v)
			}
		}
	}
}

// namesValue is the guard the drift test leans on; a weaker check would pass even
// with an entry deleted.
func TestNamesValueRejectsAPrefixOnlyMatch(t *testing.T) {
	known := tools.VolumeStorageTypes
	full := "use ESSENTIAL, BALANCED or PERFORMANCE, or the legacy HDD, SSD, SSD Standard, SSD Premium."
	dropped := "use ESSENTIAL, BALANCED or PERFORMANCE, or the legacy HDD, SSD Standard, SSD Premium."

	if !namesValue(full, "SSD", known) {
		t.Error(`"SSD" is listed in its own right but was not found`)
	}
	if namesValue(dropped, "SSD", known) {
		t.Error(`"SSD" was matched inside "SSD Standard"/"SSD Premium" although the standalone entry is gone`)
	}
	for _, v := range []string{"SSD Standard", "SSD Premium", "ESSENTIAL", "HDD"} {
		if !namesValue(full, v, known) {
			t.Errorf("%q is listed but was not found", v)
		}
	}
	if namesValue("nothing relevant here", "HDD", known) {
		t.Error("matched HDD in a description that does not mention it")
	}
}
