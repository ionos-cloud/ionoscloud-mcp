package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestK8sToolEndpoints(t *testing.T) {
	h := setup(t)

	cluster := "k8s-1"
	np := "np-1"

	tests := []toolTest{
		{"list_k8s_clusters", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/k8s"}},
		{"get_k8s_cluster", map[string]any{"k8s_cluster_id": cluster}, []string{"GET"}, []string{"/cloudapi/v6/k8s/" + cluster}},
		{"list_k8s_nodepools", map[string]any{"k8s_cluster_id": cluster}, []string{"GET"}, []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools"}},
		{"get_k8s_nodepool", map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np}, []string{"GET"}, []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools/" + np}},
		{"list_k8s_nodepool_nodes", map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np}, []string{"GET"}, []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools/" + np + "/nodes"}},
		{"get_k8s_node", map[string]any{"k8s_cluster_id": cluster, "nodepool_id": np, "node_id": "node-1"}, []string{"GET"}, []string{"/cloudapi/v6/k8s/" + cluster + "/nodepools/" + np + "/nodes/node-1"}},
		{"list_k8s_versions", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/k8s/versions"}},
		{"get_k8s_default_version", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/k8s/versions/default"}},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()

			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{
				Name:      tt.name,
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool(%q) returned protocol error: %v", tt.name, err)
			}

			reqs := h.log.allRequests()
			if len(tt.wantMethods) != len(tt.wantPaths) {
				t.Fatalf("test %q: wantMethods has %d entries, wantPaths has %d", tt.name, len(tt.wantMethods), len(tt.wantPaths))
			}
			if len(reqs) != len(tt.wantPaths) {
				t.Fatalf("CallTool(%q) made %d requests, want %d", tt.name, len(reqs), len(tt.wantPaths))
			}
			for i, req := range reqs {
				if req.Method != tt.wantMethods[i] {
					t.Errorf("CallTool(%q) request[%d] method = %q, want %q", tt.name, i, req.Method, tt.wantMethods[i])
				}
				if req.Path != tt.wantPaths[i] {
					t.Errorf("CallTool(%q) request[%d] path = %q, want %q", tt.name, i, req.Path, tt.wantPaths[i])
				}
			}
		})
	}
}
