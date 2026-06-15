package dynamic

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// callText runs callHandler and returns (isError, text) for assertions.
func callText(t *testing.T, d *dispatcher, in tools.CallToolInput) (bool, string) {
	t.Helper()
	res, _, err := d.callHandler(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("callHandler returned error: %v", err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return res.IsError, b.String()
}

func TestCallHandlerReadOnlyGuard(t *testing.T) {
	// A (hypothetical) mutating tool in the catalog must be refused before any
	// dispatch, even though it's a known name. session is nil to prove the guard
	// returns before forwarding.
	d := &dispatcher{byName: map[string]catalogEntry{
		"create_server": {Name: "create_server", Group: "compute", ReadOnly: false},
	}}
	isErr, text := callText(t, d, tools.CallToolInput{Name: "create_server"})
	if !isErr {
		t.Fatal("calling a non-read-only tool should be an error")
	}
	if !strings.Contains(text, "not read-only") {
		t.Errorf("guard message = %q, want it to mention 'not read-only'", text)
	}
}

func TestCallHandlerUnknownNoSuggestion(t *testing.T) {
	// Unknown name with no close matches exercises the no-suggestion branch.
	d := &dispatcher{entries: nil, byName: map[string]catalogEntry{}}
	isErr, text := callText(t, d, tools.CallToolInput{Name: "zzz_nonexistent"})
	if !isErr {
		t.Fatal("unknown tool should be an error")
	}
	if strings.Contains(text, "Did you mean") {
		t.Errorf("empty catalog should yield no suggestions; got %q", text)
	}
	if !strings.Contains(text, "ionos_search_tools") {
		t.Errorf("error should point to ionos_search_tools; got %q", text)
	}
}

func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "Get the billing profile.", []string{"Get the billing profile."}},
		{"two", "Do a thing. Then another.", []string{"Do a thing.", "Then another."}},
		{
			"version not split",
			"Read resource ionos://billing/focus-v1.3. Done.",
			[]string{"Read resource ionos://billing/focus-v1.3.", "Done."},
		},
		{
			"date in parens not split",
			"Get utilization for a date (YYYY-MM-DD). Use daily.",
			[]string{"Get utilization for a date (YYYY-MM-DD).", "Use daily."},
		},
		{"no trailing period", "Just a fragment", []string{"Just a fragment"}},
		{"question and bang", "Ready? Go!", []string{"Ready?", "Go!"}},
		{"collapses blank", "A.   B.", []string{"A.", "B."}},
		{"empty", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitSentences(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitSentences(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSnippet(t *testing.T) {
	desc := "Get the billing profile for your IONOS account. Call this first — the contract number is required by other tools."
	tests := []struct {
		name   string
		desc   string
		tokens []string
		want   string
	}{
		{
			"no query returns first sentence",
			desc, nil,
			"Get the billing profile for your IONOS account.",
		},
		{
			"match in first sentence returns only first",
			desc, []string{"billing"},
			"Get the billing profile for your IONOS account.",
		},
		{
			"match only in later sentence appends it",
			desc, []string{"contract"},
			"Get the billing profile for your IONOS account. Call this first — the contract number is required by other tools.",
		},
		{
			"no match returns first sentence",
			desc, []string{"kubernetes"},
			"Get the billing profile for your IONOS account.",
		},
		{
			"single-sentence desc",
			"List all DNS zones.", []string{"zones"},
			"List all DNS zones.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := snippet(tt.desc, tt.tokens); got != tt.want {
				t.Errorf("snippet() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"list_datacenters", []string{"list", "datacenters"}},
		{"DNS zone", []string{"dns", "zone"}},
		{"top_n=10", []string{"top", "n", "10"}},
		{"  ", nil},
		{"", nil},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := tokenize(tt.in)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenize(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestScoreRanking(t *testing.T) {
	list := catalogEntry{Name: "list_datacenters", Description: "List all data centers in your account."}
	getOne := catalogEntry{Name: "get_datacenter", Description: "Get a single data center by ID."}
	billing := catalogEntry{Name: "get_billing_usage_by_datacenter", Description: "Usage grouped by datacenter."}
	descOnly := catalogEntry{Name: "list_servers", Description: "Servers live inside a datacenter."}

	q := "datacenters"
	tokens := tokenize(q)
	sList := score(list, q, tokens)
	sBilling := score(billing, q, tokens)
	sDescOnly := score(descOnly, q, tokens)

	// Coverage weighting: the short, on-point name outranks the long incidental one.
	if sList <= sBilling {
		t.Errorf("list_datacenters (%d) should outrank get_billing_usage_by_datacenter (%d) for %q", sList, sBilling, q)
	}
	// A name match outranks a description-only match.
	if sList <= sDescOnly {
		t.Errorf("name match (%d) should outrank description-only match (%d)", sList, sDescOnly)
	}
	// Exact whole-name match is the strongest signal. Pin the value so a
	// regression in the token bonus is caught, not masked by a loose >=100:
	// 100 (exact name) + 10+10 (name tokens get, datacenter) + 2 (desc token
	// "get" in "Get a single data center by ID.") = 122.
	if exact := score(getOne, "get_datacenter", tokenize("get_datacenter")); exact != 122 {
		t.Errorf("exact name match score = %d, want 122", exact)
	}
}

func TestSearchPunctuationOnlyQueryMatchesNothing(t *testing.T) {
	entries := []catalogEntry{
		{Name: "list_datacenters", Group: "compute", Description: "List data centers."},
		{Name: "list_dns_zones", Group: "dns", Description: "List DNS zones."},
	}
	d := &dispatcher{entries: entries, byName: map[string]catalogEntry{}}

	// A query that tokenizes to nothing must NOT browse-all.
	if got := d.search("!!!", "", 10); len(got) != 0 {
		t.Errorf("search(%q) returned %d results, want 0 (no browse-all)", "!!!", len(got))
	}
	// An empty query DOES browse.
	if got := d.search("", "", 10); len(got) != 2 {
		t.Errorf("empty-query browse returned %d, want 2", len(got))
	}
	// Group filter is case-insensitive.
	if got := d.search("", "DNS", 10); len(got) != 1 || got[0].Name != "list_dns_zones" {
		t.Errorf("case-insensitive group filter = %v, want [list_dns_zones]", names(got))
	}
}

func TestRouterSuggestAndSearch(t *testing.T) {
	entries := []catalogEntry{
		{Name: "list_datacenters", Group: "compute", Description: "List data centers."},
		{Name: "get_datacenter", Group: "compute", Description: "Get a data center."},
		{Name: "list_dns_zones", Group: "dns", Description: "List DNS zones."},
	}
	r := &dispatcher{entries: entries, byName: map[string]catalogEntry{}}
	for _, e := range entries {
		r.byName[e.Name] = e
	}

	// suggest returns close names for a typo, capped at 3.
	got := r.suggest("list_datacenter")
	if len(got) == 0 || got[0] != "list_datacenters" {
		t.Errorf("suggest(\"list_datacenter\") = %v, want list_datacenters first", got)
	}

	// group filter restricts results.
	dns := r.search("", "dns", 10)
	if len(dns) != 1 || dns[0].Name != "list_dns_zones" {
		t.Errorf("search(group=dns) = %v, want [list_dns_zones]", names(dns))
	}

	// limit caps results.
	if all := r.search("", "", 2); len(all) != 2 {
		t.Errorf("search(limit=2) returned %d, want 2", len(all))
	}
}

func names(es []catalogEntry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Name
	}
	return out
}
