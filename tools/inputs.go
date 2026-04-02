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
