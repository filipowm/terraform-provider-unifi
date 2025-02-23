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
	ControllerVersionApiKeyAuth = asVersion("9.0.108")

	// https://community.ui.com/releases/UniFi-Network-Controller-6-1-61/62f1ad38-1ac5-430c-94b0-becbb8f71d7d
	ControllerVersionWPA3 = asVersion("6.1.61")
)

func (c *Client) IsControllerV6() bool {
	return c.Version.GreaterThanOrEqual(ControllerV6)
}

func (c *Client) IsControllerV7() bool {
	return c.Version.GreaterThanOrEqual(ControllerV7)
}

func (c *Client) SupportsApiKeyAuthentication() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionApiKeyAuth)
}

func (c *Client) SupportsWPA3() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionWPA3)
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
