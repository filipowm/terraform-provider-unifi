package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldInfoFromValidation(t *testing.T) {
	t.Parallel()

	for i, c := range []struct {
		expectedType      string
		expectedComment   string
		expectedOmitEmpty bool
		validation        interface{}
	}{
		{"string", "", true, ""},
		{"string", "default|custom", true, "default|custom"},
		{"string", ".{0,32}", true, ".{0,32}"},
		{"string", "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^$", false, "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^$"},

		{"int", "^([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^$", true, "^([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^$"},
		{"int", "", true, "^[0-9]*$"},

		{"float64", "", true, "[-+]?[0-9]*\\.?[0-9]+"},
		// this one is really an error as the . is not escaped
		{"float64", "", true, "^([-]?[\\d]+[.]?[\\d]*)$"},
		{"float64", "", true, "^([\\d]+[.]?[\\d]*)$"},

		{"bool", "", false, "false|true"},
		{"bool", "", false, "true|false"},
	} {
		t.Run(fmt.Sprintf("%d %s %s", i, c.expectedType, c.validation), func(t *testing.T) {
			t.Parallel()

			resource := &Resource{
				StructName:     "TestType",
				Types:          make(map[string]*FieldInfo),
				FieldProcessor: func(name string, f *FieldInfo) error { return nil },
			}

			fieldInfo, err := resource.fieldInfoFromValidation("fieldName", c.validation)
			// actualType, actualComment, actualOmitEmpty, err := fieldInfoFromValidation(c.validation)
			if err != nil {
				t.Fatal(err)
			}
			if fieldInfo.FieldType != c.expectedType {
				t.Fatalf("expected type %q got %q", c.expectedType, fieldInfo.FieldType)
			}
			if fieldInfo.FieldValidationComment != c.expectedComment {
				t.Fatalf("expected comment %q got %q", c.expectedComment, fieldInfo.FieldValidationComment)
			}
			if fieldInfo.OmitEmpty != c.expectedOmitEmpty {
				t.Fatalf("expected omitempty %t got %t", c.expectedOmitEmpty, fieldInfo.OmitEmpty)
			}
		})
	}
}

func TestResourceTypes(t *testing.T) {
	t.Parallel()

	testData := `
{
  "note": ".{0,1024}",
  "date": "^$|^(20[0-9]{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9])Z?$",
  "mac": "^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$",
  "number": "\\d+",
  "boolean": "true|false",
	"nested_type": {
    "nested_field": "^$"
  },
  "nested_type_array": [{
    "nested_field": "^$"
  }]
}
	`
	expectedFields := map[string]*FieldInfo{
		"Note":    NewFieldInfo("Note", "note", "string", "validate:\"omitempty,gte=0,lte=1024\"", ".{0,1024}", true, false, ""),
		"Date":    NewFieldInfo("Date", "date", "string", "", "^$|^(20[0-9]{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9])Z?$", false, false, ""),
		"MAC":     NewFieldInfo("MAC", "mac", "string", "validate:\"omitempty,mac\"", "^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$", true, false, ""),
		"Number":  NewFieldInfo("Number", "number", "int", "", "", true, false, "emptyStringInt"),
		"Boolean": NewFieldInfo("Boolean", "boolean", "bool", "", "", false, false, ""),
		"NestedType": {
			FieldName:              "NestedType",
			JSONName:               "nested_type",
			FieldType:              "StructNestedType",
			FieldValidationComment: "",
			OmitEmpty:              true,
			IsArray:                false,
			Fields: map[string]*FieldInfo{
				"NestedFieldModified": NewFieldInfo("NestedFieldModified", "nested_field", "string", "", "^$", false, false, ""),
			},
		},
		"NestedTypeArray": {
			FieldName:              "NestedTypeArray",
			JSONName:               "nested_type_array",
			FieldType:              "StructNestedTypeArray",
			FieldValidationComment: "",
			OmitEmpty:              true,
			IsArray:                true,
			Fields: map[string]*FieldInfo{
				"NestedFieldModified": NewFieldInfo("NestedFieldModified", "nested_field", "string", "", "^$", false, false, ""),
			},
		},
	}

	expectedStruct := map[string]*FieldInfo{
		"Struct": {
			FieldName:              "Struct",
			JSONName:               "path",
			FieldType:              "struct",
			FieldValidationComment: "",
			OmitEmpty:              false,
			IsArray:                false,
			Fields: map[string]*FieldInfo{
				"   ID":      NewFieldInfo("ID", "_id", "string", "", "", true, false, ""),
				"   SiteID":  NewFieldInfo("SiteID", "site_id", "string", "", "", true, false, ""),
				"   _Spacer": nil,
				"  Hidden":   NewFieldInfo("Hidden", "attr_hidden", "bool", "", "", true, false, ""),
				"  HiddenID": NewFieldInfo("HiddenID", "attr_hidden_id", "string", "", "", true, false, ""),
				"  NoDelete": NewFieldInfo("NoDelete", "attr_no_delete", "bool", "", "", true, false, ""),
				"  NoEdit":   NewFieldInfo("NoEdit", "attr_no_edit", "bool", "", "", true, false, ""),
				"  _Spacer":  nil,
				" _Spacer":   nil,
			},
		},
	}

	for k, v := range expectedFields {
		expectedStruct["Struct"].Fields[k] = v
	}

	expectation := &Resource{
		StructName:   "Struct",
		ResourcePath: "path",

		Types: map[string]*FieldInfo{
			"Struct":                expectedStruct["Struct"],
			"StructNestedType":      expectedStruct["Struct"].Fields["NestedType"],
			"StructNestedTypeArray": expectedStruct["Struct"].Fields["NestedTypeArray"],
		},

		FieldProcessor: func(name string, f *FieldInfo) error {
			if name == "NestedField" {
				f.FieldName = "NestedFieldModified"
			}
			return nil
		},
	}

	t.Run("structural test", func(t *testing.T) {
		t.Parallel()

		resource := NewResource("Struct", "path")
		resource.FieldProcessor = expectation.FieldProcessor

		err := resource.processJSON(([]byte)(testData))

		require.NoError(t, err, "No error processing JSON")
		assert.Equal(t, expectation.StructName, resource.StructName)
		assert.Equal(t, expectation.ResourcePath, resource.ResourcePath)
		assert.Equal(t, expectation.Types, resource.Types)
	})
}

func TestNormalizeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"\\d+", "09"},
		{"[-+]?[0-9]*\\.?[0-9]+", "09.09"},
		{"^([0-9]|[1-9][0-9]|25[0-5])$", "0919092505"},
		{"^(([0-9]\\.[0-9]{2})\\.){3}([0-9]\\.[0-9])$", "09.09.09.09"},
		{"[+-]?[0-9]*\\.?[0-9]+", "09.09"},
		{"[-]?[\\d]+[.]?[\\d]*", "09.09"},
		{"^$|^(20[0-9]{2}T([01][0-9]):[1-5]:[0-9])Z?$", "2009T0109:15:09Z"},
		{"false|true", "falsetrue"},
		{"true|false", "truefalse"},
		{".{0,32}", "."},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			actual := normalizeValidation(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

var testReps = []replacement{
	{"dhcpd", "DHCPD"},
	{"ip", "IP"},
}

func TestCleanName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		reps     []replacement
		expected string
	}{
		{"field replacements basic", "dhcpd_enabled", testReps, "DHCPD_enabled"},
		{"field replacements multiple", "dhcpd_ip_mac", testReps, "DHCPD_IP_mac"},
		{"field replacements no match", "something_else", testReps, "something_else"},
		{"empty string", "", fieldReps, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			actual := cleanName(tc.input, tc.reps)
			a.Equal(tc.expected, actual)
		})
	}
}

func TestIsSetting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		structName string
		expected   bool
	}{
		{"Setting", true},
		{"SettingUsg", true},
		{"SettingGlobalAp", true},
		{"Settings", true},
		{"Device", false},
		{"Network", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.structName, func(t *testing.T) {
			t.Parallel()
			resource := &Resource{StructName: tc.structName}
			assert.Equal(t, tc.expected, resource.IsSetting())
		})
	}
}

func TestFieldInfoFromValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldName     string
		validation    interface{}
		errorContains string
	}{
		{
			"invalid validation type",
			"field",
			123,
			"unable to determine type from validation",
		},
		{
			"empty array",
			"field",
			[]interface{}{},
			"",
		},
		{
			"array with multiple items",
			"field",
			[]interface{}{"item1", "item2"},
			"unknown validation",
		},
		{
			"invalid nested validation",
			"field",
			map[string]interface{}{
				"nested": 123,
			},
			"unable to determine type from validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			resource := NewResource("Test", "test")
			fieldInfo, err := resource.fieldInfoFromValidation(tc.fieldName, tc.validation)
			if tc.errorContains != "" {
				require.ErrorContains(t, err, tc.errorContains)
				a.NotNil(fieldInfo)
				a.EqualValues(&FieldInfo{}, fieldInfo)
			} else {
				require.NoError(t, err)
				a.NotNil(fieldInfo)
			}
		})
	}
}

func TestBuildResourcesFromDownloadedFields(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test JSON files
	validJSON := `{
		"name": "test",
		"value": "^[0-9]*$",
		"enabled": "true|false"
	}`
	err := os.WriteFile(filepath.Join(tmpDir, "Test.json"), []byte(validJSON), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "Invalid.json"), []byte("invalid json"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "Setting.json"), []byte(validJSON), 0o644)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		dir           string
		expectedLen   int
		errorContains string
	}{
		{
			"valid directory",
			tmpDir,
			1, // Only Test.json should be processed (Setting.json is skipped, Invalid.json fails)
			"",
		},
		{
			"non-existent directory",
			"non-existent",
			0,
			"unable to read fields directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			resources, err := buildResourcesFromDownloadedFields(tc.dir, CodeCustomizer{}, false)
			if tc.errorContains != "" {
				require.ErrorContains(t, err, tc.errorContains)
				a.Nil(resources)
			} else {
				require.NoError(t, err)
				a.Len(resources, tc.expectedLen)
				if tc.expectedLen > 0 {
					a.Equal("Test", resources[0].StructName)
					a.Equal("test", resources[0].ResourcePath)
				}
			}
		})
	}
}
