package testing

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// MarkAccTest marks the test as acceptance test. Useful when executing code before resource.ParallelTest or resource.Test
// to bring acceptance test check earlier when test environment is required
func MarkAccTest(t *testing.T) {
	t.Helper()
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc))
		return
	}
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

// AssertComputedAttributes checks that the given schema has the given computed attributes.
func AssertComputedAttributes(t *testing.T, s map[string]*schema.Schema, keys []string) {
	t.Helper()

	for _, v := range keys {
		require.NotNil(t, s[v], "Error in Schema: Missing definition for \"%s\"", v)
		assert.True(t, s[v].Computed, "Error in Schema: Attribute \"%s\" is not computed", v)
	}
}

// AssertNestedSchemaExistence checks that the given schema has a nested schema for the given key.
func AssertNestedSchemaExistence(t *testing.T, s map[string]*schema.Schema, key string) map[string]*schema.Schema {
	t.Helper()

	r, ok := s[key].Elem.(*schema.Resource)

	if !ok {
		t.Fatalf("Error in Schema: Missing nested schema for \"%s\"", key)

		return nil
	}

	return r.Schema
}

// AssertListMaxItems checks that the given schema attribute has given expected MaxItems value.
func AssertListMaxItems(t *testing.T, s map[string]*schema.Schema, key string, expectedMaxItems int) {
	t.Helper()

	require.NotNil(t, s[key], "Error in Schema: Missing definition for \"%s\"", key)
	assert.Equal(t, expectedMaxItems, s[key].MaxItems,
		"Error in Schema: Argument \"%s\" has \"MaxItems: %#v\", but value %#v is expected!",
		key, s[key].MaxItems, expectedMaxItems)
}

// AssertOptionalArguments checks that the given schema has the given optional arguments.
func AssertOptionalArguments(t *testing.T, s map[string]*schema.Schema, keys []string) {
	t.Helper()

	for _, v := range keys {
		require.NotNil(t, s[v], "Error in Schema: Missing definition for \"%s\"", v)
		assert.True(t, s[v].Optional, "Error in Schema: Argument \"%s\" is not optional", v)
	}
}

// AssertRequiredArguments checks that the given schema has the given required arguments.
func AssertRequiredArguments(t *testing.T, s map[string]*schema.Schema, keys []string) {
	t.Helper()

	for _, v := range keys {
		require.NotNil(t, s[v], "Error in Schema: Missing definition for \"%s\"", v)
		assert.True(t, s[v].Required, "Error in Schema: Argument \"%s\" is not required", v)
	}
}

// AssertValueTypes checks that the given schema has the given value types for the given fields.
func AssertValueTypes(t *testing.T, s map[string]*schema.Schema, f map[string]schema.ValueType) {
	t.Helper()

	for fn, ft := range f {
		require.NotNil(t, s[fn], "Error in Schema: Missing definition for \"%s\"", fn)
		assert.Equal(t, ft, s[fn].Type, "Error in Schema: Argument or attribute \"%s\" is not of type \"%v\"", fn, ft)
	}
}
