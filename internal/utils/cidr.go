package utils

import (
	"fmt"
	"net"
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

func CidrOneBased(cidr string) string {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}

	cidrNet.IP[3]++

	return cidrNet.String()
}
