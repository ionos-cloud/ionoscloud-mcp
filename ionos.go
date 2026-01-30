package main

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"

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

// validatePortRange validates port range values. fieldPrefix is used in error messages
// (e.g., "port_range" or "target_port_range")
func validatePortRange(start, end int32, fieldPrefix string) error {
	if start != 0 {
		if start < 1 || start > 65535 {
			return fmt.Errorf("%s_start must be between 1-65535, got %d", fieldPrefix, start)
		}
	}
	if end != 0 {
		if end < 1 || end > 65535 {
			return fmt.Errorf("%s_end must be between 1-65535, got %d", fieldPrefix, end)
		}
	}
	if start != 0 && end != 0 && start > end {
		return fmt.Errorf("%s_start (%d) cannot be greater than %s_end (%d)", fieldPrefix, start, fieldPrefix, end)
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
		return s.listDatacenters()
	case "get_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.getDatacenter(datacenterID)
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
		return s.createDatacenter(name, location, description)
	case "update_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updateDatacenter(datacenterID, name, description)
	case "delete_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.deleteDatacenter(datacenterID)
	case "list_servers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listServers(datacenterID)
	case "get_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.getServer(datacenterID, serverID)
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
		return s.createServer(datacenterID, name, int32(cores), int32(ram), cpuFamily, availabilityZone)
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
		return s.updateServer(datacenterID, serverID, name, cores, ram)
	case "delete_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.deleteServer(datacenterID, serverID)
	case "start_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.startServer(datacenterID, serverID)
	case "stop_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.stopServer(datacenterID, serverID)
	case "reboot_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.rebootServer(datacenterID, serverID)
	case "list_volumes":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listVolumes(datacenterID)
	case "get_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.getVolume(datacenterID, volumeID)
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
		return s.createVolume(datacenterID, name, float32(size), volumeType, bus, availabilityZone, image, imagePassword, licenceType)
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
		return s.updateVolume(datacenterID, volumeID, name, size)
	case "delete_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.deleteVolume(datacenterID, volumeID)
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
		return s.attachVolume(datacenterID, serverID, volumeID)
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
		return s.detachVolume(datacenterID, serverID, volumeID)
	case "list_images":
		return s.listImages()
	case "list_locations":
		return s.listLocations()
	case "list_snapshots":
		return s.listSnapshots()
	case "get_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.getSnapshot(snapshotID)
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
		return s.createSnapshot(datacenterID, volumeID, name, description)
	case "update_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updateSnapshot(snapshotID, name, description)
	case "delete_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.deleteSnapshot(snapshotID)
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
		return s.restoreSnapshot(datacenterID, volumeID, snapshotID)
	// Networking - LANs
	case "list_lans":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listLans(datacenterID)
	case "get_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		lanID, ok := arguments["lan_id"].(string)
		if !ok {
			return "", fmt.Errorf("lan_id is required")
		}
		return s.getLan(datacenterID, lanID)
	case "create_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, _ := arguments["name"].(string)
		public, _ := arguments["public"].(bool)
		return s.createLan(datacenterID, name, public)
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
		return s.updateLan(datacenterID, lanID, name, public, publicSet)
	case "delete_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		lanID, ok := arguments["lan_id"].(string)
		if !ok {
			return "", fmt.Errorf("lan_id is required")
		}
		return s.deleteLan(datacenterID, lanID)
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
		return s.listNics(datacenterID, serverID)
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
		return s.getNic(datacenterID, serverID, nicID)
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
		return s.createNic(datacenterID, serverID, name, int32(lan), dhcp, ips)
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
		return s.updateNic(datacenterID, serverID, nicID, name, lan, dhcp, dhcpSet, ips)
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
		return s.deleteNic(datacenterID, serverID, nicID)
	// Networking - IP Blocks
	case "list_ipblocks":
		return s.listIpBlocks()
	case "get_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		return s.getIpBlock(ipblockID)
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
		return s.createIpBlock(location, int32(size), name)
	case "update_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		name, _ := arguments["name"].(string)
		return s.updateIpBlock(ipblockID, name)
	case "delete_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		return s.deleteIpBlock(ipblockID)
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
		return s.listFirewallRules(datacenterID, serverID, nicID)
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
		return s.getFirewallRule(datacenterID, serverID, nicID, firewallRuleID)
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
		return s.createFirewallRule(datacenterID, serverID, nicID, name, protocol, sourceMac, sourceIP, targetIP, portRangeStart, portRangeEnd, icmpType, icmpCode, icmpTypeSet, icmpCodeSet, ruleType)
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
		return s.updateFirewallRule(datacenterID, serverID, nicID, firewallRuleID, name, protocol, sourceMac, sourceIP, targetIP, portRangeStart, portRangeEnd, portRangeStartSet, portRangeEndSet, icmpType, icmpCode, icmpTypeSet, icmpCodeSet, ruleType)
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
		return s.deleteFirewallRule(datacenterID, serverID, nicID, firewallRuleID)
	// Networking - NAT Gateways
	case "list_nat_gateways":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listNatGateways(datacenterID)
	case "get_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		return s.getNatGateway(datacenterID, natGatewayID)
	// Networking - Private Cross Connects
	case "list_pccs":
		return s.listPccs()
	case "get_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		return s.getPcc(pccID)
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
		return s.createNatGateway(datacenterID, name, publicIPs, lans)
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
		return s.updateNatGateway(datacenterID, natGatewayID, name, publicIPs, lans)
	case "delete_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		return s.deleteNatGateway(datacenterID, natGatewayID)
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
		return s.listNatGatewayRules(datacenterID, natGatewayID)
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
		return s.getNatGatewayRule(datacenterID, natGatewayID, ruleID)
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
		return s.createNatGatewayRule(datacenterID, natGatewayID, name, ruleType, protocol, sourceSubnet, publicIP, targetSubnet, targetPortRangeStart, targetPortRangeEnd)
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
		return s.updateNatGatewayRule(datacenterID, natGatewayID, ruleID, name, protocol, sourceSubnet, publicIP, targetSubnet, targetPortRangeStart, targetPortRangeEnd, targetPortRangeStartSet, targetPortRangeEndSet)
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
		return s.deleteNatGatewayRule(datacenterID, natGatewayID, ruleID)
	// Private Cross Connect CRUD
	case "create_pcc":
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		description, _ := arguments["description"].(string)
		return s.createPcc(name, description)
	case "update_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updatePcc(pccID, name, description)
	case "delete_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		return s.deletePcc(pccID)
	// Load Balancers - Application Load Balancers
	case "list_application_load_balancers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listApplicationLoadBalancers(datacenterID)
	case "get_application_load_balancer":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		return s.getApplicationLoadBalancer(datacenterID, albID)
	case "list_alb_forwarding_rules":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		return s.listAlbForwardingRules(datacenterID, albID)
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
		return s.getAlbForwardingRule(datacenterID, albID, ruleID)
	// Load Balancers - Network Load Balancers
	case "list_network_load_balancers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listNetworkLoadBalancers(datacenterID)
	case "get_network_load_balancer":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		nlbID, ok := arguments["nlb_id"].(string)
		if !ok {
			return "", fmt.Errorf("nlb_id is required")
		}
		return s.getNetworkLoadBalancer(datacenterID, nlbID)
	// Load Balancers - Target Groups
	case "list_target_groups":
		return s.listTargetGroups()
	case "get_target_group":
		targetGroupID, ok := arguments["target_group_id"].(string)
		if !ok {
			return "", fmt.Errorf("target_group_id is required")
		}
		return s.getTargetGroup(targetGroupID)
	// Kubernetes
	case "list_k8s_clusters":
		return s.listK8sClusters()
	case "get_k8s_cluster":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.getK8sCluster(k8sClusterID)
	case "get_k8s_kubeconfig":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.getK8sKubeconfig(k8sClusterID)
	case "list_k8s_nodepools":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.listK8sNodepools(k8sClusterID)
	case "get_k8s_nodepool":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		return s.getK8sNodepool(k8sClusterID, nodepoolID)
	case "list_k8s_nodes":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		return s.listK8sNodes(k8sClusterID, nodepoolID)
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
		return s.getK8sNode(k8sClusterID, nodepoolID, nodeID)
	case "list_k8s_versions":
		return s.listK8sVersions()
	// User Management - Users
	case "list_users":
		return s.listUsers()
	case "get_user":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.getUser(userID)
	// User Management - Groups
	case "list_groups":
		return s.listGroups()
	case "get_group":
		groupID, ok := arguments["group_id"].(string)
		if !ok {
			return "", fmt.Errorf("group_id is required")
		}
		return s.getGroup(groupID)
	case "list_group_members":
		groupID, ok := arguments["group_id"].(string)
		if !ok {
			return "", fmt.Errorf("group_id is required")
		}
		return s.listGroupMembers(groupID)
	case "list_user_groups":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.listUserGroups(userID)
	// User Management - S3 Keys
	case "list_s3_keys":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.listS3Keys(userID)
	case "get_s3_key":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		keyID, ok := arguments["key_id"].(string)
		if !ok {
			return "", fmt.Errorf("key_id is required")
		}
		return s.getS3Key(userID, keyID)
	// User Management - Contract
	case "get_contract":
		return s.getContract()
	case "list_resources":
		resourceType, _ := arguments["resource_type"].(string)
		return s.listResources(resourceType)
	// DNS
	case "list_dns_zones":
		return s.listDnsZones()
	case "get_dns_zone":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		return s.getDnsZone(zoneID)
	case "list_dns_records":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		return s.listDnsRecords(zoneID)
	case "get_dns_record":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		recordID, ok := arguments["record_id"].(string)
		if !ok {
			return "", fmt.Errorf("record_id is required")
		}
		return s.getDnsRecord(zoneID, recordID)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) listDatacenters() (string, error) {
	datacenters, _, err := s.client.DataCentersApi.DatacentersGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list datacenters: %w", err)
	}

	return marshalResponse(datacenters, "datacenters")
}

func (s *Server) getDatacenter(datacenterID string) (string, error) {
	datacenter, _, err := s.client.DataCentersApi.DatacentersFindById(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get datacenter: %w", err)
	}

	return marshalResponse(datacenter, "datacenter")
}

func (s *Server) createDatacenter(name, location, description string) (string, error) {
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

	result, _, err := s.client.DataCentersApi.DatacentersPost(s.ctx).Datacenter(datacenter).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create datacenter: %w", err)
	}

	return marshalResponse(result, "datacenter")
}

func (s *Server) updateDatacenter(datacenterID, name, description string) (string, error) {
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

	result, _, err := s.client.DataCentersApi.DatacentersPatch(s.ctx, datacenterID).Datacenter(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update datacenter: %w", err)
	}

	return marshalResponse(result, "datacenter")
}

func (s *Server) deleteDatacenter(datacenterID string) (string, error) {
	_, err := s.client.DataCentersApi.DatacentersDelete(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete datacenter: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "datacenter_id": datacenterID})
}

func (s *Server) listServers(datacenterID string) (string, error) {
	servers, _, err := s.client.ServersApi.DatacentersServersGet(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list servers: %w", err)
	}

	return marshalResponse(servers, "servers")
}

func (s *Server) getServer(datacenterID, serverID string) (string, error) {
	server, _, err := s.client.ServersApi.DatacentersServersFindById(s.ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get server: %w", err)
	}

	return marshalResponse(server, "server")
}

func (s *Server) createServer(datacenterID, name string, cores, ram int32, cpuFamily, availabilityZone string) (string, error) {
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

	result, _, err := s.client.ServersApi.DatacentersServersPost(s.ctx, datacenterID).Server(server).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create server: %w", err)
	}

	return marshalResponse(result, "server")
}

func (s *Server) updateServer(datacenterID, serverID, name string, cores, ram int32) (string, error) {
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

	result, _, err := s.client.ServersApi.DatacentersServersPatch(s.ctx, datacenterID, serverID).Server(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update server: %w", err)
	}

	return marshalResponse(result, "server")
}

func (s *Server) deleteServer(datacenterID, serverID string) (string, error) {
	_, err := s.client.ServersApi.DatacentersServersDelete(s.ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete server: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "server_id": serverID})
}

func (s *Server) startServer(datacenterID, serverID string) (string, error) {
	_, err := s.client.ServersApi.DatacentersServersStartPost(s.ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to start server: %w", err)
	}

	return statusResponse(map[string]string{"status": "starting", "server_id": serverID})
}

func (s *Server) stopServer(datacenterID, serverID string) (string, error) {
	_, err := s.client.ServersApi.DatacentersServersStopPost(s.ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to stop server: %w", err)
	}

	return statusResponse(map[string]string{"status": "stopping", "server_id": serverID})
}

func (s *Server) rebootServer(datacenterID, serverID string) (string, error) {
	_, err := s.client.ServersApi.DatacentersServersRebootPost(s.ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to reboot server: %w", err)
	}

	return statusResponse(map[string]string{"status": "rebooting", "server_id": serverID})
}

func (s *Server) listVolumes(datacenterID string) (string, error) {
	volumes, _, err := s.client.VolumesApi.DatacentersVolumesGet(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list volumes: %w", err)
	}

	return marshalResponse(volumes, "volumes")
}

func (s *Server) getVolume(datacenterID, volumeID string) (string, error) {
	volume, _, err := s.client.VolumesApi.DatacentersVolumesFindById(s.ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get volume: %w", err)
	}

	return marshalResponse(volume, "volume")
}

func (s *Server) createVolume(datacenterID, name string, size float32, volumeType, bus, availabilityZone, image, imagePassword, licenceType string) (string, error) {
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

	result, _, err := s.client.VolumesApi.DatacentersVolumesPost(s.ctx, datacenterID).Volume(volume).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	return marshalResponse(result, "volume")
}

func (s *Server) updateVolume(datacenterID, volumeID, name string, size float32) (string, error) {
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

	result, _, err := s.client.VolumesApi.DatacentersVolumesPatch(s.ctx, datacenterID, volumeID).Volume(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update volume: %w", err)
	}

	return marshalResponse(result, "volume")
}

func (s *Server) deleteVolume(datacenterID, volumeID string) (string, error) {
	_, err := s.client.VolumesApi.DatacentersVolumesDelete(s.ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete volume: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "volume_id": volumeID})
}

func (s *Server) attachVolume(datacenterID, serverID, volumeID string) (string, error) {
	volume := ionoscloud.Volume{
		Id: &volumeID,
	}

	result, _, err := s.client.ServersApi.DatacentersServersVolumesPost(s.ctx, datacenterID, serverID).Volume(volume).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to attach volume: %w", err)
	}

	return marshalResponse(result, "volume")
}

func (s *Server) detachVolume(datacenterID, serverID, volumeID string) (string, error) {
	_, err := s.client.ServersApi.DatacentersServersVolumesDelete(s.ctx, datacenterID, serverID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to detach volume: %w", err)
	}

	return statusResponse(map[string]string{"status": "detached", "volume_id": volumeID, "server_id": serverID})
}

func (s *Server) listImages() (string, error) {
	images, _, err := s.client.ImagesApi.ImagesGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}

	return marshalResponse(images, "images")
}

func (s *Server) listLocations() (string, error) {
	locations, _, err := s.client.LocationsApi.LocationsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list locations: %w", err)
	}

	return marshalResponse(locations, "locations")
}

func (s *Server) listSnapshots() (string, error) {
	snapshots, _, err := s.client.SnapshotsApi.SnapshotsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}

	return marshalResponse(snapshots, "snapshots")
}

func (s *Server) getSnapshot(snapshotID string) (string, error) {
	snapshot, _, err := s.client.SnapshotsApi.SnapshotsFindById(s.ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot: %w", err)
	}

	return marshalResponse(snapshot, "snapshot")
}

func (s *Server) createSnapshot(datacenterID, volumeID, name, description string) (string, error) {
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

	result, _, err := s.client.VolumesApi.DatacentersVolumesCreateSnapshotPost(s.ctx, datacenterID, volumeID).Snapshot(snapshot).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	return marshalResponse(result, "snapshot")
}

func (s *Server) updateSnapshot(snapshotID, name, description string) (string, error) {
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

	result, _, err := s.client.SnapshotsApi.SnapshotsPatch(s.ctx, snapshotID).Snapshot(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update snapshot: %w", err)
	}

	return marshalResponse(result, "snapshot")
}

func (s *Server) deleteSnapshot(snapshotID string) (string, error) {
	_, err := s.client.SnapshotsApi.SnapshotsDelete(s.ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "snapshot_id": snapshotID})
}

func (s *Server) restoreSnapshot(datacenterID, volumeID, snapshotID string) (string, error) {
	properties := ionoscloud.RestoreSnapshotProperties{
		SnapshotId: &snapshotID,
	}
	restoreSnapshot := ionoscloud.RestoreSnapshot{
		Properties: &properties,
	}

	_, err := s.client.VolumesApi.DatacentersVolumesRestoreSnapshotPost(s.ctx, datacenterID, volumeID).RestoreSnapshot(restoreSnapshot).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to restore snapshot: %w", err)
	}

	return statusResponse(map[string]string{"status": "restoring", "volume_id": volumeID, "snapshot_id": snapshotID})
}

// =============================================================================
// Networking - LANs
// =============================================================================

func (s *Server) listLans(datacenterID string) (string, error) {
	lans, _, err := s.client.LANsApi.DatacentersLansGet(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list LANs: %w", err)
	}

	return marshalResponse(lans, "LANs")
}

func (s *Server) getLan(datacenterID, lanID string) (string, error) {
	lan, _, err := s.client.LANsApi.DatacentersLansFindById(s.ctx, datacenterID, lanID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN: %w", err)
	}

	return marshalResponse(lan, "LAN")
}

func (s *Server) createLan(datacenterID, name string, public bool) (string, error) {
	properties := ionoscloud.LanProperties{
		Public: &public,
	}
	if name != "" {
		properties.Name = &name
	}

	lan := ionoscloud.Lan{
		Properties: &properties,
	}

	result, _, err := s.client.LANsApi.DatacentersLansPost(s.ctx, datacenterID).Lan(lan).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create LAN: %w", err)
	}

	return marshalResponse(result, "LAN")
}

func (s *Server) updateLan(datacenterID, lanID, name string, public, publicSet bool) (string, error) {
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

	result, _, err := s.client.LANsApi.DatacentersLansPatch(s.ctx, datacenterID, lanID).Lan(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update LAN: %w", err)
	}

	return marshalResponse(result, "LAN")
}

func (s *Server) deleteLan(datacenterID, lanID string) (string, error) {
	_, err := s.client.LANsApi.DatacentersLansDelete(s.ctx, datacenterID, lanID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete LAN: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "lan_id": lanID})
}

// =============================================================================
// Networking - NICs
// =============================================================================

func (s *Server) listNics(datacenterID, serverID string) (string, error) {
	nics, _, err := s.client.NetworkInterfacesApi.DatacentersServersNicsGet(s.ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NICs: %w", err)
	}

	return marshalResponse(nics, "NICs")
}

func (s *Server) getNic(datacenterID, serverID, nicID string) (string, error) {
	nic, _, err := s.client.NetworkInterfacesApi.DatacentersServersNicsFindById(s.ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NIC: %w", err)
	}

	return marshalResponse(nic, "NIC")
}

func (s *Server) createNic(datacenterID, serverID, name string, lan int32, dhcp bool, ips []string) (string, error) {
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

	result, _, err := s.client.NetworkInterfacesApi.DatacentersServersNicsPost(s.ctx, datacenterID, serverID).Nic(nic).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create NIC: %w", err)
	}

	return marshalResponse(result, "NIC")
}

func (s *Server) updateNic(datacenterID, serverID, nicID, name string, lan int32, dhcp, dhcpSet bool, ips []string) (string, error) {
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

	result, _, err := s.client.NetworkInterfacesApi.DatacentersServersNicsPatch(s.ctx, datacenterID, serverID, nicID).Nic(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update NIC: %w", err)
	}

	return marshalResponse(result, "NIC")
}

func (s *Server) deleteNic(datacenterID, serverID, nicID string) (string, error) {
	_, err := s.client.NetworkInterfacesApi.DatacentersServersNicsDelete(s.ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete NIC: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "nic_id": nicID})
}

// =============================================================================
// Networking - IP Blocks
// =============================================================================

func (s *Server) listIpBlocks() (string, error) {
	ipblocks, _, err := s.client.IPBlocksApi.IpblocksGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list IP blocks: %w", err)
	}

	return marshalResponse(ipblocks, "IP blocks")
}

func (s *Server) getIpBlock(ipblockID string) (string, error) {
	ipblock, _, err := s.client.IPBlocksApi.IpblocksFindById(s.ctx, ipblockID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get IP block: %w", err)
	}

	return marshalResponse(ipblock, "IP block")
}

func (s *Server) createIpBlock(location string, size int32, name string) (string, error) {
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

	result, _, err := s.client.IPBlocksApi.IpblocksPost(s.ctx).Ipblock(ipblock).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create IP block: %w", err)
	}

	return marshalResponse(result, "IP block")
}

func (s *Server) updateIpBlock(ipblockID, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name must be provided")
	}

	properties := ionoscloud.IpBlockProperties{
		Name: &name,
	}

	result, _, err := s.client.IPBlocksApi.IpblocksPatch(s.ctx, ipblockID).Ipblock(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update IP block: %w", err)
	}

	return marshalResponse(result, "IP block")
}

func (s *Server) deleteIpBlock(ipblockID string) (string, error) {
	_, err := s.client.IPBlocksApi.IpblocksDelete(s.ctx, ipblockID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete IP block: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "ipblock_id": ipblockID})
}

// =============================================================================
// Networking - Firewall Rules
// =============================================================================

func (s *Server) listFirewallRules(datacenterID, serverID, nicID string) (string, error) {
	rules, _, err := s.client.FirewallRulesApi.DatacentersServersNicsFirewallrulesGet(s.ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list firewall rules: %w", err)
	}

	return marshalResponse(rules, "firewall rules")
}

func (s *Server) getFirewallRule(datacenterID, serverID, nicID, firewallRuleID string) (string, error) {
	rule, _, err := s.client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(s.ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get firewall rule: %w", err)
	}

	return marshalResponse(rule, "firewall rule")
}

func (s *Server) createFirewallRule(datacenterID, serverID, nicID, name, protocol, sourceMac, sourceIP, targetIP string, portRangeStart, portRangeEnd, icmpType, icmpCode int32, icmpTypeSet, icmpCodeSet bool, ruleType string) (string, error) {
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
	if err := validatePortRange(portRangeStart, portRangeEnd, "port_range"); err != nil {
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
	if portRangeStart != 0 {
		properties.PortRangeStart = &portRangeStart
	}
	if portRangeEnd != 0 {
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

	result, _, err := s.client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPost(s.ctx, datacenterID, serverID, nicID).Firewallrule(rule).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create firewall rule: %w", err)
	}

	return marshalResponse(result, "firewall rule")
}

func (s *Server) updateFirewallRule(datacenterID, serverID, nicID, firewallRuleID, name, protocol, sourceMac, sourceIP, targetIP string, portRangeStart, portRangeEnd int32, portRangeStartSet, portRangeEndSet bool, icmpType, icmpCode int32, icmpTypeSet, icmpCodeSet bool, ruleType string) (string, error) {
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
	if err := validatePortRange(portRangeStart, portRangeEnd, "port_range"); err != nil {
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

	result, _, err := s.client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPatch(s.ctx, datacenterID, serverID, nicID, firewallRuleID).Firewallrule(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update firewall rule: %w", err)
	}

	return marshalResponse(result, "firewall rule")
}

func (s *Server) deleteFirewallRule(datacenterID, serverID, nicID, firewallRuleID string) (string, error) {
	_, err := s.client.FirewallRulesApi.DatacentersServersNicsFirewallrulesDelete(s.ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete firewall rule: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "firewallrule_id": firewallRuleID})
}

// =============================================================================
// Networking - NAT Gateways
// =============================================================================

func (s *Server) listNatGateways(datacenterID string) (string, error) {
	natgateways, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysGet(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NAT gateways: %w", err)
	}

	return marshalResponse(natgateways, "NAT gateways")
}

func (s *Server) getNatGateway(datacenterID, natGatewayID string) (string, error) {
	natgateway, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(s.ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NAT gateway: %w", err)
	}

	return marshalResponse(natgateway, "NAT gateway")
}

func (s *Server) createNatGateway(datacenterID, name string, publicIPs []string, lans []map[string]interface{}) (string, error) {
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

	result, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysPost(s.ctx, datacenterID).NatGateway(natGateway).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create NAT gateway: %w", err)
	}

	return marshalResponse(result, "NAT gateway")
}

func (s *Server) updateNatGateway(datacenterID, natGatewayID, name string, publicIPs []string, lans []map[string]interface{}) (string, error) {
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

	result, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysPatch(s.ctx, datacenterID, natGatewayID).NatGatewayProperties(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update NAT gateway: %w", err)
	}

	return marshalResponse(result, "NAT gateway")
}

func (s *Server) deleteNatGateway(datacenterID, natGatewayID string) (string, error) {
	_, err := s.client.NATGatewaysApi.DatacentersNatgatewaysDelete(s.ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete NAT gateway: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "nat_gateway_id": natGatewayID})
}

// =============================================================================
// NAT Gateway Rules
// =============================================================================

func (s *Server) listNatGatewayRules(datacenterID, natGatewayID string) (string, error) {
	rules, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysRulesGet(s.ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NAT gateway rules: %w", err)
	}

	return marshalResponse(rules, "NAT gateway rules")
}

func (s *Server) getNatGatewayRule(datacenterID, natGatewayID, ruleID string) (string, error) {
	rule, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysRulesFindByNatGatewayRuleId(s.ctx, datacenterID, natGatewayID, ruleID).Execute()
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

func (s *Server) createNatGatewayRule(datacenterID, natGatewayID, name, ruleType, protocol, sourceSubnet, publicIP, targetSubnet string, targetPortRangeStart, targetPortRangeEnd int32) (string, error) {
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
	if err := validatePortRange(targetPortRangeStart, targetPortRangeEnd, "target_port_range"); err != nil {
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

	result, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysRulesPost(s.ctx, datacenterID, natGatewayID).NatGatewayRule(rule).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create NAT gateway rule: %w", err)
	}

	return marshalResponse(result, "NAT gateway rule")
}

func (s *Server) updateNatGatewayRule(datacenterID, natGatewayID, ruleID, name, protocol, sourceSubnet, publicIP, targetSubnet string, targetPortRangeStart, targetPortRangeEnd int32, targetPortRangeStartSet, targetPortRangeEndSet bool) (string, error) {
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
		if err := validatePortRange(targetPortRangeStart, targetPortRangeEnd, "target_port_range"); err != nil {
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

	result, _, err := s.client.NATGatewaysApi.DatacentersNatgatewaysRulesPatch(s.ctx, datacenterID, natGatewayID, ruleID).NatGatewayRuleProperties(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update NAT gateway rule: %w", err)
	}

	return marshalResponse(result, "NAT gateway rule")
}

func (s *Server) deleteNatGatewayRule(datacenterID, natGatewayID, ruleID string) (string, error) {
	_, err := s.client.NATGatewaysApi.DatacentersNatgatewaysRulesDelete(s.ctx, datacenterID, natGatewayID, ruleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete NAT gateway rule: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "rule_id": ruleID})
}

// =============================================================================
// Networking - Private Cross Connects
// =============================================================================

func (s *Server) listPccs() (string, error) {
	pccs, _, err := s.client.PrivateCrossConnectsApi.PccsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Private Cross Connects: %w", err)
	}

	return marshalResponse(pccs, "Private Cross Connects")
}

func (s *Server) getPcc(pccID string) (string, error) {
	pcc, _, err := s.client.PrivateCrossConnectsApi.PccsFindById(s.ctx, pccID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Private Cross Connect: %w", err)
	}

	return marshalResponse(pcc, "Private Cross Connect")
}

func (s *Server) createPcc(name, description string) (string, error) {
	properties := ionoscloud.PrivateCrossConnectProperties{
		Name: &name,
	}
	if description != "" {
		properties.Description = &description
	}

	pcc := ionoscloud.PrivateCrossConnect{
		Properties: &properties,
	}

	result, _, err := s.client.PrivateCrossConnectsApi.PccsPost(s.ctx).Pcc(pcc).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create Private Cross Connect: %w", err)
	}

	return marshalResponse(result, "Private Cross Connect")
}

func (s *Server) updatePcc(pccID, name, description string) (string, error) {
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

	result, _, err := s.client.PrivateCrossConnectsApi.PccsPatch(s.ctx, pccID).Pcc(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update Private Cross Connect: %w", err)
	}

	return marshalResponse(result, "Private Cross Connect")
}

func (s *Server) deletePcc(pccID string) (string, error) {
	_, err := s.client.PrivateCrossConnectsApi.PccsDelete(s.ctx, pccID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete Private Cross Connect: %w", err)
	}

	return statusResponse(map[string]string{"status": "deleted", "pcc_id": pccID})
}

// =============================================================================
// Load Balancers - Application Load Balancers
// =============================================================================

func (s *Server) listApplicationLoadBalancers(datacenterID string) (string, error) {
	albs, _, err := s.client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersGet(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Application Load Balancers: %w", err)
	}

	return marshalResponse(albs, "Application Load Balancers")
}

func (s *Server) getApplicationLoadBalancer(datacenterID, albID string) (string, error) {
	alb, _, err := s.client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(s.ctx, datacenterID, albID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Application Load Balancer: %w", err)
	}

	return marshalResponse(alb, "Application Load Balancer")
}

func (s *Server) listAlbForwardingRules(datacenterID, albID string) (string, error) {
	rules, _, err := s.client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesGet(s.ctx, datacenterID, albID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list ALB forwarding rules: %w", err)
	}

	return marshalResponse(rules, "ALB forwarding rules")
}

func (s *Server) getAlbForwardingRule(datacenterID, albID, ruleID string) (string, error) {
	rule, _, err := s.client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId(s.ctx, datacenterID, albID, ruleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get ALB forwarding rule: %w", err)
	}

	return marshalResponse(rule, "ALB forwarding rule")
}

// =============================================================================
// Load Balancers - Network Load Balancers
// =============================================================================

func (s *Server) listNetworkLoadBalancers(datacenterID string) (string, error) {
	nlbs, _, err := s.client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(s.ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Network Load Balancers: %w", err)
	}

	return marshalResponse(nlbs, "Network Load Balancers")
}

func (s *Server) getNetworkLoadBalancer(datacenterID, nlbID string) (string, error) {
	nlb, _, err := s.client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(s.ctx, datacenterID, nlbID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Network Load Balancer: %w", err)
	}

	return marshalResponse(nlb, "Network Load Balancer")
}

// =============================================================================
// Load Balancers - Target Groups
// =============================================================================

func (s *Server) listTargetGroups() (string, error) {
	targetGroups, _, err := s.client.TargetGroupsApi.TargetgroupsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list target groups: %w", err)
	}

	return marshalResponse(targetGroups, "target groups")
}

func (s *Server) getTargetGroup(targetGroupID string) (string, error) {
	targetGroup, _, err := s.client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(s.ctx, targetGroupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get target group: %w", err)
	}

	return marshalResponse(targetGroup, "target group")
}

// =============================================================================
// Kubernetes
// =============================================================================

func (s *Server) listK8sClusters() (string, error) {
	clusters, _, err := s.client.KubernetesApi.K8sGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Kubernetes clusters: %w", err)
	}

	return marshalResponse(clusters, "Kubernetes clusters")
}

func (s *Server) getK8sCluster(k8sClusterID string) (string, error) {
	cluster, _, err := s.client.KubernetesApi.K8sFindByClusterId(s.ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Kubernetes cluster: %w", err)
	}

	return marshalResponse(cluster, "Kubernetes cluster")
}

func (s *Server) getK8sKubeconfig(k8sClusterID string) (string, error) {
	kubeconfig, _, err := s.client.KubernetesApi.K8sKubeconfigGet(s.ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	return marshalResponse(kubeconfig, "kubeconfig")
}

func (s *Server) listK8sNodepools(k8sClusterID string) (string, error) {
	nodepools, _, err := s.client.KubernetesApi.K8sNodepoolsGet(s.ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list node pools: %w", err)
	}

	return marshalResponse(nodepools, "node pools")
}

func (s *Server) getK8sNodepool(k8sClusterID, nodepoolID string) (string, error) {
	nodepool, _, err := s.client.KubernetesApi.K8sNodepoolsFindById(s.ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get node pool: %w", err)
	}

	return marshalResponse(nodepool, "node pool")
}

func (s *Server) listK8sNodes(k8sClusterID, nodepoolID string) (string, error) {
	nodes, _, err := s.client.KubernetesApi.K8sNodepoolsNodesGet(s.ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	return marshalResponse(nodes, "nodes")
}

func (s *Server) getK8sNode(k8sClusterID, nodepoolID, nodeID string) (string, error) {
	node, _, err := s.client.KubernetesApi.K8sNodepoolsNodesFindById(s.ctx, k8sClusterID, nodepoolID, nodeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get node: %w", err)
	}

	return marshalResponse(node, "node")
}

func (s *Server) listK8sVersions() (string, error) {
	versions, _, err := s.client.KubernetesApi.K8sVersionsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Kubernetes versions: %w", err)
	}

	return marshalResponse(versions, "Kubernetes versions")
}

// =============================================================================
// User Management - Users
// =============================================================================

func (s *Server) listUsers() (string, error) {
	users, _, err := s.client.UserManagementApi.UmUsersGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list users: %w", err)
	}

	return marshalResponse(users, "users")
}

func (s *Server) getUser(userID string) (string, error) {
	user, _, err := s.client.UserManagementApi.UmUsersFindById(s.ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	return marshalResponse(user, "user")
}

// =============================================================================
// User Management - Groups
// =============================================================================

func (s *Server) listGroups() (string, error) {
	groups, _, err := s.client.UserManagementApi.UmGroupsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list groups: %w", err)
	}

	return marshalResponse(groups, "groups")
}

func (s *Server) getGroup(groupID string) (string, error) {
	group, _, err := s.client.UserManagementApi.UmGroupsFindById(s.ctx, groupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get group: %w", err)
	}

	return marshalResponse(group, "group")
}

func (s *Server) listGroupMembers(groupID string) (string, error) {
	members, _, err := s.client.UserManagementApi.UmGroupsUsersGet(s.ctx, groupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list group members: %w", err)
	}

	return marshalResponse(members, "group members")
}

func (s *Server) listUserGroups(userID string) (string, error) {
	groups, _, err := s.client.UserManagementApi.UmUsersGroupsGet(s.ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list user groups: %w", err)
	}

	return marshalResponse(groups, "user groups")
}

// =============================================================================
// User Management - S3 Keys
// =============================================================================

func (s *Server) listS3Keys(userID string) (string, error) {
	keys, _, err := s.client.UserS3KeysApi.UmUsersS3keysGet(s.ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list S3 keys: %w", err)
	}

	return marshalResponse(keys, "S3 keys")
}

func (s *Server) getS3Key(userID, keyID string) (string, error) {
	key, _, err := s.client.UserS3KeysApi.UmUsersS3keysFindByKeyId(s.ctx, userID, keyID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get S3 key: %w", err)
	}

	return marshalResponse(key, "S3 key")
}

// =============================================================================
// User Management - Contract
// =============================================================================

func (s *Server) getContract() (string, error) {
	contract, _, err := s.client.ContractResourcesApi.ContractsGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get contract: %w", err)
	}

	return marshalResponse(contract, "contract")
}

func (s *Server) listResources(resourceType string) (string, error) {
	var resources ionoscloud.Resources
	var err error

	if resourceType != "" {
		resources, _, err = s.client.UserManagementApi.UmResourcesFindByType(s.ctx, resourceType).Execute()
	} else {
		resources, _, err = s.client.UserManagementApi.UmResourcesGet(s.ctx).Execute()
	}

	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	return marshalResponse(resources, "resources")
}

// =============================================================================
// DNS
// =============================================================================

func (s *Server) listDnsZones() (string, error) {
	zones, _, err := s.dnsClient.ZonesApi.ZonesGet(s.ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list DNS zones: %w", err)
	}

	return marshalResponse(zones, "DNS zones")
}

func (s *Server) getDnsZone(zoneID string) (string, error) {
	zone, _, err := s.dnsClient.ZonesApi.ZonesFindById(s.ctx, zoneID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS zone: %w", err)
	}

	return marshalResponse(zone, "DNS zone")
}

func (s *Server) listDnsRecords(zoneID string) (string, error) {
	records, _, err := s.dnsClient.RecordsApi.ZonesRecordsGet(s.ctx, zoneID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list DNS records: %w", err)
	}

	return marshalResponse(records, "DNS records")
}

func (s *Server) getDnsRecord(zoneID, recordID string) (string, error) {
	record, _, err := s.dnsClient.RecordsApi.ZonesRecordsFindById(s.ctx, zoneID, recordID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS record: %w", err)
	}

	return marshalResponse(record, "DNS record")
}
