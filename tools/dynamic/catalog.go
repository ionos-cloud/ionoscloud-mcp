package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// buildCatalog registers every product onto a private in-memory "catalog"
// server used to dispatch ionos_call_tool, and snapshots each tool's metadata.
//
// Group attribution and duplicate detection are done with a separate, exact
// pass: each product is registered on its OWN throwaway server and that
// server's tool list is read back, so every tool is attributed to exactly the
// product that registered it. This also catches a name collision across
// products — which the combined catalog could not, because mcp.AddTool silently
// replaces a tool of the same name (leaving describe/search showing one
// product's metadata while call_tool would invoke the other's handler).
func buildCatalog(ctx context.Context, products []Product, scope tools.Scope) (d *dispatcher, err error) {
	catalog := mcp.NewServer(&mcp.Implementation{
		Name:    "ionos-cloud-mcp-catalog",
		Version: "internal",
	}, nil)

	clientT, serverT := mcp.NewInMemoryTransports()
	srvSession, err := catalog.Connect(ctx, serverT, nil)
	if err != nil {
		return nil, fmt.Errorf("dynamic: connecting catalog server: %w", err)
	}
	// From here on srvSession is live; any error return must tear it down (and
	// the reader session, once established) or it leaks the server goroutines.
	defer func() {
		if err != nil {
			_ = srvSession.Close()
		}
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "ionos-cloud-mcp-catalog-reader", Version: "internal"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		return nil, fmt.Errorf("dynamic: connecting catalog reader: %w", err)
	}
	defer func() {
		if err != nil {
			_ = session.Close()
		}
	}()

	d = &dispatcher{
		byName:     make(map[string]catalogEntry),
		scope:      scope,
		session:    session,
		srvSession: srvSession,
	}

	for _, p := range products {
		// Dispatch surface: register on the shared catalog.
		p.Register(catalog)

		// Attribution + dedup: register on a throwaway server and read it back.
		prodTools, err := productTools(ctx, p)
		if err != nil {
			return nil, err
		}
		for _, tool := range prodTools {
			if prev, dup := d.byName[tool.Name]; dup {
				return nil, fmt.Errorf("dynamic: duplicate tool name %q registered by products %q and %q", tool.Name, prev.Group, p.Name)
			}
			schema, mErr := json.Marshal(tool.InputSchema)
			if mErr != nil {
				return nil, fmt.Errorf("dynamic: marshaling input schema for %q: %w", tool.Name, mErr)
			}
			e := catalogEntry{
				Name:        tool.Name,
				Group:       p.Name,
				Description: tool.Description,
				InputSchema: schema,
				Class:       tools.ClassFromName(tool.Name),
			}
			d.entries = append(d.entries, e)
			d.byName[tool.Name] = e
		}
	}

	if len(d.entries) == 0 {
		return nil, fmt.Errorf("dynamic: catalog is empty (no products registered any tools)")
	}

	sort.Slice(d.entries, func(i, j int) bool { return d.entries[i].Name < d.entries[j].Name })
	return d, nil
}

// productTools registers a single product on a throwaway in-memory server and
// returns exactly the tools it registered, so group attribution is exact rather
// than inferred from registration order.
func productTools(ctx context.Context, p Product) ([]*mcp.Tool, error) {
	srv := mcp.NewServer(&mcp.Implementation{Name: "ionos-cloud-mcp-catalog-" + p.Name, Version: "internal"}, nil)
	p.Register(srv)

	clientT, serverT := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverT, nil)
	if err != nil {
		return nil, fmt.Errorf("dynamic: connecting catalog probe for %q: %w", p.Name, err)
	}
	defer ss.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "probe", Version: "internal"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		return nil, fmt.Errorf("dynamic: connecting catalog probe reader for %q: %w", p.Name, err)
	}
	defer cs.Close()

	var out []*mcp.Tool
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			return nil, fmt.Errorf("dynamic: listing tools for %q: %w", p.Name, err)
		}
		out = append(out, tool)
	}
	return out, nil
}

// catalogSummary renders the product listing embedded in the ionos_search_tools
// description: one line per group with its tool count and summary, so the model
// knows what exists without an extra round-trip.
func catalogSummary(products []Product, r *dispatcher) string {
	counts := make(map[string]int, len(products))
	for _, e := range r.entries {
		counts[e.Group]++
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Catalog: %d tools across %d product groups.\n", len(r.entries), len(products))
	for _, p := range products {
		fmt.Fprintf(&b, "- %s (%d tools): %s\n", p.Name, counts[p.Name], p.Summary)
	}
	return b.String()
}
