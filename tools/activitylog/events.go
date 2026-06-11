package activitylog

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

const (
	defaultLimit    = int32(25)
	defaultLookback = -7 * 24 * time.Hour
	maxRangeDays    = 90
)

func RegisterEventTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_activitylog_events",
		Annotations: tools.ReadOnly,
		Description: "Query the IONOS CLOUD activity log: full audit trail of API requests made against a contract (who did what, when, on which resource). " +
			"Requires ACCESS_ACTIVITY_LOG privilege on the token. " +
			"Defaults: last 7 days, limit 25, RequestStatusUpdate events excluded. " +
			"Use user filter to narrow to a specific account. Use event_types to restrict to e.g. ['Error','RequestAccepted']. " +
			"Maximum date range is 90 days — paginate or narrow the window for longer spans. " +
			"Use list_activitylog_contracts first to look up the contract number if needed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.ActivityLogQueryInput) (*mcp.CallToolResult, any, error) {
		now := time.Now().UTC()

		// Resolve and validate date_start
		dateStart := now.Add(defaultLookback).Format("2006-01-02")
		if in.DateStart != nil {
			if err := tools.ValidateDate(*in.DateStart); err != nil {
				return tools.ToResult(nil, err)
			}
			dateStart = strings.TrimSpace(*in.DateStart)
		}

		// Resolve and validate date_end
		dateEnd := now.Format("2006-01-02")
		if in.DateEnd != nil {
			if err := tools.ValidateDate(*in.DateEnd); err != nil {
				return tools.ToResult(nil, err)
			}
			dateEnd = strings.TrimSpace(*in.DateEnd)
		}

		// Enforce 90-day max range
		start, err := time.Parse("2006-01-02", dateStart)
		if err != nil {
			return tools.ToResult(nil, fmt.Errorf("failed to parse date_start %q: %w", dateStart, err))
		}
		end, err := time.Parse("2006-01-02", dateEnd)
		if err != nil {
			return tools.ToResult(nil, fmt.Errorf("failed to parse date_end %q: %w", dateEnd, err))
		}
		if end.Before(start) {
			return tools.ToResult(nil, fmt.Errorf("date_end %q is before date_start %q", dateEnd, dateStart))
		}
		inclusiveDays := int(end.Sub(start).Hours()/24) + 1
		if inclusiveDays > maxRangeDays {
			return tools.ToResult(nil, fmt.Errorf("date range exceeds %d days (%s to %s is %d days); narrow the window or paginate", maxRangeDays, dateStart, dateEnd, inclusiveDays))
		}

		// Validate and apply limit
		limit := defaultLimit
		if in.Limit != nil {
			if *in.Limit <= 0 {
				return tools.ToResult(nil, fmt.Errorf("limit must be greater than 0, got %d", *in.Limit))
			}
			limit = *in.Limit
		}

		// Validate and apply offset
		if in.Offset != nil && *in.Offset < 0 {
			return tools.ToResult(nil, fmt.Errorf("offset must be >= 0, got %d", *in.Offset))
		}

		req := client.ContractsApi.GetByContract(ctx, in.Contract).
			DateStart(dateStart).
			DateEnd(dateEnd).
			Limit(limit)

		if in.Offset != nil {
			req = req.Offset(*in.Offset)
		}

		raw, _, err := req.Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}

		opts := CompactOptions{
			IncludeStatusUpdates: in.IncludeStatusUpdates != nil && *in.IncludeStatusUpdates,
			User:                 in.User,
			EventTypes:           in.EventTypes,
		}
		return tools.ToResult(Compact(raw, in.Contract, opts), nil)
	})
}
