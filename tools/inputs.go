package tools

// Input types for tool parameters.
// Each struct maps to the JSON schema that the MCP SDK auto-generates
// for a tool's input. Non-pointer fields are treated as required.

// Compute input types

// ListDatacentersInput is the input for list_datacenters (no required ID fields).
type ListDatacentersInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1); depth 1 includes names and basic properties"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property name to value pairs (case-sensitive contains match); e.g. {\"name\":\"prod\",\"location\":\"de/fra\"}. Filterable properties include: name, description, location, version"`
}

type DatacenterIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type ServerIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type VolumeIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	VolumeID     string `json:"volume_id" jsonschema:"the ID of the volume"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type SnapshotIDInput struct {
	SnapshotID string `json:"snapshot_id" jsonschema:"the ID of the snapshot"`
	Depth      *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type NicIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	NicID        string `json:"nic_id" jsonschema:"the ID of the network interface"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type LanIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LanID        string `json:"lan_id" jsonschema:"the ID of the LAN"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type FirewallRuleIDInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID       string `json:"server_id" jsonschema:"the ID of the server"`
	NicID          string `json:"nic_id" jsonschema:"the ID of the network interface"`
	FirewallRuleID string `json:"firewallrule_id" jsonschema:"the ID of the firewall rule"`
	Depth          *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type IpBlockIDInput struct {
	IpBlockID string `json:"ipblock_id" jsonschema:"the ID of the IP block"`
	Depth     *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type LoadBalancerIDInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID string `json:"loadbalancer_id" jsonschema:"the ID of the load balancer"`
	Depth          *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type NetworkLoadBalancerIDInput struct {
	DatacenterID          string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NetworkLoadBalancerID string `json:"network_loadbalancer_id" jsonschema:"the ID of the network load balancer"`
	Depth                 *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type ApplicationLoadBalancerIDInput struct {
	DatacenterID              string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ApplicationLoadBalancerID string `json:"application_loadbalancer_id" jsonschema:"the ID of the application load balancer"`
	Depth                     *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type TargetGroupIDInput struct {
	TargetGroupID string `json:"target_group_id" jsonschema:"the ID of the target group"`
	Depth         *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type NatGatewayIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NatGatewayID string `json:"nat_gateway_id" jsonschema:"the ID of the NAT gateway"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type PccIDInput struct {
	PccID string `json:"pcc_id" jsonschema:"the ID of the private cross-connect"`
	Depth *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type GpuIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	GpuID        string `json:"gpu_id" jsonschema:"the ID of the GPU"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type SecurityGroupIDInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group"`
	Depth           *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type SecurityGroupRuleIDInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group"`
	RuleID          string `json:"rule_id" jsonschema:"the ID of the security group rule"`
	Depth           *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type RequestIDInput struct {
	RequestID string `json:"request_id" jsonschema:"the ID of the request"`
	Depth     *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type TemplateIDInput struct {
	TemplateID string `json:"template_id" jsonschema:"the ID of the template"`
	Depth      *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// GetContractInput is the input for get_contract (no required ID fields).
type GetContractInput struct {
	Depth *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// List*Input types for list_ tools that have no required ID parameters (previously struct{}).

type ListIPBlocksInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"location\":\"de/fra\"}"`
}

type ListTargetGroupsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListPrivateCrossConnectsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListTemplatesInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"cpuFamily\":\"INTEL_SKYLAKE\"}"`
}

type ListImagesInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"ubuntu\",\"imageType\":\"HDD\",\"licenceType\":\"LINUX\",\"location\":\"de/fra\"}"`
}

type ListLocationsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"Frankfurt\"}"`
}

type ListSnapshotsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"location\":\"de/fra\"}"`
}

type ListRequestsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"method\":\"POST\",\"requestStatus\":\"DONE\",\"createdBy\":\"user@example.com\"}"`
}

type ListK8sClustersInput struct {
	Depth *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1)"`
}

type ListInDatacenterInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListInServerInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string            `json:"server_id"     jsonschema:"the ID of the server"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListNatGatewayRulesInput struct {
	DatacenterID string            `json:"datacenter_id"  jsonschema:"the ID of the data center"`
	NatGatewayID string            `json:"nat_gateway_id" jsonschema:"the ID of the NAT gateway"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"type\":\"SNAT\"}"`
}

type ListLanNicsInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LanID        string            `json:"lan_id"        jsonschema:"the ID of the LAN"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListAlbForwardingRulesInput struct {
	DatacenterID              string            `json:"datacenter_id"               jsonschema:"the ID of the data center"`
	ApplicationLoadBalancerID string            `json:"application_loadbalancer_id" jsonschema:"the ID of the application load balancer"`
	Depth                     *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters                   map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListNlbForwardingRulesInput struct {
	DatacenterID          string            `json:"datacenter_id"           jsonschema:"the ID of the data center"`
	NetworkLoadBalancerID string            `json:"network_loadbalancer_id" jsonschema:"the ID of the network load balancer"`
	Depth                 *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters               map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListLoadBalancerNicsInput struct {
	DatacenterID   string            `json:"datacenter_id"   jsonschema:"the ID of the data center"`
	LoadBalancerID string            `json:"loadbalancer_id" jsonschema:"the ID of the load balancer"`
	Depth          *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters        map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"}"`
}

type ListSecurityGroupRulesInput struct {
	DatacenterID    string            `json:"datacenter_id"     jsonschema:"the ID of the data center"`
	SecurityGroupID string            `json:"security_group_id" jsonschema:"the ID of the security group"`
	Depth           *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters         map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"type\":\"INGRESS\",\"direction\":\"INBOUND\"}"`
}

type ListFirewallRulesInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string            `json:"server_id"     jsonschema:"the ID of the server"`
	NicID        string            `json:"nic_id"        jsonschema:"the ID of the network interface"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"type\":\"INGRESS\",\"direction\":\"INBOUND\"}"`
}

// DNS input types

type ZoneIDInput struct {
	ZoneID string `json:"zone_id" jsonschema:"the ID of the DNS zone"`
}

type RecordIDInput struct {
	ZoneID   string `json:"zone_id" jsonschema:"the ID of the DNS zone"`
	RecordID string `json:"record_id" jsonschema:"the ID of the DNS record"`
}

type ReverseRecordIDInput struct {
	ReverseRecordID string `json:"reverse_record_id" jsonschema:"the ID of the reverse DNS record"`
}

type SecondaryZoneIDInput struct {
	SecondaryZoneID string `json:"secondary_zone_id" jsonschema:"the ID of the secondary DNS zone"`
}

// Billing input types
// Most billing tools require a contract number — call get_billing_profile first to get it.
// Exception: list_billing_invoices_by_period is contract-agnostic (the underlying API endpoint does not accept a contract parameter).

type BillingContractInput struct {
	Contract int32 `json:"contract" jsonschema:"contract number from get_billing_profile"`
}

type BillingContractPeriodInput struct {
	Contract int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Period   string `json:"period" jsonschema:"billing period in YYYY-MM format (e.g. 2026-04). Maximum one month per request — for wider ranges call once per month"`
}

type BillingInvoiceIDInput struct {
	Contract  int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	InvoiceID string `json:"invoice_id" jsonschema:"the invoice ID (e.g. GY00111111)"`
}

type BillingDatacenterInput struct {
	Contract     int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	DatacenterID string `json:"datacenter_id" jsonschema:"the VDC UUID"`
}

// BillingUtilizationInput is the input for list_billing_utilization.
// Compaction flags reduce response size — zero-quantity meters are dropped by default.
type BillingUtilizationInput struct {
	Contract     int32    `json:"contract" jsonschema:"contract number from get_billing_profile"`
	IncludeZero  *bool    `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false); set true to find existing resources that didn't consume in the window"`
	GroupBy      *string  `json:"group_by,omitempty" jsonschema:"aggregation level: omitted or '' = per-resource (default), 'meter' = sum per SKU per datacenter, 'datacenter' = sum per type per datacenter — coarser groupings shrink output but lose detail"`
	DatacenterID *string  `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
	MeterTypes   []string `json:"meter_types,omitempty" jsonschema:"filter to these meter type categories only (client-side); e.g. ['DBAAS','DNS','SERVER']"`
	Regions      []string `json:"regions,omitempty" jsonschema:"filter to these regions only (client-side); e.g. ['de/fra','es/vit']"`
	TopN         *int32   `json:"top_n,omitempty" jsonschema:"return only the N largest meters globally, sorted by quantity desc — flat list with dc_id/dc_name on each row, datacenters[] omitted; ideal for cost audits on contracts with many datacenters. When combined with group_by='datacenter', top_meters[] rows have no meter_id (type+unit aggregates); with group_by='meter', meter_id is the SKU"`
}

type BillingUtilizationPeriodInput struct {
	Contract     int32    `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Period       string   `json:"period" jsonschema:"billing period in YYYY-MM format (e.g. 2026-04). Maximum one month per request — for wider ranges call once per month"`
	IncludeZero  *bool    `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false)"`
	GroupBy      *string  `json:"group_by,omitempty" jsonschema:"aggregation level: omitted or '' = per-resource (default), 'meter' = sum per SKU per datacenter, 'datacenter' = sum per type per datacenter"`
	DatacenterID *string  `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
	MeterTypes   []string `json:"meter_types,omitempty" jsonschema:"filter to these meter type categories only (client-side); e.g. ['DBAAS','DNS','SERVER']"`
	Regions      []string `json:"regions,omitempty" jsonschema:"filter to these regions only (client-side); e.g. ['de/fra']"`
	TopN         *int32   `json:"top_n,omitempty" jsonschema:"return only the N largest meters globally, sorted by quantity desc (flat list, datacenters[] omitted). When combined with group_by='datacenter', top_meters[] rows have no meter_id"`
}

type BillingUtilizationDateInput struct {
	Contract     int32    `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Date         string   `json:"date" jsonschema:"date in YYYY-MM-DD format (e.g. 2026-04-15)"`
	IncludeZero  *bool    `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false)"`
	GroupBy      *string  `json:"group_by,omitempty" jsonschema:"aggregation level: omitted or '' = per-resource (default), 'meter' = sum per SKU per datacenter, 'datacenter' = sum per type per datacenter"`
	DatacenterID *string  `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
	MeterTypes   []string `json:"meter_types,omitempty" jsonschema:"filter to these meter type categories only (client-side); e.g. ['DBAAS']"`
	Regions      []string `json:"regions,omitempty" jsonschema:"filter to these regions only (client-side); e.g. ['de/fra']"`
	TopN         *int32   `json:"top_n,omitempty" jsonschema:"return only the N largest meters globally, sorted by quantity desc (flat list, datacenters[] omitted). When combined with group_by='datacenter', top_meters[] rows have no meter_id"`
}

// BillingUsageInput is the input for list_billing_usage.
// UsageMeter has fewer fields than UtilizationMeter — type/region/resource filters do not apply.
type BillingUsageInput struct {
	Contract     int32   `json:"contract" jsonschema:"contract number from get_billing_profile"`
	IncludeZero  *bool   `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false); set true to find datacenters with metered SKUs that didn't consume"`
	DatacenterID *string `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
}

type BillingUsageDatacenterInput struct {
	Contract     int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	DatacenterID string `json:"datacenter_id" jsonschema:"the VDC UUID"`
	IncludeZero  *bool  `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false)"`
}

type BillingDateInput struct {
	Contract int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Date     string `json:"date" jsonschema:"date in YYYY-MM-DD format (e.g. 2026-04-15)"`
}

type BillingPeriodOnlyInput struct {
	Period string `json:"period" jsonschema:"billing period in YYYY-MM format (e.g. 2026-04)"`
}

type BillingProductsInput struct {
	Contract int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Filter   string `json:"filter" jsonschema:"keyword to filter products by description (e.g. 'RAM', 'Kubernetes', 'Postgres', 'storage'). Use broad terms to find relevant pricing"`
}

// Object Storage input types

type ObjectStorageBucketInput struct {
	Bucket string `json:"bucket" jsonschema:"the name of the object storage bucket"`
}

type ObjectStorageObjectInput struct {
	Bucket string `json:"bucket" jsonschema:"the name of the object storage bucket"`
	Key    string `json:"key" jsonschema:"the object key (path within the bucket)"`
}

type ObjectStorageListObjectsInput struct {
	Bucket            string  `json:"bucket" jsonschema:"the name of the object storage bucket"`
	Prefix            *string `json:"prefix,omitempty" jsonschema:"optional key prefix to filter results (e.g. 'images/' to list only objects under that path)"`
	ContinuationToken *string `json:"continuation_token,omitempty" jsonschema:"optional pagination token returned by a previous list operation to continue listing objects"`
	MaxKeys           *int32  `json:"max_keys,omitempty" jsonschema:"optional maximum number of objects to return in a single page"`
}

type AccessKeyIDInput struct {
	AccessKeyID string `json:"access_key_id" jsonschema:"the ID of the object storage access key"`
}

type ObjectStorageRegionInput struct {
	Region string `json:"region" jsonschema:"the region name (e.g. eu-central-3)"`
}

type ObjectStorageListObjectVersionsInput struct {
	Bucket string  `json:"bucket" jsonschema:"the name of the object storage bucket"`
	Prefix *string `json:"prefix,omitempty" jsonschema:"optional key prefix to filter versions"`
}

// Activity Log input types

type ActivityLogQueryInput struct {
	Contract             int32    `json:"contract" jsonschema:"the contract number whose activity log to query; reseller/partner users get IDs from list_activitylog_contracts, single-contract users read it from their JWT"`
	DateStart            *string  `json:"date_start,omitempty" jsonschema:"inclusive start date YYYY-MM-DD; defaults to 7 days ago when omitted"`
	DateEnd              *string  `json:"date_end,omitempty" jsonschema:"inclusive end date YYYY-MM-DD; defaults to today when omitted; maximum range is 90 days"`
	Offset               *int32   `json:"offset,omitempty" jsonschema:"0-based pagination offset"`
	Limit                *int32   `json:"limit,omitempty" jsonschema:"max events to return; defaults to 25; increase only when the user explicitly asks for bulk data"`
	User                 *string  `json:"user,omitempty" jsonschema:"filter by username (client-side); e.g. 'ionosctl-v6@cloud.ionos.com' — drastically reduces output when investigating a specific user"`
	EventTypes           []string `json:"event_types,omitempty" jsonschema:"filter to these event types only (client-side); e.g. ['Error','RequestAccepted'] — omit Provision and RequestStatusUpdate to cut ~65% of typical log volume"`
	IncludeStatusUpdates *bool    `json:"include_status_updates,omitempty" jsonschema:"include RequestStatusUpdate events (default false); these are async provisioning echoes that account for ~55% of log volume and are rarely useful"`
}

// Certificate Manager input types

type CertificateIDInput struct {
	CertificateID string `json:"certificate_id" jsonschema:"the ID of the certificate"`
}

type AutoCertificateIDInput struct {
	AutoCertificateID string `json:"auto_certificate_id" jsonschema:"the ID of the auto-certificate"`
}

type ProviderIDInput struct {
	ProviderID string `json:"provider_id" jsonschema:"the ID of the certificate provider"`
}

// Kubernetes input types

type K8sClusterIDInput struct {
	K8sClusterID string `json:"k8s_cluster_id" jsonschema:"the ID of the Kubernetes cluster"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type K8sNodepoolIDInput struct {
	K8sClusterID string `json:"k8s_cluster_id" jsonschema:"the ID of the Kubernetes cluster"`
	NodepoolID   string `json:"nodepool_id" jsonschema:"the ID of the Kubernetes node pool"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type K8sNodeIDInput struct {
	K8sClusterID string `json:"k8s_cluster_id" jsonschema:"the ID of the Kubernetes cluster"`
	NodepoolID   string `json:"nodepool_id" jsonschema:"the ID of the Kubernetes node pool"`
	NodeID       string `json:"node_id" jsonschema:"the ID of the Kubernetes node"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// Dynamic load-mode meta-tool input types. Used only when the server runs in
// 'dynamic' mode, where these are the only tools the client sees and the full
// product catalog is browsed/invoked through them.

type SearchToolsInput struct {
	Query string  `json:"query" jsonschema:"keywords to match against tool names and descriptions; leave empty to browse, optionally with group"`
	Group *string `json:"group,omitempty" jsonschema:"restrict results to a single product group (e.g. compute, dns, billing, cert, activitylog, objectstorage, k8s)"`
	Limit *int    `json:"limit,omitempty" jsonschema:"maximum number of results to return; omit for the default of 10, or pass 0 for no limit. Must not be negative."`
}

type DescribeToolsInput struct {
	Names []string `json:"names" jsonschema:"the exact names of the tools to describe; returns each tool's description and full JSON input schema"`
}

type CallToolInput struct {
	Name      string         `json:"name" jsonschema:"the exact name of the tool to invoke (from ionos_search_tools)"`
	Arguments map[string]any `json:"arguments,omitempty" jsonschema:"the tool's arguments as a JSON object; see ionos_describe_tools for the schema"`
}
