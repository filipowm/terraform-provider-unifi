package base

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"reflect"
)

func ErrorInvalidModelMergeTarget(expectedType, actualType interface{}) diag.Diagnostic {
	e := reflect.TypeOf(&expectedType).Elem().String()
	a := reflect.TypeOf(&actualType).Elem().String()
	return diag.NewErrorDiagnostic("Invalid model merge target", "Expected target type to be the same a receiver: "+e+". Was : "+a)
}
