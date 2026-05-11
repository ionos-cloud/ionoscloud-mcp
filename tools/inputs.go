package tools

// Input types for tool parameters.
// Each struct maps to the JSON schema that the MCP SDK auto-generates
// for a tool's input. Non-pointer fields are treated as required.

// Compute input types

type DatacenterIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
}

type ServerIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
}

type VolumeIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	VolumeID     string `json:"volume_id" jsonschema:"the ID of the volume"`
}

type SnapshotIDInput struct {
	SnapshotID string `json:"snapshot_id" jsonschema:"the ID of the snapshot"`
}

type NicIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	NicID        string `json:"nic_id" jsonschema:"the ID of the network interface"`
}

type LanIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LanID        string `json:"lan_id" jsonschema:"the ID of the LAN"`
}

type FirewallRuleIDInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID       string `json:"server_id" jsonschema:"the ID of the server"`
	NicID          string `json:"nic_id" jsonschema:"the ID of the network interface"`
	FirewallRuleID string `json:"firewallrule_id" jsonschema:"the ID of the firewall rule"`
}

type IpBlockIDInput struct {
	IpBlockID string `json:"ipblock_id" jsonschema:"the ID of the IP block"`
}

type LoadBalancerIDInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID string `json:"loadbalancer_id" jsonschema:"the ID of the load balancer"`
}

type NetworkLoadBalancerIDInput struct {
	DatacenterID          string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NetworkLoadBalancerID string `json:"network_loadbalancer_id" jsonschema:"the ID of the network load balancer"`
}

type ApplicationLoadBalancerIDInput struct {
	DatacenterID              string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ApplicationLoadBalancerID string `json:"application_loadbalancer_id" jsonschema:"the ID of the application load balancer"`
}

type TargetGroupIDInput struct {
	TargetGroupID string `json:"target_group_id" jsonschema:"the ID of the target group"`
}

type NatGatewayIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NatGatewayID string `json:"nat_gateway_id" jsonschema:"the ID of the NAT gateway"`
}

type PccIDInput struct {
	PccID string `json:"pcc_id" jsonschema:"the ID of the private cross-connect"`
}

type GpuIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	GpuID        string `json:"gpu_id" jsonschema:"the ID of the GPU"`
}

type SecurityGroupIDInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group"`
}

type SecurityGroupRuleIDInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group"`
	RuleID          string `json:"rule_id" jsonschema:"the ID of the security group rule"`
}

type RequestIDInput struct {
	RequestID string `json:"request_id" jsonschema:"the ID of the request"`
}

type TemplateIDInput struct {
	TemplateID string `json:"template_id" jsonschema:"the ID of the template"`
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
