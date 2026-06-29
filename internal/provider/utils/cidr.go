package utils

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
	"strings"
)

func CidrValidate(raw interface{}, key string) ([]string, []error) {
	v, ok := raw.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected string, got %T", raw)}
	}

	_, _, err := net.ParseCIDR(v)
	if err != nil {
		return nil, []error{err}
	}

	return nil, nil
}

func CidrZeroBased(cidr string) string {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}

	return cidrNet.String()
}

// CidrListZeroBased canonicalizes each CIDR in the list to its network address
// (see CidrZeroBased). An element that fails to parse is returned unchanged.
func CidrListZeroBased(cidrs []string) []string {
	out := make([]string, len(cidrs))
	for i, cidr := range cidrs {
		if canonical := CidrZeroBased(cidr); canonical != "" {
			out[i] = canonical
		} else {
			out[i] = cidr
		}
	}
	return out
}

func CidrOneBased(cidr string) string {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}

	cidrNet.IP[3]++

	return cidrNet.String()
}

func CidrDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	_, oldNet, err := net.ParseCIDR(old)
	if err != nil {
		return false
	}

	_, newNet, err := net.ParseCIDR(new)
	if err != nil {
		return false
	}

	return oldNet.String() == newNet.String()
}

// IsIPv4 checks if the provided address is a valid IPv4 address.
// It returns true if the address is a valid IPv4 address, false otherwise.
func IsIPv4(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.To4() != nil
}

// IsIPv6 checks if the provided address is a valid IPv6 address.
// It returns true if the address is a valid IPv6 address, false otherwise.
func IsIPv6(address string) bool {

	// Handle zone index if present
	if idx := strings.Index(address, "%"); idx != -1 {
		address = address[:idx]
	}

	// Handle IPv4-mapped addresses
	isIPv4Mapped := strings.Contains(address, "::ffff:") && strings.Count(address, ".") == 3

	ip := net.ParseIP(address)
	if ip == nil || (!isIPv4Mapped && ip.To4() != nil) {
		return false
	}
	return true
}
