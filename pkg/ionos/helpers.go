package ionos

import (
	"fmt"
	"net"
	"regexp"
)

// Valid IONOS Cloud locations
var ValidLocations = map[string]bool{
	"de/fra": true, "de/txl": true, "us/las": true, "us/ewr": true,
	"gb/lhr": true, "es/vit": true, "fr/par": true,
}

// Valid firewall protocols
var ValidProtocols = map[string]bool{
	"TCP": true, "UDP": true, "ICMP": true, "ICMPv6": true,
	"GRE": true, "ESP": true, "AH": true, "ANY": true,
}

// MAC address regex pattern
var macAddressRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)

// ValidateIP validates an IP address or CIDR.
func ValidateIP(ip string) error {
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

// ValidateMAC validates a MAC address.
func ValidateMAC(mac string) error {
	if mac == "" {
		return nil
	}
	if !macAddressRegex.MatchString(mac) {
		return fmt.Errorf("invalid MAC address format: %s (expected XX:XX:XX:XX:XX:XX)", mac)
	}
	return nil
}

// ValidateLocation validates an IONOS location.
func ValidateLocation(location string) error {
	if !ValidLocations[location] {
		return fmt.Errorf("invalid location: %s (valid: de/fra, de/txl, us/las, us/ewr, gb/lhr, es/vit, fr/par)", location)
	}
	return nil
}

// ValidateProtocol validates a firewall protocol.
func ValidateProtocol(protocol string) error {
	if !ValidProtocols[protocol] {
		return fmt.Errorf("invalid protocol: %s (valid: TCP, UDP, ICMP, ICMPv6, GRE, ESP, AH, ANY)", protocol)
	}
	return nil
}

// ValidatePortRange validates port range values. fieldPrefix is used in error messages
// (e.g., "port_range" or "target_port_range").
func ValidatePortRange(start, end int32, fieldPrefix string) error {
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

// ValidateICMP validates ICMP type and code.
func ValidateICMP(icmpType, icmpCode int32, typeSet, codeSet bool) error {
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

// ValidateVolumeType validates a volume type.
func ValidateVolumeType(volumeType string) error {
	if volumeType == "" {
		return nil
	}
	validTypes := map[string]bool{"HDD": true, "SSD": true, "SSD Standard": true, "SSD Premium": true, "DAS": true}
	if !validTypes[volumeType] {
		return fmt.Errorf("invalid volume type: %s (valid: HDD, SSD, SSD Standard, SSD Premium, DAS)", volumeType)
	}
	return nil
}

// ValidateBusType validates a bus type.
func ValidateBusType(bus string) error {
	if bus == "" {
		return nil
	}
	validBus := map[string]bool{"VIRTIO": true, "IDE": true}
	if !validBus[bus] {
		return fmt.Errorf("invalid bus type: %s (valid: VIRTIO, IDE)", bus)
	}
	return nil
}
