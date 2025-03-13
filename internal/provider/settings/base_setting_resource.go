package settings

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

// NewSettingResource creates a new base setting resource
func NewSettingResource[T base.ResourceModel](
	typeName string,
	modelFactory func() T,
	getter func(context.Context, *base.Client, string) (interface{}, error),
	updater func(context.Context, *base.Client, string, interface{}) (interface{}, error),
) *base.GenericResource[T] {
	return base.NewGenericResource(
		typeName,
		modelFactory,
		base.ResourceFunctions{
			Read: func(ctx context.Context, client *base.Client, site, _ string) (interface{}, error) {
				return getter(ctx, client, site)
			},
			Create: updater,
			Update: updater,
		})
}
