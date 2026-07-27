package tools

import (
	"errors"
	"fmt"
	"strings"
)

// This file is the presentation half of the two-phase confirmation flow whose
// state half is ConfirmationStore (confirm.go): it renders the preview a write
// tool returns on its first call, and the remediation text it returns when a
// token is rejected. It is deliberately product-agnostic — nothing here imports
// a product SDK — so every product's write tools produce previews in the same
// shape and a model only has to learn one format.

// KV is one "key: value" line in a preview block. A slice of these keeps the
// order stable, which a map would not. Build them with Fields rather than by
// literal — see that function for why.
type KV struct {
	K string
	V string
}

// Fields builds a preview field list from alternating key, value arguments:
//
//	tools.Fields("name", name, "cores", tools.OptInt32(in.Cores))
//
// It exists for two reasons. Unkeyed cross-package struct literals are rejected
// by vet's composites check, and the keyed alternative
// (tools.KV{K: "name", V: name}) repeated per field would bury the field names
// that are the point of the list. Values that are empty are dropped at render
// time, so callers can list every optional field unconditionally.
//
// Panics on an odd number of arguments: that is a coding error, caught the first
// time the tool is registered or exercised, in the same spirit as RegisterTool's
// name/method assertion. In practice it rarely gets that far — staticcheck's
// SA5012 check recognises the even-pairs contract and flags a bad literal call
// at lint time.
func Fields(pairs ...string) []KV {
	if len(pairs)%2 != 0 {
		panic(fmt.Sprintf("tools.Fields: got %d arguments, want an even number of key, value pairs", len(pairs)))
	}
	out := make([]KV, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, KV{K: pairs[i], V: pairs[i+1]})
	}
	return out
}

// LabeledCount is one line of a blast-radius preview ("3 volumes").
type LabeledCount struct {
	Label string
	Count int
}

// BlastRadius accumulates the child resources a delete would destroy. Callers
// Add each collection they fetched; zero counts are dropped so a preview lists
// only what actually exists.
type BlastRadius struct {
	Counts []LabeledCount
	Total  int
}

// Add records a non-zero count under label. Zero counts are ignored.
func (b *BlastRadius) Add(label string, n int) {
	if n > 0 {
		b.Counts = append(b.Counts, LabeledCount{Label: label, Count: n})
		b.Total += n
	}
}

// Preview describes a two-phase confirmation preview: what is about to happen,
// the inputs it will act on, what it would destroy, and how to proceed.
type Preview struct {
	// Headline is the first line, e.g. "About to CREATE one server:".
	Headline string
	// Fields echoes the identity and properties the operation will use, so the
	// caller can see exactly what it asked for before committing.
	Fields []KV
	// Radius is the set of resources a delete would destroy. Leave nil for
	// create and for leaf resources that contain nothing.
	Radius *BlastRadius
	// EmptyNote is shown instead of a count list when Radius is present but
	// empty, e.g. "This server has no attached volumes or NICs."
	EmptyNote string
	// Tool is the tool name to call again, and Replay the arguments that call
	// must repeat so the token's target still matches.
	Tool   string
	Replay []KV
	// TokenNote explains precisely what the token authorizes. Keep it narrow —
	// it is the model's only signal that a token is not blanket permission.
	TokenNote string
}

// Render produces the preview text, ending with the token footer. The result is
// returned as a NON-error result: the first call of a two-phase flow succeeded,
// it just did not mutate anything yet.
func (p Preview) Render(token string) string {
	var b strings.Builder
	b.WriteString(p.Headline)
	if !strings.HasSuffix(p.Headline, "\n") {
		b.WriteString("\n")
	}
	writeFields(&b, p.Fields)

	if p.Radius != nil {
		if p.Radius.Total == 0 {
			if p.EmptyNote != "" {
				fmt.Fprintf(&b, "\n%s\n", p.EmptyNote)
			}
		} else {
			b.WriteString("\nContained resources that will be destroyed:\n")
			for _, c := range p.Radius.Counts {
				fmt.Fprintf(&b, "  - %d %s\n", c.Count, c.Label)
			}
			fmt.Fprintf(&b, "Total resources that will be destroyed: %d\n", p.Radius.Total)
		}
	}

	fmt.Fprintf(&b, "\nTo proceed, call %s again with:\n", p.Tool)
	// The token is rendered as part of the replay list so it aligns with the
	// other arguments the caller has to repeat.
	writeFields(&b, append(append([]KV{}, p.Replay...), KV{K: "confirmation_token", V: token}))
	fmt.Fprintf(&b, "%s and expires in %s.", p.TokenNote, ConfirmationTTL)
	return b.String()
}

// writeFields renders "  key: value" lines, aligned on the longest key so a
// multi-field preview stays readable. Fields with an empty value are skipped,
// which is what lets callers pass every optional field unconditionally.
func writeFields(b *strings.Builder, fields []KV) {
	width := 0
	for _, f := range fields {
		if f.V != "" && len(f.K) > width {
			width = len(f.K)
		}
	}
	for _, f := range fields {
		if f.V == "" {
			continue
		}
		fmt.Fprintf(b, "  %-*s %s\n", width+1, f.K+":", f.V)
	}
}

// ConfirmErrorText maps a ConfirmationStore failure to remediation a model can
// act on. replayArgs names the arguments to re-send for a fresh preview —
// naming them explicitly is what stops a model retrying the same dead token.
func ConfirmErrorText(tool, replayArgs string, err error) string {
	switch {
	case errors.Is(err, ErrTokenMismatch):
		return fmt.Sprintf("confirmation_token was issued for a different target; re-run %s with only %s to preview THIS target and get a fresh token", tool, replayArgs)
	case errors.Is(err, ErrTokenExpired):
		return fmt.Sprintf("confirmation_token expired; re-run %s with only %s for a fresh preview and token", tool, replayArgs)
	default: // ErrTokenUnknown
		return fmt.Sprintf("confirmation_token not recognized (already used or never issued); re-run %s with only %s for a preview and token", tool, replayArgs)
	}
}

// Target joins the parts of a confirmation target. Every part of the parent
// chain must be included so a token minted under one parent can never be
// replayed against a same-named resource under a different parent.
func Target(parts ...string) string {
	return strings.Join(parts, "|")
}

// HasToken reports whether a two-phase input carried a non-empty token, i.e.
// whether this is the execute call rather than the preview call.
func HasToken(token *string) bool {
	return token != nil && strings.TrimSpace(*token) != ""
}

// DeletedAsync is the standard success message for a DELETE, which returns no
// body. Provisioning is asynchronous, so the handler reports that the request
// was accepted rather than that the resource is gone.
func DeletedAsync(what, id string) string {
	return fmt.Sprintf("Deleted %s %s. Deletion is asynchronous; the API has accepted the request.", what, id)
}

// The Opt* helpers render an optional value as a preview field, returning ""
// when the pointer is nil so writeFields skips the line. They exist so a
// handler can list every optional field unconditionally instead of guarding
// each one.

// OptStr dereferences an optional string.
func OptStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// OptInt32 renders an optional int32.
func OptInt32(v *int32) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

// OptFloat32 renders an optional float32 without trailing zeros, so 50 prints
// as "50" rather than "50.00".
func OptFloat32(v *float32) string {
	if v == nil {
		return ""
	}
	s := strings.TrimRight(fmt.Sprintf("%.2f", *v), "0")
	return strings.TrimSuffix(s, ".")
}

// OptBool renders an optional bool.
func OptBool(v *bool) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%t", *v)
}
