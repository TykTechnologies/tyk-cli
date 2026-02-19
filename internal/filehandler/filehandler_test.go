package filehandler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data
var sampleOAS = map[string]interface{}{
	"openapi": "3.0.0",
	"info": map[string]interface{}{
		"title":   "Test API",
		"version": "1.0.0",
	},
	"paths": map[string]interface{}{
		"/test": map[string]interface{}{
			"get": map[string]interface{}{
				"summary": "Test endpoint",
			},
		},
	},
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{"valid json file", "test.json", false},
		{"valid yaml file", "test.yaml", false},
		{"valid yml file", "test.yml", false},
		{"empty path", "", true},
		{"unsupported extension", "test.txt", true},
		{"no extension", "test", true},
		{"case insensitive", "test.JSON", false},
		{"case insensitive yaml", "test.YAML", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.filePath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	tests := []struct {
		filePath    string
		expected    FileType
		expectError bool
	}{
		{"test.json", FileTypeJSON, false},
		{"test.yaml", FileTypeYAML, false},
		{"test.yml", FileTypeYAML, false},
		{"test.JSON", FileTypeJSON, false},
		{"test.YAML", FileTypeYAML, false},
		{"test.txt", 0, true},
		{"test", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result, err := getFileType(tt.filePath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoadFile_JSON(t *testing.T) {
	// Create temporary JSON file
	tmpDir, err := os.MkdirTemp("", "tyk-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonContent := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		}
	}`

	err = os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	// Test loading
	fileInfo, err := LoadFile(jsonFile)
	require.NoError(t, err)

	assert.Equal(t, jsonFile, fileInfo.Path)
	assert.Equal(t, FileTypeJSON, fileInfo.Type)
	assert.Equal(t, "3.0.0", fileInfo.Content["openapi"])
	
	info, ok := fileInfo.Content["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test API", info["title"])
	assert.Equal(t, "1.0.0", info["version"])
}

func TestLoadFile_YAML(t *testing.T) {
	// Create temporary YAML file
	tmpDir, err := os.MkdirTemp("", "tyk-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	yamlFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `
openapi: "3.0.0"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /test:
    get:
      summary: "Test endpoint"
`

	err = os.WriteFile(yamlFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Test loading
	fileInfo, err := LoadFile(yamlFile)
	require.NoError(t, err)

	assert.Equal(t, yamlFile, fileInfo.Path)
	assert.Equal(t, FileTypeYAML, fileInfo.Type)
	assert.Equal(t, "3.0.0", fileInfo.Content["openapi"])
	
	info, ok := fileInfo.Content["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test API", info["title"])
	assert.Equal(t, "1.0.0", info["version"])
}

func TestLoadFile_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tyk-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadFile(filepath.Join(tmpDir, "nonexistent.json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file does not exist")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonFile := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(jsonFile, []byte(`{invalid json`), 0644)
		require.NoError(t, err)

		_, err = LoadFile(jsonFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		yamlFile := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(yamlFile, []byte(`invalid: yaml: content:`), 0644)
		require.NoError(t, err)

		_, err = LoadFile(yamlFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})
}

func TestLoadFileAsRawJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tyk-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create YAML file
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `
openapi: "3.0.0"
info:
  title: "Test API"
  version: "1.0.0"
`

	err = os.WriteFile(yamlFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load as raw JSON
	rawJSON, err := LoadFileAsRawJSON(yamlFile)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(rawJSON, &parsed)
	require.NoError(t, err)
	
	assert.Equal(t, "3.0.0", parsed["openapi"])
	info, ok := parsed["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test API", info["title"])
}

func TestSaveFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tyk-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("save JSON", func(t *testing.T) {
		jsonFile := filepath.Join(tmpDir, "output.json")
		err := SaveFile(jsonFile, sampleOAS)
		require.NoError(t, err)

		// Verify file was created and contains correct content
		fileInfo, err := LoadFile(jsonFile)
		require.NoError(t, err)
		assert.Equal(t, "3.0.0", fileInfo.Content["openapi"])
	})

	t.Run("save YAML", func(t *testing.T) {
		yamlFile := filepath.Join(tmpDir, "output.yaml")
		err := SaveFile(yamlFile, sampleOAS)
		require.NoError(t, err)

		// Verify file was created and contains correct content
		fileInfo, err := LoadFile(yamlFile)
		require.NoError(t, err)
		assert.Equal(t, "3.0.0", fileInfo.Content["openapi"])
	})

	t.Run("create directory", func(t *testing.T) {
		nestedFile := filepath.Join(tmpDir, "subdir", "nested.json")
		err := SaveFile(nestedFile, sampleOAS)
		require.NoError(t, err)

		// Verify directory was created
		_, err = os.Stat(filepath.Dir(nestedFile))
		assert.NoError(t, err)

		// Verify file content
		fileInfo, err := LoadFile(nestedFile)
		require.NoError(t, err)
		assert.Equal(t, "3.0.0", fileInfo.Content["openapi"])
	})
}

func TestOASHelpers(t *testing.T) {
	t.Run("GetOASVersion", func(t *testing.T) {
		content1 := map[string]interface{}{"openapi": "3.0.0"}
		assert.Equal(t, "3.0.0", GetOASVersion(content1))

		content2 := map[string]interface{}{"swagger": "2.0"}
		assert.Equal(t, "2.0", GetOASVersion(content2))

		content3 := map[string]interface{}{"other": "field"}
		assert.Equal(t, "", GetOASVersion(content3))
	})

	t.Run("GetOASInfo", func(t *testing.T) {
		info := GetOASInfo(sampleOAS)
		require.NotNil(t, info)
		assert.Equal(t, "Test API", info["title"])
		assert.Equal(t, "1.0.0", info["version"])

		emptyContent := map[string]interface{}{}
		assert.Nil(t, GetOASInfo(emptyContent))
	})

	t.Run("GetOASInfoVersion", func(t *testing.T) {
		version := GetOASInfoVersion(sampleOAS)
		assert.Equal(t, "1.0.0", version)

		emptyContent := map[string]interface{}{}
		assert.Equal(t, "", GetOASInfoVersion(emptyContent))
	})

	t.Run("GetOASTitle", func(t *testing.T) {
		title := GetOASTitle(sampleOAS)
		assert.Equal(t, "Test API", title)

		emptyContent := map[string]interface{}{}
		assert.Equal(t, "", GetOASTitle(emptyContent))
	})
}

func TestConvertToJSON(t *testing.T) {
	jsonBytes, err := ConvertToJSON(sampleOAS)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "3.0.0", parsed["openapi"])
	info, ok := parsed["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test API", info["title"])
}

func TestConvertToYAML(t *testing.T) {
	yamlBytes, err := ConvertToYAML(sampleOAS)
	require.NoError(t, err)

	// Verify YAML output is not empty and contains expected content
	assert.NotEmpty(t, yamlBytes)
	assert.Contains(t, string(yamlBytes), "openapi: 3.0.0")
	assert.Contains(t, string(yamlBytes), "title: Test API")
}

func TestRealOASFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tyk-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a real-world OAS file
	realOAS := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Petstore API",
			"version": "1.0.0",
			"description": "A sample API that uses a petstore as an example",
		},
		"servers": []map[string]interface{}{
			{"url": "http://petstore.swagger.io/v1"},
		},
		"paths": map[string]interface{}{
			"/pets": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List all pets",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "A list of pets",
						},
					},
				},
				"post": map[string]interface{}{
					"summary": "Create a pet",
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Pet created",
						},
					},
				},
			},
		},
	}

	// Save as YAML and load back as JSON
	yamlFile := filepath.Join(tmpDir, "petstore.yaml")
	err = SaveFile(yamlFile, realOAS)
	require.NoError(t, err)

	rawJSON, err := LoadFileAsRawJSON(yamlFile)
	require.NoError(t, err)

	// Verify we can parse it back
	var loaded map[string]interface{}
	err = json.Unmarshal(rawJSON, &loaded)
	require.NoError(t, err)

	assert.Equal(t, "3.0.0", GetOASVersion(loaded))
	assert.Equal(t, "Petstore API", GetOASTitle(loaded))
	assert.Equal(t, "1.0.0", GetOASInfoVersion(loaded))

	t.Logf("✓ Successfully processed real OAS file with %d paths", len(loaded["paths"].(map[string]interface{})))
}