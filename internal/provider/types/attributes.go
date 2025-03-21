package types

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// ID generates an attribute definition suitable for the always-present `id` attribute.
func ID(desc ...string) schema.StringAttribute {
	a := schema.StringAttribute{
		Description: "The unique identifier of this resource.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	if len(desc) > 0 {
		a.Description = desc[0]
	}

	return a
}

func SiteAttribute(desc ...string) schema.StringAttribute {
	s := schema.StringAttribute{
		MarkdownDescription: "The name of the UniFi site where this resource should be applied. If not specified, the default site will be used.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		},
	}

	if len(desc) > 0 {
		s.MarkdownDescription = desc[0]
	}
	return s
}
