package ionos

import (
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{"empty string is valid", "", false},
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv6", "2001:db8::1", false},
		{"valid CIDR", "192.168.1.0/24", false},
		{"valid IPv6 CIDR", "2001:db8::/32", false},
		{"invalid IP", "not-an-ip", true},
		{"invalid CIDR", "192.168.1.0/99", true},
		{"SQL injection attempt", "'; DROP TABLE users; --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIP(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMAC(t *testing.T) {
	tests := []struct {
		name    string
		mac     string
		wantErr bool
	}{
		{"empty string is valid", "", false},
		{"valid MAC lowercase", "aa:bb:cc:dd:ee:ff", false},
		{"valid MAC uppercase", "AA:BB:CC:DD:EE:FF", false},
		{"valid MAC mixed case", "Aa:Bb:Cc:Dd:Ee:Ff", false},
		{"invalid MAC - too short", "aa:bb:cc", true},
		{"invalid MAC - wrong separator", "aa-bb-cc-dd-ee-ff", true},
		{"invalid MAC - no separators", "aabbccddeeff", true},
		{"invalid MAC - invalid hex", "gg:hh:ii:jj:kk:ll", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMAC(tt.mac)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMAC(%q) error = %v, wantErr %v", tt.mac, err, tt.wantErr)
			}
		})
	}
}

func TestValidateLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		wantErr  bool
	}{
		{"valid de/fra", "de/fra", false},
		{"valid de/txl", "de/txl", false},
		{"valid us/las", "us/las", false},
		{"valid us/ewr", "us/ewr", false},
		{"valid gb/lhr", "gb/lhr", false},
		{"valid es/vit", "es/vit", false},
		{"valid fr/par", "fr/par", false},
		{"invalid location", "invalid/loc", true},
		{"empty location", "", true},
		{"SQL injection", "'; DROP TABLE --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocation(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLocation(%q) error = %v, wantErr %v", tt.location, err, tt.wantErr)
			}
		})
	}
}

func TestValidateProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{"valid TCP", "TCP", false},
		{"valid UDP", "UDP", false},
		{"valid ICMP", "ICMP", false},
		{"valid ICMPv6", "ICMPv6", false},
		{"valid GRE", "GRE", false},
		{"valid ESP", "ESP", false},
		{"valid AH", "AH", false},
		{"valid ANY", "ANY", false},
		{"invalid protocol", "INVALID", true},
		{"lowercase tcp", "tcp", true},
		{"empty protocol", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProtocol(tt.protocol)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProtocol(%q) error = %v, wantErr %v", tt.protocol, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePortRange(t *testing.T) {
	tests := []struct {
		name    string
		start   int32
		end     int32
		wantErr bool
	}{
		{"both unset is valid", 0, 0, false},
		{"valid single port", 22, 22, false},
		{"valid port range", 80, 443, false},
		{"only start set", 22, 0, false},
		{"only end set", 0, 443, false},
		{"start too low", -1, 0, true},
		{"start too high", 70000, 0, true},
		{"end too low", 0, -1, true},
		{"end too high", 0, 70000, true},
		{"start greater than end", 443, 80, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortRange(tt.start, tt.end, "port_range")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortRange(%d, %d, %q) error = %v, wantErr %v",
					tt.start, tt.end, "port_range", err, tt.wantErr)
			}
		})
	}
}

func TestValidateICMP(t *testing.T) {
	tests := []struct {
		name     string
		icmpType int32
		icmpCode int32
		typeSet  bool
		codeSet  bool
		wantErr  bool
	}{
		{"both unset is valid", 0, 0, false, false, false},
		{"valid type and code", 8, 0, true, true, false},
		{"type only", 8, 0, true, false, false},
		{"code only", 0, 0, false, true, false},
		{"type too low", -1, 0, true, false, true},
		{"type too high", 256, 0, true, false, true},
		{"code too low", 0, -1, false, true, true},
		{"code too high", 0, 256, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateICMP(tt.icmpType, tt.icmpCode, tt.typeSet, tt.codeSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateICMP(%d, %d, %v, %v) error = %v, wantErr %v",
					tt.icmpType, tt.icmpCode, tt.typeSet, tt.codeSet, err, tt.wantErr)
			}
		})
	}
}

func TestValidateVolumeType(t *testing.T) {
	tests := []struct {
		name       string
		volumeType string
		wantErr    bool
	}{
		{"valid HDD", "HDD", false},
		{"valid SSD", "SSD", false},
		{"valid SSD Standard", "SSD Standard", false},
		{"valid SSD Premium", "SSD Premium", false},
		{"valid DAS", "DAS", false},
		{"invalid type", "INVALID", true},
		{"lowercase hdd", "hdd", true},
		{"empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVolumeType(tt.volumeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVolumeType(%q) error = %v, wantErr %v", tt.volumeType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBusType(t *testing.T) {
	tests := []struct {
		name    string
		busType string
		wantErr bool
	}{
		{"valid VIRTIO", "VIRTIO", false},
		{"valid IDE", "IDE", false},
		{"invalid type", "SCSI", true},
		{"lowercase virtio", "virtio", true},
		{"empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBusType(tt.busType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBusType(%q) error = %v, wantErr %v", tt.busType, err, tt.wantErr)
			}
		})
	}
}
