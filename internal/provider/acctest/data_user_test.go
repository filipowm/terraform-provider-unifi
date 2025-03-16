package acctest

import (
	"context"
	"fmt"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"net"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataUser_default(t *testing.T) {
	mac, unallocateTestMac := pt.AllocateTestMac(t)
	defer unallocateTestMac()
	name := acctest.RandomWithPrefix("tfacc")

	AcceptanceTest(t, AcceptanceTestCase{
		PreCheck: func() {
			_, err := testClient.CreateUser(context.Background(), "default", &unifi.User{
				MAC:  mac,
				Name: name,
				Note: name,
			})
			if err != nil {
				t.Fatal(err)
			}
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataUserConfig_default(mac),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_user.test", "id"),
					resource.TestCheckResourceAttr("data.unifi_user.test", "mac", mac),
					resource.TestCheckResourceAttr("data.unifi_user.test", "name", name),
				),
			},
		},
	})
}

func TestAccDataUser_localDnsRecord(t *testing.T) {
	mac, unallocateTestMac := pt.AllocateTestMac(t)
	defer unallocateTestMac()
	name := acctest.RandomWithPrefix("tfacc")
	ctx := context.Background()
	n, err := testClient.ListNetwork(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(n) == 0 {
		t.Fatal("no networks found, but default should exist")
	}
	_, subnet, err := net.ParseCIDR(n[0].IPSubnet)
	if err != nil {
		t.Fatal(err)
	}
	ip, err := cidr.Host(subnet, 1)
	if err != nil {
		t.Fatal(err)
	}
	AcceptanceTest(t, AcceptanceTestCase{
		PreCheck: func() {
			_, err = testClient.CreateUser(ctx, "default", &unifi.User{
				MAC:                   mac,
				Name:                  name,
				UseFixedIP:            true,
				NetworkID:             n[0].ID,
				FixedIP:               ip.String(),
				LocalDNSRecord:        "myuser.local",
				LocalDNSRecordEnabled: true,
				Note:                  name,
			})
			if err != nil {
				t.Fatal(err)
			}
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataUserConfig_default(mac),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_user.test", "id"),
					resource.TestCheckResourceAttr("data.unifi_user.test", "mac", mac),
					resource.TestCheckResourceAttr("data.unifi_user.test", "name", name),
					resource.TestCheckResourceAttr("data.unifi_user.test", "local_dns_record", "myuser.local"),
					resource.TestCheckResourceAttr("data.unifi_user.test", "fixed_ip", ip.String()),
				),
			},
		},
	})
}

func testAccDataUserConfig_default(mac string) string {
	return fmt.Sprintf(`
data "unifi_user" "test" {
mac = "%s"
}
`, mac)
}
