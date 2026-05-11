package loader

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterLoaderTools adds three always-visible tools to the server:
//   - ionos_load_compute_tools: dynamically registers all 50 Compute tools
//   - ionos_load_objectstorage_tools: dynamically registers all 23 Object Storage tools
//   - ionos_list_available_domains: shows loaded vs deferred domains with tool counts
//
// These tools are safe to register regardless of which domains were loaded at startup.
// Loader tools for already-loaded domains simply report "already loaded" when called.
func RegisterLoaderTools(server *mcp.Server, dl *DomainLoader) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "ionos_load_compute_tools",
		Description: `Load IONOS Compute Engine tools (50 tools) into this session.

Call this tool before managing any virtual infrastructure. After it returns, the
following tools become available:

Datacenters : list_datacenters, get_datacenter
Servers     : list_servers, get_server, list_server_volumes, list_server_cdroms,
              list_server_gpus, get_server_gpu, get_server_remote_console
Volumes     : list_volumes, get_volume
NICs        : list_nics, get_nic
LANs        : list_lans, get_lan, list_lan_nics
Firewall    : list_firewall_rules, get_firewall_rule
IP Blocks   : list_ip_blocks, get_ip_block
LBs         : list_loadbalancers, get_loadbalancer, list_loadbalancer_nics
Network LBs : list_network_loadbalancers, get_network_loadbalancer, list_nlb_forwarding_rules
App LBs     : list_application_loadbalancers, get_application_loadbalancer, list_alb_forwarding_rules
Target Grps : list_target_groups, get_target_group
NAT         : list_nat_gateways, get_nat_gateway, list_nat_gateway_rules
Cross-Conn  : list_private_cross_connects, get_private_cross_connect
Sec Groups  : list_security_groups, get_security_group, list_security_group_rules, get_security_group_rule
Other       : get_contract, list_requests, get_request, get_request_status,
              list_templates, get_template, list_images, list_locations,
              list_snapshots, get_snapshot`,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		newly, err := dl.Load(DomainCompute)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		var text string
		if newly {
			text = "Compute tools loaded (50 tools). The tool list has been updated — you can now call list_servers, list_datacenters, and all other compute tools listed above."
		} else {
			text = "Compute tools are already loaded."
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "ionos_load_objectstorage_tools",
		Description: `Load IONOS Object Storage tools (23 tools) into this session.

Call this tool before managing S3-compatible object storage. After it returns, the
following tools become available:

Buckets     : list_object_storage_buckets, head_object_storage_bucket,
              get_object_storage_bucket_location
Bucket Cfg  : get_object_storage_bucket_cors, get_object_storage_bucket_encryption,
              get_object_storage_bucket_lifecycle, get_object_storage_bucket_lock_configuration,
              get_object_storage_bucket_policy, get_object_storage_bucket_policy_status,
              get_object_storage_bucket_public_access_block, get_object_storage_bucket_replication,
              get_object_storage_bucket_tagging, get_object_storage_bucket_versioning
Objects     : list_object_storage_objects, head_object_storage_object,
              list_object_storage_object_versions, get_object_storage_object_legal_hold,
              get_object_storage_object_retention, get_object_storage_object_tagging
Access Keys : list_object_storage_access_keys, get_object_storage_access_key
Regions     : list_object_storage_regions, get_object_storage_region`,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		newly, err := dl.Load(DomainObjectStorage)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		var text string
		if newly {
			text = "Object Storage tools loaded (23 tools). The tool list has been updated — you can now call list_object_storage_buckets and all other object storage tools listed above."
		} else {
			text = "Object Storage tools are already loaded."
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ionos_list_available_domains",
		Description: "List IONOS Cloud tool domains and their loading status. Shows which domains are already active and which are deferred, with tool counts and the loader tool name to enable each deferred domain.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		type deferredEntry struct {
			Domain      string `json:"domain"`
			ToolCount   int    `json:"tool_count"`
			Description string `json:"description"`
			LoadTool    string `json:"load_tool"`
		}
		type response struct {
			Loaded   []string        `json:"loaded"`
			Deferred []deferredEntry `json:"deferred"`
		}

		resp := response{
			Loaded:   []string{},
			Deferred: []deferredEntry{},
		}
		for _, d := range AllDomains {
			meta := domainMeta[d]
			if dl.IsLoaded(d) {
				resp.Loaded = append(resp.Loaded, string(d))
			} else {
				resp.Deferred = append(resp.Deferred, deferredEntry{
					Domain:      string(d),
					ToolCount:   meta.ToolCount,
					Description: meta.Description,
					LoadTool:    meta.LoadTool,
				})
			}
		}

		b, err := json.Marshal(resp)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal response: %w", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
		}, nil, nil
	})
}
