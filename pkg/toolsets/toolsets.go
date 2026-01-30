// Package toolsets provides the toolset registry for IONOS Cloud MCP tools.
package toolsets

import (
	"sync"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

var (
	mu        sync.RWMutex
	toolsets  []api.Toolset
	toolIndex map[string]api.ServerTool
)

func init() {
	toolIndex = make(map[string]api.ServerTool)
}

// Register adds a toolset to the global registry.
// This is typically called from toolset init() functions.
func Register(ts api.Toolset) {
	mu.Lock()
	defer mu.Unlock()

	toolsets = append(toolsets, ts)

	// Index all tools by name for fast lookup
	for _, tool := range ts.GetTools() {
		toolIndex[tool.Tool.Name] = tool
	}
}

// Toolsets returns all registered toolsets.
func Toolsets() []api.Toolset {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]api.Toolset, len(toolsets))
	copy(result, toolsets)
	return result
}

// AllTools returns all tools from all registered toolsets.
func AllTools() []api.ServerTool {
	mu.RLock()
	defer mu.RUnlock()

	var result []api.ServerTool
	for _, ts := range toolsets {
		result = append(result, ts.GetTools()...)
	}
	return result
}

// GetTool returns a specific tool by name.
func GetTool(name string) (api.ServerTool, bool) {
	mu.RLock()
	defer mu.RUnlock()

	tool, ok := toolIndex[name]
	return tool, ok
}

// ToolCount returns the total number of registered tools.
func ToolCount() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(toolIndex)
}

// ToolNames returns the names of all registered tools.
func ToolNames() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(toolIndex))
	for name := range toolIndex {
		names = append(names, name)
	}
	return names
}

// Reset clears all registered toolsets (for testing).
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	toolsets = nil
	toolIndex = make(map[string]api.ServerTool)
}
