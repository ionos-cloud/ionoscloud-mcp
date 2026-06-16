package test

import (
	"net/url"
	"testing"
)

func TestObjectStorageToolEndpoints(t *testing.T) {
	h := setup(t)

	bucket := "my-bucket"
	key := "images/photo.jpg"
	accessKeyID := "ak-1"
	region := "eu-central-3"

	tests := []toolTest{
		// Buckets
		{name: "list_object_storage_buckets", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/"}},
		{name: "get_object_storage_bucket_location", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		// forBucket resolves location on first access (GET /{bucket}?location), then HEAD /{bucket}.
		// This is the first forBucket call for my-bucket, so it pays the location lookup and
		// caches the region — subsequent cases below see a single request.
		{name: "head_object_storage_bucket", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET", "HEAD"}, wantPaths: []string{"/" + bucket, "/" + bucket}},

		// Bucket configuration
		{name: "get_object_storage_bucket_cors", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_encryption", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_lifecycle", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_policy", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_policy_status", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_replication", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_tagging", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_versioning", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_public_access_block", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_bucket_lock_configuration", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},

		// Objects
		{name: "list_object_storage_objects", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "head_object_storage_object", args: map[string]any{"bucket": bucket, "key": key}, wantMethods: []string{"HEAD"}, wantPaths: []string{"/" + bucket + "/" + key}},
		{name: "list_object_storage_object_versions", args: map[string]any{"bucket": bucket}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket}},
		{name: "get_object_storage_object_tagging", args: map[string]any{"bucket": bucket, "key": key}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket + "/" + key}},
		{name: "get_object_storage_object_retention", args: map[string]any{"bucket": bucket, "key": key}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket + "/" + key}},
		{name: "get_object_storage_object_legal_hold", args: map[string]any{"bucket": bucket, "key": key}, wantMethods: []string{"GET"}, wantPaths: []string{"/" + bucket + "/" + key}},

		// Access Keys
		{name: "list_object_storage_access_keys", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/accesskeys"}},
		{name: "get_object_storage_access_key", args: map[string]any{"access_key_id": accessKeyID}, wantMethods: []string{"GET"}, wantPaths: []string{"/accesskeys/" + accessKeyID}},

		// Regions
		{name: "list_object_storage_regions", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/regions"}},
		{name: "get_object_storage_region", args: map[string]any{"region": region}, wantMethods: []string{"GET"}, wantPaths: []string{"/regions/" + region}},
	}

	h.run(t, tests)
}

// TestObjectStorageListObjectsQuery asserts the prefix filter is forwarded as a
// query parameter. Uses a fresh setup so the clientCache is empty: the first
// forBucket access pays the location lookup, so the listing is request index 1.
func TestObjectStorageListObjectsQuery(t *testing.T) {
	h := setup(t)

	tests := []toolTest{
		{
			name:        "list_object_storage_objects",
			args:        map[string]any{"bucket": "my-bucket", "prefix": "images/"},
			wantMethods: []string{"GET", "GET"},
			wantPaths:   []string{"/my-bucket", "/my-bucket"},
			wantQuery:   []url.Values{nil, {"prefix": {"images/"}}},
		},
	}

	h.run(t, tests)
}
