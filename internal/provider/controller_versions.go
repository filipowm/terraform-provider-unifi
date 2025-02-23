package provider

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func asVersion(versionString string) *version.Version {
	return version.Must(version.NewVersion(versionString))
}

var (
	controllerV6                = asVersion("6.0.0")
	controllerV7                = asVersion("7.0.0")
	controllerVersionApiKeyAuth = asVersion("9.0.108")

	// https://community.ui.com/releases/UniFi-Network-Controller-6-1-61/62f1ad38-1ac5-430c-94b0-becbb8f71d7d
	controllerVersionWPA3 = asVersion("6.1.61")
)

func (c *client) IsControllerV6() bool {
	return c.version.GreaterThanOrEqual(controllerV6)
}

func (c *client) IsControllerV7() bool {
	return c.version.GreaterThanOrEqual(controllerV7)
}

func (c *client) SupportsApiKeyAuthentication() bool {
	return c.version.GreaterThanOrEqual(controllerVersionApiKeyAuth)
}

func (c *client) SupportsWPA3() bool {
	return c.version.GreaterThanOrEqual(controllerVersionWPA3)
}

func checkMinimumControllerVersion(versionString string) error {
	v, err := version.NewVersion(versionString)
	if err != nil {
		return err
	}
	if v.LessThan(controllerV6) {
		return fmt.Errorf("Controller version %q or greater is required to use the provider, found %q.", controllerV6, v)
	}
	return nil
}
