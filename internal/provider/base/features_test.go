package base

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// MockUnifiClient provides a minimal implementation of unifi.Client for testing
type MockUnifiClient struct {
	unifi.Client // Embed the interface to satisfy all methods (they'll panic if called)
	featuresFunc func(ctx context.Context, site string) ([]unifi.DescribedFeature, error)
}

// ListFeatures implements the only unifi.Client method we care about for testing
func (m *MockUnifiClient) ListFeatures(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
	if m.featuresFunc != nil {
		return m.featuresFunc(ctx, site)
	}
	return nil, errors.New("ListFeatures not implemented")
}

// TestFeaturesIsEnabled tests the IsEnabled method of Features
func TestFeaturesIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		features Features
		feature  string
		expected bool
	}{
		{
			name:     "feature is enabled",
			features: Features{"feature1": featureEnabled, "feature2": featureDisabled},
			feature:  "feature1",
			expected: true,
		},
		{
			name:     "feature is disabled",
			features: Features{"feature1": featureEnabled, "feature2": featureDisabled},
			feature:  "feature2",
			expected: false,
		},
		{
			name:     "feature does not exist",
			features: Features{"feature1": featureEnabled},
			feature:  "feature2",
			expected: false,
		},
		{
			name:     "empty features map",
			features: Features{},
			feature:  "feature1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.features.IsEnabled(tt.feature)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFeaturesIsDisabled tests the IsDisabled method of Features
func TestFeaturesIsDisabled(t *testing.T) {
	tests := []struct {
		name     string
		features Features
		feature  string
		expected bool
	}{
		{
			name:     "feature is enabled",
			features: Features{"feature1": featureEnabled, "feature2": featureDisabled},
			feature:  "feature1",
			expected: false,
		},
		{
			name:     "feature is disabled",
			features: Features{"feature1": featureEnabled, "feature2": featureDisabled},
			feature:  "feature2",
			expected: true,
		},
		{
			name:     "feature does not exist",
			features: Features{"feature1": featureEnabled},
			feature:  "feature2",
			expected: false,
		},
		{
			name:     "empty features map",
			features: Features{},
			feature:  "feature1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.features.IsDisabled(tt.feature)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFeaturesIsUnavailable tests the IsUnavailable method of Features
func TestFeaturesIsUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		features Features
		feature  string
		expected bool
	}{
		{
			name:     "feature is enabled",
			features: Features{"feature1": featureEnabled, "feature2": featureDisabled},
			feature:  "feature2",
			expected: false,
		},
		{
			name:     "feature is disabled",
			features: Features{"feature1": featureEnabled},
			feature:  "feature2",
			expected: true,
		},
		{
			name:     "feature does not exist",
			features: Features{},
			feature:  "feature2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.features.IsUnavailable(tt.feature)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func newTestClient(mock *MockUnifiClient) *Client {
	return &Client{
		Client: mock,
		Site:   "default",
	}
}

// TestNewFeatureValidator tests the NewFeatureValidator function
func TestNewFeatureValidator(t *testing.T) {
	mockUnifiClient := &MockUnifiClient{
		featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
			return []unifi.DescribedFeature{}, nil
		},
	}

	client := newTestClient(mockUnifiClient)

	validator := NewFeatureValidator(client)

	assert.NotNil(t, validator, "Validator should not be nil")
	featureValidator, ok := validator.(*featureEnabledValidator)
	assert.True(t, ok, "Validator should be of type *featureEnabledValidator")
	assert.Equal(t, client, featureValidator.client, "Client should be set correctly")
	assert.NotNil(t, featureValidator.cache, "Cache should be initialized")
}

// TestGetFeatures tests the getFeatures method of featureEnabledValidator
func TestGetFeatures(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *MockUnifiClient
		site     string
		expected Features
	}{
		{
			name: "successfully get features",
			setup: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return []unifi.DescribedFeature{
							{Name: "feature1", FeatureExists: true},
							{Name: "feature2", FeatureExists: false},
						}, nil
					},
				}
			},
			site: "site1",
			expected: Features{
				"feature1": featureEnabled,
				"feature2": featureDisabled,
			},
		},
		{
			name: "error getting features",
			setup: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return nil, errors.New("error listing features")
					},
				}
			},
			site:     "site2",
			expected: Features{}, // Now returns empty Features instead of nil
		},
		{
			name: "no features returned",
			setup: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return []unifi.DescribedFeature{}, nil
					},
				}
			},
			site:     "site3",
			expected: Features{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUnifiClient := tt.setup()
			client := newTestClient(mockUnifiClient)

			validator := &featureEnabledValidator{
				client: client,
				cache:  make(map[string]Features),
				lock:   sync.Mutex{},
			}

			result := validator.getFeatures(context.Background(), tt.site)
			assert.Equal(t, tt.expected, result)

			// Test caching - if we call again, we should get the cached result
			if tt.expected != nil {
				// Replace the test client with one that fails
				client.Client = &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return nil, errors.New("should not be called")
					},
				}
				result = validator.getFeatures(context.Background(), tt.site)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestGetFeaturesConcurrent tests the concurrency safety of getFeatures
func TestGetFeaturesConcurrent(t *testing.T) {
	callCount := 0
	mockUnifiClient := &MockUnifiClient{
		featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
			callCount++
			return []unifi.DescribedFeature{
				{Name: "feature1", FeatureExists: true},
				{Name: "feature2", FeatureExists: false},
			}, nil
		},
	}

	client := newTestClient(mockUnifiClient)

	validator := &featureEnabledValidator{
		client: client,
		cache:  make(map[string]Features),
		lock:   sync.Mutex{},
	}

	var wg sync.WaitGroup
	// Launch 10 concurrent goroutines to call getFeatures
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			features := validator.getFeatures(context.Background(), "site1")
			assert.NotNil(t, features)
			assert.True(t, features.IsEnabled("feature1"))
			assert.False(t, features.IsEnabled("feature2"))
		}()
	}
	wg.Wait()

	// Verify ListFeatures was called exactly once
	assert.Equal(t, 1, callCount, "ListFeatures should be called exactly once")
}

// TestRequireFeatures tests the requireFeatures method of featureEnabledValidator
func TestRequireFeatures(t *testing.T) {
	tests := []struct {
		name              string
		features          Features
		site              string
		attrPath          *path.Path
		requiredFeatures  []string
		expectedHasErrors bool
	}{
		{
			name:              "all features enabled",
			features:          Features{"feature1": featureEnabled, "feature2": featureEnabled},
			site:              "site1",
			attrPath:          nil,
			requiredFeatures:  []string{"feature1", "feature2"},
			expectedHasErrors: false,
		},
		{
			name:              "one feature disabled",
			features:          Features{"feature1": featureEnabled, "feature2": featureDisabled},
			site:              "site1",
			attrPath:          nil,
			requiredFeatures:  []string{"feature1", "feature2"},
			expectedHasErrors: true,
		},
		{
			name:              "all features disabled",
			features:          Features{"feature1": featureDisabled, "feature2": featureDisabled},
			site:              "site1",
			attrPath:          nil,
			requiredFeatures:  []string{"feature1", "feature2"},
			expectedHasErrors: true,
		},
		{
			name:              "empty required features",
			features:          Features{"feature1": featureEnabled, "feature2": featureEnabled},
			site:              "site1",
			attrPath:          nil,
			requiredFeatures:  []string{},
			expectedHasErrors: false,
		},
		{
			name:              "nil required features",
			features:          Features{"feature1": featureEnabled, "feature2": featureEnabled},
			site:              "site1",
			attrPath:          nil,
			requiredFeatures:  nil,
			expectedHasErrors: false,
		},
		{
			name:              "with attribute path",
			features:          Features{"feature1": featureDisabled},
			site:              "site1",
			attrPath:          &path.Path{},
			requiredFeatures:  []string{"feature1"},
			expectedHasErrors: true,
		},
		{
			name:              "feature not in map",
			features:          Features{"feature1": featureEnabled},
			site:              "site1",
			attrPath:          nil,
			requiredFeatures:  []string{"feature2"},
			expectedHasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUnifiClient := &MockUnifiClient{}
			client := newTestClient(mockUnifiClient)

			validator := &featureEnabledValidator{
				client: client,
				cache:  map[string]Features{tt.site: tt.features},
				lock:   sync.Mutex{},
			}

			diags := validator.requireFeatures(context.Background(), tt.site, tt.attrPath, tt.requiredFeatures...)
			assert.Equal(t, tt.expectedHasErrors, diags.HasError())

			if tt.expectedHasErrors {
				// Verify error message contains appropriate information
				assert.Contains(t, diags[0].Detail(), "Features", "Error detail should mention 'Features'")

				if tt.attrPath != nil {
					assert.Contains(t, diags[0].Detail(), "is not supported", "Error should mention path is not supported")
				}
			}
		})
	}
}

// TestRequireFeaturesEnabledForPath tests the RequireFeaturesEnabledForPath method
func TestRequireFeaturesEnabledForPath(t *testing.T) {
	tests := []struct {
		name              string
		setupClient       func() *MockUnifiClient
		attrValue         attr.Value
		attrPath          path.Path
		requiredFeatures  []string
		configError       bool
		expectedHasErrors bool
	}{
		{
			name: "attribute not set",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{}
			},
			attrValue:         types.StringNull(),
			attrPath:          path.Root("test"),
			requiredFeatures:  []string{"feature1"},
			configError:       false,
			expectedHasErrors: false,
		},
		{
			name: "error getting attribute",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{}
			},
			attrValue:         types.StringNull(),
			attrPath:          path.Root("test"),
			requiredFeatures:  []string{"feature1"},
			configError:       true,
			expectedHasErrors: true,
		},
		{
			name: "attribute set, feature enabled",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return []unifi.DescribedFeature{
							{Name: "feature1", FeatureExists: true},
						}, nil
					},
				}
			},
			attrValue:         types.StringValue("test"),
			attrPath:          path.Root("test"),
			requiredFeatures:  []string{"feature1"},
			configError:       false,
			expectedHasErrors: false,
		},
		{
			name: "attribute set, feature disabled",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return []unifi.DescribedFeature{
							{Name: "feature1", FeatureExists: false},
						}, nil
					},
				}
			},
			attrValue:         types.StringValue("test"),
			attrPath:          path.Root("test"),
			requiredFeatures:  []string{"feature1"},
			configError:       false,
			expectedHasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUnifiClient := tt.setupClient()
			client := newTestClient(mockUnifiClient)

			// Create a wrapper FeatureValidator that provides a minimal implementation
			// of RequireFeaturesEnabledForPath without needing a real tfsdk.Config
			validator := &testFeatureValidator{
				base:      NewFeatureValidator(client),
				attrValue: tt.attrValue,
				configErr: tt.configError,
			}

			diags := validator.TestRequireFeaturesEnabledForPath(context.Background(), "site1", tt.attrPath, tt.requiredFeatures...)

			assert.Equal(t, tt.expectedHasErrors, diags.HasError())
		})
	}
}

// TestRequireFeaturesEnabled tests the RequireFeaturesEnabled method
func TestRequireFeaturesEnabled(t *testing.T) {
	tests := []struct {
		name              string
		setupClient       func() *MockUnifiClient
		requiredFeatures  []string
		expectedHasErrors bool
	}{
		{
			name: "feature enabled",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return []unifi.DescribedFeature{
							{Name: "feature1", FeatureExists: true},
						}, nil
					},
				}
			},
			requiredFeatures:  []string{"feature1"},
			expectedHasErrors: false,
		},
		{
			name: "feature disabled",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return []unifi.DescribedFeature{
							{Name: "feature1", FeatureExists: false},
						}, nil
					},
				}
			},
			requiredFeatures:  []string{"feature1"},
			expectedHasErrors: true,
		},
		{
			name: "feature error",
			setupClient: func() *MockUnifiClient {
				return &MockUnifiClient{
					featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
						return nil, errors.New("error listing features")
					},
				}
			},
			requiredFeatures:  []string{"feature1"},
			expectedHasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUnifiClient := tt.setupClient()
			client := newTestClient(mockUnifiClient)
			validator := NewFeatureValidator(client)
			diags := validator.RequireFeaturesEnabled(context.Background(), "site1", tt.requiredFeatures...)
			assert.Equal(t, tt.expectedHasErrors, diags.HasError())
		})
	}
}

// TestIsDefined is used in RequireFeaturesEnabledForPath to check if a value is defined
func TestIsDefined(t *testing.T) {
	tests := []struct {
		name     string
		value    attr.Value
		expected bool
	}{
		{
			name:     "null",
			value:    types.StringNull(),
			expected: false,
		},
		{
			name:     "unknown",
			value:    types.StringUnknown(),
			expected: false,
		},
		{
			name:     "null list",
			value:    types.ListNull(types.StringType),
			expected: false,
		},
		{
			name:     "empty list",
			value:    types.ListValueMust(types.StringType, []attr.Value{}),
			expected: true,
		},
		{
			name:     "defined value",
			value:    types.StringValue("test"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsDefined(tt.value))
		})
	}
}

// TestFeatureValidatorCache specifically tests the caching behavior of the FeatureValidator
// It verifies that multiple calls with the same site only result in one API call
func TestFeatureValidatorCache(t *testing.T) {
	// Create a mock client with a counter for API calls
	callCount := 0
	mockUnifiClient := &MockUnifiClient{
		featuresFunc: func(ctx context.Context, site string) ([]unifi.DescribedFeature, error) {
			callCount++
			return []unifi.DescribedFeature{
				{Name: "feature1", FeatureExists: true},
				{Name: "feature2", FeatureExists: false},
			}, nil
		},
	}

	client := newTestClient(mockUnifiClient)

	validator := NewFeatureValidator(client)

	// First call to check features should trigger an API call
	diags1 := validator.RequireFeaturesEnabled(context.Background(), "site1", "feature1")
	assert.Equal(t, 1, callCount, "First call should trigger an API call")
	assert.False(t, diags1.HasError(), "Feature1 should be enabled")

	// Second call with the same site should use the cache
	diags2 := validator.RequireFeaturesEnabled(context.Background(), "site1", "feature2")
	assert.Equal(t, 1, callCount, "Second call should use cached data")
	assert.True(t, diags2.HasError(), "Feature2 should be disabled")

	// Call with a different site should trigger another API call
	diags3 := validator.RequireFeaturesEnabled(context.Background(), "site2", "feature1")
	assert.Equal(t, 2, callCount, "Call with different site should trigger an API call")
	assert.False(t, diags3.HasError(), "Feature1 should be enabled")

	// Multiple calls using the same site should still use the cache
	for i := 0; i < 5; i++ {
		validator.RequireFeaturesEnabled(context.Background(), "site1", "feature1")
	}
	assert.Equal(t, 2, callCount, "Multiple calls with same site should use cached data")
}

// testFeatureValidator wraps a real FeatureValidator but has a special method for testing
// that doesn't require a real tfsdk.Config
type testFeatureValidator struct {
	base      FeatureValidator
	attrValue attr.Value
	configErr bool
}

// TestRequireFeaturesEnabledForPath is a test-specific version that doesn't need a real tfsdk.Config
func (v *testFeatureValidator) TestRequireFeaturesEnabledForPath(ctx context.Context, site string,
	attrPath path.Path, features ...string) diag.Diagnostics {

	diags := diag.Diagnostics{}

	// This simulates what happens in RequireFeaturesEnabledForPath without needing a real Config
	if v.configErr {
		diags.AddError("Error", "Error getting attribute")
		return diags
	}

	if !IsDefined(v.attrValue) {
		return diags
	}

	// Call the underlying validator's RequireFeaturesEnabled
	fv, ok := v.base.(*featureEnabledValidator)
	if !ok {
		diags.AddError("Error", "Invalid validator type")
		return diags
	}

	diags.Append(fv.requireFeatures(ctx, site, &attrPath, features...)...)
	return diags
}
