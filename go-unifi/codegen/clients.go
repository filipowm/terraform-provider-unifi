package main

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
)

// ClientFunction is the interface for client functions.
type ClientFunction interface {
	Name() string
	ResourceName() string
	Comment() string
	Signature() string
}

type FunctionParam struct {
	Name string
	Type string
}

type Comment struct {
	comment      string
	resourceName string
}

func (c *Comment) Name() string {
	return ""
}

func (c *Comment) Comment() string {
	return c.comment
}

func (c *Comment) Signature() string {
	return ""
}

func (c *Comment) ResourceName() string {
	return c.resourceName
}

// CustomClientFunction represents a custom client function definition.
type CustomClientFunction struct {
	Resource         string          `yaml:"resourceName"`
	FunctionName     string          `yaml:"name"`
	Parameters       []FunctionParam `yaml:"params"`
	ReturnParameters []string        `yaml:"returns"`
	FunctionComment  string          `yaml:"comment"`
}

func (c *CustomClientFunction) Name() string {
	return c.FunctionName
}

func (c *CustomClientFunction) ResourceName() string {
	return c.Resource
}

// Signature returns the signature string for the custom client function.
func (c *CustomClientFunction) Signature() string {
	if c.Name() == "" {
		return ""
	}
	var b strings.Builder
	//if c.comment != "" {
	//	b.WriteString(fmt.Sprintf("// %s %s\n", c.Name, c.Comment))
	//}
	b.WriteString(c.Name())
	b.WriteString("(")

	// Build parameters without trailing comma
	params := make([]string, 0, len(c.Parameters))
	for _, v := range c.Parameters {
		params = append(params, fmt.Sprintf("%s %s", v.Name, v.Type))
	}
	b.WriteString(strings.Join(params, ", "))
	b.WriteString(")")

	if len(c.ReturnParameters) > 1 {
		b.WriteString(" (")
		b.WriteString(strings.Join(c.ReturnParameters, ", "))
		b.WriteString(")")
	} else if len(c.ReturnParameters) == 1 {
		b.WriteString(" " + c.ReturnParameters[0])
	}
	return b.String()
}

func (c *CustomClientFunction) Comment() string {
	return c.FunctionComment
}

// ClientInfo represents the client information used for code generation.
type ClientInfo struct {
	Imports   []string
	Functions []ClientFunction
}

type ClientInfoBuilder struct {
	imports   []string
	functions []ClientFunction
}

func (c *ClientInfoBuilder) AddFunction(f ClientFunction) *ClientInfoBuilder { //nolint: unparam
	c.functions = append(c.functions, f)
	return c
}

func (c *ClientInfoBuilder) AddFunctions(f []CustomClientFunction) *ClientInfoBuilder {
	for _, v := range f {
		c.functions = append(c.functions, &v)
	}
	return c
}

func (c *ClientInfoBuilder) addResourceFunction(actionName, resourceName, comment string, additionalParams []FunctionParam, additionalReturns []string) {
	fName := fmt.Sprintf("%s%s", actionName, resourceName)
	params := []FunctionParam{
		{"ctx", "context.Context"},
		{"site", "string"},
	}
	params = append(params, additionalParams...)
	returns := additionalReturns
	returns = append(returns, "error")
	f := CustomClientFunction{
		FunctionName:     fName,
		Resource:         resourceName,
		Parameters:       params,
		ReturnParameters: returns,
		FunctionComment:  fmt.Sprintf("%s %s", fName, comment),
	}
	c.AddFunction(&f)
}

func singlePointerReturn(name string) []string {
	return []string{"*" + name}
}

func singlePointerParam(name string) []FunctionParam {
	return []FunctionParam{{strings.ToLower(name[0:1]), "*" + name}}
}

func (c *ClientInfoBuilder) AddResource(r *Resource) *ClientInfoBuilder {
	c.AddFunction(&Comment{comment: fmt.Sprintf("==== client methods for %s resource ====", r.Name()), resourceName: r.Name()})
	if r.IsSetting() {
		c.addResourceFunction("Get", r.Name(), "retrieves the settings for a resource", nil, singlePointerReturn(r.Name()))
		c.addResourceFunction("Update", r.Name(), "updates a resource", singlePointerParam(r.Name()), singlePointerReturn(r.Name()))
		return c
	}
	c.addResourceFunction("Get", r.Name(), "retrieves a resource", []FunctionParam{{"id", "string"}}, singlePointerReturn(r.Name()))
	c.addResourceFunction("List", r.Name(), "lists the resources", nil, []string{"[]" + r.Name()})
	c.addResourceFunction("Create", r.Name(), "creates a resource", singlePointerParam(r.Name()), singlePointerReturn(r.Name()))
	c.addResourceFunction("Update", r.Name(), "updates a resource", singlePointerParam(r.Name()), singlePointerReturn(r.Name()))
	c.addResourceFunction("Delete", r.Name(), "deletes a resource", []FunctionParam{{"id", "string"}}, nil)
	c.AddFunction(&Comment{comment: fmt.Sprintf("==== end of client methods for %s resource ====", r.Name()), resourceName: r.Name() + "_end"})
	return c
}

func (c *ClientInfoBuilder) AddImport(i string) *ClientInfoBuilder {
	c.imports = append(c.imports, i)
	return c
}

func (c *ClientInfoBuilder) AddImports(i []string) *ClientInfoBuilder {
	c.imports = append(c.imports, i...)
	return c
}

func (c *ClientInfoBuilder) Build() *ClientInfo {
	// Sort the functions by resource name and then by name.
	sort.Slice(c.functions, func(i, j int) bool {
		if c.functions[i].ResourceName() == c.functions[j].ResourceName() {
			return c.functions[i].Signature() < c.functions[j].Signature()
		}
		return c.functions[i].ResourceName() < c.functions[j].ResourceName()
	})

	return newClientInfo(c.imports, c.functions)
}

func NewClientInfoBuilder() *ClientInfoBuilder {
	return &ClientInfoBuilder{}
}

// newClientInfo creates ClientInfo from the provided resources.
func newClientInfo(imports []string, functions []ClientFunction) *ClientInfo {
	return &ClientInfo{imports, functions}
}

//go:embed client.go.tmpl
var clientGoTemplate string

// GenerateCode generates the code for the client using a template.
func (c *ClientInfo) GenerateCode() (string, error) {
	return generateCodeFromTemplate("client.go.tmpl", clientGoTemplate, c)
}

// Name returns the name of the client.
func (c *ClientInfo) Name() string {
	return "Client"
}
