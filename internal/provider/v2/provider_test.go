package v2

import (
	"context"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	v1 "github.com/filipowm/terraform-provider-unifi/internal/provider/v1"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
)

func TestProviderInstantiation(t *testing.T) {
	t.Parallel()
	p := New("acctest")()
	if p == nil {
		t.Fatal("cannot instantiate UniFi Provider")
	}
}

func TestMain(m *testing.M) {
	os.Exit(pt.Run(m))
}

var (
	providersMutex = sync.Mutex{}
	providers      map[string]func() (tfprotov6.ProviderServer, error)
)

func MuxProviders(t *testing.T) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	providersMutex.Lock()
	defer providersMutex.Unlock()
	if len(providers) > 0 {
		return providers
	}
	ctx := context.Background()
	// Init mux servers
	p := map[string]func() (tfprotov6.ProviderServer, error){
		"unifi": func() (tfprotov6.ProviderServer, error) {
			return tf6muxserver.NewMuxServer(ctx,
				providerserver.NewProtocol6(New("acctestv2")()),
				func() tfprotov6.ProviderServer {
					sdkV2Provider, err := tf5to6server.UpgradeServer(
						ctx,
						func() tfprotov5.ProviderServer {
							return schema.NewGRPCProviderServer(
								v1.New("acctestv1")(),
							)
						},
					)
					require.NoError(t, err)

					return sdkV2Provider
				},
			)
		},
	}
	providers = p
	return providers
}
