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

const (
	featureEnabled featureStatus = iota
	featureDisabled
)

type featureStatus int

type FeatureValidator interface {
	RequireFeaturesEnabled(ctx context.Context, site string, features ...string) diag.Diagnostics
	RequireFeaturesEnabledForPath(ctx context.Context, site string, attrPath path.Path, config tfsdk.Config, features ...string) diag.Diagnostics
}

type Features map[string]featureStatus

func (v Features) IsEnabled(feature string) bool {
	return !v.IsUnavailable(feature) && v[feature] == featureEnabled
}

func (v Features) IsDisabled(feature string) bool {
	return !v.IsUnavailable(feature) && v[feature] == featureDisabled
}

func (v Features) IsUnavailable(feature string) bool {
	if _, ok := v[feature]; ok {
		return false
	}
	return true
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
	cache := make(map[string]featureStatus)
	features, err := v.client.ListFeatures(ctx, site)
	if err != nil {
		// Return an empty Features map instead of nil to avoid potential nil pointer dereference
		return Features{}
	}
	for _, feature := range features {
		if feature.FeatureExists {
			cache[feature.Name] = featureEnabled
		} else {
			cache[feature.Name] = featureDisabled
		}
	}
	v.cache[site] = cache
	return v.cache[site]
}

func (v *featureEnabledValidator) requireFeatures(ctx context.Context, site string, attrPath *path.Path, features ...string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if len(features) == 0 {
		return diags
	}

	f := v.getFeatures(ctx, site)
	var unavailableFeatures, disabledFeatures []string
	for _, feature := range features {
		if f.IsUnavailable(feature) {
			unavailableFeatures = append(unavailableFeatures, feature)
		}
		if f.IsDisabled(feature) {
			disabledFeatures = append(disabledFeatures, feature)
		}
	}
	pathInfo := ""
	if attrPath != nil {
		pathInfo = fmt.Sprintf("%s is not supported. ", attrPath.String())
	}
	if len(unavailableFeatures) > 0 {
		diags.AddError("Controller features not available", fmt.Sprintf("%sFeatures %s must be available on controller, but %s are not", pathInfo, strings.Join(features, ", "), strings.Join(unavailableFeatures, ", ")))
	}
	if len(disabledFeatures) > 0 {
		diags.AddError("Controller features not disabled", fmt.Sprintf("%sFeatures %s must be enabled on controller, but %s are disabled", pathInfo, strings.Join(features, ", "), strings.Join(disabledFeatures, ", ")))
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
