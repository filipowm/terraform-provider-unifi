package utils

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ListToStringSlice(src []interface{}) ([]string, error) {
	dst := make([]string, 0, len(src))
	for _, s := range src {
		d, ok := s.(string)
		if !ok {
			return nil, fmt.Errorf("unale to convert %v (%T) to string", s, s)
		}
		dst = append(dst, d)
	}
	return dst, nil
}

func SetToStringSlice(src *schema.Set) ([]string, error) {
	return ListToStringSlice(src.List())
}

func StringSliceToList(list []string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}

func StringSliceToSet(src []string) *schema.Set {
	return schema.NewSet(schema.HashString, StringSliceToList(src))
}

func IsStringValueNotEmpty(s basetypes.StringValue) bool {
	return !s.IsUnknown() && !s.IsNull() && s.ValueString() != ""
}

// JoinNonEmpty joins non-empty strings from a slice with the specified separator.
// Empty strings in the slice are filtered out.
func JoinNonEmpty(elements []string, separator string) string {
	var nonEmpty []string
	for _, elem := range elements {
		if elem != "" {
			nonEmpty = append(nonEmpty, elem)
		}
	}
	return strings.Join(nonEmpty, separator)
}

// SplitAndTrim splits a string by the specified separator and trims whitespace from each element.
// Empty strings after trimming are filtered out.
func SplitAndTrim(s string, separator string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, separator)
	var result []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func RemoveElements[S ~[]E, E comparable](first S, second S) S {
	var result S
	for _, category := range first {
		if !slices.Contains(second, category) {
			result = append(result, category)
		}
	}
	return result
}
