package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCodeFromTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		templateName  string
		template      string
		data          interface{}
		expectedCode  string
		expectedError bool
		errorContains string
	}{
		{
			name:         "valid template",
			templateName: "simple",
			template: `package main

const greeting = "{{.Greeting}}"`,
			data:         struct{ Greeting string }{Greeting: "hello"},
			expectedCode: "const greeting = \"hello\"",
		},
		{
			name:          "invalid go code output",
			templateName:  "invalid_code",
			template:      `not valid {{ .Value }} go code`,
			data:          struct{ Value string }{Value: "test"},
			expectedError: true,
			errorContains: "failed to format source",
		},
		{
			name:         "no data",
			templateName: "nil_data",
			template:     `package main`,
			data:         nil,
			expectedCode: "package main",
		},
		{
			name:         "complex template",
			templateName: "complex",
			template: `package main

type {{.TypeName}} struct {
	{{range .Fields}}
	{{.Name}} {{.Type}}
	{{end}}
}`,
			data: struct {
				TypeName string
				Fields   []struct{ Name, Type string }
			}{
				TypeName: "Person",
				Fields: []struct{ Name, Type string }{
					{Name: "Name", Type: "string"},
					{Name: "Age", Type: "int"},
				},
			},
			expectedCode: "type Person struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			code, err := generateCodeFromTemplate(tt.templateName, tt.template, tt.data)

			if tt.expectedError {
				require.ErrorContains(t, err, tt.errorContains)
			} else {
				require.NoError(t, err)
			}
			a.Contains(code, tt.expectedCode)
		})
	}
}

func TestWriteGeneratedFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		fileName         string
		content          string
		expectedFileName string
		expectError      bool
	}{
		{
			name:             "valid file",
			fileName:         "TestFile",
			content:          "package main\n\n// Code content",
			expectedFileName: "test_file.generated.go",
			expectError:      false,
		},
		{
			name:             "empty content",
			fileName:         "EmptyFile",
			content:          "",
			expectedFileName: "empty_file.generated.go",
			expectError:      true,
		},
		{
			name:             "file with spaces",
			fileName:         "Test File",
			content:          "package main",
			expectedFileName: "test_file.generated.go",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			tempDir := t.TempDir()

			fileName, err := writeGeneratedFile(tempDir, tt.fileName, tt.content)
			require.NoError(t, err)
			a.Equal(tt.expectedFileName, fileName)

			expectedFile := filepath.Join(tempDir, tt.expectedFileName)
			dataBytes, err := os.ReadFile(expectedFile)
			require.NoError(t, err)
			a.Equal(tt.content, string(dataBytes))
		})
	}
}

func TestWriteGeneratedFile_OverrideExistingFile(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tempDir := t.TempDir()
	fileName := "test"

	_, err := writeGeneratedFile(tempDir, fileName, "starting content")
	require.NoError(t, err)

	_, err = writeGeneratedFile(tempDir, fileName, "updated content")
	require.NoError(t, err)

	expectedFile := filepath.Join(tempDir, "test.generated.go")
	dataBytes, err := os.ReadFile(expectedFile)
	require.NoError(t, err)
	a.Equal("updated content", string(dataBytes))
}

func TestWriteGeneratedFile_InvalidPath(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	invalidDir := filepath.Join(tempDir, "nonexistent")

	_, err := writeGeneratedFile(invalidDir, "test", "content")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write file")
}

func TestGenerateCodeFromFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		fieldsDir      string
		outDir         string
		expectedError  bool
		errorContains  string
		setupMockFiles func(string)
	}{
		{
			name:          "invalid fields directory",
			fieldsDir:     "nonexistent",
			outDir:        t.TempDir(),
			expectedError: true,
			errorContains: "failed to build resources from downloaded fields",
		},
		{
			name:      "valid empty fields directory",
			fieldsDir: t.TempDir(),
			outDir:    t.TempDir(),
			setupMockFiles: func(dir string) {
				// Create empty directory structure
				_ = os.MkdirAll(dir, 0o755)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.setupMockFiles != nil {
				tt.setupMockFiles(tt.fieldsDir)
			}

			err := generateCode(tt.fieldsDir, tt.outDir, CodeCustomizer{})

			if tt.expectedError {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
