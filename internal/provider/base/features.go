package base

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type FeatureValidator interface {
	RequireFeaturesEnabled(ctx context.Context, site string, features ...string) diag.Diagnostics
	RequireFeaturesEnabledForPath(ctx context.Context, site string, attrPath path.Path, config tfsdk.Config, features ...string) diag.Diagnostics
}

type Features map[string]bool

func (v Features) IsEnabled(feature string) bool {
	return v[feature]
}

func (v Features) IsDisabled(feature string) bool {
	return !v[feature]
}

type featureEnabledValidator struct {
	client *Client
	cache  map[string]Features

	lock sync.Mutex
}

func NewFeatureValidator(client *Client) FeatureValidator {
	return &featureEnabledValidator{client: client, cache: make(map[string]Features), lock: sync.Mutex{}}
}

func (v *featureEnabledValidator) getFeatures(ctx context.Context, site string) Features {
	if v.cache[site] != nil {
		return v.cache[site]
	}
	v.lock.Lock()
	defer v.lock.Unlock()
	if v.cache[site] != nil {
		return v.cache[site]
	}
	cache := make(map[string]bool)
	features, err := v.client.ListFeatures(ctx, site)
	if err != nil {
		// Return an empty Features map instead of nil to avoid potential nil pointer dereference
		return Features{}
	}
	for _, feature := range features {
		cache[feature.Name] = feature.FeatureExists
	}
	v.cache[site] = cache
	return v.cache[site]
}

func (v *featureEnabledValidator) requireFeatures(ctx context.Context, site string, attrPath *path.Path, features ...string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if len(features) == 0 {
		return diags
	}

	errorBuilder := strings.Builder{}
	f := v.getFeatures(ctx, site)
	disabledFeatures := make([]string, 0)
	for _, feature := range features {
		if !f.IsEnabled(feature) {
			disabledFeatures = append(disabledFeatures, feature)
		}
	}
	if len(disabledFeatures) > 0 {
		if attrPath != nil {
			errorBuilder.WriteString(fmt.Sprintf("%s is not supported. ", attrPath.String()))
		}
		errorBuilder.WriteString(fmt.Sprintf("Features %s must be enabled, but %s are disabled", strings.Join(features, ", "), strings.Join(disabledFeatures, ", ")))
		diags.AddError("Features not enabled", errorBuilder.String())
	}

	return diags

}

func (v *featureEnabledValidator) RequireFeaturesEnabled(ctx context.Context, site string, features ...string) diag.Diagnostics {
	return v.requireFeatures(ctx, site, nil, features...)
}

func (v *featureEnabledValidator) RequireFeaturesEnabledForPath(ctx context.Context, site string, attrPath path.Path, config tfsdk.Config, features ...string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	var val attr.Value
	diags.Append(config.GetAttribute(context.Background(), attrPath, &val)...)
	if diags.HasError() {
		return diags
	}
	if !IsDefined(val) {
		return diags
	}
	diags.Append(v.requireFeatures(ctx, site, &attrPath, features...)...)
	return diags
}
