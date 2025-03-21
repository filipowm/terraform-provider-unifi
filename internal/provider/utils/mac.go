package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"regexp"
	"strings"
)

var MacAddressRegexp = regexp.MustCompile("^([0-9a-fA-F][0-9a-fA-F][-:]){5}([0-9a-fA-F][0-9a-fA-F])$")

func CleanMAC(mac string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ToLower(mac), "-", ":"))
}

func MacDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	old = CleanMAC(old)
	new = CleanMAC(new)
	return old == new
}
