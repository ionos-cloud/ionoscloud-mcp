package compute

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterIpBlockWriteTools registers the create and delete IP block tools. IP
// blocks are account-level, so these take no datacenter_id.
//
// There is no update_ip_block: the API forbids location in update requests, but the
// SDK always serializes it, so no typed call can produce an acceptable body.
func RegisterIpBlockWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateIpBlock(server, client, scope, confirm)
	registerDeleteIpBlock(server, client, scope, confirm)
}

func registerCreateIpBlock(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_ip_block",
		Description: "Reserve a block of public IPv4 addresses. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same location and size) to reserve it. " +
			"An IP block belongs to the account rather than to a data center, but its location must match the data center whose resources will use the addresses. " +
			"Both location and size are fixed once reserved — to change either, reserve a new block and release this one. The block is billed from the moment it exists, whether or not the addresses are assigned to anything. Reserves exactly one block per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateIpBlockInput) (*mcp.CallToolResult, any, error) {
		location := strings.TrimSpace(input.Location)
		if location == "" {
			return tools.ErrorText("location is required to reserve an IP block (e.g. de/fra); it must match the location of the data center that will use the addresses"), nil, nil
		}
		if input.Size <= 0 {
			return tools.ErrorText("size must be at least 1: it is the number of public IPv4 addresses to reserve"), nil, nil
		}
		target := tools.Target(location, strconv.FormatInt(int64(input.Size), 10))

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_ip_block", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_ip_block", "location and size", err)), nil, nil
			}
			props := ionos.NewIpBlockProperties(location, input.Size)
			if input.Name != nil {
				props.SetName(*input.Name)
			}
			body := ionos.NewIpBlock(*props)
			created, _, err := client.IPBlocksApi.IpblocksPost(ctx).Ipblock(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_ip_block", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to RESERVE one block of public IPv4 addresses:",
			Fields: tools.Fields(
				"location", location,
				"size (addresses)", strconv.FormatInt(int64(input.Size), 10),
				"name", tools.OptStr(input.Name),
			),
			Tool:      "create_ip_block",
			Replay:    tools.Fields("location", location, "size", strconv.FormatInt(int64(input.Size), 10)),
			TokenNote: "The block is billed from creation even while its addresses are unused, and neither location nor size can be changed afterwards. The token authorizes reserving only this location+size",
		}.Render(token)), nil, nil
	})
}

func registerDeleteIpBlock(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_ip_block",
		Description: "Release a block of public IPv4 addresses. Two-phase: call first WITHOUT confirmation_token to get a preview listing every resource still using the addresses, plus a one-time token, then call again WITH the token to release it. " +
			"Releasing a block whose addresses are still assigned breaks connectivity for those resources, and the addresses go back to the pool — you cannot reclaim the same ones. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteIpBlockInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.IpBlockID)
		if id == "" {
			return tools.ErrorText("ipblock_id is required"), nil, nil
		}
		target := tools.Target(id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_ip_block", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_ip_block", "ipblock_id", err)), nil, nil
			}
			_, err := client.IPBlocksApi.IpblocksDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("IP block", id)), nil, nil
		}

		// Phase 1: no token -> report who is using the addresses, then mint a token.
		block, _, err := client.IPBlocksApi.IpblocksFindById(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("IP block %s does not exist; nothing to release", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := block.GetProperties()
		radius, inUse := ipBlockBlastRadius(props)
		token, mErr := confirm.Mint("delete_ip_block", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to RELEASE a block of public IPv4 addresses. This is IRREVERSIBLE — a block is requested by location and size only, so there is no way to ask for these same addresses back."
		if inUse > 0 {
			headline += fmt.Sprintf("\nWARNING: %d of these addresses are still assigned. Releasing the block breaks connectivity for the resources listed below.", inUse)
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"ipblock_id", id,
				"name", props.GetName(),
				"location", props.GetLocation(),
				"size (addresses)", strconv.FormatInt(int64(props.GetSize()), 10),
				"addresses", ipSummary(props.GetIps()),
			),
			Radius:    radius,
			EmptyNote: "None of these addresses are currently assigned to a resource.",
			Tool:      "delete_ip_block",
			Replay:    tools.Fields("ipblock_id", id),
			TokenNote: "This token authorizes releasing ONLY this IP block",
		}.Render(token)), nil, nil
	})
}

// ipBlockBlastRadius reports which resources are using the block's addresses, and
// how many addresses are assigned.
func ipBlockBlastRadius(props ionos.IpBlockProperties) (*tools.BlastRadius, int) {
	r := tools.AffectedRadius()
	consumers := props.GetIpConsumers()
	if len(consumers) == 0 {
		return r, 0
	}

	// One consumer entry may name a NIC, a server and a node pool at once, so
	// count distinct IDs per kind.
	servers, nics, k8s := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, c := range consumers {
		if id := c.GetServerId(); id != "" {
			servers[id] = true
		}
		if id := c.GetNicId(); id != "" {
			nics[id] = true
		}
		if id := c.GetK8sNodePoolUuid(); id != "" {
			k8s[id] = true
		}
	}
	r.Add("addresses currently assigned", len(consumers))
	r.Add("servers that lose these addresses", len(servers))
	r.Add("NICs that lose these addresses", len(nics))
	r.Add("Kubernetes node pools that lose these addresses", len(k8s))
	return r, len(consumers)
}
