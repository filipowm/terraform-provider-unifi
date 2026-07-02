package acctest

import (
	"fmt"
	"regexp"
	"sync"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// using dedicated site for each test, because USG settings might interfere with parallel tests of other resources

// using an additional lock to the one around the resource to avoid deadlocking accidentally.
var settingUsgLock = sync.Mutex{}

func TestAccSettingUsg_mdns_v6(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "< 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigMdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigMdns(false),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigMdns(true),
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
				Config:      testAccSettingUsgSite() + testAccSettingUsgConfigMdns(true),
				ExpectError: regexp.MustCompile("multicast_dns_enabled is not supported"),
			},
		},
	})
}

func TestAccSettingUsg_dhcpRelayServers(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping: dhcp_relay_servers requires an adopted gateway not available on the Docker test controller")
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDhcpRelay(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigGeoIPFilteringBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.mode", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "KP"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigGeoIPFilteringAllow(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.mode", "allow"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "US"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CA"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "GB"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigGeoIPFilteringDirections(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.mode", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "ingress"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigGeoIPFilteringDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering_enabled", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigGeoIPFilteringBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.mode", "block"),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUpnpBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_enabled", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUpnpAdvanced(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.nat_pmp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.secure_mode", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.wan_interface", "WAN"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUpnpDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_enabled", "false"),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDNSVerification(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "dns_verification.domain"),
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "dns_verification.primary_dns_server"),
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "dns_verification.secondary_dns_server"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dns_verification.setting_preference", "auto"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDNSVerificationUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigTCPTimeouts(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigTCPTimeoutsUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigArpCache(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_base_reachable", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "arp_cache_timeout", "custom"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigArpCacheUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDhcpConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "broadcast_ping", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_hostfile_update", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dhcpd_use_dnsmasq", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "dnsmasq_all_servers", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDhcpConfigUpdated(),
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
	pt.SkipIfEnvLocalMissing(t, "Skipping: dhcp_relay_servers requires an adopted gateway not available on the Docker test controller")
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDhcpRelayConfig(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigDhcpRelayConfigUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigNetworkTools(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigProtocolModules(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigProtocolModulesUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigIcmpAndLldp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "icmp_timeout", "60"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "lldp_enable_all", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigIcmpAndLldpUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigMssClamp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mss_clamp", "auto"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mss_clamp_mss", "1452"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigMssClampUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigOffloadSettings(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_accounting", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_l2_blocking", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "offload_sch", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigOffloadSettingsUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigTimeoutSettings(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "other_timeout", "600"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "timeout_setting_preference", "auto"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigTimeoutSettingsUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigRedirectsAndSecurity(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "receive_redirects", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "send_redirects", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "syn_cookies", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigRedirectsAndSecurityUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUDP(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_other_timeout", "30"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "udp_stream_timeout", "120"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUDPUpdated(),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUnbindWanMonitor(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "unbind_wan_monitors", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigUnbindWanMonitor(false),
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
				Config: testAccSettingUsgSite() + testAccSettingUsgConfigComprehensive(),
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

func testAccSettingUsgConfigMdns(mdns bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usg" "test" {
	multicast_dns_enabled = %t
	site = unifi_site.test.name
}
`, mdns)
}

func testAccSettingUsgConfigDhcpRelay() string {
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

func testAccSettingUsgConfigGeoIPFilteringBasic() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		countries = ["RU", "CN", "KP"]
	}
}
`
}

func testAccSettingUsgConfigGeoIPFilteringAllow() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		mode = "allow"
		countries = ["US", "CA", "GB"]
	}
}
`
}

func testAccSettingUsgConfigGeoIPFilteringDirections() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	geo_ip_filtering = {
		traffic_direction = "ingress"
		countries = ["RU", "CN"]
	}
}
`
}

func testAccSettingUsgConfigGeoIPFilteringDisabled() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
}
`
}

func testAccSettingUsgConfigUpnpBasic() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	upnp = {
	}
}
`
}

func testAccSettingUsgConfigUpnpAdvanced() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
	upnp = {
		nat_pmp_enabled = true
		secure_mode = true
		wan_interface = "WAN"
	}
}
`
}

func testAccSettingUsgConfigUpnpDisabled() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
}
`
}

func testAccSettingUsgConfigDNSVerification() string {
	return `
resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
  	dns_verification = {
    	setting_preference  = "auto"
  	}
}
`
}

func testAccSettingUsgConfigDNSVerificationUpdated() string {
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

func testAccSettingUsgConfigTCPTimeouts() string {
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

func testAccSettingUsgConfigTCPTimeoutsUpdated() string {
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

func testAccSettingUsgConfigArpCache() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  arp_cache_base_reachable = 60
  arp_cache_timeout = "custom"
}
`
}

func testAccSettingUsgConfigDhcpConfig() string {
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

func testAccSettingUsgConfigDhcpRelayConfig() string {
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

func testAccSettingUsgConfigDhcpRelayConfigUpdated() string {
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

func testAccSettingUsgConfigNetworkTools() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  echo_server = "echo.example.com"
}
`
}

func testAccSettingUsgConfigProtocolModules() string {
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

func testAccSettingUsgConfigIcmpAndLldp() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  icmp_timeout = 60
  lldp_enable_all = true
}
`
}

func testAccSettingUsgConfigMssClamp() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  mss_clamp = "auto"
  mss_clamp_mss = 1452
}
`
}

func testAccSettingUsgConfigOffloadSettings() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  offload_accounting = true
  offload_l2_blocking = true
  offload_sch = true
}
`
}

func testAccSettingUsgConfigTimeoutSettings() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  other_timeout = 600
  timeout_setting_preference = "auto"
}
`
}

func testAccSettingUsgConfigRedirectsAndSecurity() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  receive_redirects = false
  send_redirects = true
  syn_cookies = true
}
`
}

func testAccSettingUsgConfigUDP() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  udp_other_timeout = 30
  udp_stream_timeout = 120
}
`
}

func testAccSettingUsgConfigComprehensive() string {
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
    mode = "block"
    countries = ["RU", "CN"]
    traffic_direction = "both"
  }

  // UPNP Settings
  upnp = {
    nat_pmp_enabled = true
    secure_mode = true
  }
}
`
}

func testAccSettingUsgConfigArpCacheUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  arp_cache_base_reachable = 120
  arp_cache_timeout = "normal"
}
`
}

func testAccSettingUsgConfigDhcpConfigUpdated() string {
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

func testAccSettingUsgConfigProtocolModulesUpdated() string {
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

func testAccSettingUsgConfigIcmpAndLldpUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  icmp_timeout = 120
  lldp_enable_all = false
}
`
}

func testAccSettingUsgConfigMssClampUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  mss_clamp = "custom"
  mss_clamp_mss = 1400
}
`
}

func testAccSettingUsgConfigOffloadSettingsUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  offload_accounting = false
  offload_l2_blocking = false
  offload_sch = false
}
`
}

func testAccSettingUsgConfigTimeoutSettingsUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  other_timeout = 1200
  timeout_setting_preference = "manual"
}
`
}

func testAccSettingUsgConfigRedirectsAndSecurityUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  receive_redirects = true
  send_redirects = false
  syn_cookies = false
}
`
}

func testAccSettingUsgConfigUDPUpdated() string {
	return `
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  udp_other_timeout = 60
  udp_stream_timeout = 240
}
`
}

func testAccSettingUsgConfigUnbindWanMonitor(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usg" "test" {
  site = unifi_site.test.name
  unbind_wan_monitors = %t
}
`, enabled)
}
