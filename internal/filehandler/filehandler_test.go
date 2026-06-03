package filehandler

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

// reqproof:req REQ-API-031
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

// reqproof:req REQ-API-031
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

// reqproof:req REQ-API-031
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

// reqproof:req REQ-API-031
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

// reqproof:req REQ-API-031
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

	// Exercises the os.ReadFile error branch (filehandler.go:41 err != nil = T).
	// os.Stat reports the file exists, but the read fails because permissions
	// prevent access. Skipped when running as root since root bypasses chmod.
	t.Run("read failure (permission denied)", func(t *testing.T) {
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("chmod permission semantics not applicable on this platform/user")
		}

		jsonFile := filepath.Join(tmpDir, "noread.json")
		err := os.WriteFile(jsonFile, []byte(`{"openapi":"3.0.0"}`), 0644)
		require.NoError(t, err)
		require.NoError(t, os.Chmod(jsonFile, 0))
		defer func() { _ = os.Chmod(jsonFile, 0644) }()

		_, err = LoadFile(jsonFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	// Exercises the getFileType error branch (filehandler.go:47 err != nil = T).
	// The file exists (os.Stat OK) and is readable, but the extension is not
	// in SupportedExtensions, so getFileType returns an error.
	t.Run("unsupported extension on existing file", func(t *testing.T) {
		txtFile := filepath.Join(tmpDir, "data.txt")
		err := os.WriteFile(txtFile, []byte(`anything`), 0644)
		require.NoError(t, err)

		_, err = LoadFile(txtFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported file extension")
	})
}

// reqproof:req REQ-API-031
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

// reqproof:req REQ-API-031
// Covers error branches in LoadFileAsRawJSON that the happy-path test cannot
// reach: the LoadFile failure path and the json.Marshal failure path.
func TestLoadFileAsRawJSON_Errors(t *testing.T) {
	tmpDir := t.TempDir()

	// Exercises filehandler.go:77 err != nil = T: LoadFile fails on a missing
	// file, so LoadFileAsRawJSON propagates the error before reaching marshal.
	t.Run("LoadFile failure propagates", func(t *testing.T) {
		_, err := LoadFileAsRawJSON(filepath.Join(tmpDir, "missing.json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file does not exist")
	})

	// Exercises filehandler.go:83 err != nil = T: yaml.v3 happily decodes
	// `.nan` into math.NaN(), but json.Marshal rejects NaN/Inf, so the final
	// marshal step fails.
	t.Run("json.Marshal fails on NaN from YAML", func(t *testing.T) {
		yamlFile := filepath.Join(tmpDir, "nan.yaml")
		err := os.WriteFile(yamlFile, []byte("value: .nan\n"), 0644)
		require.NoError(t, err)

		_, err = LoadFileAsRawJSON(yamlFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal to JSON")
	})
}

// reqproof:req REQ-POL-005
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

	// Exercises filehandler.go:93 err != nil = T: getFileType inside SaveFile
	// rejects unsupported extensions before any marshalling happens.
	t.Run("unsupported extension", func(t *testing.T) {
		err := SaveFile(filepath.Join(tmpDir, "out.txt"), sampleOAS)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported file extension")
	})

	// Exercises filehandler.go:102 err != nil = T: json.MarshalIndent fails on
	// values that JSON cannot represent (NaN here).
	t.Run("json marshal failure on NaN", func(t *testing.T) {
		bad := map[string]interface{}{"value": math.NaN()}
		err := SaveFile(filepath.Join(tmpDir, "nan.json"), bad)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal JSON")
	})

	// Exercises filehandler.go:116 err != nil = T: os.MkdirAll fails because
	// the would-be parent directory already exists as a regular file.
	t.Run("mkdir failure (parent is a file)", func(t *testing.T) {
		blocker := filepath.Join(tmpDir, "blocker")
		require.NoError(t, os.WriteFile(blocker, []byte("not a dir"), 0644))

		// Target asks SaveFile to MkdirAll(<tmpDir>/blocker) which is a file.
		target := filepath.Join(blocker, "child.json")
		err := SaveFile(target, sampleOAS)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})

	// Exercises filehandler.go:122 err != nil = T: os.WriteFile fails when the
	// final path is itself an existing directory.
	t.Run("write failure (path is a directory)", func(t *testing.T) {
		dirAsFile := filepath.Join(tmpDir, "dir-as-file.json")
		require.NoError(t, os.Mkdir(dirAsFile, 0755))

		err := SaveFile(dirAsFile, sampleOAS)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write file")
	})
}

// reqproof:req REQ-API-013
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

		// Exercises filehandler.go:167 ok = F: info.version is present but
		// is not a string, so the type assertion fails and we fall through
		// to the empty-string return.
		nonStringVersion := map[string]interface{}{
			"info": map[string]interface{}{
				"version": 1.0,
			},
		}
		assert.Equal(t, "", GetOASInfoVersion(nonStringVersion))
	})

	t.Run("GetOASTitle", func(t *testing.T) {
		title := GetOASTitle(sampleOAS)
		assert.Equal(t, "Test API", title)

		emptyContent := map[string]interface{}{}
		assert.Equal(t, "", GetOASTitle(emptyContent))

		// Exercises filehandler.go:177 ok = F: info.title is present but is
		// not a string, so the type assertion fails and we return "".
		nonStringTitle := map[string]interface{}{
			"info": map[string]interface{}{
				"title": 42,
			},
		}
		assert.Equal(t, "", GetOASTitle(nonStringTitle))
	})
}

// reqproof:req REQ-API-031
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

// reqproof:req REQ-API-031
func TestConvertToYAML(t *testing.T) {
	yamlBytes, err := ConvertToYAML(sampleOAS)
	require.NoError(t, err)

	// Verify YAML output is not empty and contains expected content
	assert.NotEmpty(t, yamlBytes)
	assert.Contains(t, string(yamlBytes), "openapi: 3.0.0")
	assert.Contains(t, string(yamlBytes), "title: Test API")
}

// reqproof:req REQ-API-031
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