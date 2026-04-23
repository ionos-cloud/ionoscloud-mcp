package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fetchBillingRaw calls the IONOS Billing v3 API directly and returns the
// response body unparsed. This bypasses the SDK's typed deserialization, which
// breaks whenever the backend introduces a new ResourceType enum value (e.g.
// "NFS", "ML") that the SDK's generated enum has not yet learned about.
//
// The SDK's own client goes through the same endpoints and token but decodes
// into strongly-typed structs; since MCP tools ultimately return JSON text to
// the LLM anyway, skipping that decode step costs nothing and sidesteps the
// bug entirely. This helper is intentionally self-contained (env-based token)
// so existing tool-registration signatures (which only take *sdk.APIClient)
// do not need to change.
func fetchBillingRaw(ctx context.Context, path string) (json.RawMessage, error) {
	token := os.Getenv("IONOS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("IONOS_TOKEN not set")
	}
	host := os.Getenv("IONOS_API_URL")
	if host == "" {
		host = "https://api.ionos.com"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, host+"/billing"+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("billing API returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	return json.RawMessage(body), nil
}

// RegisterUtilizationTools registers the three billing-utilization tools.
//
// All three bypass sdk.UtilizationApi.* and call the IONOS Billing v3 API
// directly — the SDK's typed deserialization rejects any ResourceType enum
// value it was not generated with, so adding a new product server-side
// (NFS, ML, etc.) silently breaks these endpoints. Raw-HTTP passthrough
// immunises us against future enum additions at the cost of returning
// unstructured JSON, which is fine because MCP tool results are text anyway.
// The sdk.APIClient parameter is retained for signature compatibility with
// RegisterAll but is intentionally unused.
func RegisterUtilizationTools(server *mcp.Server, _ *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_utilization",
		Description: "Get high-granularity resource utilization for your contract for the current billing period. Shows per-resource metrics (CPU, RAM, storage, DNS) grouped by datacenter. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractInput) (*mcp.CallToolResult, any, error) {
		raw, err := fetchBillingRaw(ctx, "/"+strconv.Itoa(int(input.Contract))+"/utilization")
		return tools.ToResult(raw, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_utilization_by_period",
		Description: "Get high-granularity resource utilization for a specific billing period (YYYY-MM). One month per call. If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractPeriodInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidatePeriod(input.Period); err != nil {
			return tools.ToResult(nil, err)
		}
		raw, err := fetchBillingRaw(ctx, "/"+strconv.Itoa(int(input.Contract))+"/utilization/"+input.Period)
		return tools.ToResult(raw, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_utilization_daily",
		Description: "Get high-granularity resource utilization for a specific date (YYYY-MM-DD). Use this for day-level analysis within a month. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingDateInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidateDate(input.Date); err != nil {
			return tools.ToResult(nil, err)
		}
		raw, err := fetchBillingRaw(ctx, "/"+strconv.Itoa(int(input.Contract))+"/utilization/daily/"+input.Date)
		return tools.ToResult(raw, err)
	})
}
