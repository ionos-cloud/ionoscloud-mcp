package activitylog

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterEventTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_activitylog_events",
		Description: "Query the IONOS CLOUD activity log: full audit trail of API requests made against a contract (who did what, when, on which resource). " +
			"Requires ACCESS_ACTIVITY_LOG privilege on the token. " +
			"ALWAYS narrow results with date_start and date_end — logs span years with thousands of events per day. " +
			"Pass a small limit (e.g. 25) unless the user explicitly asks for bulk data. " +
			"Response is compacted: _source wrapper removed, auditVersion dropped, redundant contractNumber and duplicate sourceService fields stripped. " +
			"Use list_activitylog_contracts first to look up the contract number if needed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ActivityLogQueryInput) (*mcp.CallToolResult, any, error) {
		if in.DateStart != nil {
			if err := tools.ValidateDate(*in.DateStart); err != nil {
				return tools.ToResult(nil, err)
			}
		}
		if in.DateEnd != nil {
			if err := tools.ValidateDate(*in.DateEnd); err != nil {
				return tools.ToResult(nil, err)
			}
		}

		req := client.ContractsApi.GetByContract(ctx, in.Contract)
		if in.DateStart != nil {
			req = req.DateStart(*in.DateStart)
		}
		if in.DateEnd != nil {
			req = req.DateEnd(*in.DateEnd)
		}
		if in.Offset != nil {
			req = req.Offset(*in.Offset)
		}
		if in.Limit != nil {
			req = req.Limit(*in.Limit)
		}

		raw, _, err := req.Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(Compact(raw, in.Contract), nil)
	})
}
