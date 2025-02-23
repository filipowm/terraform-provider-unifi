package provider

const (
	ProviderUsernameDescription = "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` environment variable."
	ProviderPasswordDescription = "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` environment variable."
	ProviderAPIKeyDescription   = "API Key for the user accessing the API. Can be specified with the `UNIFI_API_KEY` environment variable. Controller version 9.0.108 or later is required."
	ProviderAPIURLDescription   = "URL of the controller API. Can be specified with the `UNIFI_API` environment variable. " +
		"You should **NOT** supply the path (`/api`), the SDK will discover the appropriate paths. This is to support UDM Pro style API paths as well as more standard controller paths."
	ProviderSiteDescription          = "The site in the Unifi controller this provider will manage. Can be specified with the `UNIFI_SITE` environment variable. Default: `default`"
	ProviderAllowInsecureDescription = "Skip verification of TLS certificates of API requests. You may need to set this to `true` " +
		"if you are using your local API without setting up a signed certificate. Can be specified with the " +
		"`UNIFI_INSECURE` environment variable."
)
