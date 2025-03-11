package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// using dedicated site for each test, because USG settings might interfere with parallel tests of other resources

// using an additional lock to the one around the resource to avoid deadlocking accidentally
var settingUsgLock = sync.Mutex{}

func TestAccSettingUsg_mdns_v6(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "< 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_mdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_mdns(false),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_mdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_mdns_v7(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingUsgSite() + testAccSettingUsgConfig_mdns(true),
				ExpectError: regexp.MustCompile("multicast_dns_enabled is not supported"),
			},
		},
	})
}

func TestAccSettingUsg_dhcpRelayServers(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dhcpRelay(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_geoIpFiltering(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_geoIpFilteringBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "KP"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_geoIpFilteringAllow(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "allow"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "US"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CA"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "GB"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_geoIpFilteringDirections(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "ingress"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_geoIpFilteringDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_geoIpFilteringBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "KP"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_upnp(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_upnpBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.enabled", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_upnpAdvanced(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.nat_pmp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.secure_mode", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.wan_interface", "WAN"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_upnpDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.enabled", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_dnsVerification(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dnsVerification(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "dns_verification.domain"),
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "dns_verification.primary_dns_server"),
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "dns_verification.secondary_dns_server"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dns_verification.setting_preference", "auto"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dnsVerificationUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dns_verification.domain", "example.com"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dns_verification.primary_dns_server", "1.1.1.1"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dns_verification.secondary_dns_server", "1.0.0.1"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dns_verification.setting_preference", "manual"),
				),
			},
		},
	})
}
func TestAccSettingUsg_tcpTimeouts(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_tcpTimeouts(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.close_timeout", "10"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.established_timeout", "3600"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.close_wait_timeout", "20"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.fin_wait_timeout", "30"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.last_ack_timeout", "30"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.syn_recv_timeout", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.syn_sent_timeout", "120"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.time_wait_timeout", "120"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_tcpTimeoutsUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.close_timeout", "20"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.established_timeout", "7200"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.close_wait_timeout", "40"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.fin_wait_timeout", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.last_ack_timeout", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.syn_recv_timeout", "120"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.syn_sent_timeout", "240"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tcp_timeouts.time_wait_timeout", "240"),
				),
			},
		},
	})
}

func TestAccSettingUsg_arpCache(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_arpCache(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_base_reachable", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_timeout", "custom"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_arpCacheUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_base_reachable", "120"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_timeout", "normal"),
				),
			},
		},
	})
}

func TestAccSettingUsg_dhcpConfig(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dhcpConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "broadcast_ping", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_hostfile_update", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_use_dnsmasq", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dnsmasq_all_servers", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dhcpConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "broadcast_ping", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_hostfile_update", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_use_dnsmasq", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dnsmasq_all_servers", "false"),
				),
			},
		},
	})
}

func TestAccSettingUsg_dhcpRelayConfig(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dhcpRelayConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.agents_packets", "forward"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.hop_count", "5"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.max_size", "1400"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.port", "67"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay_servers.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "dhcp_relay_servers.*", "10.1.2.3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "dhcp_relay_servers.*", "10.1.2.4"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_dhcpRelayConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.agents_packets", "replace"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.hop_count", "10"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.max_size", "64"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay.port", "68"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcp_relay_servers.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "dhcp_relay_servers.*", "10.1.2.5"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "dhcp_relay_servers.*", "10.1.2.6"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "dhcp_relay_servers.*", "10.1.2.7"),
				),
			},
		},
	})
}

func TestAccSettingUsg_networkTools(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_networkTools(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "echo_server", "echo.example.com"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_protocolModules(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_protocolModules(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "ftp_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "gre_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "h323_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "pptp_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "sip_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tftp_module", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_protocolModulesUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "ftp_module", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "gre_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "h323_module", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "pptp_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "sip_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tftp_module", "false"),
				),
			},
		},
	})
}

func TestAccSettingUsg_icmpAndLldp(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_icmpAndLldp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "icmp_timeout", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "lldp_enable_all", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_icmpAndLldpUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "icmp_timeout", "120"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "lldp_enable_all", "false"),
				),
			},
		},
	})
}

func TestAccSettingUsg_mssClamp(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_mssClamp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mss_clamp", "auto"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mss_clamp_mss", "1452"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_mssClampUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mss_clamp", "custom"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mss_clamp_mss", "1400"),
				),
			},
		},
	})
}

func TestAccSettingUsg_offloadSettings(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_offloadSettings(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_accounting", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_l2_blocking", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_sch", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_offloadSettingsUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_accounting", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_l2_blocking", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_sch", "false"),
				),
			},
		},
	})
}

func TestAccSettingUsg_timeoutSettings(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_timeoutSettings(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "other_timeout", "600"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "timeout_setting_preference", "auto"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_timeoutSettingsUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "other_timeout", "1200"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "timeout_setting_preference", "manual"),
				),
			},
		},
	})
}

func TestAccSettingUsg_redirectsAndSecurity(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_redirectsAndSecurity(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "receive_redirects", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "send_redirects", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "syn_cookies", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_redirectsAndSecurityUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "receive_redirects", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "send_redirects", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "syn_cookies", "false"),
				),
			},
		},
	})
}

func TestAccSettingUsg_udp(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_udp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_other_timeout", "30"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_stream_timeout", "120"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_udpUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_other_timeout", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_stream_timeout", "240"),
				),
			},
		},
	})
}

func TestAccSettingUsg_unbindWanMonitor(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_unbindWanMonitor(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "unbind_wan_monitors", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_unbindWanMonitor(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "unbind_wan_monitors", "false"),
				),
			},
		},
	})
}

func TestAccSettingUsg_comprehensive(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfig_comprehensive(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_site.test", "id"),
					// ARP Cache
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_base_reachable", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_timeout", "custom"),

					// DHCP Config
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "broadcast_ping", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_hostfile_update", "true"),

					// Protocol Modules (sample)
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "ftp_module", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "tftp_module", "true"),

					// Timeouts
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "other_timeout", "600"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_stream_timeout", "120"),

					// Security
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "syn_cookies", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}
func testAccSettingUsgSite() string {
	return `
resource "unifi_site" "test" {
	description = "tfacc-setting-usg"
}
`
}

func testAccSettingUsgConfig_mdns(mdns bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usg" "test" {
	multicast_dns_enabled = %t
	site = unifi_site.test.name
}
`, mdns)
}

func testAccSettingUsgConfig_dhcpRelay() string {
	return `
resource "unifi_setting_usg" "test" {
	dhcp_relay_servers = [
		"10.1.2.3",
		"10.1.2.4",
	]
	site = unifi_site.test.name
}
`
}
func testAccSettingUsgConfig_geoIpFilteringBasic() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		enabled = true
		countries = ["RU", "CN", "KP"]
	}
}
`
}

func testAccSettingUsgConfig_geoIpFilteringAllow() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		enabled = true
		block = "allow"
		countries = ["US", "CA", "GB"]
	}
}
`
}

func testAccSettingUsgConfig_geoIpFilteringDirections() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		enabled = true
		traffic_direction = "ingress"
		countries = ["RU", "CN"]
	}
}
`
}

func testAccSettingUsgConfig_geoIpFilteringDisabled() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		enabled = false
	}
}
`
}

func testAccSettingUsgConfig_upnpBasic() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	upnp = {
		enabled = true
	}
}
`
}

func testAccSettingUsgConfig_upnpAdvanced() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	upnp = {
		enabled = true
		nat_pmp_enabled = true
		secure_mode = true
		wan_interface = "WAN"
	}
}
`
}

func testAccSettingUsgConfig_upnpDisabled() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	upnp = {
		enabled = false
	}
}
`
}

func testAccSettingUsgConfig_dnsVerification() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
  	dns_verification = {
    	setting_preference  = "auto"
  	}
}
`
}

func testAccSettingUsgConfig_dnsVerificationUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  dns_verification = {
    domain              = "example.com"
    primary_dns_server  = "1.1.1.1"
    secondary_dns_server = "1.0.0.1"
    setting_preference  = "manual"
  }
}
`
}

func testAccSettingUsgConfig_tcpTimeouts() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  tcp_timeouts = {
    close_timeout       = 10
    established_timeout = 3600
    close_wait_timeout  = 20
    fin_wait_timeout    = 30
    last_ack_timeout    = 30
    syn_recv_timeout    = 60
    syn_sent_timeout    = 120
    time_wait_timeout   = 120
  }
}
`
}

func testAccSettingUsgConfig_tcpTimeoutsUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  tcp_timeouts = {
    close_timeout       = 20
    established_timeout = 7200
    close_wait_timeout  = 40
    fin_wait_timeout    = 60
    last_ack_timeout    = 60
    syn_recv_timeout    = 120
    syn_sent_timeout    = 240
    time_wait_timeout   = 240
  }
}
`
}

func testAccSettingUsgConfig_arpCache() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  arp_cache_base_reachable = 60
  arp_cache_timeout = "custom"
}
`
}

func testAccSettingUsgConfig_dhcpConfig() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  broadcast_ping = true
  dhcpd_hostfile_update = true
  dhcpd_use_dnsmasq = true
  dnsmasq_all_servers = true
}
`
}

func testAccSettingUsgConfig_dhcpRelayConfig() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  dhcp_relay = {
	agents_packets = "forward"
	hop_count = 5
	max_size = 1400
	port = 67
  }
  dhcp_relay_servers = ["10.1.2.3","10.1.2.4"]
}
`
}

func testAccSettingUsgConfig_dhcpRelayConfigUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  dhcp_relay = {
	agents_packets = "replace"
	hop_count = 10
	max_size = 64
	port = 68
  }
  dhcp_relay_servers = ["10.1.2.5","10.1.2.6","10.1.2.7"]
}
`
}

func testAccSettingUsgConfig_networkTools() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  echo_server = "echo.example.com"
}
`
}

func testAccSettingUsgConfig_protocolModules() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  ftp_module = true
  gre_module = true
  h323_module = true
  pptp_module = true
  sip_module = true
  tftp_module = true
}
`
}

func testAccSettingUsgConfig_icmpAndLldp() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  icmp_timeout = 60
  lldp_enable_all = true
}
`
}

func testAccSettingUsgConfig_mssClamp() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  mss_clamp = "auto"
  mss_clamp_mss = 1452
}
`
}

func testAccSettingUsgConfig_offloadSettings() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  offload_accounting = true
  offload_l2_blocking = true
  offload_sch = true
}
`
}

func testAccSettingUsgConfig_timeoutSettings() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  other_timeout = 600
  timeout_setting_preference = "auto"
}
`
}

func testAccSettingUsgConfig_redirectsAndSecurity() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  receive_redirects = false
  send_redirects = true
  syn_cookies = true
}
`
}

func testAccSettingUsgConfig_udp() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  udp_other_timeout = 30
  udp_stream_timeout = 120
}
`
}

func testAccSettingUsgConfig_comprehensive() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  // ARP Cache Configuration
  arp_cache_base_reachable = 60
  arp_cache_timeout = "custom"

  // DHCP Configuration
  broadcast_ping = true
  dhcpd_hostfile_update = true
  dhcpd_use_dnsmasq = true
  dnsmasq_all_servers = true

  // DHCP Relay
  dhcp_relay = {
	agents_packets = "forward"
	hop_count = 5
  }
  dhcp_relay_servers = ["10.1.2.3", "10.1.2.4"]

  // Network Tools
  echo_server = "echo.example.com"

  // Protocol Modules
  ftp_module = true
  gre_module = true
  tftp_module = true

  // ICMP & LLDP
  icmp_timeout = 20
  lldp_enable_all = true

  // MSS Clamp
  mss_clamp = "auto"
  mss_clamp_mss = 1452

  // Offload Settings
  offload_accounting = true
  offload_l2_blocking = true

  // Timeout Settings
  other_timeout = 600
  timeout_setting_preference = "auto"

  // TCP Settings
  tcp_timeouts = {
    close_timeout = 10
    established_timeout = 3600
    close_wait_timeout = 20
    fin_wait_timeout = 30
    last_ack_timeout = 30
    syn_recv_timeout = 60
    syn_sent_timeout = 120
    time_wait_timeout = 120
  }

  // Redirects & Security
  receive_redirects = false
  send_redirects = true
  syn_cookies = true

  // UDP
  udp_other_timeout = 30
  udp_stream_timeout = 120

  // Geo IP Filtering
  geo_ip_filtering = {
    enabled = true
    block = "block"
    countries = ["RU", "CN"]
    traffic_direction = "both"
  }

  // UPNP Settings
  upnp = {
    enabled = true
    nat_pmp_enabled = true
    secure_mode = true
  }
}
`
}

func testAccSettingUsgConfig_arpCacheUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  arp_cache_base_reachable = 120
  arp_cache_timeout = "normal"
}
`
}

func testAccSettingUsgConfig_dhcpConfigUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  broadcast_ping = false
  dhcpd_hostfile_update = false
  dhcpd_use_dnsmasq = false
  dnsmasq_all_servers = false
}
`
}

func testAccSettingUsgConfig_protocolModulesUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  ftp_module = false
  gre_module = true
  h323_module = false
  pptp_module = true
  sip_module = true
  tftp_module = false
}
`
}

func testAccSettingUsgConfig_icmpAndLldpUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  icmp_timeout = 120
  lldp_enable_all = false
}
`
}

func testAccSettingUsgConfig_mssClampUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  mss_clamp = "custom"
  mss_clamp_mss = 1400
}
`
}

func testAccSettingUsgConfig_offloadSettingsUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  offload_accounting = false
  offload_l2_blocking = false
  offload_sch = false
}
`
}

func testAccSettingUsgConfig_timeoutSettingsUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  other_timeout = 1200
  timeout_setting_preference = "manual"
}
`
}

func testAccSettingUsgConfig_redirectsAndSecurityUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  receive_redirects = true
  send_redirects = false
  syn_cookies = false
}
`
}

func testAccSettingUsgConfig_udpUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  udp_other_timeout = 60
  udp_stream_timeout = 240
}
`
}

func testAccSettingUsgConfig_unbindWanMonitor(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usg" "test" {
  unbind_wan_monitors = %t
}
`, enabled)
}
