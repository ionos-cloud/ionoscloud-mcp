package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/mail"
	"net/url"
	"slices"
	"strings"
	"time"
)

// Input validation for the Certificate Manager write tools. Each function returns
// (value, message); a non-empty message means reject the call with that text.
// The SDK types all of these as plain strings, so an unchecked mistake surfaces as a
// round-trip 422 that never names the field.

// keyAlgorithms is the enum the spec allows for an auto-certificate's key.
var keyAlgorithms = []string{"rsa2048", "rsa3072", "rsa4096"}

func normalizeKeyAlgorithm(raw string) (string, string) {
	a := strings.ToLower(strings.TrimSpace(raw))
	if !slices.Contains(keyAlgorithms, a) {
		return "", fmt.Sprintf("key_algorithm %q is not supported; use one of: %s", raw, strings.Join(keyAlgorithms, ", "))
	}
	return a, ""
}

// validatePEM checks that a field holds PEM text. The value is never quoted back:
// one of the three fields it guards is the private key.
func validatePEM(field, value string) string {
	if strings.TrimSpace(value) == "" {
		return field + " is required"
	}
	if block, _ := pem.Decode([]byte(strings.TrimSpace(value))); block == nil {
		return fmt.Sprintf("%s is not PEM; pass the file's contents including the -----BEGIN ... ----- lines, not a path or bare base64", field)
	}
	return ""
}

func validateEmail(raw string) (string, string) {
	e := strings.TrimSpace(raw)
	// ParseAddress alone would accept "Name <a@b>", which is not what the API stores.
	if addr, err := mail.ParseAddress(e); err != nil || addr.Address != e {
		return "", fmt.Sprintf("email %q is not a plain email address, e.g. user@example.com", raw)
	}
	return e, ""
}

// validateACMEServer requires https: the ACME account key is exchanged over this
// URL, and every public provider's directory is https.
func validateACMEServer(raw string) (string, string) {
	s := strings.TrimSpace(raw)
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", fmt.Sprintf("server %q is not an https ACME directory URL, e.g. https://acme-v02.api.letsencrypt.org/directory", raw)
	}
	return s, ""
}

// cleanNames trims a DNS name list and drops blanks, keeping the caller's order.
func cleanNames(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, n := range raw {
		if n = strings.TrimSpace(n); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// certificateSummary describes the leaf certificate for a preview without dumping
// it, so a caller can spot the wrong PEM before authorizing the upload.
func certificateSummary(pemText string) string {
	block, _ := pem.Decode([]byte(strings.TrimSpace(pemText)))
	if block == nil {
		return ""
	}
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "1 PEM block, not parseable as a certificate"
	}
	summary := fmt.Sprintf("CN=%s, expires %s", crt.Subject.CommonName, crt.NotAfter.UTC().Format(time.RFC3339))
	if len(crt.DNSNames) > 0 {
		summary += ", SAN " + strings.Join(crt.DNSNames, " ")
	}
	return summary
}

// pemBlockCount reports how many PEM blocks a chain holds, for the same reason.
func pemBlockCount(pemText string) string {
	rest, n := []byte(strings.TrimSpace(pemText)), 0
	for {
		var block *pem.Block
		if block, rest = pem.Decode(rest); block == nil {
			break
		}
		n++
	}
	switch n {
	case 0:
		return ""
	case 1:
		return "1 PEM block"
	default:
		return fmt.Sprintf("%d PEM blocks", n)
	}
}
