package loader

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DomainName identifies an IONOS Cloud tool domain.
type DomainName string

const (
	DomainCompute       DomainName = "compute"
	DomainDNS           DomainName = "dns"
	DomainBilling       DomainName = "billing"
	DomainCert          DomainName = "cert"
	DomainObjectStorage DomainName = "objectstorage"
)

// AllDomains is the complete ordered list of supported domains.
var AllDomains = []DomainName{
	DomainCompute, DomainDNS, DomainBilling, DomainCert, DomainObjectStorage,
}

// EagerDomains are loaded at startup by default (small domains, ≤14 tools each).
var EagerDomains = []DomainName{DomainDNS, DomainBilling, DomainCert}

// DomainInfo holds static metadata for a domain.
type DomainInfo struct {
	Description string
	ToolCount   int
	LoadTool    string // MCP tool name to load this domain; empty if always loaded eagerly
}

var domainMeta = map[DomainName]DomainInfo{
	DomainCompute: {
		Description: "Virtual infrastructure: servers, volumes, networks, load balancers, NAT, security groups",
		ToolCount:   50,
		LoadTool:    "ionos_load_compute_tools",
	},
	DomainDNS: {
		Description: "DNS zones, records, reverse records, secondary zones, DNSSEC, quota",
		ToolCount:   14,
	},
	DomainBilling: {
		Description: "Invoices, usage, traffic, utilization, EVN, products, billing profile",
		ToolCount:   14,
	},
	DomainCert: {
		Description: "TLS certificates, auto-certificates, ACME providers",
		ToolCount:   6,
	},
	DomainObjectStorage: {
		Description: "S3-compatible buckets, objects, access keys, regions, bucket configuration",
		ToolCount:   23,
		LoadTool:    "ionos_load_objectstorage_tools",
	},
}

// Meta returns the static metadata for a domain.
func Meta(d DomainName) DomainInfo {
	return domainMeta[d]
}

// DomainClients bundles all IONOS Cloud SDK clients needed across domains.
type DomainClients struct {
	Compute *computeSDK.APIClient
	DNS     *dnsSDK.APIClient
	Billing *billSDK.APIClient
	Cert    *certSDK.APIClient
	ObjSt   *objstSDK.APIClient
	ObjMgmt *objmgmtSDK.APIClient
}

// DomainLoader manages lazy, thread-safe registration of IONOS Cloud tool domains.
type DomainLoader struct {
	server  *mcp.Server
	clients DomainClients
	mu      sync.Mutex
	loaded  map[DomainName]bool
}

// NewDomainLoader creates a loader. No domains are registered until Load is called.
func NewDomainLoader(server *mcp.Server, clients DomainClients) *DomainLoader {
	return &DomainLoader{
		server:  server,
		clients: clients,
		loaded:  make(map[DomainName]bool),
	}
}

// Load registers all tools for the given domain if not already loaded.
// Returns true if tools were newly registered, false if already loaded.
// Safe for concurrent use; idempotent.
func (dl *DomainLoader) Load(domain DomainName) (bool, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	if dl.loaded[domain] {
		return false, nil
	}
	switch domain {
	case DomainCompute:
		compute.RegisterAll(dl.server, dl.clients.Compute)
	case DomainDNS:
		dns.RegisterAll(dl.server, dl.clients.DNS)
	case DomainBilling:
		billing.RegisterAll(dl.server, dl.clients.Billing)
	case DomainCert:
		cert.RegisterAll(dl.server, dl.clients.Cert)
	case DomainObjectStorage:
		objectstorage.RegisterAll(dl.server, dl.clients.ObjSt, dl.clients.ObjMgmt)
	default:
		return false, fmt.Errorf("unknown domain %q", domain)
	}
	dl.loaded[domain] = true
	return true, nil
}

// IsLoaded reports whether a domain's tools have been registered.
func (dl *DomainLoader) IsLoaded(domain DomainName) bool {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.loaded[domain]
}

// LoadedDomains returns a snapshot of the currently registered domain names.
func (dl *DomainLoader) LoadedDomains() []DomainName {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	var out []DomainName
	for _, d := range AllDomains {
		if dl.loaded[d] {
			out = append(out, d)
		}
	}
	return out
}

// UnloadedDomains returns the names of domains not yet registered.
func (dl *DomainLoader) UnloadedDomains() []DomainName {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	var out []DomainName
	for _, d := range AllDomains {
		if !dl.loaded[d] {
			out = append(out, d)
		}
	}
	return out
}

// ParseDomains parses a comma-separated domain list from envVal.
// Falls back to defaults when envVal is empty.
// The special value "all" expands to AllDomains.
// Unknown names are silently skipped.
func ParseDomains(envVal string, defaults []DomainName) []DomainName {
	v := strings.TrimSpace(envVal)
	if v == "" {
		return defaults
	}
	if v == "all" {
		return AllDomains
	}
	known := make(map[DomainName]bool, len(AllDomains))
	for _, d := range AllDomains {
		known[d] = true
	}
	seen := make(map[DomainName]bool)
	var result []DomainName
	for _, part := range strings.Split(v, ",") {
		name := DomainName(strings.TrimSpace(part))
		if known[name] && !seen[name] {
			result = append(result, name)
			seen[name] = true
		}
	}
	return result
}
