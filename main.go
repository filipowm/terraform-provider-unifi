package main // import "github.com/paultyng/terraform-provider-unifi"

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/paultyng/terraform-provider-unifi/internal/provider"
)

// Generate docs for website
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	ctx := context.Background()

	sdkv2Provider := provider.New(version)()
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		context.Background(),
		sdkv2Provider.GRPCProvider,
	)
	if err != nil {
		panic(err)
	}

	frameworkProvider := provider.NewFrameworkProvider(version)()
	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
		providerserver.NewProtocol6(frameworkProvider),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		panic(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts,
			tf6server.WithManagedDebug(),
			tf6server.WithGoDebug(),
		)
	}

	err = tf6server.Serve(
		"registry.terraform.io/paultyng/unifi",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		panic(err)
	}
}
