package base

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func ImportSiteAndID(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if id := d.Id(); strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		d.SetId(importParts[1])
		d.Set("site", importParts[0])
	}
	return []*schema.ResourceData{d}, nil
}

func ImportIDWithSite(req resource.ImportStateRequest, resp *resource.ImportStateResponse) (string, string) {
	id := req.ID
	if id == "" {
		resp.Diagnostics.AddError("Invalid ID", "ID is required")
		return "", ""
	}

	if strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		if len(importParts) == 2 {
			return importParts[1], importParts[0]
		}
		resp.Diagnostics.AddError("Invalid ID", "ID contains too many colon-separated parts. Format should be 'site:id'")
		return "", ""
	}
	resp.Diagnostics.AddError("Invalid ID", "ID does not contain site part. Format should be 'site:id'")
	return id, ""
}
