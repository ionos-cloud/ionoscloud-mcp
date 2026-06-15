package dynamic

import (
	"reflect"
	"testing"
)

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
	// Exact whole-name match is the strongest signal.
	if exact := score(getOne, "get_datacenter", tokenize("get_datacenter")); exact < 100 {
		t.Errorf("exact name match score = %d, want >= 100", exact)
	}
}

func TestRouterSuggestAndSearch(t *testing.T) {
	entries := []catalogEntry{
		{Name: "list_datacenters", Group: "compute", Description: "List data centers."},
		{Name: "get_datacenter", Group: "compute", Description: "Get a data center."},
		{Name: "list_dns_zones", Group: "dns", Description: "List DNS zones."},
	}
	r := &router{entries: entries, byName: map[string]catalogEntry{}}
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
