package test

import (
	"context"
	"testing"

	activitylogpkg "github.com/ionos-cloud/ionoscloud-mcp/tools/activitylog"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestActivityLogToolEndpoints(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	c := int32(12345678)

	tests := []toolTest{
		{
			name:        "list_activitylog_contracts",
			args:        map[string]any{},
			wantMethods: []string{"GET"},
			wantPaths:   []string{"/activitylog/v1/contracts"},
		},
		{
			name:        "list_activitylog_events",
			args:        map[string]any{"contract": c},
			wantMethods: []string{"GET"},
			wantPaths:   []string{"/activitylog/v1/contracts/12345678"},
		},
	}

	// date-filtered and paginated calls still route to the same path
	dateTests := []struct {
		subtest string
		args    map[string]any
	}{
		{"with_dates", map[string]any{"contract": c, "date_start": "2026-05-01", "date_end": "2026-05-11"}},
		{"with_pagination", map[string]any{"contract": c, "offset": int32(0), "limit": int32(25)}},
	}

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

	for _, tt := range dateTests {
		t.Run(tt.subtest, func(t *testing.T) {
			h.log.clear()

			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "list_activitylog_events",
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool returned protocol error: %v", err)
			}

			reqs := h.log.allRequests()
			if len(reqs) != 1 {
				t.Fatalf("want 1 request, got %d", len(reqs))
			}
			if reqs[0].Path != "/activitylog/v1/contracts/12345678" {
				t.Errorf("path = %q, want %q", reqs[0].Path, "/activitylog/v1/contracts/12345678")
			}
		})
	}
}

func TestActivityLogDateValidation(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	c := int32(1)

	cases := []struct {
		name string
		args map[string]any
	}{
		{"bad date_start", map[string]any{"contract": c, "date_start": "not-a-date"}},
		{"bad date_end", map[string]any{"contract": c, "date_end": "2026-13-01"}},
		{"wrong format date_start", map[string]any{"contract": c, "date_start": "01-05-2026"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()

			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "list_activitylog_events",
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool returned protocol error: %v", err)
			}

			reqs := h.log.allRequests()
			if len(reqs) != 0 {
				t.Errorf("invalid date made %d HTTP requests, want 0", len(reqs))
			}
		})
	}
}

func TestCompact(t *testing.T) {
	contract := int32(31909628)

	str := func(s string) *string { return &s }
	i32 := func(n int32) *int32 { return &n }

	t.Run("empty response", func(t *testing.T) {
		out := activitylogpkg.Compact(sdk.GetByContractResponse{}, contract, activitylogpkg.CompactOptions{IncludeStatusUpdates: true})
		if len(out.Events) != 0 {
			t.Errorf("want 0 events, got %d", len(out.Events))
		}
		if out.Total != 0 {
			t.Errorf("want total 0, got %d", out.Total)
		}
	})

	t.Run("total propagated", func(t *testing.T) {
		raw := sdk.GetByContractResponse{
			Hits: &sdk.GetByContractResponseHits{
				Total: i32(42),
			},
		}
		out := activitylogpkg.Compact(raw, contract, activitylogpkg.CompactOptions{IncludeStatusUpdates: true})
		if out.Total != 42 {
			t.Errorf("want total 42, got %d", out.Total)
		}
	})

	t.Run("drops auditVersion and contractNumber", func(t *testing.T) {
		auditVer := float32(0.1)
		raw := sdk.GetByContractResponse{
			Hits: &sdk.GetByContractResponseHits{
				Total: i32(1),
				Hits: []sdk.GetByContractResponseHitsHits{
					{
						Source: &sdk.GetByContractResponseHitsHitsSource{
							Meta: &sdk.GetByContractResponseHitsHitsSourceMeta{
								AuditVersion: &auditVer,
								Time:         str("2026-05-01T07:42:20.530Z"),
								RequestId:    str("req-uuid-1"),
							},
							Principal: &sdk.GetByContractResponseHitsHitsSourcePrincipal{
								SourceService: str("PUBLIC_REST_V6"),
								Identity: &sdk.GetByContractResponseHitsHitsSourcePrincipalIdentity{
									ContractNumber: &contract,
									Username:       str("user@example.com"),
								},
							},
							Event: &sdk.GetByContractResponseHitsHitsSourceEvent{
								Type: str("RequestAccepted"),
								Param: &sdk.GetByContractResponseHitsHitsSourceEventParam{
									Action: str("GET"),
								},
							},
						},
					},
				},
			},
		}

		out := activitylogpkg.Compact(raw, contract, activitylogpkg.CompactOptions{IncludeStatusUpdates: true})
		if len(out.Events) != 1 {
			t.Fatalf("want 1 event, got %d", len(out.Events))
		}
		ev := out.Events[0]

		if ev.Time != "2026-05-01T07:42:20.530Z" {
			t.Errorf("time = %q, want %q", ev.Time, "2026-05-01T07:42:20.530Z")
		}
		if ev.RequestID != "req-uuid-1" {
			t.Errorf("request_id = %q, want %q", ev.RequestID, "req-uuid-1")
		}
		if ev.User != "user@example.com" {
			t.Errorf("user = %q, want %q", ev.User, "user@example.com")
		}
		if ev.Service != "PUBLIC_REST_V6" {
			t.Errorf("service = %q, want %q", ev.Service, "PUBLIC_REST_V6")
		}
		if ev.Action != "GET" {
			t.Errorf("action = %q, want %q", ev.Action, "GET")
		}
		// auditVersion is dropped — no field to check
		// contractNumber is dropped — user should not contain it
		if ev.User == "31909628" {
			t.Errorf("contractNumber leaked into User field")
		}
	})

	t.Run("drops redundant param sourceService", func(t *testing.T) {
		raw := sdk.GetByContractResponse{
			Hits: &sdk.GetByContractResponseHits{
				Total: i32(1),
				Hits: []sdk.GetByContractResponseHitsHits{
					{
						Source: &sdk.GetByContractResponseHitsHitsSource{
							Meta: &sdk.GetByContractResponseHitsHitsSourceMeta{
								Time: str("2026-05-01T07:00:00Z"),
							},
							Principal: &sdk.GetByContractResponseHitsHitsSourcePrincipal{
								SourceService: str("PUBLIC_REST"),
								Identity:      &sdk.GetByContractResponseHitsHitsSourcePrincipalIdentity{},
							},
							Event: &sdk.GetByContractResponseHitsHitsSourceEvent{
								Type: str("RequestStatusUpdate"),
								Param: &sdk.GetByContractResponseHitsHitsSourceEventParam{
									Initiator:     str("PUBLIC_REST"),
									SourceService: str("PUBLIC_REST"),
								},
							},
						},
					},
				},
			},
		}

		out := activitylogpkg.Compact(raw, contract, activitylogpkg.CompactOptions{IncludeStatusUpdates: true})
		if len(out.Events) != 1 {
			t.Fatalf("want 1 event, got %d", len(out.Events))
		}
		ev := out.Events[0]
		if ev.Initiator != "" {
			t.Errorf("Initiator should be stripped when equal to principal.sourceService, got %q", ev.Initiator)
		}
		if ev.ParamSvc != "" {
			t.Errorf("ParamSvc should be stripped when equal to principal.sourceService, got %q", ev.ParamSvc)
		}
	})

	t.Run("keeps non-redundant initiator", func(t *testing.T) {
		raw := sdk.GetByContractResponse{
			Hits: &sdk.GetByContractResponseHits{
				Total: i32(1),
				Hits: []sdk.GetByContractResponseHitsHits{
					{
						Source: &sdk.GetByContractResponseHitsHitsSource{
							Meta: &sdk.GetByContractResponseHitsHitsSourceMeta{
								Time: str("2026-05-01T07:00:00Z"),
							},
							Principal: &sdk.GetByContractResponseHitsHitsSourcePrincipal{
								SourceService: str("PUBLIC_REST_V6"),
								Identity:      &sdk.GetByContractResponseHitsHitsSourcePrincipalIdentity{},
							},
							Event: &sdk.GetByContractResponseHitsHitsSourceEvent{
								Type: str("RequestStatusUpdate"),
								Param: &sdk.GetByContractResponseHitsHitsSourceEventParam{
									Initiator: str("PUBLIC_REST"),
								},
							},
						},
					},
				},
			},
		}

		out := activitylogpkg.Compact(raw, contract, activitylogpkg.CompactOptions{IncludeStatusUpdates: true})
		if len(out.Events) != 1 {
			t.Fatalf("want 1 event, got %d", len(out.Events))
		}
		if out.Events[0].Initiator != "PUBLIC_REST" {
			t.Errorf("Initiator should be kept when different from service, got %q", out.Events[0].Initiator)
		}
	})

	t.Run("resources with empty action stripped", func(t *testing.T) {
		raw := sdk.GetByContractResponse{
			Hits: &sdk.GetByContractResponseHits{
				Total: i32(1),
				Hits: []sdk.GetByContractResponseHitsHits{
					{
						Source: &sdk.GetByContractResponseHitsHitsSource{
							Meta:      &sdk.GetByContractResponseHitsHitsSourceMeta{Time: str("2026-05-01T07:00:00Z")},
							Principal: &sdk.GetByContractResponseHitsHitsSourcePrincipal{},
							Event: &sdk.GetByContractResponseHitsHitsSourceEvent{
								Type: str("RequestAccepted"),
								Resources: []sdk.GetByContractResponseHitsHitsSourceEventResources{
									{Type: str("datacenter"), Id: str("dc-uuid"), Action: []string{}},
									{Type: str("user"), Id: str("user-uuid"), Action: []string{"sec.user.create"}},
								},
							},
						},
					},
				},
			},
		}

		out := activitylogpkg.Compact(raw, contract, activitylogpkg.CompactOptions{IncludeStatusUpdates: true})
		if len(out.Events) != 1 {
			t.Fatalf("want 1 event, got %d", len(out.Events))
		}
		res := out.Events[0].Resources
		if len(res) != 2 {
			t.Fatalf("want 2 resources, got %d", len(res))
		}
		if len(res[0].Action) != 0 {
			t.Errorf("empty action should be nil/empty, got %v", res[0].Action)
		}
		if len(res[1].Action) != 1 || res[1].Action[0] != "sec.user.create" {
			t.Errorf("non-empty action should be preserved, got %v", res[1].Action)
		}
	})
}
