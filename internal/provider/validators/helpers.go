package validators

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// conditionValueMatches checks if the condition value matches the expected value
func conditionValueMatches(ctx context.Context, condition, expected attr.Value) bool {
	// If types don't match, can't be equal
	if condition.Type(ctx) != expected.Type(ctx) {
		return false
	}
	if condition.IsNull() {
		return true
	}
	return condition.Equal(expected)
}
