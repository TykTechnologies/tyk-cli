package filehandler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SupportedExtensions lists the file extensions we support
var SupportedExtensions = []string{".yaml", ".yml", ".json"}

// FileType represents the type of file content
type FileType int

const (
	FileTypeJSON FileType = iota
	FileTypeYAML
)

// FileInfo contains information about a loaded file
type FileInfo struct {
	Path     string
	Type     FileType
	Content  map[string]interface{}
	RawBytes []byte
}

// Implements: SYS-REQ-022
func LoadFile(filePath string) (*FileInfo, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Determine file type from extension
	fileType, err := getFileType(filePath)
	if err != nil {
		return nil, err
	}

	// Parse content based on file type
	var parsedContent map[string]interface{}
	switch fileType {
	case FileTypeJSON:
		if err := json.Unmarshal(content, &parsedContent); err != nil {
			return nil, fmt.Errorf("failed to parse JSON file %s: %w", filePath, err)
		}
	case FileTypeYAML:
		if err := yaml.Unmarshal(content, &parsedContent); err != nil {
			return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
		}
	default:
		return nil, fmt.Errorf("unsupported file type for %s (supported: %v)", filePath, SupportedExtensions)
	}

	return &FileInfo{
		Path:     filePath,
		Type:     fileType,
		Content:  parsedContent,
		RawBytes: content,
	}, nil
}

// Implements: SYS-REQ-022
func LoadFileAsRawJSON(filePath string) (json.RawMessage, error) {
	fileInfo, err := LoadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Convert parsed content back to JSON
	jsonBytes, err := json.Marshal(fileInfo.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return json.RawMessage(jsonBytes), nil
}

// Implements: SYS-REQ-028
func SaveFile(filePath string, content map[string]interface{}) error {
	fileType, err := getFileType(filePath)
	if err != nil {
		return err
	}

	var data []byte

	switch fileType {
	case FileTypeJSON:
		data, err = json.MarshalIndent(content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	case FileTypeYAML:
		data, err = yaml.Marshal(content)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file type for %s (supported: %v)", filePath, SupportedExtensions)
	}

	// Create directory if it doesn't exist
	if dir := filepath.Dir(filePath); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// Implements: SYS-REQ-022
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supportedExt := range SupportedExtensions {
		if ext == supportedExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file extension %s (supported: %v)", ext, SupportedExtensions)
}

// Implements: SYS-REQ-011
func GetOASVersion(content map[string]interface{}) string {
	if openapi, ok := content["openapi"].(string); ok {
		return openapi
	}
	if swagger, ok := content["swagger"].(string); ok {
		return swagger
	}
	return ""
}

// Implements: SYS-REQ-011
func GetOASInfo(content map[string]interface{}) map[string]interface{} {
	if info, ok := content["info"].(map[string]interface{}); ok {
		return info
	}
	return nil
}

// Implements: SYS-REQ-011
func GetOASInfoVersion(content map[string]interface{}) string {
	if info := GetOASInfo(content); info != nil {
		if version, ok := info["version"].(string); ok {
			return version
		}
	}
	return ""
}

// Implements: SYS-REQ-011
func GetOASTitle(content map[string]interface{}) string {
	if info := GetOASInfo(content); info != nil {
		if title, ok := info["title"].(string); ok {
			return title
		}
	}
	return ""
}

// Implements: SYS-REQ-022
func getFileType(filePath string) (FileType, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return FileTypeJSON, nil
	case ".yaml", ".yml":
		return FileTypeYAML, nil
	default:
		return 0, fmt.Errorf("unsupported file extension %q (supported: %v)", ext, SupportedExtensions)
	}
}

// Implements: SYS-REQ-022
func ConvertToJSON(content map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(content, "", "  ")
}

// Implements: SYS-REQ-022
func ConvertToYAML(content map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(content)
}