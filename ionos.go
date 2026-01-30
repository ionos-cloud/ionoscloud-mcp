package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"

	dns "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// Valid IONOS Cloud locations
var validLocations = map[string]bool{
	"de/fra": true, "de/txl": true, "us/las": true, "us/ewr": true,
	"gb/lhr": true, "es/vit": true, "fr/par": true,
}

// Valid firewall protocols
var validProtocols = map[string]bool{
	"TCP": true, "UDP": true, "ICMP": true, "ICMPv6": true,
	"GRE": true, "ESP": true, "AH": true, "ANY": true,
}

// MAC address regex pattern
var macAddressRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)

// validateIP validates an IP address or CIDR
func validateIP(ip string) error {
	if ip == "" {
		return nil
	}
	// Try parsing as IP address
	if net.ParseIP(ip) != nil {
		return nil
	}
	// Try parsing as CIDR
	if _, _, err := net.ParseCIDR(ip); err == nil {
		return nil
	}
	return fmt.Errorf("invalid IP address or CIDR: %s", ip)
}

// validateMAC validates a MAC address
func validateMAC(mac string) error {
	if mac == "" {
		return nil
	}
	if !macAddressRegex.MatchString(mac) {
		return fmt.Errorf("invalid MAC address format: %s (expected XX:XX:XX:XX:XX:XX)", mac)
	}
	return nil
}

// validateLocation validates an IONOS location
func validateLocation(location string) error {
	if !validLocations[location] {
		return fmt.Errorf("invalid location: %s (valid: de/fra, de/txl, us/las, us/ewr, gb/lhr, es/vit, fr/par)", location)
	}
	return nil
}

// validateProtocol validates a firewall protocol
func validateProtocol(protocol string) error {
	if !validProtocols[protocol] {
		return fmt.Errorf("invalid protocol: %s (valid: TCP, UDP, ICMP, ICMPv6, GRE, ESP, AH, ANY)", protocol)
	}
	return nil
}

// validatePortRange validates port range values
func validatePortRange(start, end int32, startSet, endSet bool) error {
	if startSet {
		if start < 1 || start > 65535 {
			return fmt.Errorf("port_range_start must be between 1-65535, got %d", start)
		}
	}
	if endSet {
		if end < 1 || end > 65535 {
			return fmt.Errorf("port_range_end must be between 1-65535, got %d", end)
		}
	}
	if startSet && endSet && start > end {
		return fmt.Errorf("port_range_start (%d) cannot be greater than port_range_end (%d)", start, end)
	}
	return nil
}

// validateICMP validates ICMP type and code
func validateICMP(icmpType, icmpCode int32, typeSet, codeSet bool) error {
	if typeSet {
		if icmpType < 0 || icmpType > 255 {
			return fmt.Errorf("icmp_type must be between 0-255, got %d", icmpType)
		}
	}
	if codeSet {
		if icmpCode < 0 || icmpCode > 255 {
			return fmt.Errorf("icmp_code must be between 0-255, got %d", icmpCode)
		}
	}
	return nil
}

// marshalResponse marshals any value to indented JSON
func marshalResponse(v interface{}, resourceName string) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal %s: %w", resourceName, err)
	}
	return string(data), nil
}

// statusResponse creates a JSON response for status messages
func statusResponse(fields map[string]string) (string, error) {
	data, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}
	return string(data), nil
}

func (s *Server) executeTool(name string, arguments map[string]interface{}) (string, error) {
	switch name {
	case "list_datacenters":
		return s.listDatacenters(s.client, s.ctx)
	case "get_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.getDatacenter(s.client, s.ctx, datacenterID)
	case "create_datacenter":
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		location, ok := arguments["location"].(string)
		if !ok {
			return "", fmt.Errorf("location is required")
		}
		description, _ := arguments["description"].(string)
		return s.createDatacenter(s.client, s.ctx, name, location, description)
	case "update_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updateDatacenter(s.client, s.ctx, datacenterID, name, description)
	case "delete_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.deleteDatacenter(s.client, s.ctx, datacenterID)
	case "list_servers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listServers(s.client, s.ctx, datacenterID)
	case "get_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.getServer(s.client, s.ctx, datacenterID, serverID)
	case "create_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		cores, ok := arguments["cores"].(float64)
		if !ok {
			return "", fmt.Errorf("cores is required")
		}
		ram, ok := arguments["ram"].(float64)
		if !ok {
			return "", fmt.Errorf("ram is required")
		}
		cpuFamily, _ := arguments["cpu_family"].(string)
		availabilityZone, _ := arguments["availability_zone"].(string)
		return s.createServer(s.client, s.ctx, datacenterID, name, int32(cores), int32(ram), cpuFamily, availabilityZone)
	case "update_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		name, _ := arguments["name"].(string)
		var cores, ram int32
		if c, ok := arguments["cores"].(float64); ok {
			cores = int32(c)
		}
		if r, ok := arguments["ram"].(float64); ok {
			ram = int32(r)
		}
		return s.updateServer(s.client, s.ctx, datacenterID, serverID, name, cores, ram)
	case "delete_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.deleteServer(s.client, s.ctx, datacenterID, serverID)
	case "start_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.startServer(s.client, s.ctx, datacenterID, serverID)
	case "stop_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.stopServer(s.client, s.ctx, datacenterID, serverID)
	case "reboot_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.rebootServer(s.client, s.ctx, datacenterID, serverID)
	case "list_volumes":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listVolumes(s.client, s.ctx, datacenterID)
	case "get_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.getVolume(s.client, s.ctx, datacenterID, volumeID)
	case "create_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		size, ok := arguments["size"].(float64)
		if !ok {
			return "", fmt.Errorf("size is required")
		}
		volumeType, _ := arguments["type"].(string)
		bus, _ := arguments["bus"].(string)
		availabilityZone, _ := arguments["availability_zone"].(string)
		image, _ := arguments["image"].(string)
		imagePassword, _ := arguments["image_password"].(string)
		licenceType, _ := arguments["licence_type"].(string)
		return s.createVolume(s.client, s.ctx, datacenterID, name, float32(size), volumeType, bus, availabilityZone, image, imagePassword, licenceType)
	case "update_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		name, _ := arguments["name"].(string)
		var size float32
		if sizeFloat, ok := arguments["size"].(float64); ok {
			size = float32(sizeFloat)
		}
		return s.updateVolume(s.client, s.ctx, datacenterID, volumeID, name, size)
	case "delete_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.deleteVolume(s.client, s.ctx, datacenterID, volumeID)
	case "attach_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.attachVolume(s.client, s.ctx, datacenterID, serverID, volumeID)
	case "detach_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.detachVolume(s.client, s.ctx, datacenterID, serverID, volumeID)
	case "list_images":
		return s.listImages(s.client, s.ctx)
	case "list_locations":
		return s.listLocations(s.client, s.ctx)
	case "list_snapshots":
		return s.listSnapshots(s.client, s.ctx)
	case "get_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.getSnapshot(s.client, s.ctx, snapshotID)
	case "create_snapshot":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.createSnapshot(s.client, s.ctx, datacenterID, volumeID, name, description)
	case "update_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updateSnapshot(s.client, s.ctx, snapshotID, name, description)
	case "delete_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.deleteSnapshot(s.client, s.ctx, snapshotID)
	case "restore_snapshot":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.restoreSnapshot(s.client, s.ctx, datacenterID, volumeID, snapshotID)
	// Networking - LANs
	case "list_lans":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listLans(s.client, s.ctx, datacenterID)
	case "get_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		lanID, ok := arguments["lan_id"].(string)
		if !ok {
			return "", fmt.Errorf("lan_id is required")
		}
		return s.getLan(s.client, s.ctx, datacenterID, lanID)
	case "create_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, _ := arguments["name"].(string)
		public, _ := arguments["public"].(bool)
		return s.createLan(s.client, s.ctx, datacenterID, name, public)
	case "update_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		lanID, ok := arguments["lan_id"].(string)
		if !ok {
			return "", fmt.Errorf("lan_id is required")
		}
		name, _ := arguments["name"].(string)
		public, publicSet := arguments["public"].(bool)
		return s.updateLan(s.client, s.ctx, datacenterID, lanID, name, public, publicSet)
	case "delete_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		lanID, ok := arguments["lan_id"].(string)
		if !ok {
			return "", fmt.Errorf("lan_id is required")
		}
		return s.deleteLan(s.client, s.ctx, datacenterID, lanID)
	// Networking - NICs
	case "list_nics":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.listNics(s.client, s.ctx, datacenterID, serverID)
	case "get_nic":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		return s.getNic(s.client, s.ctx, datacenterID, serverID, nicID)
	case "create_nic":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		lan, ok := arguments["lan"].(float64)
		if !ok {
			return "", fmt.Errorf("lan is required")
		}
		name, _ := arguments["name"].(string)
		dhcp, _ := arguments["dhcp"].(bool)
		// Handle ips as a slice of interfaces and convert to strings
		var ips []string
		if ipsRaw, ok := arguments["ips"].([]interface{}); ok {
			for _, ip := range ipsRaw {
				if ipStr, ok := ip.(string); ok {
					ips = append(ips, ipStr)
				}
			}
		}
		return s.createNic(s.client, s.ctx, datacenterID, serverID, name, int32(lan), dhcp, ips)
	case "update_nic":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		name, _ := arguments["name"].(string)
		var lan int32
		if lanFloat, ok := arguments["lan"].(float64); ok {
			lan = int32(lanFloat)
		}
		dhcp, dhcpSet := arguments["dhcp"].(bool)
		var ips []string
		if ipsRaw, ok := arguments["ips"].([]interface{}); ok {
			for _, ip := range ipsRaw {
				if ipStr, ok := ip.(string); ok {
					ips = append(ips, ipStr)
				}
			}
		}
		return s.updateNic(s.client, s.ctx, datacenterID, serverID, nicID, name, lan, dhcp, dhcpSet, ips)
	case "delete_nic":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		return s.deleteNic(s.client, s.ctx, datacenterID, serverID, nicID)
	// Networking - IP Blocks
	case "list_ipblocks":
		return s.listIpBlocks(s.client, s.ctx)
	case "get_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		return s.getIpBlock(s.client, s.ctx, ipblockID)
	case "create_ipblock":
		location, ok := arguments["location"].(string)
		if !ok {
			return "", fmt.Errorf("location is required")
		}
		size, ok := arguments["size"].(float64)
		if !ok {
			return "", fmt.Errorf("size is required")
		}
		name, _ := arguments["name"].(string)
		return s.createIpBlock(s.client, s.ctx, location, int32(size), name)
	case "update_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		name, _ := arguments["name"].(string)
		return s.updateIpBlock(s.client, s.ctx, ipblockID, name)
	case "delete_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		return s.deleteIpBlock(s.client, s.ctx, ipblockID)
	// Networking - Firewall Rules
	case "list_firewall_rules":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		return s.listFirewallRules(s.client, s.ctx, datacenterID, serverID, nicID)
	case "get_firewall_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		firewallRuleID, ok := arguments["firewallrule_id"].(string)
		if !ok {
			return "", fmt.Errorf("firewallrule_id is required")
		}
		return s.getFirewallRule(s.client, s.ctx, datacenterID, serverID, nicID, firewallRuleID)
	case "create_firewall_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		protocol, ok := arguments["protocol"].(string)
		if !ok {
			return "", fmt.Errorf("protocol is required")
		}
		name, _ := arguments["name"].(string)
		sourceMac, _ := arguments["source_mac"].(string)
		sourceIP, _ := arguments["source_ip"].(string)
		targetIP, _ := arguments["target_ip"].(string)
		var portRangeStart, portRangeEnd int32
		if prs, ok := arguments["port_range_start"].(float64); ok {
			portRangeStart = int32(prs)
		}
		if pre, ok := arguments["port_range_end"].(float64); ok {
			portRangeEnd = int32(pre)
		}
		var icmpType, icmpCode int32
		icmpTypeSet, icmpCodeSet := false, false
		if it, ok := arguments["icmp_type"].(float64); ok {
			icmpType = int32(it)
			icmpTypeSet = true
		}
		if ic, ok := arguments["icmp_code"].(float64); ok {
			icmpCode = int32(ic)
			icmpCodeSet = true
		}
		ruleType, _ := arguments["type"].(string)
		return s.createFirewallRule(s.client, s.ctx, datacenterID, serverID, nicID, name, protocol, sourceMac, sourceIP, targetIP, portRangeStart, portRangeEnd, icmpType, icmpCode, icmpTypeSet, icmpCodeSet, ruleType)
	case "update_firewall_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		firewallRuleID, ok := arguments["firewallrule_id"].(string)
		if !ok {
			return "", fmt.Errorf("firewallrule_id is required")
		}
		name, _ := arguments["name"].(string)
		protocol, _ := arguments["protocol"].(string)
		sourceMac, _ := arguments["source_mac"].(string)
		sourceIP, _ := arguments["source_ip"].(string)
		targetIP, _ := arguments["target_ip"].(string)
		var portRangeStart, portRangeEnd int32
		portRangeStartSet, portRangeEndSet := false, false
		if prs, ok := arguments["port_range_start"].(float64); ok {
			portRangeStart = int32(prs)
			portRangeStartSet = true
		}
		if pre, ok := arguments["port_range_end"].(float64); ok {
			portRangeEnd = int32(pre)
			portRangeEndSet = true
		}
		var icmpType, icmpCode int32
		icmpTypeSet, icmpCodeSet := false, false
		if it, ok := arguments["icmp_type"].(float64); ok {
			icmpType = int32(it)
			icmpTypeSet = true
		}
		if ic, ok := arguments["icmp_code"].(float64); ok {
			icmpCode = int32(ic)
			icmpCodeSet = true
		}
		ruleType, _ := arguments["type"].(string)
		return s.updateFirewallRule(s.client, s.ctx, datacenterID, serverID, nicID, firewallRuleID, name, protocol, sourceMac, sourceIP, targetIP, portRangeStart, portRangeEnd, portRangeStartSet, portRangeEndSet, icmpType, icmpCode, icmpTypeSet, icmpCodeSet, ruleType)
	case "delete_firewall_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		firewallRuleID, ok := arguments["firewallrule_id"].(string)
		if !ok {
			return "", fmt.Errorf("firewallrule_id is required")
		}
		return s.deleteFirewallRule(s.client, s.ctx, datacenterID, serverID, nicID, firewallRuleID)
	// Networking - NAT Gateways
	case "list_nat_gateways":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listNatGateways(s.client, s.ctx, datacenterID)
	case "get_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		return s.getNatGateway(s.client, s.ctx, datacenterID, natGatewayID)
	// Networking - Private Cross Connects
	case "list_pccs":
		return s.listPccs(s.client, s.ctx)
	case "get_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		return s.getPcc(s.client, s.ctx, pccID)
	case "create_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		publicIPsInterface, ok := arguments["public_ips"].([]interface{})
		if !ok {
			return "", fmt.Errorf("public_ips is required")
		}
		publicIPs := make([]string, len(publicIPsInterface))
		for i, ip := range publicIPsInterface {
			ipStr, ok := ip.(string)
			if !ok {
				return "", fmt.Errorf("public_ips[%d] must be a string", i)
			}
			publicIPs[i] = ipStr
		}
		var lans []map[string]interface{}
		if lansInterface, ok := arguments["lans"].([]interface{}); ok {
			for i, l := range lansInterface {
				lanMap, ok := l.(map[string]interface{})
				if !ok {
					return "", fmt.Errorf("lans[%d] must be an object", i)
				}
				lans = append(lans, lanMap)
			}
		}
		return s.createNatGateway(s.client, s.ctx, datacenterID, name, publicIPs, lans)
	case "update_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		name, _ := arguments["name"].(string)
		var publicIPs []string
		if publicIPsInterface, ok := arguments["public_ips"].([]interface{}); ok {
			publicIPs = make([]string, len(publicIPsInterface))
			for i, ip := range publicIPsInterface {
				ipStr, ok := ip.(string)
				if !ok {
					return "", fmt.Errorf("public_ips[%d] must be a string", i)
				}
				publicIPs[i] = ipStr
			}
		}
		var lans []map[string]interface{}
		if lansInterface, ok := arguments["lans"].([]interface{}); ok {
			for i, l := range lansInterface {
				lanMap, ok := l.(map[string]interface{})
				if !ok {
					return "", fmt.Errorf("lans[%d] must be an object", i)
				}
				lans = append(lans, lanMap)
			}
		}
		return s.updateNatGateway(s.client, s.ctx, datacenterID, natGatewayID, name, publicIPs, lans)
	case "delete_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		return s.deleteNatGateway(s.client, s.ctx, datacenterID, natGatewayID)
	// NAT Gateway Rules
	case "list_nat_gateway_rules":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		return s.listNatGatewayRules(s.client, s.ctx, datacenterID, natGatewayID)
	case "get_nat_gateway_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		ruleID, ok := arguments["rule_id"].(string)
		if !ok {
			return "", fmt.Errorf("rule_id is required")
		}
		return s.getNatGatewayRule(s.client, s.ctx, datacenterID, natGatewayID, ruleID)
	case "create_nat_gateway_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		sourceSubnet, ok := arguments["source_subnet"].(string)
		if !ok {
			return "", fmt.Errorf("source_subnet is required")
		}
		publicIP, ok := arguments["public_ip"].(string)
		if !ok {
			return "", fmt.Errorf("public_ip is required")
		}
		ruleType, _ := arguments["type"].(string)
		protocol, _ := arguments["protocol"].(string)
		targetSubnet, _ := arguments["target_subnet"].(string)
		var targetPortRangeStart, targetPortRangeEnd int32
		if v, ok := arguments["target_port_range_start"].(float64); ok {
			targetPortRangeStart = int32(v)
		}
		if v, ok := arguments["target_port_range_end"].(float64); ok {
			targetPortRangeEnd = int32(v)
		}
		return s.createNatGatewayRule(s.client, s.ctx, datacenterID, natGatewayID, name, ruleType, protocol, sourceSubnet, publicIP, targetSubnet, targetPortRangeStart, targetPortRangeEnd)
	case "update_nat_gateway_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		ruleID, ok := arguments["rule_id"].(string)
		if !ok {
			return "", fmt.Errorf("rule_id is required")
		}
		name, _ := arguments["name"].(string)
		protocol, _ := arguments["protocol"].(string)
		sourceSubnet, _ := arguments["source_subnet"].(string)
		publicIP, _ := arguments["public_ip"].(string)
		targetSubnet, _ := arguments["target_subnet"].(string)
		var targetPortRangeStart, targetPortRangeEnd int32
		var targetPortRangeStartSet, targetPortRangeEndSet bool
		if v, ok := arguments["target_port_range_start"].(float64); ok {
			targetPortRangeStart = int32(v)
			targetPortRangeStartSet = true
		}
		if v, ok := arguments["target_port_range_end"].(float64); ok {
			targetPortRangeEnd = int32(v)
			targetPortRangeEndSet = true
		}
		return s.updateNatGatewayRule(s.client, s.ctx, datacenterID, natGatewayID, ruleID, name, protocol, sourceSubnet, publicIP, targetSubnet, targetPortRangeStart, targetPortRangeEnd, targetPortRangeStartSet, targetPortRangeEndSet)
	case "delete_nat_gateway_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		ruleID, ok := arguments["rule_id"].(string)
		if !ok {
			return "", fmt.Errorf("rule_id is required")
		}
		return s.deleteNatGatewayRule(s.client, s.ctx, datacenterID, natGatewayID, ruleID)
	// Private Cross Connect CRUD
	case "create_pcc":
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		description, _ := arguments["description"].(string)
		return s.createPcc(s.client, s.ctx, name, description)
	case "update_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updatePcc(s.client, s.ctx, pccID, name, description)
	case "delete_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		return s.deletePcc(s.client, s.ctx, pccID)
	// Load Balancers - Application Load Balancers
	case "list_application_load_balancers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listApplicationLoadBalancers(s.client, s.ctx, datacenterID)
	case "get_application_load_balancer":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		return s.getApplicationLoadBalancer(s.client, s.ctx, datacenterID, albID)
	case "list_alb_forwarding_rules":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		return s.listAlbForwardingRules(s.client, s.ctx, datacenterID, albID)
	case "get_alb_forwarding_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		ruleID, ok := arguments["rule_id"].(string)
		if !ok {
			return "", fmt.Errorf("rule_id is required")
		}
		return s.getAlbForwardingRule(s.client, s.ctx, datacenterID, albID, ruleID)
	// Load Balancers - Network Load Balancers
	case "list_network_load_balancers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listNetworkLoadBalancers(s.client, s.ctx, datacenterID)
	case "get_network_load_balancer":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		nlbID, ok := arguments["nlb_id"].(string)
		if !ok {
			return "", fmt.Errorf("nlb_id is required")
		}
		return s.getNetworkLoadBalancer(s.client, s.ctx, datacenterID, nlbID)
	// Load Balancers - Target Groups
	case "list_target_groups":
		return s.listTargetGroups(s.client, s.ctx)
	case "get_target_group":
		targetGroupID, ok := arguments["target_group_id"].(string)
		if !ok {
			return "", fmt.Errorf("target_group_id is required")
		}
		return s.getTargetGroup(s.client, s.ctx, targetGroupID)
	// Kubernetes
	case "list_k8s_clusters":
		return s.listK8sClusters(s.client, s.ctx)
	case "get_k8s_cluster":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.getK8sCluster(s.client, s.ctx, k8sClusterID)
	case "get_k8s_kubeconfig":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.getK8sKubeconfig(s.client, s.ctx, k8sClusterID)
	case "list_k8s_nodepools":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.listK8sNodepools(s.client, s.ctx, k8sClusterID)
	case "get_k8s_nodepool":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		return s.getK8sNodepool(s.client, s.ctx, k8sClusterID, nodepoolID)
	case "list_k8s_nodes":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		return s.listK8sNodes(s.client, s.ctx, k8sClusterID, nodepoolID)
	case "get_k8s_node":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		nodeID, ok := arguments["node_id"].(string)
		if !ok {
			return "", fmt.Errorf("node_id is required")
		}
		return s.getK8sNode(s.client, s.ctx, k8sClusterID, nodepoolID, nodeID)
	case "list_k8s_versions":
		return s.listK8sVersions(s.client, s.ctx)
	// User Management - Users
	case "list_users":
		return s.listUsers(s.client, s.ctx)
	case "get_user":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.getUser(s.client, s.ctx, userID)
	// User Management - Groups
	case "list_groups":
		return s.listGroups(s.client, s.ctx)
	case "get_group":
		groupID, ok := arguments["group_id"].(string)
		if !ok {
			return "", fmt.Errorf("group_id is required")
		}
		return s.getGroup(s.client, s.ctx, groupID)
	case "list_group_members":
		groupID, ok := arguments["group_id"].(string)
		if !ok {
			return "", fmt.Errorf("group_id is required")
		}
		return s.listGroupMembers(s.client, s.ctx, groupID)
	case "list_user_groups":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.listUserGroups(s.client, s.ctx, userID)
	// User Management - S3 Keys
	case "list_s3_keys":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.listS3Keys(s.client, s.ctx, userID)
	case "get_s3_key":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		keyID, ok := arguments["key_id"].(string)
		if !ok {
			return "", fmt.Errorf("key_id is required")
		}
		return s.getS3Key(s.client, s.ctx, userID, keyID)
	// User Management - Contract
	case "get_contract":
		return s.getContract(s.client, s.ctx)
	case "list_resources":
		resourceType, _ := arguments["resource_type"].(string)
		return s.listResources(s.client, s.ctx, resourceType)
	// DNS
	case "list_dns_zones":
		return s.listDnsZones(s.dnsClient, s.ctx)
	case "get_dns_zone":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		return s.getDnsZone(s.dnsClient, s.ctx, zoneID)
	case "list_dns_records":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		return s.listDnsRecords(s.dnsClient, s.ctx, zoneID)
	case "get_dns_record":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		recordID, ok := arguments["record_id"].(string)
		if !ok {
			return "", fmt.Errorf("record_id is required")
		}
		return s.getDnsRecord(s.dnsClient, s.ctx, zoneID, recordID)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) listDatacenters(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	datacenters, _, err := client.DataCentersApi.DatacentersGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list datacenters: %w", err)
	}

	data, err := json.MarshalIndent(datacenters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenters: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	datacenter, _, err := client.DataCentersApi.DatacentersFindById(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get datacenter: %w", err)
	}

	data, err := json.MarshalIndent(datacenter, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) createDatacenter(client *ionoscloud.APIClient, ctx context.Context, name, location, description string) (string, error) {
	properties := ionoscloud.DatacenterPropertiesPost{
		Name:     &name,
		Location: &location,
	}
	if description != "" {
		properties.Description = &description
	}

	datacenter := ionoscloud.DatacenterPost{
		Properties: &properties,
	}

	result, _, err := client.DataCentersApi.DatacentersPost(ctx).Datacenter(datacenter).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create datacenter: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name, description string) (string, error) {
	if name == "" && description == "" {
		return "", fmt.Errorf("at least one of name or description must be provided")
	}

	properties := ionoscloud.DatacenterPropertiesPut{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := client.DataCentersApi.DatacentersPatch(ctx, datacenterID).Datacenter(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update datacenter: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	_, err := client.DataCentersApi.DatacentersDelete(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete datacenter: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "datacenter_id": datacenterID})
}

func (s *Server) listServers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	servers, _, err := client.ServersApi.DatacentersServersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list servers: %w", err)
	}

	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal servers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	server, _, err := client.ServersApi.DatacentersServersFindById(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get server: %w", err)
	}

	data, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) createServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name string, cores, ram int32, cpuFamily, availabilityZone string) (string, error) {
	// Validate inputs
	if cores < 1 {
		return "", fmt.Errorf("cores must be at least 1, got %d", cores)
	}
	if ram < 256 {
		return "", fmt.Errorf("ram must be at least 256 MB, got %d", ram)
	}
	if ram%256 != 0 {
		return "", fmt.Errorf("ram must be a multiple of 256 MB, got %d", ram)
	}

	properties := ionoscloud.ServerProperties{
		Name:  &name,
		Cores: &cores,
		Ram:   &ram,
	}
	if cpuFamily != "" {
		properties.CpuFamily = &cpuFamily
	}
	if availabilityZone != "" {
		properties.AvailabilityZone = &availabilityZone
	}

	server := ionoscloud.Server{
		Properties: &properties,
	}

	result, _, err := client.ServersApi.DatacentersServersPost(ctx, datacenterID).Server(server).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create server: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, name string, cores, ram int32) (string, error) {
	if name == "" && cores == 0 && ram == 0 {
		return "", fmt.Errorf("at least one of name, cores, or ram must be provided")
	}

	properties := ionoscloud.ServerProperties{}
	if name != "" {
		properties.Name = &name
	}
	if cores > 0 {
		properties.Cores = &cores
	}
	if ram > 0 {
		properties.Ram = &ram
	}

	result, _, err := client.ServersApi.DatacentersServersPatch(ctx, datacenterID, serverID).Server(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update server: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersDelete(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete server: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "server_id": serverID})
}

func (s *Server) startServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersStartPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to start server: %w", err)
	}

	return statusResponse(map[string]string{"status": "starting", "server_id": serverID})
}

func (s *Server) stopServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersStopPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to stop server: %w", err)
	}

	return statusResponse(map[string]string{"status": "stopping", "server_id": serverID})
}

func (s *Server) rebootServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersRebootPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to reboot server: %w", err)
	}

	return statusResponse(map[string]string{"status": "rebooting", "server_id": serverID})
}

func (s *Server) listVolumes(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	volumes, _, err := client.VolumesApi.DatacentersVolumesGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list volumes: %w", err)
	}

	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volumes: %w", err)
	}

	return string(data), nil
}

func (s *Server) getVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID string) (string, error) {
	volume, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get volume: %w", err)
	}

	data, err := json.MarshalIndent(volume, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) createVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name string, size float32, volumeType, bus, availabilityZone, image, imagePassword, licenceType string) (string, error) {
	// Validate inputs
	if size < 1 {
		return "", fmt.Errorf("size must be at least 1 GB, got %.1f", size)
	}

	// Validate volume type if provided
	validTypes := map[string]bool{"HDD": true, "SSD": true, "SSD Standard": true, "SSD Premium": true, "DAS": true}
	if volumeType != "" && !validTypes[volumeType] {
		return "", fmt.Errorf("invalid volume type: %s (valid: HDD, SSD, SSD Standard, SSD Premium, DAS)", volumeType)
	}

	// Validate bus type if provided
	validBus := map[string]bool{"VIRTIO": true, "IDE": true}
	if bus != "" && !validBus[bus] {
		return "", fmt.Errorf("invalid bus type: %s (valid: VIRTIO, IDE)", bus)
	}

	properties := ionoscloud.VolumeProperties{
		Name: &name,
		Size: &size,
	}
	if volumeType != "" {
		properties.Type = &volumeType
	}
	if bus != "" {
		properties.Bus = &bus
	}
	if availabilityZone != "" {
		properties.AvailabilityZone = &availabilityZone
	}
	if image != "" {
		properties.Image = &image
	}
	if imagePassword != "" {
		properties.ImagePassword = &imagePassword
	}
	if licenceType != "" {
		properties.LicenceType = &licenceType
	}

	volume := ionoscloud.Volume{
		Properties: &properties,
	}

	result, _, err := client.VolumesApi.DatacentersVolumesPost(ctx, datacenterID).Volume(volume).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID, name string, size float32) (string, error) {
	if name == "" && size == 0 {
		return "", fmt.Errorf("at least one of name or size must be provided")
	}

	properties := ionoscloud.VolumeProperties{}
	if name != "" {
		properties.Name = &name
	}
	if size > 0 {
		properties.Size = &size
	}

	result, _, err := client.VolumesApi.DatacentersVolumesPatch(ctx, datacenterID, volumeID).Volume(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update volume: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID string) (string, error) {
	_, err := client.VolumesApi.DatacentersVolumesDelete(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete volume: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "volume_id": volumeID})
}

func (s *Server) attachVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, volumeID string) (string, error) {
	volume := ionoscloud.Volume{
		Id: &volumeID,
	}

	result, _, err := client.ServersApi.DatacentersServersVolumesPost(ctx, datacenterID, serverID).Volume(volume).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to attach volume: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) detachVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, volumeID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersVolumesDelete(ctx, datacenterID, serverID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to detach volume: %w", err)
	}

	return statusResponse(map[string]string{"status": "detached", "volume_id": volumeID, "server_id": serverID})
}

func (s *Server) listImages(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	images, _, err := client.ImagesApi.ImagesGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}

	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal images: %w", err)
	}

	return string(data), nil
}

func (s *Server) listLocations(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	locations, _, err := client.LocationsApi.LocationsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list locations: %w", err)
	}

	data, err := json.MarshalIndent(locations, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal locations: %w", err)
	}

	return string(data), nil
}

func (s *Server) listSnapshots(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}

	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshots: %w", err)
	}

	return string(data), nil
}

func (s *Server) getSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID string) (string, error) {
	snapshot, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}

func (s *Server) createSnapshot(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID, name, description string) (string, error) {
	properties := ionoscloud.CreateSnapshotProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	snapshot := ionoscloud.CreateSnapshot{
		Properties: &properties,
	}

	result, _, err := client.VolumesApi.DatacentersVolumesCreateSnapshotPost(ctx, datacenterID, volumeID).Snapshot(snapshot).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID, name, description string) (string, error) {
	if name == "" && description == "" {
		return "", fmt.Errorf("at least one of name or description must be provided")
	}

	properties := ionoscloud.SnapshotProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := client.SnapshotsApi.SnapshotsPatch(ctx, snapshotID).Snapshot(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update snapshot: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID string) (string, error) {
	_, err := client.SnapshotsApi.SnapshotsDelete(ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "snapshot_id": snapshotID})
}

func (s *Server) restoreSnapshot(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID, snapshotID string) (string, error) {
	properties := ionoscloud.RestoreSnapshotProperties{
		SnapshotId: &snapshotID,
	}
	restoreSnapshot := ionoscloud.RestoreSnapshot{
		Properties: &properties,
	}

	_, err := client.VolumesApi.DatacentersVolumesRestoreSnapshotPost(ctx, datacenterID, volumeID).RestoreSnapshot(restoreSnapshot).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to restore snapshot: %w", err)
	}

	return statusResponse(map[string]string{"status": "restoring", "volume_id": volumeID, "snapshot_id": snapshotID})
}

// =============================================================================
// Networking - LANs
// =============================================================================

func (s *Server) listLans(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	lans, _, err := client.LANsApi.DatacentersLansGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list LANs: %w", err)
	}

	data, err := json.MarshalIndent(lans, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal LANs: %w", err)
	}

	return string(data), nil
}

func (s *Server) getLan(client *ionoscloud.APIClient, ctx context.Context, datacenterID, lanID string) (string, error) {
	lan, _, err := client.LANsApi.DatacentersLansFindById(ctx, datacenterID, lanID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN: %w", err)
	}

	data, err := json.MarshalIndent(lan, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal LAN: %w", err)
	}

	return string(data), nil
}

func (s *Server) createLan(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name string, public bool) (string, error) {
	properties := ionoscloud.LanProperties{
		Public: &public,
	}
	if name != "" {
		properties.Name = &name
	}

	lan := ionoscloud.Lan{
		Properties: &properties,
	}

	result, _, err := client.LANsApi.DatacentersLansPost(ctx, datacenterID).Lan(lan).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create LAN: %w", err)
	}

	return marshalResponse(result, "LAN")
}

func (s *Server) updateLan(client *ionoscloud.APIClient, ctx context.Context, datacenterID, lanID, name string, public, publicSet bool) (string, error) {
	if name == "" && !publicSet {
		return "", fmt.Errorf("at least one of name or public must be provided")
	}

	properties := ionoscloud.LanProperties{}
	if name != "" {
		properties.Name = &name
	}
	if publicSet {
		properties.Public = &public
	}

	result, _, err := client.LANsApi.DatacentersLansPatch(ctx, datacenterID, lanID).Lan(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update LAN: %w", err)
	}

	return marshalResponse(result, "LAN")
}

func (s *Server) deleteLan(client *ionoscloud.APIClient, ctx context.Context, datacenterID, lanID string) (string, error) {
	_, err := client.LANsApi.DatacentersLansDelete(ctx, datacenterID, lanID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete LAN: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "lan_id": lanID})
}

// =============================================================================
// Networking - NICs
// =============================================================================

func (s *Server) listNics(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	nics, _, err := client.NetworkInterfacesApi.DatacentersServersNicsGet(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NICs: %w", err)
	}

	data, err := json.MarshalIndent(nics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NICs: %w", err)
	}

	return string(data), nil
}

func (s *Server) getNic(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID string) (string, error) {
	nic, _, err := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NIC: %w", err)
	}

	data, err := json.MarshalIndent(nic, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NIC: %w", err)
	}

	return string(data), nil
}

func (s *Server) createNic(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, name string, lan int32, dhcp bool, ips []string) (string, error) {
	// Validate LAN ID
	if lan < 1 {
		return "", fmt.Errorf("lan must be at least 1, got %d", lan)
	}

	// Validate IP addresses
	for _, ip := range ips {
		if err := validateIP(ip); err != nil {
			return "", fmt.Errorf("invalid IP in ips list: %w", err)
		}
	}

	properties := ionoscloud.NicProperties{
		Lan:  &lan,
		Dhcp: &dhcp,
	}
	if name != "" {
		properties.Name = &name
	}
	if len(ips) > 0 {
		properties.Ips = &ips
	}

	nic := ionoscloud.Nic{
		Properties: &properties,
	}

	result, _, err := client.NetworkInterfacesApi.DatacentersServersNicsPost(ctx, datacenterID, serverID).Nic(nic).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create NIC: %w", err)
	}

	return marshalResponse(result, "NIC")
}

func (s *Server) updateNic(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID, name string, lan int32, dhcp, dhcpSet bool, ips []string) (string, error) {
	if name == "" && lan == 0 && !dhcpSet && len(ips) == 0 {
		return "", fmt.Errorf("at least one of name, lan, dhcp, or ips must be provided")
	}

	// Validate IP addresses
	for _, ip := range ips {
		if err := validateIP(ip); err != nil {
			return "", fmt.Errorf("invalid IP in ips list: %w", err)
		}
	}

	properties := ionoscloud.NicProperties{}
	if name != "" {
		properties.Name = &name
	}
	if lan > 0 {
		properties.Lan = &lan
	}
	if dhcpSet {
		properties.Dhcp = &dhcp
	}
	if len(ips) > 0 {
		properties.Ips = &ips
	}

	result, _, err := client.NetworkInterfacesApi.DatacentersServersNicsPatch(ctx, datacenterID, serverID, nicID).Nic(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update NIC: %w", err)
	}

	return marshalResponse(result, "NIC")
}

func (s *Server) deleteNic(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID string) (string, error) {
	_, err := client.NetworkInterfacesApi.DatacentersServersNicsDelete(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete NIC: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "nic_id": nicID})
}

// =============================================================================
// Networking - IP Blocks
// =============================================================================

func (s *Server) listIpBlocks(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	ipblocks, _, err := client.IPBlocksApi.IpblocksGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list IP blocks: %w", err)
	}

	data, err := json.MarshalIndent(ipblocks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal IP blocks: %w", err)
	}

	return string(data), nil
}

func (s *Server) getIpBlock(client *ionoscloud.APIClient, ctx context.Context, ipblockID string) (string, error) {
	ipblock, _, err := client.IPBlocksApi.IpblocksFindById(ctx, ipblockID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get IP block: %w", err)
	}

	data, err := json.MarshalIndent(ipblock, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal IP block: %w", err)
	}

	return string(data), nil
}

func (s *Server) createIpBlock(client *ionoscloud.APIClient, ctx context.Context, location string, size int32, name string) (string, error) {
	// Validate location
	if err := validateLocation(location); err != nil {
		return "", err
	}

	// Validate size
	if size < 1 {
		return "", fmt.Errorf("size must be at least 1, got %d", size)
	}

	properties := ionoscloud.IpBlockProperties{
		Location: &location,
		Size:     &size,
	}
	if name != "" {
		properties.Name = &name
	}

	ipblock := ionoscloud.IpBlock{
		Properties: &properties,
	}

	result, _, err := client.IPBlocksApi.IpblocksPost(ctx).Ipblock(ipblock).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create IP block: %w", err)
	}

	return marshalResponse(result, "IP block")
}

func (s *Server) updateIpBlock(client *ionoscloud.APIClient, ctx context.Context, ipblockID, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name must be provided")
	}

	properties := ionoscloud.IpBlockProperties{
		Name: &name,
	}

	result, _, err := client.IPBlocksApi.IpblocksPatch(ctx, ipblockID).Ipblock(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update IP block: %w", err)
	}

	return marshalResponse(result, "IP block")
}

func (s *Server) deleteIpBlock(client *ionoscloud.APIClient, ctx context.Context, ipblockID string) (string, error) {
	_, err := client.IPBlocksApi.IpblocksDelete(ctx, ipblockID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete IP block: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "ipblock_id": ipblockID})
}

// =============================================================================
// Networking - Firewall Rules
// =============================================================================

func (s *Server) listFirewallRules(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID string) (string, error) {
	rules, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesGet(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list firewall rules: %w", err)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal firewall rules: %w", err)
	}

	return string(data), nil
}

func (s *Server) getFirewallRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID, firewallRuleID string) (string, error) {
	rule, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get firewall rule: %w", err)
	}

	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal firewall rule: %w", err)
	}

	return string(data), nil
}

func (s *Server) createFirewallRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID, name, protocol, sourceMac, sourceIP, targetIP string, portRangeStart, portRangeEnd, icmpType, icmpCode int32, icmpTypeSet, icmpCodeSet bool, ruleType string) (string, error) {
	// Validate protocol
	if err := validateProtocol(protocol); err != nil {
		return "", err
	}

	// Validate MAC address
	if err := validateMAC(sourceMac); err != nil {
		return "", err
	}

	// Validate IP addresses
	if err := validateIP(sourceIP); err != nil {
		return "", fmt.Errorf("invalid source_ip: %w", err)
	}
	if err := validateIP(targetIP); err != nil {
		return "", fmt.Errorf("invalid target_ip: %w", err)
	}

	// Validate port range
	portStartSet := portRangeStart > 0
	portEndSet := portRangeEnd > 0
	if err := validatePortRange(portRangeStart, portRangeEnd, portStartSet, portEndSet); err != nil {
		return "", err
	}

	// Validate ICMP parameters
	if err := validateICMP(icmpType, icmpCode, icmpTypeSet, icmpCodeSet); err != nil {
		return "", err
	}

	properties := ionoscloud.FirewallruleProperties{
		Protocol: &protocol,
	}
	if name != "" {
		properties.Name = &name
	}
	if sourceMac != "" {
		properties.SourceMac = &sourceMac
	}
	if sourceIP != "" {
		properties.SourceIp = &sourceIP
	}
	if targetIP != "" {
		properties.TargetIp = &targetIP
	}
	if portStartSet {
		properties.PortRangeStart = &portRangeStart
	}
	if portEndSet {
		properties.PortRangeEnd = &portRangeEnd
	}
	if icmpTypeSet {
		properties.IcmpType = &icmpType
	}
	if icmpCodeSet {
		properties.IcmpCode = &icmpCode
	}
	if ruleType != "" {
		properties.Type = &ruleType
	}

	rule := ionoscloud.FirewallRule{
		Properties: &properties,
	}

	result, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPost(ctx, datacenterID, serverID, nicID).Firewallrule(rule).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create firewall rule: %w", err)
	}

	return marshalResponse(result, "firewall rule")
}

func (s *Server) updateFirewallRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID, firewallRuleID, name, protocol, sourceMac, sourceIP, targetIP string, portRangeStart, portRangeEnd int32, portRangeStartSet, portRangeEndSet bool, icmpType, icmpCode int32, icmpTypeSet, icmpCodeSet bool, ruleType string) (string, error) {
	// Check if at least one field is provided
	if name == "" && protocol == "" && sourceMac == "" && sourceIP == "" && targetIP == "" && !portRangeStartSet && !portRangeEndSet && !icmpTypeSet && !icmpCodeSet && ruleType == "" {
		return "", fmt.Errorf("at least one field must be provided for update")
	}

	// Validate protocol if provided
	if protocol != "" {
		if err := validateProtocol(protocol); err != nil {
			return "", err
		}
	}

	// Validate MAC address if provided
	if err := validateMAC(sourceMac); err != nil {
		return "", err
	}

	// Validate IP addresses if provided
	if err := validateIP(sourceIP); err != nil {
		return "", fmt.Errorf("invalid source_ip: %w", err)
	}
	if err := validateIP(targetIP); err != nil {
		return "", fmt.Errorf("invalid target_ip: %w", err)
	}

	// Validate port range
	if err := validatePortRange(portRangeStart, portRangeEnd, portRangeStartSet, portRangeEndSet); err != nil {
		return "", err
	}

	// Validate ICMP parameters
	if err := validateICMP(icmpType, icmpCode, icmpTypeSet, icmpCodeSet); err != nil {
		return "", err
	}

	properties := ionoscloud.FirewallruleProperties{}
	if name != "" {
		properties.Name = &name
	}
	if protocol != "" {
		properties.Protocol = &protocol
	}
	if sourceMac != "" {
		properties.SourceMac = &sourceMac
	}
	if sourceIP != "" {
		properties.SourceIp = &sourceIP
	}
	if targetIP != "" {
		properties.TargetIp = &targetIP
	}
	if portRangeStartSet {
		properties.PortRangeStart = &portRangeStart
	}
	if portRangeEndSet {
		properties.PortRangeEnd = &portRangeEnd
	}
	if icmpTypeSet {
		properties.IcmpType = &icmpType
	}
	if icmpCodeSet {
		properties.IcmpCode = &icmpCode
	}
	if ruleType != "" {
		properties.Type = &ruleType
	}

	result, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPatch(ctx, datacenterID, serverID, nicID, firewallRuleID).Firewallrule(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update firewall rule: %w", err)
	}

	return marshalResponse(result, "firewall rule")
}

func (s *Server) deleteFirewallRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID, firewallRuleID string) (string, error) {
	_, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesDelete(ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete firewall rule: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "firewallrule_id": firewallRuleID})
}

// =============================================================================
// Networking - NAT Gateways
// =============================================================================

func (s *Server) listNatGateways(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	natgateways, _, err := client.NATGatewaysApi.DatacentersNatgatewaysGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NAT gateways: %w", err)
	}

	data, err := json.MarshalIndent(natgateways, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NAT gateways: %w", err)
	}

	return string(data), nil
}

func (s *Server) getNatGateway(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID string) (string, error) {
	natgateway, _, err := client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NAT gateway: %w", err)
	}

	data, err := json.MarshalIndent(natgateway, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NAT gateway: %w", err)
	}

	return string(data), nil
}

func (s *Server) createNatGateway(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name string, publicIPs []string, lans []map[string]interface{}) (string, error) {
	// Validate public IPs
	for i, ip := range publicIPs {
		if err := validateIP(ip); err != nil {
			return "", fmt.Errorf("public_ips[%d] invalid: %w", i, err)
		}
	}

	properties := ionoscloud.NatGatewayProperties{
		Name:      &name,
		PublicIps: &publicIPs,
	}

	// Convert and validate LAN configurations if provided
	if len(lans) > 0 {
		natGatewayLans, err := parseLanConfigurations(lans)
		if err != nil {
			return "", err
		}
		properties.Lans = &natGatewayLans
	}

	natGateway := ionoscloud.NatGateway{
		Properties: &properties,
	}

	result, _, err := client.NATGatewaysApi.DatacentersNatgatewaysPost(ctx, datacenterID).NatGateway(natGateway).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create NAT gateway: %w", err)
	}

	return marshalResponse(result, "NAT gateway")
}

func (s *Server) updateNatGateway(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID, name string, publicIPs []string, lans []map[string]interface{}) (string, error) {
	// Check if at least one field is provided
	if name == "" && len(publicIPs) == 0 && len(lans) == 0 {
		return "", fmt.Errorf("at least one field must be provided for update")
	}

	// Validate public IPs
	for i, ip := range publicIPs {
		if err := validateIP(ip); err != nil {
			return "", fmt.Errorf("public_ips[%d] invalid: %w", i, err)
		}
	}

	properties := ionoscloud.NatGatewayProperties{}
	if name != "" {
		properties.Name = &name
	}
	if len(publicIPs) > 0 {
		properties.PublicIps = &publicIPs
	}

	// Convert and validate LAN configurations if provided
	if len(lans) > 0 {
		natGatewayLans, err := parseLanConfigurations(lans)
		if err != nil {
			return "", err
		}
		properties.Lans = &natGatewayLans
	}

	result, _, err := client.NATGatewaysApi.DatacentersNatgatewaysPatch(ctx, datacenterID, natGatewayID).NatGatewayProperties(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update NAT gateway: %w", err)
	}

	return marshalResponse(result, "NAT gateway")
}

func (s *Server) deleteNatGateway(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID string) (string, error) {
	_, err := client.NATGatewaysApi.DatacentersNatgatewaysDelete(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete NAT gateway: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "nat_gateway_id": natGatewayID})
}

// =============================================================================
// NAT Gateway Rules
// =============================================================================

func (s *Server) listNatGatewayRules(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID string) (string, error) {
	rules, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesGet(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NAT gateway rules: %w", err)
	}

	return marshalResponse(rules, "NAT gateway rules")
}

func (s *Server) getNatGatewayRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID, ruleID string) (string, error) {
	rule, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesFindByNatGatewayRuleId(ctx, datacenterID, natGatewayID, ruleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NAT gateway rule: %w", err)
	}

	return marshalResponse(rule, "NAT gateway rule")
}

// Valid NAT gateway rule protocols
var validNatProtocols = map[string]bool{
	"TCP": true, "UDP": true, "ICMP": true, "ALL": true,
}

// Valid NAT gateway rule types
var validNatRuleTypes = map[string]bool{
	"SNAT": true,
}

// parseLanConfigurations parses and validates LAN configurations for NAT gateways
func parseLanConfigurations(lans []map[string]interface{}) ([]ionoscloud.NatGatewayLanProperties, error) {
	natGatewayLans := make([]ionoscloud.NatGatewayLanProperties, len(lans))
	for i, lan := range lans {
		// Validate LAN ID
		idFloat, ok := lan["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("lan[%d].id is required and must be a number", i)
		}
		lanID := int32(idFloat)
		if lanID <= 0 {
			return nil, fmt.Errorf("lan[%d].id must be positive, got %d", i, lanID)
		}

		lanProps := ionoscloud.NatGatewayLanProperties{
			Id: &lanID,
		}

		// Validate gateway IPs if provided
		if gatewayIPs, ok := lan["gateway_ips"].([]interface{}); ok {
			ips := make([]string, len(gatewayIPs))
			for j, ip := range gatewayIPs {
				ipStr, ok := ip.(string)
				if !ok {
					return nil, fmt.Errorf("lan[%d].gateway_ips[%d] must be a string", i, j)
				}
				if err := validateIP(ipStr); err != nil {
					return nil, fmt.Errorf("lan[%d].gateway_ips[%d] invalid: %w", i, j, err)
				}
				ips[j] = ipStr
			}
			lanProps.GatewayIps = &ips
		}

		natGatewayLans[i] = lanProps
	}
	return natGatewayLans, nil
}

// validateNatPortRange validates port range for NAT gateway rules
func validateNatPortRange(start, end int32) error {
	if start != 0 {
		if start < 1 || start > 65535 {
			return fmt.Errorf("target_port_range_start must be between 1-65535, got %d", start)
		}
	}
	if end != 0 {
		if end < 1 || end > 65535 {
			return fmt.Errorf("target_port_range_end must be between 1-65535, got %d", end)
		}
	}
	if start != 0 && end != 0 && start > end {
		return fmt.Errorf("target_port_range_start (%d) cannot be greater than target_port_range_end (%d)", start, end)
	}
	return nil
}

func (s *Server) createNatGatewayRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID, name, ruleType, protocol, sourceSubnet, publicIP, targetSubnet string, targetPortRangeStart, targetPortRangeEnd int32) (string, error) {
	// Validate source subnet (CIDR)
	if err := validateIP(sourceSubnet); err != nil {
		return "", fmt.Errorf("invalid source_subnet: %w", err)
	}

	// Validate public IP
	if err := validateIP(publicIP); err != nil {
		return "", fmt.Errorf("invalid public_ip: %w", err)
	}

	// Validate protocol if provided
	if protocol != "" && !validNatProtocols[protocol] {
		return "", fmt.Errorf("invalid protocol: %s (valid: TCP, UDP, ICMP, ALL)", protocol)
	}

	// Validate rule type if provided
	if ruleType != "" && !validNatRuleTypes[ruleType] {
		return "", fmt.Errorf("invalid type: %s (valid: SNAT)", ruleType)
	}

	// Validate port range
	if err := validateNatPortRange(targetPortRangeStart, targetPortRangeEnd); err != nil {
		return "", err
	}

	properties := ionoscloud.NatGatewayRuleProperties{
		Name:         &name,
		SourceSubnet: &sourceSubnet,
		PublicIp:     &publicIP,
	}

	if ruleType != "" {
		natRuleType := ionoscloud.NatGatewayRuleType(ruleType)
		properties.Type = &natRuleType
	}
	if protocol != "" {
		natProtocol := ionoscloud.NatGatewayRuleProtocol(protocol)
		properties.Protocol = &natProtocol
	}
	if targetSubnet != "" {
		if err := validateIP(targetSubnet); err != nil {
			return "", fmt.Errorf("invalid target_subnet: %w", err)
		}
		properties.TargetSubnet = &targetSubnet
	}
	if targetPortRangeStart > 0 || targetPortRangeEnd > 0 {
		portRange := ionoscloud.TargetPortRange{}
		if targetPortRangeStart > 0 {
			portRange.Start = &targetPortRangeStart
		}
		if targetPortRangeEnd > 0 {
			portRange.End = &targetPortRangeEnd
		}
		properties.TargetPortRange = &portRange
	}

	rule := ionoscloud.NatGatewayRule{
		Properties: &properties,
	}

	result, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesPost(ctx, datacenterID, natGatewayID).NatGatewayRule(rule).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create NAT gateway rule: %w", err)
	}

	return marshalResponse(result, "NAT gateway rule")
}

func (s *Server) updateNatGatewayRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID, ruleID, name, protocol, sourceSubnet, publicIP, targetSubnet string, targetPortRangeStart, targetPortRangeEnd int32, targetPortRangeStartSet, targetPortRangeEndSet bool) (string, error) {
	// Check if at least one field is provided
	if name == "" && protocol == "" && sourceSubnet == "" && publicIP == "" && targetSubnet == "" && !targetPortRangeStartSet && !targetPortRangeEndSet {
		return "", fmt.Errorf("at least one field must be provided for update")
	}

	// Validate inputs
	if sourceSubnet != "" {
		if err := validateIP(sourceSubnet); err != nil {
			return "", fmt.Errorf("invalid source_subnet: %w", err)
		}
	}
	if publicIP != "" {
		if err := validateIP(publicIP); err != nil {
			return "", fmt.Errorf("invalid public_ip: %w", err)
		}
	}
	if targetSubnet != "" {
		if err := validateIP(targetSubnet); err != nil {
			return "", fmt.Errorf("invalid target_subnet: %w", err)
		}
	}
	if protocol != "" && !validNatProtocols[protocol] {
		return "", fmt.Errorf("invalid protocol: %s (valid: TCP, UDP, ICMP, ALL)", protocol)
	}

	// Validate port range if provided
	if targetPortRangeStartSet || targetPortRangeEndSet {
		var start, end int32
		if targetPortRangeStartSet {
			start = targetPortRangeStart
		}
		if targetPortRangeEndSet {
			end = targetPortRangeEnd
		}
		if err := validateNatPortRange(start, end); err != nil {
			return "", err
		}
	}

	properties := ionoscloud.NatGatewayRuleProperties{}
	if name != "" {
		properties.Name = &name
	}
	if protocol != "" {
		natProtocol := ionoscloud.NatGatewayRuleProtocol(protocol)
		properties.Protocol = &natProtocol
	}
	if sourceSubnet != "" {
		properties.SourceSubnet = &sourceSubnet
	}
	if publicIP != "" {
		properties.PublicIp = &publicIP
	}
	if targetSubnet != "" {
		properties.TargetSubnet = &targetSubnet
	}
	if targetPortRangeStartSet || targetPortRangeEndSet {
		portRange := ionoscloud.TargetPortRange{}
		if targetPortRangeStartSet {
			portRange.Start = &targetPortRangeStart
		}
		if targetPortRangeEndSet {
			portRange.End = &targetPortRangeEnd
		}
		properties.TargetPortRange = &portRange
	}

	result, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesPatch(ctx, datacenterID, natGatewayID, ruleID).NatGatewayRuleProperties(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update NAT gateway rule: %w", err)
	}

	return marshalResponse(result, "NAT gateway rule")
}

func (s *Server) deleteNatGatewayRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID, ruleID string) (string, error) {
	_, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesDelete(ctx, datacenterID, natGatewayID, ruleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete NAT gateway rule: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "rule_id": ruleID})
}

// =============================================================================
// Networking - Private Cross Connects
// =============================================================================

func (s *Server) listPccs(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	pccs, _, err := client.PrivateCrossConnectsApi.PccsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Private Cross Connects: %w", err)
	}

	data, err := json.MarshalIndent(pccs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Private Cross Connects: %w", err)
	}

	return string(data), nil
}

func (s *Server) getPcc(client *ionoscloud.APIClient, ctx context.Context, pccID string) (string, error) {
	pcc, _, err := client.PrivateCrossConnectsApi.PccsFindById(ctx, pccID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Private Cross Connect: %w", err)
	}

	data, err := json.MarshalIndent(pcc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Private Cross Connect: %w", err)
	}

	return string(data), nil
}

func (s *Server) createPcc(client *ionoscloud.APIClient, ctx context.Context, name, description string) (string, error) {
	properties := ionoscloud.PrivateCrossConnectProperties{
		Name: &name,
	}
	if description != "" {
		properties.Description = &description
	}

	pcc := ionoscloud.PrivateCrossConnect{
		Properties: &properties,
	}

	result, _, err := client.PrivateCrossConnectsApi.PccsPost(ctx).Pcc(pcc).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create Private Cross Connect: %w", err)
	}

	return marshalResponse(result, "Private Cross Connect")
}

func (s *Server) updatePcc(client *ionoscloud.APIClient, ctx context.Context, pccID, name, description string) (string, error) {
	// Check if at least one field is provided
	if name == "" && description == "" {
		return "", fmt.Errorf("at least one field must be provided for update")
	}

	properties := ionoscloud.PrivateCrossConnectProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := client.PrivateCrossConnectsApi.PccsPatch(ctx, pccID).Pcc(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update Private Cross Connect: %w", err)
	}

	return marshalResponse(result, "Private Cross Connect")
}

func (s *Server) deletePcc(client *ionoscloud.APIClient, ctx context.Context, pccID string) (string, error) {
	_, err := client.PrivateCrossConnectsApi.PccsDelete(ctx, pccID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete Private Cross Connect: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "pcc_id": pccID})
}

// =============================================================================
// Load Balancers - Application Load Balancers
// =============================================================================

func (s *Server) listApplicationLoadBalancers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	albs, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Application Load Balancers: %w", err)
	}

	data, err := json.MarshalIndent(albs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Application Load Balancers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getApplicationLoadBalancer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, albID string) (string, error) {
	alb, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, datacenterID, albID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Application Load Balancer: %w", err)
	}

	data, err := json.MarshalIndent(alb, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Application Load Balancer: %w", err)
	}

	return string(data), nil
}

func (s *Server) listAlbForwardingRules(client *ionoscloud.APIClient, ctx context.Context, datacenterID, albID string) (string, error) {
	rules, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesGet(ctx, datacenterID, albID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list ALB forwarding rules: %w", err)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ALB forwarding rules: %w", err)
	}

	return string(data), nil
}

func (s *Server) getAlbForwardingRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, albID, ruleID string) (string, error) {
	rule, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId(ctx, datacenterID, albID, ruleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get ALB forwarding rule: %w", err)
	}

	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ALB forwarding rule: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Load Balancers - Network Load Balancers
// =============================================================================

func (s *Server) listNetworkLoadBalancers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	nlbs, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Network Load Balancers: %w", err)
	}

	data, err := json.MarshalIndent(nlbs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Network Load Balancers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getNetworkLoadBalancer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, nlbID string) (string, error) {
	nlb, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, datacenterID, nlbID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Network Load Balancer: %w", err)
	}

	data, err := json.MarshalIndent(nlb, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Network Load Balancer: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Load Balancers - Target Groups
// =============================================================================

func (s *Server) listTargetGroups(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	targetGroups, _, err := client.TargetGroupsApi.TargetgroupsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list target groups: %w", err)
	}

	data, err := json.MarshalIndent(targetGroups, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal target groups: %w", err)
	}

	return string(data), nil
}

func (s *Server) getTargetGroup(client *ionoscloud.APIClient, ctx context.Context, targetGroupID string) (string, error) {
	targetGroup, _, err := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, targetGroupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get target group: %w", err)
	}

	data, err := json.MarshalIndent(targetGroup, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal target group: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Kubernetes
// =============================================================================

func (s *Server) listK8sClusters(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	clusters, _, err := client.KubernetesApi.K8sGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Kubernetes clusters: %w", err)
	}

	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Kubernetes clusters: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sCluster(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID string) (string, error) {
	cluster, _, err := client.KubernetesApi.K8sFindByClusterId(ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Kubernetes cluster: %w", err)
	}

	data, err := json.MarshalIndent(cluster, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Kubernetes cluster: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sKubeconfig(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID string) (string, error) {
	kubeconfig, _, err := client.KubernetesApi.K8sKubeconfigGet(ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	data, err := json.MarshalIndent(kubeconfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal kubeconfig: %w", err)
	}

	return string(data), nil
}

func (s *Server) listK8sNodepools(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID string) (string, error) {
	nodepools, _, err := client.KubernetesApi.K8sNodepoolsGet(ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list node pools: %w", err)
	}

	data, err := json.MarshalIndent(nodepools, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal node pools: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sNodepool(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID, nodepoolID string) (string, error) {
	nodepool, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get node pool: %w", err)
	}

	data, err := json.MarshalIndent(nodepool, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal node pool: %w", err)
	}

	return string(data), nil
}

func (s *Server) listK8sNodes(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID, nodepoolID string) (string, error) {
	nodes, _, err := client.KubernetesApi.K8sNodepoolsNodesGet(ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal nodes: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sNode(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID, nodepoolID, nodeID string) (string, error) {
	node, _, err := client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, k8sClusterID, nodepoolID, nodeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get node: %w", err)
	}

	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal node: %w", err)
	}

	return string(data), nil
}

func (s *Server) listK8sVersions(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	versions, _, err := client.KubernetesApi.K8sVersionsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Kubernetes versions: %w", err)
	}

	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Kubernetes versions: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - Users
// =============================================================================

func (s *Server) listUsers(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	users, _, err := client.UserManagementApi.UmUsersGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list users: %w", err)
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal users: %w", err)
	}

	return string(data), nil
}

func (s *Server) getUser(client *ionoscloud.APIClient, ctx context.Context, userID string) (string, error) {
	user, _, err := client.UserManagementApi.UmUsersFindById(ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - Groups
// =============================================================================

func (s *Server) listGroups(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	groups, _, err := client.UserManagementApi.UmGroupsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list groups: %w", err)
	}

	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal groups: %w", err)
	}

	return string(data), nil
}

func (s *Server) getGroup(client *ionoscloud.APIClient, ctx context.Context, groupID string) (string, error) {
	group, _, err := client.UserManagementApi.UmGroupsFindById(ctx, groupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get group: %w", err)
	}

	data, err := json.MarshalIndent(group, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal group: %w", err)
	}

	return string(data), nil
}

func (s *Server) listGroupMembers(client *ionoscloud.APIClient, ctx context.Context, groupID string) (string, error) {
	members, _, err := client.UserManagementApi.UmGroupsUsersGet(ctx, groupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list group members: %w", err)
	}

	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal group members: %w", err)
	}

	return string(data), nil
}

func (s *Server) listUserGroups(client *ionoscloud.APIClient, ctx context.Context, userID string) (string, error) {
	groups, _, err := client.UserManagementApi.UmUsersGroupsGet(ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list user groups: %w", err)
	}

	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal user groups: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - S3 Keys
// =============================================================================

func (s *Server) listS3Keys(client *ionoscloud.APIClient, ctx context.Context, userID string) (string, error) {
	keys, _, err := client.UserS3KeysApi.UmUsersS3keysGet(ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list S3 keys: %w", err)
	}

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal S3 keys: %w", err)
	}

	return string(data), nil
}

func (s *Server) getS3Key(client *ionoscloud.APIClient, ctx context.Context, userID, keyID string) (string, error) {
	key, _, err := client.UserS3KeysApi.UmUsersS3keysFindByKeyId(ctx, userID, keyID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get S3 key: %w", err)
	}

	data, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal S3 key: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - Contract
// =============================================================================

func (s *Server) getContract(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	contract, _, err := client.ContractResourcesApi.ContractsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get contract: %w", err)
	}

	data, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal contract: %w", err)
	}

	return string(data), nil
}

func (s *Server) listResources(client *ionoscloud.APIClient, ctx context.Context, resourceType string) (string, error) {
	var resources ionoscloud.Resources
	var err error

	if resourceType != "" {
		resources, _, err = client.UserManagementApi.UmResourcesFindByType(ctx, resourceType).Execute()
	} else {
		resources, _, err = client.UserManagementApi.UmResourcesGet(ctx).Execute()
	}

	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	data, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal resources: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// DNS
// =============================================================================

func (s *Server) listDnsZones(client *dns.APIClient, ctx context.Context) (string, error) {
	zones, _, err := client.ZonesApi.ZonesGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list DNS zones: %w", err)
	}

	data, err := json.MarshalIndent(zones, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS zones: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDnsZone(client *dns.APIClient, ctx context.Context, zoneID string) (string, error) {
	zone, _, err := client.ZonesApi.ZonesFindById(ctx, zoneID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS zone: %w", err)
	}

	data, err := json.MarshalIndent(zone, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS zone: %w", err)
	}

	return string(data), nil
}

func (s *Server) listDnsRecords(client *dns.APIClient, ctx context.Context, zoneID string) (string, error) {
	records, _, err := client.RecordsApi.ZonesRecordsGet(ctx, zoneID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list DNS records: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS records: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDnsRecord(client *dns.APIClient, ctx context.Context, zoneID, recordID string) (string, error) {
	record, _, err := client.RecordsApi.ZonesRecordsFindById(ctx, zoneID, recordID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS record: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS record: %w", err)
	}

	return string(data), nil
}
