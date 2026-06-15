package dynamic

import (
	"sort"
	"strings"
)

// search ranks catalog entries against a free-text query, optionally restricted
// to a single product group, returning at most limit results. An empty query
// (typically combined with a group) browses: it returns entries unscored, in
// the catalog's stable name order. A non-empty query drops zero-score entries.
func (r *router) search(query, group string, limit int) []catalogEntry {
	tokens := tokenize(query)

	type scored struct {
		entry catalogEntry
		score int
	}
	var hits []scored
	for _, e := range r.entries {
		if group != "" && e.Group != group {
			continue
		}
		if len(tokens) == 0 {
			hits = append(hits, scored{entry: e})
			continue
		}
		if s := score(e, query, tokens); s > 0 {
			hits = append(hits, scored{entry: e, score: s})
		}
	}

	// Sort by descending score; break ties by shorter name (a query covers more
	// of a short, specific name like list_datacenters than a long incidental one
	// like get_billing_usage_by_datacenter), then alphabetically for stability.
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		if len(hits[i].entry.Name) != len(hits[j].entry.Name) {
			return len(hits[i].entry.Name) < len(hits[j].entry.Name)
		}
		return hits[i].entry.Name < hits[j].entry.Name
	})

	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]catalogEntry, len(hits))
	for i, h := range hits {
		out[i] = h.entry
	}
	return out
}

// score weights a query against one entry. Name matches dominate description
// matches so that, e.g., "datacenter" surfaces list_datacenters above tools
// that merely mention datacenters in their description.
func score(e catalogEntry, query string, tokens []string) int {
	name := strings.ToLower(e.Name)
	desc := strings.ToLower(e.Description)
	q := strings.ToLower(strings.TrimSpace(query))

	s := 0
	switch {
	case name == q:
		s += 100
	case strings.Contains(name, q):
		// Coverage-weighted: the larger a fraction of the name the query spans,
		// the more relevant. Keeps the query's primary noun (e.g. "datacenter"
		// → list_datacenters) above tools that merely contain it in a longer name.
		s += 10 + 40*len(q)/len(name)
	}
	for _, t := range tokens {
		if strings.Contains(name, t) {
			s += 10
		}
		if strings.Contains(desc, t) {
			s += 2
		}
	}
	return s
}

// suggest returns up to three tool names closest to name, for "did you mean"
// hints on unknown-tool errors. It reuses the search scorer with name as the
// query, falling back to substring matches.
func (r *router) suggest(name string) []string {
	matches := r.search(name, "", 3)
	out := make([]string, 0, len(matches))
	for _, e := range matches {
		out = append(out, e.Name)
	}
	return out
}

// tokenize lowercases query and splits it into alphanumeric tokens, dropping
// punctuation and separators (so "list_datacenters" and "list datacenters"
// tokenize the same way).
func tokenize(query string) []string {
	fields := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	return fields
}
