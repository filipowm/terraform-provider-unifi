package testing

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func RandHostname() string {
	return RandHostnameWithSuffix("test.com")
}

func RandHostnameWithSuffix(suffix string) string {
	return fmt.Sprintf("%s.%s", RandAlpha(10), suffix)
}

func RandAlpha(n int) string {
	return acctest.RandStringFromCharSet(n, acctest.CharSetAlpha)
}

func RandIPAddress() string {
	ip, _ := acctest.RandIpAddress("192.168.0.1/24")
	return ip
}
