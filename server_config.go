package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	serverName    = "ionoscloud-mcp"
	serverVersion = "1.0.0"
)

func buildUserAgent() string {
	bundleVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/ionos-cloud/sdk-go-bundle/shared" {
				bundleVersion = strings.TrimPrefix(dep.Version, "v")
				break
			}
		}
	}

	return fmt.Sprintf("ionoscloud-mcp/%s_ionos-cloud-sdk-go-bundle/%s_os/%s_arch/%s",
		serverVersion, bundleVersion, runtime.GOOS, runtime.GOARCH)
}
