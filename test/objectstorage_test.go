package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestObjectStorageToolEndpoints(t *testing.T) {
	h := setup(t)

	bucket := "my-bucket"
	key := "images/photo.jpg"
	accessKeyID := "ak-1"
	region := "eu-central-3"

	tests := []toolTest{
		// Buckets
		{"list_object_storage_buckets", map[string]any{}, []string{"GET"}, []string{"/"}},
		{"get_object_storage_bucket_location", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		// forBucket resolves location on first access (GET /{bucket}?location), then HEAD /{bucket}
		{"head_object_storage_bucket", map[string]any{"bucket": bucket}, []string{"GET", "HEAD"}, []string{"/" + bucket, "/" + bucket}},

		// Bucket configuration
		{"get_object_storage_bucket_cors", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_encryption", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_lifecycle", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_policy", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_policy_status", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_replication", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_tagging", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_versioning", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_public_access_block", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_bucket_lock_configuration", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},

		// Objects
		{"list_object_storage_objects", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"head_object_storage_object", map[string]any{"bucket": bucket, "key": key}, []string{"HEAD"}, []string{"/" + bucket + "/" + key}},
		{"list_object_storage_object_versions", map[string]any{"bucket": bucket}, []string{"GET"}, []string{"/" + bucket}},
		{"get_object_storage_object_tagging", map[string]any{"bucket": bucket, "key": key}, []string{"GET"}, []string{"/" + bucket + "/" + key}},
		{"get_object_storage_object_retention", map[string]any{"bucket": bucket, "key": key}, []string{"GET"}, []string{"/" + bucket + "/" + key}},
		{"get_object_storage_object_legal_hold", map[string]any{"bucket": bucket, "key": key}, []string{"GET"}, []string{"/" + bucket + "/" + key}},

		// Access Keys
		{"list_object_storage_access_keys", map[string]any{}, []string{"GET"}, []string{"/accesskeys"}},
		{"get_object_storage_access_key", map[string]any{"access_key_id": accessKeyID}, []string{"GET"}, []string{"/accesskeys/" + accessKeyID}},

		// Regions
		{"list_object_storage_regions", map[string]any{}, []string{"GET"}, []string{"/regions"}},
		{"get_object_storage_region", map[string]any{"region": region}, []string{"GET"}, []string{"/regions/" + region}},
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
