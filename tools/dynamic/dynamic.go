// Package dynamic implements the 'dynamic' load mode: instead of exposing the
// full product tool catalog (110+ tools) to the MCP client, the server exposes
// only three meta-tools — ionos_search_tools, ionos_describe_tools and
// ionos_call_tool — through which the model discovers and invokes the real
// tools at runtime. The real catalog never enters the client's tool list, so
// this works for clients with hard tool caps and no client-side tool search of
// their own (e.g. Cursor, Windsurf), without relying on
// notifications/tools/list_changed.
//
// Mechanism: the full catalog is registered onto a private, in-memory "catalog"
// server (reusing each product's existing RegisterAll, unchanged). The dynamic
// package self-connects to that server over an in-memory transport, snapshots
// the tool metadata once at startup, and forwards ionos_call_tool invocations
// to it. Input validation, schema inference and error enrichment all run inside
// the catalog server exactly as they would in eager mode.
package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Product is one registrable product group (compute, dns, ...). Register adds
// the product's tools to the given server; it is the same closure the eager
// path runs, so the catalog stays in lockstep with eager mode.
type Product struct {
	Name     string
	Summary  string
	Register func(s *mcp.Server)
}

// catalogEntry is the snapshot of a single real tool, captured from the catalog
// server at startup.
type catalogEntry struct {
	Name        string
	Group       string
	Description string
	InputSchema json.RawMessage
	Class       tools.Class // mutation class (read/write/destructive); enforced by callHandler
}

// dispatcher holds the immutable post-startup state shared by the three
// meta-tool handlers: the searchable index and the live session to the private
// catalog server. (Named "dispatcher", not "router", to avoid confusion with
// the retired "router" load mode.)
type dispatcher struct {
	entries    []catalogEntry
	byName     map[string]catalogEntry
	scope      tools.Scope        // enabled tool classes; re-checked in callHandler (defense-in-depth)
	session    *mcp.ClientSession // client side of the self-connection to the catalog
	srvSession *mcp.ServerSession // server side; retained so Close can tear it down
}

const defaultSearchLimit = 10

// Close tears down the private catalog server connection (both halves of the
// in-memory self-connection). main defers it for a clean shutdown when
// server.Run returns; since the dispatcher is a process-lifetime singleton the
// OS would reclaim everything at exit anyway, but tests also rely on it to avoid
// leaking a catalog server + sessions per run. nil-safe on each half.
func (d *dispatcher) Close() error {
	if d.session != nil {
		_ = d.session.Close()
	}
	if d.srvSession != nil {
		_ = d.srvSession.Close()
	}
	return nil
}

// Register builds the private catalog from products and registers the three
// dynamic meta-tools on the public server. It returns an io.Closer that tears
// down the catalog connection; main defers it for a clean shutdown and tests
// Close it to avoid leaks. Returns an error if the catalog cannot be built.
func Register(ctx context.Context, public *mcp.Server, products []Product, scope tools.Scope) (io.Closer, error) {
	d, err := buildCatalog(ctx, products, scope)
	if err != nil {
		return nil, err
	}

	summary := catalogSummary(products, d)

	mcp.AddTool(public, &mcp.Tool{
		Name: "ionos_search_tools",
		Description: "Search the IONOS Cloud tool catalog by keyword. Returns matching tool names, " +
			"their product group, and a one-line description — but NOT their input schemas (use " +
			"ionos_describe_tools for those, then ionos_call_tool to invoke). Leave query empty and " +
			"set group to browse a whole product.\n\n" + summary,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, d.searchHandler)

	mcp.AddTool(public, &mcp.Tool{
		Name: "ionos_describe_tools",
		Description: "Return the full JSON input schema and description for one or more IONOS Cloud " +
			"tools (names come from ionos_search_tools). Call this before ionos_call_tool to learn a " +
			"tool's required and optional arguments.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, d.describeHandler)

	mcp.AddTool(public, &mcp.Tool{
		Name: "ionos_call_tool",
		Description: "Invoke an IONOS Cloud tool by name with the given arguments and return its " +
			"result. The name must be an exact tool name from ionos_search_tools; arguments must match " +
			"the schema from ionos_describe_tools. Most tools are read-only; create_/update_/delete_ tools mutate and are only callable when IONOS_MCP_TOOL_SCOPE enables them.\n\n" +
			"USE OPTIONAL PARAMETERS IF NEEDED. Beyond the required IDs, many tools (especially list_/get_) accept " +
			"optional arguments that are easy to overlook but are often what makes a single call answer " +
			"the request. Always read the tool's schema via ionos_describe_tools first, then pass every " +
			"optional argument the request needs or clearly implies — do not call with only the required " +
			"fields when an optional one fits. Two are especially impactful where a tool's schema offers " +
			"them:\n" +
			"  - filters (on list_ tools): server-side filtering as property→value pairs (contains match). " +
			"Filter at the API instead of listing everything and filtering the result yourself.\n" +
			"  - depth (on both list_ and get_ tools): nesting depth of returned objects (default 1). " +
			"depth 1 returns only the object and its basic properties; depth 2 embeds its child resources in " +
			"the SAME response. Prefer depth 2 so nested resources come back in one round-trip instead of a " +
			"follow-up call per parent; depth 2 is enough for this and higher values add nothing useful. " +
			"When a user asks to browse or inspect a resource, assume they will also want the resources " +
			"nested inside it and default to depth 2.\n" +
			"Other tools expose different optional arguments (filtering, date ranges, pagination, " +
			"aggregation, and so on) — the set varies per tool, so always check the schema and use what the " +
			"request implies.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, d.callHandler)

	return d, nil
}

// searchResult is one row of an ionos_search_tools response.
type searchResult struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

func (r *dispatcher) searchHandler(_ context.Context, _ *mcp.CallToolRequest, in tools.SearchToolsInput) (*mcp.CallToolResult, any, error) {
	limit := defaultSearchLimit
	if in.Limit != nil {
		if *in.Limit < 0 {
			return errorResult(fmt.Sprintf("limit must not be negative (got %d); omit it for the default of %d, or pass 0 for no limit", *in.Limit, defaultSearchLimit)), nil, nil
		}
		// 0 means "no limit": search treats limit<=0 as uncapped.
		limit = *in.Limit
	}
	group := ""
	if in.Group != nil {
		group = strings.ToLower(strings.TrimSpace(*in.Group))
	}

	matches := r.search(in.Query, group, limit)
	tokens := tokenize(in.Query)
	results := make([]searchResult, 0, len(matches))
	for _, e := range matches {
		// Scoring runs over the full description (see search/score), so a tool
		// matched via a later sentence is still found; we only shorten what's
		// shown. Prefer the first sentence that contains a query term, so the
		// snippet always explains the match; fall back to the first sentence.
		results = append(results, searchResult{Name: e.Name, Group: e.Group, Description: snippet(e.Description, tokens)})
	}
	return tools.ToResult(map[string]any{
		"count": len(results),
		"tools": results,
		"hint":  "Descriptions are shortened. Use ionos_describe_tools for a tool's full description and input schema, then ionos_call_tool to invoke it.",
	}, nil)
}

// describedTool is one row of an ionos_describe_tools response.
type describedTool struct {
	Name        string          `json:"name"`
	Group       string          `json:"group,omitempty"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
	Error       string          `json:"error,omitempty"`
	DidYouMean  []string        `json:"did_you_mean,omitempty"`
}

func (r *dispatcher) describeHandler(_ context.Context, _ *mcp.CallToolRequest, in tools.DescribeToolsInput) (*mcp.CallToolResult, any, error) {
	out := make([]describedTool, 0, len(in.Names))
	for _, name := range in.Names {
		name = strings.TrimSpace(name)
		e, ok := r.byName[name]
		if !ok {
			out = append(out, describedTool{
				Name:       name,
				Error:      "no such tool",
				DidYouMean: r.suggest(name),
			})
			continue
		}
		out = append(out, describedTool{
			Name:        e.Name,
			Group:       e.Group,
			Description: e.Description,
			InputSchema: e.InputSchema,
		})
	}
	return tools.ToResult(map[string]any{"tools": out}, nil)
}

func (r *dispatcher) callHandler(ctx context.Context, _ *mcp.CallToolRequest, in tools.CallToolInput) (*mcp.CallToolResult, any, error) {
	name := strings.TrimSpace(in.Name)
	entry, ok := r.byName[name]
	if !ok {
		msg := fmt.Sprintf("no such tool %q.", name)
		if s := r.suggest(name); len(s) > 0 {
			msg += " Did you mean: " + strings.Join(s, ", ") + "? Use ionos_search_tools to find the right tool."
		} else {
			msg += " Use ionos_search_tools to find the right tool."
		}
		return errorResult(msg), nil, nil
	}

	// Defense-in-depth: enforce the scope gate here too. Registration already
	// keeps classes the scope disallows out of the catalog, so this is a second
	// gate — a mutating tool is only dispatched when IONOS_MCP_TOOL_SCOPE opts its
	// class in, and it fails closed otherwise.
	if !r.scope.Allows(entry.Class) {
		return errorResult(fmt.Sprintf("tool %q requires IONOS_MCP_TOOL_SCOPE to include the %q capability; current scope is %q", name, entry.Class.String(), r.scope.String())), nil, nil
	}

	// Forward to the catalog server. Input validation, schema enforcement and
	// IONOS error enrichment all happen inside that handler; we relay its
	// result verbatim, preserving IsError.
	res, err := r.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: in.Arguments,
	})
	if err != nil {
		return errorResult(fmt.Sprintf("calling %q failed: %v", name, err)), nil, nil
	}
	return res, nil, nil
}

// errorResult builds a tool result flagged as an error (IsError) carrying msg.
func errorResult(msg string) *mcp.CallToolResult {
	res := tools.TextResult(msg)
	res.IsError = true
	return res
}
