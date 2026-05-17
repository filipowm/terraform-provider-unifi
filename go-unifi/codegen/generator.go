package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/iancoleman/strcase"
)

// Generatable is the interface for generation sources.
type Generatable interface {
	Name() string
	GenerateCode() (string, error)
}

// generateCodeFromTemplate renders a template with provided content and formats the source.
func generateCodeFromTemplate(templateName, templateContent string, toWrite any) (string, error) {
	var err error
	var buf bytes.Buffer
	writer := io.Writer(&buf)

	tpl := template.Must(template.New(templateName).Parse(templateContent))

	err = tpl.Execute(writer, toWrite)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format source: %w", err)
	}

	return string(src), err
}

// generateCode generates code for each generation source and writes it to file.
func generateCode(fieldsDir, outDir string, customizer CodeCustomizer) error {
	if _, err := ensurePath(outDir); err != nil {
		return fmt.Errorf("unable to create output directory %s: %w", outDir, err)
	}

	generators := make([]Generatable, 0)
	resources, err := buildResourcesFromDownloadedFields(fieldsDir, customizer, false)
	if err != nil {
		return fmt.Errorf("failed to build resources from downloaded fields: %w", err)
	}

	codegenPath, err := findCodegenDir()
	if err != nil {
		return fmt.Errorf("failed to find codegen directory: %w", err)
	}
	resourcesCustomV2, err := buildCustomResources(filepath.Join(codegenPath, "v2"), customizer, true)
	if err != nil {
		return fmt.Errorf("failed to build resources from downloaded fields: %w", err)
	}
	resources = append(resources, resourcesCustomV2...)
	cb := NewClientInfoBuilder()
	customizer.ApplyToClient(cb)
	for _, resource := range resources {
		if !customizer.IsExcludedFromClient(resource.Name()) {
			cb.AddResource(resource)
		}
		customizer.ApplyToResource(resource)
		generators = append(generators, resource)
	}
	generators = append(generators, cb.Build())

	for _, g := range generators {
		var code string
		if code, err = g.GenerateCode(); err != nil {
			log.Errorf("failed to generate code for %s: %s", g.Name(), err)
			continue
		}

		filename, err := writeGeneratedFile(outDir, g.Name(), code)
		if err != nil {
			log.Errorf("failed to write file %s: %s", g.Name(), err)
			continue
		}
		log.Debugf("Generated %s with resource %s\n\n", filename, g.Name())
	}
	return nil
}

// writeGeneratedFile writes generated file content to a file.
func writeGeneratedFile(outDir string, name string, content string) (string, error) {
	goFile := strcase.ToSnake(name) + ".generated.go"
	goFilePath := filepath.Join(outDir, goFile)
	_ = os.Remove(goFilePath)
	if err := os.WriteFile(goFilePath, []byte(content), 0o644); err != nil {
		return goFile, fmt.Errorf("failed to write file %s: %w", goFile, err)
	}
	return goFile, nil
}
