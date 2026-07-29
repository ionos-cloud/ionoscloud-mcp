package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Write-tool tests for the core compute resources (server, volume, NIC, LAN).
// The datacenter blueprint's own tests live in compute_write_test.go.

const (
	dcID       = "dc-1"
	srvID      = "srv-1"
	serversAPI = "/cloudapi/v6/datacenters/dc-1/servers"
	volumesAPI = "/cloudapi/v6/datacenters/dc-1/volumes"
	nicsAPI    = "/cloudapi/v6/datacenters/dc-1/servers/srv-1/nics"
	lansAPI    = "/cloudapi/v6/datacenters/dc-1/lans"
)

// destructiveSetup is the scope every two-phase test needs.
func destructiveSetup(t *testing.T) *testSetup {
	t.Helper()
	return setupWithScope(t, tools.Scope{Write: true, Destructive: true})
}

// callTool invokes a tool and fails the test on a transport error.
func callTool(t *testing.T, h *testSetup, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("calling %s: %v", name, err)
	}
	return res
}

// assertNoMutation fails if any request so far used a mutating HTTP method. This
// is the core two-phase guarantee: the preview call must not change anything.
func assertNoMutation(t *testing.T, h *testSetup, phase string) {
	t.Helper()
	for _, r := range h.log.allRequests() {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			t.Fatalf("%s must not issue %s %s", phase, r.Method, r.Path)
		}
	}
}

// previewThenExecute runs both phases of a two-phase tool: it asserts the preview
// mutates nothing, extracts the minted token, clears the request log, then calls
// again with the token. It returns the preview text and the execute result so a
// caller can assert on both.
func previewThenExecute(t *testing.T, h *testSetup, tool string, args map[string]any) (string, *mcp.CallToolResult) {
	t.Helper()

	res := callTool(t, h, tool, args)
	if res.IsError {
		t.Fatalf("%s preview must not be an error: %s", tool, resultText(res))
	}
	previewText := resultText(res)
	assertNoMutation(t, h, tool+" preview phase")
	token := extractToken(t, previewText)

	// Clear so the caller's assertions see only the execute phase's requests.
	h.log.clear()

	execArgs := map[string]any{"confirmation_token": token}
	for k, v := range args {
		execArgs[k] = v
	}
	return previewText, callTool(t, h, tool, execArgs)
}

// singleRequest asserts exactly one request was made, with the given method, and
// returns it.
func singleRequest(t *testing.T, h *testSetup, method string) recordedRequest {
	t.Helper()
	reqs := h.log.allRequests()
	if len(reqs) != 1 {
		t.Fatalf("expected exactly 1 request, got %d: %+v", len(reqs), reqs)
	}
	if reqs[0].Method != method {
		t.Fatalf("expected a %s, got %s %s", method, reqs[0].Method, reqs[0].Path)
	}
	return reqs[0]
}

// ---------- create_server ----------

func TestCreateServerTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "web-1", "cores": 4, "ram": 2048,
	})

	for _, want := range []string{"web-1", "4", "2048", dcID} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	// With no boot_volume the server comes up with no disk and no NIC, which the
	// preview must say — see also TestCreateServerWithoutBootVolumeSendsNoEntities.
	// Assert the actionable substance, not the exact phrasing: the caller must
	// learn the server has no disk or network and which tools fix that.
	for _, want := range []string{"no disk and no network", "attach_server_volume", "create_nic"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview should tell the caller %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != serversAPI {
		t.Errorf("POST path = %s, want %s", req.Path, serversAPI)
	}
	for _, want := range []string{`"name":"web-1"`, `"cores":4`, `"ram":2048`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateServerRequiresASize covers the pre-flight validation: a server with
// neither cores+ram nor a template has no size, and the API can only report that
// after a round trip.
func TestCreateServerRequiresASize(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{"datacenter_id": dcID, "name": "web-1"})
	if !res.IsError {
		t.Fatal("creating a server with no cores/ram and no template_uuid should be rejected")
	}
	if !strings.Contains(resultText(res), "template_uuid") {
		t.Errorf("error should point at cores/ram or template_uuid: %s", resultText(res))
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestCreateServerRejectsBothSizings guards the other half: CUBE/GPU servers take
// their size from the template, so cores+ram together with a template is a
// contradiction rather than a merge.
func TestCreateServerRejectsBothSizings(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "web-1", "cores": 4, "ram": 2048, "template_uuid": "tpl-1",
	})
	if !res.IsError {
		t.Fatal("cores+ram together with template_uuid should be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestCreateCubeServerRequiresDasBootVolume covers the failure that made CUBE
// servers uncreatable: a CUBE server's storage is Direct Attached Storage, which
// the API accepts ONLY inside the composite create request. Attaching a volume
// afterwards does not work, so without an inline boot_volume there is no recovery
// path — the error has to say exactly what to pass.
func TestCreateCubeServerRequiresDasBootVolume(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "cube-1", "type": "CUBE", "template_uuid": "tpl-cube-xs",
	})
	if !res.IsError {
		t.Fatal("a CUBE server without a boot_volume must be rejected up front")
	}
	txt := resultText(res)
	for _, want := range []string{"boot_volume", "DAS"} {
		if !strings.Contains(txt, want) {
			t.Errorf("error should name boot_volume and DAS, got: %s", txt)
		}
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestCreateCubeServerWithDasBootVolume is the fix: the DAS volume rides along in
// entities.volumes of the same POST.
func TestCreateCubeServerWithDasBootVolume(t *testing.T) {
	h := destructiveSetup(t)
	args := map[string]any{
		"datacenter_id": dcID, "name": "cube-1", "type": "CUBE", "template_uuid": "tpl-cube-xs",
		"boot_volume": map[string]any{"type": "DAS", "name": "cube-1-das", "image_alias": "ubuntu:latest"},
	}
	preview, res := previewThenExecute(t, h, "create_server", args)

	for _, want := range []string{"CUBE", "DAS", "fixed by the template", "ubuntu:latest"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}

	req := singleRequest(t, h, http.MethodPost)
	if req.Path != serversAPI {
		t.Errorf("POST path = %s, want %s", req.Path, serversAPI)
	}

	// The composite shape is the whole point: one request carrying both the
	// server properties and the volume under entities.volumes.items.
	var body struct {
		Properties struct {
			Name         string `json:"name"`
			Type         string `json:"type"`
			TemplateUuid string `json:"templateUuid"`
		} `json:"properties"`
		Entities struct {
			Volumes struct {
				Items []struct {
					Properties map[string]any `json:"properties"`
				} `json:"items"`
			} `json:"volumes"`
		} `json:"entities"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("POST body is not JSON (%v): %s", err, req.Body)
	}
	if body.Properties.Type != "CUBE" || body.Properties.TemplateUuid != "tpl-cube-xs" {
		t.Errorf("server properties wrong: %+v", body.Properties)
	}
	items := body.Entities.Volumes.Items
	if len(items) != 1 {
		t.Fatalf("expected exactly one inline volume, got %d: %s", len(items), req.Body)
	}
	vol := items[0].Properties
	if vol["type"] != "DAS" {
		t.Errorf("inline volume type = %v, want DAS", vol["type"])
	}
	if vol["imageAlias"] != "ubuntu:latest" {
		t.Errorf("inline volume imageAlias = %v, want ubuntu:latest", vol["imageAlias"])
	}
	// A DAS volume's size comes from the template; sending one is rejected.
	if _, ok := vol["size"]; ok {
		t.Errorf("inline DAS volume must not carry a size: %s", req.Body)
	}
}

// TestCreateGpuServerRequiresBootVolume covers the GPU half of the same trap. GPU
// servers are template-sized like CUBE, so the API accepts their storage only
// inside the composite create — a GPU server created without a boot_volume hits
// the identical dead end.
//
// Evidence for the requirement, since it is not stated in the Go SDK: the team's
// Terraform provider marks the inline volume Required in resource_gpu_server.go
// (Optional in the ENTERPRISE/VCPU resource), and ionosctl builds the volume into
// entities for GPU as well as CUBE.
func TestCreateGpuServerRequiresBootVolume(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "gpu-1", "type": "GPU", "template_uuid": "tpl-gpu",
	})
	if !res.IsError {
		t.Fatal("a GPU server without a boot_volume must be rejected up front")
	}
	if !strings.Contains(resultText(res), "boot_volume") {
		t.Errorf("error should name boot_volume, got: %s", resultText(res))
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestCreateGpuServerWithSsdPremium mirrors the config the Terraform provider's
// GPU acceptance test runs against the real API: SSD Premium, licence_type LINUX,
// availability_zone AUTO, and NO size — the template fixes it.
func TestCreateGpuServerWithSsdPremium(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "gpu-1", "type": "GPU",
		"template_uuid": "6913ed82-a143-4c15-89ac-08fb375a97c5", "availability_zone": "AUTO",
		"boot_volume": map[string]any{
			"type": "SSD Premium", "name": "system", "licence_type": "LINUX", "bus": "VIRTIO",
		},
	})

	for _, want := range []string{"GPU", "SSD Premium", "fixed by the template"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}

	req := singleRequest(t, h, http.MethodPost)
	var body struct {
		Properties struct {
			Type         string `json:"type"`
			TemplateUuid string `json:"templateUuid"`
		} `json:"properties"`
		Entities struct {
			Volumes struct {
				Items []struct {
					Properties map[string]any `json:"properties"`
				} `json:"items"`
			} `json:"volumes"`
		} `json:"entities"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("POST body is not JSON (%v): %s", err, req.Body)
	}
	if body.Properties.Type != "GPU" {
		t.Errorf("server type = %q, want GPU", body.Properties.Type)
	}
	items := body.Entities.Volumes.Items
	if len(items) != 1 {
		t.Fatalf("expected exactly one inline volume, got %d: %s", len(items), req.Body)
	}
	vol := items[0].Properties
	if vol["type"] != "SSD Premium" {
		t.Errorf("inline volume type = %v, want SSD Premium", vol["type"])
	}
	if vol["licenceType"] != "LINUX" {
		t.Errorf("inline volume licenceType = %v, want LINUX", vol["licenceType"])
	}
	// A template-sized server's storage size comes from the template; neither the
	// TF provider nor ionosctl sends one for GPU.
	if _, ok := vol["size"]; ok {
		t.Errorf("inline GPU volume must not carry a size: %s", req.Body)
	}
}

// TestCreateGpuServerWithoutBootVolumeType covers ionosctl's actual behaviour: for
// GPU it sends no storage type at all and lets the API choose, so boot_volume.type
// must be optional for this server type (and only this one, besides CUBE's
// mandatory DAS).
func TestCreateGpuServerWithoutBootVolumeType(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "gpu-1", "type": "GPU", "template_uuid": "tpl-gpu",
		"boot_volume": map[string]any{"licence_type": "LINUX", "name": "system"},
	})
	if !strings.Contains(preview, "chosen by the API") {
		t.Errorf("preview should say the storage type is left to the API:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("a GPU boot_volume with no type should be accepted: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	// No type key at all, rather than an empty string the API would reject.
	if strings.Contains(req.Body, `"type":""`) {
		t.Errorf("an unset boot_volume.type must be omitted, not sent empty:\n%s", req.Body)
	}
}

// TestCreateServerBootVolumeTypeRules pins the cross-field rules that the API only
// reports as a late, generic rejection.
func TestCreateServerBootVolumeTypeRules(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantMsg string
	}{
		{
			name: "CUBE rejects a non-DAS boot volume",
			args: map[string]any{
				"datacenter_id": dcID, "name": "c", "type": "CUBE", "template_uuid": "tpl",
				"boot_volume": map[string]any{"type": "SSD", "size": 50, "image_alias": "ubuntu:latest"},
			},
			wantMsg: "must be DAS",
		},
		{
			name: "CUBE size must be omitted",
			args: map[string]any{
				"datacenter_id": dcID, "name": "c", "type": "CUBE", "template_uuid": "tpl",
				"boot_volume": map[string]any{"type": "DAS", "size": 60, "image_alias": "ubuntu:latest"},
			},
			wantMsg: "size must be omitted",
		},
		{
			// The GPU half of the same rule: template-sized, so no size.
			name: "GPU size must be omitted",
			args: map[string]any{
				"datacenter_id": dcID, "name": "g", "type": "GPU", "template_uuid": "tpl",
				"boot_volume": map[string]any{"type": "SSD Premium", "size": 100, "licence_type": "LINUX"},
			},
			wantMsg: "size must be omitted for a GPU server",
		},
		{
			// CUBE must state DAS explicitly, unlike GPU which may omit the type.
			name: "CUBE requires an explicit DAS type",
			args: map[string]any{
				"datacenter_id": dcID, "name": "c", "type": "CUBE", "template_uuid": "tpl",
				"boot_volume": map[string]any{"licence_type": "LINUX"},
			},
			wantMsg: "must be set to DAS",
		},
		{
			name: "boot volume needs an image or a licence type",
			args: map[string]any{
				"datacenter_id": dcID, "name": "e", "cores": 2, "ram": 1024,
				"boot_volume": map[string]any{"type": "SSD", "size": 50},
			},
			wantMsg: "licence_type",
		},
		{
			name: "template_uuid without type cannot be validated",
			args: map[string]any{
				"datacenter_id": dcID, "name": "c", "template_uuid": "tpl",
			},
			wantMsg: "type is required when template_uuid is set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_server", tt.args)
			if !res.IsError {
				t.Fatalf("expected rejection, got success: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantMsg) {
				t.Errorf("error should mention %q, got: %s", tt.wantMsg, resultText(res))
			}
			if len(h.log.allRequests()) != 0 {
				t.Error("validation failure must not reach the API")
			}
		})
	}
}

// TestCreateServerBootVolumeWarningsDoNotBlock covers the rules that are inferred
// rather than documented. They must surface in the preview but must NOT reject the
// request: a mistaken block would break a valid request with no way around it,
// while an unwanted storage type is trivially recoverable — the API rejects it and
// the caller retries. Only the documented rules (CUBE's DAS type, the missing
// inline volume on a template-sized server) are hard errors.
func TestCreateServerBootVolumeWarningsDoNotBlock(t *testing.T) {
	tests := []struct {
		name     string
		bootVol  map[string]any
		wantWarn string
	}{
		{
			// The specific case: DAS on an ENTERPRISE server. Implied by the spec
			// ("DAS could be used only in a composite call with a Cube server") but
			// never stated as an ENTERPRISE rejection, so the API decides.
			name:     "DAS on ENTERPRISE warns",
			bootVol:  map[string]any{"type": "DAS", "image_alias": "ubuntu:latest"},
			wantWarn: "DAS storage is documented for template-sized CUBE servers",
		},
		{
			// Contradicted by the Terraform provider, which marks the inline
			// volume's size Optional+Computed.
			name:     "missing size on ENTERPRISE warns",
			bootVol:  map[string]any{"type": "SSD", "image_alias": "ubuntu:latest"},
			wantWarn: "boot_volume.size is not set",
		},
		{
			name:     "missing type on ENTERPRISE warns",
			bootVol:  map[string]any{"size": 50, "image_alias": "ubuntu:latest"},
			wantWarn: "boot_volume.type is not set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			args := map[string]any{
				"datacenter_id": dcID, "name": "web-1", "cores": 2, "ram": 1024,
				"boot_volume": tt.bootVol,
			}
			res := callTool(t, h, "create_server", args)
			if res.IsError {
				t.Fatalf("this combination is inferred, not documented — it must warn, not block: %s", resultText(res))
			}
			preview := resultText(res)
			if !strings.Contains(preview, "NOTE:") || !strings.Contains(preview, tt.wantWarn) {
				t.Errorf("preview should carry the advisory %q:\n%s", tt.wantWarn, preview)
			}
			// A warning still mints a token, so the caller can proceed.
			token := extractToken(t, preview)
			h.log.clear()
			args["confirmation_token"] = token
			if res := callTool(t, h, "create_server", args); res.IsError {
				t.Fatalf("execute after a warning should still work: %s", resultText(res))
			}
			singleRequest(t, h, http.MethodPost)
		})
	}
}

// TestTemplateSizedServersGetNoBootVolumeWarnings guards against the warnings
// firing where they make no sense: a GPU server legitimately omits both the
// storage type and the size.
func TestTemplateSizedServersGetNoBootVolumeWarnings(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "gpu-1", "type": "GPU", "template_uuid": "tpl",
		"boot_volume": map[string]any{"licence_type": "LINUX"},
	})
	if res.IsError {
		t.Fatalf("a GPU boot volume with no type and no size is valid: %s", resultText(res))
	}
	if strings.Contains(resultText(res), "NOTE:") {
		t.Errorf("template-sized servers must not get the non-template warnings:\n%s", resultText(res))
	}
}

// TestCreateServerTokenBoundToBootVolume stops a preview shown "with a DAS disk"
// from executing as "with no disk", which would silently produce an unbootable
// server — and for CUBE, one the API rejects outright.
func TestCreateServerTokenBoundToBootVolume(t *testing.T) {
	h := destructiveSetup(t)
	withVolume := map[string]any{
		"datacenter_id": dcID, "name": "cube-1", "type": "CUBE", "template_uuid": "tpl",
		"boot_volume": map[string]any{"type": "DAS", "image_alias": "ubuntu:latest"},
	}
	res := callTool(t, h, "create_server", withVolume)
	token := extractToken(t, resultText(res))
	h.log.clear()

	// Replay the token against an ENTERPRISE server with no boot volume: same
	// datacenter and name, different thing being created.
	res = callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "cube-1", "cores": 2, "ram": 1024,
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token previewed with a boot volume must not create a server without one")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("a mismatched token must not reach the API")
	}
}

// TestCreateServerWithoutBootVolumeSendsNoEntities keeps the non-composite path
// clean: a server created without a boot volume must not send an empty entities
// object.
func TestCreateServerWithoutBootVolumeSendsNoEntities(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "web-1", "cores": 2, "ram": 1024,
	})
	// Assert the actionable substance, not the exact phrasing: the caller must
	// learn the server has no disk or network and which tools fix that.
	for _, want := range []string{"no disk and no network", "attach_server_volume", "create_nic"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview should tell the caller %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if strings.Contains(req.Body, "entities") {
		t.Errorf("a server with no boot volume must send no entities key:\n%s", req.Body)
	}
}

func TestCreateServerBadToken(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "web-1", "cores": 4, "ram": 2048,
		"confirmation_token": "not-a-real-token",
	})
	if !res.IsError {
		t.Fatal("a bogus confirmation_token must be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("a rejected token must not reach the API")
	}
}

// TestCreateServerTokenBoundToName proves the token is target-bound: one minted
// for web-1 cannot create db-1.
func TestCreateServerTokenBoundToName(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "web-1", "cores": 4, "ram": 2048,
	})
	token := extractToken(t, resultText(res))
	h.log.clear()

	res = callTool(t, h, "create_server", map[string]any{
		"datacenter_id": dcID, "name": "db-1", "cores": 4, "ram": 2048,
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token minted for web-1 must not create db-1")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("a mismatched token must not reach the API")
	}
}

// ---------- update_server ----------

func TestUpdateServer(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "name": "renamed", "cores": 8,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != serversAPI+"/"+srvID {
		t.Errorf("PATCH path = %s, want %s", req.Path, serversAPI+"/"+srvID)
	}
	for _, want := range []string{`"name":"renamed"`, `"cores":8`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("PATCH body missing %s:\n%s", want, req.Body)
		}
	}
	// Partial update: fields the caller did not pass must be absent, so the API
	// leaves them alone rather than receiving a zero value.
	for _, absent := range []string{"ram", "hostname", "cpuFamily", "nicMultiQueue"} {
		if strings.Contains(req.Body, absent) {
			t.Errorf("PATCH body should omit un-passed field %q:\n%s", absent, req.Body)
		}
	}
}

// TestUpdateServerSetsBootVolume covers the capability whose absence stranded a
// real disk-swap: attaching a volume does not make it bootable, and detaching the
// previous boot volume clears the server's boot setting outright. With no way to
// set it, a server could be driven into an unbootable state the toolset could not
// repair — the only remaining option was to delete and recreate it.
//
// The API's mechanism is a server PATCH carrying properties.bootVolume as a
// reference to an already-attached volume, which is what the Terraform provider's
// UpdateBootDevice does.
func TestUpdateServerSetsBootVolume(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "boot_volume_id": "vol-new",
	})
	if res.IsError {
		t.Fatalf("setting the boot volume failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != serversAPI+"/"+srvID {
		t.Errorf("PATCH path = %s, want the server path", req.Path)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("PATCH body is not JSON (%v): %s", err, req.Body)
	}
	bootVol, ok := body["bootVolume"].(map[string]any)
	if !ok {
		t.Fatalf("PATCH body should carry bootVolume as a reference object: %s", req.Body)
	}
	if bootVol["id"] != "vol-new" {
		t.Errorf("bootVolume.id = %v, want vol-new", bootVol["id"])
	}
	// Still a partial update: nothing the caller did not ask for.
	if len(body) != 1 {
		t.Errorf("PATCH body should contain only bootVolume, got %s", req.Body)
	}
}

func TestUpdateServerRejectsBlankBootVolumeID(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "boot_volume_id": "   ",
	})
	if !res.IsError {
		t.Fatal("a blank boot_volume_id should be rejected rather than sent")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestUpdateVolumeSetsBootOrder covers the complementary mechanism the agent could
// see (bootOrder on the volume) but not act on.
func TestUpdateVolumeSetsBootOrder(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_volume", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "boot_order": "primary",
	})
	if res.IsError {
		t.Fatalf("setting boot_order failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	// Normalised to upper case, since the API's enum is upper case.
	if !strings.Contains(req.Body, `"bootOrder":"PRIMARY"`) {
		t.Errorf("PATCH body should carry bootOrder PRIMARY:\n%s", req.Body)
	}
}

func TestUpdateVolumeRejectsBadBootOrder(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_volume", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "boot_order": "FIRST",
	})
	if !res.IsError {
		t.Fatal("an invalid boot_order should be rejected before the request")
	}
	if !strings.Contains(resultText(res), "PRIMARY") {
		t.Errorf("error should list the valid values: %s", resultText(res))
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestDiskSwapIsReachable walks the whole sequence from the reported flow, in the
// order the tools now recommend: attach the replacement, point the server at it,
// then detach the old one. The point is that every step exists — previously the
// middle one did not, which is what turned a disk swap into "delete and recreate".
func TestDiskSwapIsReachable(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()
	names := toolNames(t, ctx, h)
	for _, needed := range []string{"attach_server_volume", "update_server", "detach_server_volume"} {
		if !names[needed] {
			t.Fatalf("%q is required to swap a server's disk without recreating it", needed)
		}
	}

	// 1. attach the replacement volume
	if res := callTool(t, h, "attach_server_volume", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "volume_id": "vol-new",
	}); res.IsError {
		t.Fatalf("attach failed: %s", resultText(res))
	}
	// 2. make it the boot device BEFORE detaching, so the server is never left
	//    without one
	h.log.clear()
	if res := callTool(t, h, "update_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "boot_volume_id": "vol-new",
	}); res.IsError {
		t.Fatalf("setting the boot volume failed: %s", resultText(res))
	}
	patch := singleRequest(t, h, http.MethodPatch)
	if !strings.Contains(patch.Body, "vol-new") {
		t.Errorf("boot volume PATCH should reference vol-new:\n%s", patch.Body)
	}
	// 3. detach the old volume
	h.resp.serve(volumesAPI+"/vol-old", `{"id":"vol-old","properties":{"name":"old","bootServer":"srv-1"}}`)
	h.log.clear()
	_, res := previewThenExecute(t, h, "detach_server_volume", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "volume_id": "vol-old",
	})
	if res.IsError {
		t.Fatalf("detach failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// TestVolumeToolsPointAtTheBootRecoveryPath checks the descriptions and preview
// name the tool that fixes an unbootable server. The gap in the reported flow was
// as much discoverability as capability: the agent inspected bootVolume and
// bootOrder, found no tool for either, and gave up.
func TestVolumeToolsPointAtTheBootRecoveryPath(t *testing.T) {
	ctx := context.Background()
	byName := map[string]*mcp.Tool{}
	for _, tool := range computeOnlyTools(t, ctx, tools.Scope{Write: true, Destructive: true}) {
		byName[tool.Name] = tool
	}
	for _, name := range []string{"attach_server_volume", "detach_server_volume"} {
		tool, ok := byName[name]
		if !ok {
			t.Fatalf("%q not registered", name)
		}
		if !strings.Contains(tool.Description, "boot_volume_id") {
			t.Errorf("%s's description should name update_server boot_volume_id as the way to set the boot device:\n%s", name, tool.Description)
		}
	}

	// The detach preview must carry the recovery instruction too, since that is
	// what a caller reads at the moment of the decision.
	h := destructiveSetup(t)
	h.resp.serve(volumesAPI+"/vol-1", `{"id":"vol-1","properties":{"name":"data","bootServer":"srv-1"}}`)
	preview := resultText(callTool(t, h, "detach_server_volume", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "volume_id": "vol-1",
	}))
	if !strings.Contains(preview, "boot_volume_id") {
		t.Errorf("detach preview should name the recovery path:\n%s", preview)
	}
}

func TestUpdateServerRejectsEmptyUpdate(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_server", map[string]any{"datacenter_id": dcID, "server_id": srvID})
	if !res.IsError {
		t.Fatal("an update with no fields should be rejected rather than sent")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("an empty update must not reach the API")
	}
}

// ---------- delete_server ----------

func TestDeleteServerTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(serversAPI+"/"+srvID, `{
		"id":"srv-1","properties":{"name":"web-1","type":"ENTERPRISE","vmState":"RUNNING"},
		"entities":{
			"volumes":{"items":[{"id":"v1"},{"id":"v2"}]},
			"nics":{"items":[{"id":"n1"}]}
		}}`)

	preview, res := previewThenExecute(t, h, "delete_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID,
	})

	// delete_volumes defaults to false, so the volumes survive — and the preview
	// must say so, including that they keep costing money.
	for _, want := range []string{"web-1", "1 NICs", "SURVIVE", "cost"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != serversAPI+"/"+srvID {
		t.Errorf("DELETE path = %s, want %s", req.Path, serversAPI+"/"+srvID)
	}
	if got := req.Query.Get("deleteVolumes"); got != "false" {
		t.Errorf("deleteVolumes query = %q, want \"false\" (the safe default)", got)
	}
	// DELETE returns no body, so the tool reports acceptance in text and must
	// explain what happened to the volumes.
	if txt := resultText(res); !strings.Contains(txt, "kept") {
		t.Errorf("success text should say the volumes were kept: %s", txt)
	}
}

func TestDeleteServerWithVolumes(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(serversAPI+"/"+srvID, `{
		"id":"srv-1","properties":{"name":"web-1"},
		"entities":{"volumes":{"items":[{"id":"v1"},{"id":"v2"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "delete_volumes": true,
	})

	for _, want := range []string{"2 attached volumes", "DESTROYED"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if got := req.Query.Get("deleteVolumes"); got != "true" {
		t.Errorf("deleteVolumes query = %q, want \"true\"", got)
	}
}

// TestDeleteServerTokenBoundToDeleteVolumes is the reason delete_volumes is part
// of the confirmation target: a token previewed as "keep the volumes" must not be
// replayable as "destroy them", or the two-phase preview would be describing
// something other than what executes.
func TestDeleteServerTokenBoundToDeleteVolumes(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(serversAPI+"/"+srvID, `{"id":"srv-1","properties":{"name":"web-1"},
		"entities":{"volumes":{"items":[{"id":"v1"}]}}}`)

	res := callTool(t, h, "delete_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "delete_volumes": false,
	})
	token := extractToken(t, resultText(res))
	h.log.clear()

	res = callTool(t, h, "delete_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "delete_volumes": true,
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token previewed with delete_volumes=false must not execute with delete_volumes=true")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("a mismatched token must not issue a DELETE")
		}
	}
}

func TestDeleteServerNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(serversAPI+"/"+srvID, http.StatusNotFound, `{"messages":[{"message":"not found"}]}`)

	res := callTool(t, h, "delete_server", map[string]any{"datacenter_id": dcID, "server_id": srvID})
	if !res.IsError {
		t.Fatal("deleting a non-existent server should report an error")
	}
	if !strings.Contains(resultText(res), "does not exist") {
		t.Errorf("want a friendly does-not-exist message, got: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("a 404 preview must not issue a DELETE")
		}
	}
}

// ---------- create_volume ----------

func TestCreateVolumeTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_volume", map[string]any{
		"datacenter_id": dcID, "name": "data-1", "size": 50, "type": "SSD",
		"image": "img-1", "image_password": "sup3rsecret", "ssh_keys": []string{"ssh-rsa AAA", "ssh-rsa BBB"},
	})

	for _, want := range []string{"data-1", "50", "SSD", "img-1"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	// Previews are shown to the model and logged by clients, so secrets must be
	// acknowledged without being echoed.
	if strings.Contains(preview, "sup3rsecret") {
		t.Errorf("image_password must not be echoed into the preview:\n%s", preview)
	}
	if !strings.Contains(preview, "not shown") {
		t.Errorf("preview should report that image_password was set:\n%s", preview)
	}
	if !strings.Contains(preview, "2 key(s)") {
		t.Errorf("preview should summarise the SSH keys rather than list them:\n%s", preview)
	}

	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != volumesAPI {
		t.Errorf("POST path = %s, want %s", req.Path, volumesAPI)
	}
	// The secret is redacted in the preview but must still reach the API.
	for _, want := range []string{`"name":"data-1"`, `"size":50`, `"type":"SSD"`, `"image":"img-1"`, `"imagePassword":"sup3rsecret"`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateVolumeRequiresImageOrLicence covers the pre-flight check: an empty
// volume has no image to infer a licence type from.
func TestCreateVolumeRequiresImageOrLicence(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_volume", map[string]any{
		"datacenter_id": dcID, "name": "data-1", "size": 50, "type": "SSD",
	})
	if !res.IsError {
		t.Fatal("a volume with no image and no licence_type should be rejected")
	}
	if !strings.Contains(resultText(res), "licence_type") {
		t.Errorf("error should mention licence_type: %s", resultText(res))
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

func TestUpdateVolume(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_volume", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "size": 100,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != volumesAPI+"/vol-1" {
		t.Errorf("PATCH path = %s, want %s", req.Path, volumesAPI+"/vol-1")
	}
	if !strings.Contains(req.Body, `"size":100`) {
		t.Errorf("PATCH body missing size:\n%s", req.Body)
	}
	// image_password and user_data are creation-only; they must never appear in
	// an update body even as empty values.
	for _, absent := range []string{"imagePassword", "userData", "image", "licenceType"} {
		if strings.Contains(req.Body, absent) {
			t.Errorf("PATCH body must omit creation-only field %q:\n%s", absent, req.Body)
		}
	}
}

// TestVolumeHotPlugFlags covers the capability flags that decide whether the
// server a volume is attached to can change CPU, RAM, NICs or disks without a
// reboot. They live on the VOLUME, not the server, which is easy to miss — and
// without them a running server cannot be resized at all.
//
// They are exposed on all three paths that carry volume properties, matching
// ionosctl (volume create and update) and the Terraform provider (the volume
// resource plus the inline volume block of server, cube_server and gpu_server).
func TestVolumeHotPlugFlags(t *testing.T) {
	allFlags := map[string]any{
		"cpu_hot_plug": true, "ram_hot_plug": true,
		"nic_hot_plug": true, "nic_hot_unplug": false,
		"disc_virtio_hot_plug": true, "disc_virtio_hot_unplug": false,
	}
	wantJSON := []string{
		`"cpuHotPlug":true`, `"ramHotPlug":true`,
		`"nicHotPlug":true`, `"nicHotUnplug":false`,
		`"discVirtioHotPlug":true`, `"discVirtioHotUnplug":false`,
	}

	t.Run("update_volume", func(t *testing.T) {
		h := destructiveSetup(t)
		args := map[string]any{"datacenter_id": dcID, "volume_id": "vol-1"}
		for k, v := range allFlags {
			args[k] = v
		}
		if res := callTool(t, h, "update_volume", args); res.IsError {
			t.Fatalf("update failed: %s", resultText(res))
		}
		req := singleRequest(t, h, http.MethodPatch)
		for _, want := range wantJSON {
			if !strings.Contains(req.Body, want) {
				t.Errorf("PATCH body missing %s:\n%s", want, req.Body)
			}
		}
	})

	t.Run("create_volume", func(t *testing.T) {
		h := destructiveSetup(t)
		args := map[string]any{
			"datacenter_id": dcID, "name": "data-1", "size": 50, "type": "SSD",
			"image_alias": "ubuntu:latest",
		}
		for k, v := range allFlags {
			args[k] = v
		}
		preview, res := previewThenExecute(t, h, "create_volume", args)
		if !strings.Contains(preview, "cpu_hot_plug") {
			t.Errorf("preview should list the capability flags it will set:\n%s", preview)
		}
		if res.IsError {
			t.Fatalf("create failed: %s", resultText(res))
		}
		req := singleRequest(t, h, http.MethodPost)
		for _, want := range wantJSON {
			if !strings.Contains(req.Body, want) {
				t.Errorf("POST body missing %s:\n%s", want, req.Body)
			}
		}
	})

	t.Run("create_server inline boot_volume", func(t *testing.T) {
		h := destructiveSetup(t)
		bootVol := map[string]any{"type": "SSD", "size": 30, "image_alias": "ubuntu:latest"}
		for k, v := range allFlags {
			bootVol[k] = v
		}
		_, res := previewThenExecute(t, h, "create_server", map[string]any{
			"datacenter_id": dcID, "name": "web-1", "cores": 2, "ram": 1024,
			"boot_volume": bootVol,
		})
		if res.IsError {
			t.Fatalf("create failed: %s", resultText(res))
		}
		req := singleRequest(t, h, http.MethodPost)
		for _, want := range wantJSON {
			if !strings.Contains(req.Body, want) {
				t.Errorf("composite POST body missing %s:\n%s", want, req.Body)
			}
		}
	})
}

// TestUpdateVolumeHotPlugFlagsAreIndependent checks each flag can be set alone —
// so a caller turning on RAM hot-plug does not silently reset the other five, which
// is exactly the failure mode the PATCH-purity rule exists to prevent.
func TestUpdateVolumeHotPlugFlagsAreIndependent(t *testing.T) {
	cases := map[string]string{
		"cpu_hot_plug":           "cpuHotPlug",
		"ram_hot_plug":           "ramHotPlug",
		"nic_hot_plug":           "nicHotPlug",
		"nic_hot_unplug":         "nicHotUnplug",
		"disc_virtio_hot_plug":   "discVirtioHotPlug",
		"disc_virtio_hot_unplug": "discVirtioHotUnplug",
	}
	for field, jsonKey := range cases {
		t.Run(field, func(t *testing.T) {
			h := destructiveSetup(t)
			if res := callTool(t, h, "update_volume", map[string]any{
				"datacenter_id": dcID, "volume_id": "vol-1", field: true,
			}); res.IsError {
				t.Fatalf("update failed: %s", resultText(res))
			}
			req := singleRequest(t, h, http.MethodPatch)
			var body map[string]any
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				t.Fatalf("PATCH body is not JSON (%v): %s", err, req.Body)
			}
			if body[jsonKey] != true {
				t.Errorf("PATCH body should set %s: %s", jsonKey, req.Body)
			}
			// Exactly one property: none of the other five leaked in.
			if len(body) != 1 {
				t.Errorf("setting %s alone must send only that field, got %s", field, req.Body)
			}
		})
	}
}

// TestUpdateVolumeStillRejectsEmptyUpdate guards the guard: adding six optional
// fields must not make an all-empty update look like a real change.
func TestUpdateVolumeStillRejectsEmptyUpdate(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_volume", map[string]any{"datacenter_id": dcID, "volume_id": "vol-1"})
	if !res.IsError {
		t.Fatal("an update with no fields should still be rejected")
	}
	if !strings.Contains(resultText(res), "cpu_hot_plug") {
		t.Errorf("the error should list the newly available fields too: %s", resultText(res))
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("an empty update must not reach the API")
	}
}

func TestDeleteVolumeWarnsWhenAttached(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(volumesAPI+"/vol-1", `{"id":"vol-1","properties":{
		"name":"data-1","size":50,"type":"SSD","bootServer":"srv-9"}}`)

	preview, res := previewThenExecute(t, h, "delete_volume", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1",
	})

	// A volume still attached to a server may be its boot disk; deleting it can
	// leave that server unbootable, which the preview has to surface.
	for _, want := range []string{"srv-9", "WARNING", "boot"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != volumesAPI+"/vol-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, volumesAPI+"/vol-1")
	}
}

// ---------- create/update/delete NIC ----------

func TestCreateNicTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "lan": 2, "name": "eth0",
	})

	for _, want := range []string{"eth0", "2"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != nicsAPI {
		t.Errorf("POST path = %s, want %s", req.Path, nicsAPI)
	}
	for _, want := range []string{`"name":"eth0"`, `"lan":2`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateNicWarnsAboutEmptyFirewall covers the trap where switching the
// firewall on before any rules exist silently blocks all inbound traffic.
func TestCreateNicWarnsAboutEmptyFirewall(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "lan": 2, "firewall_active": true,
	})
	if !strings.Contains(resultText(res), "blocked") {
		t.Errorf("preview should warn that an empty active firewall blocks traffic:\n%s", resultText(res))
	}
}

func TestCreateNicRequiresLan(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "lan": 0,
	})
	if !res.IsError {
		t.Fatal("lan 0 should be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// TestUpdateNicPreservesLan is the most important test in this file. NicProperties
// serializes lan unconditionally, so a PATCH built without it sends "lan":0 and
// moves the NIC off its LAN as a side effect of an unrelated change. update_nic
// therefore reads the current lan first and sends it back unchanged.
func TestUpdateNicPreservesLan(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nicsAPI+"/nic-1", `{"id":"nic-1","properties":{"name":"eth0","lan":7}}`)

	res := callTool(t, h, "update_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	reqs := h.log.allRequests()
	if len(reqs) != 2 {
		t.Fatalf("expected a GET then a PATCH, got %d requests: %+v", len(reqs), reqs)
	}
	if reqs[0].Method != http.MethodGet {
		t.Errorf("first request should read the NIC's current lan, got %s", reqs[0].Method)
	}
	patch := reqs[1]
	if patch.Method != http.MethodPatch {
		t.Fatalf("second request should be the PATCH, got %s", patch.Method)
	}
	if !strings.Contains(patch.Body, `"lan":7`) {
		t.Errorf("PATCH must carry the NIC's existing lan 7 forward:\n%s", patch.Body)
	}
	if strings.Contains(patch.Body, `"lan":0`) {
		t.Errorf("PATCH must never send lan 0, which would move the NIC off its LAN:\n%s", patch.Body)
	}
	if !strings.Contains(patch.Body, `"name":"renamed"`) {
		t.Errorf("PATCH missing the requested rename:\n%s", patch.Body)
	}
}

// TestUpdateNicExplicitLanSkipsRead is the other half: when the caller does ask to
// move the NIC, no read is needed.
func TestUpdateNicExplicitLanSkipsRead(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "lan": 9,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if !strings.Contains(req.Body, `"lan":9`) {
		t.Errorf("PATCH should carry the requested lan 9:\n%s", req.Body)
	}
}

func TestUpdateNicNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(nicsAPI+"/nic-1", http.StatusNotFound, `{"messages":[{"message":"not found"}]}`)

	res := callTool(t, h, "update_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "name": "renamed",
	})
	if !res.IsError {
		t.Fatal("updating a missing NIC should report an error")
	}
	if !strings.Contains(resultText(res), "does not exist") {
		t.Errorf("want a friendly does-not-exist message, got: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPatch {
			t.Fatal("a 404 on the lan-preserving read must not issue a PATCH")
		}
	}
}

func TestDeleteNicTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nicsAPI+"/nic-1", `{"id":"nic-1","properties":{"name":"eth0","lan":2,"ips":["10.0.0.5"]},
		"entities":{"firewallrules":{"items":[{"id":"fw1"},{"id":"fw2"}]},"flowlogs":{"items":[{"id":"fl1"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1",
	})

	for _, want := range []string{"eth0", "10.0.0.5", "2 firewall rules", "1 flow logs"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != nicsAPI+"/nic-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, nicsAPI+"/nic-1")
	}
}

// TestDeleteNicTokenBoundToParents proves the target includes the full parent
// chain, so a token minted for a NIC under one server cannot delete a
// same-numbered NIC under another.
func TestDeleteNicTokenBoundToParents(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nicsAPI+"/nic-1", `{"id":"nic-1","properties":{"name":"eth0","lan":2}}`)

	res := callTool(t, h, "delete_nic", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1",
	})
	token := extractToken(t, resultText(res))
	h.log.clear()

	res = callTool(t, h, "delete_nic", map[string]any{
		"datacenter_id": dcID, "server_id": "srv-OTHER", "nic_id": "nic-1",
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token minted under srv-1 must not delete a NIC under srv-OTHER")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("a mismatched token must not issue a DELETE")
		}
	}
}

// ---------- create/update/delete LAN ----------

func TestCreateLanTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_lan", map[string]any{
		"datacenter_id": dcID, "name": "public-lan", "public": true,
	})

	for _, want := range []string{"public-lan", "true"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != lansAPI {
		t.Errorf("POST path = %s, want %s", req.Path, lansAPI)
	}
	for _, want := range []string{`"name":"public-lan"`, `"public":true`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateLanUnnamed covers the LAN-specific case that name is optional, so the
// confirmation target needs a stand-in rather than an empty string.
func TestCreateLanUnnamed(t *testing.T) {
	h := destructiveSetup(t)
	_, res := previewThenExecute(t, h, "create_lan", map[string]any{"datacenter_id": dcID})
	if res.IsError {
		t.Fatalf("an unnamed LAN should still be creatable: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if strings.Contains(req.Body, `"name"`) {
		t.Errorf("an omitted name must not be sent as an empty string:\n%s", req.Body)
	}
}

func TestUpdateLan(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_lan", map[string]any{
		"datacenter_id": dcID, "lan_id": "3", "public": false,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != lansAPI+"/3" {
		t.Errorf("PATCH path = %s, want %s", req.Path, lansAPI+"/3")
	}
	if !strings.Contains(req.Body, `"public":false`) {
		t.Errorf("PATCH body missing public:\n%s", req.Body)
	}
	// ipv4_cidr_block is read-only; the tool exposes no way to send it.
	if strings.Contains(req.Body, "ipv4CidrBlock") {
		t.Errorf("PATCH must not send the read-only ipv4CidrBlock:\n%s", req.Body)
	}
}

func TestDeleteLanCountsAttachedNics(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(lansAPI+"/3", `{"id":"3","properties":{"name":"public-lan","public":true},
		"entities":{"nics":{"items":[{"id":"n1"},{"id":"n2"},{"id":"n3"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_lan", map[string]any{
		"datacenter_id": dcID, "lan_id": "3",
	})

	// The NICs are not deleted, but they go dark — the count is the point.
	if !strings.Contains(preview, "3 NICs that will lose their network connection") {
		t.Errorf("preview should count the NICs losing connectivity:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != lansAPI+"/3" {
		t.Errorf("DELETE path = %s, want %s", req.Path, lansAPI+"/3")
	}
}

// ---------- cross-cutting ----------

// TestUpdateBodiesContainOnlyRequestedFields is the regression guard for a whole
// class of silent-corruption bug. The SDK's generated New*Properties[WithDefaults]
// constructors pre-set documented API defaults — NicProperties gets dhcp=true,
// VolumeProperties gets exposeSerial=false, requireLegacyBios=true and
// bootOrder="AUTO". Building a PATCH body from one of those constructors sends
// those defaults as though the caller had asked for them, so renaming a volume
// would also force legacy BIOS on and reset its boot order, which can stop a
// server booting. Update handlers must therefore build a zero-valued literal and
// set only the fields the caller supplied.
//
// Asserting on the exact JSON key set (rather than probing for known-bad keys)
// means a future SDK bump that adds a new default is caught too.
func TestUpdateBodiesContainOnlyRequestedFields(t *testing.T) {
	tests := []struct {
		tool     string
		args     map[string]any
		wantKeys []string
		// prime pins a GET response for handlers that read before writing.
		primePath string
		primeBody string
	}{
		{
			tool:     "update_server",
			args:     map[string]any{"datacenter_id": dcID, "server_id": srvID, "name": "renamed"},
			wantKeys: []string{"name"},
		},
		{
			tool:     "update_volume",
			args:     map[string]any{"datacenter_id": dcID, "volume_id": "vol-1", "name": "renamed"},
			wantKeys: []string{"name"},
		},
		{
			tool:     "update_lan",
			args:     map[string]any{"datacenter_id": dcID, "lan_id": "3", "name": "renamed"},
			wantKeys: []string{"name"},
		},
		{
			// update_nic legitimately also sends lan, because the SDK serializes
			// that field unconditionally and omitting it would mean lan=0.
			tool:      "update_nic",
			args:      map[string]any{"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "name": "renamed"},
			wantKeys:  []string{"name", "lan"},
			primePath: nicsAPI + "/nic-1",
			primeBody: `{"id":"nic-1","properties":{"name":"eth0","lan":7}}`,
		},
		{
			tool:     "update_datacenter",
			args:     map[string]any{"datacenter_id": dcID, "name": "renamed"},
			wantKeys: []string{"name"},
		},
		{
			// update_security_group also carries name, for the same reason.
			tool:      "update_security_group",
			args:      map[string]any{"datacenter_id": dcID, "security_group_id": "sg-1", "description": "new"},
			wantKeys:  []string{"name", "description"},
			primePath: "/cloudapi/v6/datacenters/dc-1/securitygroups/sg-1",
			primeBody: `{"id":"sg-1","properties":{"name":"web-sg","description":"old"}}`,
		},
		{
			tool: "update_firewall_rule",
			args: map[string]any{
				"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1",
				"firewallrule_id": "fw-1", "name": "renamed",
			},
			wantKeys: []string{"name"},
		},
		{
			tool: "update_security_group_rule",
			args: map[string]any{
				"datacenter_id": dcID, "security_group_id": "sg-1", "rule_id": "r-1", "name": "renamed",
			},
			wantKeys: []string{"name"},
		},
		{
			tool:     "update_pcc",
			args:     map[string]any{"pcc_id": "pcc-1", "description": "new"},
			wantKeys: []string{"description"},
		},
		{
			tool:     "update_snapshot",
			args:     map[string]any{"snapshot_id": "snap-1", "name": "renamed"},
			wantKeys: []string{"name"},
		},
		{
			// update_image also carries licenceType, which the SDK always sends.
			tool:      "update_image",
			args:      map[string]any{"image_id": "img-1", "description": "new"},
			wantKeys:  []string{"description", "licenceType"},
			primePath: "/cloudapi/v6/images/img-1",
			primeBody: `{"id":"img-1","properties":{"name":"custom","licenceType":"LINUX"}}`,
		},
		{
			tool:     "update_loadbalancer",
			args:     map[string]any{"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "renamed"},
			wantKeys: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			if tt.primePath != "" {
				h.resp.serve(tt.primePath, tt.primeBody)
			}
			res := callTool(t, h, tt.tool, tt.args)
			if res.IsError {
				t.Fatalf("%s failed: %s", tt.tool, resultText(res))
			}

			var patch recordedRequest
			found := false
			for _, r := range h.log.allRequests() {
				if r.Method == http.MethodPatch {
					patch, found = r, true
				}
			}
			if !found {
				t.Fatalf("%s issued no PATCH", tt.tool)
			}

			var body map[string]any
			if err := json.Unmarshal([]byte(patch.Body), &body); err != nil {
				t.Fatalf("%s PATCH body is not JSON (%v): %s", tt.tool, err, patch.Body)
			}
			// Some resources wrap properties; unwrap so the comparison is on the
			// property set either way.
			if props, ok := body["properties"].(map[string]any); ok {
				body = props
			}

			want := map[string]bool{}
			for _, k := range tt.wantKeys {
				want[k] = true
			}
			for k := range body {
				if !want[k] {
					t.Errorf("%s PATCH body contains unrequested field %q — an SDK-injected default would silently overwrite it on the resource. Body: %s",
						tt.tool, k, patch.Body)
				}
			}
			for k := range want {
				if _, ok := body[k]; !ok {
					t.Errorf("%s PATCH body missing requested field %q: %s", tt.tool, k, patch.Body)
				}
			}
		})
	}
}

// TestCoreWriteToolsScopeGating pins which of the new tools each scope exposes.
func TestCoreWriteToolsScopeGating(t *testing.T) {
	ctx := context.Background()
	creates := []string{"create_server", "create_volume", "create_nic", "create_lan"}
	updates := []string{"update_server", "update_volume", "update_nic", "update_lan"}
	deletes := []string{"delete_server", "delete_volume", "delete_nic", "delete_lan"}

	readOnly := toolNames(t, ctx, setup(t))
	for _, name := range append(append(append([]string{}, creates...), updates...), deletes...) {
		if readOnly[name] {
			t.Errorf("read-only scope must not expose %q", name)
		}
	}

	write := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true}))
	for _, name := range append(append([]string{}, creates...), updates...) {
		if !write[name] {
			t.Errorf("write scope should expose %q", name)
		}
	}
	for _, name := range deletes {
		if write[name] {
			t.Errorf("write scope must not expose destructive %q", name)
		}
	}

	destructive := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	for _, name := range deletes {
		if !destructive[name] {
			t.Errorf("destructive scope should expose %q", name)
		}
	}
}

// TestCoreWriteDynamicParity runs a full two-phase delete through ionos_call_tool,
// proving the shared confirmation store is reached identically in dynamic mode.
func TestCoreWriteDynamicParity(t *testing.T) {
	h := setupDynamicWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()
	h.resp.serve(lansAPI+"/3", `{"id":"3","properties":{"name":"public-lan"}}`)

	call := func(args map[string]any) *mcp.CallToolResult {
		res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "ionos_call_tool",
			Arguments: map[string]any{"name": "delete_lan", "arguments": args},
		})
		if err != nil {
			t.Fatalf("ionos_call_tool: %v", err)
		}
		return res
	}

	res := call(map[string]any{"datacenter_id": dcID, "lan_id": "3"})
	if res.IsError {
		t.Fatalf("dynamic preview failed: %s", resultText(res))
	}
	token := extractToken(t, resultText(res))
	h.log.clear()

	res = call(map[string]any{"datacenter_id": dcID, "lan_id": "3", "confirmation_token": token})
	if res.IsError {
		t.Fatalf("dynamic execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}
