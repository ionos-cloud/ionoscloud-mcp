package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Write-tool tests for snapshots and images — the two compute resources with no
// create tool. A snapshot comes from a volume via create_volume_snapshot; the API
// exposes no way to create an image at all.

const (
	snapshotsAPI = "/cloudapi/v6/snapshots"
	imagesAPI    = "/cloudapi/v6/images"
)

// TestNoCreateSnapshotOrImageTool documents the deliberate absence. Both resources
// are reachable in other ways, and inventing a create_ tool for either would be a
// tool that always fails.
func TestNoCreateSnapshotOrImageTool(t *testing.T) {
	names := toolNames(t, context.Background(), setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	for _, absent := range []string{"create_snapshot", "create_image"} {
		if names[absent] {
			t.Errorf("%q must not exist: the API offers no such operation", absent)
		}
	}
	// The supported route to a snapshot does exist.
	if !names["create_volume_snapshot"] {
		t.Error("create_volume_snapshot is the only way to make a snapshot and must be present")
	}
}

// ---------- snapshots ----------

func TestUpdateSnapshot(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_snapshot", map[string]any{
		"snapshot_id": "snap-1", "name": "renamed", "sec_auth_protection": true,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != snapshotsAPI+"/snap-1" {
		t.Errorf("PATCH path = %s, want %s", req.Path, snapshotsAPI+"/snap-1")
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("PATCH body is not JSON (%v): %s", err, req.Body)
	}
	if body["name"] != "renamed" || body["secAuthProtection"] != true {
		t.Errorf("PATCH body missing the requested fields: %s", req.Body)
	}
	// SnapshotProperties' WithDefaults constructor pre-sets exposeSerial=false and
	// requireLegacyBios=true; building from a zero literal keeps them out.
	for _, injected := range []string{"exposeSerial", "requireLegacyBios"} {
		if _, present := body[injected]; present {
			t.Errorf("PATCH must not carry the SDK-injected default %q:\n%s", injected, req.Body)
		}
	}
	if len(body) != 2 {
		t.Errorf("PATCH should contain exactly the two requested fields, got %s", req.Body)
	}
}

// TestUpdateSnapshotAcceptsTheWiderFlagSet checks the four capability flags that
// snapshots and images support but volumes do not.
func TestUpdateSnapshotAcceptsTheWiderFlagSet(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_snapshot", map[string]any{
		"snapshot_id": "snap-1",
		// the six shared with volumes
		"cpu_hot_plug": true, "ram_hot_plug": true, "nic_hot_plug": true,
		"nic_hot_unplug": true, "disc_virtio_hot_plug": true, "disc_virtio_hot_unplug": true,
		// the four only snapshots and images have
		"cpu_hot_unplug": true, "ram_hot_unplug": true,
		"disc_scsi_hot_plug": true, "disc_scsi_hot_unplug": true,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	for _, want := range []string{
		`"cpuHotPlug":true`, `"cpuHotUnplug":true`, `"ramHotUnplug":true`,
		`"discScsiHotPlug":true`, `"discScsiHotUnplug":true`,
	} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("PATCH body missing %s:\n%s", want, req.Body)
		}
	}
}

func TestUpdateSnapshotRejectsEmptyUpdate(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_snapshot", map[string]any{"snapshot_id": "snap-1"})
	if !res.IsError {
		t.Fatal("an update with no fields should be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("an empty update must not reach the API")
	}
}

func TestDeleteSnapshotTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(snapshotsAPI+"/snap-1", `{"id":"snap-1","properties":{
		"name":"pre-upgrade","description":"before the 24.04 upgrade","location":"de/fra",
		"size":50,"licenceType":"LINUX"}}`)

	preview, res := previewThenExecute(t, h, "delete_snapshot", map[string]any{"snapshot_id": "snap-1"})

	for _, want := range []string{"pre-upgrade", "before the 24.04 upgrade", "50", "only way back"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if !strings.Contains(preview, "restore_volume_snapshot") {
		t.Errorf("preview should name what stops working:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != snapshotsAPI+"/snap-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, snapshotsAPI+"/snap-1")
	}
}

// TestDeleteSnapshotFlagsProtectedSnapshot surfaces the signal that someone marked
// this snapshot as important.
func TestDeleteSnapshotFlagsProtectedSnapshot(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(snapshotsAPI+"/snap-1", `{"id":"snap-1","properties":{"name":"golden","secAuthProtection":true}}`)
	res := callTool(t, h, "delete_snapshot", map[string]any{"snapshot_id": "snap-1"})
	if !strings.Contains(resultText(res), "sec_auth_protection") {
		t.Errorf("preview should flag a protected snapshot:\n%s", resultText(res))
	}
}

// ---------- images ----------

// TestUpdateImagePreservesLicenceType is the carry-forward case for this batch:
// ImageProperties serializes licenceType unconditionally, so a PATCH built without it
// would send an empty OS type.
func TestUpdateImagePreservesLicenceType(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(imagesAPI+"/img-1", `{"id":"img-1","properties":{
		"name":"custom-ubuntu","licenceType":"LINUX","public":false}}`)

	res := callTool(t, h, "update_image", map[string]any{
		"image_id": "img-1", "description": "internal base image",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet {
		t.Fatalf("expected a GET (to read licenceType) then a PATCH, got %+v", reqs)
	}
	patch := reqs[1]
	if !strings.Contains(patch.Body, `"licenceType":"LINUX"`) {
		t.Errorf("PATCH must carry the existing licenceType forward:\n%s", patch.Body)
	}
	if strings.Contains(patch.Body, `"licenceType":""`) {
		t.Errorf("PATCH must never send an empty licenceType:\n%s", patch.Body)
	}
	if !strings.Contains(patch.Body, `"description":"internal base image"`) {
		t.Errorf("PATCH missing the requested description:\n%s", patch.Body)
	}
	// The injected defaults must still be absent.
	for _, injected := range []string{"exposeSerial", "requireLegacyBios"} {
		if strings.Contains(patch.Body, injected) {
			t.Errorf("PATCH must not carry the SDK-injected default %q:\n%s", injected, patch.Body)
		}
	}
}

func TestUpdateImageExplicitLicenceTypeSkipsRead(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_image", map[string]any{
		"image_id": "img-1", "licence_type": "WINDOWS2022",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if !strings.Contains(req.Body, `"licenceType":"WINDOWS2022"`) {
		t.Errorf("PATCH should carry the requested licenceType:\n%s", req.Body)
	}
}

func TestUpdateImageRejectsBlankLicenceType(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_image", map[string]any{"image_id": "img-1", "licence_type": "  "})
	if !res.IsError {
		t.Fatal("a blank licence_type should be rejected rather than sent")
	}
	if !strings.Contains(resultText(res), "omit it entirely") {
		t.Errorf("error should say how to keep the current value: %s", resultText(res))
	}
}

func TestDeleteImageTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(imagesAPI+"/img-1", `{"id":"img-1","properties":{
		"name":"custom-ubuntu","location":"de/fra","size":10,"licenceType":"LINUX",
		"imageType":"HDD","public":false,"imageAliases":["mycorp:ubuntu-base"]}}`)

	preview, res := previewThenExecute(t, h, "delete_image", map[string]any{"image_id": "img-1"})

	// The aliases matter: they are how scripts refer to the image, so deleting it
	// breaks anything using one.
	for _, want := range []string{"custom-ubuntu", "mycorp:ubuntu-base", "Terraform", "autoscaling"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != imagesAPI+"/img-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, imagesAPI+"/img-1")
	}
}

// TestDeleteImageWarnsOnPublicImage catches the case up front rather than after a
// token round trip: public IONOS images cannot be deleted at all.
func TestDeleteImageWarnsOnPublicImage(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(imagesAPI+"/img-pub", `{"id":"img-pub","properties":{
		"name":"ubuntu-24.04","licenceType":"LINUX","public":true}}`)

	res := callTool(t, h, "delete_image", map[string]any{"image_id": "img-pub"})
	preview := resultText(res)
	if !strings.Contains(preview, "cannot be deleted") {
		t.Errorf("preview should say a public image cannot be deleted:\n%s", preview)
	}
	if !strings.Contains(preview, "uploaded yourself") {
		t.Errorf("preview should say which images can be removed:\n%s", preview)
	}
}

func TestSnapshotImageScopeGating(t *testing.T) {
	ctx := context.Background()
	readOnly := toolNames(t, ctx, setup(t))
	write := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true}))
	destructive := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true, Destructive: true}))

	for _, name := range []string{"update_snapshot", "update_image"} {
		if readOnly[name] {
			t.Errorf("read-only scope must not expose %q", name)
		}
		if !write[name] {
			t.Errorf("write scope should expose %q", name)
		}
	}
	for _, name := range []string{"delete_snapshot", "delete_image"} {
		if write[name] {
			t.Errorf("write scope must not expose destructive %q", name)
		}
		if !destructive[name] {
			t.Errorf("destructive scope should expose %q", name)
		}
	}
}
