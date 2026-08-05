package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dynamic"
)

// Confirmation tokens must not cross between clients. Only the Streamable HTTP
// transport issues session ids (stdio and in-memory both return ""), so these
// tests run over a real HTTP handler — an in-memory harness cannot tell two
// sessions apart and would pass no matter what the code did.

var tokenRE = regexp.MustCompile(`confirmation_token: (\w+)`)

// httpMCPServer serves the compute tools over Streamable HTTP against a mocked
// IONOS API, and returns the base URL.
func httpMCPServer(t *testing.T, mode string) string {
	t.Helper()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"dc-new","properties":{"name":"iso-vdc","location":"de/txl"}}`))
	}))
	t.Cleanup(api.Close)

	cfg := &shared.Configuration{
		Token:              "test-token",
		DefaultHeader:      map[string]string{},
		DefaultQueryParams: make(map[string][]string),
		Servers:            shared.ServerConfigurations{{URL: api.URL}},
		OperationServers:   map[string]shared.ServerConfigurations{},
	}
	client := ionos.NewAPIClient(cfg)
	scope := tools.Scope{Write: true, Destructive: true}
	store := tools.NewConfirmationStore()

	srv := mcp.NewServer(&mcp.Implementation{Name: "iso-test", Version: "test"}, nil)
	if mode == "dynamic" {
		products := []dynamic.Product{
			{Name: "compute", Summary: "Compute Engine.", Register: func(s *mcp.Server) { compute.RegisterAll(s, client, scope, store) }},
		}
		closer, err := dynamic.Register(context.Background(), srv, products, scope)
		if err != nil {
			t.Fatalf("dynamic.Register failed: %v", err)
		}
		t.Cleanup(func() { closer.Close() })
	} else {
		compute.RegisterAll(srv, client, scope, store)
	}

	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
	mcpSrv := httptest.NewServer(h)
	t.Cleanup(mcpSrv.Close)
	return mcpSrv.URL
}

// session is one client with its own Mcp-Session-Id.
type session struct {
	t    *testing.T
	cs   *mcp.ClientSession
	mode string
}

func connect(t *testing.T, url, mode string) *session {
	t.Helper()
	c := mcp.NewClient(&mcp.Implementation{Name: "iso-client", Version: "test"}, nil)
	cs, err := c.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: url}, nil)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return &session{t: t, cs: cs, mode: mode}
}

// createDatacenter calls create_datacenter, optionally with a forged caller id
// in _meta, and returns the tool's text output.
func (s *session) createDatacenter(args map[string]any, forgedCallerID string) string {
	s.t.Helper()
	params := &mcp.CallToolParams{Name: "create_datacenter", Arguments: args}
	if s.mode == "dynamic" {
		params = &mcp.CallToolParams{
			Name:      "ionos_call_tool",
			Arguments: map[string]any{"name": "create_datacenter", "arguments": args},
		}
	}
	if forgedCallerID != "" {
		params.Meta = mcp.Meta{tools.CallerIDMetaKey: forgedCallerID}
	}
	res, err := s.cs.CallTool(context.Background(), params)
	if err != nil {
		s.t.Fatalf("CallTool failed: %v", err)
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}

func (s *session) mintToken(args map[string]any) string {
	s.t.Helper()
	out := s.createDatacenter(args, "")
	m := tokenRE.FindStringSubmatch(out)
	if m == nil {
		s.t.Fatalf("no token in preview:\n%s", out)
	}
	return m[1]
}

func withToken(args map[string]any, tok string) map[string]any {
	out := map[string]any{}
	for k, v := range args {
		out[k] = v
	}
	out["confirmation_token"] = tok
	return out
}

func refused(out string) bool {
	return strings.Contains(out, "different target") || strings.Contains(out, "not recognized")
}

// TestConfirmationTokensAreSessionScoped is the regression guard: before tokens
// were bound to the caller, session B could execute a delete that session A had
// previewed and approved.
func TestConfirmationTokensAreSessionScoped(t *testing.T) {
	for _, mode := range []string{"eager", "dynamic"} {
		t.Run(mode, func(t *testing.T) {
			url := httpMCPServer(t, mode)
			args := map[string]any{"name": "iso-vdc", "location": "de/txl"}

			a := connect(t, url, mode)
			b := connect(t, url, mode)

			// B must not be able to spend a token A minted.
			tok := a.mintToken(args)
			if out := b.createDatacenter(withToken(args, tok), ""); !refused(out) {
				t.Errorf("session B spent session A's token:\n%s", out)
			}

			// A must still be able to spend its own, or the two-phase flow is broken.
			tok = a.mintToken(args)
			if out := a.createDatacenter(withToken(args, tok), ""); refused(out) {
				t.Errorf("session A could not spend its own token:\n%s", out)
			}

			// _meta is client-controlled, so forging a caller id must not work.
			// The transport-supplied session id has to win.
			tok = a.mintToken(args)
			if out := b.createDatacenter(withToken(args, tok), callerIDOf(t, a)); !refused(out) {
				t.Errorf("session B spoofed A's caller id via _meta:\n%s", out)
			}
		})
	}
}

// callerIDOf returns a session's id as the server sees it. A real attacker would
// have to guess it; the test hands it over to prove that even knowing it is not
// enough.
func callerIDOf(t *testing.T, s *session) string {
	t.Helper()
	id := s.cs.ID()
	if id == "" {
		t.Fatal("expected a non-empty session id over HTTP; the isolation test would be vacuous")
	}
	return id
}
