package tools

import (
	"errors"
	"strings"
	"testing"
)

func TestFieldsBuildsPairs(t *testing.T) {
	got := Fields("name", "web-1", "cores", "4")
	want := []KV{{K: "name", V: "web-1"}, {K: "cores", V: "4"}}
	if len(got) != len(want) {
		t.Fatalf("Fields returned %d entries, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Fields()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFieldsPanicsOnOddArgs(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Fields with an odd argument count should panic")
		}
	}()
	// The argument count is deliberately hidden behind a Split so it is not
	// statically known: staticcheck's SA5012 check recognises the even-pairs
	// contract and rejects an odd call it can see through, which is a feature for
	// real callers but means this test has to reach the runtime panic another way.
	args := strings.Split("name,web-1,orphan", ",")
	Fields(args...)
}

func TestFieldsEmpty(t *testing.T) {
	if got := Fields(); len(got) != 0 {
		t.Errorf("Fields() = %+v, want empty", got)
	}
}

// TestPreviewSkipsEmptyValues is what lets a handler pass every optional field
// unconditionally: nil options render as "" and must not produce a blank line.
func TestPreviewSkipsEmptyValues(t *testing.T) {
	out := Preview{
		Headline:  "About to CREATE one thing:",
		Fields:    Fields("name", "web-1", "description", "", "cores", "4"),
		Tool:      "create_thing",
		Replay:    Fields("name", "web-1"),
		TokenNote: "The token authorizes creating only this thing",
	}.Render("tok123")

	if strings.Contains(out, "description") {
		t.Errorf("empty field should be omitted entirely:\n%s", out)
	}
	for _, want := range []string{"web-1", "cores", "4", "create_thing", "tok123", "5m0s"} {
		if !strings.Contains(out, want) {
			t.Errorf("preview missing %q:\n%s", want, out)
		}
	}
}

// TestPreviewAlignsFields pins the alignment, since a ragged preview is what the
// shared formatter exists to prevent.
func TestPreviewAlignsFields(t *testing.T) {
	out := Preview{
		Headline:  "About to CREATE one thing:",
		Fields:    Fields("a", "1", "longer_key", "2"),
		Tool:      "create_thing",
		TokenNote: "note",
	}.Render("tok")
	if !strings.Contains(out, "  a:          1\n") {
		t.Errorf("short key not padded to the longest key width:\n%s", out)
	}
	if !strings.Contains(out, "  longer_key: 2\n") {
		t.Errorf("longest key should have a single trailing space:\n%s", out)
	}
}

// TestPreviewTokenIsAlignedWithReplay checks the token line participates in the
// replay block's alignment, so the caller sees one coherent argument list.
func TestPreviewTokenIsAlignedWithReplay(t *testing.T) {
	out := Preview{
		Headline:  "About to DELETE a thing.",
		Tool:      "delete_thing",
		Replay:    Fields("datacenter_id", "dc-1"),
		TokenNote: "note",
	}.Render("tok")
	if !strings.Contains(out, "  datacenter_id:      dc-1\n") {
		t.Errorf("replay field not aligned against confirmation_token:\n%s", out)
	}
	if !strings.Contains(out, "  confirmation_token: tok\n") {
		t.Errorf("token line missing or misaligned:\n%s", out)
	}
}

func TestPreviewBlastRadius(t *testing.T) {
	r := &BlastRadius{}
	r.Add("servers", 2)
	r.Add("volumes", 0) // zero counts are dropped
	r.Add("LANs", 1)
	if r.Total != 3 {
		t.Errorf("Total = %d, want 3", r.Total)
	}
	if len(r.Counts) != 2 {
		t.Errorf("Counts = %+v, want only the non-zero entries", r.Counts)
	}

	out := Preview{
		Headline:  "About to DELETE a thing.",
		Radius:    r,
		EmptyNote: "nothing inside",
		Tool:      "delete_thing",
		TokenNote: "note",
	}.Render("tok")
	for _, want := range []string{"2 servers", "1 LANs", "Total resources that will be destroyed: 3"} {
		if !strings.Contains(out, want) {
			t.Errorf("preview missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "volumes") {
		t.Errorf("zero-count collection should not appear:\n%s", out)
	}
	if strings.Contains(out, "nothing inside") {
		t.Errorf("EmptyNote must not show when the radius is non-empty:\n%s", out)
	}
}

func TestPreviewEmptyBlastRadiusUsesEmptyNote(t *testing.T) {
	out := Preview{
		Headline:  "About to DELETE a thing.",
		Radius:    &BlastRadius{},
		EmptyNote: "This thing is empty.",
		Tool:      "delete_thing",
		TokenNote: "note",
	}.Render("tok")
	if !strings.Contains(out, "This thing is empty.") {
		t.Errorf("EmptyNote should show for an empty radius:\n%s", out)
	}
	if strings.Contains(out, "will be destroyed") {
		t.Errorf("empty radius must not render a count list:\n%s", out)
	}
}

// TestPreviewNilRadiusOmitsSection covers create previews, which have no radius.
func TestPreviewNilRadiusOmitsSection(t *testing.T) {
	out := Preview{
		Headline:  "About to CREATE one thing:",
		Fields:    Fields("name", "x"),
		EmptyNote: "should never appear",
		Tool:      "create_thing",
		TokenNote: "note",
	}.Render("tok")
	if strings.Contains(out, "should never appear") || strings.Contains(out, "will be destroyed") {
		t.Errorf("a nil radius must render no blast-radius section at all:\n%s", out)
	}
}

func TestConfirmErrorText(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"mismatch", ErrTokenMismatch, "different target"},
		{"expired", ErrTokenExpired, "expired"},
		{"unknown", ErrTokenUnknown, "not recognized"},
		{"wrapped mismatch", errors.New("x: " + ErrTokenMismatch.Error()), "not recognized"}, // not wrapped via %w -> default
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConfirmErrorText("delete_thing", "thing_id", tt.err)
			if !strings.Contains(got, tt.want) {
				t.Errorf("ConfirmErrorText(%v) = %q, want it to mention %q", tt.err, got, tt.want)
			}
			// Every branch must name the tool and the arguments to re-send;
			// that is what stops a model retrying a dead token.
			if !strings.Contains(got, "delete_thing") || !strings.Contains(got, "thing_id") {
				t.Errorf("ConfirmErrorText should name the tool and replay args: %q", got)
			}
		})
	}
}

func TestTargetBindsParentChain(t *testing.T) {
	if got := Target("dc-1", "srv-2", "nic-3"); got != "dc-1|srv-2|nic-3" {
		t.Errorf("Target = %q, want the full chain joined by |", got)
	}
	// Different parents with the same leaf name must not collide.
	if Target("dc-1", "web") == Target("dc-2", "web") {
		t.Error("targets under different parents must differ")
	}
}

func TestHasToken(t *testing.T) {
	empty, blank, real := "", "   ", "tok"
	tests := map[string]struct {
		in   *string
		want bool
	}{
		"nil":        {nil, false},
		"empty":      {&empty, false},
		"whitespace": {&blank, false},
		"real":       {&real, true},
	}
	for name, tt := range tests {
		if got := HasToken(tt.in); got != tt.want {
			t.Errorf("HasToken(%s) = %v, want %v", name, got, tt.want)
		}
	}
}

func TestOptFormatters(t *testing.T) {
	i := int32(4)
	f := float32(50)
	fFrac := float32(10.5)
	b := true
	s := "x"
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"nil int32", OptInt32(nil), ""},
		{"int32", OptInt32(&i), "4"},
		{"nil float32", OptFloat32(nil), ""},
		{"whole float32 drops decimals", OptFloat32(&f), "50"},
		{"fractional float32", OptFloat32(&fFrac), "10.5"},
		{"nil bool", OptBool(nil), ""},
		{"bool", OptBool(&b), "true"},
		{"nil string", OptStr(nil), ""},
		{"string", OptStr(&s), "x"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestDeletedAsync(t *testing.T) {
	got := DeletedAsync("server", "srv-1")
	for _, want := range []string{"server", "srv-1", "asynchronous"} {
		if !strings.Contains(got, want) {
			t.Errorf("DeletedAsync missing %q: %q", want, got)
		}
	}
}
