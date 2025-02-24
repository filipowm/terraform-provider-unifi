package provider

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func asVersion(versionString string) *version.Version {
	return version.Must(version.NewVersion(versionString))
}

var (
	ControllerV6                = asVersion("6.0.0")
	ControllerV7                = asVersion("7.0.0")
	ControllerV9                = asVersion("9.0.0")
	ControllerVersionApiKeyAuth = asVersion("9.0.108")
	// https://community.ui.com/releases/UniFi-Network-Application-8-2-93/fce86dc6-897a-4944-9c53-1eec7e37e738
	ControllerVersionDnsRecords = asVersion("8.2.93")

	// https://community.ui.com/releases/UniFi-Network-Controller-6-1-61/62f1ad38-1ac5-430c-94b0-becbb8f71d7d
	ControllerVersionWPA3 = asVersion("6.1.61")
)

func (c *Client) IsControllerV6() bool {
	return c.Version.GreaterThanOrEqual(ControllerV6)
}

func (c *Client) IsControllerV7() bool {
	return c.Version.GreaterThanOrEqual(ControllerV7)
}

func (c *Client) IsControllerV9() bool {
	return c.Version.GreaterThanOrEqual(ControllerV9)
}

func (c *Client) SupportsApiKeyAuthentication() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionApiKeyAuth)
}

func (c *Client) SupportsWPA3() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionWPA3)
}

func (c *Client) SupportsDnsRecords() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionDnsRecords)
}

func CheckMinimumControllerVersion(versionString string) error {
	v, err := version.NewVersion(versionString)
	if err != nil {
		return err
	}
	if v.LessThan(ControllerV6) {
		return fmt.Errorf("Controller version %q or greater is required to use the provider, found %q.", ControllerV6, v)
	}
	return nil
}
