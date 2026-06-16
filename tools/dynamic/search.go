package dynamic

import (
	"sort"
	"strings"
)

// snippet returns a short description for search results. It always includes
// the first sentence (the tool's summary). When the query matched on a LATER
// sentence (e.g. "idle" → "...set include_zero=true to find idle resources"),
// it appends that matching sentence so the shown text explains the match —
// scoring (in score) runs over the full description, so the tool is found
// regardless; this only controls what's displayed. With no query (group
// browse), only the first sentence is shown.
func snippet(desc string, tokens []string) string {
	parts := splitSentences(desc)
	if len(parts) == 0 {
		return strings.TrimSpace(desc)
	}
	out := parts[0]
	if len(tokens) == 0 || containsAny(strings.ToLower(parts[0]), tokens) {
		return out
	}
	for _, s := range parts[1:] {
		if containsAny(strings.ToLower(s), tokens) {
			return out + " " + s
		}
	}
	return out
}

// containsAny reports whether s contains any of the (already lowercased) tokens.
func containsAny(s string, tokens []string) bool {
	for _, t := range tokens {
		if strings.Contains(s, t) {
			return true
		}
	}
	return false
}

// splitSentences splits desc into trimmed, non-empty sentences, breaking after
// sentence-ending punctuation (. ! ?) that is followed by whitespace. A period
// with no trailing space does not split, so "v1.3." and "(YYYY-MM-DD)." stay
// intact. (Go's RE2 has no lookbehind, so this is a manual scan.)
func splitSentences(desc string) []string {
	desc = strings.TrimSpace(desc)
	var out []string
	start := 0
	runes := []rune(desc)
	for i := range len(runes) {
		c := runes[i]
		if (c == '.' || c == '!' || c == '?') && i+1 < len(runes) && isSpace(runes[i+1]) {
			if s := strings.TrimSpace(string(runes[start : i+1])); s != "" {
				out = append(out, s)
			}
			start = i + 1
		}
	}
	if s := strings.TrimSpace(string(runes[start:])); s != "" {
		out = append(out, s)
	}
	return out
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// search ranks catalog entries against a free-text query, optionally restricted
// to a single product group, returning at most limit results. An empty query
// (typically combined with a group) browses: it returns entries unscored, in
// the catalog's stable name order. A non-empty query drops zero-score entries.
func (r *dispatcher) search(query, group string, limit int) []catalogEntry {
	tokens := tokenize(query)
	browse := strings.TrimSpace(query) == ""

	type scored struct {
		entry catalogEntry
		score int
	}
	var hits []scored
	for _, e := range r.entries {
		if group != "" && !strings.EqualFold(e.Group, group) {
			continue
		}
		if browse {
			// No query (typically a group browse): list entries unscored.
			hits = append(hits, scored{entry: e})
			continue
		}
		if len(tokens) == 0 {
			// Query present but only punctuation/separators (e.g. "!!!"): it
			// matched nothing, so return no results rather than browsing all.
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
func (r *dispatcher) suggest(name string) []string {
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
