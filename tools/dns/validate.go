package dns

import (
	"fmt"
	"net"
	"sort"
	"strings"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// The DNS SDK models the spec's numeric bounds and enums as plain ints and string
// aliases with no validation, so every constraint below would otherwise surface as
// a round-trip 400. Shared by the create and update tools so both normalise alike.

const (
	minTTL = 60
	maxTTL = 604800

	maxPriority = 65535

	minValidity = 90
	maxValidity = 365

	maxNsec3Iterations = 50
	minNsec3SaltBits   = 64
	maxNsec3SaltBits   = 128

	defaultKskBits         = 4096
	defaultZskBits         = 2048
	defaultNsec3SaltBits   = minNsec3SaltBits
	defaultNsec3Iterations = 0
)

// recordTypes is the set the SDK's RecordType accepts. Its UnmarshalJSON rejects
// anything else, so an unknown type breaks reads as well as writes.
var recordTypes = map[string]dnsSDK.RecordType{
	"A": dnsSDK.RECORDTYPE_A, "AAAA": dnsSDK.RECORDTYPE_AAAA, "CNAME": dnsSDK.RECORDTYPE_CNAME,
	"ALIAS": dnsSDK.RECORDTYPE_ALIAS, "MX": dnsSDK.RECORDTYPE_MX, "NS": dnsSDK.RECORDTYPE_NS,
	"SRV": dnsSDK.RECORDTYPE_SRV, "TXT": dnsSDK.RECORDTYPE_TXT, "CAA": dnsSDK.RECORDTYPE_CAA,
	"SSHFP": dnsSDK.RECORDTYPE_SSHFP, "TLSA": dnsSDK.RECORDTYPE_TLSA, "SMIMEA": dnsSDK.RECORDTYPE_SMIMEA,
	"DS": dnsSDK.RECORDTYPE_DS, "HTTPS": dnsSDK.RECORDTYPE_HTTPS, "SVCB": dnsSDK.RECORDTYPE_SVCB,
	"OPENPGPKEY": dnsSDK.RECORDTYPE_OPENPGPKEY, "CERT": dnsSDK.RECORDTYPE_CERT,
	"URI": dnsSDK.RECORDTYPE_URI, "RP": dnsSDK.RECORDTYPE_RP, "LOC": dnsSDK.RECORDTYPE_LOC,
}

// priorityTypes are the types for which the API requires a priority. Every other
// type ignores it.
var priorityTypes = map[dnsSDK.RecordType]bool{
	dnsSDK.RECORDTYPE_MX: true, dnsSDK.RECORDTYPE_SRV: true, dnsSDK.RECORDTYPE_URI: true,
}

// normalizeRecordType upper-cases and validates a record type, returning a message
// naming the accepted values when it does not match.
func normalizeRecordType(raw string) (dnsSDK.RecordType, string) {
	t, ok := recordTypes[strings.ToUpper(strings.TrimSpace(raw))]
	if !ok {
		names := make([]string, 0, len(recordTypes))
		for n := range recordTypes {
			names = append(names, n)
		}
		sort.Strings(names)
		return "", fmt.Sprintf("type %q is not a DNS record type; use one of: %s", raw, strings.Join(names, ", "))
	}
	return t, ""
}

func validateTTL(ttl *int32) string {
	if ttl == nil {
		return ""
	}
	if *ttl < minTTL || *ttl > maxTTL {
		return fmt.Sprintf("ttl %d is out of range; it must be between %d and %d seconds", *ttl, minTTL, maxTTL)
	}
	return ""
}

// validatePriority checks the range and enforces the API's requirement that MX,
// SRV and URI records carry one. required is false on an update, where an omitted
// priority is carried forward from the existing record instead.
func validatePriority(priority *int32, t dnsSDK.RecordType, required bool) string {
	if priority == nil {
		if required && priorityTypes[t] {
			return fmt.Sprintf("priority is required for a %s record; it must be between 0 and %d", t, maxPriority)
		}
		return ""
	}
	if *priority < 0 || *priority > maxPriority {
		return fmt.Sprintf("priority %d is out of range; it must be between 0 and %d", *priority, maxPriority)
	}
	return ""
}

// validatePrimaryIps enforces the spec's minItems and uniqueItems and parses each
// address. A nil slice would serialize as "primaryIps":null, which the API rejects.
func validatePrimaryIps(ips []string) ([]string, string) {
	if len(ips) == 0 {
		return nil, "primary_ips must list at least one primary nameserver IP"
	}
	seen := make(map[string]bool, len(ips))
	out := make([]string, 0, len(ips))
	for _, raw := range ips {
		ip := strings.TrimSpace(raw)
		if net.ParseIP(ip) == nil {
			return nil, fmt.Sprintf("primary_ips entry %q is not an IPv4 or IPv6 address", raw)
		}
		if seen[ip] {
			return nil, fmt.Sprintf("primary_ips contains %s twice; the API requires unique entries", ip)
		}
		seen[ip] = true
		out = append(out, ip)
	}
	return out, ""
}

func validateIP(field, raw string) (string, string) {
	ip := strings.TrimSpace(raw)
	if net.ParseIP(ip) == nil {
		return "", fmt.Sprintf("%s %q is not an IPv4 or IPv6 address", field, raw)
	}
	return ip, ""
}

// buildDnssecParams assembles the signing parameters, applying defaults for the
// optional fields and enforcing the bounds and the kskBits >= zskBits invariant
// that the spec states in prose and the SDK does not check.
func buildDnssecParams(in tools.CreateDnsDnssecKeyInput) (dnsSDK.DnssecKeyParameters, string) {
	var zero dnsSDK.DnssecKeyParameters

	if in.Validity < minValidity || in.Validity > maxValidity {
		return zero, fmt.Sprintf("validity %d is out of range; it must be between %d and %d days", in.Validity, minValidity, maxValidity)
	}

	algorithm := dnsSDK.ALGORITHM_RSASHA256
	if in.Algorithm != nil {
		if got := strings.ToUpper(strings.TrimSpace(*in.Algorithm)); got != string(dnsSDK.ALGORITHM_RSASHA256) {
			return zero, fmt.Sprintf("algorithm %q is not supported; RSASHA256 is the only value the API accepts", *in.Algorithm)
		}
	}

	ksk, msg := validateKeyBits("ksk_bits", in.KskBits, defaultKskBits)
	if msg != "" {
		return zero, msg
	}
	zsk, msg := validateKeyBits("zsk_bits", in.ZskBits, defaultZskBits)
	if msg != "" {
		return zero, msg
	}
	if ksk < zsk {
		return zero, fmt.Sprintf("ksk_bits (%d) must be greater than or equal to zsk_bits (%d)", ksk, zsk)
	}

	nsecMode := dnsSDK.NSECMODE_NSEC3
	if in.NsecMode != nil {
		switch strings.ToUpper(strings.TrimSpace(*in.NsecMode)) {
		case string(dnsSDK.NSECMODE_NSEC):
			nsecMode = dnsSDK.NSECMODE_NSEC
		case string(dnsSDK.NSECMODE_NSEC3):
			nsecMode = dnsSDK.NSECMODE_NSEC3
		default:
			return zero, fmt.Sprintf("nsec_mode %q is not supported; use NSEC or NSEC3", *in.NsecMode)
		}
	}

	iterations := int32(defaultNsec3Iterations)
	if in.Nsec3Iterations != nil {
		if *in.Nsec3Iterations < 0 || *in.Nsec3Iterations > maxNsec3Iterations {
			return zero, fmt.Sprintf("nsec3_iterations %d is out of range; it must be between 0 and %d", *in.Nsec3Iterations, maxNsec3Iterations)
		}
		iterations = *in.Nsec3Iterations
	}

	saltBits := int32(defaultNsec3SaltBits)
	if in.Nsec3SaltBits != nil {
		switch {
		case *in.Nsec3SaltBits < minNsec3SaltBits || *in.Nsec3SaltBits > maxNsec3SaltBits:
			return zero, fmt.Sprintf("nsec3_salt_bits %d is out of range; it must be between %d and %d", *in.Nsec3SaltBits, minNsec3SaltBits, maxNsec3SaltBits)
		case *in.Nsec3SaltBits%8 != 0:
			return zero, fmt.Sprintf("nsec3_salt_bits %d must be a multiple of 8", *in.Nsec3SaltBits)
		}
		saltBits = *in.Nsec3SaltBits
	}

	return dnsSDK.DnssecKeyParameters{
		KeyParameters: dnsSDK.KeyParameters{
			Algorithm: algorithm,
			KskBits:   dnsSDK.KskBits(ksk),
			ZskBits:   dnsSDK.ZskBits(zsk),
		},
		NsecParameters: dnsSDK.NsecParameters{
			NsecMode:        nsecMode,
			Nsec3Iterations: iterations,
			Nsec3SaltBits:   saltBits,
		},
		Validity: in.Validity,
	}, ""
}

func validateKeyBits(field string, bits *int32, def int32) (int32, string) {
	if bits == nil {
		return def, ""
	}
	switch *bits {
	case 1024, 2048, 4096:
		return *bits, ""
	}
	return 0, fmt.Sprintf("%s %d is not supported; use 1024, 2048 or 4096", field, *bits)
}
