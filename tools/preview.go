package tools

import (
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Preview rendering for the two-phase confirmation flow; ConfirmationStore
// (confirm.go) holds its state. Product-agnostic, so every product's write tools
// produce previews in one shape.

// KV is one "key: value" line in a preview block.
type KV struct {
	K string
	V string
}

// Fields builds a preview field list from alternating key, value arguments:
//
//	tools.Fields("name", name, "cores", tools.OptInt32(in.Cores))
//
// Empty values are dropped at render time, so callers can list every optional
// field unconditionally. Panics on an odd number of arguments.
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

// BlastRadius accumulates the resources an operation touches. Callers Add each
// collection they fetched; zero counts are dropped.
//
// Whether entries are destroyed or merely affected is a flag rather than free text,
// because labelling a survivor "will be destroyed" is a false claim in the one place
// a caller looks before authorizing a change.
type BlastRadius struct {
	// Destroys marks entries that cease to exist with their parent. The zero value is
	// the safe one: an unset radius reads as merely affected.
	Destroys bool
	Counts   []LabeledCount
	Total    int
}

// DestroyedRadius is for entries that cease to exist with their parent.
func DestroyedRadius() *BlastRadius { return &BlastRadius{Destroys: true} }

// AffectedRadius is for entries that survive. Say what each one loses in its label.
func AffectedRadius() *BlastRadius { return &BlastRadius{} }

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
	// Fields echoes what the operation will act on.
	Fields []KV
	// Radius is what the operation touches. Nil for creates and leaf resources.
	Radius *BlastRadius
	// EmptyNote replaces the count list when Radius is present but empty.
	EmptyNote string
	// Tool is the tool to call again; Replay the arguments it must repeat.
	Tool   string
	Replay []KV
	// TokenNote says what the token authorizes. Keep it narrow — it is the only
	// signal that a token is not blanket permission.
	TokenNote string
}

// Render produces the preview text with the token footer. Returned as a NON-error
// result: the first call of a two-phase flow succeeded, it just changed nothing.
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
			if p.Radius.Destroys {
				b.WriteString("\nContained resources that will be destroyed:\n")
			} else {
				b.WriteString("\nNot deleted, but affected:\n")
			}
			for _, c := range p.Radius.Counts {
				fmt.Fprintf(&b, "  - %d %s\n", c.Count, c.Label)
			}
			// A total means something for a destroy; summing unlike survivors does not.
			if p.Radius.Destroys {
				fmt.Fprintf(&b, "Total resources that will be destroyed: %d\n", p.Radius.Total)
			}
		}
	}

	fmt.Fprintf(&b, "\nTo proceed, call %s again with:\n", p.Tool)
	// Rendered inside the replay list so it aligns with the other arguments.
	writeFields(&b, append(append([]KV{}, p.Replay...), KV{K: "confirmation_token", V: token}))
	fmt.Fprintf(&b, "%s and expires in %s.", p.TokenNote, ConfirmationTTL)
	return b.String()
}

// writeFields renders aligned "key: value" lines, skipping empty values.
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

// ConfirmErrorText turns a ConfirmationStore failure into remediation, naming the
// arguments to re-send so a caller does not retry a dead token.
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

// CallerIDMetaKey carries the originating client's session id when a call is
// forwarded through a shared session (see tools/dynamic).
const CallerIDMetaKey = "ionos/callerID"

// CallerID identifies the client session a call arrived on, so a confirmation
// token cannot cross from one client to another. Empty on stdio, where the
// process serves exactly one client and there is nothing to isolate.
func CallerID(req *mcp.CallToolRequest) string {
	if req == nil {
		return ""
	}
	// The session id is authoritative because it comes from the transport, so a
	// client cannot forge it. Order matters: _meta travels in the request and is
	// therefore client-controlled, so trusting it first would let one client
	// claim another's id and spend its tokens.
	if req.Session != nil {
		if id := req.Session.ID(); id != "" {
			return id
		}
	}
	// No session id of its own: stdio, or the in-memory catalog session that
	// dynamic mode forwards through. Only there is the dispatcher-supplied id used.
	if req.Params != nil {
		if id, ok := req.Params.Meta[CallerIDMetaKey].(string); ok {
			return id
		}
	}
	return ""
}

// Target joins the parts of a confirmation target. Include every identifier that
// makes the operation unique, so a token cannot be replayed elsewhere. The
// calling session is part of the target, so one client cannot spend a token
// another client minted.
func Target(req *mcp.CallToolRequest, parts ...string) string {
	return strings.Join(append([]string{CallerID(req)}, parts...), "|")
}

// HasToken reports whether a two-phase input carried a non-empty token, i.e.
// whether this is the execute call rather than the preview call.
func HasToken(token *string) bool {
	return token != nil && strings.TrimSpace(*token) != ""
}

// DeletedAsync is the success message for a DELETE, which returns no body.
func DeletedAsync(what, id string) string {
	return fmt.Sprintf("Deleted %s %s. Deletion is asynchronous; the API has accepted the request.", what, id)
}

// The Opt* helpers render an optional value for a preview field, returning "" when
// unset so writeFields drops the line.

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

// Redacted acknowledges a secret without echoing it — clients log previews.
func Redacted(v *string) string {
	if v == nil || *v == "" {
		return ""
	}
	return "(set, not shown)"
}

// IncompleteRadiusNote warns that a blast radius could not be fully determined. An
// unreadable collection must never render as an empty one.
func IncompleteRadiusNote(what ...string) string {
	var named []string
	for _, w := range what {
		if w != "" {
			named = append(named, w)
		}
	}
	if len(named) == 0 {
		return ""
	}
	return fmt.Sprintf("\nWARNING: could not read this resource's %s, so the list below is INCOMPLETE — this may destroy more than it shows.", strings.Join(named, " or "))
}

// CappedCountNote warns that a preview count is a floor rather than an exact
// figure, for list endpoints that cap a page and report no total.
func CappedCountNote(capped bool, label string, limit int) string {
	if !capped {
		return ""
	}
	return fmt.Sprintf("\nNOTE: the %s count below is a floor — the API returns at most %d per page and reports no total.", label, limit)
}

// ErrLabel names a collection when reading it failed, and "" when it succeeded, so
// callers can build one warning covering however many lookups went wrong.
func ErrLabel(err error, label string) string {
	if err == nil {
		return ""
	}
	return label
}
