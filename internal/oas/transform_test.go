package oas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasTykExtensions(t *testing.T) {
	tests := []struct {
		name     string
		oasDoc   map[string]interface{}
		expected bool
	}{
		{
			name: "has extensions",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "Test API"},
				},
			},
			expected: true,
		},
		{
			name: "no extensions",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"title": "Test API"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasTykExtensions(tt.oasDoc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAPIIDFromTykExtensions(t *testing.T) {
	tests := []struct {
		name     string
		oasDoc   map[string]interface{}
		expectedID string
		expectedFound bool
	}{
		{
			name: "has ID",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"id": "api-123"},
				},
			},
			expectedID: "api-123",
			expectedFound: true,
		},
		{
			name: "no extensions",
			oasDoc: map[string]interface{}{
				"info": map[string]interface{}{"title": "Test API"},
			},
			expectedID: "",
			expectedFound: false,
		},
		{
			name: "has extensions but no ID",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "Test API"},
				},
			},
			expectedID: "",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, found := ExtractAPIIDFromTykExtensions(tt.oasDoc)
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

func TestAddTykExtensions(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		checkResult func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "plain OAS document",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title":   "Swagger Petstore",
					"version": "1.0.0",
				},
				"servers": []interface{}{
					map[string]interface{}{"url": "http://petstore.swagger.io/v2"},
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Should have extensions now
				assert.True(t, HasTykExtensions(result))
				
				// Check generated extensions
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				info, ok := tykExt["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Swagger Petstore", info["name"])
				
				upstream, ok := tykExt["upstream"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "http://petstore.swagger.io/v2", upstream["url"])
				
				server, ok := tykExt["server"].(map[string]interface{})
				require.True(t, ok)
				listenPath, ok := server["listenPath"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "/swagger-petstore/", listenPath["value"])
				assert.Equal(t, true, listenPath["strip"])
			},
		},
		{
			name: "already has extensions",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title": "Test API",
				},
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"name": "Existing"},
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Should keep existing extensions unchanged
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				info, ok := tykExt["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Existing", info["name"])
			},
		},
		{
			name: "missing info section",
			input: map[string]interface{}{
				"openapi": "3.0.0",
			},
			expectError: true,
		},
		{
			name: "missing title",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info":    map[string]interface{}{"version": "1.0.0"},
			},
			expectError: true,
		},
		{
			name: "missing servers",
			input: map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title": "Test API",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddTykExtensions(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestExtractVersioningMetadata(t *testing.T) {
	tests := []struct {
		name     string
		oasDoc   map[string]interface{}
		expected VersioningMetadata
	}{
		{
			name: "full versioning section",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{
						"versioning": map[string]interface{}{
							"base_api_id":  "abc123",
							"version_name": "v3",
							"set_default":  false,
						},
					},
				},
			},
			expected: VersioningMetadata{
				BaseAPIID:     "abc123",
				VersionName:   "v3",
				SetDefault:    false,
				IsVersionFile: true,
			},
		},
		{
			name: "no tyk extensions",
			oasDoc: map[string]interface{}{
				"openapi": "3.0.0",
			},
			expected: VersioningMetadata{IsVersionFile: false},
		},
		{
			name: "no versioning section",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{
						"id": "some-id",
					},
				},
			},
			expected: VersioningMetadata{IsVersionFile: false},
		},
		{
			name: "partial versioning - only base_api_id",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{
						"versioning": map[string]interface{}{
							"base_api_id": "xyz789",
						},
					},
				},
			},
			expected: VersioningMetadata{
				BaseAPIID:     "xyz789",
				IsVersionFile: true,
			},
		},
		{
			name: "versioning with set_default true",
			oasDoc: map[string]interface{}{
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{
						"versioning": map[string]interface{}{
							"base_api_id":  "def456",
							"version_name": "v1",
							"set_default":  true,
						},
					},
				},
			},
			expected: VersioningMetadata{
				BaseAPIID:     "def456",
				VersionName:   "v1",
				SetDefault:    true,
				IsVersionFile: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVersioningMetadata(tt.oasDoc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateListenPath(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Swagger Petstore", "/swagger-petstore/"},
		{"My API", "/my-api/"},
		{"User Management API v2", "/user-management-api-v2/"},
		{"123 Test API", "/api-123-test-api/"},
		{"", "/api/"},
		{"Special!@#$%Characters", "/special-characters/"},
		{"API", "/api/"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := GenerateListenPath(tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}