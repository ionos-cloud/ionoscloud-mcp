// Package main provides the entry point for the IONOS Cloud MCP server.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/config"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/mcp"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"

	// Import all toolsets for init() registration
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/compute"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/dns"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/iam"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/kubernetes"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/loadbalancing"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/networking"
)

var (
	// Version is set by ldflags at build time
	Version = "0.2.0"
	// BuildDate is set by ldflags at build time
	BuildDate = "unknown"
)

var (
	enabledToolsets []string
	readOnly        bool
	enableLogging   bool
	enableMetrics   bool
)

var rootCmd = &cobra.Command{
	Use:   "ionoscloud-mcp",
	Short: "IONOS Cloud MCP Server",
	Long: `IONOS Cloud MCP Server provides Model Context Protocol (MCP) access
to IONOS Cloud infrastructure resources.

The server communicates via JSON-RPC over stdio and requires
authentication via environment variables:
  - IONOS_USERNAME + IONOS_PASSWORD for username/password auth
  - IONOS_TOKEN for token-based auth`,
	Version: Version,
	RunE:    runServer,
}

var listToolsetsCmd = &cobra.Command{
	Use:   "list-toolsets",
	Short: "List available toolsets",
	Run: func(cmd *cobra.Command, args []string) {
		for _, ts := range toolsets.Toolsets() {
			fmt.Printf("%-15s %s (%d tools)\n", ts.GetName(), ts.GetDescription(), len(ts.GetTools()))
		}
	},
}

var listToolsCmd = &cobra.Command{
	Use:   "list-tools",
	Short: "List all available tools",
	Run: func(cmd *cobra.Command, args []string) {
		for _, ts := range toolsets.Toolsets() {
			fmt.Printf("\n[%s] %s\n", ts.GetName(), ts.GetDescription())
			for _, tool := range ts.GetTools() {
				annotation := ""
				if tool.Tool.Annotations != nil {
					if tool.Tool.Annotations.ReadOnlyHint != nil && *tool.Tool.Annotations.ReadOnlyHint {
						annotation = " (read-only)"
					} else if tool.Tool.Annotations.DestructiveHint != nil && *tool.Tool.Annotations.DestructiveHint {
						annotation = " (destructive)"
					}
				}
				fmt.Printf("  - %s%s\n", tool.Tool.Name, annotation)
			}
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&enabledToolsets, "toolsets", nil,
		"Enabled toolsets (comma-separated, e.g., 'compute,dns'). Default: all")
	rootCmd.PersistentFlags().BoolVar(&readOnly, "read-only", false,
		"Only expose read-only tools (no create, update, delete)")
	rootCmd.PersistentFlags().BoolVar(&enableLogging, "logging", false,
		"Enable request/response logging to stderr")
	rootCmd.PersistentFlags().BoolVar(&enableMetrics, "metrics", false,
		"Enable basic timing metrics")

	rootCmd.AddCommand(listToolsetsCmd)
	rootCmd.AddCommand(listToolsCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load base config from environment
	cfg := config.LoadFromEnv()

	// Override with command-line flags
	if cmd.Flags().Changed("toolsets") {
		cfg.EnabledToolsets = enabledToolsets
	}
	if cmd.Flags().Changed("read-only") {
		cfg.ReadOnly = readOnly
	}
	if cmd.Flags().Changed("logging") {
		cfg.EnableLogging = enableLogging
	}
	if cmd.Flags().Changed("metrics") {
		cfg.EnableMetrics = enableMetrics
	}

	// Create IONOS client
	client, err := ionos.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create IONOS client: %w", err)
	}

	// Create and configure MCP server
	mcpCfg := &mcp.Config{
		EnableLogging:   cfg.EnableLogging,
		EnableMetrics:   cfg.EnableMetrics,
		ReadOnly:        cfg.ReadOnly,
		EnabledToolsets: cfg.EnabledToolsets,
	}

	server := mcp.NewServer(client, mcpCfg)

	// Run the server
	return server.Run()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
