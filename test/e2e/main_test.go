//go:build e2e

// Package e2e drives the actual shipped binary over real stdio JSON-RPC framing,
// with a local HTTP mock standing in for the IONOS API (injected via
// IONOS_API_URL). It is gated behind the `e2e` build tag so the default
// `go test ./...` stays fast and hermetic; run it with `make test-e2e`.
package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverVersion = "e2e-test"

var (
	binPath string

	mu          sync.Mutex
	lastUA      string
	statusByPfx = map[string]int{} // path-prefix → status override
)

// TestMain builds the server binary once and starts the shared API mock.
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "ionos-mcp-e2e")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	binPath = tmp + "/ionoscloud-mcp"
	build := exec.Command("go", "build",
		"-ldflags", "-X main.serverVersion="+serverVersion,
		"-o", binPath, ".")
	build.Dir = "../.."
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("build failed: " + err.Error())
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()
	apiURL = ts.URL

	os.Exit(m.Run())
}

var apiURL string

func mockHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	lastUA = r.Header.Get("User-Agent")
	status := http.StatusOK
	for pfx, code := range statusByPfx {
		if strings.HasPrefix(r.URL.Path, pfx) {
			status = code
			break
		}
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte("{}"))
}

func setStatus(pfx string, code int) {
	mu.Lock()
	statusByPfx[pfx] = code
	mu.Unlock()
}

func clearStatus() {
	mu.Lock()
	statusByPfx = map[string]int{}
	mu.Unlock()
}

func userAgent() string {
	mu.Lock()
	defer mu.Unlock()
	return lastUA
}

// syncBuffer is a goroutine-safe sink for subprocess stderr: os/exec writes to
// it from a copier goroutine while the test reads it, so both ends lock.
type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// waitStderrContains polls the buffer until it contains sub or the deadline
// passes — the subprocess copier goroutine writes asynchronously, so a bare
// read right after the handshake can miss a startup log line.
func waitStderrContains(buf *syncBuffer, sub string, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for {
		if strings.Contains(buf.String(), sub) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// spawn starts the binary with the given extra env and connects an MCP client
// over stdio. stderrBuf (if non-nil) captures the subprocess stderr. The
// returned channel receives a value on each notifications/tools/list_changed.
func spawn(t *testing.T, extraEnv map[string]string, stderrBuf *syncBuffer, args ...string) (*mcp.ClientSession, <-chan struct{}) {
	t.Helper()

	cmd := exec.Command(binPath, args...)

	// Strip keys that the test controls from the ambient environment. On Linux
	// getenv() returns the first match, so ambient values would otherwise win
	// over anything appended later. IONOS_MCP_LOAD_MODE is always stripped:
	// tests that need a specific mode set it via extraEnv; tests that omit it
	// get the binary's compiled-in default (eager).
	overrideKeys := map[string]bool{
		"IONOS_TOKEN":         true,
		"IONOS_API_URL":       true,
		"IONOS_S3_ACCESS_KEY": true,
		"IONOS_S3_SECRET_KEY": true,
		"IONOS_MCP_LOAD_MODE": true,
	}
	for k := range extraEnv {
		overrideKeys[k] = true
	}
	base := os.Environ()
	filtered := make([]string, 0, len(base))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if !overrideKeys[key] {
			filtered = append(filtered, entry)
		}
	}
	env := append(filtered,
		"IONOS_TOKEN=test-token",
		"IONOS_API_URL="+apiURL,
		"IONOS_S3_ACCESS_KEY=ak",
		"IONOS_S3_SECRET_KEY=sk",
	)
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	cmd.Env = env
	if stderrBuf != nil {
		cmd.Stderr = stderrBuf
	}

	notify := make(chan struct{}, 16)
	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-client", Version: "1.0.0"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(_ context.Context, _ *mcp.ToolListChangedRequest) {
			select {
			case notify <- struct{}{}:
			default:
			}
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect to binary: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session, notify
}

func toolNameSet(t *testing.T, s *mcp.ClientSession) map[string]bool {
	t.Helper()
	res, err := s.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	return names
}

func textOf(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}
