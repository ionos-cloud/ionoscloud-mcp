package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// buildCatalog registers every product onto a private in-memory catalog server,
// self-connects to it, and snapshots each tool's metadata. Group attribution
// works by registering products one at a time and treating every tool that
// appears after a product's Register call (and was not present before) as
// belonging to that product — the product tools carry no name prefix, so
// registration order is the only signal.
func buildCatalog(ctx context.Context, products []Product) (*router, error) {
	catalog := mcp.NewServer(&mcp.Implementation{
		Name:    "ionos-cloud-mcp-catalog",
		Version: "internal",
	}, nil)

	clientT, serverT := mcp.NewInMemoryTransports()
	if _, err := catalog.Connect(ctx, serverT, nil); err != nil {
		return nil, fmt.Errorf("dynamic: connecting catalog server: %w", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "ionos-cloud-mcp-catalog-reader", Version: "internal"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		return nil, fmt.Errorf("dynamic: connecting catalog reader: %w", err)
	}

	r := &router{
		byName:  make(map[string]catalogEntry),
		session: session,
	}

	seen := make(map[string]bool)
	for _, p := range products {
		p.Register(catalog)
		for tool, err := range session.Tools(ctx, nil) {
			if err != nil {
				return nil, fmt.Errorf("dynamic: listing catalog tools after %q: %w", p.Name, err)
			}
			if seen[tool.Name] {
				continue
			}
			seen[tool.Name] = true
			schema, mErr := json.Marshal(tool.InputSchema)
			if mErr != nil {
				return nil, fmt.Errorf("dynamic: marshaling input schema for %q: %w", tool.Name, mErr)
			}
			e := catalogEntry{
				Name:        tool.Name,
				Group:       p.Name,
				Description: tool.Description,
				InputSchema: schema,
			}
			r.entries = append(r.entries, e)
			r.byName[tool.Name] = e
		}
	}

	if len(r.entries) == 0 {
		return nil, fmt.Errorf("dynamic: catalog is empty (no products registered any tools)")
	}

	sort.Slice(r.entries, func(i, j int) bool { return r.entries[i].Name < r.entries[j].Name })
	return r, nil
}

// catalogSummary renders the product listing embedded in the ionos_search_tools
// description: one line per group with its tool count and summary, so the model
// knows what exists without an extra round-trip.
func catalogSummary(products []Product, r *router) string {
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
