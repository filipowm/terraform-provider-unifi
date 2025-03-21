package testing

import (
	"context"
	"errors"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"strings"
	"testing"
)

const TfAccLocal = "TF_ACC_LOCAL"

// MarkAccTest marks the test as acceptance test. Useful when executing code before resource.ParallelTest or resource.Test
// to bring acceptance test check earlier when test environment is required
func MarkAccTest(t *testing.T) {
	t.Helper()
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
		return
	}
}

func ImportStepWithSite(name string, ignore ...string) resource.TestStep {
	step := &resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: SiteAndIDImportStateIDFunc(name),
	}

	if len(ignore) > 0 {
		step.ImportStateVerifyIgnore = ignore
	}

	return *step
}

func ImportStep(name string, ignore ...string) resource.TestStep {
	step := resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}

	if len(ignore) > 0 {
		step.ImportStateVerifyIgnore = ignore
	}

	return step
}

// SiteAndIDImportStateIDFunc returns a function that can be used to import resources that require site and id.
func SiteAndIDImportStateIDFunc(resourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		id := rs.Primary.Attributes["id"]
		site := rs.Primary.Attributes["site"]
		return site + ":" + id, nil
	}
}

// PreCheck checks if provided environment variables are set. If not, it will fail the test.
func PreCheck(t *testing.T) {
	variables := []string{
		"UNIFI_USERNAME",
		"UNIFI_PASSWORD",
		"UNIFI_API",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for acceptance tests!", variable)
		}
	}
}

func CheckPlanPreApply(checks ...plancheck.PlanCheck) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PreApply: checks,
	}
}

func CheckResourceActions(resourceAddress string, actions ...plancheck.ResourceActionType) resource.ConfigPlanChecks {
	var checks []plancheck.PlanCheck
	for _, a := range actions {
		checks = append(checks, plancheck.ExpectResourceAction(resourceAddress, a))
	}
	return CheckPlanPreApply(checks...)
}

func ComposeConfig(configs ...string) string {
	return strings.Join(configs, "\n")
}

func SkipIfEnvMissing(t *testing.T, msg string, env string) {
	t.Helper()
	if os.Getenv(env) == "" {
		t.Skip(msg)
	}
}

func SkipIfEnvLocalMissing(t *testing.T, msg string) {
	t.Helper()
	SkipIfEnvMissing(t, msg, TfAccLocal)
}

func CheckDestroy(resourceType string, read func(ctx context.Context, site, id string) error) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type == "" || rs.Type != resourceType {
				continue
			}
			site := "default"
			if s, ok := rs.Primary.Attributes["site"]; ok {
				if s != "" {
					site = s
				}
			}
			err := read(context.Background(), site, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Resource with id %q still exists.", rs.Primary.ID)
			}
			if utils.IsServerErrorStatusCode(err, 404) || errors.Is(err, unifi.ErrNotFound) {
				continue
			}
			return err
		}
		return nil
	}
}
