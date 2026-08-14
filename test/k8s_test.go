package test

import (
	"context"
	"net/url"
	"testing"
)

func TestK8sToolEndpoints(t *testing.T) {
	h := setup(t)

	cluster := "k8s-1"
	np := "np-1"

	d1 := url.Values{"depth": []string{"1"}}

	tests := []toolTest{
		{name: "list_k8s_clusters", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s"}, wantQuery: []url.Values{d1}},
		{name: "get_k8s_cluster", args: map[string]any{"k8s_cluster_id": cluster}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/" + cluster}},
		{name: "list_k8s_nodepools", args: map[string]any{"k8s_cluster_id": cluster}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools"}, wantQuery: []url.Values{d1}},
		{name: "get_k8s_nodepool", args: map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools/" + np}},
		{name: "list_k8s_nodepool_nodes", args: map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools/" + np + "/nodes"}, wantQuery: []url.Values{d1}},
		{name: "get_k8s_node", args: map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np, "node_id": "node-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools/" + np + "/nodes/node-1"}},
		{name: "list_k8s_versions", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/versions"}},
		{name: "get_k8s_default_version", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/k8s/versions/default"}},
	}

	h.run(t, tests)
}

// TestK8sReadToolsAreAnnotatedReadOnly pins the annotations on the Kubernetes read
// tools. Without them a client cannot tell that these tools are safe to call
// without asking, which is the whole point of the read-only hint.
func TestK8sReadToolsAreAnnotatedReadOnly(t *testing.T) {
	h := setup(t) // read-only scope: only the read tools register
	ctx := context.Background()

	want := map[string]bool{
		"list_k8s_clusters": true, "get_k8s_cluster": true,
		"list_k8s_nodepools": true, "get_k8s_nodepool": true,
		"list_k8s_nodepool_nodes": true, "get_k8s_node": true,
		"list_k8s_versions": true, "get_k8s_default_version": true,
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if !want[tool.Name] {
			continue
		}
		seen[tool.Name] = true
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations; it must carry ReadOnlyHint", tool.Name)
			continue
		}
		if !tool.Annotations.ReadOnlyHint {
			t.Errorf("%s ReadOnlyHint = false, want true", tool.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("read tool %s was not registered", name)
		}
	}
}
