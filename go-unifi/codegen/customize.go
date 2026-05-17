package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	AllFieldsCustomizationKeyword = "_all"
	defaultCustomizationsPath     = "customizations.yml"
)

type Customizations struct {
	Resources map[string]*ResourceCustomization `yaml:"resources"`
	Client    *ClientCustomization              `yaml:"client"`
}

type Generate struct {
	Customizations *Customizations `yaml:"customizations"`
}

type ResourceCustomization struct {
	ResourceName string                         `yaml:"-"`
	Fields       map[string]*FieldCustomization `yaml:"fields"`
	ResourcePath string                         `yaml:"resourcePath"`
}

type ClientCustomization struct {
	Imports          []string               `yaml:"imports"`
	Functions        []CustomClientFunction `yaml:"functions"`
	ExcludeResources []string               `yaml:"excludeResources"`
}

type FieldCustomization struct {
	FieldName   string             `yaml:"-"`
	Overrides   *FieldInfoOverride `yaml:",inline"`
	IfFieldType string             `yaml:"ifFieldType"`
}

type FieldInfoOverride struct {
	FieldName           *string `yaml:"fieldName"`
	FieldType           *string `yaml:"fieldType"`
	OmitEmpty           *bool   `yaml:"omitEmpty"`
	CustomUnmarshalType *string `yaml:"customUnmarshalType"`
	CustomUnmarshalFunc *string `yaml:"customUnmarshalFunc"`
	JsonPath            *string `yaml:"jsonPath"`
}

func compositeCustomizationsProcessor(customizationsProcessor FieldProcessor) FieldProcessor {
	return func(name string, f *FieldInfo) error {
		err := customizationsProcessor(AllFieldsCustomizationKeyword, f)
		if err != nil {
			return fmt.Errorf("failed applying all fields customization to %s field: %w", name, err)
		}
		err = customizationsProcessor(name, f)
		if err != nil {
			return fmt.Errorf("failed applying customization to %s fields: %w", name, err)
		}
		return nil
	}
}

func (r *ResourceCustomization) ApplyTo(resource *Resource) {
	if resource.StructName == r.ResourceName {
		currentProcessor := resource.FieldProcessor
		customizationsProcessor := r.toFieldProcessor()
		if currentProcessor != nil {
			// create composite processor with existing processor, first running pre-defined customizations, then user-defined
			resource.FieldProcessor = func(name string, f *FieldInfo) error {
				err := compositeCustomizationsProcessor(customizationsProcessor)(name, f)
				if err != nil {
					return err
				}
				return currentProcessor(name, f)
			}
			if r.ResourcePath != "" {
				resource.ResourcePath = r.ResourcePath
			}
		} else {
			resource.FieldProcessor = compositeCustomizationsProcessor(customizationsProcessor)
		}
	}
}

func (r *ResourceCustomization) toFieldProcessor() FieldProcessor {
	return func(name string, f *FieldInfo) error {
		if fc, ok := r.Fields[name]; ok && fc.Overrides != nil && (fc.IfFieldType == "" || fc.IfFieldType == f.FieldType) {
			if fc.Overrides.FieldType != nil {
				f.FieldType = *fc.Overrides.FieldType
			}
			if fc.Overrides.CustomUnmarshalType != nil {
				f.CustomUnmarshalType = *fc.Overrides.CustomUnmarshalType
			}
			if fc.Overrides.OmitEmpty != nil {
				f.OmitEmpty = *fc.Overrides.OmitEmpty
			}
			if fc.Overrides.CustomUnmarshalFunc != nil {
				f.CustomUnmarshalFunc = *fc.Overrides.CustomUnmarshalFunc
			}
			if fc.Overrides.FieldName != nil {
				f.FieldName = *fc.Overrides.FieldName
			}
			if fc.Overrides.JsonPath != nil {
				f.JSONName = *fc.Overrides.JsonPath
			}
		}
		return nil
	}
}

//go:embed customizations.yml
var defaultCustomizationYml []byte

func readCustomizationsYml(customizationsPath string) ([]byte, error) {
	if customizationsPath == "" || customizationsPath == defaultCustomizationsPath {
		return defaultCustomizationYml, nil
	}
	customizations, err := os.ReadFile(customizationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading customizations file %s: %w", customizationsPath, err)
	}
	return customizations, nil
}

func unmarshalCustomizationYaml(customizationsPath string) (*Generate, error) {
	var generate Generate
	customizationsYml, err := readCustomizationsYml(customizationsPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(customizationsYml, &generate) //nolint: musttag
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling YAML to Generate structure: %w", err)
	}
	// Assign ResourceName and FieldName based on the map keys
	for resourceName, resource := range generate.Customizations.Resources {
		resource.ResourceName = resourceName
		for fieldName, field := range resource.Fields {
			field.FieldName = fieldName
		}
	}

	return &generate, nil
}

type CodeCustomizer struct {
	Customizations Customizations
}

func NewCodeCustomizer(customizationsPath string) (*CodeCustomizer, error) {
	generate, err := unmarshalCustomizationYaml(customizationsPath)
	if err != nil {
		return nil, err
	}
	if generate.Customizations == nil {
		generate.Customizations = &Customizations{}
	}
	return &CodeCustomizer{*generate.Customizations}, nil
}

func (r *CodeCustomizer) IsExcludedFromClient(resourceName string) bool {
	if r.Customizations.Client == nil || r.Customizations.Client.ExcludeResources == nil {
		return false
	}
	for _, excludedResource := range r.Customizations.Client.ExcludeResources {
		prefixedAll := strings.HasPrefix(excludedResource, "*")
		suffixedAll := strings.HasSuffix(excludedResource, "*")
		if prefixedAll && suffixedAll && strings.Contains(resourceName, excludedResource[1:len(excludedResource)-1]) {
			return true
		} else if prefixedAll && strings.HasSuffix(resourceName, excludedResource[1:]) {
			return true
		} else if suffixedAll && strings.HasPrefix(resourceName, excludedResource[:len(excludedResource)-1]) {
			return true
		} else if resourceName == excludedResource {
			return true
		}
	}
	return false
}

func (r *CodeCustomizer) ApplyToResource(resource *Resource) {
	for resourceName, resourceCustomization := range r.Customizations.Resources {
		if resource.StructName == resourceName {
			resourceCustomization.ApplyTo(resource)
		}
	}
}

func (r *CodeCustomizer) ApplyToClient(client *ClientInfoBuilder) {
	if client == nil || r.Customizations.Client == nil {
		return
	}
	client.AddFunctions(r.Customizations.Client.Functions)
	client.AddImports(r.Customizations.Client.Imports)
}
