package acctest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type providersMap map[string]func() (tfprotov6.ProviderServer, error)

var (
	providers  = createProviders()
	testClient unifi.Client
)

type Steps []resource.TestStep

type AcceptanceTestCase struct {
	CheckDestroy      resource.TestCheckFunc
	VersionConstraint string
	MinVersion        *version.Version
	PreCheck          func()
	Steps             Steps
	Lock              *sync.Mutex
}

func AcceptanceTest(t *testing.T, testCase AcceptanceTestCase) {
	t.Helper()
	if len(testCase.Steps) == 0 {
		t.Fatal("missing test steps")
	}

	// Core/matrix scope gating. UNIFI_ACCTEST_SCOPE selects which subset of tests this
	// binary runs so CI can split version-independent tests (run once, on "latest") from
	// version-specific tests (run across the controller matrix):
	//   - "core"           -> run only version-INDEPENDENT tests (skip version-gated ones).
	//   - "matrix"         -> run only version-SPECIFIC tests (skip version-independent ones).
	//   - "all"/unset/else -> run everything (default; preserves local `make testacc` and full/release CI).
	// A test case is "version-gated" iff it pins a controller version via MinVersion or
	// VersionConstraint. The skip happens here, before resource.ParallelTest, so it occurs
	// before any controller interaction (PreCheck / provider configuration / API calls).
	SkipForScope(t, testCase.MinVersion != nil || testCase.VersionConstraint != "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			if testCase.VersionConstraint != "" {
				PreCheckVersionConstraint(t, testCase.VersionConstraint)
			}
			if testCase.MinVersion != nil {
				PreCheckMinVersion(t, testCase.MinVersion)
			}
			if testCase.PreCheck != nil {
				testCase.PreCheck()
			}
			if testCase.Lock != nil {
				testCase.Lock.Lock()
				t.Cleanup(func() {
					testCase.Lock.Unlock()
				})
			}
		},
		ProtoV6ProviderFactories: providers,
		CheckDestroy:             testCase.CheckDestroy,
		Steps:                    testCase.Steps,
	})
}

// SkipForScope applies the UNIFI_ACCTEST_SCOPE gating (see AcceptanceTest): in
// the "core" scope a version-gated test is skipped, and in the "matrix" scope a
// version-independent test is skipped. AcceptanceTest calls this itself, but a
// test that performs expensive setup BEFORE calling AcceptanceTest (e.g. device
// discovery) must call it first so that setup is not run for a test that is about
// to be skipped.
func SkipForScope(t *testing.T, gated bool) {
	t.Helper()
	switch strings.ToLower(strings.TrimSpace(os.Getenv("UNIFI_ACCTEST_SCOPE"))) {
	case "core":
		if gated {
			t.Skip("skipped in core scope: version-gated test runs in the matrix job")
		}
	case "matrix":
		if !gated {
			t.Skip("skipped in matrix scope: version-independent test runs in the core job")
		}
	}
}

func TestMain(m *testing.M) {
	providers = createProviders()
	os.Exit(pt.Run(m, func(env *pt.TestEnvironment) {
		testClient = env.Client
	}))
}

func createProviders() providersMap {
	ctx := context.Background()
	// Init mux servers
	return map[string]func() (tfprotov6.ProviderServer, error){
		"unifi": func() (tfprotov6.ProviderServer, error) {
			return tf6muxserver.NewMuxServer(ctx,
				providerserver.NewProtocol6(provider.NewV2("acctestv2")()),
				func() tfprotov6.ProviderServer {
					sdkV2Provider, err := tf5to6server.UpgradeServer(
						ctx,
						func() tfprotov5.ProviderServer {
							return schema.NewGRPCProviderServer(
								provider.New("acctestv1")(),
							)
						},
					)
					if err != nil {
						panic(fmt.Errorf("failed to create test providers: %w", err))
					}

					return sdkV2Provider
				},
			)
		},
	}
}
