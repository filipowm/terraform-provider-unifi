package utils

import (
	"errors"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"reflect"
	"strings"
)

func ErrorInvalidModelMergeTarget(expectedType, actualType interface{}) diag.Diagnostic {
	e := reflect.TypeOf(&expectedType).Elem().String()
	a := reflect.TypeOf(&actualType).Elem().String()
	return diag.NewErrorDiagnostic("Invalid model merge target", "Expected target type to be the same a receiver: "+e+". Was : "+a)
}

func IsServerErrorStatusCode(err error, statusCode int) bool {
	if err == nil {
		return false
	}
	var se *unifi.ServerError
	if errors.As(err, &se) {
		return se.StatusCode == statusCode
	}
	return false
}

func IsServerErrorContains(err error, messageContains string) bool {
	if err == nil {
		return false
	}
	var se *unifi.ServerError
	if errors.As(err, &se) {
		if strings.Contains(se.Message, messageContains) {
			return true
		}
		// check details
		if se.Details != nil {
			for _, m := range se.Details {
				if strings.Contains(m.Message, messageContains) {
					return true
				}
			}
		}
	}
	return false
}
