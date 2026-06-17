package test

import (
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
