package v1

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
)

func cidrDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
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
