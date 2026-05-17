package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomClientFunctionSignature(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		fn          CustomClientFunction
		wantComment string // expected comment in the signature
		wantFunc    string // expected function signature
	}{
		{
			name: "no comment, no params, no returns",
			fn: CustomClientFunction{
				FunctionName: "Foo",
			},
			wantFunc: "Foo()",
		},
		{
			name: "with comment, no params, no returns",
			fn: CustomClientFunction{
				FunctionName: "Bar",
			},
			wantFunc: "Bar()",
		},
		{
			name: "with one param and one return",
			fn: CustomClientFunction{
				FunctionName:     "Baz",
				Parameters:       []FunctionParam{{"a", "int"}},
				ReturnParameters: []string{"error"},
			},
			wantFunc: "Baz(a int) error",
		},
		{
			name: "with multiple returns",
			fn: CustomClientFunction{
				FunctionName:     "Qux",
				Parameters:       []FunctionParam{{"x", "string"}},
				ReturnParameters: []string{"int", "error"},
			},
			wantFunc: "Qux(x string) (int, error)",
		},
		{
			name: "with multiple params",
			fn: CustomClientFunction{
				FunctionName:     "MultiParams",
				Parameters:       []FunctionParam{{"x", "string"}, {"y", "int"}},
				ReturnParameters: []string{},
			},
			wantFunc: "MultiParams(x string, y int)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			a.Equal(tt.wantFunc, tt.fn.Signature())
		})
	}
}

func TestGenerateCode(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	b := NewClientInfoBuilder()
	b.AddImport("fmt")
	b.AddFunction(&CustomClientFunction{
		FunctionName:     "TestFunc",
		Parameters:       []FunctionParam{{"x", "int"}},
		ReturnParameters: []string{"error"},
		FunctionComment:  "This is a test function",
	})
	ci := b.Build()
	code, err := ci.GenerateCode()
	require.NoError(t, err)
	a.NotEmpty(code, "GenerateCode() returned empty code")
	a.Contains(code, "TestFunc")
}
