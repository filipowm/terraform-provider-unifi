package utils

import (
	"fmt"

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
