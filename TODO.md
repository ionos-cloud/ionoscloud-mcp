# IONOS Cloud MCP Server - Implementation TODO

This document tracks the implementation progress for adding full CRUD operations to the IONOS Cloud MCP server. Each task includes implementing the function AND its corresponding end-to-end tests.

## Current Status

The server currently has **42 READ-only tools** (list/get operations). This plan adds **CREATE, UPDATE, DELETE, and ACTION operations** for full infrastructure management.

---

## Priority 1: Core Infrastructure CRUD ✅ COMPLETED

### 1.1 Datacenter Operations
- [x] `create_datacenter` - Create a new virtual datacenter
  - Parameters: name, description, location
  - Test: Create datacenter, verify, then delete
- [x] `update_datacenter` - Update datacenter name/description
  - Parameters: datacenter_id, name, description
  - Test: Update existing datacenter, verify changes
- [x] `delete_datacenter` - Delete a datacenter
  - Parameters: datacenter_id
  - Test: Delete datacenter created in tests

### 1.2 Server Operations
- [x] `create_server` - Create a new server in a datacenter
  - Parameters: datacenter_id, name, cores, ram, availability_zone, cpu_family
  - Test: Create server, verify properties
- [x] `update_server` - Update server properties
  - Parameters: datacenter_id, server_id, name, cores, ram
  - Test: Update server, verify changes
- [x] `delete_server` - Delete a server
  - Parameters: datacenter_id, server_id
  - Test: Delete server created in tests
- [x] `start_server` - Power on a server
  - Parameters: datacenter_id, server_id
  - Test: Start stopped server, verify state
- [x] `stop_server` - Power off a server
  - Parameters: datacenter_id, server_id
  - Test: Stop running server, verify state
- [x] `reboot_server` - Reboot a server
  - Parameters: datacenter_id, server_id
  - Test: Reboot server, verify state transitions

### 1.3 Volume Operations
- [x] `create_volume` - Create a new volume
  - Parameters: datacenter_id, name, size, type (HDD/SSD), image (optional), image_password (optional)
  - Test: Create volume, verify properties
- [x] `update_volume` - Update volume properties
  - Parameters: datacenter_id, volume_id, name, size
  - Test: Update volume size/name
- [x] `delete_volume` - Delete a volume
  - Parameters: datacenter_id, volume_id
  - Test: Delete volume created in tests
- [x] `attach_volume` - Attach volume to a server
  - Parameters: datacenter_id, server_id, volume_id
  - Test: Attach volume, verify attachment
- [x] `detach_volume` - Detach volume from a server
  - Parameters: datacenter_id, server_id, volume_id
  - Test: Detach volume, verify detachment

### 1.4 Snapshot Operations
- [x] `create_snapshot` - Create snapshot of a volume
  - Parameters: datacenter_id, volume_id, name, description
  - Test: Create snapshot, verify it exists
- [x] `update_snapshot` - Update snapshot metadata
  - Parameters: snapshot_id, name, description
  - Test: Update snapshot name/description
- [x] `delete_snapshot` - Delete a snapshot
  - Parameters: snapshot_id
  - Test: Delete snapshot created in tests
- [x] `restore_snapshot` - Restore volume from snapshot
  - Parameters: snapshot_id, volume_id
  - Test: Restore snapshot to volume

---

## Priority 2: Networking CRUD ✅ COMPLETED

### 2.1 LAN Operations
- [x] `create_lan` - Create a LAN in a datacenter
  - Parameters: datacenter_id, name, public (boolean)
  - Test: Create LAN, verify properties
- [x] `update_lan` - Update LAN properties
  - Parameters: datacenter_id, lan_id, name, public
  - Test: Update LAN, verify changes
- [x] `delete_lan` - Delete a LAN
  - Parameters: datacenter_id, lan_id
  - Test: Delete LAN created in tests

### 2.2 NIC Operations
- [x] `create_nic` - Create a NIC on a server
  - Parameters: datacenter_id, server_id, name, lan, dhcp (boolean), ips (optional)
  - Test: Create NIC, verify attachment to server
- [x] `update_nic` - Update NIC properties
  - Parameters: datacenter_id, server_id, nic_id, name, lan, ips
  - Test: Update NIC, verify changes
- [x] `delete_nic` - Delete a NIC
  - Parameters: datacenter_id, server_id, nic_id
  - Test: Delete NIC created in tests

### 2.3 IP Block Operations
- [x] `create_ipblock` - Reserve a block of public IPs
  - Parameters: location, size, name
  - Test: Reserve IPs, verify allocation
- [x] `update_ipblock` - Update IP block name
  - Parameters: ipblock_id, name
  - Test: Update IP block name
- [x] `delete_ipblock` - Release an IP block
  - Parameters: ipblock_id
  - Test: Release IP block created in tests

### 2.4 Firewall Rule Operations
- [x] `create_firewall_rule` - Create a firewall rule on a NIC
  - Parameters: datacenter_id, server_id, nic_id, name, protocol, source_mac, source_ip, target_ip, port_range_start, port_range_end, icmp_type, icmp_code, type (INGRESS/EGRESS)
  - Test: Create firewall rule, verify it exists
- [x] `update_firewall_rule` - Update firewall rule
  - Parameters: datacenter_id, server_id, nic_id, firewallrule_id, + same as create
  - Test: Update rule, verify changes
- [x] `delete_firewall_rule` - Delete a firewall rule
  - Parameters: datacenter_id, server_id, nic_id, firewallrule_id
  - Test: Delete rule created in tests

---

## Priority 3: Advanced Networking

### 3.1 NAT Gateway Operations
- [ ] `create_nat_gateway` - Create a NAT gateway
  - Parameters: datacenter_id, name, public_ips, lans
  - Test: Create NAT gateway, verify configuration
- [ ] `update_nat_gateway` - Update NAT gateway
  - Parameters: datacenter_id, nat_gateway_id, name, public_ips, lans
  - Test: Update NAT gateway, verify changes
- [ ] `delete_nat_gateway` - Delete a NAT gateway
  - Parameters: datacenter_id, nat_gateway_id
  - Test: Delete NAT gateway created in tests

### 3.2 NAT Gateway Rules
- [ ] `list_nat_gateway_rules` - List NAT gateway rules
  - Parameters: datacenter_id, nat_gateway_id
  - Test: List rules for existing gateway
- [ ] `get_nat_gateway_rule` - Get specific NAT rule
  - Parameters: datacenter_id, nat_gateway_id, rule_id
  - Test: Get rule details
- [ ] `create_nat_gateway_rule` - Create NAT rule
  - Parameters: datacenter_id, nat_gateway_id, name, type, protocol, source_subnet, public_ip, target_subnet, target_port_range
  - Test: Create rule, verify it exists
- [ ] `update_nat_gateway_rule` - Update NAT rule
  - Parameters: datacenter_id, nat_gateway_id, rule_id, + same as create
  - Test: Update rule, verify changes
- [ ] `delete_nat_gateway_rule` - Delete NAT rule
  - Parameters: datacenter_id, nat_gateway_id, rule_id
  - Test: Delete rule created in tests

### 3.3 Private Cross-Connect Operations
- [ ] `create_pcc` - Create a Private Cross-Connect
  - Parameters: name, description
  - Test: Create PCC, verify properties
- [ ] `update_pcc` - Update PCC
  - Parameters: pcc_id, name, description
  - Test: Update PCC, verify changes
- [ ] `delete_pcc` - Delete a PCC
  - Parameters: pcc_id
  - Test: Delete PCC created in tests

---

## Priority 4: Load Balancers

### 4.1 Application Load Balancer Operations
- [ ] `create_application_load_balancer` - Create an ALB
  - Parameters: datacenter_id, name, listener_lan, target_lan, ips
  - Test: Create ALB, verify configuration
- [ ] `update_application_load_balancer` - Update ALB
  - Parameters: datacenter_id, alb_id, name, listener_lan, target_lan
  - Test: Update ALB, verify changes
- [ ] `delete_application_load_balancer` - Delete an ALB
  - Parameters: datacenter_id, alb_id
  - Test: Delete ALB created in tests

### 4.2 ALB Forwarding Rules
- [ ] `create_alb_forwarding_rule` - Create ALB forwarding rule
  - Parameters: datacenter_id, alb_id, name, protocol, listener_ip, listener_port, client_timeout, server_certificates, http_rules
  - Test: Create rule, verify configuration
- [ ] `update_alb_forwarding_rule` - Update ALB forwarding rule
  - Parameters: datacenter_id, alb_id, rule_id, + same as create
  - Test: Update rule, verify changes
- [ ] `delete_alb_forwarding_rule` - Delete ALB forwarding rule
  - Parameters: datacenter_id, alb_id, rule_id
  - Test: Delete rule created in tests

### 4.3 Network Load Balancer Operations
- [ ] `create_network_load_balancer` - Create an NLB
  - Parameters: datacenter_id, name, listener_lan, target_lan, ips
  - Test: Create NLB, verify configuration
- [ ] `update_network_load_balancer` - Update NLB
  - Parameters: datacenter_id, nlb_id, name, listener_lan, target_lan, ips
  - Test: Update NLB, verify changes
- [ ] `delete_network_load_balancer` - Delete an NLB
  - Parameters: datacenter_id, nlb_id
  - Test: Delete NLB created in tests

### 4.4 NLB Forwarding Rules
- [ ] `list_nlb_forwarding_rules` - List NLB forwarding rules
  - Parameters: datacenter_id, nlb_id
  - Test: List rules for existing NLB
- [ ] `get_nlb_forwarding_rule` - Get specific NLB forwarding rule
  - Parameters: datacenter_id, nlb_id, rule_id
  - Test: Get rule details
- [ ] `create_nlb_forwarding_rule` - Create NLB forwarding rule
  - Parameters: datacenter_id, nlb_id, name, algorithm, protocol, listener_ip, listener_port, health_check, targets
  - Test: Create rule, verify configuration
- [ ] `update_nlb_forwarding_rule` - Update NLB forwarding rule
  - Parameters: datacenter_id, nlb_id, rule_id, + same as create
  - Test: Update rule, verify changes
- [ ] `delete_nlb_forwarding_rule` - Delete NLB forwarding rule
  - Parameters: datacenter_id, nlb_id, rule_id
  - Test: Delete rule created in tests

### 4.5 Target Group Operations
- [ ] `create_target_group` - Create a target group
  - Parameters: name, algorithm, protocol, health_check, http_health_check, targets
  - Test: Create target group, verify configuration
- [ ] `update_target_group` - Update target group
  - Parameters: target_group_id, name, algorithm, protocol, health_check, targets
  - Test: Update target group, verify changes
- [ ] `delete_target_group` - Delete a target group
  - Parameters: target_group_id
  - Test: Delete target group created in tests

---

## Priority 5: Kubernetes Operations

### 5.1 K8s Cluster Operations
- [ ] `create_k8s_cluster` - Create a Kubernetes cluster
  - Parameters: name, k8s_version, maintenance_window
  - Test: Create cluster, verify state (long running)
- [ ] `update_k8s_cluster` - Update cluster properties
  - Parameters: k8s_cluster_id, name, k8s_version, maintenance_window
  - Test: Update cluster, verify changes
- [ ] `delete_k8s_cluster` - Delete a Kubernetes cluster
  - Parameters: k8s_cluster_id
  - Test: Delete cluster created in tests

### 5.2 K8s Node Pool Operations
- [ ] `create_k8s_nodepool` - Create a node pool
  - Parameters: k8s_cluster_id, name, datacenter_id, node_count, cpu_family, cores_count, ram_size, availability_zone, storage_type, storage_size, k8s_version
  - Test: Create node pool, verify configuration
- [ ] `update_k8s_nodepool` - Update node pool
  - Parameters: k8s_cluster_id, nodepool_id, name, node_count, maintenance_window, auto_scaling
  - Test: Update node pool, verify changes
- [ ] `delete_k8s_nodepool` - Delete a node pool
  - Parameters: k8s_cluster_id, nodepool_id
  - Test: Delete node pool created in tests

### 5.3 K8s Node Operations
- [ ] `delete_k8s_node` - Delete a specific node
  - Parameters: k8s_cluster_id, nodepool_id, node_id
  - Test: Delete node (triggers replacement)
- [ ] `replace_k8s_node` - Replace a node (recreate)
  - Parameters: k8s_cluster_id, nodepool_id, node_id
  - Test: Replace node, verify new node created

---

## Priority 6: User & Access Management

### 6.1 User Operations
- [ ] `create_user` - Create a new user
  - Parameters: firstname, lastname, email, password, administrator, force_sec_auth
  - Test: Create user, verify properties
- [ ] `update_user` - Update user properties
  - Parameters: user_id, firstname, lastname, email, administrator, force_sec_auth
  - Test: Update user, verify changes
- [ ] `delete_user` - Delete a user
  - Parameters: user_id
  - Test: Delete user created in tests

### 6.2 Group Operations
- [ ] `create_group` - Create a group
  - Parameters: name, create_datacenter, create_snapshot, reserve_ip, access_activity_log
  - Test: Create group, verify permissions
- [ ] `update_group` - Update group properties
  - Parameters: group_id, name, create_datacenter, create_snapshot, reserve_ip, access_activity_log
  - Test: Update group, verify changes
- [ ] `delete_group` - Delete a group
  - Parameters: group_id
  - Test: Delete group created in tests
- [ ] `add_user_to_group` - Add a user to a group
  - Parameters: group_id, user_id
  - Test: Add user, verify membership
- [ ] `remove_user_from_group` - Remove user from group
  - Parameters: group_id, user_id
  - Test: Remove user, verify removal

### 6.3 S3 Key Operations
- [ ] `create_s3_key` - Create S3 key for user
  - Parameters: user_id
  - Test: Create key, verify it exists
- [ ] `update_s3_key` - Update S3 key (activate/deactivate)
  - Parameters: user_id, key_id, active (boolean)
  - Test: Toggle key state
- [ ] `delete_s3_key` - Delete an S3 key
  - Parameters: user_id, key_id
  - Test: Delete key created in tests

### 6.4 Resource Shares (Group Permissions)
- [ ] `list_group_shares` - List resources shared with group
  - Parameters: group_id
  - Test: List shares for existing group
- [ ] `get_group_share` - Get specific share
  - Parameters: group_id, resource_id
  - Test: Get share details
- [ ] `create_group_share` - Share resource with group
  - Parameters: group_id, resource_id, edit_privilege, share_privilege
  - Test: Create share, verify permissions
- [ ] `update_group_share` - Update share permissions
  - Parameters: group_id, resource_id, edit_privilege, share_privilege
  - Test: Update share, verify changes
- [ ] `delete_group_share` - Remove resource share
  - Parameters: group_id, resource_id
  - Test: Delete share created in tests

---

## Priority 7: DNS Operations

### 7.1 DNS Zone Operations
- [ ] `create_dns_zone` - Create a DNS zone
  - Parameters: name, description
  - Test: Create zone, verify properties
- [ ] `update_dns_zone` - Update DNS zone
  - Parameters: zone_id, description
  - Test: Update zone, verify changes
- [ ] `delete_dns_zone` - Delete a DNS zone
  - Parameters: zone_id
  - Test: Delete zone created in tests

### 7.2 DNS Record Operations
- [ ] `create_dns_record` - Create a DNS record
  - Parameters: zone_id, name, type (A/AAAA/CNAME/MX/TXT/etc), content, ttl, priority, enabled
  - Test: Create record, verify it resolves
- [ ] `update_dns_record` - Update DNS record
  - Parameters: zone_id, record_id, name, content, ttl, priority, enabled
  - Test: Update record, verify changes
- [ ] `delete_dns_record` - Delete a DNS record
  - Parameters: zone_id, record_id
  - Test: Delete record created in tests

### 7.3 Secondary Zone Operations
- [ ] `list_secondary_zones` - List secondary DNS zones
  - Test: List existing secondary zones
- [ ] `create_secondary_zone` - Create secondary zone
  - Parameters: name, description, primary_ips
  - Test: Create zone, verify configuration
- [ ] `get_secondary_zone` - Get secondary zone details
  - Parameters: zone_id
  - Test: Get zone details
- [ ] `update_secondary_zone` - Update secondary zone
  - Parameters: zone_id, description, primary_ips
  - Test: Update zone, verify changes
- [ ] `delete_secondary_zone` - Delete secondary zone
  - Parameters: zone_id
  - Test: Delete zone created in tests

### 7.4 Reverse DNS Operations
- [ ] `list_reverse_records` - List reverse DNS records
  - Test: List existing reverse records
- [ ] `create_reverse_record` - Create reverse DNS record
  - Parameters: ip, name, description
  - Test: Create record, verify configuration
- [ ] `get_reverse_record` - Get reverse record
  - Parameters: record_id
  - Test: Get record details
- [ ] `update_reverse_record` - Update reverse record
  - Parameters: record_id, name, description
  - Test: Update record, verify changes
- [ ] `delete_reverse_record` - Delete reverse record
  - Parameters: record_id
  - Test: Delete record created in tests

---

## Priority 8: Additional Features

### 8.1 Image Operations
- [ ] `update_image` - Update image metadata
  - Parameters: image_id, name, description, cpu_hot_plug, ram_hot_plug, nic_hot_plug, disc_virtio_hot_plug
  - Test: Update custom image metadata
- [ ] `delete_image` - Delete a custom image
  - Parameters: image_id
  - Test: Delete custom image (if exists)

### 8.2 Flow Logs (Network Observability)
- [ ] `list_nic_flowlogs` - List flow logs for a NIC
  - Parameters: datacenter_id, server_id, nic_id
  - Test: List flow logs
- [ ] `create_nic_flowlog` - Create flow log for NIC
  - Parameters: datacenter_id, server_id, nic_id, name, action, direction, bucket
  - Test: Create flow log, verify configuration
- [ ] `get_nic_flowlog` - Get flow log details
  - Parameters: datacenter_id, server_id, nic_id, flowlog_id
  - Test: Get flow log details
- [ ] `update_nic_flowlog` - Update flow log
  - Parameters: datacenter_id, server_id, nic_id, flowlog_id, name, action, direction, bucket
  - Test: Update flow log, verify changes
- [ ] `delete_nic_flowlog` - Delete flow log
  - Parameters: datacenter_id, server_id, nic_id, flowlog_id
  - Test: Delete flow log created in tests

### 8.3 Request Status (Async Operations)
- [ ] `get_request_status` - Check status of async operation
  - Parameters: request_id
  - Test: Check status of create/update/delete operations

### 8.4 Labels
- [ ] `list_labels` - List labels on a resource
  - Parameters: resource_type, resource_id
  - Test: List labels
- [ ] `create_label` - Add label to a resource
  - Parameters: resource_type, resource_id, key, value
  - Test: Create label, verify it exists
- [ ] `delete_label` - Remove label from a resource
  - Parameters: resource_type, resource_id, key
  - Test: Delete label created in tests

---

## Implementation Notes

### Async Operations
Most CREATE/UPDATE/DELETE operations return HTTP 202 (Accepted) and are asynchronous. The implementation should:
1. Submit the request
2. Poll the request status endpoint until complete
3. Return the final result or error

### Testing Strategy
- Tests should be organized by resource type
- Use a dedicated test datacenter to avoid affecting production
- Clean up all resources created during tests
- Tests can run against real IONOS Cloud API (requires credentials)

### Error Handling
All operations should handle common IONOS Cloud error codes:
- 400: Bad Request (invalid parameters)
- 401: Unauthorized (invalid/missing credentials)
- 403: Forbidden (insufficient permissions)
- 404: Not Found (resource doesn't exist)
- 422: Unprocessable Entity (validation failed)
- 429: Too Many Requests (rate limited)
- 500: Internal Server Error
- 503: Service Unavailable

---

## Summary

| Priority | Category | Tools to Add | Est. Complexity |
|----------|----------|--------------|-----------------|
| 1 | Core Infrastructure | 17 | High |
| 2 | Networking | 12 | Medium |
| 3 | Advanced Networking | 11 | Medium |
| 4 | Load Balancers | 17 | High |
| 5 | Kubernetes | 7 | High |
| 6 | User Management | 15 | Medium |
| 7 | DNS | 15 | Medium |
| 8 | Additional | 12 | Low |
| **Total** | | **106** | |

The implementation adds **106 new tools** to complement the existing **42 read-only tools**, bringing the total to **148 tools** for comprehensive IONOS Cloud infrastructure management.
